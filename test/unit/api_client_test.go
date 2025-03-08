package unit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hakongo/kubernetes-connector/internal/api"
	"github.com/hakongo/kubernetes-connector/internal/collector"
	"github.com/stretchr/testify/assert"
)

func TestSendMetrics(t *testing.T) {
	// Create a sample set of metrics
	metrics := []collector.ResourceMetrics{
		{
			Name:      "test-pod",
			Namespace: "default",
			Kind:      "Pod",
			Labels: map[string]string{
				"app": "test",
			},
			CollectedAt: time.Now().UTC(),
			CPU: collector.CPUMetrics{
				UsageNanoCores:    int64(100000000),
				RequestMilliCores: int64(200),
				LimitMilliCores:   int64(500),
				UsageCorePercent:  2.5,
				ThrottlingSeconds: 0.0,
			},
			Memory: collector.MemoryMetrics{
				UsageBytes:      int64(104857600), // 100 MiB
				RequestBytes:    int64(209715200), // 200 MiB
				LimitBytes:      int64(524288000), // 500 MiB
				RSSBytes:        int64(104857600),
				PageFaults:      0,
				MajorPageFaults: 0,
			},
			Storage: collector.StorageMetrics{
				UsageBytes:    int64(1073741824), // 1 GiB
				CapacityBytes: int64(10737418240), // 10 GiB
				Available:     int64(9663676416),
				DiskPressure:  false,
			},
			Network: collector.NetworkMetrics{
				RxBytes: int64(1048576), // 1 MiB
				TxBytes: int64(524288),  // 0.5 MiB
			},
			Cost: collector.CostMetrics{
				CPUCost:     0.05,
				MemoryCost:  0.02,
				StorageCost: 0.01,
				NetworkCost: 0.001,
				TotalCost:   0.081,
				Currency:    "USD",
			},
			Status: map[string]interface{}{
				"phase": "Running",
			},
		},
		{
			Name:      "test-node",
			Kind:      "Node",
			Labels: map[string]string{
				"kubernetes.io/hostname": "test-node",
			},
			CollectedAt: time.Now().UTC(),
			CPU: collector.CPUMetrics{
				UsageNanoCores:    int64(2000000000),
				UsageCorePercent:  50.0,
				RequestMilliCores: int64(0),
				LimitMilliCores:   int64(4000),
				ThrottlingSeconds: 0.0,
			},
			Memory: collector.MemoryMetrics{
				UsageBytes:      int64(4294967296), // 4 GiB
				RequestBytes:    int64(0),
				LimitBytes:      int64(8589934592), // 8 GiB
				RSSBytes:        int64(4294967296),
				PageFaults:      0,
				MajorPageFaults: 0,
			},
			Cost: collector.CostMetrics{
				CPUCost:     1.2,
				MemoryCost:  0.4,
				TotalCost:   1.6,
				Currency:    "USD",
			},
			Status: map[string]interface{}{
				"ready": true,
			},
		},
		{
			Name:      "test-service",
			Namespace: "default",
			Kind:      "Service",
			CollectedAt: time.Now().UTC(),
			Network: collector.NetworkMetrics{
				RxBytes: int64(2097152), // 2 MiB
				TxBytes: int64(1048576), // 1 MiB
			},
			Cost: collector.CostMetrics{
				NetworkCost: 0.003,
				TotalCost:   0.003,
				Currency:    "USD",
			},
			Status: map[string]interface{}{
				"type": "ClusterIP",
				"ports": []map[string]interface{}{
					{
						"name":     "http",
						"port":     80,
						"protocol": "TCP",
					},
				},
			},
		},
		{
			Name:      "test-ingress",
			Namespace: "default",
			Kind:      "Ingress",
			CollectedAt: time.Now().UTC(),
			Status: map[string]interface{}{
				"rules": []map[string]interface{}{
					{
						"host": "test.example.com",
						"paths": []map[string]interface{}{
							{
								"path":     "/",
								"pathType": "Prefix",
								"service":  "test-service",
								"port":     80,
							},
						},
					},
				},
				"tls": []map[string]interface{}{
					{
						"hosts":      []string{"test.example.com"},
						"secretName": "test-tls",
					},
				},
			},
		},
		{
			Name:      "test-namespace",
			Kind:      "Namespace",
			CollectedAt: time.Now().UTC(),
			Status: map[string]interface{}{
				"phase": "Active",
				"age":   "2d",
			},
		},
		{
			Name:      "test-pv",
			Kind:      "PersistentVolume",
			CollectedAt: time.Now().UTC(),
			Storage: collector.StorageMetrics{
				UsageBytes:    int64(5368709120), // 5 GiB
				CapacityBytes: int64(10737418240), // 10 GiB
				Available:     int64(5368709120),
				DiskPressure:  false,
			},
			Cost: collector.CostMetrics{
				StorageCost: 0.5,
				TotalCost:   0.5,
				Currency:    "USD",
			},
			Status: map[string]interface{}{
				"storageClass": "standard",
				"phase":        "Bound",
				"claim":        "default/test-pvc",
			},
		},
		{
			Name:      "test-deployment",
			Namespace: "default",
			Kind:      "Deployment",
			CollectedAt: time.Now().UTC(),
			Status: map[string]interface{}{
				"replicas":        3,
				"readyReplicas":   3,
				"updatedReplicas": 3,
			},
		},
	}

	// Create a test server
	var receivedMetrics []collector.ResourceMetrics
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/metrics" {
			// Decode the request body
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&receivedMetrics)
			if err != nil {
				t.Errorf("Failed to decode request body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Check if the API key is present
			apiKey := r.Header.Get("X-API-Key")
			if apiKey != "test-api-key" {
				t.Errorf("Expected API key 'test-api-key', got '%s'", apiKey)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Return success response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := map[string]interface{}{
				"status":  "ok",
				"message": "Metrics received successfully",
			}
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create the API client
	client := api.NewClient(api.ClientConfig{
		BaseURL:   server.URL,
		APIKey:    "test-api-key",
		Timeout:   5 * time.Second,
		MaxRetries: 3,
	})

	// Send the metrics
	err := client.SendMetrics(context.Background(), metrics)
	assert.NoError(t, err, "SendMetrics should not return an error")

	// Verify the received metrics
	assert.Equal(t, len(metrics), len(receivedMetrics), "Number of metrics should match")

	// Check each metric type is present
	resourceTypes := make(map[string]bool)
	for _, metric := range receivedMetrics {
		resourceTypes[metric.Kind] = true
	}

	assert.True(t, resourceTypes["Pod"], "Should have Pod metrics")
	assert.True(t, resourceTypes["Node"], "Should have Node metrics")
	assert.True(t, resourceTypes["Service"], "Should have Service metrics")
	assert.True(t, resourceTypes["Ingress"], "Should have Ingress metrics")
	assert.True(t, resourceTypes["Namespace"], "Should have Namespace metrics")
	assert.True(t, resourceTypes["PersistentVolume"], "Should have PersistentVolume metrics")
	// WorkloadDeployment is not part of the current implementation
	// assert.True(t, resourceTypes["WorkloadDeployment"], "Should have WorkloadDeployment metrics")

	// Verify specific metric values for a Pod
	for _, metric := range receivedMetrics {
		if metric.Kind == "Pod" && metric.Name == "test-pod" {
			assert.Equal(t, "default", metric.Namespace, "Namespace should match")
			// ClusterName is no longer in the ResourceMetrics struct
			
			assert.NotNil(t, metric.CPU, "CPU metrics should not be nil")
			assert.Equal(t, int64(100000000), metric.CPU.UsageNanoCores, "CPU usage should match")
			assert.Equal(t, int64(200), metric.CPU.RequestMilliCores, "CPU request should match")
			assert.Equal(t, int64(500), metric.CPU.LimitMilliCores, "CPU limit should match")
			
			assert.NotNil(t, metric.Memory, "Memory metrics should not be nil")
			assert.Equal(t, int64(104857600), metric.Memory.UsageBytes, "Memory usage should match")
			
			assert.NotNil(t, metric.Storage, "Storage metrics should not be nil")
			assert.Equal(t, int64(1073741824), metric.Storage.UsageBytes, "Storage usage should match")
			
			assert.NotNil(t, metric.Network, "Network metrics should not be nil")
			assert.Equal(t, int64(1048576), metric.Network.RxBytes, "Network RX should match")
			assert.Equal(t, int64(524288), metric.Network.TxBytes, "Network TX should match")
		}
	}
}

func TestGetClusterConfig(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/clusters/test-cluster/config" {
			// Check if the API key is present
			apiKey := r.Header.Get("X-API-Key")
			if apiKey != "test-api-key" {
				t.Errorf("Expected API key 'test-api-key', got '%s'", apiKey)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Return a mock cluster config
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			config := api.ClusterConfig{
				CollectionInterval: 60 * time.Second,
				IncludeNamespaces:  []string{},
				ExcludeNamespaces:  []string{"kube-system"},
				ResourceTypes:      []string{"Pod", "Node", "Service", "Ingress", "Namespace", "PersistentVolume", "Workload"},
				CostingConfiguration: api.CostingConfiguration{
					Currency:         "USD",
					CPUCostPerCore:   30.0,
					MemoryCostPerGB:  5.0,
					StorageCostPerGB: 0.1,
					NetworkCostPerGB: 0.01,
				},
			}
			json.NewEncoder(w).Encode(config)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create the API client
	client := api.NewClient(api.ClientConfig{
		BaseURL:   server.URL,
		APIKey:    "test-api-key",
		Timeout:   5 * time.Second,
		MaxRetries: 3,
	})

	// Get the cluster config
	config, err := client.GetClusterConfig(context.Background(), "test-cluster")
	assert.NoError(t, err, "GetClusterConfig should not return an error")
	assert.NotNil(t, config, "Config should not be nil")

	// Verify the config values
	assert.Equal(t, 60*time.Second, config.CollectionInterval, "CollectionInterval should match")
	assert.Equal(t, []string{"kube-system"}, config.ExcludeNamespaces, "ExcludeNamespaces should match")
	assert.Equal(t, "USD", config.CostingConfiguration.Currency, "Currency should match")
	assert.Equal(t, 30.0, config.CostingConfiguration.CPUCostPerCore, "CPUCostPerCore should match")
	assert.Equal(t, 5.0, config.CostingConfiguration.MemoryCostPerGB, "MemoryCostPerGB should match")
	assert.Equal(t, 0.1, config.CostingConfiguration.StorageCostPerGB, "StorageCostPerGB should match")
	assert.Equal(t, 0.01, config.CostingConfiguration.NetworkCostPerGB, "NetworkCostPerGB should match")
}
