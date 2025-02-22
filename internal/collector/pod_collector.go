package collector

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

// PodCollector collects metrics for Kubernetes pods
type PodCollector struct {
	kubeClient    kubernetes.Interface
	metricsClient versioned.Interface
	config        CollectorConfig
}

// NewPodCollector creates a new pod collector
func NewPodCollector(
	kubeClient kubernetes.Interface,
	metricsClient versioned.Interface,
	config CollectorConfig,
) *PodCollector {
	return &PodCollector{
		kubeClient:    kubeClient,
		metricsClient: metricsClient,
		config:        config,
	}
}

// Name returns the collector name
func (pc *PodCollector) Name() string {
	return "pod-collector"
}

// Description returns the collector description
func (pc *PodCollector) Description() string {
	return "Collects resource usage metrics for Kubernetes pods"
}

// Collect gathers metrics for all pods in the cluster
func (pc *PodCollector) Collect(ctx context.Context) ([]ResourceMetrics, error) {
	var metrics []ResourceMetrics

	// List pods in all or specified namespaces
	pods, err := pc.kubeClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	// Get pod metrics
	podMetrics, err := pc.metricsClient.MetricsV1beta1().PodMetricses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %w", err)
	}

	// Create a map for quick lookup of pod metrics
	podMetricsMap := make(map[string]*corev1.Pod)
	for _, pod := range pods.Items {
		key := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		podMetricsMap[key] = &pod
	}

	// Process each pod
	for _, podMetric := range podMetrics.Items {
		// Skip if pod is in excluded namespace
		if pc.isNamespaceExcluded(podMetric.Namespace) {
			continue
		}

		pod, exists := podMetricsMap[fmt.Sprintf("%s/%s", podMetric.Namespace, podMetric.Name)]
		if !exists {
			continue
		}

		metric := ResourceMetrics{
			Name:        podMetric.Name,
			Namespace:   podMetric.Namespace,
			Kind:        "Pod",
			Labels:      podMetric.Labels,
			CollectedAt: time.Now(),
		}

		// Calculate CPU metrics
		metric.CPU = pc.calculateCPUMetrics(pod, &podMetric)

		// Calculate Memory metrics
		metric.Memory = pc.calculateMemoryMetrics(pod, &podMetric)

		// Calculate Storage metrics
		metric.Storage = pc.calculateStorageMetrics(pod)

		// Calculate Network metrics
		metric.Network = pc.calculateNetworkMetrics(&podMetric)

		// Calculate Cost metrics
		metric.Cost = pc.calculateCostMetrics(metric.CPU, metric.Memory, metric.Storage, metric.Network)

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (pc *PodCollector) isNamespaceExcluded(namespace string) bool {
	// Check if namespace is in exclude list
	for _, excluded := range pc.config.ExcludeNamespaces {
		if namespace == excluded {
			return true
		}
	}

	// If include list is empty, include all namespaces
	if len(pc.config.IncludeNamespaces) == 0 {
		return false
	}

	// Check if namespace is in include list
	for _, included := range pc.config.IncludeNamespaces {
		if namespace == included {
			return false
		}
	}

	return true
}

// Helper functions to calculate specific metrics
func (pc *PodCollector) calculateCPUMetrics(pod *corev1.Pod, metrics *corev1.Pod) CPUMetrics {
	// Implementation for CPU metrics calculation
	return CPUMetrics{
		// Add implementation details
	}
}

func (pc *PodCollector) calculateMemoryMetrics(pod *corev1.Pod, metrics *corev1.Pod) MemoryMetrics {
	// Implementation for Memory metrics calculation
	return MemoryMetrics{
		// Add implementation details
	}
}

func (pc *PodCollector) calculateStorageMetrics(pod *corev1.Pod) StorageMetrics {
	// Implementation for Storage metrics calculation
	return StorageMetrics{
		// Add implementation details
	}
}

func (pc *PodCollector) calculateNetworkMetrics(metrics *corev1.Pod) NetworkMetrics {
	// Implementation for Network metrics calculation
	return NetworkMetrics{
		// Add implementation details
	}
}

func (pc *PodCollector) calculateCostMetrics(cpu CPUMetrics, memory MemoryMetrics, storage StorageMetrics, network NetworkMetrics) CostMetrics {
	// Implementation for Cost metrics calculation
	return CostMetrics{
		Currency: "USD",
		// Add implementation details
	}
}
