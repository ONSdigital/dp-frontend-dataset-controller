package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-net/v2/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// FilterableLanding will load a filterable landing page
func FilterableLanding(dc DatasetClient, rend RenderClient, zc ZebedeeClient, cfg config.Config, apiRouterVersion string) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		filterableLanding(w, req, dc, rend, zc, cfg, collectionID, lang, apiRouterVersion, userAccessToken)
	})
}

func filterableLanding(w http.ResponseWriter, req *http.Request, dc DatasetClient, rend RenderClient, zc ZebedeeClient, cfg config.Config, collectionID, lang, apiRouterVersion, userAccessToken string) {
	vars := mux.Vars(req)
	datasetID := vars["datasetID"]
	edition := vars["editionID"]
	version := vars["versionID"]
	ctx := req.Context()

	datasetModel, err := dc.Get(ctx, userAccessToken, "", collectionID, datasetID)
	if err != nil {
		setStatusCode(ctx, w, err)
		return
	}

	homepageContent, err := zc.GetHomepageContent(ctx, userAccessToken, collectionID, lang, homepagePath)
	if err != nil {
		log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{err}), log.Data{"homepage_content": err})
	}

	if len(edition) == 0 {
		latestVersionURL, err := url.Parse(datasetModel.Links.LatestVersion.URL)
		if err != nil {
			setStatusCode(ctx, w, err)
			return
		}

		_, edition, version, err = helpers.ExtractDatasetInfoFromPath(latestVersionURL.Path)
		if err != nil {
			setStatusCode(ctx, w, err)
			return
		}
	}

	q := dataset.QueryParams{Offset: 0, Limit: 1000}
	allVers, err := dc.GetVersions(ctx, userAccessToken, "", "", collectionID, datasetID, edition, &q)
	if err != nil {
		setStatusCode(ctx, w, err)
		return
	}

	var displayOtherVersionsLink bool
	if len(allVers.Items) > 1 {
		displayOtherVersionsLink = true
	}

	latestVersionNumber := 1
	for _, singleVersion := range allVers.Items {
		if singleVersion.Version > latestVersionNumber {
			latestVersionNumber = singleVersion.Version
		}
	}

	latestVersionURL := helpers.DatasetVersionUrl(datasetID, edition, strconv.Itoa(latestVersionNumber))

	if version == "" {
		log.Info(ctx, "no version provided, therefore redirecting to latest version", log.Data{"latestVersionURL": latestVersionURL})
		http.Redirect(w, req, latestVersionURL, http.StatusFound)
		return
	}

	ver, err := dc.GetVersion(ctx, userAccessToken, "", "", collectionID, datasetID, edition, version)
	if err != nil {
		setStatusCode(ctx, w, err)
		return
	}

	if cfg.EnableCensusPages && strings.Contains(datasetModel.Type, "cantabular") {
		censusLanding(cfg.EnableMultivariate, ctx, w, req, dc, datasetModel, rend, edition, ver, displayOtherVersionsLink, allVers.Items, latestVersionNumber, latestVersionURL, collectionID, lang, userAccessToken, homepageContent.ServiceMessage, homepageContent.EmergencyBanner)
		return
	}

	dims := dataset.VersionDimensions{Items: nil}
	if datasetModel.Type != "nomis" {
		dims, err = dc.GetVersionDimensions(ctx, userAccessToken, "", collectionID, datasetID, edition, version)
		if err != nil {
			setStatusCode(ctx, w, err)
			return
		}
	}

	opts, err := getOptionsSummary(ctx, dc, userAccessToken, collectionID, datasetID, edition, version, dims, numOptsSummary)
	if err != nil {
		setStatusCode(ctx, w, err)
		return
	}

	metadata, err := dc.GetVersionMetadata(ctx, userAccessToken, "", collectionID, datasetID, edition, version)
	if err != nil {
		setStatusCode(ctx, w, err)
		return
	}

	// get metadata file content. If a dimension has too many options, ignore the error and a size 0 will be shown to the user
	textBytes, err := getText(dc, userAccessToken, collectionID, datasetID, edition, version, metadata, dims, req)
	if err != nil {
		if err != errTooManyOptions {
			setStatusCode(ctx, w, err)
			return
		}
	}

	if ver.Downloads == nil {
		ver.Downloads = make(map[string]dataset.Download)
	}

	var bc []zebedee.Breadcrumb
	if datasetModel.Type != "nomis" {
		bc, err = zc.GetBreadcrumb(ctx, userAccessToken, collectionID, lang, datasetModel.Links.Taxonomy.URL)
		if err != nil {
			log.Warn(ctx, "unable to get breadcrumb for dataset uri", log.FormatErrors([]error{err}), log.Data{"taxonomy_url": datasetModel.Links.Taxonomy.URL})
		}
	}

	basePage := rend.NewBasePageModel()
	m := mapper.CreateFilterableLandingPage(basePage, ctx, req, datasetModel, ver, datasetID, opts, dims, displayOtherVersionsLink, bc, latestVersionNumber, latestVersionURL, lang, apiRouterVersion, numOptsSummary, homepageContent.ServiceMessage, homepageContent.EmergencyBanner)

	for i, d := range m.DatasetLandingPage.Version.Downloads {
		if len(cfg.DownloadServiceURL) > 0 {
			downloadURL, err := url.Parse(d.URI)
			if err != nil {
				setStatusCode(ctx, w, err)
				return
			}

			d.URI = cfg.DownloadServiceURL + downloadURL.Path
			m.DatasetLandingPage.Version.Downloads[i] = d
		}
	}

	// This needs to be after the for-loop to add the download files,
	// because the loop adds the download services domain to the URLs
	// which this text file doesn't need because it's created on-the-fly
	// by this app
	m.DatasetLandingPage.Version.Downloads = append(m.DatasetLandingPage.Version.Downloads, model.Download{
		Extension: "txt",
		Size:      strconv.Itoa(len(textBytes)),
		URI:       fmt.Sprintf("/datasets/%s/editions/%s/versions/%s/metadata.txt", datasetID, edition, version),
	})

	templateName := "filterable"
	if datasetModel.Type == "nomis" {
		templateName = "nomis"
	}

	rend.BuildPage(w, m, templateName)
}

func censusLanding(isEnableMultivariate bool, ctx context.Context, w http.ResponseWriter, req *http.Request, dc DatasetClient, datasetModel dataset.DatasetDetails, rend RenderClient, edition string, version dataset.Version, hasOtherVersions bool, allVersions []dataset.Version, latestVersionNumber int, latestVersionURL, collectionID, lang, userAccessToken string, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) {
	const numOptsSummary = 1000
	var initialVersion dataset.Version
	var initialVersionReleaseDate string
	var err error
	var form = req.URL.Query().Get("f")
	var format = req.URL.Query().Get("format")
	var isValidationError bool

	if version.Version != 1 {
		initialVersion, err = dc.GetVersion(ctx, userAccessToken, "", "", collectionID, datasetModel.ID, edition, "1")
		initialVersionReleaseDate = initialVersion.ReleaseDate
	}

	if err != nil {
		setStatusCode(ctx, w, err)
		return
	}

	dims := dataset.VersionDimensions{Items: version.Dimensions}

	opts, err := getOptionsSummary(ctx, dc, userAccessToken, collectionID, datasetModel.ID, edition, fmt.Sprint(version.Version), dims, numOptsSummary)
	if err != nil {
		setStatusCode(ctx, w, err)
		return
	}
	opts = sortedOpts(opts)

	if version.Downloads == nil {
		log.Warn(ctx, "version downloads are nil", log.Data{"version_id": version.ID})
		version.Downloads = make(map[string]dataset.Download)
	}

	if form == "get-data" && format == "" {
		isValidationError = true
	}
	if form == "get-data" && format != "" {
		getDownloadFile(version.Downloads, format, w, req)
	}

	showAll := req.URL.Query()[queryStrKey]
	basePage := rend.NewBasePageModel()
	m := mapper.CreateCensusLandingPage(ctx, req, basePage, datasetModel, version, opts, initialVersionReleaseDate, hasOtherVersions, allVersions, latestVersionNumber, latestVersionURL, lang, showAll, numOptsSummary, isValidationError, serviceMessage, emergencyBannerContent, isEnableMultivariate)
	rend.BuildPage(w, m, "census-landing")
}

func getDownloadFile(downloads map[string]dataset.Download, format string, w http.ResponseWriter, req *http.Request) {
	for ext, download := range downloads {
		if strings.EqualFold(ext, format) {
			http.Redirect(w, req, download.URL, http.StatusFound)
		}
	}
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
