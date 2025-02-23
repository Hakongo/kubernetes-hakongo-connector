package collector

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type WorkloadCollector struct {
	kubeClient kubernetes.Interface
	config     CollectorConfig
}

func NewWorkloadCollector(kubeClient kubernetes.Interface, config CollectorConfig) *WorkloadCollector {
	return &WorkloadCollector{
		kubeClient: kubeClient,
		config:     config,
	}
}

func (wc *WorkloadCollector) Name() string { return "workload-collector" }

func (wc *WorkloadCollector) Description() string {
	return "Collects metrics for Kubernetes workloads (Deployments, StatefulSets, DaemonSets)"
}

func (wc *WorkloadCollector) Collect(ctx context.Context) ([]ResourceMetrics, error) {
	var metrics []ResourceMetrics

	// Collect Deployments
	deployments, err := wc.kubeClient.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	for _, deploy := range deployments.Items {
		if wc.isNamespaceExcluded(deploy.Namespace) {
			continue
		}

		metric := ResourceMetrics{
			Name:        deploy.Name,
			Namespace:   deploy.Namespace,
			Kind:        "Deployment",
			Labels:      deploy.Labels,
			CollectedAt: time.Now(),
			Status: map[string]interface{}{
				"replicas":            deploy.Status.Replicas,
				"availableReplicas":   deploy.Status.AvailableReplicas,
				"updatedReplicas":     deploy.Status.UpdatedReplicas,
				"readyReplicas":       deploy.Status.ReadyReplicas,
				"observedGeneration":  deploy.Status.ObservedGeneration,
				"conditions":          deploy.Status.Conditions,
				"collisionCount":      deploy.Status.CollisionCount,
				"strategy":            deploy.Spec.Strategy.Type,
				"minReadySeconds":     deploy.Spec.MinReadySeconds,
				"revisionHistoryLimit": deploy.Spec.RevisionHistoryLimit,
			},
		}

		metrics = append(metrics, metric)
	}

	// Collect StatefulSets
	statefulsets, err := wc.kubeClient.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list statefulsets: %w", err)
	}

	for _, sts := range statefulsets.Items {
		if wc.isNamespaceExcluded(sts.Namespace) {
			continue
		}

		metric := ResourceMetrics{
			Name:        sts.Name,
			Namespace:   sts.Namespace,
			Kind:        "StatefulSet",
			Labels:      sts.Labels,
			CollectedAt: time.Now(),
			Status: map[string]interface{}{
				"replicas":           sts.Status.Replicas,
				"readyReplicas":      sts.Status.ReadyReplicas,
				"currentReplicas":    sts.Status.CurrentReplicas,
				"updatedReplicas":    sts.Status.UpdatedReplicas,
				"observedGeneration": sts.Status.ObservedGeneration,
				"conditions":         sts.Status.Conditions,
				"updateStrategy":     sts.Spec.UpdateStrategy.Type,
				"serviceName":        sts.Spec.ServiceName,
			},
		}

		metrics = append(metrics, metric)
	}

	// Collect DaemonSets
	daemonsets, err := wc.kubeClient.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list daemonsets: %w", err)
	}

	for _, ds := range daemonsets.Items {
		if wc.isNamespaceExcluded(ds.Namespace) {
			continue
		}

		metric := ResourceMetrics{
			Name:        ds.Name,
			Namespace:   ds.Namespace,
			Kind:        "DaemonSet",
			Labels:      ds.Labels,
			CollectedAt: time.Now(),
			Status: map[string]interface{}{
				"desiredNumberScheduled": ds.Status.DesiredNumberScheduled,
				"currentNumberScheduled": ds.Status.CurrentNumberScheduled,
				"numberReady":           ds.Status.NumberReady,
				"updatedNumberScheduled": ds.Status.UpdatedNumberScheduled,
				"numberAvailable":       ds.Status.NumberAvailable,
				"numberUnavailable":     ds.Status.NumberUnavailable,
				"observedGeneration":    ds.Status.ObservedGeneration,
				"conditions":            ds.Status.Conditions,
				"updateStrategy":        ds.Spec.UpdateStrategy.Type,
			},
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (wc *WorkloadCollector) isNamespaceExcluded(namespace string) bool {
	for _, excluded := range wc.config.ExcludeNamespaces {
		if namespace == excluded {
			return true
		}
	}

	if len(wc.config.IncludeNamespaces) == 0 {
		return false
	}

	for _, included := range wc.config.IncludeNamespaces {
		if namespace == included {
			return false
		}
	}

	return true
}
