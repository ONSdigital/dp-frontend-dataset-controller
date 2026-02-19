package handlers

import (
	"errors"
	"net/http"
)

// List of errors used within the handlers package
var (
	errDatasetTypeNotSupported = errors.New("dataset type is not supported")
	errDatasetTopicMismatch    = errors.New("dataset topic does not match topic in URL")
)

// Map of errors to HTTP status codes
var errorToStatusCodeMap = map[error]int{
	errDatasetTypeNotSupported: http.StatusNotFound,
	errDatasetTopicMismatch:    http.StatusNotFound,
}
