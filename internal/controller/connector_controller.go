package controller

import (
	"context"
	"fmt"
	"time"

	hakongov1alpha1 "github.com/hakongo/kubernetes-connector/api/v1alpha1"
	"github.com/hakongo/kubernetes-connector/internal/collector"
	"github.com/hakongo/kubernetes-connector/internal/hakongo"
	"github.com/hakongo/kubernetes-connector/internal/cluster"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	defaultRequeueAfter = time.Second * 300 // 5 minutes
)

// ConnectorConfigReconciler reconciles a ConnectorConfig object
// +kubebuilder:rbac:groups=hakongo.io,resources=connectorconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hakongo.io,resources=connectorconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hakongo.io,resources=connectorconfigs/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=persistentvolumes,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=endpoints,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="metrics.k8s.io",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="metrics.k8s.io",resources=nodes,verbs=get;list;watch
type ConnectorConfigReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	Config       *rest.Config
	kubeClient    kubernetes.Interface
	metricsClient versioned.Interface
	hakongoClient *hakongo.Client
	collectors    map[string]collector.Collector
	contextProvider *cluster.ContextProvider
}

// +kubebuilder:rbac:groups=hakongo.io,resources=connectorconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hakongo.io,resources=connectorconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hakongo.io,resources=connectorconfigs/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods;nodes;persistentvolumeclaims;services,verbs=get;list;watch
// +kubebuilder:rbac:groups=metrics.k8s.io,resources=pods;nodes,verbs=get;list

func (r *ConnectorConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var config hakongov1alpha1.ConnectorConfig
	if err := r.Get(ctx, req.NamespacedName, &config); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Initialize HakonGo client if needed
	if r.hakongoClient == nil {
		if err := r.initHakonGoClient(ctx, &config); err != nil {
			log.Error(err, "failed to initialize HakonGo client")
			r.updateStatus(ctx, &config, err.Error())
			return ctrl.Result{}, err
		}
	}

	// Initialize cluster context provider if needed
	if r.contextProvider == nil {
		r.contextProvider = cluster.NewContextProvider(r.kubeClient, &cluster.Config{
			ClusterName:  config.Spec.ClusterContext.Name,
			ProviderName: config.Spec.ClusterContext.ProviderName,
			Region:      config.Spec.ClusterContext.Region,
			Zone:        config.Spec.ClusterContext.Zone,
			Labels:      config.Spec.ClusterContext.Labels,
			Metadata:    config.Spec.ClusterContext.Metadata,
		})
	}

	// Get current cluster context
	clusterCtx, err := r.contextProvider.GetContext(ctx)
	if err != nil {
		log.Error(err, "failed to get cluster context")
		r.updateStatus(ctx, &config, err.Error())
		return ctrl.Result{}, err
	}

	// Update collectors based on configuration
	if err := r.updateCollectors(&config); err != nil {
		log.Error(err, "failed to update collectors")
		r.updateStatus(ctx, &config, err.Error())
		return ctrl.Result{}, err
	}

	// Collect metrics from all enabled collectors
	var allMetrics []collector.ResourceMetrics
	for _, c := range r.collectors {
		metrics, err := c.Collect(ctx)
		if err != nil {
			log.Error(err, "failed to collect metrics", "collector", c.Name())
			r.updateStatus(ctx, &config, err.Error())
			return ctrl.Result{}, err
		}
		allMetrics = append(allMetrics, metrics...)
	}

	// Send metrics to HakonGo with cluster context
	if err := r.hakongoClient.SendMetrics(ctx, &hakongo.MetricsData{
		ClusterID:     config.Spec.HakonGo.ClusterID,
		Context:       clusterCtx,
		CollectedAt:   time.Now(),
		Resources:     allMetrics,
	}); err != nil {
		log.Error(err, "failed to send metrics")
		r.updateStatus(ctx, &config, err.Error())
		return ctrl.Result{}, err
	}

	// Update status with success
	config.Status.MetricsCollected = int64(len(allMetrics))
	config.Status.LastError = ""
	if err := r.Status().Update(ctx, &config); err != nil {
		log.Error(err, "failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
}

func (r *ConnectorConfigReconciler) ensureClients() error {
	if r.kubeClient == nil {
		client, err := kubernetes.NewForConfig(r.Config)
		if err != nil {
			return fmt.Errorf("failed to create kubernetes client: %w", err)
		}
		r.kubeClient = client
	}

	if r.metricsClient == nil {
		client, err := versioned.NewForConfig(r.Config)
		if err != nil {
			return fmt.Errorf("failed to create metrics client: %w", err)
		}
		r.metricsClient = client
	}

	return nil
}

func (r *ConnectorConfigReconciler) initHakonGoClient(ctx context.Context, config *hakongov1alpha1.ConnectorConfig) error {
	// Get API key from secret
	var secret corev1.Secret
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: config.Namespace,
		Name:      config.Spec.HakonGo.APIKeySecret.Name,
	}, &secret); err != nil {
		return fmt.Errorf("failed to get API key secret: %w", err)
	}

	apiKey := string(secret.Data[config.Spec.HakonGo.APIKeySecret.Key])
	if apiKey == "" {
		return fmt.Errorf("API key not found in secret")
	}

	// Create HakonGo client
	r.hakongoClient = hakongo.NewClient(hakongo.ClientConfig{
		BaseURL: config.Spec.HakonGo.BaseURL,
		APIKey:  apiKey,
	})

	return nil
}

func (r *ConnectorConfigReconciler) updateCollectors(config *hakongov1alpha1.ConnectorConfig) error {
	if r.collectors == nil {
		r.collectors = make(map[string]collector.Collector)
	}

	// Reset hakongoClient to force re-initialization
	r.hakongoClient = nil

	return nil
}

func (r *ConnectorConfigReconciler) updateStatus(ctx context.Context, config *hakongov1alpha1.ConnectorConfig, lastError string) {
	config.Status.LastError = lastError
	if err := r.Status().Update(ctx, config); err != nil {
		log.FromContext(ctx).Error(err, "failed to update status")
	}
}

func (r *ConnectorConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Config = mgr.GetConfig()
	return ctrl.NewControllerManagedBy(mgr).
		For(&hakongov1alpha1.ConnectorConfig{}).
		Complete(r)
}

func SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hakongov1alpha1.ConnectorConfig{}).
		Complete(&ConnectorConfigReconciler{
			Client:        mgr.GetClient(),
			Scheme:       mgr.GetScheme(),
			Config:       mgr.GetConfig(),
			metricsClient: versioned.NewForConfigOrDie(mgr.GetConfig()),
			kubeClient:   kubernetes.NewForConfigOrDie(mgr.GetConfig()),
			collectors:   make(map[string]collector.Collector),
		})
}
