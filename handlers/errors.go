package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	datasetAPIErrors "github.com/ONSdigital/dp-dataset-api/apierrors"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-frontend-dataset-controller/clients"
	"github.com/ONSdigital/log.go/v2/log"
)

// List of errors used within the handlers package
var (
	errTooManyOptions           = errors.New("too many options in dimension")
	errDatasetTypeNotSupported  = errors.New("dataset type is not supported")
	errDatasetHasNoTopics       = errors.New("no topics found for dataset")
	errMissingLatestVersionLink = errors.New("latest version link is missing from dataset API response")
)

// Map of errors to HTTP status codes
var errorToStatusCodeMap = map[error]int{
	errDatasetTypeNotSupported:  http.StatusNotFound,
	errTooManyOptions:           http.StatusRequestEntityTooLarge,
	errDatasetHasNoTopics:       http.StatusInternalServerError,
	errMissingLatestVersionLink: http.StatusInternalServerError,
}

// 404 errors within the dataset API.
var datasetAPINotFoundErrors = []error{
	datasetAPIErrors.ErrDatasetNotFound,
	datasetAPIErrors.ErrEditionNotFound,
	datasetAPIErrors.ErrEditionsNotFound,
	datasetAPIErrors.ErrVersionNotFound,
	datasetAPIErrors.ErrVersionsNotFound,
	datasetAPIErrors.ErrInstanceNotFound,
	datasetAPIErrors.ErrDimensionNotFound,
	datasetAPIErrors.ErrDimensionsNotFound,
	datasetAPIErrors.ErrDimensionNodeNotFound,
	datasetAPIErrors.ErrDimensionOptionNotFound,
	datasetAPIErrors.ErrFileMetadataNotFound,
}

// setStatusCode sets the appropriate HTTP status code based on the error type.
func setStatusCode(ctx context.Context, w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError

	if clientErr, ok := err.(clients.ClientError); ok {
		if clientErr.Code() == http.StatusNotFound {
			status = clientErr.Code()
		}
	}

	if datasetErr, ok := err.(datasetAPIModels.Error); ok {
		errCode, atoiErr := strconv.Atoi(datasetErr.Code)
		if atoiErr != nil {
			log.Error(ctx, "failed to convert error code to int", atoiErr, log.Data{"code": datasetErr.Code})
		} else if errCode == http.StatusNotFound {
			status = errCode
		}
	}

	// The dataset API SDK can return errors either as a datasetAPIModels.Error (handled above) or as a plain error
	// constructed from the response body string (handled here).
	for _, notFoundErr := range datasetAPINotFoundErrors {
		if err.Error() == notFoundErr.Error() {
			status = http.StatusNotFound
			break
		}
	}

	// Check if the error is a known error with a mapped status code.
	if mappedStatusCode, exists := errorToStatusCodeMap[err]; exists {
		status = mappedStatusCode
	}

	log.Error(ctx, "client error", err, log.Data{"setting-response-status": status})
	w.WriteHeader(status)
}
