package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	backupv1 "github.com/opdev/backup-handler/api/v1"
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

	if restore.ID.String() != "" {
		if export.Status.RestoreID == "" {
			export.Status.RestoreID = restore.ID.String()
		}

		if err := r.Status().Update(ctx, export); err != nil {
			return err
		}
	}

	return nil
}

func requestRestore(export *aimlv1beta1.PachydermExport) (*backupv1.Restore, error) {
	restore := &backupv1.Restore{
		Destination: backupv1.Destination{
			Name:      export.Spec.Restore.Destination.Name,
			Namespace: export.Spec.Restore.Destination.Namespace,
		},
		Backup: export.Spec.Restore.BackupName,
	}

	payload, err := json.Marshal(restore)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, "http://pachyderm-operator-pachyderm-backup-manager:8890/restore", bytes.NewBuffer(payload))
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

	if err := json.Unmarshal(body, restore); err != nil {
		return nil, err
	}

	return restore, nil
}

func getRestore(export *aimlv1beta1.PachydermExport) (*backupv1.Restore, error) {
	return nil, nil
}
