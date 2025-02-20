package handlers

import (
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-net/v2/handlers"
	"github.com/gorilla/mux"
)

// Redirects to the latest version associated with a valid edition
func LatestVersionRedirect(dc DatasetClient, pc PopulationClient, rend RenderClient, zc ZebedeeClient, cfg config.Config, apiRouterVersion string) http.HandlerFunc {
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

	// Fetch dataset and error if it isn't found
	datasetDetails, err := datasetClient.Get(ctx, userAccessToken, serviceAuthToken, collectionId, datasetId)
	if err != nil {
		setStatusCode(ctx, responseWriter, err)
		return
	}

	// Fetch all versions associated with the dataset to determine latest
	versionsList, err := datasetClient.GetVersions(ctx, userAccessToken, serviceAuthToken,
		downloadServiceAuthToken, collectionId, datasetId, editionId, &getVersionsQueryParams,
	)
	if err != nil {
		setStatusCode(ctx, responseWriter, err)
		return
	}

	// Build latest version url

	// Redirect to latest version
	http.Redirect(responseWriter, request, latestVersionURL, http.StatusFound)
	return
}
