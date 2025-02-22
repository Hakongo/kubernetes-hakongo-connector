package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	hakongov1alpha1 "github.com/hakongo/kubernetes-connector/api/v1alpha1"
)

// ConnectorConfigReconciler reconciles a ConnectorConfig object
type ConnectorConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hakongo.io,resources=connectorconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hakongo.io,resources=connectorconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hakongo.io,resources=connectorconfigs/finalizers,verbs=update

func (r *ConnectorConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("reconciling ConnectorConfig", "name", req.Name)

	var config hakongov1alpha1.ConnectorConfig
	if err := r.Get(ctx, req.NamespacedName, &config); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// TODO: Implement reconciliation logic
	// 1. Validate configuration
	// 2. Setup collectors based on config
	// 3. Update status

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConnectorConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hakongov1alpha1.ConnectorConfig{}).
		Complete(r)
}
