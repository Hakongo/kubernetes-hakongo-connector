package hakongo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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
	ClusterID   string      `json:"clusterId"`
	CollectedAt time.Time   `json:"collectedAt"`
	Resources   interface{} `json:"resources"`
}

func (c *Client) SendMetrics(ctx context.Context, data MetricsData) error {
	url := fmt.Sprintf("%s/api/v1/metrics", c.baseURL)
	
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
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

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to send metrics: status code %d", resp.StatusCode)
	}

	return nil
}
