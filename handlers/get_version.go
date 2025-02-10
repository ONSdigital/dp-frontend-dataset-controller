package handlers

import (
	"net/http"
	"strconv"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-net/v2/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

func GetVersion(dc DatasetClient, pc PopulationClient, rend RenderClient, zc ZebedeeClient, cfg config.Config, apiRouterVersion string) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		getVersion(w, req, dc, pc, rend, zc, cfg, collectionID, lang, apiRouterVersion, userAccessToken)
	})
}

// Returns page to render a version associated with an edition. If a version url param is not
// supplied, returns the latest version associated with the edition.
func getVersion(responseWriter http.ResponseWriter, request *http.Request, datasetClient DatasetClient, populationClient PopulationClient, renderClient RenderClient, zebedeeClient ZebedeeClient, cfg config.Config, collectionId, lang, apiRouterVersion, userAccessToken string) {
	// Defaults
	downloadServiceAuthToken := ""
	serviceAuthToken := ""

	vars := mux.Vars(request)

	datasetId := vars["datasetID"]
	editionId := vars["editionID"]
	versionId := vars["versionID"]

	context := request.Context()

	// Fetch the dataset
	datasetDetails, err := datasetClient.Get(context, userAccessToken, serviceAuthToken, collectionId, datasetId)
	if err != nil {
		setStatusCode(context, responseWriter, err)
		return
	}

	// Fetch the versions associated with the dataset edition
	getVersionsQueryParams := dataset.QueryParams{Offset: 0, Limit: 1000}
	versionsList, err := datasetClient.GetVersions(context, userAccessToken, serviceAuthToken, downloadServiceAuthToken, collectionId, datasetId, editionId, &getVersionsQueryParams)
	if err != nil {
		setStatusCode(context, responseWriter, err)
		return
	}

	hasOtherVersions := false
	if len(allVers.Items) > 1 {
		hasOtherVersions = true
	}
	allVersions := allVers.Items

	var displayOtherVersionsLink bool
	if len(allVers.Items) > 1 {
		displayOtherVersionsLink = true
	}

	latestVersionNumber := 1
	for _, singleVersion := range allVers.Items {
		if singleVersion.Version > latestVersionNumber {
			latestVersionNumber = singleVersion.Version
		}
	}

	latestVersionURL := helpers.DatasetVersionURL(datasetID, edition, strconv.Itoa(latestVersionNumber))

	if len(versionId) > 0 {
		// `versionId` url param is not set as part of the request, so return that version

	} else {
		// `versionId` url param is not set as part of the request, so redirect to the latest
		// version
		log.Info(context, "no version provided, therefore redirecting to latest version", log.Data{"latestVersionURL": latestVersionURL})
		http.Redirect(responseWriter, request, latestVersionURL, http.StatusFound)
		return
	}
}
