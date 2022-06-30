package controllers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	goerrors "errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	restoreservice "github.com/opdev/backup-handler/gen/restore_service"
	aimlv1beta1 "github.com/pachyderm/openshift-operator/api/v1beta1"
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

	if export.Status.Restore.ID != "" {
		restore, err := getRestore(export)
		if err != nil {
			return err
		}

		if restore.CreatedAt != nil {
			export.Status.Restore.StartedAt = *restore.CreatedAt
		}

		if restore.DeletedAt != nil {
			bk, err := retrieveBackupContent(export, restore)
			if err != nil {
				return err
			}

			// create the pachyderm object returned from backup
			if err := func(ctx context.Context, backup *backupContent) error {
				if err := r.Create(ctx, bk.object); err != nil {
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
					Name:      bk.object.Name,
					Namespace: bk.object.Namespace,
				}
				if err := r.Get(ctx, restoreKey, restored); err != nil {
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

			export.Status.Restore.CompletedAt = *restore.DeletedAt
			export.Status.Restore.Status = "completed"
		}

		return r.Status().Update(ctx, export)
	}

	return nil
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

// retrieve backup content returns the content of the backup
func retrieveBackupContent(export *aimlv1beta1.PachydermExport, restore *restoreservice.Restoreresult) (*backupContent, error) {
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
