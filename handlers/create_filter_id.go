package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-net/v2/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
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
func CreateFilterFlexID(fc FilterClient, dc DatasetClient) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		datasetID := vars["datasetID"]
		edition := vars["editionID"]
		version := vars["versionID"]
		ctx := req.Context()

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

		dimensionName := url.QueryEscape(req.FormValue("dimension"))
		filterPath := createFilterPath(fid, dimensionName)

		log.Info(ctx, "created filter id", log.Data{"filter_id": fid})
		http.Redirect(w, req, filterPath, http.StatusMovedPermanently)
	})
}

// CreateFilterFlexIDFromOutput creates a new filter ID for filter flex journeys from the user's filter output
func CreateFilterFlexIDFromOutput(fc FilterClient) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		filterOutputID := vars["filterOutputID"]
		ctx := req.Context()

		if err := req.ParseForm(); err != nil {
			log.Error(ctx, "unable to parse request form", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fo, err := fc.GetOutput(ctx, userAccessToken, "", "", collectionID, filterOutputID)
		if err != nil {
			setStatusCode(ctx, w, err)
			return
		}

		dims := []filter.ModelDimension{}
		for _, verDim := range fo.Dimensions {
			var dim = filter.ModelDimension{}
			dim.Name = verDim.Name
			dim.URI = verDim.URI
			dim.IsAreaType = verDim.IsAreaType
			dim.Options = verDim.Options
			dim.FilterByParent = verDim.FilterByParent
			dims = append(dims, dim)
		}

		fid, _, err := fc.CreateFlexibleBlueprint(ctx, userAccessToken, "", "", collectionID, fo.Dataset.DatasetID, fo.Dataset.Edition, strconv.Itoa(fo.Dataset.Version), dims, fo.PopulationType)
		if err != nil {
			setStatusCode(ctx, w, err)
			return
		}

		dimensionName := url.QueryEscape(req.FormValue("dimension"))
		filterPath := createFilterPath(fid, dimensionName)

		log.Info(ctx, "created filter id", log.Data{"filter_id": fid})
		http.Redirect(w, req, filterPath, http.StatusMovedPermanently)
	})
}

func createFilterPath(fid, dimensionName string) string {
	filterPath := fmt.Sprintf("/filters/%s/dimensions", fid)
	if dimensionName != "" {
		if dimensionName == "coverage" {
			filterPath += fmt.Sprintf("/geography/%s", dimensionName)
		} else {
			filterPath += fmt.Sprintf("/%s", dimensionName)
		}
	}
	return filterPath
}
