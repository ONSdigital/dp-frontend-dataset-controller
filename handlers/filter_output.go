package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/dimension"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/dp-net/v2/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// FilterOutput will load a filtered landing page
func FilterOutput(fc FilterClient, dimsc DimensionClient, dc DatasetClient, rend RenderClient, cfg config.Config, apiRouterVersion string) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		filterOutput(w, req, dc, fc, dimsc, rend, cfg, collectionID, lang, apiRouterVersion, userAccessToken)
	})
}

func filterOutput(w http.ResponseWriter, req *http.Request, dc DatasetClient, fc FilterClient, dimsc DimensionClient, rend RenderClient, cfg config.Config, collectionID, lang, apiRouterVersion, userAccessToken string) {
	const numOptsSummary = 1000
	var initialVersion dataset.Version
	var initialVersionReleaseDate string
	var form = req.URL.Query().Get("f")
	var format = req.URL.Query().Get("format")
	var isValidationError bool

	vars := mux.Vars(req)
	datasetID := vars["datasetID"]
	edition := vars["editionID"]
	version := vars["versionID"]
	filterOutputID := vars["filterOutputID"]
	ctx := req.Context()

	datasetModel, err := dc.Get(ctx, userAccessToken, "", collectionID, datasetID)
	if err != nil {
		log.Error(ctx, "failed to get dataset", err, log.Data{"dataset": datasetID})
		setStatusCode(ctx, w, err)
		return
	}

	q := dataset.QueryParams{Offset: 0, Limit: 1000}
	allVers, err := dc.GetVersions(ctx, userAccessToken, "", "", collectionID, datasetID, edition, &q)
	if err != nil {
		log.Error(ctx, "failed to get dataset versions", err, log.Data{
			"dataset": datasetID,
			"edition": edition,
		})
		setStatusCode(ctx, w, err)
		return
	}

	// TODO: Inherited from census landing, refactor to check in mapper
	var hasOtherVersions bool
	if len(allVers.Items) > 1 {
		hasOtherVersions = true
	}

	latestVersionNumber := 1
	for _, singleVersion := range allVers.Items {
		if singleVersion.Version > latestVersionNumber {
			latestVersionNumber = singleVersion.Version
		}
	}

	latestVersionURL := helpers.DatasetVersionUrl(datasetID, edition, strconv.Itoa(latestVersionNumber))

	ver, err := dc.GetVersion(ctx, userAccessToken, "", "", collectionID, datasetID, edition, version)
	if err != nil {
		log.Error(ctx, "failed to get dataset version", err, log.Data{
			"dataset": datasetID,
			"edition": edition,
			"version": version,
		})
		setStatusCode(ctx, w, err)
		return
	}

	// TODO: inherited from censusLanding handler refactor to get initial release date in the mapper
	if ver.Version != 1 {
		initialVersion, err = dc.GetVersion(ctx, userAccessToken, "", "", collectionID, datasetID, edition, "1")
		if err != nil {
			setStatusCode(ctx, w, err)
			return
		}
		initialVersionReleaseDate = initialVersion.ReleaseDate
	}

	filterOutput := filter.Model{}

	getDimensionOptions := func(dim filter.ModelDimension) ([]string, error) {
		q := dataset.QueryParams{Offset: 0, Limit: 1000}
		opts, err := dc.GetOptions(ctx, userAccessToken, "", collectionID, datasetModel.ID, edition, strconv.Itoa(ver.Version), dim.Name, &q)
		if err != nil {
			return nil, fmt.Errorf("failed to get options for dimension: %w", err)
		}

		var options []string
		for _, opt := range opts.Items {
			options = append(options, opt.Label)
		}

		return options, nil
	}

	getAreaOptions := func(dim filter.ModelDimension) ([]string, error) {
		areas, err := dimsc.GetAreas(ctx, dimension.GetAreasInput{
			UserAuthToken: userAccessToken,
			DatasetID:     filterOutput.PopulationType,
			AreaTypeID:    dim.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get dimension areas: %w", err)
		}

		var options []string
		for _, area := range areas.Areas {
			options = append(options, area.Label)
		}

		return options, nil
	}

	getOptions := func(dim filter.ModelDimension) ([]string, error) {
		if dim.IsAreaType != nil && *dim.IsAreaType {
			return getAreaOptions(dim)
		}

		return getDimensionOptions(dim)
	}

	filterOutput, err = fc.GetOutput(ctx, userAccessToken, "", "", collectionID, filterOutputID)
	if err != nil {
		log.Error(ctx, "failed to get filter output", err, log.Data{"filter_output_id": filterOutputID})
		setStatusCode(ctx, w, err)
		return
	}

	for i, dim := range filterOutput.Dimensions {
		options, err := getOptions(dim)
		if err != nil {
			log.Error(ctx, "failed to get options for dimension", err, log.Data{"dimension_name": dim.Name})
			setStatusCode(ctx, w, err)
			return
		}

		dim.Options = options
		filterOutput.Dimensions[i] = dim

	}

	if filterOutput.Downloads == nil {
		log.Warn(ctx, "filter output downloads are nil", log.Data{"filter_output_id": filterOutputID})
		filterOutput.Downloads = make(map[string]filter.Download)
	}

	if form == "get-data" && format == "" {
		isValidationError = true
	}

	if form == "get-data" && format != "" {
		for ext, download := range filterOutput.Downloads {
			if strings.EqualFold(ext, format) {
				http.Redirect(w, req, download.URL, http.StatusFound)
			}
		}
	}

	basePage := rend.NewBasePageModel()
	m := mapper.CreateCensusDatasetLandingPage(ctx, req, basePage, datasetModel, ver, []dataset.Options{}, initialVersionReleaseDate, hasOtherVersions, allVers.Items, latestVersionNumber, latestVersionURL, lang, numOptsSummary, isValidationError, true, filterOutput)
	rend.BuildPage(w, m, "census-landing")
}
