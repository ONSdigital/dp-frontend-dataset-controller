package handlers

import (
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/dp-net/v2/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"net/http"
)

// VersionsList will load a list of versions for a filterable dataset
func VersionsList(dc DatasetClient, zc ZebedeeClient, rend RenderClient, cfg config.Config) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		versionsList(w, req, dc, zc, rend, collectionID, userAccessToken, lang)
	})
}

func versionsList(w http.ResponseWriter, req *http.Request, dc DatasetClient, zc ZebedeeClient, rend RenderClient, collectionID, userAccessToken, lang string) {
	vars := mux.Vars(req)
	datasetID := vars["datasetID"]
	edition := vars["edition"]
	ctx := req.Context()

	d, err := dc.Get(ctx, userAccessToken, "", collectionID, datasetID)
	if err != nil {
		setStatusCode(ctx, w, err)
		return
	}

	homepageContent, err := zc.GetHomepageContent(ctx, userAccessToken, collectionID, lang, homepagePath)
	if err != nil {
		log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{err}), log.Data{"homepage_content": err})
	}

	q := dataset.QueryParams{Offset: 0, Limit: 1000}
	versions, err := dc.GetVersions(ctx, userAccessToken, "", "", collectionID, datasetID, edition, &q)
	if err != nil {
		setStatusCode(ctx, w, err)
		return
	}

	e, err := dc.GetEdition(ctx, userAccessToken, "", collectionID, datasetID, edition)
	if err != nil {
		setStatusCode(ctx, w, err)
		return
	}

	basePage := rend.NewBasePageModel()
	m := mapper.CreateVersionsList(basePage, req, d, e, versions.Items, homepageContent.ServiceMessage, homepageContent.EmergencyBanner)
	rend.BuildPage(w, m, "version-list")
}
