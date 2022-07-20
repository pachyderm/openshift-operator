package controllers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	goerrors "errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	restoreservice "github.com/opdev/backup-handler/gen/restore_service"
	aimlv1beta1 "github.com/pachyderm/openshift-operator/api/v1beta1"
)

var (
	ErrPostgresNotReady = goerrors.New("postgres pod not ready")
)

func (r *PachydermExportReconciler) restorePachyderm(ctx context.Context, export *aimlv1beta1.PachydermExport) error {
	if export.Spec.Restore != nil && export.Status.Restore.ID == "" {
		restore, err := requestRestore(export)
		if err != nil {
			return err
		}

		if restore.ID != nil {
			if export.Status.Restore.ID == "" {
				export.Status.Restore.ID = *restore.ID
			}

			if err := r.Status().Update(ctx, export); err != nil {
				return err
			}
		}
	}

	restore, err := getRestore(export)
	if err != nil {
		return err
	}

	if restore.CreatedAt != nil {
		export.Status.Restore.StartedAt = *restore.CreatedAt
	}

	bk, err := decodeBackupContent(export, restore)
	if err != nil {
		return err
	}

	// create the pachyderm object returned from backup
	if err := func(ctx context.Context, backup *backupContent) error {
		if err := r.Create(ctx, backup.object); err != nil {
			if errors.IsAlreadyExists(err) {
				return nil
			}
			return err
		}
		return nil
	}(ctx, bk); err != nil {
		return err
	}

	// set pachyderm instance in maintenanace mode
	restored := &aimlv1beta1.Pachyderm{}
	if err := func(name, namespace string, pd *aimlv1beta1.Pachyderm) error {
		restoreKey := types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		}
		if err := r.Get(ctx, restoreKey, pd); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return err
		}

		if restored.Annotations == nil {
			restored.Annotations = map[string]string{
				aimlv1beta1.PachydermPauseAnnotation: "true",
			}
		}

		return r.Update(ctx, restored)
	}(bk.object.Name, bk.object.Namespace, restored); err != nil {
		return err
	}

	// restore the pachyderm database
	// remove from maintenance mode
	if err := r.initiateDBRestore(ctx, restored, restore); err != nil {
		return err
	}
	log.Println("Database restore completed")

	if restore.DeletedAt != nil {
		export.Status.Restore.CompletedAt = *restore.DeletedAt
		export.Status.Restore.Status = "completed"
	}

	if err := r.Status().Update(ctx, export); err != nil {
		return err
	}

	return r.exitMaintenanceMode(ctx, restored)
}

func requestRestore(export *aimlv1beta1.PachydermExport) (*restoreservice.Restoreresult, error) {
	payload, err := newRestorequest(export)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, "http://localhost:8890/restores", bytes.NewBuffer(payload))
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

	return parseRestoreresult(body)
}

func getRestore(export *aimlv1beta1.PachydermExport) (*restoreservice.Restoreresult, error) {
	url := fmt.Sprintf("http://localhost:8890/restores/%s", export.Status.Restore.ID)
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
		export.Status.Restore.CompletedAt = time.Now().UTC().String()
		return nil, goerrors.New("restore not found")
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return parseRestoreresult(body)
}

func parseRestoreresult(body []byte) (*restoreservice.Restoreresult, error) {
	payload := &restore{}
	if err := json.Unmarshal(body, payload); err != nil {
		return nil, err
	}

	response := restoreservice.Restoreresult(*payload)

	return &response, nil
}

func newRestorequest(export *aimlv1beta1.PachydermExport) ([]byte, error) {
	restoreObj := &restore{
		Name:                 &export.Name,
		Namespace:            &export.Namespace,
		DestinationName:      &export.Spec.Restore.Destination.Name,
		DestinationNamespace: &export.Spec.Restore.Destination.Namespace,
		BackupLocation:       &export.Spec.Restore.BackupName,
		StorageSecret:        &export.Spec.StorageSecret,
	}

	return json.Marshal(restoreObj)
}

// backupContent returns the backup contents
type backupContent struct {
	// holds database dump
	database []byte
	// holds pachyderm object backup
	object *aimlv1beta1.Pachyderm
}

// decode backup content returns the base64 decoded contents of the backup
func decodeBackupContent(export *aimlv1beta1.PachydermExport, restore *restoreservice.Restoreresult) (*backupContent, error) {
	if restore.KubernetesResource == nil {
		return nil, ErrPachydermNotFound
	}

	if restore.Database == nil {
		return nil, ErrDatabaseNotFound
	}

	cr, err := decode(restore.KubernetesResource)
	if err != nil {
		return nil, err
	}

	db, err := decode(restore.Database)
	if err != nil {
		return nil, err
	}

	pd, err := func(name, namespace string, payload []byte) (*aimlv1beta1.Pachyderm, error) {
		pd := &aimlv1beta1.Pachyderm{}
		if err := json.Unmarshal(payload, pd); err != nil {
			return nil, err
		}

		if name != "" && namespace != "" {
			pd.ObjectMeta = metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			}
		}

		pd.Status = aimlv1beta1.PachydermStatus{}

		return pd, nil
	}(
		export.Spec.Restore.Destination.Name,
		export.Spec.Restore.Destination.Namespace,
		cr,
	)
	if err != nil {
		return nil, err
	}

	return &backupContent{
		database: db,
		object:   pd,
	}, nil
}

func decode(payload *string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(*payload)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *PachydermExportReconciler) initiateDBRestore(ctx context.Context, pd *aimlv1beta1.Pachyderm, restore *restoreservice.Restoreresult) error {
	pachd := &appsv1.Deployment{}
	pachdKey := types.NamespacedName{
		Namespace: pd.Namespace,
		Name:      "pachd",
	}
	if err := r.Get(ctx, pachdKey, pachd); err != nil {
		return err
	}

	if *pachd.Spec.Replicas != 0 {
		return ErrPachdPodsRunning
	}

	url := fmt.Sprintf("http://localhost:8890/restores/%s", *restore.ID)
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	result, err := parseRestoreresult(body)
	if err != nil {
		return err
	}
	restore.DeletedAt = result.DeletedAt

	return nil
}

func (r *PachydermExportReconciler) exitMaintenanceMode(ctx context.Context, pd *aimlv1beta1.Pachyderm) error {
	delete(pd.Annotations, aimlv1beta1.PachydermPauseAnnotation)
	return r.Update(ctx, pd)
}
