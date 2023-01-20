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

	getDimensionOptions := func(dim filter.ModelDimension) ([]string, int, error) {
		q := dataset.QueryParams{Offset: 0, Limit: 1000}
		opts, err := dc.GetOptions(ctx, userAccessToken, "", collectionID, datasetModel.ID, edition, strconv.Itoa(ver.Version), dim.Name, &q)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get options for dimension: %w", err)
		}

		var options []string
		for _, opt := range sortOptionsByCode(opts.Items) {
			options = append(options, opt.Label)
		}

		return options, opts.TotalCount, nil
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

		// TODO: pc.GetParentAreaCount is causing production issues
		// if dim.FilterByParent != "" {
		// 	count, err := pc.GetParentAreaCount(ctx, population.GetParentAreaCountInput{
		// 		AuthTokens: population.AuthTokens{
		// 			UserAuthToken: userAccessToken,
		// 		},
		// 		PopulationType:   filterOutput.PopulationType,
		// 		AreaTypeID:       dim.ID,
		// 		ParentAreaTypeID: dim.FilterByParent,
		// 		Areas:            optsIDs,
		// 		SVarID:           supVar,
		// 	})
		// 	if err != nil {
		// 		log.Error(ctx, "failed to get parent area count", err, log.Data{
		// 			"dataset_id":                filterOutput.PopulationType,
		// 			"area_type_id":              dim.ID,
		// 			"parent_area_type_id":       dim.FilterByParent,
		// 			"areas":                     optsIDs,
		// 			"supplementary_variable_id": supVar,
		// 		})
		// 		return nil, 0, err
		// 	}

		// 	totalCount = count
		// }

		return options, totalCount, nil
	}

	getOptions := func(dim filter.ModelDimension) ([]string, int, error) {
		if dim.IsAreaType != nil && *dim.IsAreaType {
			return getAreaOptions(dim)
		}

		return getDimensionOptions(dim)
	}

	var fDims []model.FilterDimension
	for i := len(filterOutput.Dimensions) - 1; i >= 0; i-- {
		// TODO: pc.GetParentAreaCount is causing production issues
		// if filterOutput.Dimensions[i].IsAreaType == nil || !*filterOutput.Dimensions[i].IsAreaType {
		// 	supVar = filterOutput.Dimensions[i].ID
		// }
		options, count, err := getOptions(filterOutput.Dimensions[i])
		if err != nil {
			log.Error(ctx, "failed to get options for dimension", err, log.Data{"dimension_name": filterOutput.Dimensions[i].Name})
			setStatusCode(ctx, w, err)
			return
		}

		filterOutput.Dimensions[i].Options = options
		fDims = append(fDims, model.FilterDimension{
			ModelDimension: filterOutput.Dimensions[i],
			OptionsCount:   count,
		})
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
	m := mapper.CreateCensusFilterOutputsPage(ctx, req, basePage, datasetModel, ver, initialVersionReleaseDate, hasOtherVersions, allVers.Items, latestVersionNumber, latestVersionURL, lang, showAll, numOptsSummary, isValidationError, hasNoAreaOptions, filterOutput.Downloads, fDims, homepageContent.ServiceMessage, homepageContent.EmergencyBanner, cfg.EnableMultivariate)
	rend.BuildPage(w, m, "census-landing")
}
