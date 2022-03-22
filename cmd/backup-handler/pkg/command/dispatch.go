package command

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	backupv1 "github.com/pachyderm/openshift-operator/cmd/backup-handler/api/v1"
)

func BackupDispatch() {
	time.Sleep(10 * time.Second)

	for {
		backups := fetchBackupBatch()

		for _, backup := range backups.Items {
			results, err := execBackup(backup)
			if err != nil {
				log.Fatalf("error running backup; %v\n", err)
			}

			if err := writeBackup(backup, results.Output()); err != nil {
				log.Printf("error writing backup.\n%v.\n", err)
			}

			fmt.Println("stderr: ", results.Error())
			if err := setBackupResults(backup); err != nil {
				log.Println("error updating backup response.", err.Error())
			}

			if err := markBackupCompleted(backup); err != nil {
				log.Fatal("error marking backup completed")
			}

			log.Printf("backup %s has completed.", backup.ID.String())
		}
		time.Sleep(3 * time.Second)
	}
}

func writeBackup(backup *backupv1.Backup, output string) error {
	var buffer bytes.Buffer
	gw := gzip.NewWriter(&buffer)
	gw.Comment = "Created by pachyderm backup helper"
	gw.ModTime = time.Now().UTC()
	gw.Name = fmt.Sprintf("%s-%s.sql", backup.Name, gw.ModTime.Format("200601021504"))

	if _, err := gw.Write([]byte(output)); err != nil {
		return err
	}

	if err := gw.Close(); err != nil {
		return err
	}

	f, err := os.Create(path.Join("/", "tmp", fmt.Sprintf("%s.gz", gw.Name)))
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(buffer.Bytes()); err != nil {
		return err
	}

	return nil
}

func setBackupResults(backup *backupv1.Backup) error {
	payload, err := json.Marshal(backup)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPut, "http://localhost:8890/backup", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	defer request.Body.Close()

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &backup); err != nil {
		return err
	}

	return nil
}

func markBackupCompleted(backup *backupv1.Backup) error {
	url := fmt.Sprintf("http://localhost:8890/backup/%s", backup.ID.String())
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

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &backup); err != nil {
		return err
	}

	return nil
}

func fetchBackupBatch() backupv1.BackupList {
	resp, err := http.Get("http://localhost:8890/next-batch")
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error retrieving backup jobs: %v\n", err)
	}

	backups := backupv1.BackupList{}
	if err := json.Unmarshal(body, &backups); err != nil {
		log.Fatalf("error unmarshaling json to struct. %v\n", err)
	}

	return backups
}

func execBackup(backup *backupv1.Backup) (*ExecResponse, error) {
	return ExecuteCommand(
		ExecOptions{
			Pod:       backup.PodName,
			Container: backup.ContainerName,
			Namespace: backup.Namespace,
			Command:   backup.Command,
		},
	)
}
