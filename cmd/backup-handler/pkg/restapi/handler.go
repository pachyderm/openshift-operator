package restapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-migrate/migrate"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	backupv1 "github.com/pachyderm/openshift-operator/cmd/backup-handler/api/v1"
	"github.com/pachyderm/openshift-operator/cmd/backup-handler/pkg/models"
)

func Start() error {
	err := models.MigrateDB()
	if err != nil && errors.Is(err, migrate.ErrNoChange) {
		log.Fatal(err.Error())
	}

	r := mux.NewRouter()
	r.HandleFunc("/backup", createBackup).Methods(http.MethodPost)
	r.HandleFunc("/backup", updateBackup).Methods(http.MethodPut)
	r.HandleFunc("/backup/{id}", getBackup).Methods(http.MethodGet)
	r.HandleFunc("/backup/{id}", deleteBackup).Methods(http.MethodDelete)
	r.HandleFunc("/next-batch", backupScheduler).Methods(http.MethodGet)

	log.Println("Starting API...")
	return http.ListenAndServe(":8890", r)
}

func createBackup(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	backup := &backupv1.Backup{}
	if err := json.NewDecoder(r.Body).Decode(backup); err != nil {
		fmt.Printf("error: %v\n", err)
	}

	if err := models.CreateBackup(backup); err != nil {
		fmt.Println(err.Error())
	}

	json.NewEncoder(w).Encode(backup)
}

func updateBackup(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	backup := &backupv1.Backup{}
	json.NewDecoder(r.Body).Decode(backup)

	records, err := models.UpdateBackup(backup)
	if err != nil {
		fmt.Println(err.Error())
	}

	if records > 0 {
		if err := models.GetBackup(backup); err != nil {
			fmt.Println(err.Error())
		}
	}

	json.NewEncoder(w).Encode(backup)
}

func getBackup(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	params := mux.Vars(r)
	id, ok := params["id"]
	if !ok {
		fmt.Println("id not found")
	}

	backup := &backupv1.Backup{
		Metadata: backupv1.Metadata{
			ID: uuid.MustParse(id),
		},
	}

	if err := models.GetBackup(backup); err != nil {
		fmt.Fprint(w, err.Error())
	}
	json.NewEncoder(w).Encode(backup)
}

func deleteBackup(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	params := mux.Vars(r)
	id, ok := params["id"]
	if !ok {
		fmt.Println("not found")
	}

	backup := &backupv1.Backup{
		Metadata: backupv1.Metadata{
			ID: uuid.MustParse(id),
		},
	}

	if err := models.GetBackup(backup); err != nil {
		log.Println("error getting backup: ", err.Error())
	}

	records, err := models.DeleteBackup(backup)
	if err != nil || records == 0 {
		log.Println("error deleting backup: ", err.Error())
	}

	json.NewEncoder(w).Encode(backup)
}

func backupScheduler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	jobs, err := models.BackupBatch()
	if err != nil {
		log.Println("error getting next job", err.Error())
	}

	json.NewEncoder(w).Encode(jobs)
}
