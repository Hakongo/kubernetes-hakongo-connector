package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	hakongov1alpha1 "github.com/hakongo/kubernetes-connector/api/v1alpha1"
	"github.com/hakongo/kubernetes-connector/internal/api"
	"github.com/hakongo/kubernetes-connector/internal/cluster"
	"github.com/hakongo/kubernetes-connector/internal/collector"
	"github.com/hakongo/kubernetes-connector/internal/metrics"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ConnectorConfigReconciler reconciles a ConnectorConfig object
type ConnectorConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	kubeClient       kubernetes.Interface
	metricsClient    versioned.Interface
	prometheusClient *metrics.PrometheusClient
	apiClient        *api.Client
	collectors       []collector.Collector
	contextProvider  *cluster.ContextProvider
}

func (r *ConnectorConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the ConnectorConfig instance
	connConfig := &hakongov1alpha1.ConnectorConfig{}
	if err := r.Get(ctx, req.NamespacedName, connConfig); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Reconciling ConnectorConfig", "name", req.Name)

	// Initialize clients if needed
	if err := r.ensureClients(ctx, connConfig); err != nil {
		logger.Error(err, "Failed to initialize clients")
		return ctrl.Result{}, err
	}

	// Initialize cluster context provider
	if r.contextProvider == nil {
		metadata := make(map[string]interface{})
		for k, v := range connConfig.Spec.ClusterContext.Metadata {
			metadata[k] = v
		}
		r.contextProvider = cluster.NewContextProvider(r.kubeClient, &cluster.Config{
			ClusterName:  connConfig.Spec.ClusterContext.Name,
			ProviderName: connConfig.Spec.ClusterContext.Type,
			Region:       connConfig.Spec.ClusterContext.Region,
			Zone:         connConfig.Spec.ClusterContext.Zone,
			Labels:       connConfig.Spec.ClusterContext.Labels,
			Metadata:     metadata,
		})
	}

	// Get cluster context
	clusterCtx, err := r.contextProvider.GetContext(ctx)
	if err != nil {
		logger.Error(err, "Failed to get cluster context")
		return ctrl.Result{}, err
	}

	// Setup collectors with cluster context
	if err := r.setupCollectors(ctx, connConfig, clusterCtx); err != nil {
		logger.Error(err, "Failed to setup collectors")
		return ctrl.Result{}, err
	}

	// Start collecting metrics
	if err := r.collectMetrics(ctx, clusterCtx); err != nil {
		logger.Error(err, "Failed to collect metrics")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

func (r *ConnectorConfigReconciler) setupCollectors(_ context.Context, config *hakongov1alpha1.ConnectorConfig, clusterCtx *cluster.ClusterContext) error {
	// Determine which metrics sources to use based on user configuration
	usePrometheus := false
	useMetricsServer := false

	// Initialize Prometheus client if configured
	if config.Spec.Prometheus != nil {
		usePrometheus = true
		if r.prometheusClient == nil {
			var err error
			r.prometheusClient, err = metrics.NewPrometheusClient(config.Spec.Prometheus.URL)
			if err != nil {
				return fmt.Errorf("failed to create Prometheus client: %w", err)
			}
		}
	}

	// Check if Metrics Server is enabled
	if config.Spec.MetricsServer != nil && config.Spec.MetricsServer.Enabled {
		useMetricsServer = true
	}

	// If neither is configured, default to Metrics Server
	if !usePrometheus && !useMetricsServer {
		useMetricsServer = true
	}

	// Create base collector config
	collectorConfig := collector.CollectorConfig{
		CollectionInterval:       time.Duration(60) * time.Second, // Default to 60s
		IncludeNamespaces:        []string{},                      // Collect from all namespaces
		ExcludeNamespaces:        []string{"kube-system"},         // Exclude system namespaces by default
		IncludeLabels:            make(map[string]string),
		MaxConcurrentCollections: 5,
	}

	// Add cluster context labels
	if clusterCtx != nil {
		collectorConfig.IncludeLabels["cluster_name"] = clusterCtx.Name
		if clusterCtx.Labels != nil {
			for k, v := range clusterCtx.Labels {
				collectorConfig.IncludeLabels[k] = v
			}
		}
	}

	// Override config from spec if provided
	if len(config.Spec.Collectors) > 0 {
		for _, c := range config.Spec.Collectors {
			if c.Interval > 0 {
				collectorConfig.CollectionInterval = time.Duration(c.Interval) * time.Second
			}
			if c.Labels != nil {
				// Merge labels
				for k, v := range c.Labels {
					collectorConfig.IncludeLabels[k] = v
				}
			}
		}
	}

	// Create collectors
	r.collectors = []collector.Collector{
		collector.NewPodCollector(r.kubeClient, r.prometheusClient, collectorConfig, usePrometheus),
		collector.NewNodeCollector(r.kubeClient, r.metricsClient, r.prometheusClient, collectorConfig, usePrometheus, useMetricsServer),
		collector.NewPVCollector(r.kubeClient, r.metricsClient, collectorConfig),
		collector.NewServiceCollector(r.kubeClient, r.metricsClient, collectorConfig),
		collector.NewNamespaceCollector(r.kubeClient, collectorConfig),
		collector.NewWorkloadCollector(r.kubeClient, collectorConfig),
		collector.NewIngressCollector(r.kubeClient, collectorConfig),
		collector.NewEventCollector(r.kubeClient, collectorConfig),
	}

	return nil
}

func (r *ConnectorConfigReconciler) collectMetrics(ctx context.Context, clusterCtx *cluster.ClusterContext) error {
	logger := log.FromContext(ctx)
	var allMetrics []collector.ResourceMetrics

	// Log Prometheus client status with detailed information
	if r.prometheusClient != nil {
		logger.Info("Prometheus client is configured", "url", r.prometheusClient.GetBaseURL())
		
		// Test Prometheus connectivity
		testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		
		// Simple query to check if Prometheus is responding
		testResult, err := r.prometheusClient.Query(testCtx, "up", time.Now())
		if err != nil {
			logger.Error(err, "Failed to connect to Prometheus server", "url", r.prometheusClient.GetBaseURL())
		} else {
			logger.Info("Successfully connected to Prometheus server", "url", r.prometheusClient.GetBaseURL(), "result", testResult.String())
			
			// Get some sample metrics to verify Prometheus data collection
			// Query CPU usage for pods
			cpuQuery := "sum(rate(container_cpu_usage_seconds_total{container!='POD',container!=''}[5m])) by (pod, namespace)"
			cpuResult, cpuErr := r.prometheusClient.Query(testCtx, cpuQuery, time.Now())
			if cpuErr != nil {
				logger.Error(cpuErr, "Failed to query CPU metrics from Prometheus", "query", cpuQuery)
			} else {
				logger.Info("Prometheus CPU metrics sample", "query", cpuQuery, "result_type", cpuResult.Type().String(), "sample_count", len(cpuResult.String()) > 0)
			}
			
			// Query memory usage for pods
			memQuery := "sum(container_memory_working_set_bytes{container!='POD',container!=''}) by (pod, namespace)"
			memResult, memErr := r.prometheusClient.Query(testCtx, memQuery, time.Now())
			if memErr != nil {
				logger.Error(memErr, "Failed to query memory metrics from Prometheus", "query", memQuery)
			} else {
				logger.Info("Prometheus memory metrics sample", "query", memQuery, "result_type", memResult.Type().String(), "sample_count", len(memResult.String()) > 0)
			}
		}
	} else {
		logger.Info("Prometheus client is not configured, metrics will be limited")
	}

	// Separate metrics for events and other resources
	var eventMetrics []collector.ResourceMetrics
	var regularMetrics []collector.ResourceMetrics

	// Collect metrics from all collectors with detailed logging
	for _, c := range r.collectors {
		logger.Info("Starting metrics collection", "collector", c.Name(), "description", c.Description())
		
		// Measure collection time
		startTime := time.Now()
		metrics, err := c.Collect(ctx)
		collectionDuration := time.Since(startTime)
		
		if err != nil {
			logger.Error(err, "Failed to collect metrics", "collector", c.Name(), "duration_ms", collectionDuration.Milliseconds())
			return fmt.Errorf("failed to collect metrics from %s: %w", c.Name(), err)
		}
		
		logger.Info("Successfully collected metrics", 
			"collector", c.Name(), 
			"count", len(metrics), 
			"duration_ms", collectionDuration.Milliseconds())
		
		// Log a sample of metrics for debugging
		if len(metrics) > 0 {
			sampleSize := 2 // Increase sample size for better visibility
			if len(metrics) < sampleSize {
				sampleSize = len(metrics)
			}
			
			logger.Info("Metrics sample", "collector", c.Name(), "sample_size", sampleSize, "total", len(metrics))
			
			for i := 0; i < sampleSize; i++ {
				// Log basic resource info
				logger.Info("Resource metrics", 
					"collector", c.Name(),
					"resource", metrics[i].Kind+"/"+metrics[i].Name,
					"namespace", metrics[i].Namespace,
					"labels", fmt.Sprintf("%v", metrics[i].Labels),
					"collected_at", metrics[i].CollectedAt.Format(time.RFC3339))
				
				// Log resource usage metrics if not an event
				if metrics[i].Kind != "Event" {
					logger.Info("Resource usage", 
						"resource", metrics[i].Kind+"/"+metrics[i].Name,
						"cpu_usage_cores", float64(metrics[i].CPU.UsageNanoCores) / 1e9,
						"cpu_usage_percent", metrics[i].CPU.UsageCorePercent,
						"memory_usage_mb", float64(metrics[i].Memory.UsageBytes) / (1024 * 1024),
						"memory_request_mb", float64(metrics[i].Memory.RequestBytes) / (1024 * 1024),
						"memory_limit_mb", float64(metrics[i].Memory.LimitBytes) / (1024 * 1024))
				} else {
					// Log event-specific information
					if metrics[i].Status != nil {
						logger.Info("Event details",
							"resource", metrics[i].Kind+"/"+metrics[i].Name,
							"type", metrics[i].Status["type"],
							"reason", metrics[i].Status["reason"],
							"count", metrics[i].Status["count"],
							"severity", metrics[i].Status["severity"])
					}
				}
				
				// Log container metrics if available
				if len(metrics[i].Containers) > 0 {
					containerSampleSize := 2
					if len(metrics[i].Containers) < containerSampleSize {
						containerSampleSize = len(metrics[i].Containers)
					}
					
					logger.Info("Container metrics", 
						"resource", metrics[i].Kind+"/"+metrics[i].Name, 
						"container_count", len(metrics[i].Containers),
						"sample_size", containerSampleSize)
					
					for j := 0; j < containerSampleSize; j++ {
						logger.Info("Container details", 
							"resource", metrics[i].Kind+"/"+metrics[i].Name,
							"container", metrics[i].Containers[j].Name,
							"cpu_usage_cores", float64(metrics[i].Containers[j].CPU.UsageNanoCores) / 1e9,
							"memory_usage_mb", float64(metrics[i].Containers[j].Memory.UsageBytes) / (1024 * 1024),
							"ready", metrics[i].Containers[j].Ready,
							"restarts", metrics[i].Containers[j].Restarts,
							"state", metrics[i].Containers[j].State)
					}
				}
			}
		}
		
		// Separate event metrics from other resource metrics
		for _, metric := range metrics {
			if metric.Kind == "Event" {
				eventMetrics = append(eventMetrics, metric)
			} else {
				regularMetrics = append(regularMetrics, metric)
			}
		}
		
		// Add all metrics to the combined list for backward compatibility
		allMetrics = append(allMetrics, metrics...)
	}

	// Log the total number of metrics collected with summary
	logger.Info("Total metrics collected", 
		"count", len(allMetrics), 
		"regular_metrics", len(regularMetrics),
		"event_metrics", len(eventMetrics),
		"collector_count", len(r.collectors),
		"cluster_name", clusterCtx.Name,
		"timestamp", time.Now().Format(time.RFC3339))

	// Send regular metrics to API with detailed logging
	if len(regularMetrics) > 0 {
		logger.Info("Sending regular metrics to HakonGo API", 
			"count", len(regularMetrics), 
			"api_url", r.apiClient.GetBaseURL())
		
		startTime := time.Now()
		if err := r.apiClient.SendMetrics(ctx, regularMetrics); err != nil {
			// Log the error but don't fail the reconciliation
			logger.Error(err, "Failed to send regular metrics to HakonGo API", 
				"duration_ms", time.Since(startTime).Milliseconds())
			// Continue processing even if sending metrics fails
		} else {
			logger.Info("Successfully sent regular metrics to HakonGo API", 
				"count", len(regularMetrics), 
				"duration_ms", time.Since(startTime).Milliseconds())
		}
	}

	// Send event metrics to API with detailed logging
	if len(eventMetrics) > 0 {
		logger.Info("Sending event metrics to HakonGo API", 
			"count", len(eventMetrics), 
			"api_url", r.apiClient.GetBaseURL())
		
		// Create cluster context map for event metrics
		clusterContextMap := map[string]interface{}{
			"name":     clusterCtx.Name,
			"provider": clusterCtx.Provider.Name,
			"region":   clusterCtx.Provider.Region,
			"zone":     clusterCtx.Provider.Zone,
			"labels":   clusterCtx.Labels,
		}
		
		startTime := time.Now()
		if err := r.apiClient.SendEventMetrics(ctx, clusterCtx.Name, clusterContextMap, eventMetrics); err != nil {
			// Log the error but don't fail the reconciliation
			logger.Error(err, "Failed to send event metrics to HakonGo API", 
				"duration_ms", time.Since(startTime).Milliseconds())
			// Continue processing even if sending metrics fails
		} else {
			logger.Info("Successfully sent event metrics to HakonGo API", 
				"count", len(eventMetrics), 
				"duration_ms", time.Since(startTime).Milliseconds())
		}
	}

	return nil
}

func (r *ConnectorConfigReconciler) ensureClients(ctx context.Context, config *hakongov1alpha1.ConnectorConfig) error {
	// Initialize Kubernetes clients if needed
	if r.kubeClient == nil {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			// Fall back to kubeconfig
			kubeconfigPath := os.Getenv("KUBECONFIG")
			if kubeconfigPath == "" {
				home, _ := os.UserHomeDir()
				kubeconfigPath = filepath.Join(home, ".kube", "config")
			}
			cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
			if err != nil {
				return fmt.Errorf("failed to get kubeconfig: %w", err)
			}
		}

		r.kubeClient, err = kubernetes.NewForConfig(cfg)
		if err != nil {
			return fmt.Errorf("failed to create kubernetes client: %w", err)
		}

		r.metricsClient, err = versioned.NewForConfig(cfg)
		if err != nil {
			return fmt.Errorf("failed to create metrics client: %w", err)
		}
	}

	// Get API key from secret
	var apiKeySecret corev1.Secret
	namespace := config.Namespace
	if namespace == "" {
		namespace = "default" // Use default namespace if not specified
	}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      config.Spec.HakonGo.APIKey.Name,
		Namespace: namespace,
	}, &apiKeySecret); err != nil {
		return fmt.Errorf("failed to get API key secret: %w", err)
	}

	apiKey := string(apiKeySecret.Data[config.Spec.HakonGo.APIKey.Key])
	if apiKey == "" {
		return fmt.Errorf("API key not found in secret %s at key %s",
			config.Spec.HakonGo.APIKey.Name, config.Spec.HakonGo.APIKey.Key)
	}

	// Initialize API client if needed
	if r.apiClient == nil {
		r.apiClient = api.NewClient(api.ClientConfig{
			BaseURL: config.Spec.HakonGo.BaseURL,
			APIKey:  apiKey,
			Timeout: 30 * time.Second,
		})
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConnectorConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hakongov1alpha1.ConnectorConfig{}).
		Complete(r)
}
