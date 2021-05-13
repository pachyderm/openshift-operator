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
	"encoding/base64"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	aimlv1beta1 "github.com/OchiengEd/pachyderm-operator/api/v1beta1"
	"github.com/OchiengEd/pachyderm-operator/controllers/generators"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

const (
	finalizer string = "operator.pachyderm.com"

	// ErrEtcdNotReady is returned when Etcd is not ready
	ErrEtcdNotReady generators.PachydermError = "waiting for etcd"
)

// PachydermReconciler reconciles a Pachyderm object
type PachydermReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=aiml.pachyderm.com,resources=pachyderms,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=aiml.pachyderm.com,resources=pachyderms/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=aiml.pachyderm.com,resources=pachyderms/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *PachydermReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("pachyderm", req.NamespacedName)

	pd := &aimlv1beta1.Pachyderm{}
	if err := r.Get(ctx, req.NamespacedName, pd); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if err := r.reconcileFinalizer(ctx, pd); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileStatus(ctx, pd); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		if errors.IsResourceExpired(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}

	if err := r.reconcilePachydermObj(ctx, pd); err != nil {
		if err == ErrEtcdNotReady {
			return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PachydermReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aimlv1beta1.Pachyderm{}).
		Owns(&networkingv1.Ingress{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		WithEventFilter(filterEvents()).
		Complete(r)
}

func filterEvents() predicate.Funcs {
	return predicate.Funcs{
		DeleteFunc: func(event.DeleteEvent) bool {
			// enable sending delete functions
			// to the reconcile function
			return true
		},
	}
}

func (r *PachydermReconciler) reconcilePachydermObj(ctx context.Context, pd *aimlv1beta1.Pachyderm) error {
	components := generators.Prepare(pd)

	// Deploy service accounts
	if err := r.reconcileServiceAccounts(ctx, components); err != nil {
		return err
	}

	// Deploy secrets
	if err := r.reconcileSecrets(ctx, components); err != nil {
		return err
	}

	// Deploy services
	if err := r.reconcileServices(ctx, components); err != nil {
		return err
	}

	// Deploy storage class
	if err := r.reconcileStorageClass(ctx, components); err != nil {
		return err
	}

	if err := r.deployEtcd(ctx, components); err != nil {
		return err
	}

	if err := r.deployPachd(ctx, components); err != nil {
		return err
	}

	if err := r.deployDash(ctx, components); err != nil {
		return err
	}

	return nil
}

// TODO: cleanup Pachyderm objects
// - service accounts
func (r *PachydermReconciler) cleanupPachydermObj(ctx context.Context, pd *aimlv1beta1.Pachyderm) error {
	fmt.Println("delete pachyderm child resources")
	return nil
}

// TODO: set finalizer and status for Pachyderm resource
func (r *PachydermReconciler) reconcileStatus(ctx context.Context, pd *aimlv1beta1.Pachyderm) error {
	current := &aimlv1beta1.Pachyderm{}
	pdKey := types.NamespacedName{
		Name:      pd.Name,
		Namespace: pd.Namespace,
	}

	if err := r.Get(ctx, pdKey, current); err != nil {
		if errors.IsNotFound(err) {
			// TODO: do something
			return nil
		}
		return err
	}

	if reflect.DeepEqual(current.Status, aimlv1beta1.PachydermStatus{}) &&
		pd.DeletionTimestamp == nil {
		current.Status.Phase = aimlv1beta1.PhaseInitializing
	}

	return r.Status().Patch(ctx, current, client.MergeFrom(pd))
	// return r.Status().Update(ctx, current)
}

func (r *PachydermReconciler) reconcileFinalizer(ctx context.Context, pd *aimlv1beta1.Pachyderm) error {
	currentFinalizers := pd.Finalizers

	// add finalizer for new Pachyderm resource
	if pd.DeletionTimestamp == nil && len(pd.ObjectMeta.Finalizers) == 0 {
		pd.ObjectMeta.Finalizers = []string{finalizer}
	}

	// perform clean up and delete finalizer otherwise
	if len(pd.ObjectMeta.Finalizers) > 0 && pd.DeletionTimestamp != nil {
		if err := r.cleanupPachydermObj(ctx, pd); err != nil {
			return err
		}
		// remove finalizer if clean up is successful
		pd.Finalizers = []string{}
	}

	if reflect.DeepEqual(pd.Finalizers, currentFinalizers) {
		return nil
	}

	return r.Update(ctx, pd)
}

func (r *PachydermReconciler) reconcileServiceAccounts(ctx context.Context, components generators.PachydermComponents) error {
	pd := components.Parent()

	for _, sa := range components.ServiceAccounts {
		sa.Namespace = pd.Namespace
		// add owner references
		if err := controllerutil.SetControllerReference(pd, &sa, r.Scheme); err != nil {
			return err
		}

		if err := r.Create(ctx, &sa); err != nil {
			if errors.IsAlreadyExists(err) {
				return nil
			}

			return err
		}
	}

	return nil
}

func (r *PachydermReconciler) reconcileServices(ctx context.Context, components generators.PachydermComponents) error {
	pd := components.Parent()

	for _, svc := range components.Services {
		svc.Namespace = pd.Namespace
		// add owner references
		if err := controllerutil.SetControllerReference(pd, &svc, r.Scheme); err != nil {
			return err
		}

		if err := r.Create(ctx, &svc); err != nil {
			if errors.IsAlreadyExists(err) {
				// Check if the secret contents have changed
				current := &corev1.Service{}
				svcKey := types.NamespacedName{
					Name:      svc.Name,
					Namespace: pd.Namespace,
				}

				if err := r.Get(ctx, svcKey, current); err != nil {
					return err
				}

				if serviceChanged(current, &svc) {
					if err := r.Update(ctx, current); err != nil {
						return err
					}
				}

				return nil
			}

			return err
		}
	}

	return nil
}

func (r *PachydermReconciler) reconcileSecrets(ctx context.Context, components generators.PachydermComponents) error {
	pd := components.Parent()

	for _, secret := range components.Secrets() {
		// set owner reference
		if err := controllerutil.SetControllerReference(pd, secret, r.Scheme); err != nil {
			return err
		}

		if err := r.Create(ctx, secret); err != nil {
			if errors.IsAlreadyExists(err) {
				// Check if the secret contents have changed
				currentSecret := &corev1.Secret{}
				secretKey := types.NamespacedName{
					Name:      secret.Name,
					Namespace: pd.Namespace,
				}

				if err := r.Get(ctx, secretKey, currentSecret); err != nil {
					return err
				}

				if !reflect.DeepEqual(secret.Data, currentSecret.Data) {
					currentSecret.Data = secret.Data

					if err := r.Update(ctx, currentSecret); err != nil {
						return err
					}
				}
				// secret exists
				return nil
			}

			return err
		}
	}

	return nil
}

func (r *PachydermReconciler) deployEtcd(ctx context.Context, components generators.PachydermComponents) error {
	pd := components.Parent()

	etcd := components.EtcdStatefulSet()
	etcd.Namespace = pd.Namespace

	if err := controllerutil.SetControllerReference(pd, etcd, r.Scheme); err != nil {
		return err
	}

	if err := r.Create(ctx, etcd); err != nil {
		if errors.IsAlreadyExists(err) {
			// TODO: add update logic
			return nil
		}
		return err
	}

	return nil
}

func (r *PachydermReconciler) deployPachd(ctx context.Context, components generators.PachydermComponents) error {
	pd := components.Parent()

	// Check Etcd is ready before deploying pachd
	etcdSvc := types.NamespacedName{
		Name:      "etcd",
		Namespace: pd.Namespace,
	}
	if !r.isServiceReady(ctx, etcdSvc) {
		return ErrEtcdNotReady
	}

	pachd := components.PachdDeployment()
	pachd.Namespace = pd.Namespace

	if err := controllerutil.SetControllerReference(pd, pachd, r.Scheme); err != nil {
		return err
	}

	if err := r.Create(ctx, pachd); err != nil {
		if errors.IsAlreadyExists(err) {
			// TODO: add update logic
			return nil
		}
		return err
	}

	return nil
}

func (r *PachydermReconciler) deployDash(ctx context.Context, components generators.PachydermComponents) error {
	pd := components.Parent()

	dash := components.DashDeployment()
	dash.Namespace = pd.Namespace

	if err := controllerutil.SetControllerReference(pd, dash, r.Scheme); err != nil {
		return err
	}

	if err := r.Create(ctx, dash); err != nil {
		if errors.IsAlreadyExists(err) {
			// TODO: add update logic
			return nil
		}
		return err
	}

	return nil
}

func (r *PachydermReconciler) reconcileStorageClass(ctx context.Context, components generators.PachydermComponents) error {
	sc := components.StorageClass()
	if sc == nil {
		return nil
	}

	if err := r.Create(ctx, sc); err != nil {
		if errors.IsAlreadyExists(err) {
			// TODO: implement update logic
			return nil
		}
	}

	return nil
}

func (r *PachydermReconciler) isServiceReady(ctx context.Context, service types.NamespacedName) bool {
	ep := &corev1.Endpoints{}
	if err := r.Get(ctx, service, ep); err != nil {
		return false
	}

	addresses := []corev1.EndpointAddress{}

	for _, subset := range ep.Subsets {
		addresses = append(addresses, subset.Addresses...)
	}

	return len(addresses) > 0
}

func base64Encode(input []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(input))
}
