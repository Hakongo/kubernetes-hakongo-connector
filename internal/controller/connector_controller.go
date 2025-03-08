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

func (r *ConnectorConfigReconciler) setupCollectors(ctx context.Context, config *hakongov1alpha1.ConnectorConfig, clusterCtx *cluster.ClusterContext) error {
	// Initialize Prometheus client if needed
	if r.prometheusClient == nil && config.Spec.Prometheus != nil {
		var err error
		r.prometheusClient, err = metrics.NewPrometheusClient(config.Spec.Prometheus.URL)
		if err != nil {
			return fmt.Errorf("failed to create Prometheus client: %w", err)
		}
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
		collector.NewPodCollector(r.kubeClient, r.prometheusClient, collectorConfig),
		collector.NewNodeCollector(r.kubeClient, r.metricsClient, collectorConfig),
		collector.NewPVCollector(r.kubeClient, r.metricsClient, collectorConfig),
		collector.NewServiceCollector(r.kubeClient, r.metricsClient, collectorConfig),
		collector.NewNamespaceCollector(r.kubeClient, collectorConfig),
		collector.NewWorkloadCollector(r.kubeClient, collectorConfig),
		collector.NewIngressCollector(r.kubeClient, collectorConfig),
	}

	return nil
}

func (r *ConnectorConfigReconciler) collectMetrics(ctx context.Context, clusterCtx *cluster.ClusterContext) error {
	var allMetrics []collector.ResourceMetrics

	// Collect metrics from all collectors
	for _, c := range r.collectors {
		metrics, err := c.Collect(ctx)
		if err != nil {
			return fmt.Errorf("failed to collect metrics from %s: %w", c.Name(), err)
		}
		allMetrics = append(allMetrics, metrics...)
	}

	// Send metrics to API
	if err := r.apiClient.SendMetrics(ctx, allMetrics); err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
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
	if err := r.Get(ctx, types.NamespacedName{
		Name:      config.Spec.HakonGo.APIKey.Name,
		Namespace: config.Namespace,
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
