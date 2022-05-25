package handlers

import (
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-net/v2/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"net/http"
)

// MetadataText generates a metadata text file
func MetadataText(dc DatasetClient, cfg config.Config) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		metadataText(w, req, dc, cfg, userAccessToken, collectionID)
	})
}

func metadataText(w http.ResponseWriter, req *http.Request, dc DatasetClient, cfg config.Config, userAccessToken, collectionID string) {
	vars := mux.Vars(req)
	datasetID := vars["datasetID"]
	edition := vars["edition"]
	version := vars["version"]
	ctx := req.Context()

	metadata, err := dc.GetVersionMetadata(ctx, userAccessToken, "", collectionID, datasetID, edition, version)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	dimensions, err := dc.GetVersionDimensions(ctx, userAccessToken, "", collectionID, datasetID, edition, version)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	b, err := getText(dc, userAccessToken, collectionID, datasetID, edition, version, metadata, dimensions, req)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	w.Header().Set("Content-Type", "plain/text")
	_, err = w.Write(b)
	if err != nil {
		setStatusCode(req, w, errors.Wrap(err, "failed to write metadata text response"))
	}
}
