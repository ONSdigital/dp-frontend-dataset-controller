package handlers

import (
	"fmt"
	"net/http"

	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-net/v3/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// ApproveDatasetVersion handles the approve button click on static dataset version pages
func ApproveDatasetVersion(dc DatasetAPISdkClient, cfg config.Config) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		approveDatasetVersion(w, req, dc, collectionID, userAccessToken, lang)
	})
}

func approveDatasetVersion(w http.ResponseWriter, req *http.Request, dc DatasetAPISdkClient, collectionID, userAccessToken, lang string) {
	ctx := req.Context()

	vars := mux.Vars(req)
	datasetID := vars["datasetID"]
	editionID := vars["editionID"]
	versionID := vars["versionID"]

	headers := dpDatasetApiSdk.Headers{
		CollectionID:         collectionID,
		DownloadServiceToken: "",
		AccessToken:          userAccessToken,
	}

	logData := log.Data{"datasetID": datasetID, "editionID": editionID, "versionID": versionID}

	err := dc.PutVersionState(ctx, headers, datasetID, editionID, versionID, "approved")
	if err != nil {
		log.Error(ctx, "dataset version approval failed", err, logData)
	} else {
		log.Info(ctx, "dataset version approval successful", logData)
	}

	uri := fmt.Sprintf("/datasets/%s/editions/%s/versions/%s", datasetID, editionID, versionID)
	http.Redirect(w, req, uri, http.StatusTemporaryRedirect)
}
