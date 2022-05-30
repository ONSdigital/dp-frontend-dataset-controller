package handlers

import (
	"errors"
	"fmt"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-net/v2/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"strings"
)

// CreateFilterID controls the creating of a filter idea when a new user journey is requested
func CreateFilterID(c FilterClient, dc DatasetClient) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		datasetID := vars["datasetID"]
		edition := vars["editionID"]
		version := vars["versionID"]
		ctx := req.Context()

		dimensions, err := dc.GetVersionDimensions(ctx, userAccessToken, "", collectionID, datasetID, edition, version)
		if err != nil {
			setStatusCode(ctx, w, err)
			return
		}

		var names []string
		for _, dim := range dimensions.Items {
			// we are only interested in the totalCount, limit=0 will always return an empty list of items and the total count
			q := dataset.QueryParams{Offset: 0, Limit: 0}
			opts, err := dc.GetOptions(ctx, userAccessToken, "", collectionID, datasetID, edition, version, dim.Name, &q)
			if err != nil {
				setStatusCode(ctx, w, err)
				return
			}

			if opts.TotalCount > 1 { // If there is only one option then it can't be filterable so don't add to filter api
				names = append(names, dim.Name)
			}
		}
		fid, _, err := c.CreateBlueprint(ctx, userAccessToken, "", "", collectionID, datasetID, edition, version, names)
		if err != nil {
			setStatusCode(ctx, w, err)
			return
		}

		log.Info(ctx, "created filter id", log.Data{"filter_id": fid})
		http.Redirect(w, req, "/filters/"+fid+"/dimensions", http.StatusMovedPermanently)
	})
}

// CreateFilterFlexID creates a new filter ID for filter flex journeys
func CreateFilterFlexID(fc FilterClient, dc DatasetClient, cfg config.Config) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		datasetID := vars["datasetID"]
		edition := vars["editionID"]
		version := vars["versionID"]
		ctx := req.Context()

		if !cfg.EnableCensusPages {
			err := errors.New("not implemented")
			log.Error(ctx, "route not implemented", err)
			setStatusCode(ctx, w, err)
			return
		}

		if err := req.ParseForm(); err != nil {
			log.Error(ctx, "unable to parse request form", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ver, err := dc.GetVersion(ctx, userAccessToken, "", "", collectionID, datasetID, edition, version)
		if err != nil {
			setStatusCode(ctx, w, err)
			return
		}

		dims := []filter.ModelDimension{}
		for _, verDim := range ver.Dimensions {
			var dim = filter.ModelDimension{}
			dim.Name = verDim.Name
			dim.URI = verDim.URL
			dim.IsAreaType = verDim.IsAreaType
			q := dataset.QueryParams{Offset: 0, Limit: 1000}
			opts, err := dc.GetOptions(ctx, userAccessToken, "", collectionID, datasetID, edition, version, dim.Name, &q)
			if err != nil {
				setStatusCode(ctx, w, err)
				return
			}
			var labels, options []string
			for _, opt := range opts.Items {
				labels = append(labels, opt.Label)
				options = append(options, opt.Option)
			}
			dim.Options = options
			dim.Values = labels
			dims = append(dims, dim)
		}

		datasetModel, err := dc.Get(ctx, userAccessToken, "", collectionID, datasetID)
		if err != nil {
			setStatusCode(ctx, w, err)
			return
		}

		popType := datasetModel.IsBasedOn.ID
		fid, _, err := fc.CreateFlexibleBlueprint(ctx, userAccessToken, "", "", collectionID, datasetID, edition, version, dims, popType)
		if err != nil {
			setStatusCode(ctx, w, err)
			return
		}

		filterPath := fmt.Sprintf("/filters/%s/dimensions", fid)
		dimensionName := req.FormValue("dimension")
		if dimensionName != "" {
			filterPath += fmt.Sprintf("/%s", strings.ToLower(url.QueryEscape(dimensionName)))
		}

		log.Info(ctx, "created filter id", log.Data{"filter_id": fid})
		http.Redirect(w, req, filterPath, http.StatusMovedPermanently)
	})
}
