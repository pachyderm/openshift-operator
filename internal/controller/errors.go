package controller

import (
	"errors"
)

var (
	// ErrServiceNotReady returns when a microservice is not ready
	ErrServiceNotReady = errors.New("waiting for core microservice")
	// ErrPasswordNotFound returns when the postgres database password is not found
	ErrPasswordNotFound = errors.New("postgres-password key not found")
	// ErrorPachydermNotFound is returned when the pachyderm object is not found in the restore
	ErrPachydermNotFound = errors.New("pachyderm resource not found")
	// ErrDatabaseNotFound is returned when the database dump is not found in the restore
	ErrDatabaseNotFound = errors.New("database restore not found")
	// ErrPachdPodsRunning is returned when pachd pods are running while in maintenance mode
	ErrPachdPodsRunning = errors.New("pachd pods still running")
)
