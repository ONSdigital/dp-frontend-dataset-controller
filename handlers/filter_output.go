package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-net/v2/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// FilterOutput will load a filtered landing page
func FilterOutput(zc ZebedeeClient, fc FilterClient, pc PopulationClient, dc DatasetClient, rend RenderClient, cfg config.Config, apiRouterVersion string) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		filterOutput(w, req, zc, dc, fc, pc, rend, cfg, collectionID, lang, apiRouterVersion, userAccessToken)
	})
}

func filterOutput(w http.ResponseWriter, req *http.Request, zc ZebedeeClient, dc DatasetClient, fc FilterClient, pc PopulationClient, rend RenderClient, cfg config.Config, collectionID, lang, apiRouterVersion, userAccessToken string) {
	const numOptsSummary = 1000
	var form = req.URL.Query().Get("f")
	var format = req.URL.Query().Get("format")
	var isValidationError bool
	var datasetModel dataset.DatasetDetails
	var allVers dataset.VersionsList
	var ver dpDatasetApiModels.Version
	var filterOutput filter.Model
	var dimDescriptions population.GetDimensionsResponse
	var sdc *cantabular.GetBlockedAreaCountResult
	var areaTypeID, parent string
	var dimCategories population.GetDimensionCategoriesResponse
	var pop population.GetPopulationTypeResponse
	var dimIds, nonAreaDimIds, areaOpts []string
	var dmErr, versErr, verErr, fErr, dErr, sErr, dcErr, pErr error
	var wg sync.WaitGroup
	var isSpinner = req.URL.Query().Get("spinner") == "true"

	vars := mux.Vars(req)
	ctx := req.Context()
	datasetID := vars["datasetID"]
	edition := vars["editionID"]
	version := vars["versionID"]
	filterOutputID := vars["filterOutputID"]

	wg.Add(1)
	go func() {
		defer wg.Done()
		filterOutput, fErr = fc.GetOutput(ctx, userAccessToken, "", "", collectionID, filterOutputID)
		for _, dim := range filterOutput.Dimensions {
			dimIds = append(dimIds, dim.ID)
			if !helpers.IsBoolPtr(dim.IsAreaType) {
				nonAreaDimIds = append(nonAreaDimIds, dim.ID)
			}
		}
	}()

	// If url doesn't include datasetID, edition and version wait and retrieve from the filter object
	if datasetID == "" {
		wg.Wait()
		if logError(ctx, w, fErr, "failed to get filter output", log.Data{"filterOutputID": filterOutputID}) {
			return
		}

		datasetID = filterOutput.Dataset.DatasetID
		edition = filterOutput.Dataset.Edition
		version = strconv.Itoa(filterOutput.Dataset.Version)
	}

	wg.Add(3)
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
	}()

	wg.Wait()

	if logError(ctx, w, dmErr, "failed to get dataset", log.Data{"dataset": datasetID}) {
		return
	}
	if logError(ctx, w, versErr, "failed to get dataset versions", log.Data{"dataset": datasetID, "edition": edition}) {
		return
	}
	if logError(ctx, w, verErr, "failed to get dataset version", log.Data{"dataset": datasetID, "edition": edition, "version": version}) {
		return
	}
	if logError(ctx, w, fErr, "failed to get filter output", log.Data{"filterOutputID": filterOutputID}) {
		return
	}

	wg.Add(3)
	go func() {
		defer wg.Done()
		dimDescriptions, dErr = pc.GetDimensionsDescription(ctx, population.GetDimensionsDescriptionInput{
			AuthTokens: population.AuthTokens{
				UserAuthToken: userAccessToken,
			},
			PopulationType: filterOutput.PopulationType,
			DimensionIDs:   dimIds,
		})
		if dErr != nil {
			log.Error(ctx, "failed to get dimension descriptions", dErr, log.Data{
				"population_type": filterOutput.PopulationType,
				"dimension_ids":   dimIds,
			})
			setStatusCode(ctx, w, dErr)
			return
		}
	}()

	go func() {
		defer wg.Done()
		pop, pErr = pc.GetPopulationType(ctx, population.GetPopulationTypeInput{
			AuthTokens: population.AuthTokens{
				UserAuthToken: userAccessToken,
			},
			PopulationType: filterOutput.PopulationType,
		})
	}()

	go func() {
		defer wg.Done()
		if len(nonAreaDimIds) > 0 {
			dimCategories, dcErr = pc.GetDimensionCategories(ctx, population.GetDimensionCategoryInput{
				AuthTokens: population.AuthTokens{
					UserAuthToken: userAccessToken,
				},
				PaginationParams: population.PaginationParams{
					Limit:  1000,
					Offset: 0,
				},
				PopulationType: filterOutput.PopulationType,
				Dimensions:     nonAreaDimIds,
			})
		} else {
			dimCategories = population.GetDimensionCategoriesResponse{}
		}
	}()

	wg.Wait()

	if dErr != nil {
		log.Error(ctx, "failed to get dimension descriptions", dErr, log.Data{
			"population_type": filterOutput.PopulationType,
			"dimension_ids":   dimIds,
		})
		setStatusCode(ctx, w, dErr)
		return
	}

	if pErr != nil {
		log.Error(ctx, "failed to get population type", pErr, log.Data{
			"filterOutputID":  filterOutputID,
			"population_type": filterOutput.PopulationType,
		})
		setStatusCode(ctx, w, pErr)
		return
	}

	if dcErr != nil {
		log.Error(ctx, "failed to get dimension categories", dErr, log.Data{
			"population_type": filterOutput.PopulationType,
			"dimension_ids":   dimIds,
		})
		setStatusCode(ctx, w, dcErr)
		return
	}
	dimensionCategoriesMap := mapDimensionCategories(dimCategories)

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

	latestVersionURL := helpers.DatasetVersionURL(datasetID, edition, strconv.Itoa(latestVersionNumber))

	getDimensionOptions := func(dim filter.ModelDimension) ([]string, int, error) {
		dimensionCategory := dimensionCategoriesMap[dim.ID]

		var options []string
		for _, opt := range sortCategoriesByID(dimensionCategory.Categories) {
			options = append(options, opt.Label)
		}

		return options, len(options), nil
	}

	var hasNoAreaOptions bool
	getAreaOptions := func(dim filter.ModelDimension) ([]string, int, error) {
		q := filter.QueryParams{
			Limit: 500,
		}
		opts, _, err := fc.GetDimensionOptions(ctx, userAccessToken, "", collectionID, filterOutput.FilterID, dim.Name, &q)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get options for dimension: %w", err)
		}

		var options []string
		if opts.TotalCount == 0 {
			// TODO: GetAreas has been updated with a Text parameter - this needs identifying and updating
			areas, err := pc.GetAreas(ctx, population.GetAreasInput{
				AuthTokens: population.AuthTokens{
					UserAuthToken: userAccessToken,
				},
				PaginationParams: population.PaginationParams{
					Limit:  opts.Limit,
					Offset: opts.Offset,
				},
				PopulationType: filterOutput.PopulationType,
				AreaTypeID:     dim.ID,
			})

			if err != nil {
				return nil, 0, fmt.Errorf("failed to get dimension areas: %w", err)
			}

			for _, area := range areas.Areas {
				options = append(options, area.Label)
			}

			hasNoAreaOptions = true
			return options, areas.TotalCount, nil
		}

		var wg sync.WaitGroup
		areaErrs := make([]error, len(opts.Items))
		optsIDs := []string{}
		totalCount := opts.TotalCount
		for i, opt := range opts.Items {
			wg.Add(1)
			go func(opt filter.DimensionOption, i int) {
				defer wg.Done()
				optsIDs = append(optsIDs, opt.Option)
				var areaTypeID string
				if dim.FilterByParent != "" {
					areaTypeID = dim.FilterByParent
				} else {
					areaTypeID = dim.ID
				}

				area, err := pc.GetArea(ctx, population.GetAreaInput{
					AuthTokens: population.AuthTokens{
						UserAuthToken: userAccessToken,
					},
					PopulationType: filterOutput.PopulationType,
					AreaType:       areaTypeID,
					Area:           opt.Option,
				})

				if err != nil {
					areaErrs[i] = err
				}

				options = append(options, area.Area.Label)

			}(opt, i)
		}
		wg.Wait()

		var hasErrs bool
		for _, err := range areaErrs {
			if err != nil {
				log.Error(ctx, "failed to get areas for options", err, log.Data{
					"dimension_name": dim.Name,
					"options":        opts.Items,
				})
				hasErrs = true
			}
		}

		if hasErrs {
			return nil, 0, fmt.Errorf("failed to get dimension areas")
		}
		areaOpts = optsIDs

		return options, totalCount, nil
	}

	getOptions := func(dim filter.ModelDimension) ([]string, int, error) {
		if dim.IsAreaType != nil && *dim.IsAreaType {
			areaTypeID = dim.ID
			parent = dim.FilterByParent
			return getAreaOptions(dim)
		}

		return getDimensionOptions(dim)
	}

	getDimensionCategorisations := func(populationType string, dimension string) (int, error) {
		cats, err := pc.GetCategorisations(ctx, population.GetCategorisationsInput{
			AuthTokens: population.AuthTokens{
				UserAuthToken: userAccessToken,
			},
			PaginationParams: population.PaginationParams{
				Limit: 1000,
			},
			PopulationType: populationType,
			Dimension:      dimension,
		})
		return cats.PaginationResponse.TotalCount, err
	}

	var fDims []model.FilterDimension
	for i := len(filterOutput.Dimensions) - 1; i >= 0; i-- {
		options, count, err := getOptions(filterOutput.Dimensions[i])
		if err != nil {
			log.Error(ctx, "failed to get options for dimension", err, log.Data{"dimension_name": filterOutput.Dimensions[i].Name})
			setStatusCode(ctx, w, err)
			return
		}

		categorisationCount := 0
		if !helpers.IsBoolPtr(filterOutput.Dimensions[i].IsAreaType) {
			categorisationCount, err = getDimensionCategorisations(filterOutput.PopulationType, filterOutput.Dimensions[i].Name)
		}
		if err != nil {
			log.Error(ctx, "failed to get categorisations for dimension", err, log.Data{"dimension_name": filterOutput.Dimensions[i].Name})
			setStatusCode(ctx, w, err)
			return
		}

		filterOutput.Dimensions[i].Options = options
		fDims = append(fDims, model.FilterDimension{
			ModelDimension:      filterOutput.Dimensions[i],
			OptionsCount:        count,
			CategorisationCount: categorisationCount,
		})
	}

	if strings.Contains(datasetModel.Type, "multivariate") {
		sort.Slice(dimIds, func(i, j int) bool {
			return dimIds[i] == areaTypeID || dimIds[i] == parent
		})

		if parent != "" {
			areaTypeID = parent
		}

		// set default coverage
		if len(areaOpts) == 0 {
			areaOpts = []string{"K04000001"}
			areaTypeID = "nat"
		}

		if isSpinner {
			sdc = &cantabular.GetBlockedAreaCountResult{}
		} else {
			sdc, sErr = pc.GetBlockedAreaCount(ctx, population.GetBlockedAreaCountInput{
				AuthTokens: population.AuthTokens{
					UserAuthToken: userAccessToken,
				},
				PopulationType: filterOutput.PopulationType,
				Variables:      dimIds,
				Filter: population.Filter{
					Codes:    areaOpts,
					Variable: areaTypeID,
				}})
			if sErr != nil {
				log.Error(ctx, "failed to get blocked area count", sErr, log.Data{
					"population_type": filterOutput.PopulationType,
					"variables":       dimIds,
					"area_codes":      areaOpts,
					"area_type_id":    areaTypeID,
				})
				setStatusCode(ctx, w, sErr)
				return
			}
		}
	} else {
		sdc = &cantabular.GetBlockedAreaCountResult{}
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

	homepageContent, err := zc.GetHomepageContent(ctx, userAccessToken, collectionID, lang, homepagePath)
	if err != nil {
		log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{err}), log.Data{"homepage_content": err})
	}

	showAll := req.URL.Query()[queryStrKey]
	basePage := rend.NewBasePageModel()
	m := mapper.CreateCensusFilterOutputsPage(req, basePage, datasetModel, ver, hasOtherVersions, allVers.Items, latestVersionNumber, latestVersionURL,
		lang, showAll, isValidationError, hasNoAreaOptions, filterOutput, fDims, homepageContent.ServiceMessage, homepageContent.EmergencyBanner,
		cfg.EnableMultivariate, dimDescriptions, *sdc, pop)
	m.DatasetLandingPage.OSRLogo = helpers.GetOSRLogoDetails(m.Language)

	rend.BuildPage(w, m, "census-landing")
}

func logError(ctx context.Context, w http.ResponseWriter, err error, msg string, data log.Data) bool {
	if err != nil {
		log.Error(ctx, msg, err, data)
		setStatusCode(ctx, w, err)
		return true
	}
	return false
}
