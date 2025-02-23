package collector

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type NamespaceCollector struct {
	kubeClient kubernetes.Interface
	config     CollectorConfig
}

func NewNamespaceCollector(kubeClient kubernetes.Interface, config CollectorConfig) *NamespaceCollector {
	return &NamespaceCollector{
		kubeClient: kubeClient,
		config:     config,
	}
}

func (nc *NamespaceCollector) Name() string { return "namespace-collector" }

func (nc *NamespaceCollector) Description() string {
	return "Collects metadata and status for Kubernetes namespaces"
}

func (nc *NamespaceCollector) Collect(ctx context.Context) ([]ResourceMetrics, error) {
	var metrics []ResourceMetrics

	namespaces, err := nc.kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	for _, ns := range namespaces.Items {
		if nc.isNamespaceExcluded(ns.Name) {
			continue
		}

		metric := ResourceMetrics{
			Name:        ns.Name,
			Kind:        "Namespace",
			Labels:      ns.Labels,
			CollectedAt: time.Now(),
			Status: map[string]interface{}{
				"phase":     string(ns.Status.Phase),
				"age":       time.Since(ns.CreationTimestamp.Time).String(),
				"finalizer": ns.Spec.Finalizers,
			},
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (nc *NamespaceCollector) isNamespaceExcluded(namespace string) bool {
	for _, excluded := range nc.config.ExcludeNamespaces {
		if namespace == excluded {
			return true
		}
	}

	if len(nc.config.IncludeNamespaces) == 0 {
		return false
	}

	for _, included := range nc.config.IncludeNamespaces {
		if namespace == included {
			return false
		}
	}

	return true
}
