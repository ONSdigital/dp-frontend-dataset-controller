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
	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-net/v2/handlers"
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

	var bc []zebedee.Breadcrumb
	var dims dataset.VersionDimensions
	var displayOtherVersionsLink bool
	var numOpts int
	var pageModel interface{}
	var templateName string

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
		fileDownloadUrl := ""
		// Try to get the download url based on dataset type. Static will be in distributions, otherwise downloads
		if datasetDetails.Type == "static" {
			fileDownloadUrl = helpers.GetDistributionFileUrl(version.Distributions, format)
		} else {
			fileDownloadUrl = helpers.GetDownloadFileUrl(version.Downloads, format)
		}

		if fileDownloadUrl == "" {
			// If download url is empty string, file not found so error
			isValidationError = true
		} else {
			// Otherwise redirect to valid file location
			http.Redirect(responseWriter, request, fileDownloadUrl, http.StatusFound)
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

	if datasetDetails.Type == "static" {
		pageModel = mapper.CreateStaticOverviewPage(basePage, datasetDetails, version, allVersions, cfg.EnableMultivariate)
		templateName = "static"
	} else {
		// Update dimensions based on dataset type
		if datasetDetails.Type == "nomis" {
			dims = dataset.VersionDimensions{Items: nil}
		} else {
			dims, err = datasetClient.GetVersionDimensions(ctx, userAccessToken, serviceAuthToken, collectionId,
				datasetId, editionId, versionId)
			if err != nil {
				setStatusCode(ctx, responseWriter, err)
				return
			}
		}

		// options
		// Set number of options to request based on typeUpdate numOptsSummary if cantabular type
		if strings.Contains(datasetDetails.Type, "cantabular") {
			numOpts = 1000
		} else {
			// Load from constant
			numOpts = numOptsSummary
		}
		opts, err := getOptionsSummary(ctx, datasetClient, userAccessToken, collectionId, datasetId, editionId, versionId, dims, numOpts)
		if err != nil {
			setStatusCode(ctx, responseWriter, err)
			return
		}

		if strings.Contains(datasetDetails.Type, "cantabular") {
			idOfVersionBasedOn := version.IsBasedOn.ID
			// population client stuff
			pop, err := populationClient.GetPopulationType(ctx, population.GetPopulationTypeInput{
				PopulationType: idOfVersionBasedOn,
				AuthTokens: population.AuthTokens{
					UserAuthToken: userAccessToken,
				},
			})
			if err != nil {
				log.Error(ctx, "failed to get population types", err)
				setStatusCode(ctx, responseWriter, err)
				return
			}

			categorisationsMap := getDimensionCategorisationCountMap(ctx, populationClient, userAccessToken, idOfVersionBasedOn, version.Dimensions)

			// census mapper
			opts = sortedOpts(opts)

			showAll := request.URL.Query()[queryStrKey]

			m := mapper.CreateCensusLandingPage(basePage, datasetDetails, version, opts, categorisationsMap, allVersions, showAll, cfg.EnableMultivariate, pop)
			m.DatasetLandingPage.OSRLogo = helpers.GetOSRLogoDetails(m.Language)

			pageModel = m
			templateName = "census-landing"
		} else {
			// Update breadcrumbs if not nomis
			if !(datasetDetails.Type == "nomis") {
				bc, err = zebedeeClient.GetBreadcrumb(ctx, userAccessToken, collectionId, lang, datasetDetails.Links.Taxonomy.URL)
				if err != nil {
					log.Warn(ctx, "unable to get breadcrumb for dataset uri", log.FormatErrors([]error{err}), log.Data{"taxonomy_url": datasetDetails.Links.Taxonomy.URL})
				}
			}
			// filterable landing mapper
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

			pageModel = m
			if datasetDetails.Type == "nomis" {
				templateName = "nomis"
			} else {
				templateName = "filterable"
			}
		}
	}
	// Render the page
	renderClient.BuildPage(responseWriter, pageModel, templateName)
}

func getDimensionCategorisationCountMap(ctx context.Context, pc PopulationClient, userAccessToken string, populationType string, dims []dpDatasetApiModels.Dimension) map[string]int {
	m := make(map[string]int)
	var mutex sync.Mutex
	var wg sync.WaitGroup

	for _, dim := range dims {
		if !helpers.IsBoolPtr(dim.IsAreaType) {
			wg.Add(1)
			go func(dim dpDatasetApiModels.Dimension) {
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
