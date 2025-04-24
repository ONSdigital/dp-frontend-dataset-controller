package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-net/v2/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"

	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
)

// CreateFilterID controls the creating of a filter idea when a new user journey is requested
func CreateFilterID(c FilterClient, dc ApiClientsGoDatasetClient) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		datasetID := vars["datasetID"]
		edition := vars["editionID"]
		version := vars["versionID"]
		ctx := req.Context()

		headers := dpDatasetApiSdk.Headers{
			CollectionID:    collectionID,
			UserAccessToken: userAccessToken,
		}

		dimensions, err := dc.GetVersionDimensions(ctx, headers, datasetID, edition, version)
		if err != nil {
			setStatusCode(ctx, w, err)
			return
		}

		var names []string
		for i := range dimensions.Items {
			dimension := &dimensions.Items[i]

			// we are only interested in the totalCount, limit=0 will always return an empty list of items and the total count
			q := dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 0}
			opts, err := dc.GetVersionDimensionOptions(ctx, headers, datasetID, edition, version, dimension.Name, &q)
			if err != nil {
				setStatusCode(ctx, w, err)
				return
			}

			if len(opts.Items) > 1 { // If there is only one option then it can't be filterable so don't add to filter api
				names = append(names, dimension.Name)
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
func CreateFilterFlexID(fc FilterClient, dc ApiClientsGoDatasetClient) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		datasetID := vars["datasetID"]
		edition := vars["editionID"]
		version := vars["versionID"]
		ctx := req.Context()

		headers := dpDatasetApiSdk.Headers{
			CollectionID:    collectionID,
			UserAccessToken: userAccessToken,
		}

		if err := req.ParseForm(); err != nil {
			log.Error(ctx, "unable to parse request form", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ver, err := dc.GetVersion(ctx, headers, datasetID, edition, version)
		if err != nil {
			setStatusCode(ctx, w, err)
			return
		}

		dims := []filter.ModelDimension{}
		for i := range ver.Dimensions {
			versionDimension := &ver.Dimensions[i]

			var dim = filter.ModelDimension{}
			dim.Name = versionDimension.Name
			dim.URI = versionDimension.HRef
			dim.IsAreaType = versionDimension.IsAreaType
			dims = append(dims, dim)
		}

		datasetModel, err := dc.GetDataset(ctx, headers, datasetID)
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
			log.Error(ctx, "unable to get filter output", err)
			setStatusCode(ctx, w, err)
			return
		}

		dims := []filter.ModelDimension{}
		for i := range fo.Dimensions {
			versionDimension := &fo.Dimensions[i]

			var dim = filter.ModelDimension{}
			dim.Name = versionDimension.Name
			dim.URI = versionDimension.URI
			dim.IsAreaType = versionDimension.IsAreaType
			dim.Options = versionDimension.Options
			dim.FilterByParent = versionDimension.FilterByParent
			dim.QualityStatementText = versionDimension.QualityStatementText
			dim.QualitySummaryURL = versionDimension.QualitySummaryURL
			dims = append(dims, dim)
		}

		fid := ""
		isCustom := helpers.IsBoolPtr(fo.Custom)
		if isCustom {
			fid, _, err = fc.CreateFlexibleBlueprintCustom(ctx, userAccessToken, "", "", filter.CreateFlexBlueprintCustomRequest{
				Dataset:        fo.Dataset,
				Dimensions:     dims,
				PopulationType: fo.PopulationType,
				CollectionID:   collectionID,
			})
		} else {
			fid, _, err = fc.CreateFlexibleBlueprint(ctx, userAccessToken, "", "", collectionID, fo.Dataset.DatasetID, fo.Dataset.Edition, strconv.Itoa(fo.Dataset.Version), dims, fo.PopulationType)
		}
		if err != nil {
			log.Error(ctx, "unable to create new filter", err)
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
		switch dimensionName {
		case "coverage":
			filterPath += fmt.Sprintf("/geography/%s", dimensionName)
		case "change":
			filterPath += "/change"
		default:
			filterPath += fmt.Sprintf("/%s", dimensionName)
		}
	}
	return filterPath
}
