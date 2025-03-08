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
