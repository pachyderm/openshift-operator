package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	goerrors "errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	restoreservice "github.com/opdev/backup-handler/gen/restore_service"
	aimlv1beta1 "github.com/pachyderm/openshift-operator/api/v1beta1"
)

func (r *PachydermExportReconciler) restorePachyderm(ctx context.Context, export *aimlv1beta1.PachydermExport) error {
	if export.Status.RestoreID != "" {
		restore, err := getRestore(export)
		if err != nil {
			return err
		}

		if restore.DeletedAt != nil {
			var t bool = true
			export.Status.RestoreCompleted = &t
		}

		return r.Status().Update(ctx, export)
	}

	restore, err := requestRestore(export)
	if err != nil {
		return err
	}

	fmt.Printf("restore response ==> %+v\n", restore)

	if restore.ID != nil {
		if export.Status.RestoreID == "" {
			export.Status.RestoreID = *restore.ID
		}

		if err := r.Status().Update(ctx, export); err != nil {
			return err
		}
	}

	return nil
}

func requestRestore(export *aimlv1beta1.PachydermExport) (*restoreservice.Restoreresult, error) {
	result := &restoreservice.Restoreresult{}
	restore := &restoreservice.Restore{
		Name:                 &export.Name,
		Namespace:            &export.Namespace,
		DestinationName:      &export.Spec.Restore.Destination.Name,
		DestinationNamespace: &export.Spec.Restore.Destination.Namespace,
		BackupLocation:       &export.Spec.Restore.BackupName,
		StorageSecret:        &export.Spec.StorageSecret,
	}

	payload, err := json.Marshal(restore)
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

	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}

	return result, nil
}

func getRestore(export *aimlv1beta1.PachydermExport) (*restoreservice.Restoreresult, error) {
	result := &restoreservice.Restoreresult{}
	url := fmt.Sprintf("http://localhost:8890/restores/%s", export.Status.RestoreID)
	request, err := http.NewRequest(http.MethodGet, url, nil)
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

	if response.StatusCode == http.StatusNotFound {
		export.Status.CompletedAt = time.Now().UTC().String()
		return nil, goerrors.New("restore not found")
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}

	return result, nil
}
