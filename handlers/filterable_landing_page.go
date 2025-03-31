package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-net/v2/handlers"
	dpRendererModel "github.com/ONSdigital/dp-renderer/v2/model"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// FilterableLanding will load a filterable landing page
func FilterableLanding(dc DatasetClient, pc PopulationClient, rend RenderClient, zc ZebedeeClient, cfg config.Config, apiRouterVersion string) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		filterableLanding(w, req, dc, pc, rend, zc, cfg, collectionID, lang, apiRouterVersion, userAccessToken)
	})
}

func filterableLanding(responseWriter http.ResponseWriter, request *http.Request, datasetClient DatasetClient,
	populationClient PopulationClient, renderClient RenderClient, zebedeeClient ZebedeeClient, cfg config.Config,
	collectionId string, lang string, apiRouterVersion string, userAccessToken string) {

	downloadServiceAuthToken := ""
	serviceAuthToken := ""

	ctx := request.Context()
	vars := mux.Vars(request)

	datasetId := vars["datasetID"]
	editionId := vars["editionID"]
	versionId := vars["versionID"]

	form := request.URL.Query().Get("f")
	format := request.URL.Query().Get("format")
	isValidationError := false

	// Fetch the dataset
	datasetDetails, err := datasetClient.Get(ctx, userAccessToken, serviceAuthToken, collectionId, datasetId)
	if err != nil {
		setStatusCode(ctx, responseWriter, err)
		return
	}

	if len(editionId) == 0 {
		latestVersionURL, err := url.Parse(datasetDetails.Links.LatestVersion.URL)
		if err != nil {
			setStatusCode(ctx, responseWriter, err)
			return
		}

		_, editionId, versionId, err = helpers.ExtractDatasetInfoFromPath(latestVersionURL.Path)
		if err != nil {
			setStatusCode(ctx, responseWriter, err)
			return
		}
	}

	// Fetch versions associated with dataset and redirect to latest if specific version isn't requested
	getVersionsQueryParams := dataset.QueryParams{Offset: 0, Limit: 1000}
	versionsList, err := datasetClient.GetVersions(ctx, userAccessToken, serviceAuthToken, downloadServiceAuthToken,
		collectionId, datasetId, editionId, &getVersionsQueryParams)
	if err != nil {
		setStatusCode(ctx, responseWriter, err)
		return
	}
	allVersions := versionsList.Items

	var displayOtherVersionsLink bool
	if len(allVersions) > 1 {
		displayOtherVersionsLink = true
	}

	latestVersionNumber := 1
	for _, singleVersion := range versionsList.Items {
		if singleVersion.Version > latestVersionNumber {
			latestVersionNumber = singleVersion.Version
		}
	}

	latestVersionURL := helpers.DatasetVersionURL(datasetId, editionId, strconv.Itoa(latestVersionNumber))

	if versionId == "" {
		log.Info(ctx, "no version provided, therefore redirecting to latest version", log.Data{"latestVersionURL": latestVersionURL})
		http.Redirect(responseWriter, request, latestVersionURL, http.StatusFound)
		return
	}

	version, err := datasetClient.GetVersion(ctx, userAccessToken, serviceAuthToken, downloadServiceAuthToken, collectionId, datasetId, editionId, versionId)
	if err != nil {
		setStatusCode(ctx, responseWriter, err)
		return
	}

	// Check if this is a download request and redirect to get file if so
	if form == "get-data" {
		if format == "" {
			// Format not valid so raise error
			isValidationError = true
		} else {
			getDownloadFile(version.Downloads, format, responseWriter, request)
		}
	}

	// Fetch homepage content
	homepageContent, err := zebedeeClient.GetHomepageContent(ctx, userAccessToken, collectionId, lang, homepagePath)
	if err != nil {
		log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{err}), log.Data{"homepage_content": err})
	}

	// Build page context
	basePage := renderClient.NewBasePageModel()
	// Update basePage common parameters
	mapper.UpdateBasePage(&basePage, datasetDetails, homepageContent, isValidationError, lang, request)

	if strings.Contains(datasetDetails.Type, "cantabular") {
		censusLanding(basePage, cfg, ctx, responseWriter, request, datasetClient, populationClient, datasetDetails,
			renderClient, editionId, version, allVersions, collectionId, userAccessToken)
		return
	}

	if version.Downloads == nil {
		version.Downloads = make(map[string]dataset.Download)
	}

	if datasetDetails.Type == "static" {
		m := mapper.CreateStaticOverviewPage(basePage, datasetDetails, version, allVersions, cfg.EnableMultivariate)
		renderClient.BuildPage(responseWriter, m, "static")
	} else {
		dims := dataset.VersionDimensions{Items: nil}
		var bc []zebedee.Breadcrumb

		// Unless type is nomis, update values of bc and dims
		if !(datasetDetails.Type == "nomis") {
			dims, err = datasetClient.GetVersionDimensions(ctx, userAccessToken, serviceAuthToken, collectionId,
				datasetId, editionId, versionId)
			if err != nil {
				setStatusCode(ctx, responseWriter, err)
				return
			}
			bc, err = zebedeeClient.GetBreadcrumb(ctx, userAccessToken, collectionId, lang, datasetDetails.Links.Taxonomy.URL)
			if err != nil {
				log.Warn(ctx, "unable to get breadcrumb for dataset uri", log.FormatErrors([]error{err}), log.Data{"taxonomy_url": datasetDetails.Links.Taxonomy.URL})
			}
		}

		opts, err := getOptionsSummary(ctx, datasetClient, userAccessToken, collectionId, datasetId, editionId, versionId, dims, numOptsSummary)
		if err != nil {
			setStatusCode(ctx, responseWriter, err)
			return
		}

		m := mapper.CreateFilterableLandingPage(ctx, basePage, datasetDetails, version, datasetId, opts,
			dims, displayOtherVersionsLink, bc, latestVersionNumber, latestVersionURL, apiRouterVersion,
			numOptsSummary)

		for i, d := range m.DatasetLandingPage.Version.Downloads {
			if len(cfg.DownloadServiceURL) > 0 {
				downloadURL, err := url.Parse(d.URI)
				if err != nil {
					setStatusCode(ctx, responseWriter, err)
					return
				}

				d.URI = cfg.DownloadServiceURL + downloadURL.Path
				m.DatasetLandingPage.Version.Downloads[i] = d
			}
		}

		metadata, err := datasetClient.GetVersionMetadata(ctx, userAccessToken, serviceAuthToken, collectionId, datasetId, editionId, versionId)
		if err != nil {
			setStatusCode(ctx, responseWriter, err)
			return
		}

		// get metadata file content. If a dimension has too many options, ignore the error and a size 0 will be shown to the user
		textBytes, err := getText(datasetClient, userAccessToken, collectionId, datasetId, editionId, versionId, metadata, dims, request)
		if err != nil {
			if err != errTooManyOptions {
				setStatusCode(ctx, responseWriter, err)
				return
			}
		}

		// This needs to be after the for-loop to add the download files,
		// because the loop adds the download services domain to the URLs
		// which this text file doesn't need because it's created on-the-fly
		// by this app
		m.DatasetLandingPage.Version.Downloads = append(m.DatasetLandingPage.Version.Downloads, model.Download{
			Extension: "txt",
			Size:      strconv.Itoa(len(textBytes)),
			URI:       fmt.Sprintf("/datasets/%s/editions/%s/versions/%s/metadata.txt", datasetId, editionId, versionId),
		})

		m.DatasetLandingPage.OSRLogo = helpers.GetOSRLogoDetails(m.Language)

		templateName := "filterable"
		if datasetDetails.Type == "nomis" {
			templateName = "nomis"
		}

		renderClient.BuildPage(responseWriter, m, templateName)
	}
}

func censusLanding(basePage dpRendererModel.Page, cfg config.Config, ctx context.Context, w http.ResponseWriter, req *http.Request,
	dc DatasetClient, pc PopulationClient, datasetModel dataset.DatasetDetails, rend RenderClient, edition string,
	version dataset.Version, allVersions []dataset.Version, collectionID, userAccessToken string) {
	const numOptsSummary = 1000
	var err error
	idOfVersionBasedOn := version.IsBasedOn.ID

	pop, err := pc.GetPopulationType(ctx, population.GetPopulationTypeInput{
		PopulationType: idOfVersionBasedOn,
		AuthTokens: population.AuthTokens{
			UserAuthToken: userAccessToken,
		},
	})
	if err != nil {
		log.Error(ctx, "failed to get population types", err)
		setStatusCode(ctx, w, err)
		return
	}

	dims := dataset.VersionDimensions{Items: version.Dimensions}
	categorisationsMap := getDimensionCategorisationCountMap(
		ctx,
		pc,
		userAccessToken,
		idOfVersionBasedOn,
		version.Dimensions,
	)

	opts, err := getOptionsSummary(
		ctx,
		dc,
		userAccessToken,
		collectionID,
		datasetModel.ID,
		edition,
		fmt.Sprint(version.Version),
		dims,
		numOptsSummary,
	)
	if err != nil {
		setStatusCode(ctx, w, err)
		return
	}
	opts = sortedOpts(opts)

	if version.Downloads == nil {
		log.Warn(ctx, "version downloads are nil", log.Data{"version_id": version.ID})
		version.Downloads = make(map[string]dataset.Download)
	}

	showAll := req.URL.Query()[queryStrKey]

	m := mapper.CreateCensusLandingPage(basePage, datasetModel, version, opts, categorisationsMap, allVersions, showAll, cfg.EnableMultivariate, pop)
	m.DatasetLandingPage.OSRLogo = helpers.GetOSRLogoDetails(m.Language)

	rend.BuildPage(w, m, "census-landing")
}

func getDownloadFile(downloads map[string]dataset.Download, format string, w http.ResponseWriter, req *http.Request) {
	for ext, download := range downloads {
		if strings.EqualFold(ext, format) {
			http.Redirect(w, req, download.URL, http.StatusFound)
		}
	}
}

func getDimensionCategorisationCountMap(ctx context.Context, pc PopulationClient, userAccessToken string, populationType string, dims []dataset.VersionDimension) map[string]int {
	m := make(map[string]int)
	var mutex sync.Mutex
	var wg sync.WaitGroup

	for _, dim := range dims {
		if !helpers.IsBoolPtr(dim.IsAreaType) {
			wg.Add(1)
			go func(dim dataset.VersionDimension) {
				defer wg.Done()
				cats, err := pc.GetCategorisations(ctx, population.GetCategorisationsInput{
					AuthTokens: population.AuthTokens{
						UserAuthToken: userAccessToken,
					},
					PaginationParams: population.PaginationParams{
						Limit: 1000,
					},
					PopulationType: populationType,
					Dimension:      dim.ID,
				})
				defer mutex.Unlock()
				mutex.Lock()

				if err != nil {
					m[dim.ID] = 1
				} else {
					m[dim.ID] = cats.PaginationResponse.TotalCount
				}
			}(dim)
		}
	}
	wg.Wait()

	return m
}

func sortedOpts(opts []dataset.Options) []dataset.Options {
	sorted := []dataset.Options{}
	for _, opt := range opts {
		sorted = append(sorted, dataset.Options{
			Items:      sortOptionsByCode(opt.Items),
			Count:      opt.Count,
			Offset:     opt.Offset,
			Limit:      opt.Limit,
			TotalCount: opt.TotalCount,
		})
	}
	return sorted
}
