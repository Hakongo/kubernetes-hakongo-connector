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
)

const (
	defaultRequeueAfter = time.Second * 300 // 5 minutes
)

// ConnectorConfigReconciler reconciles a ConnectorConfig object
type ConnectorConfigReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	Config        *rest.Config
	collectors    map[string]collector.Collector
	metricsClient versioned.Interface
}

// +kubebuilder:rbac:groups=hakongo.io,resources=connectorconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hakongo.io,resources=connectorconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hakongo.io,resources=connectorconfigs/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods;nodes;persistentvolumeclaims;services,verbs=get;list;watch
// +kubebuilder:rbac:groups=metrics.k8s.io,resources=pods;nodes,verbs=get;list

func (r *ConnectorConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("reconciling ConnectorConfig", "name", req.Name)

	// Get the ConnectorConfig
	var config hakongov1alpha1.ConnectorConfig
	if err := r.Get(ctx, req.NamespacedName, &config); err != nil {
		if errors.IsNotFound(err) {
			// Stop all collectors if config is deleted
			r.stopCollectors()
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Initialize clients if needed
	if err := r.ensureClients(); err != nil {
		logger.Error(err, "failed to initialize clients")
		return ctrl.Result{}, err
	}

	// Update collectors based on config
	if err := r.updateCollectors(&config); err != nil {
		logger.Error(err, "failed to update collectors")
		r.updateStatus(&config, err.Error())
		return ctrl.Result{}, err
	}

	// Collect metrics
	metrics, err := r.collectMetrics(ctx, &config)
	if err != nil {
		logger.Error(err, "failed to collect metrics")
		r.updateStatus(&config, err.Error())
		return ctrl.Result{}, err
	}

	// TODO: Send metrics to SaaS platform
	logger.Info("collected metrics", "count", len(metrics))

	// Update status
	r.updateStatus(&config, "Metrics collected successfully")

	// Requeue after collection interval or default
	interval := time.Duration(config.Spec.CollectionInterval) * time.Second
	if interval < defaultRequeueAfter {
		interval = defaultRequeueAfter
	}
	return ctrl.Result{RequeueAfter: interval}, nil
}

func (r *ConnectorConfigReconciler) ensureClients() error {
	if r.metricsClient == nil {
		metricsClient, err := versioned.NewForConfig(r.Config)
		if err != nil {
			return fmt.Errorf("failed to create metrics client: %w", err)
		}
		r.metricsClient = metricsClient
	}
	return nil
}

func (r *ConnectorConfigReconciler) updateCollectors(config *hakongov1alpha1.ConnectorConfig) error {
	if r.collectors == nil {
		r.collectors = make(map[string]collector.Collector)
	}

	kubeClient, err := kubernetes.NewForConfig(r.Config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	collectorConfig := collector.CollectorConfig{
		IncludeNamespaces: config.Spec.IncludeNamespaces,
		ExcludeNamespaces: config.Spec.ExcludeNamespaces,
	}

	// Update Pod collector
	if config.Spec.Collectors.EnablePodMetrics {
		if _, exists := r.collectors["pod"]; !exists {
			r.collectors["pod"] = collector.NewPodCollector(kubeClient, r.metricsClient, collectorConfig)
		}
	} else {
		delete(r.collectors, "pod")
	}

	// Update Node collector
	if config.Spec.Collectors.EnableNodeMetrics {
		if _, exists := r.collectors["node"]; !exists {
			r.collectors["node"] = collector.NewNodeCollector(kubeClient, r.metricsClient, collectorConfig)
		}
	} else {
		delete(r.collectors, "node")
	}

	// Update PV collector
	if config.Spec.Collectors.EnablePVMetrics {
		if _, exists := r.collectors["pv"]; !exists {
			r.collectors["pv"] = collector.NewPVCollector(kubeClient, r.metricsClient, collectorConfig)
		}
	} else {
		delete(r.collectors, "pv")
	}

	// Update Service collector
	if config.Spec.Collectors.EnableServiceMetrics {
		if _, exists := r.collectors["service"]; !exists {
			r.collectors["service"] = collector.NewServiceCollector(kubeClient, r.metricsClient, collectorConfig)
		}
	} else {
		delete(r.collectors, "service")
	}

	return nil
}

func (r *ConnectorConfigReconciler) collectMetrics(ctx context.Context, config *hakongov1alpha1.ConnectorConfig) ([]collector.ResourceMetrics, error) {
	var allMetrics []collector.ResourceMetrics

	for name, c := range r.collectors {
		metrics, err := c.Collect(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to collect metrics from %s: %w", name, err)
		}
		allMetrics = append(allMetrics, metrics...)
	}

	return allMetrics, nil
}

func (r *ConnectorConfigReconciler) stopCollectors() {
	r.collectors = nil
	r.metricsClient = nil
}

func (r *ConnectorConfigReconciler) updateStatus(config *hakongov1alpha1.ConnectorConfig, status string) {
	config.Status.LastCollectionTime = &metav1.Time{Time: time.Now()}
	config.Status.LastCollectionStatus = status

	if err := r.Status().Update(context.Background(), config); err != nil {
		log.FromContext(context.Background()).Error(err, "failed to update status")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConnectorConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Config = mgr.GetConfig()
	return ctrl.NewControllerManagedBy(mgr).
		For(&hakongov1alpha1.ConnectorConfig{}).
		Complete(r)
}
