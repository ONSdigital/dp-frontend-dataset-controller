package handlers

import (
	"net/http"
	"strings"

	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"

	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
)

// FilterPageHandler handles requests to /filters/{filterID}
func FilterPageHandler(f FilterClient, datasetClient DatasetAPISdkClient, filter, filterFlex http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)

		headers := dpDatasetApiSdk.Headers{
			CollectionID:         "",
			DownloadServiceToken: "",
			ServiceToken:         "",
			UserAccessToken:      "",
		}

		filterID := vars["filterID"]
		if filterID == "" {
			http.Error(w, "missing filter ID", http.StatusBadRequest)
			return
		}

		filterModel, _, err := f.GetJobState(
			ctx, headers.UserAccessToken, headers.ServiceToken, headers.DownloadServiceToken, headers.CollectionID, filterID,
		)
		if err != nil {
			log.Error(ctx, "failed to get filter job state", err)
			http.Error(w, "failed to get filter", http.StatusInternalServerError)
			return
		}

		datasetDetails, err := datasetClient.GetDataset(ctx, headers, headers.CollectionID, filterModel.Dataset.DatasetID)
		if err != nil {
			log.Error(ctx, "failed to get dataset details", err)
			http.Error(w, "failed to get dataset", http.StatusInternalServerError)
			return
		}

		if strings.Contains(datasetDetails.Type, "cantabular") {
			// Redirect to dp-frontend-filter-flex-dataset and continue filter-flex journey
			filterFlex.ServeHTTP(w, r)
			return
		}

		// If CMD type, the CMD filter journey works as it currently does i.e. to frontend-filter-dataset-controller
		filter.ServeHTTP(w, r)
	}
}
