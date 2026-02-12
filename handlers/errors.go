package handlers

import (
	"errors"
	"net/http"
)

// List of errors used within the handlers package
var (
	errDatasetNotStatic     = errors.New("dataset is not of type static")
	errDatasetTopicMismatch = errors.New("dataset topic does not match topic in URL")
)

// Map of errors to HTTP status codes
var errorToStatusCodeMap = map[error]int{
	errDatasetNotStatic:     http.StatusNotFound,
	errDatasetTopicMismatch: http.StatusNotFound,
}
