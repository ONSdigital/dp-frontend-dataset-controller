package handlers

import (
	"net/http"
	"strconv"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-net/v2/handlers"
	"github.com/gorilla/mux"
)

// Redirects to the latest version associated with a valid edition
func LatestVersionRedirect(dc DatasetClient) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		latestVersionRedirect(w, req, dc, userAccessToken, collectionID)
	})
}

func latestVersionRedirect(responseWriter http.ResponseWriter, request *http.Request, datasetClient DatasetClient,
	userAccessToken string, collectionId string) {

	downloadServiceAuthToken := ""
	getVersionsQueryParams := dataset.QueryParams{Offset: 0, Limit: 1000}
	serviceAuthToken := ""

	ctx := request.Context()
	vars := mux.Vars(request)

	datasetId := vars["datasetID"]
	editionId := vars["editionID"]

	// Fetch all versions associated with the dataset and error if not found
	versionsList, err := datasetClient.GetVersions(ctx, userAccessToken, serviceAuthToken,
		downloadServiceAuthToken, collectionId, datasetId, editionId, &getVersionsQueryParams,
	)
	if err != nil {
		setStatusCode(ctx, responseWriter, err)
		return
	}

	// Build latest version url
	latestVersionNumber := 1
	for _, singleVersion := range versionsList.Items {
		if singleVersion.Version > latestVersionNumber {
			latestVersionNumber = singleVersion.Version
		}
	}
	latestVersionURL := helpers.DatasetVersionURL(datasetId, editionId, strconv.Itoa(latestVersionNumber))

	// Redirect to latest version
	http.Redirect(responseWriter, request, latestVersionURL, http.StatusFound)
}
