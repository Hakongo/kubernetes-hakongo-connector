package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hakongo/kubernetes-connector/internal/collector"
)

// Client represents the SaaS API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// ClientConfig contains configuration for the API client
type ClientConfig struct {
	BaseURL            string
	APIKey             string
	Timeout            time.Duration
	MaxRetries         int
	RetryWaitDuration  time.Duration
	CompressionEnabled bool
}

// NewClient creates a new SaaS API client
func NewClient(config ClientConfig) *Client {
	return &Client{
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// SendMetrics sends collected metrics to the SaaS platform
func (c *Client) SendMetrics(ctx context.Context, metrics []collector.ResourceMetrics) error {
	payload, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/v1/metrics", c.baseURL),
		bytes.NewReader(payload),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// EventMetricsPayload represents the payload for sending event metrics
type EventMetricsPayload struct {
	ClusterID   string                 `json:"cluster_id"`
	Context     map[string]interface{} `json:"context"`
	CollectedAt time.Time              `json:"collected_at"`
	Resources   []EventResource        `json:"resources"`
}

// EventResource represents a Kubernetes event resource
type EventResource struct {
	Namespace      string                 `json:"namespace"`
	Name           string                 `json:"name"`
	UID            string                 `json:"uid"`
	Type           string                 `json:"type"`
	Reason         string                 `json:"reason"`
	Message        string                 `json:"message"`
	Source         EventSource            `json:"source"`
	InvolvedObject EventInvolvedObject    `json:"involved_object"`
	Metadata       EventMetadata          `json:"metadata"`
	Metrics        EventMetrics           `json:"metrics"`
}

// EventSource represents the source of a Kubernetes event
type EventSource struct {
	Component string `json:"component"`
	Host      string `json:"host"`
}

// EventInvolvedObject represents the object involved in a Kubernetes event
type EventInvolvedObject struct {
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	UID       string `json:"uid"`
}

// EventMetadata represents metadata for a Kubernetes event
type EventMetadata struct {
	Labels            map[string]string `json:"labels"`
	Annotations       map[string]string `json:"annotations"`
	CreationTimestamp string             `json:"creation_timestamp"`
}

// EventMetrics represents metrics for a Kubernetes event
type EventMetrics struct {
	Count           int    `json:"count"`
	FirstTimestamp  string `json:"first_timestamp"`
	LastTimestamp   string `json:"last_timestamp"`
	DurationSeconds int64  `json:"duration_seconds"`
	Severity        string `json:"severity"`
}

// SendEventMetrics sends collected event metrics to the SaaS platform
func (c *Client) SendEventMetrics(ctx context.Context, clusterID string, clusterContext map[string]interface{}, eventMetrics []collector.ResourceMetrics) error {
	// Convert ResourceMetrics to EventResource format
	resources := make([]EventResource, 0, len(eventMetrics))
	
	for _, metric := range eventMetrics {
		// Skip non-event resources
		if metric.Kind != "Event" {
			continue
		}
		
		// Check if status is nil
		if metric.Status == nil {
			continue
		}
		
		// Use the status map directly
		status := metric.Status
		
		// Extract source information
		sourceMap, _ := status["source"].(map[string]string)
		source := EventSource{
			Component: sourceMap["component"],
			Host:      sourceMap["host"],
		}
		
		// Extract involved object information
		involvedMap, _ := status["involvedObject"].(map[string]string)
		involved := EventInvolvedObject{
			Kind:      involvedMap["kind"],
			Namespace: involvedMap["namespace"],
			Name:      involvedMap["name"],
			UID:       involvedMap["uid"],
		}
		
		// Extract event metrics
		eventMetrics := EventMetrics{
			Count:           getIntFromMap(status, "count"),
			FirstTimestamp:  getStringFromMap(status, "firstTimestamp"),
			LastTimestamp:   getStringFromMap(status, "lastTimestamp"),
			DurationSeconds: getInt64FromMap(status, "durationSeconds"),
			Severity:        getStringFromMap(status, "severity"),
		}
		
		// Create event resource
		eventResource := EventResource{
			Namespace:      metric.Namespace,
			Name:           metric.Name,
			UID:            "", // UID not available in ResourceMetrics
			Type:           getStringFromMap(status, "type"),
			Reason:         getStringFromMap(status, "reason"),
			Message:        getStringFromMap(status, "message"),
			Source:         source,
			InvolvedObject: involved,
			Metadata: EventMetadata{
				Labels:            metric.Labels,
				Annotations:       make(map[string]string),
				CreationTimestamp: metric.CollectedAt.Format(time.RFC3339),
			},
			Metrics: eventMetrics,
		}
		
		resources = append(resources, eventResource)
	}
	
	// Create payload
	payload := EventMetricsPayload{
		ClusterID:   clusterID,
		Context:     clusterContext,
		CollectedAt: time.Now().UTC(),
		Resources:   resources,
	}
	
	// Marshal payload
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event metrics: %w", err)
	}
	
	// Create request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/v1/metrics/events", c.baseURL),
		bytes.NewReader(data),
	)
	if err != nil {
		return fmt.Errorf("failed to create event metrics request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	
	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send event metrics: %w", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code for event metrics: %d", resp.StatusCode)
	}
	
	return nil
}

// Helper functions for extracting values from maps
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getIntFromMap(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int32:
			return int(v)
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

func getInt64FromMap(m map[string]interface{}, key string) int64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return int64(v)
		case int32:
			return int64(v)
		case int64:
			return v
		case float64:
			return int64(v)
		}
	}
	return 0
}

// GetClusterConfig retrieves cluster-specific configuration from the SaaS platform
func (c *Client) GetClusterConfig(ctx context.Context, clusterID string) (*ClusterConfig, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/v1/clusters/%s/config", c.baseURL, clusterID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var config ClusterConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &config, nil
}

// GetBaseURL returns the base URL of the API
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// ClusterConfig represents cluster-specific configuration from the SaaS platform
type ClusterConfig struct {
	CollectionInterval   time.Duration        `json:"collectionInterval"`
	IncludeNamespaces    []string             `json:"includeNamespaces"`
	ExcludeNamespaces    []string             `json:"excludeNamespaces"`
	ResourceTypes        []string             `json:"resourceTypes"`
	CostingConfiguration CostingConfiguration `json:"costingConfiguration"`
	AlertingRules        []AlertingRule       `json:"alertingRules"`
	CustomMetrics        []CustomMetricConfig `json:"customMetrics"`
}

// CostingConfiguration contains settings for cost calculation
type CostingConfiguration struct {
	Currency          string             `json:"currency"`
	CPUCostPerCore    float64            `json:"cpuCostPerCore"`
	MemoryCostPerGB   float64            `json:"memoryCostPerGB"`
	StorageCostPerGB  float64            `json:"storageCostPerGB"`
	NetworkCostPerGB  float64            `json:"networkCostPerGB"`
	CustomCostFactors map[string]float64 `json:"customCostFactors"`
}

// AlertingRule defines when and how to generate alerts
type AlertingRule struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Metric      string            `json:"metric"`
	Threshold   float64           `json:"threshold"`
	Operator    string            `json:"operator"`
	Duration    time.Duration     `json:"duration"`
	Labels      map[string]string `json:"labels"`
	Severity    string            `json:"severity"`
}

// CustomMetricConfig defines configuration for custom metrics collection
type CustomMetricConfig struct {
	Name       string            `json:"name"`
	Query      string            `json:"query"`
	Labels     map[string]string `json:"labels"`
	Interval   time.Duration     `json:"interval"`
	Aggregator string            `json:"aggregator"`
}
