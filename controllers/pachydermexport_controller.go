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
	"strings"
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

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
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
		if export.Status.CompletedAt != "" {
			return ctrl.Result{}, nil
		}

		if err := r.checkBackupStatus(ctx, export); err != nil {
			return ctrl.Result{}, err
		}

		if reflect.DeepEqual(export.Status, current.Status) &&
			export.Status.CompletedAt == "" {
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

type backup struct {
	CreatedAt *string `form:"created_at,omitempty" json:"created_at,omitempty" xml:"created_at,omitempty"`
	UpdatedAt *string `form:"updated_at,omitempty" json:"updated_at,omitempty" xml:"updated_at,omitempty"`
	DeletedAt *string `form:"deleted_at,omitempty" json:"deleted_at,omitempty" xml:"deleted_at,omitempty"`
	ID        *string `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// Current state of the job
	State *string `form:"state,omitempty" json:"state,omitempty" xml:"state,omitempty"`
	// Name of pachyderm instance backed up
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Namespace of resource backed up
	Namespace *string `form:"namespace,omitempty" json:"namespace,omitempty" xml:"namespace,omitempty"`
	// Name of target pod
	Pod *string `form:"pod,omitempty" json:"pod,omitempty" xml:"pod,omitempty"`
	// Name of container in pod
	Container *string `form:"container,omitempty" json:"container,omitempty" xml:"container,omitempty"`
	// base64 encoded command to run in pod
	Command *string `form:"command,omitempty" json:"command,omitempty" xml:"command,omitempty"`
	// Kubernetes secret containing S3 storage credentials
	StorageSecret *string `form:"storage_secret,omitempty" json:"storage_secret,omitempty" xml:"storage_secret,omitempty"`
	// base64 encoded json representation of object
	KubernetesResource *string `form:"kubernetes_resource,omitempty" json:"kubernetes_resource,omitempty" xml:"kubernetes_resource,omitempty"`
	// BackupLocation represents the URL for the uploaded backup object
	BackupLocation *string `form:"backup_location,omitempty" json:"backup_location,omitempty" xml:"backup_location,omitempty"`
}

func newBackupRequest(export *aimlv1beta1.PachydermExport, pd *aimlv1beta1.Pachyderm, pod, container string, command []string) ([]byte, error) {
	cr, err := json.Marshal(pd)
	if err != nil {
		return nil, err
	}

	encodedCR := base64.StdEncoding.EncodeToString(cr)

	cmd := []byte(strings.Join(command, "[/]"))
	encodedCmd := base64.StdEncoding.EncodeToString(cmd)

	return json.Marshal(
		&backup{
			Name:               &export.Name,
			Namespace:          &export.Namespace,
			Pod:                &pod,
			Container:          &container,
			StorageSecret:      &export.Spec.StorageSecret,
			KubernetesResource: &encodedCR,
			Command:            &encodedCmd,
		},
	)
}

func createBackup(export *aimlv1beta1.PachydermExport, pd *aimlv1beta1.Pachyderm, pods *corev1.PodList) (*backupservice.Backupresult, error) {
	res := &backupservice.Backupresult{}
	if export.Status.BackupID != "" {
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

	request, err := http.NewRequest(http.MethodPost, "http://pachyderm-operator-pachyderm-backup-manager:8890/backups", bytes.NewBuffer(payload))
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

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, res); err != nil {
		return nil, err
	}

	return res, nil
}

func getBackup(export *aimlv1beta1.PachydermExport) (*backupservice.Backupresult, error) {
	backup := &backupservice.Backupresult{}
	url := fmt.Sprintf("http://pachyderm-operator-pachyderm-backup-manager:8890/backups/%s", export.Status.BackupID)
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
		export.Status.CompletedAt = time.Now().UTC().String()
		return nil, goerrors.New("backup not found")
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, backup); err != nil {
		return nil, err
	}

	return backup, nil
}

func (r *PachydermExportReconciler) newBackupTask(ctx context.Context, export *aimlv1beta1.PachydermExport) error {
	// return nil if the backup already exists
	if export.Status.BackupID != "" {
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
		export.Status.BackupID = *backup.ID
		export.Status.Backup = *backup.Name
		export.Status.StartedAt = *backup.CreatedAt
		export.Status.Phase = aimlv1beta1.ExportRunningStatus
	}

	return nil
}

func (r *PachydermExportReconciler) checkBackupStatus(ctx context.Context, export *aimlv1beta1.PachydermExport) error {
	if export.Status.BackupID != "" {
		backup, err := getBackup(export)
		if err != nil {
			return err
		}

		if backup != nil {
			if backup.DeletedAt != nil {
				export.Status.CompletedAt = *backup.DeletedAt
				export.Status.Phase = aimlv1beta1.ExportCompletedStatus
				export.Status.BackupLocation = *backup.BackupLocation
			}

			if export.Status.CompletedAt != "" {
				// remove pachyderm resource from maintenance mode
				pd, err := r.pachydermForBackup(ctx, export)
				if err != nil {
					return err
				}

				if pd.IsPaused() {
					delete(pd.Annotations, aimlv1beta1.PachydermPauseAnnotation)
					return r.Update(ctx, pd)
				}
			}
		}
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
