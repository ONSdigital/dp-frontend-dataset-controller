package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-net/v2/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

const (
	DatasetTypeNomis  = "nomis"
	DatasetTypeStatic = "static"
)

// FilterableLanding will load a filterable landing page
func FilterableLanding(dc DatasetAPISdkClient, pc PopulationClient, rend RenderClient, zc ZebedeeClient, cfg config.Config, apiRouterVersion string) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		filterableLanding(w, req, dc, pc, rend, zc, cfg, collectionID, lang, apiRouterVersion, userAccessToken)
	})
}

// nolint:gocognit,gocyclo // In future the redirect part should be handled in a different file to reduce complexity
func filterableLanding(responseWriter http.ResponseWriter, request *http.Request, dc DatasetAPISdkClient,
	populationClient PopulationClient, renderClient RenderClient, zebedeeClient ZebedeeClient, cfg config.Config,
	collectionID string, lang string, apiRouterVersion string, userAccessToken string) {
	var bc []zebedee.Breadcrumb
	var dims dpDatasetApiSdk.VersionDimensionsList
	var displayOtherVersionsLink bool
	var numOpts int
	var pageModel interface{}
	var templateName string

	downloadServiceAuthToken := ""
	serviceAuthToken := ""
	taxonomyURL := ""

	ctx := request.Context()
	vars := mux.Vars(request)

	datasetID := vars["datasetID"]
	editionID := vars["editionID"]
	versionID := vars["versionID"]

	form := request.URL.Query().Get("f")
	format := request.URL.Query().Get("format")
	isValidationError := false

	headers := dpDatasetApiSdk.Headers{
		CollectionID:         collectionID,
		DownloadServiceToken: downloadServiceAuthToken,
		ServiceToken:         serviceAuthToken,
		UserAccessToken:      userAccessToken,
	}

	// Fetch the dataset
	datasetDetails, err := dc.GetDataset(ctx, headers, collectionID, datasetID)
	if err != nil {
		setStatusCode(ctx, responseWriter, err)
		return
	}

	if editionID == "" {
		latestVersionURL, err := url.Parse(datasetDetails.Links.LatestVersion.HRef)
		if err != nil {
			setStatusCode(ctx, responseWriter, err)
			return
		}

		_, editionID, versionID, err = helpers.ExtractDatasetInfoFromPath(latestVersionURL.Path)
		if err != nil {
			setStatusCode(ctx, responseWriter, err)
			return
		}
	}

	// Fetch versions associated with dataset and redirect to latest if specific version isn't requested
	getVersionsQueryParams := dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}
	versionsList, err := dc.GetVersions(ctx, headers, datasetID, editionID, &getVersionsQueryParams)
	if err != nil {
		setStatusCode(ctx, responseWriter, err)
		return
	}
	allVersions := versionsList.Items

	if len(allVersions) > 1 {
		displayOtherVersionsLink = true
	}

	latestVersionNumber := 1
	for i := range versionsList.Items {
		singleVersion := &versionsList.Items[i]

		if singleVersion.Version > latestVersionNumber {
			latestVersionNumber = singleVersion.Version
		}
	}

	latestVersionURL := helpers.DatasetVersionURL(datasetID, editionID, strconv.Itoa(latestVersionNumber))

	if versionID == "" {
		log.Info(ctx, "no version provided, therefore redirecting to latest version", log.Data{"latestVersionURL": latestVersionURL})
		http.Redirect(responseWriter, request, latestVersionURL, http.StatusFound)
		return
	}

	version, err := dc.GetVersion(ctx, headers, datasetID, editionID, versionID)
	if err != nil {
		setStatusCode(ctx, responseWriter, err)
		return
	}

	// Check if this is a download request and redirect to get file if so
	if form == "get-data" {
		fileDownloadURL := ""
		// Try to get the download url based on dataset type. Static will be in distributions, otherwise downloads
		if datasetDetails.Type == DatasetTypeStatic {
			fileDownloadURL = helpers.GetDistributionFileURL(version.Distributions, format)
		} else {
			fileDownloadURL = helpers.GetDownloadFileURL(version.Downloads, format)
		}

		if fileDownloadURL == "" {
			// If download url is empty string, file not found so error
			isValidationError = true
		} else {
			// Otherwise redirect to valid file location
			http.Redirect(responseWriter, request, fileDownloadURL, http.StatusFound)
		}
	}

	// Fetch homepage content
	homepageContent, err := zebedeeClient.GetHomepageContent(ctx, userAccessToken, collectionID, lang, homepagePath)
	if err != nil {
		log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{err}), log.Data{"homepage_content": err})
	}

	// Build page context
	basePage := renderClient.NewBasePageModel()
	// Update basePage common parameters
	mapper.UpdateBasePage(&basePage, datasetDetails, homepageContent, isValidationError, lang, request)

	if datasetDetails.Type == DatasetTypeStatic {
		pageModel = mapper.CreateStaticOverviewPage(basePage, datasetDetails, version, allVersions, cfg.EnableMultivariate)
		templateName = DatasetTypeStatic
	} else {
		// Update dimensions based on dataset type
		if datasetDetails.Type == DatasetTypeNomis {
			dims = dpDatasetApiSdk.VersionDimensionsList{Items: nil}
		} else {
			dims, err = dc.GetVersionDimensions(ctx, headers, datasetID, editionID, versionID)
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
		opts, err := getOptionsSummary(ctx, dc, userAccessToken, collectionID, datasetID, editionID, versionID, dims, numOpts)
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
			if !(datasetDetails.Type == DatasetTypeNomis) {
				if datasetDetails.Links.Taxonomy != nil {
					taxonomyURL = datasetDetails.Links.Taxonomy.HRef
				}
				bc, err = zebedeeClient.GetBreadcrumb(ctx, userAccessToken, collectionID, lang, taxonomyURL)
				if err != nil {
					log.Warn(ctx, "unable to get breadcrumb for dataset uri", log.FormatErrors([]error{err}), log.Data{"taxonomy_url": taxonomyURL})
				}
			}
			// filterable landing mapper
			m := mapper.CreateFilterableLandingPage(ctx, basePage, datasetDetails, version, datasetID, opts,
				dims, displayOtherVersionsLink, bc, latestVersionNumber, latestVersionURL, apiRouterVersion,
				numOptsSummary)

			for i, d := range m.DatasetLandingPage.Version.Downloads {
				if cfg.DownloadServiceURL != "" {
					downloadURL, err := url.Parse(d.URI)
					if err != nil {
						setStatusCode(ctx, responseWriter, err)
						return
					}

					d.URI = cfg.DownloadServiceURL + downloadURL.Path
					m.DatasetLandingPage.Version.Downloads[i] = d
				}
			}

			m.DatasetLandingPage.Version.Downloads = append(m.DatasetLandingPage.Version.Downloads, model.Download{
				Extension: "txt",
				Size:      "0",
				URI:       fmt.Sprintf("/datasets/%s/editions/%s/versions/%s/metadata.txt", datasetID, editionID, versionID),
			})

			m.DatasetLandingPage.OSRLogo = helpers.GetOSRLogoDetails(m.Language)

			pageModel = m
			if datasetDetails.Type == DatasetTypeNomis {
				templateName = DatasetTypeNomis
			} else {
				templateName = "filterable"
			}
		}
	}
	// Render the page
	renderClient.BuildPage(responseWriter, pageModel, templateName)
}

func getDimensionCategorisationCountMap(ctx context.Context, pc PopulationClient, userAccessToken, populationType string, dims []dpDatasetApiModels.Dimension) map[string]int {
	m := make(map[string]int)
	var mutex sync.Mutex
	var wg sync.WaitGroup

	//nolint:gocritic //don't want to modify this goroutine
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

func sortedOpts(opts []dpDatasetApiSdk.VersionDimensionOptionsList) []dpDatasetApiSdk.VersionDimensionOptionsList {
	sorted := []dpDatasetApiSdk.VersionDimensionOptionsList{}
	for _, opt := range opts {
		sorted = append(sorted, dpDatasetApiSdk.VersionDimensionOptionsList{
			Items: sortOptionsByCode(opt.Items),
		})
	}
	return sorted
}
