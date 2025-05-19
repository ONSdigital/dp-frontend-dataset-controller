package handlers

import (
	"net/http"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/dp-net/v3/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// EditionsList will load a list of editions for a filterable dataset
func EditionsList(dc APIClientsGoDatasetClient, zc ZebedeeClient, rend RenderClient, cfg config.Config, apiRouterVersion string) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		editionsList(w, req, dc, zc, rend, collectionID, lang, apiRouterVersion, userAccessToken)
	})
}

func editionsList(w http.ResponseWriter, req *http.Request, dc APIClientsGoDatasetClient, zc ZebedeeClient, rend RenderClient, collectionID, lang, apiRouterVersion, userAccessToken string) {
	vars := mux.Vars(req)
	datasetID := vars["datasetID"]
	ctx := req.Context()

	const (
		DatasetTypeStatic = "static"
	)

	datasetDetails, err := dc.Get(ctx, userAccessToken, "", collectionID, datasetID)
	if err != nil {
		setStatusCode(ctx, w, err)
		return
	}

	datasetEditions, err := dc.GetEditions(ctx, userAccessToken, "", collectionID, datasetID)
	if err != nil {
		if err, ok := err.(ClientError); ok {
			if err.Code() != http.StatusNotFound {
				setStatusCode(ctx, w, err)
				return
			}
		}
	}

	homepageContent, err := zc.GetHomepageContent(ctx, userAccessToken, collectionID, lang, homepagePath)
	if err != nil {
		log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{err}), log.Data{"homepage_content": err})
	}

	bc, err := zc.GetBreadcrumb(ctx, userAccessToken, userAccessToken, collectionID, datasetDetails.Links.Taxonomy.URL)
	if err != nil {
		log.Warn(ctx, "unable to get breadcrumb for dataset uri", log.FormatErrors([]error{err}), log.Data{"taxonomy_url": datasetDetails.Links.Taxonomy.URL})
	}

	numberOfEditions := len(datasetEditions)
	if numberOfEditions == 1 {
		latestVersionPath := helpers.DatasetVersionURL(datasetID, datasetEditions[0].Edition, datasetEditions[0].Links.LatestVersion.ID)
		log.Info(ctx, "only one edition, therefore redirecting to latest version", log.Data{"latestVersionPath": latestVersionPath})
		http.Redirect(w, req, latestVersionPath, http.StatusFound)
	}

	basePage := rend.NewBasePageModel()

	m := mapper.CreateEditionsList(ctx, basePage, req, datasetDetails, datasetEditions, datasetID, bc, lang, apiRouterVersion, homepageContent.ServiceMessage, homepageContent.EmergencyBanner)
	
	if (datasetDetails.Type == DatasetTypeStatic){
		m.FeatureFlags.SixteensVersion = "" // When unset, dp-design-system is used 
		m.Type =  DatasetTypeStatic
		rend.BuildPage(w, m, "edition-list-static")
	} else {
		rend.BuildPage(w, m, "edition-list")
	}
}
