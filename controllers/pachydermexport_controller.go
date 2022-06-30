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
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	goerrors "errors"

	backupcmd "github.com/opdev/backup-handler/cmd/command"
	backupservice "github.com/opdev/backup-handler/gen/backup_service"
	aimlv1beta1 "github.com/pachyderm/openshift-operator/api/v1beta1"
)

// PachydermExportReconciler reconciles a PachydermExport object
type PachydermExportReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=aiml.pachyderm.com,resources=pachydermexports,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=aiml.pachyderm.com,resources=pachydermexports/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=aiml.pachyderm.com,resources=pachydermexports/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods/exec,verbs=create;get
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile method runs when an event is triggered for the watched reesources
func (r *PachydermExportReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	export := &aimlv1beta1.PachydermExport{}
	if err := r.Client.Get(ctx, req.NamespacedName, export); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// get status of the export
	exportKey := types.NamespacedName{
		Namespace: export.Namespace,
		Name:      export.Name,
	}

	current := &aimlv1beta1.PachydermExport{}
	if err := r.Get(ctx, exportKey, current); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// restore pachyderm from export
	if export.Spec.Restore != nil {
		if export.Spec.StorageSecret == "" {
			return ctrl.Result{}, goerrors.New("storage secret name required")
		}
		if err := r.restorePachyderm(ctx, export); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if !reflect.DeepEqual(current.Status, aimlv1beta1.PachydermExportStatus{}) {
		if export.Status.Restore.CompletedAt != "" {
			return ctrl.Result{}, nil
		}

		if err := r.checkBackupStatus(ctx, export); err != nil {
			return ctrl.Result{}, err
		}

		if reflect.DeepEqual(export.Status, current.Status) &&
			export.Status.Restore.CompletedAt == "" {
			return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
		}
	}

	// If status is empty, create new backup
	if err := r.newBackupTask(ctx, export); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.Status().Patch(ctx, export, client.MergeFrom(current)); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PachydermExportReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aimlv1beta1.PachydermExport{}).
		Complete(r)
}

func (r *PachydermExportReconciler) getStatefulSetPods(ctx context.Context, sts *appsv1.StatefulSet) (*corev1.PodList, error) {
	listOptions := &client.ListOptions{
		Namespace:     sts.Namespace,
		LabelSelector: labels.SelectorFromSet(sts.Spec.Template.ObjectMeta.Labels),
	}
	pods := &corev1.PodList{}
	if err := r.List(ctx, pods, listOptions); err != nil {
		return nil, err
	}

	return pods, nil
}

func newBackupRequest(export *aimlv1beta1.PachydermExport, pd *aimlv1beta1.Pachyderm, pod, container string, commands []string) ([]byte, error) {
	cr, err := json.Marshal(pd)
	if err != nil {
		return nil, err
	}

	c := backupcmd.Marshal(commands).String()

	storageSecret := export.Spec.StorageSecret
	encodedCR := base64.StdEncoding.EncodeToString(cr)

	return json.Marshal(
		&backup{
			Name:               &export.Name,
			Namespace:          &export.Namespace,
			Pod:                &pod,
			Container:          &container,
			StorageSecret:      &storageSecret,
			KubernetesResource: &encodedCR,
			Command:            &c,
		},
	)
}

func createBackup(export *aimlv1beta1.PachydermExport, pd *aimlv1beta1.Pachyderm, pods *corev1.PodList) (*backupservice.Backupresult, error) {
	if export.Status.Backup.ID != "" {
		return nil, nil
	}

	// Backup the pachyderm resource
	payload, err := newBackupRequest(
		export,
		pd,
		pods.Items[0].Name,
		"postgres",
		[]string{"bash", "-c", "pg_dump --dbname \"pachyderm\" --dbname \"dex\""},
	)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, "http://localhost:8890/backups", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	defer request.Body.Close()

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return parseBackupresult(body)
}

func parseBackupresult(body []byte) (*backupservice.Backupresult, error) {
	temp := backup{}
	if err := json.Unmarshal(body, &temp); err != nil {
		return nil, err
	}
	result := backupservice.Backupresult(temp)
	return &result, nil
}

func getBackup(export *aimlv1beta1.PachydermExport) (*backupservice.Backupresult, error) {
	url := fmt.Sprintf("http://localhost:8890/backups/%s", export.Status.Backup.ID)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		export.Status.Backup.CompletedAt = time.Now().UTC().String()
		msg := fmt.Sprintf("backup %s not found", export.Status.Backup.ID)
		return nil, goerrors.New(msg)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return parseBackupresult(body)
}

func (r *PachydermExportReconciler) newBackupTask(ctx context.Context, export *aimlv1beta1.PachydermExport) error {
	// return nil if the backup already exists
	if export.Status.Backup.ID != "" {
		return nil
	}

	pd, err := r.pachydermForBackup(ctx, export)
	if err != nil {
		return err
	}

	if err := r.pausePachydermAnnotation(ctx, pd); err != nil {
		return err
	}

	if pd.Spec.Postgres.Disable {
		return nil
	}

	pg := &appsv1.StatefulSet{}
	pgKey := types.NamespacedName{
		Namespace: pd.Namespace,
		Name:      "postgres",
	}
	if err := r.Get(ctx, pgKey, pg); err != nil {
		return err
	}

	pods, err := r.getStatefulSetPods(ctx, pg)
	if err != nil {
		return err
	}

	backup, err := createBackup(export, pd, pods)
	if err != nil {
		return err
	}

	if backup != nil {
		if backup.ID != nil {
			export.Status.Backup.ID = *backup.ID
		}

		if backup.Name != nil {
			export.Status.Backup.Name = *backup.Name
		}

		if backup.CreatedAt != nil {
			export.Status.Backup.StartedAt = *backup.CreatedAt
		}

		if backup.State != nil {
			export.Status.Phase = *backup.State
		}
	}

	return nil
}

func (r *PachydermExportReconciler) checkBackupStatus(ctx context.Context, export *aimlv1beta1.PachydermExport) error {
	if export.Status.Backup.ID == "" || export.Status.Backup.CompletedAt != "" {
		return nil
	}

	backup, err := getBackup(export)
	if err != nil {
		return err
	}

	if backup == nil {
		return nil
	}

	if backup.State != nil {
		export.Status.Phase = *backup.State
	}

	if backup.Location != nil {
		export.Status.Backup.Location = *backup.Location
	}

	if backup.DeletedAt != nil {
		export.Status.Backup.CompletedAt = *backup.DeletedAt
	}

	return nil
}

func (r *PachydermExportReconciler) pachydermForBackup(ctx context.Context, export *aimlv1beta1.PachydermExport) (*aimlv1beta1.Pachyderm, error) {
	pd := &aimlv1beta1.Pachyderm{}
	pdKey := types.NamespacedName{
		Namespace: export.Namespace,
		Name:      export.Spec.Backup.Target,
	}
	if err := r.Client.Get(ctx, pdKey, pd); err != nil {
		return nil, err
	}

	return pd, nil
}

func (r *PachydermExportReconciler) pausePachydermAnnotation(ctx context.Context, pd *aimlv1beta1.Pachyderm) error {
	if pd.Annotations == nil {
		pd.Annotations = map[string]string{
			aimlv1beta1.PachydermPauseAnnotation: "true",
		}
	}

	return r.Update(ctx, pd)
}
