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

type ServiceCollector struct {
	kubeClient    kubernetes.Interface
	metricsClient versioned.Interface
	config        CollectorConfig
}

func NewServiceCollector(kubeClient kubernetes.Interface, metricsClient versioned.Interface, config CollectorConfig) *ServiceCollector {
	return &ServiceCollector{
		kubeClient:    kubeClient,
		metricsClient: metricsClient,
		config:        config,
	}
}

func (sc *ServiceCollector) Name() string { return "service-collector" }

func (sc *ServiceCollector) Description() string { return "Collects metrics for Kubernetes Services" }

func (sc *ServiceCollector) Collect(ctx context.Context) ([]ResourceMetrics, error) {
	var metrics []ResourceMetrics

	services, err := sc.kubeClient.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	// Get endpoints to check service health and connectivity
	endpoints, err := sc.kubeClient.CoreV1().Endpoints("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list endpoints: %w", err)
	}

	// Create endpoints map for quick lookups
	endpointMap := make(map[string]*corev1.Endpoints)
	for i := range endpoints.Items {
		endpoint := &endpoints.Items[i]
		key := fmt.Sprintf("%s/%s", endpoint.Namespace, endpoint.Name)
		endpointMap[key] = endpoint
	}

	for _, svc := range services.Items {
		// Skip services in excluded namespaces
		if sc.shouldSkipNamespace(svc.Namespace) {
			continue
		}

		metric := ResourceMetrics{
			Name:        svc.Name,
			Namespace:   svc.Namespace,
			Kind:        "Service",
			Labels:      svc.Labels,
			CollectedAt: time.Now(),
		}

		// Calculate network metrics based on service type and endpoints
		metric.Network = sc.calculateNetworkMetrics(&svc, endpointMap[fmt.Sprintf("%s/%s", svc.Namespace, svc.Name)])

		// Calculate cost based on service type and configuration
		metric.Cost = sc.calculateCostMetrics(&svc)

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (sc *ServiceCollector) shouldSkipNamespace(namespace string) bool {
	// Check if namespace is in exclude list
	for _, excluded := range sc.config.ExcludeNamespaces {
		if namespace == excluded {
			return true
		}
	}

	// If include list is empty, collect from all non-excluded namespaces
	if len(sc.config.IncludeNamespaces) == 0 {
		return false
	}

	// Check if namespace is in include list
	for _, included := range sc.config.IncludeNamespaces {
		if namespace == included {
			return false
		}
	}

	return true
}

func (sc *ServiceCollector) calculateNetworkMetrics(svc *corev1.Service, endpoints *corev1.Endpoints) NetworkMetrics {
	var net NetworkMetrics

	// Check if service has endpoints
	if endpoints != nil {
		for _, subset := range endpoints.Subsets {
			net.TxPackets += int64(len(subset.Addresses)) // Count ready endpoints
			if len(subset.NotReadyAddresses) > 0 {
				net.TxErrors += int64(len(subset.NotReadyAddresses))
			}
		}
	}

	// Check if service has external IP
	if len(svc.Status.LoadBalancer.Ingress) > 0 {
		net.RxPackets++ // Increment to indicate external connectivity
	}

	return net
}

func (sc *ServiceCollector) calculateCostMetrics(svc *corev1.Service) CostMetrics {
	var cost CostMetrics
	cost.Currency = "USD"

	// Base cost calculation based on service type
	switch svc.Spec.Type {
	case corev1.ServiceTypeLoadBalancer:
		cost.NetworkCost = 0.025 // $0.025 per hour for load balancer
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			cost.NetworkCost *= float64(len(svc.Status.LoadBalancer.Ingress))
		}
	case corev1.ServiceTypeNodePort:
		cost.NetworkCost = 0.010 // $0.010 per hour for NodePort
	}

	// Additional cost for session affinity
	if svc.Spec.SessionAffinity == corev1.ServiceAffinityClientIP {
		cost.NetworkCost *= 1.1 // 10% premium for session affinity
	}

	cost.TotalCost = cost.NetworkCost

	return cost
}
