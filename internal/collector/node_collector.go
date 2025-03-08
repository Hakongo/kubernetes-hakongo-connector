package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/hakongo/kubernetes-connector/internal/metrics"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

type NodeCollector struct {
	kubeClient       kubernetes.Interface
	metricsClient    versioned.Interface
	prometheusClient *metrics.PrometheusClient
	config           CollectorConfig
	usePrometheus    bool
	useMetricsServer bool
}

func NewNodeCollector(kubeClient kubernetes.Interface, metricsClient versioned.Interface, prometheusClient *metrics.PrometheusClient, config CollectorConfig, usePrometheus bool, useMetricsServer bool) *NodeCollector {
	return &NodeCollector{
		kubeClient:       kubeClient,
		metricsClient:    metricsClient,
		prometheusClient: prometheusClient,
		config:           config,
		usePrometheus:    usePrometheus,
		useMetricsServer: useMetricsServer,
	}
}

func (nc *NodeCollector) Name() string { return "node-collector" }

func (nc *NodeCollector) Description() string {
	return "Collects resource usage metrics for Kubernetes nodes"
}

func (nc *NodeCollector) Collect(ctx context.Context) ([]ResourceMetrics, error) {
	var metrics []ResourceMetrics

	nodes, err := nc.kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Get metrics based on user configuration
	var nodePrometheusMetrics map[string]*PrometheusNodeMetrics
	var nodeK8sMetricsMap map[string]*metricsv1beta1.NodeMetrics

	// Get Prometheus metrics if configured
	if nc.usePrometheus && nc.prometheusClient != nil {
		nodePrometheusMetrics = make(map[string]*PrometheusNodeMetrics)

		// Get metrics for each node from Prometheus
		for _, node := range nodes.Items {
			promNodeMetric, err := nc.prometheusClient.GetNodeMetrics(ctx, node.Name)
			if err != nil {
				// Log error but continue with other nodes
				fmt.Printf("Warning: failed to get Prometheus metrics for node %s: %v\n", node.Name, err)
				continue
			}
			// Convert from Prometheus client NodeMetrics to our internal PrometheusNodeMetrics
			nodeMetric := &PrometheusNodeMetrics{
				CPUUsage:    promNodeMetric.CPUUsage,
				MemoryUsage: promNodeMetric.MemoryUsage,
			}
			nodePrometheusMetrics[node.Name] = nodeMetric
		}
	}

	// Get Kubernetes Metrics Server metrics if configured
	if nc.useMetricsServer {
		nodeMetrics, err := nc.metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
		if err != nil {
			// Log error but continue
			fmt.Printf("Warning: failed to get node metrics from Kubernetes metrics API: %v\n", err)
		} else {
			nodeK8sMetricsMap = make(map[string]*metricsv1beta1.NodeMetrics)
			for _, nodeMetric := range nodeMetrics.Items {
				nodeK8sMetricsMap[nodeMetric.Name] = &nodeMetric
			}
		}
	}

	for _, node := range nodes.Items {
		metric := ResourceMetrics{
			Name:        node.Name,
			Kind:        "Node",
			Labels:      node.Labels,
			CollectedAt: time.Now(),
		}

		// Get metrics from configured sources
		// First try Prometheus if enabled
		if nc.usePrometheus && nodePrometheusMetrics != nil {
			if promMetric, exists := nodePrometheusMetrics[node.Name]; exists {
				metric.CPU = nc.calculateCPUMetricsFromPrometheus(promMetric)
				metric.Memory = nc.calculateMemoryMetricsFromPrometheus(promMetric)
			}
		}

		// Then try Metrics Server if enabled
		// If we already have CPU/Memory from Prometheus, we'll keep those values
		if nc.useMetricsServer && nodeK8sMetricsMap != nil {
			if k8sMetric, exists := nodeK8sMetricsMap[node.Name]; exists {
				// Only set metrics if not already set by Prometheus
				if metric.CPU.UsageNanoCores == 0 {
					metric.CPU = nc.calculateCPUMetrics(k8sMetric)
				}
				if metric.Memory.UsageBytes == 0 {
					metric.Memory = nc.calculateMemoryMetrics(k8sMetric)
				}
			}
		}

		// These metrics don't depend on Prometheus or Kubernetes metrics API
		metric.Storage = nc.calculateStorageMetrics(&node)
		metric.Network = nc.calculateNetworkMetrics(&node)
		metric.Cost = nc.calculateCostMetrics(metric.CPU, metric.Memory, &node)

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (nc *NodeCollector) calculateCPUMetrics(metrics *metricsv1beta1.NodeMetrics) CPUMetrics {
	var cpu CPUMetrics
	cpu.UsageNanoCores = metrics.Usage.Cpu().Value()
	if cpu.UsageNanoCores > 0 {
		cpu.UsageCorePercent = float64(cpu.UsageNanoCores) / float64(1e9)
	}
	return cpu
}

// PrometheusNodeMetrics represents resource usage metrics for nodes from Prometheus
type PrometheusNodeMetrics struct {
	CPUUsage    float64 // CPU usage in cores
	MemoryUsage float64 // Memory usage in bytes
}

func (nc *NodeCollector) calculateCPUMetricsFromPrometheus(metrics *PrometheusNodeMetrics) CPUMetrics {
	var cpu CPUMetrics
	cpu.UsageNanoCores = int64(metrics.CPUUsage * 1e9) // Convert cores to nanocores
	cpu.UsageCorePercent = metrics.CPUUsage
	return cpu
}

func (nc *NodeCollector) calculateMemoryMetrics(metrics *metricsv1beta1.NodeMetrics) MemoryMetrics {
	return MemoryMetrics{
		UsageBytes: metrics.Usage.Memory().Value(),
	}
}

func (nc *NodeCollector) calculateMemoryMetricsFromPrometheus(metrics *PrometheusNodeMetrics) MemoryMetrics {
	return MemoryMetrics{
		UsageBytes: int64(metrics.MemoryUsage),
	}
}

func (nc *NodeCollector) calculateStorageMetrics(node *corev1.Node) StorageMetrics {
	var storage StorageMetrics
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeDiskPressure && condition.Status == corev1.ConditionTrue {
			storage.DiskPressure = true
			break
		}
	}
	return storage
}

func (nc *NodeCollector) calculateNetworkMetrics(node *corev1.Node) NetworkMetrics {
	var net NetworkMetrics
	// Get network metrics from node conditions and status
	for _, address := range node.Status.Addresses {
		if address.Type == corev1.NodeInternalIP {
			// Node has internal network connectivity
			net.TxPackets++ // Increment to indicate network is available
			break
		}
	}
	return net
}

func (nc *NodeCollector) calculateCostMetrics(cpu CPUMetrics, memory MemoryMetrics, node *corev1.Node) CostMetrics {
	// Calculate cost based on node size, usage, and instance type
	var allocatable = node.Status.Allocatable
	cpuCores := float64(allocatable.Cpu().Value()) / 1e9
	memoryGB := float64(allocatable.Memory().Value()) / float64(1<<30)

	// Base cost calculation
	cpuCost := cpuCores * 0.04    // $0.04 per core hour
	memoryCost := memoryGB * 0.01 // $0.01 per GB hour

	// Adjust cost based on actual usage
	if cpu.UsageCorePercent > 0 {
		cpuCost *= cpu.UsageCorePercent
	}
	if memory.UsageBytes > 0 {
		usageGB := float64(memory.UsageBytes) / float64(1<<30)
		memoryCost *= (usageGB / memoryGB)
	}

	// Get instance type for more accurate pricing
	instanceType := node.Labels["node.kubernetes.io/instance-type"]
	if instanceType != "" {
		// TODO: Implement more accurate pricing based on instance type
		// This would require a pricing table for different instance types
	}

	return CostMetrics{
		Currency:   "USD",
		CPUCost:    cpuCost,
		MemoryCost: memoryCost,
		TotalCost:  cpuCost + memoryCost,
	}
}
