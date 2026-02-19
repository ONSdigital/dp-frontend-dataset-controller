package handlers

import (
	"net/http"

	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/dp-net/v3/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// VersionsList will load a list of versions for a filterable dataset
func VersionsList(dc DatasetAPISdkClient, zc ZebedeeClient, rend RenderClient) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		versionsList(w, req, dc, zc, rend, collectionID, userAccessToken, lang)
	})
}

func versionsList(responseWriter http.ResponseWriter, request *http.Request, dc DatasetAPISdkClient, zc ZebedeeClient, rend RenderClient, collectionID, userAccessToken, lang string) {
	vars := mux.Vars(request)
	datasetID := vars["datasetID"]
	editionID := vars["edition"]
	ctx := request.Context()

	headers := dpDatasetApiSdk.Headers{
		CollectionID: collectionID,
		AccessToken:  userAccessToken,
	}

	datasetDetails, err := dc.GetDataset(ctx, headers, datasetID)
	if err != nil {
		setStatusCode(ctx, responseWriter, err)
		return
	}

	if datasetDetails.Type == DatasetTypeStatic {
		log.Error(ctx, "handler does not support static datasets", errDatasetTypeNotSupported)
		setStatusCode(ctx, responseWriter, errDatasetTypeNotSupported)
		return
	}

	getVersionsQueryParams := dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}

	versionsList, err := dc.GetVersions(ctx, headers, datasetID, editionID, &getVersionsQueryParams)
	if err != nil {
		setStatusCode(ctx, responseWriter, err)
		return
	}

	homepageContent, err := zc.GetHomepageContent(ctx, userAccessToken, collectionID, lang, homepagePath)
	if err != nil {
		log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{err}), log.Data{"homepage_content": err})
	}

	editionDetails, err := dc.GetEdition(ctx, headers, datasetID, editionID)
	if err != nil {
		setStatusCode(ctx, responseWriter, err)
		return
	}

	basePage := rend.NewBasePageModel()
	m := mapper.CreateVersionsList(basePage, request, datasetDetails, editionDetails, versionsList.Items, homepageContent.ServiceMessage, homepageContent.EmergencyBanner)
	rend.BuildPage(responseWriter, m, "version-list")
}
