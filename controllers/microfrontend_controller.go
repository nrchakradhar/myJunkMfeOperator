package controllers

import (
	context "context"
	"fmt"
	"time"

	"mfe-operator/api/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// MicroFrontendReconciler reconciles a MicroFrontend object
type MicroFrontendReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=platform.mycorp.com,resources=microfrontends,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=platform.mycorp.com,resources=microfrontends/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=platform.mycorp.com,resources=microfrontends/finalizers,verbs=update

func (r *MicroFrontendReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the MicroFrontend object
	var mfe v1alpha1.MicroFrontend
	if err := r.Get(ctx, req.NamespacedName, &mfe); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("MicroFrontend resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get MicroFrontend")
		return ctrl.Result{}, err
	}

	// Simulated: Process the OCI artifact, extract, and upload to CDN
	logger.Info("Processing MicroFrontend", "name", mfe.Name, "oci", mfe.Spec.OCIArtifact)

	// Update status
	mfe.Status.Synced = true
	mfe.Status.LastSyncedAt = time.Now().Format(time.RFC3339)
	mfe.Status.Message = "Successfully processed (simulated)"
	if err := r.Status().Update(ctx, &mfe); err != nil {
		logger.Error(err, "Failed to update MicroFrontend status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 10 * time.Minute}, nil
}

func (r *MicroFrontendReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.MicroFrontend{}).
		Complete(r)
}
