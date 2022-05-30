package handlers

import (
	"fmt"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/dp-net/v2/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
	"net/http"
)

func datasetPage(w http.ResponseWriter, req *http.Request, zc ZebedeeClient, rend RenderClient, fac FilesAPIClient, collectionID, lang, userAccessToken string) {
	path := req.URL.Path
	ctx := req.Context()

	if handleRequestForZebedeeJsonData(w, req, zc, path, ctx, userAccessToken) {
		return
	}

	ds, err := zc.GetDataset(ctx, userAccessToken, collectionID, lang, path)
	if err != nil {
		setStatusCode(ctx, w, errors.Wrap(err, "zebedee client get dataset returned an error"))
		return
	}

	bc, err := zc.GetBreadcrumb(ctx, userAccessToken, collectionID, lang, ds.URI)
	if err != nil {
		setStatusCode(ctx, w, err)
		return
	}

	if len(bc) < 2 {
		setStatusCode(ctx, w, fmt.Errorf("invalid breadcrumb length"))
		return
	}

	parentPath := bc[len(bc)-1].URI

	homepageContent, err := zc.GetHomepageContent(ctx, userAccessToken, collectionID, lang, homepagePath)
	if err != nil {
		log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{err}), log.Data{"homepage_content": err})
	}

	dlp, err := zc.GetDatasetLandingPage(ctx, userAccessToken, collectionID, lang, parentPath)
	if err != nil {
		setStatusCode(ctx, w, err)
		return
	}

	var versions []zebedee.Dataset
	for _, ver := range ds.Versions {
		version, err := zc.GetDataset(ctx, userAccessToken, collectionID, lang, ver.URI)
		if err != nil {
			setStatusCode(ctx, w, errors.Wrap(err, "zebedee client get previous dataset versions returned an error"))
			return
		}

		version, err = addFileSizesToDataset(ctx, fac, version, userAccessToken)
		if err != nil {
			log.Error(ctx, "failed to get file size from files API", err, log.Data{"version": version})
		}

		versions = append(versions, version)
	}

	basePage := rend.NewBasePageModel()
	m := mapper.CreateDatasetPage(basePage, ctx, req, ds, dlp, bc, versions, lang, homepageContent.ServiceMessage, homepageContent.EmergencyBanner)

	rend.BuildPage(w, m, "dataset")
}

// DataSet will load a legacy dataset page
func DatasetPage(zc ZebedeeClient, rend RenderClient, fac FilesAPIClient) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		datasetPage(w, req, zc, rend, fac, collectionID, lang, userAccessToken)
	})
}
