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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	aimlv1beta1 "github.com/pachyderm/openshift-operator/api/v1beta1"
)

// PachydermVaultReconciler reconciles a PachydermVault object
type PachydermVaultReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=aiml.pachyderm.com,resources=pachydermvaults,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=aiml.pachyderm.com,resources=pachydermvaults/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=aiml.pachyderm.com,resources=pachydermvaults/finalizers,verbs=update

func (r *PachydermVaultReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	vault := &aimlv1beta1.PachydermVault{}
	if err := r.Client.Get(ctx, req.NamespacedName, vault); err != nil {
		return ctrl.Result{}, err
	}

	pd := &aimlv1beta1.Pachyderm{}
	pdKey := types.NamespacedName{
		Namespace: vault.Namespace,
		Name:      vault.Spec.Backup.Pachyderm,
	}
	if err := r.Client.Get(ctx, pdKey, pd); err != nil {
		return ctrl.Result{RequeueAfter: 3 * time.Second}, err
	}

	pg := &appsv1.StatefulSet{}
	pgKey := types.NamespacedName{
		Namespace: pd.Namespace,
		Name:      "postgres",
	}
	if err := r.Get(ctx, pgKey, pg); err != nil {
		return ctrl.Result{}, err
	}

	// if pg.Status.ReadyReplicas > 0 {
	// }

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PachydermVaultReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aimlv1beta1.PachydermVault{}).
		Complete(r)
}
