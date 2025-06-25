package handlers

import (
	"net/http"
	"strings"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"

	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
)

// FilterPageHandler handles requests to /filters/{filterID}
func FilterPageHandler(f FilterClient, datasetClient DatasetAPISdkClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)
		cfg, err := config.Get()

		if err != nil {
			log.Error(ctx, "unable to retrieve service configuration", err)
			return
		}

		filterID := vars["filterID"]
		if filterID == "" {
			http.Error(w, "missing filter ID", http.StatusBadRequest)
			return
		}

		jobStateUserAuthToken := ""
		jobStateServiceAuthToken := "" 
		jobStateDownloadServiceToken := "" 
		jobStateCollectionID := ""

		filterModel, _, err := f.GetJobState(
			ctx, jobStateUserAuthToken, jobStateServiceAuthToken, jobStateDownloadServiceToken, jobStateCollectionID, filterID,
		)
		if err != nil {
			log.Error(ctx, "failed to get filter job state", err)
			http.Error(w, "failed to get filter", http.StatusInternalServerError)
			return
		}

		getDatasetHeaders := dpDatasetApiSdk.Headers{}
		datasetCollectionID := ""

		datasetDetails, err := datasetClient.GetDataset(ctx, getDatasetHeaders, datasetCollectionID, filterModel.Dataset.DatasetID)
		if err != nil {
			log.Error(ctx, "failed to get dataset details", err)
			http.Error(w, "failed to get dataset", http.StatusInternalServerError)
			return
		}

		if strings.Contains(datasetDetails.Type, "cantabular") {
			// Redirect to dp-frontend-filter-flex-dataset and continue filter-flex journey
			target := cfg.FilterFlexDatasetServiceURL + r.URL.Path
			http.Redirect(w, r, target, http.StatusTemporaryRedirect)
			return
		}

		// If CMD type, the CMD filter journey works as it currently does i.e. to frontend-filter-dataset-controller
		target := cfg.FrontendFilterDatasetControllerURL + r.URL.Path
		http.Redirect(w, r, target, http.StatusTemporaryRedirect)
	}
}
