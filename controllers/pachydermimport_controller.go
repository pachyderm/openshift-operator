/*
Copyright 2021 Pachyderm.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"reflect"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	aimlv1beta1 "github.com/pachyderm/openshift-operator/api/v1beta1"
)

// PachydermImportReconciler reconciles a PachydermImport object
type PachydermImportReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=aiml.pachyderm.com,resources=pachydermimports,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=aiml.pachyderm.com,resources=pachydermimports/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=aiml.pachyderm.com,resources=pachydermimports/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PachydermImport object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *PachydermImportReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	restore := &aimlv1beta1.PachydermImport{}
	if err := r.Client.Get(ctx, req.NamespacedName, restore); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	logger.Info("Starting restore of backup", "backup", restore.Spec.BackupName)

	// if strings.EqualFold(restore.Status.Phase, aimlv1beta1.ExportCompletedStatus) {
	// 	return ctrl.Result{}, nil
	// }

	if err := r.restorePachyderm(ctx, restore); err != nil {
		// If the pachd deployment is not found, requeque the request
		if errors.IsNotFound(err) {
			return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		}
		if err == ErrPachdPodsRunning {
			return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		}
		if err == ErrDatabaseNotFound {
			return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		}
		if strings.Contains(err.Error(), "pachyderm resource not found") {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Retrieve the most up to date instance of the pachyderm import object
	current := &aimlv1beta1.PachydermImport{}
	if err := r.Get(ctx, req.NamespacedName, current); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// check the status of the pachyderm restore request
	if !reflect.DeepEqual(restore.Status, aimlv1beta1.PachydermImportStatus{}) {
		if restore.Status.CompletedAt != "" {
			return ctrl.Result{}, nil
		}

		if reflect.DeepEqual(restore.Status, current.Status) &&
			restore.Status.CompletedAt == "" {
			return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PachydermImportReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aimlv1beta1.PachydermImport{}).
		Complete(r)
}

func (r *PachydermImportReconciler) exitMaintenanceMode(ctx context.Context, pd *aimlv1beta1.Pachyderm) error {
	delete(pd.Annotations, aimlv1beta1.PachydermPauseAnnotation)
	return r.Update(ctx, pd)
}
