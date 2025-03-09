package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hakongo/kubernetes-connector/internal/collector"
	"github.com/stretchr/testify/assert"
)

func TestClient_SendEventMetrics(t *testing.T) {
	// Create test event metrics
	eventMetrics := []collector.ResourceMetrics{
		{
			Name:      "test-event",
			Namespace: "default",
			Kind:      "Event",
			Labels: map[string]string{
				"app": "test",
			},
			CollectedAt: time.Now().UTC(),
			Status: map[string]interface{}{
				"type":    "Normal",
				"reason":  "Started",
				"message": "Started container",
				"count":   1,
				"source": map[string]string{
					"component": "kubelet",
					"host":      "node-1",
				},
				"involvedObject": map[string]string{
					"kind":      "Pod",
					"name":      "test-pod",
					"namespace": "default",
					"uid":       "test-pod-uid",
				},
				"firstTimestamp":  time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
				"lastTimestamp":   time.Now().Format(time.RFC3339),
				"durationSeconds": int64(600),
				"severity":        "info",
			},
		},
		{
			Name:      "warning-event",
			Namespace: "default",
			Kind:      "Event",
			Labels: map[string]string{
				"app": "test",
			},
			CollectedAt: time.Now().UTC(),
			Status: map[string]interface{}{
				"type":    "Warning",
				"reason":  "Failed",
				"message": "Failed to start container",
				"count":   3,
				"source": map[string]string{
					"component": "kubelet",
					"host":      "node-2",
				},
				"involvedObject": map[string]string{
					"kind":      "Pod",
					"name":      "test-pod-2",
					"namespace": "default",
					"uid":       "test-pod-2-uid",
				},
				"firstTimestamp":  time.Now().Add(-20 * time.Minute).Format(time.RFC3339),
				"lastTimestamp":   time.Now().Format(time.RFC3339),
				"durationSeconds": int64(1200),
				"severity":        "warning",
			},
		},
		// Add a non-event metric to ensure it's filtered out
		{
			Name:      "test-pod",
			Namespace: "default",
			Kind:      "Pod",
			Labels: map[string]string{
				"app": "test",
			},
			CollectedAt: time.Now().UTC(),
			Status: map[string]interface{}{
				"phase": "Running",
			},
		},
	}

	// Create a test server
	var receivedPayload EventMetricsPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request path
		assert.Equal(t, "/v1/metrics/events", r.URL.Path)
		
		// Verify request method
		assert.Equal(t, http.MethodPost, r.Method)
		
		// Verify content type
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		
		// Verify API key
		assert.Equal(t, "test-api-key", r.Header.Get("X-API-Key"))
		
		// Decode the payload
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&receivedPayload)
		assert.NoError(t, err)
		
		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	// Create the API client
	client := NewClient(ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	})

	// Create cluster context
	clusterID := "test-cluster"
	clusterContext := map[string]interface{}{
		"name":     "test-cluster",
		"provider": "test-provider",
		"region":   "test-region",
		"zone":     "test-zone",
		"labels": map[string]string{
			"environment": "test",
		},
	}

	// Send event metrics
	err := client.SendEventMetrics(context.Background(), clusterID, clusterContext, eventMetrics)
	assert.NoError(t, err)

	// Verify the payload
	assert.Equal(t, clusterID, receivedPayload.ClusterID)
	
	// Check context fields individually since JSON unmarshaling may change map[string]string to map[string]interface{}
	assert.Equal(t, clusterContext["name"], receivedPayload.Context["name"])
	assert.Equal(t, clusterContext["provider"], receivedPayload.Context["provider"])
	assert.Equal(t, clusterContext["region"], receivedPayload.Context["region"])
	assert.Equal(t, clusterContext["zone"], receivedPayload.Context["zone"])
	
	// Check that the labels exist and contain the expected environment value
	contextLabels, ok := receivedPayload.Context["labels"].(map[string]interface{})
	assert.True(t, ok, "Context labels should be a map")
	assert.Equal(t, "test", contextLabels["environment"])
	
	assert.NotZero(t, receivedPayload.CollectedAt)
	
	// Should only include event metrics (not the Pod)
	assert.Equal(t, 2, len(receivedPayload.Resources))
	
	// Verify the first event
	assert.Equal(t, "default", receivedPayload.Resources[0].Namespace)
	assert.Equal(t, "test-event", receivedPayload.Resources[0].Name)
	assert.Equal(t, "Normal", receivedPayload.Resources[0].Type)
	assert.Equal(t, "Started", receivedPayload.Resources[0].Reason)
	assert.Equal(t, "Started container", receivedPayload.Resources[0].Message)
	
	// Verify source
	assert.Equal(t, "kubelet", receivedPayload.Resources[0].Source.Component)
	assert.Equal(t, "node-1", receivedPayload.Resources[0].Source.Host)
	
	// Verify involved object
	assert.Equal(t, "Pod", receivedPayload.Resources[0].InvolvedObject.Kind)
	assert.Equal(t, "test-pod", receivedPayload.Resources[0].InvolvedObject.Name)
	assert.Equal(t, "default", receivedPayload.Resources[0].InvolvedObject.Namespace)
	
	// Verify metrics
	assert.Equal(t, 1, receivedPayload.Resources[0].Metrics.Count)
	assert.Equal(t, int64(600), receivedPayload.Resources[0].Metrics.DurationSeconds)
	assert.Equal(t, "info", receivedPayload.Resources[0].Metrics.Severity)
	
	// Verify the second event (warning)
	assert.Equal(t, "Warning", receivedPayload.Resources[1].Type)
	assert.Equal(t, "Failed", receivedPayload.Resources[1].Reason)
	assert.Equal(t, "warning", receivedPayload.Resources[1].Metrics.Severity)
	assert.Equal(t, 3, receivedPayload.Resources[1].Metrics.Count)
}

func TestClient_SendEventMetrics_Error(t *testing.T) {
	// Create test event metrics
	eventMetrics := []collector.ResourceMetrics{
		{
			Name:      "test-event",
			Namespace: "default",
			Kind:      "Event",
			Labels: map[string]string{
				"app": "test",
			},
			CollectedAt: time.Now().UTC(),
			Status: map[string]interface{}{
				"type":    "Normal",
				"reason":  "Started",
				"message": "Started container",
			},
		},
	}

	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":"error","message":"Internal server error"}`))
	}))
	defer server.Close()

	// Create the API client
	client := NewClient(ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	})

	// Send event metrics
	err := client.SendEventMetrics(context.Background(), "test-cluster", map[string]interface{}{}, eventMetrics)
	
	// Verify that an error was returned
	assert.Error(t, err)
}
