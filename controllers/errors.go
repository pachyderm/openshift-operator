package controllers

import "errors"

var (
	// ErrServiceNotReady returns when a microservice is not ready
	ErrServiceNotReady = errors.New("waiting for core microservice")
	// ErrPasswordNotFound returns when the postgres database password is not found
	ErrPasswordNotFound = errors.New("postgres-password key not found")
)
