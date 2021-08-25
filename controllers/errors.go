package controllers

import "errors"

var (
	ErrServiceNotReady = errors.New("waiting for core microservice")
)
