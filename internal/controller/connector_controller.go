package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	hakongov1alpha1 "github.com/hakongo/kubernetes-connector/api/v1alpha1"
	"github.com/hakongo/kubernetes-connector/internal/collector"
	"github.com/hakongo/kubernetes-connector/internal/hakongo"
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
	Config        *rest.Config
	collectors    map[string]collector.Collector
	metricsClient versioned.Interface
	hakongoClient *hakongo.Client
	kubeClient    kubernetes.Interface
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
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if err := r.ensureClients(); err != nil {
		log.Error(err, "failed to initialize clients")
		return ctrl.Result{}, err
	}

	if err := r.initHakonGoClient(ctx, &config); err != nil {
		log.Error(err, "failed to initialize HakonGo client")
		r.updateStatus(ctx, &config, err.Error())
		return ctrl.Result{}, err
	}

	if err := r.updateCollectors(&config); err != nil {
		log.Error(err, "failed to update collectors")
		r.updateStatus(ctx, &config, err.Error())
		return ctrl.Result{}, err
	}

	var allMetrics []collector.ResourceMetrics
	for _, c := range r.collectors {
		metrics, err := c.Collect(ctx)
		if err != nil {
			log.Error(err, "failed to collect metrics", "collector", c.Name())
			continue
		}
		allMetrics = append(allMetrics, metrics...)
	}

	if err := r.hakongoClient.SendMetrics(ctx, hakongo.MetricsData{
		ClusterID:   config.Spec.HakonGo.ClusterID,
		CollectedAt: time.Now(),
		Resources:   allMetrics,
	}); err != nil {
		log.Error(err, "failed to send metrics to HakonGo")
		r.updateStatus(ctx, &config, err.Error())
		return ctrl.Result{}, err
	}

	r.updateStatus(ctx, &config, "")
	config.Status.MetricsCollected = int32(len(allMetrics))
	if err := r.Status().Update(ctx, &config); err != nil {
		log.Error(err, "failed to update status")
		return ctrl.Result{}, err
	}

	interval, err := time.ParseDuration(config.Spec.CollectionInterval)
	if err != nil {
		interval = 5 * time.Minute
	}

	return ctrl.Result{RequeueAfter: interval}, nil
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
	secret, err := r.kubeClient.CoreV1().Secrets(config.Spec.HakonGo.APIKeySecret.Namespace).Get(
		ctx,
		config.Spec.HakonGo.APIKeySecret.Name,
		metav1.GetOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to get API key secret: %w", err)
	}

	apiKey := string(secret.Data[config.Spec.HakonGo.APIKeySecret.Key])
	if apiKey == "" {
		return fmt.Errorf("API key not found in secret")
	}

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

	collectorConfig := collector.CollectorConfig{
		IncludeNamespaces: config.Spec.IncludeNamespaces,
		ExcludeNamespaces: config.Spec.ExcludeNamespaces,
	}

	// Update Pod collector
	if config.Spec.Collectors.EnablePodMetrics {
		if _, exists := r.collectors["pod"]; !exists {
			r.collectors["pod"] = collector.NewPodCollector(r.kubeClient, r.metricsClient, collectorConfig)
		}
	} else {
		delete(r.collectors, "pod")
	}

	// Update Node collector
	if config.Spec.Collectors.EnableNodeMetrics {
		if _, exists := r.collectors["node"]; !exists {
			r.collectors["node"] = collector.NewNodeCollector(r.kubeClient, r.metricsClient, collectorConfig)
		}
	} else {
		delete(r.collectors, "node")
	}

	// Update PV collector
	if config.Spec.Collectors.EnablePVMetrics {
		if _, exists := r.collectors["pv"]; !exists {
			r.collectors["pv"] = collector.NewPVCollector(r.kubeClient, r.metricsClient, collectorConfig)
		}
	} else {
		delete(r.collectors, "pv")
	}

	// Update Service collector
	if config.Spec.Collectors.EnableServiceMetrics {
		if _, exists := r.collectors["service"]; !exists {
			r.collectors["service"] = collector.NewServiceCollector(r.kubeClient, r.metricsClient, collectorConfig)
		}
	} else {
		delete(r.collectors, "service")
	}

	return nil
}

func (r *ConnectorConfigReconciler) updateStatus(ctx context.Context, config *hakongov1alpha1.ConnectorConfig, lastError string) {
	now := metav1.Now()
	config.Status.LastCollectionTime = &now
	config.Status.LastError = lastError
}

func (r *ConnectorConfigReconciler) stopCollectors() {
	r.collectors = nil
	r.metricsClient = nil
	r.hakongoClient = nil
}

func (r *ConnectorConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Config = mgr.GetConfig()
	return ctrl.NewControllerManagedBy(mgr).
		For(&hakongov1alpha1.ConnectorConfig{}).
		Complete(r)
}

// SetupWithManager sets up the controller with the Manager.
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
