package models

import (
	"database/sql"
	"errors"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/mattn/go-sqlite3"
)

const BackupDB = "/var/backup-handler/backup.db"

func connect() (*sql.DB, error) {
	connection, err := sql.Open("sqlite3", BackupDB)
	if err != nil {
		return nil, err
	}

	if err := connection.Ping(); err != nil {
		return nil, err
	}

	return connection, nil
}

// MigrateDB runs database migrations
func MigrateDB() error {
	if _, err := os.Stat(BackupDB); errors.Is(err, os.ErrNotExist) {
		_, err = os.Create(BackupDB)
		if err != nil {
			return err
		}
	}

	db, err := connect()
	if err != nil {
		return err
	}
	defer db.Close()

	log.Println("Starting database migrations...")
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}

	handler, err := migrate.NewWithDatabaseInstance(
		"file:///migrations",
		"sqlite3",
		driver,
	)
	if err != nil {
		return err
	}

	return handler.Up()
}
