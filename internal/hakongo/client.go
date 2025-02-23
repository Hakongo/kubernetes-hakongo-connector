package hakongo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hakongo/kubernetes-connector/internal/cluster"
	"github.com/hakongo/kubernetes-connector/internal/collector"
)

const defaultTimeout = 30 * time.Second

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type ClientConfig struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

func NewClient(config ClientConfig) *Client {
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	return &Client{
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

type MetricsData struct {
	ClusterID   string                     `json:"cluster_id"`
	Context     *cluster.ClusterContext    `json:"context"`
	CollectedAt time.Time                  `json:"collected_at"`
	Resources   []collector.ResourceMetrics `json:"resources"`
}

func (c *Client) SendMetrics(ctx context.Context, data *MetricsData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/v1/metrics", c.baseURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
