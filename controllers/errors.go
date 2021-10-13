package controllers

import "errors"

var (
	ErrServiceNotReady  = errors.New("waiting for core microservice")
	ErrPasswordNotFound = errors.New("postgres-password key not found")
)
