package handlers

import (
	"errors"
	"net/http"
)

// List of errors used within the handlers package
var (
	errDatasetTypeNotSupported  = errors.New("dataset type is not supported")
	errDatasetHasNoTopics       = errors.New("no topics found for dataset")
	errMissingLatestVersionLink = errors.New("latest version link is missing from dataset API response")
)

// Map of errors to HTTP status codes
var errorToStatusCodeMap = map[error]int{
	errDatasetTypeNotSupported:  http.StatusNotFound,
	errDatasetHasNoTopics:       http.StatusInternalServerError,
	errMissingLatestVersionLink: http.StatusInternalServerError,
}
