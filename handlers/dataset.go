package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/ONSdigital/dp-net/handlers"

	"github.com/pkg/errors"

	"github.com/ONSdigital/dp-api-clients-go/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/log.go/log"
)

// DatasetPage will load a zebedee dataset page
func DatasetPage(zc ZebedeeClient, dc DatasetClient, rend RenderClient, cfg config.Config, apiRouterVersion string) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		datasetPage(w, req, zc, dc, rend, cfg, collectionID, lang, userAccessToken, apiRouterVersion)
	})
}

func datasetPage(w http.ResponseWriter, req *http.Request, zc ZebedeeClient, dc DatasetClient, rend RenderClient, cfg config.Config, collectionID, lang, userAccessToken string, apiRouterVersion string) {
	path := req.URL.Path
	ctx := req.Context()

	// Since MatchString will only error if the regex is invalid, and the regex is
	// constant, don't capture the error
	if ok, _ := regexp.MatchString(dataEndpoint, path); ok {
		b, err := zc.Get(ctx, userAccessToken, "/data?uri="+path)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}
		w.Write(b)
		return
	}

	ds, err := zc.GetDataset(ctx, userAccessToken, collectionID, lang, path)
	if err != nil {
		setStatusCode(req, w, errors.Wrap(err, "zebedee client get dataset returned an error"))
		return
	}

	bc, err := zc.GetBreadcrumb(ctx, userAccessToken, collectionID, lang, ds.URI)
	if err != nil {
		log.Event(ctx, "unable to get breadcrumb for dataset page uri", log.WARN, log.Error(err))
		setStatusCode(req, w, err)
		return
	}

	if len(bc) < 2 {
		log.Event(ctx, "invalid breadcrumb length for dataset page uri")
		setStatusCode(req, w, fmt.Errorf("invalid breadcrumb length"))
		return
	}

	parentPath := bc[len(bc)-1].URI

	dlp, err := zc.GetDatasetLandingPage(ctx, userAccessToken, collectionID, lang, parentPath)
	if err != nil {
		log.Event(ctx, "unable to get dataset page parent", log.WARN, log.Error(err))
		setStatusCode(req, w, err)
		return
	}

	var versions []zebedee.Dataset
	for _, ver := range ds.Versions {
		version, err := zc.GetDataset(ctx, userAccessToken, collectionID, lang, ver.URI)
		if err != nil {
			setStatusCode(req, w, errors.Wrap(err, "zebedee client get previous dataset versions returned an error"))
			return
		}
		versions = append(versions, version)
	}

	m := mapper.CreateDatasetPage(ctx, req, ds, dlp, bc, versions, lang, apiRouterVersion)
	templateJSON, err := json.Marshal(m)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	templateHTML, err := rend.Do("dataset-page", templateJSON)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	w.Write(templateHTML)
	return
}
