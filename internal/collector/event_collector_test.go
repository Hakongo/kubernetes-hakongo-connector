package collector

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

func TestEventCollector_Collect(t *testing.T) {
	// Create a fake kubernetes client
	client := fake.NewSimpleClientset()

	// Create test events
	firstTimestamp := metav1.NewTime(time.Now().Add(-10 * time.Minute))
	lastTimestamp := metav1.NewTime(time.Now())

	normalEvent := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "normal-event",
			Namespace: "default",
			Labels: map[string]string{
				"app": "test",
			},
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "Pod",
			Name:      "test-pod",
			Namespace: "default",
			UID:       types.UID("test-pod-uid"),
		},
		Type:           "Normal",
		Reason:         "Started",
		Message:        "Started container",
		Count:          1,
		FirstTimestamp: firstTimestamp,
		LastTimestamp:  lastTimestamp,
		Source: corev1.EventSource{
			Component: "kubelet",
			Host:      "node-1",
		},
	}

	warningEvent := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "warning-event",
			Namespace: "kube-system",
			Labels: map[string]string{
				"app": "system",
			},
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "Pod",
			Name:      "system-pod",
			Namespace: "kube-system",
			UID:       types.UID("system-pod-uid"),
		},
		Type:           "Warning",
		Reason:         "Failed",
		Message:        "Failed to start container",
		Count:          3,
		FirstTimestamp: firstTimestamp,
		LastTimestamp:  lastTimestamp,
		Source: corev1.EventSource{
			Component: "kubelet",
			Host:      "node-2",
		},
	}

	// Create the events in the fake client
	_, err := client.CoreV1().Events("default").Create(context.Background(), normalEvent, metav1.CreateOptions{})
	assert.NoError(t, err)

	_, err = client.CoreV1().Events("kube-system").Create(context.Background(), warningEvent, metav1.CreateOptions{})
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name              string
		config            CollectorConfig
		expectedEventCount int
		expectedTypes     []string
	}{
		{
			name: "collect all events",
			config: CollectorConfig{
				IncludeNamespaces: []string{},
				ExcludeNamespaces: []string{},
			},
			expectedEventCount: 2,
			expectedTypes:      []string{"Normal", "Warning"},
		},
		{
			name: "exclude kube-system namespace",
			config: CollectorConfig{
				IncludeNamespaces: []string{},
				ExcludeNamespaces: []string{"kube-system"},
			},
			expectedEventCount: 1,
			expectedTypes:      []string{"Normal"},
		},
		{
			name: "include only default namespace",
			config: CollectorConfig{
				IncludeNamespaces: []string{"default"},
				ExcludeNamespaces: []string{},
			},
			expectedEventCount: 1,
			expectedTypes:      []string{"Normal"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create the event collector with the test config
			collector := NewEventCollector(client, tc.config)

			// Collect events
			metrics, err := collector.Collect(context.Background())
			assert.NoError(t, err)

			// Verify the number of events collected
			assert.Equal(t, tc.expectedEventCount, len(metrics))

			// Verify each event has the expected fields
			if len(metrics) > 0 {
				foundTypes := make([]string, 0, len(metrics))
				
				for _, metric := range metrics {
					// Verify common fields
					assert.Equal(t, "Event", metric.Kind)
					assert.NotEmpty(t, metric.Name)
					assert.NotEmpty(t, metric.Namespace)
					assert.NotNil(t, metric.Status)
					
					// Access the status map directly
					status := metric.Status
					if assert.NotNil(t, status, "Status should not be nil") {
						assert.Contains(t, status, "type")
						assert.Contains(t, status, "reason")
						assert.Contains(t, status, "message")
						assert.Contains(t, status, "count")
						assert.Contains(t, status, "severity")
						assert.Contains(t, status, "firstTimestamp")
						assert.Contains(t, status, "lastTimestamp")
						assert.Contains(t, status, "durationSeconds")
						
						// Add the event type to our found types
						if eventType, ok := status["type"].(string); ok {
							foundTypes = append(foundTypes, eventType)
						}
						
						// Verify severity based on type
						if eventType, ok := status["type"].(string); ok {
							expectedSeverity := "info"
							if eventType == "Warning" {
								expectedSeverity = "warning"
							}
							assert.Equal(t, expectedSeverity, status["severity"].(string))
						}
					}
				}
				
				// Verify we found all the expected event types
				for _, expectedType := range tc.expectedTypes {
					assert.Contains(t, foundTypes, expectedType)
				}
			}
		})
	}
}
