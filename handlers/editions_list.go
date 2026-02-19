package handlers

import (
	"net/http"

	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/dp-net/v3/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// EditionsList will load a list of editions for a filterable dataset
func EditionsList(dc DatasetAPISdkClient, zc ZebedeeClient, rend RenderClient, apiRouterVersion string) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		editionsList(w, req, dc, zc, rend, collectionID, lang, apiRouterVersion, userAccessToken)
	})
}

func editionsList(w http.ResponseWriter, req *http.Request, dc DatasetAPISdkClient, zc ZebedeeClient, rend RenderClient, collectionID, lang, apiRouterVersion, userAccessToken string) {
	vars := mux.Vars(req)
	datasetID := vars["datasetID"]
	ctx := req.Context()

	headers := dpDatasetApiSdk.Headers{
		AccessToken:  userAccessToken,
		CollectionID: collectionID,
	}

	datasetDetails, err := dc.GetDataset(ctx, headers, datasetID)
	if err != nil {
		setStatusCode(ctx, w, err)
		return
	}

	if datasetDetails.Type == DatasetTypeStatic {
		log.Error(ctx, "handler does not support static datasets", errDatasetTypeNotSupported)
		setStatusCode(ctx, w, errDatasetTypeNotSupported)
		return
	}

	queryParams := dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}
	datasetEditions, err := dc.GetEditions(ctx, headers, datasetID, &queryParams)
	if err != nil {
		if err, ok := err.(ClientError); ok {
			if err.Code() != http.StatusNotFound {
				setStatusCode(ctx, w, err)
				return
			}
		}
	}

	// redirect to latest version if number of editions is one or less
	numberOfEditions := len(datasetEditions.Items)
	if numberOfEditions == 1 {
		latestVersionPath := helpers.DatasetVersionURL(datasetID, datasetEditions.Items[0].Edition, datasetEditions.Items[0].Links.LatestVersion.ID)
		log.Info(ctx, "only one edition, therefore redirecting to latest version", log.Data{"latestVersionPath": latestVersionPath})
		http.Redirect(w, req, latestVersionPath, http.StatusFound)
		return
	}

	// Fetch homepage content
	homepageContent, err := zc.GetHomepageContent(ctx, userAccessToken, collectionID, lang, homepagePath)
	if err != nil {
		log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{err}), log.Data{"homepage_content": err})
	}

	// Build page context
	basePage := rend.NewBasePageModel()
	// Update basePage common parameters
	mapper.UpdateBasePage(&basePage, datasetDetails, homepageContent, false, lang, req)

	bc, err := zc.GetBreadcrumb(ctx, userAccessToken, userAccessToken, collectionID, datasetDetails.Links.Taxonomy.HRef)
	if err != nil {
		log.Warn(ctx, "unable to get breadcrumb for dataset uri", log.FormatErrors([]error{err}), log.Data{"taxonomy_url": datasetDetails.Links.Taxonomy.HRef})
	}

	m := mapper.CreateEditionsList(ctx, basePage, req, datasetDetails, datasetEditions, datasetID, bc, apiRouterVersion)
	rend.BuildPage(w, m, "edition-list")
}
