package collector

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type IngressCollector struct {
	kubeClient kubernetes.Interface
	config     CollectorConfig
}

func NewIngressCollector(kubeClient kubernetes.Interface, config CollectorConfig) *IngressCollector {
	return &IngressCollector{
		kubeClient: kubeClient,
		config:     config,
	}
}

func (ic *IngressCollector) Name() string { return "ingress-collector" }

func (ic *IngressCollector) Description() string {
	return "Collects metrics for Kubernetes ingresses"
}

func (ic *IngressCollector) Collect(ctx context.Context) ([]ResourceMetrics, error) {
	var metrics []ResourceMetrics

	ingresses, err := ic.kubeClient.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list ingresses: %w", err)
	}

	for _, ing := range ingresses.Items {
		if ic.isNamespaceExcluded(ing.Namespace) {
			continue
		}

		rules := make([]map[string]interface{}, 0)
		for _, rule := range ing.Spec.Rules {
			paths := make([]map[string]interface{}, 0)
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					paths = append(paths, map[string]interface{}{
						"path":     path.Path,
						"pathType": path.PathType,
						"backend": map[string]interface{}{
							"service": map[string]interface{}{
								"name": path.Backend.Service.Name,
								"port": path.Backend.Service.Port,
							},
						},
					})
				}
			}

			rules = append(rules, map[string]interface{}{
				"host":  rule.Host,
				"paths": paths,
			})
		}

		tls := make([]map[string]interface{}, 0)
		for _, t := range ing.Spec.TLS {
			tls = append(tls, map[string]interface{}{
				"hosts":      t.Hosts,
				"secretName": t.SecretName,
			})
		}

		metric := ResourceMetrics{
			Name:        ing.Name,
			Namespace:   ing.Namespace,
			Kind:        "Ingress",
			Labels:      ing.Labels,
			CollectedAt: time.Now(),
			Status: map[string]interface{}{
				"loadBalancer": ing.Status.LoadBalancer,
				"class":        ing.Spec.IngressClassName,
				"rules":        rules,
				"tls":          tls,
			},
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (ic *IngressCollector) isNamespaceExcluded(namespace string) bool {
	for _, excluded := range ic.config.ExcludeNamespaces {
		if namespace == excluded {
			return true
		}
	}

	if len(ic.config.IncludeNamespaces) == 0 {
		return false
	}

	for _, included := range ic.config.IncludeNamespaces {
		if namespace == included {
			return false
		}
	}

	return true
}
