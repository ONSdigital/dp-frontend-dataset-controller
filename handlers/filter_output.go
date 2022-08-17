package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/dp-net/v2/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// FilterOutput will load a filtered landing page
func FilterOutput(fc FilterClient, pc PopulationClient, dc DatasetClient, rend RenderClient, cfg config.Config, apiRouterVersion string) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		filterOutput(w, req, dc, fc, pc, rend, cfg, collectionID, lang, apiRouterVersion, userAccessToken)
	})
}

func filterOutput(w http.ResponseWriter, req *http.Request, dc DatasetClient, fc FilterClient, pc PopulationClient, rend RenderClient, cfg config.Config, collectionID, lang, apiRouterVersion, userAccessToken string) {
	const numOptsSummary = 1000
	var initialVersion dataset.Version
	var initialVersionReleaseDate string
	var form = req.URL.Query().Get("f")
	var format = req.URL.Query().Get("format")
	var isValidationError bool
	var datasetModel dataset.DatasetDetails
	var allVers dataset.VersionsList
	var ver dataset.Version
	var filterOutput filter.Model
	var dmErr, versErr, verErr, fErr error

	vars := mux.Vars(req)
	datasetID := vars["datasetID"]
	edition := vars["editionID"]
	version := vars["versionID"]
	filterOutputID := vars["filterOutputID"]
	ctx := req.Context()

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		datasetModel, dmErr = dc.Get(ctx, userAccessToken, "", collectionID, datasetID)
	}()

	go func() {
		defer wg.Done()
		q := dataset.QueryParams{Offset: 0, Limit: 1000}
		allVers, versErr = dc.GetVersions(ctx, userAccessToken, "", "", collectionID, datasetID, edition, &q)
	}()

	go func() {
		defer wg.Done()
		ver, verErr = dc.GetVersion(ctx, userAccessToken, "", "", collectionID, datasetID, edition, version)
		if ver.Version != 1 {
			initialVersion, verErr = dc.GetVersion(ctx, userAccessToken, "", "", collectionID, datasetID, edition, "1")
			initialVersionReleaseDate = initialVersion.ReleaseDate
		}
	}()

	go func() {
		defer wg.Done()
		filterOutput, fErr = fc.GetOutput(ctx, userAccessToken, "", "", collectionID, filterOutputID)
	}()

	wg.Wait()

	if dmErr != nil {
		log.Error(ctx, "failed to get dataset", dmErr, log.Data{"dataset": datasetID})
		setStatusCode(ctx, w, dmErr)
		return
	}
	if versErr != nil {
		log.Error(ctx, "failed to get dataset versions", versErr, log.Data{
			"dataset": datasetID,
			"edition": edition,
		})
		setStatusCode(ctx, w, versErr)
		return
	}
	if verErr != nil {
		log.Error(ctx, "failed to get dataset version", verErr, log.Data{
			"dataset": datasetID,
			"edition": edition,
			"version": version,
		})
		setStatusCode(ctx, w, verErr)
		return
	}
	if fErr != nil {
		log.Error(ctx, "failed to get filter output", fErr, log.Data{"filter_output_id": filterOutputID})
		setStatusCode(ctx, w, fErr)
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

	var hasNoAreaOptions bool
	getAreaOptions := func(dim filter.ModelDimension) ([]string, error) {
		q := filter.QueryParams{
			Limit: 500,
		}
		opts, _, err := fc.GetDimensionOptions(ctx, userAccessToken, "", collectionID, filterOutput.FilterID, dim.Name, &q)
		if err != nil {
			return nil, fmt.Errorf("failed to get options for dimension: %w", err)
		}

		var options []string
		if opts.TotalCount == 0 {
			areas, err := pc.GetAreas(ctx, population.GetAreasInput{
				UserAuthToken: userAccessToken,
				DatasetID:     filterOutput.PopulationType,
				AreaTypeID:    dim.ID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get dimension areas: %w", err)
			}

			for _, area := range areas.Areas {
				options = append(options, area.Label)
			}

			hasNoAreaOptions = true
			return options, nil
		}

		var wg sync.WaitGroup
		var areaErr error
		var areas population.GetAreasResponse
		for _, opt := range opts.Items {
			wg.Add(1)
			go func(opt filter.DimensionOption) {
				defer wg.Done()
				// TODO: Temporary fix until GetArea endpoint is created
				areas, areaErr = pc.GetAreas(ctx, population.GetAreasInput{
					UserAuthToken: userAccessToken,
					DatasetID:     filterOutput.PopulationType,
					AreaTypeID:    dim.ID,
					Text:          opt.Option,
				})

				for _, area := range areas.Areas {
					if area.ID == opt.Option {
						options = append(options, area.Label)
						break
					}
				}
			}(opt)
		}
		wg.Wait()

		if areaErr != nil {
			return nil, fmt.Errorf("failed to get dimension areas: %w", areaErr)
		}

		return options, nil
	}

	getOptions := func(dim filter.ModelDimension) ([]string, error) {
		if dim.IsAreaType != nil && *dim.IsAreaType {
			return getAreaOptions(dim)
		}

		return getDimensionOptions(dim)
	}

	for i, dim := range filterOutput.Dimensions {
		wg.Add(1)
		go func(i int, dim filter.ModelDimension) {
			defer wg.Done()
			options, err := getOptions(dim)
			if err != nil {
				log.Error(ctx, "failed to get options for dimension", err, log.Data{"dimension_name": dim.Name})
				setStatusCode(ctx, w, err)
				return
			}

			dim.Options = options
			filterOutput.Dimensions[i] = dim
		}(i, dim)
	}
	wg.Wait()

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

	showAll := req.URL.Query()[queryStrKey]
	basePage := rend.NewBasePageModel()
	m := mapper.CreateCensusDatasetLandingPage(ctx, req, basePage, datasetModel, ver, []dataset.Options{}, initialVersionReleaseDate, hasOtherVersions, allVers.Items, latestVersionNumber, latestVersionURL, lang, showAll, numOptsSummary, isValidationError, true, hasNoAreaOptions, filterOutput)
	rend.BuildPage(w, m, "census-landing")
}
