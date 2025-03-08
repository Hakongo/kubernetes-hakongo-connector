package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/hakongo/kubernetes-connector/internal/metrics"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PodCollector struct {
	kubeClient       kubernetes.Interface
	prometheusClient *metrics.PrometheusClient
	config           CollectorConfig
}

func NewPodCollector(kubeClient kubernetes.Interface, prometheusClient *metrics.PrometheusClient, config CollectorConfig) *PodCollector {
	return &PodCollector{
		kubeClient:       kubeClient,
		prometheusClient: prometheusClient,
		config:           config,
	}
}

func (c *PodCollector) Name() string {
	return "pod_collector"
}

func (c *PodCollector) Description() string {
	return "Collects resource metrics from Kubernetes pods"
}

func (c *PodCollector) Collect(ctx context.Context) ([]ResourceMetrics, error) {
	pods, err := c.kubeClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var metrics []ResourceMetrics

	for _, pod := range pods.Items {
		// Skip pods in excluded namespaces
		if contains(c.config.ExcludeNamespaces, pod.Namespace) {
			continue
		}

		// Skip if namespace is not included (when inclusion list is not empty)
		if len(c.config.IncludeNamespaces) > 0 && !contains(c.config.IncludeNamespaces, pod.Namespace) {
			continue
		}

		// Get pod metrics from Prometheus
		containerMetrics, err := c.prometheusClient.GetContainerMetrics(ctx, pod.Namespace, pod.Name)
		if err != nil {
			// Log error but continue collecting other pod metrics
			fmt.Printf("Error getting metrics for pod %s/%s: %v\n", pod.Namespace, pod.Name, err)
			continue
		}

		// Create pod metrics
		podMetrics := ResourceMetrics{
			Name:        pod.Name,
			Namespace:   pod.Namespace,
			Kind:        "Pod",
			Labels:      pod.Labels,
			CollectedAt: time.Now(),
			Status: map[string]interface{}{
				"phase":     string(pod.Status.Phase),
				"hostIP":    pod.Status.HostIP,
				"podIP":     pod.Status.PodIP,
				"startTime": pod.Status.StartTime,
			},
		}

		// Add container metrics
		for _, container := range pod.Spec.Containers {
			containerName := container.Name
			cpuUsage := containerMetrics.CPU[containerName]
			memoryUsage := containerMetrics.Memory[containerName]

			containerStatus := getContainerStatus(pod.Status.ContainerStatuses, containerName)

			podMetrics.Containers = append(podMetrics.Containers, ContainerMetrics{
				Name: containerName,
				CPU: CPUMetrics{
					UsageNanoCores:    int64(cpuUsage * 1e9), // Convert cores to nanocores
					UsageCorePercent:  cpuUsage * 100,
					RequestMilliCores: getResourceMilliValue(container.Resources.Requests, corev1.ResourceCPU),
					LimitMilliCores:   getResourceMilliValue(container.Resources.Limits, corev1.ResourceCPU),
				},
				Memory: MemoryMetrics{
					UsageBytes:   int64(memoryUsage),
					RequestBytes: getResourceByteValue(container.Resources.Requests, corev1.ResourceMemory),
					LimitBytes:   getResourceByteValue(container.Resources.Limits, corev1.ResourceMemory),
				},
				Ready:    containerStatus.Ready,
				Restarts: containerStatus.RestartCount,
				State:    getContainerState(containerStatus.State),
			})
		}

		metrics = append(metrics, podMetrics)
	}

	return metrics, nil
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func getContainerStatus(statuses []corev1.ContainerStatus, name string) corev1.ContainerStatus {
	for _, status := range statuses {
		if status.Name == name {
			return status
		}
	}
	return corev1.ContainerStatus{}
}

func getContainerState(state corev1.ContainerState) string {
	switch {
	case state.Running != nil:
		return "running"
	case state.Waiting != nil:
		return "waiting"
	case state.Terminated != nil:
		return "terminated"
	default:
		return "unknown"
	}
}

func getResourceMilliValue(resources corev1.ResourceList, resourceName corev1.ResourceName) int64 {
	if val, ok := resources[resourceName]; ok {
		return val.MilliValue()
	}
	return 0
}

func getResourceByteValue(resources corev1.ResourceList, resourceName corev1.ResourceName) int64 {
	if val, ok := resources[resourceName]; ok {
		return val.Value()
	}
	return 0
}
