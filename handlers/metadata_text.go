package handlers

import (
	"net/http"

	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-net/v3/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// MetadataText generates a metadata text file
func MetadataText(dc DatasetAPISdkClient, cfg config.Config) http.HandlerFunc {
	return handlers.ControllerHandler(func(responseWriter http.ResponseWriter, request *http.Request, lang, collectionID, userAccessToken string) {
		metadataText(responseWriter, request, dc, cfg, userAccessToken, collectionID)
	})
}

func metadataText(responseWriter http.ResponseWriter, request *http.Request, dc DatasetAPISdkClient, cfg config.Config, userAccessToken, collectionID string) {
	downloadServiceAuthToken := ""
	serviceAuthToken := ""

	ctx := request.Context()
	vars := mux.Vars(request)

	datasetID := vars["datasetID"]
	editionID := vars["editionID"]
	versionID := vars["versionID"]

	headers := dpDatasetApiSdk.Headers{
		CollectionID:         collectionID,
		DownloadServiceToken: downloadServiceAuthToken,
		ServiceToken:         serviceAuthToken,
		UserAccessToken:      userAccessToken,
	}

	metadata, err := dc.GetVersionMetadata(ctx, headers, datasetID, editionID, versionID)
	if err != nil {
		setStatusCode(ctx, responseWriter, err)
		return
	}

	dimensions, err := dc.GetVersionDimensions(ctx, headers, datasetID, editionID, versionID)
	if err != nil {
		setStatusCode(ctx, responseWriter, err)
		return
	}

	b, err := getText(ctx, dc, headers, datasetID, editionID, versionID, metadata, dimensions)
	if err != nil {
		setStatusCode(ctx, responseWriter, err)
		return
	}

	responseWriter.Header().Set("Content-Type", "plain/text")
	_, err = responseWriter.Write(b)
	if err != nil {
		setStatusCode(ctx, responseWriter, errors.Wrap(err, "failed to write metadata text response"))
	}
}
