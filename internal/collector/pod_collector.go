package collector

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

type PodCollector struct {
	kubeClient    kubernetes.Interface
	metricsClient versioned.Interface
	config        CollectorConfig
}

func NewPodCollector(kubeClient kubernetes.Interface, metricsClient versioned.Interface, config CollectorConfig) *PodCollector {
	return &PodCollector{
		kubeClient:    kubeClient,
		metricsClient: metricsClient,
		config:        config,
	}
}

func (pc *PodCollector) Name() string { return "pod-collector" }

func (pc *PodCollector) Description() string { return "Collects resource usage metrics for Kubernetes pods" }

func (pc *PodCollector) Collect(ctx context.Context) ([]ResourceMetrics, error) {
	var metrics []ResourceMetrics

	pods, err := pc.kubeClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	podMetrics, err := pc.metricsClient.MetricsV1beta1().PodMetricses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %w", err)
	}

	podMetricsMap := make(map[string]*metricsv1beta1.PodMetrics)
	for _, metric := range podMetrics.Items {
		key := fmt.Sprintf("%s/%s", metric.Namespace, metric.Name)
		podMetricsMap[key] = &metric
	}

	for _, pod := range pods.Items {
		if pc.isNamespaceExcluded(pod.Namespace) {
			continue
		}

		metric := ResourceMetrics{
			Name:        pod.Name,
			Namespace:   pod.Namespace,
			Kind:        "Pod",
			Labels:      pod.Labels,
			CollectedAt: time.Now(),
		}

		podMetric, hasMetrics := podMetricsMap[fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)]
		if hasMetrics {
			metric.CPU = pc.calculateCPUMetrics(podMetric)
			metric.Memory = pc.calculateMemoryMetrics(podMetric)
			metric.Network = pc.calculateNetworkMetrics(podMetric)
			metric.Cost = pc.calculateCostMetrics(metric.CPU, metric.Memory)
		}

		metric.Storage = pc.calculateStorageMetrics(&pod)
		metric.Containers = pc.calculateContainerMetrics(&pod, podMetric)

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (pc *PodCollector) isNamespaceExcluded(namespace string) bool {
	for _, excluded := range pc.config.ExcludeNamespaces {
		if namespace == excluded {
			return true
		}
	}

	if len(pc.config.IncludeNamespaces) == 0 {
		return false
	}

	for _, included := range pc.config.IncludeNamespaces {
		if namespace == included {
			return false
		}
	}

	return true
}

func (pc *PodCollector) calculateCPUMetrics(metrics *metricsv1beta1.PodMetrics) CPUMetrics {
	var cpu CPUMetrics
	for _, container := range metrics.Containers {
		cpu.UsageNanoCores += container.Usage.Cpu().Value()
		if container.Usage.Cpu().Value() > 0 {
			cpu.UsageCorePercent = float64(container.Usage.Cpu().Value()) / float64(1e9)
		}
	}
	return cpu
}

func (pc *PodCollector) calculateMemoryMetrics(metrics *metricsv1beta1.PodMetrics) MemoryMetrics {
	var mem MemoryMetrics
	for _, container := range metrics.Containers {
		mem.UsageBytes += container.Usage.Memory().Value()
	}
	return mem
}

func (pc *PodCollector) calculateStorageMetrics(pod *corev1.Pod) StorageMetrics {
	var storage StorageMetrics
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			storage.PVCName = volume.PersistentVolumeClaim.ClaimName
			storage.DiskPressure = pod.Status.Phase == corev1.PodPending &&
				pod.Status.Reason == "Unschedulable" &&
				pod.Status.Message != "" &&
				strings.Contains(pod.Status.Message, "disk pressure")
		}
	}
	return storage
}

func (pc *PodCollector) calculateNetworkMetrics(metrics *metricsv1beta1.PodMetrics) NetworkMetrics {
	var network NetworkMetrics
	for _, container := range metrics.Containers {
		if container.Usage.Memory().Value() > 0 {
			// Estimate network usage based on memory activity
			network.RxBytes += container.Usage.Memory().Value() / 10
			network.TxBytes += container.Usage.Memory().Value() / 20
		}
	}
	return network
}

func (pc *PodCollector) calculateCostMetrics(cpu CPUMetrics, memory MemoryMetrics) CostMetrics {
	const (
		cpuCostPerCore    = 0.04 // $0.04 per core hour
		memoryCostPerGB   = 0.01 // $0.01 per GB hour
		storageCostPerGB  = 0.002 // $0.002 per GB hour
		networkCostPerGB  = 0.05 // $0.05 per GB
	)

	cpuCost := cpu.UsageCorePercent * cpuCostPerCore
	memoryCost := float64(memory.UsageBytes) / float64(1<<30) * memoryCostPerGB

	return CostMetrics{
		Currency:    "USD",
		CPUCost:    cpuCost,
		MemoryCost: memoryCost,
		TotalCost:  cpuCost + memoryCost,
	}
}

func (pc *PodCollector) calculateContainerMetrics(pod *corev1.Pod, metrics *metricsv1beta1.PodMetrics) []ContainerMetrics {
	var containers []ContainerMetrics

	containerMetrics := make(map[string]*metricsv1beta1.ContainerMetrics)
	if metrics != nil {
		for i := range metrics.Containers {
			containerMetrics[metrics.Containers[i].Name] = &metrics.Containers[i]
		}
	}

	for _, container := range pod.Status.ContainerStatuses {
		metric := ContainerMetrics{
			Name:     container.Name,
			Ready:    container.Ready,
			Restarts: container.RestartCount,
		}

		if container.State.Running != nil {
			metric.State = "Running"
		} else if container.State.Waiting != nil {
			metric.State = "Waiting"
		} else if container.State.Terminated != nil {
			metric.State = "Terminated"
		}

		if containerMetric, ok := containerMetrics[container.Name]; ok {
			metric.CPU = CPUMetrics{
				UsageNanoCores:   containerMetric.Usage.Cpu().Value(),
				UsageCorePercent: float64(containerMetric.Usage.Cpu().Value()) / float64(1e9),
			}
			metric.Memory = MemoryMetrics{
				UsageBytes: containerMetric.Usage.Memory().Value(),
			}
		}

		containers = append(containers, metric)
	}

	return containers
}
