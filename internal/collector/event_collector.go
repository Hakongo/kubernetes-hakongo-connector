package collector

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// EventCollector collects Kubernetes events
type EventCollector struct {
	kubeClient kubernetes.Interface
	config     CollectorConfig
}

// NewEventCollector creates a new event collector
func NewEventCollector(kubeClient kubernetes.Interface, config CollectorConfig) *EventCollector {
	return &EventCollector{
		kubeClient: kubeClient,
		config:     config,
	}
}

// Name returns the name of the collector
func (c *EventCollector) Name() string {
	return "event-collector"
}

// Description returns a description of the collector
func (c *EventCollector) Description() string {
	return "Collects Kubernetes events"
}

// Collect gathers events from the Kubernetes cluster
func (c *EventCollector) Collect(ctx context.Context) ([]ResourceMetrics, error) {
	// List events from all namespaces
	events, err := c.kubeClient.CoreV1().Events("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	var metrics []ResourceMetrics

	// Process each event
	for _, event := range events.Items {
		// Skip events in excluded namespaces
		if contains(c.config.ExcludeNamespaces, event.Namespace) {
			continue
		}

		// Skip if namespace is not included (when inclusion list is not empty)
		if len(c.config.IncludeNamespaces) > 0 && !contains(c.config.IncludeNamespaces, event.Namespace) {
			continue
		}

		// Calculate event duration
		var durationSeconds int64
		if event.FirstTimestamp.Time.Before(event.LastTimestamp.Time) {
			durationSeconds = int64(event.LastTimestamp.Sub(event.FirstTimestamp.Time).Seconds())
		}

		// Determine event severity based on type
		severity := "info"
		if event.Type == "Warning" {
			severity = "warning"
		}

		// Create event metrics
		eventMetrics := ResourceMetrics{
			Name:        event.Name,
			Namespace:   event.Namespace,
			Kind:        "Event",
			Labels:      event.Labels,
			CollectedAt: time.Now(),
			Status: map[string]interface{}{
				"type":    event.Type,
				"reason":  event.Reason,
				"message": event.Message,
				"count":   event.Count,
				"source": map[string]string{
					"component": event.Source.Component,
					"host":      event.Source.Host,
				},
				"involvedObject": map[string]string{
					"kind":      event.InvolvedObject.Kind,
					"name":      event.InvolvedObject.Name,
					"namespace": event.InvolvedObject.Namespace,
					"uid":       string(event.InvolvedObject.UID),
				},
				"firstTimestamp": event.FirstTimestamp.Time.Format(time.RFC3339),
				"lastTimestamp":  event.LastTimestamp.Time.Format(time.RFC3339),
				"durationSeconds": durationSeconds,
				"severity":        severity,
			},
		}

		metrics = append(metrics, eventMetrics)
	}

	return metrics, nil
}
