package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// PrometheusClient wraps the Prometheus API client
type PrometheusClient struct {
	api     v1.API
	baseURL string
}

// NewPrometheusClient creates a new Prometheus client
func NewPrometheusClient(baseURL string) (*PrometheusClient, error) {
	client, err := api.NewClient(api.Config{
		Address: baseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	return &PrometheusClient{
		api:     v1.NewAPI(client),
		baseURL: baseURL,
	}, nil
}

// QueryRange performs a range query against Prometheus
func (c *PrometheusClient) QueryRange(ctx context.Context, query string, r v1.Range) (model.Value, error) {
	result, _, err := c.api.QueryRange(ctx, query, r)
	if err != nil {
		return nil, fmt.Errorf("error querying Prometheus: %w", err)
	}
	return result, nil
}

// Query performs an instant query against Prometheus
func (c *PrometheusClient) Query(ctx context.Context, query string, ts time.Time) (model.Value, error) {
	result, _, err := c.api.Query(ctx, query, ts)
	if err != nil {
		return nil, fmt.Errorf("error querying Prometheus: %w", err)
	}
	return result, nil
}

// GetContainerMetrics returns container resource usage metrics
func (c *PrometheusClient) GetContainerMetrics(ctx context.Context, namespace, pod string) (*ContainerMetrics, error) {
	timeNow := time.Now()

	// CPU usage in cores
	cpuQuery := fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{namespace="%s",pod="%s"}[5m])) by (container)`, namespace, pod)
	cpuResult, err := c.Query(ctx, cpuQuery, timeNow)
	if err != nil {
		return nil, fmt.Errorf("error querying CPU metrics: %w", err)
	}

	// Memory usage in bytes
	memQuery := fmt.Sprintf(`sum(container_memory_working_set_bytes{namespace="%s",pod="%s"}) by (container)`, namespace, pod)
	memResult, err := c.Query(ctx, memQuery, timeNow)
	if err != nil {
		return nil, fmt.Errorf("error querying memory metrics: %w", err)
	}

	metrics := &ContainerMetrics{
		CPU:    make(map[string]float64),
		Memory: make(map[string]float64),
	}

	// Parse CPU results
	if vector, ok := cpuResult.(model.Vector); ok {
		for _, sample := range vector {
			container := string(sample.Metric["container"])
			metrics.CPU[container] = float64(sample.Value)
		}
	}

	// Parse memory results
	if vector, ok := memResult.(model.Vector); ok {
		for _, sample := range vector {
			container := string(sample.Metric["container"])
			metrics.Memory[container] = float64(sample.Value)
		}
	}

	return metrics, nil
}

// GetNodeMetrics returns node resource usage metrics
func (c *PrometheusClient) GetNodeMetrics(ctx context.Context, nodeName string) (*NodeMetrics, error) {
	timeNow := time.Now()

	// CPU usage in cores
	cpuQuery := fmt.Sprintf(`sum(rate(node_cpu_seconds_total{mode!="idle",node="%s"}[5m]))`, nodeName)
	cpuResult, err := c.Query(ctx, cpuQuery, timeNow)
	if err != nil {
		return nil, fmt.Errorf("error querying CPU metrics: %w", err)
	}

	// Memory usage in bytes
	memQuery := fmt.Sprintf(`node_memory_MemTotal_bytes{node="%s"} - node_memory_MemAvailable_bytes{node="%s"}`, nodeName, nodeName)
	memResult, err := c.Query(ctx, memQuery, timeNow)
	if err != nil {
		return nil, fmt.Errorf("error querying memory metrics: %w", err)
	}

	metrics := &NodeMetrics{}

	// Parse CPU results
	if vector, ok := cpuResult.(model.Vector); ok && len(vector) > 0 {
		metrics.CPUUsage = float64(vector[0].Value)
	}

	// Parse memory results
	if vector, ok := memResult.(model.Vector); ok && len(vector) > 0 {
		metrics.MemoryUsage = float64(vector[0].Value)
	}

	return metrics, nil
}

// ContainerMetrics represents resource usage metrics for containers
type ContainerMetrics struct {
	CPU    map[string]float64 // CPU usage in cores by container
	Memory map[string]float64 // Memory usage in bytes by container
}

// NodeMetrics represents resource usage metrics for nodes
type NodeMetrics struct {
	CPUUsage    float64 // CPU usage in cores
	MemoryUsage float64 // Memory usage in bytes
}

// GetBaseURL returns the base URL of the Prometheus server
func (c *PrometheusClient) GetBaseURL() string {
	return c.baseURL
}
