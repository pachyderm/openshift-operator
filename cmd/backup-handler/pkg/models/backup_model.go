package models

import (
	"context"
	"database/sql"
	"errors"

	backupv1 "github.com/pachyderm/openshift-operator/cmd/backup-handler/api/v1"
)

func CreateBackup(backup *backupv1.Backup) error {
	db, err := connect()
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	backup.New()
	return createBackup(context.Background(), db, backup)
}

func createBackup(ctx context.Context, db *sql.DB, backup *backupv1.Backup) error {
	query := "INSERT INTO backups(created_at, id, name, namespace, is_running, pod_name, container_name, command) VALUES(?, ?, ?, ?, ?, ?, ?, ?)"
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer db.Close()

	backup.EncodeCmd()
	results, err := stmt.ExecContext(ctx,
		backup.CreatedAt,
		backup.ID,
		backup.Name,
		backup.Namespace,
		backup.IsRunning,
		backup.PodName,
		backup.ContainerName,
		backup.Cmd,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if count, _ := results.RowsAffected(); count == 0 {
		return errors.New("error adding record")
	}

	return nil
}

func GetBackup(backup *backupv1.Backup) error {
	db, err := connect()
	if err != nil {
		return err
	}

	return getBackup(context.Background(), db, backup)
}

func getBackup(ctx context.Context, db *sql.DB, backup *backupv1.Backup) error {
	query := "SELECT created_at, updated_at, deleted_at, id, name, namespace, is_running, pod_name, container_name, command FROM backups WHERE id = ?"
	row := db.QueryRowContext(ctx, query, backup.ID)
	if err := row.Scan(
		&backup.CreatedAt,
		&backup.UpdatedAt,
		&backup.DeletedAt,
		&backup.ID,
		&backup.Name,
		&backup.Namespace,
		&backup.IsRunning,
		&backup.PodName,
		&backup.ContainerName,
		&backup.Cmd,
	); err != nil {
		return err
	}
	backup.DecodeCmd()

	return nil
}

func UpdateBackup(backup *backupv1.Backup) (int64, error) {
	db, err := connect()
	if err != nil {
		return 0, err
	}

	if err := db.Ping(); err != nil {
		return 0, err
	}

	backup.SetUpdatedTime()
	return updateBackup(context.Background(), db, backup)
}

// TODO: cleanup function implementation
func updateBackup(ctx context.Context, db *sql.DB, backup *backupv1.Backup) (int64, error) {
	query := "UPDATE backups SET updated_at = ?, name = ?,  is_running = ?,  pod_name = ?,  container_name = ?, command = ? WHERE id = ?"
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	defer stmt.Close()

	backup.EncodeCmd()
	results, err := stmt.ExecContext(ctx,
		backup.UpdatedAt,
		backup.Name,
		backup.IsRunning,
		backup.PodName,
		backup.ContainerName,
		backup.Cmd,
		backup.ID,
	)
	if err != nil {
		return 0, err
	}

	return results.RowsAffected()
}

func DeleteBackup(backup *backupv1.Backup) (int64, error) {
	db, err := connect()
	if err != nil {
		return 0, err
	}

	if err := db.Ping(); err != nil {
		return 0, err
	}

	backup.SetDeletedTime()
	return deleteBackup(context.Background(), db, backup)
}

func deleteBackup(ctx context.Context, db *sql.DB, backup *backupv1.Backup) (int64, error) {
	query := "UPDATE backups SET deleted_at = ? WHERE id = ?"
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	defer stmt.Close()

	results, err := stmt.ExecContext(ctx,
		backup.DeletedAt,
		backup.ID)
	if err != nil {
		return 0, err
	}

	return results.RowsAffected()
}

func BackupBatch() (*backupv1.BackupList, error) {
	db, err := connect()
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return nextBackupBatch(context.Background(), db)
}

func nextBackupBatch(ctx context.Context, db *sql.DB) (*backupv1.BackupList, error) {
	query := `SELECT created_at, updated_at, deleted_at, id, name, namespace, is_running, pod_name,
container_name, command FROM backups WHERE deleted_at is null ORDER BY created_at ASC LIMIT 5`
	results, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer results.Close()

	backups := &backupv1.BackupList{}
	for results.Next() {
		backup := backupv1.Backup{}
		err := results.Scan(
			&backup.CreatedAt,
			&backup.UpdatedAt,
			&backup.DeletedAt,
			&backup.ID,
			&backup.Name,
			&backup.Namespace,
			&backup.IsRunning,
			&backup.PodName,
			&backup.ContainerName,
			&backup.Cmd,
		)
		if err != nil {
			return nil, err
		}
		backup.DecodeCmd()
		backups.Items = append(backups.Items, &backup)
	}
	if err := results.Err(); err != nil {
		return nil, err
	}

	return backups, nil
}
