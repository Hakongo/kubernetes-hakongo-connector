package collector

import (
	"context"
	"fmt"
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

	podMetricsMap := make(map[string]*corev1.Pod)
	for _, pod := range pods.Items {
		key := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		podMetricsMap[key] = &pod
	}

	for _, podMetric := range podMetrics.Items {
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

		metric.CPU = pc.calculateCPUMetrics(&podMetric)
		metric.Memory = pc.calculateMemoryMetrics(&podMetric)
		metric.Storage = pc.calculateStorageMetrics(pod)
		metric.Network = pc.calculateNetworkMetrics(&podMetric)
		metric.Cost = pc.calculateCostMetrics(metric.CPU, metric.Memory)

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
		}
	}
	return storage
}

func (pc *PodCollector) calculateNetworkMetrics(_ *metricsv1beta1.PodMetrics) NetworkMetrics {
	return NetworkMetrics{}
}

func (pc *PodCollector) calculateCostMetrics(cpu CPUMetrics, memory MemoryMetrics) CostMetrics {
	return CostMetrics{
		Currency:    "USD",
		CPUCost:    cpu.UsageCorePercent * 0.04,
		MemoryCost: float64(memory.UsageBytes) / float64(1<<30) * 0.01,
	}
}
