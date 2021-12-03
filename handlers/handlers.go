package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/ONSdigital/dp-net/handlers"
	coreModel "github.com/ONSdigital/dp-renderer/model"

	"github.com/pkg/errors"

	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model"

	"github.com/gorilla/mux"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/log.go/v2/log"
)

const dataEndpoint = `\/data$`
const numOptsSummary = 50
const maxMetadataOptions = 1000
const maxAgeAndTimeOptions = 1000

// To mock interfaces in this file
//go:generate mockgen -source=handlers.go -destination=mock_handlers.go -package=handlers github.com/ONSdigital/dp-frontend-dataset-controller/handlers FilterClient,DatasetClient,RenderClient

// FilterClient is an interface with the methods required for a filter client
type FilterClient interface {
	CreateBlueprint(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version string, names []string) (filterID, eTag string, err error)
}

// DatasetClient is an interface with methods required for a dataset client
type DatasetClient interface {
	Get(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID string) (m dataset.DatasetDetails, err error)
	GetByPath(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, path string) (m dataset.DatasetDetails, err error)
	GetEditions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID string) (m []dataset.Edition, err error)
	GetEdition(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID, edition string) (dataset.Edition, error)
	GetVersions(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition string) (m []dataset.Version, err error)
	GetVersion(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, version string) (m dataset.Version, err error)
	GetVersionMetadata(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version string) (m dataset.Metadata, err error)
	GetVersionDimensions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version string) (m dataset.VersionDimensions, err error)
	GetOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension string, q *dataset.QueryParams) (m dataset.Options, err error)
}

// RenderClient is an interface with methods for require for rendering a template
type RenderClient interface {
	BuildPage(w io.Writer, pageModel interface{}, templateName string)
	NewBasePageModel() coreModel.Page
}

// ClientError is an interface that can be used to retrieve the status code if a client has errorred
type ClientError interface {
	Error() string
	Code() int
}

// errTooManyOptions is an error returned when a request can't complete because the dimension has too many options
var errTooManyOptions = errors.New("too many options in dimension")

func setStatusCode(req *http.Request, w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if err == errTooManyOptions {
		status = http.StatusRequestEntityTooLarge
	}
	if err, ok := err.(ClientError); ok {
		if err.Code() == http.StatusNotFound {
			status = err.Code()
		}
	}
	log.Error(req.Context(), "client error", err, log.Data{"setting-response-status": status})
	w.WriteHeader(status)
}

// CreateFilterID controls the creating of a filter idea when a new user journey is
// requested
func CreateFilterID(c FilterClient, dc DatasetClient) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		vars := mux.Vars(req)
		datasetID := vars["datasetID"]
		edition := vars["editionID"]
		version := vars["versionID"]
		ctx := req.Context()

		dimensions, err := dc.GetVersionDimensions(ctx, userAccessToken, "", collectionID, datasetID, edition, version)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}

		var names []string
		for _, dim := range dimensions.Items {
			// we are only interested in the totalCount, limit=0 will always return an empty list of items and the total count
			q := dataset.QueryParams{Offset: 0, Limit: 0}
			opts, err := dc.GetOptions(ctx, userAccessToken, "", collectionID, datasetID, edition, version, dim.Name, &q)
			if err != nil {
				setStatusCode(req, w, err)
				return
			}

			if opts.TotalCount > 1 { // If there is only one option then it can't be filterable so don't add to filter api
				names = append(names, dim.Name)
			}
		}
		fid, _, err := c.CreateBlueprint(ctx, userAccessToken, "", "", collectionID, datasetID, edition, version, names)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}

		log.Info(ctx, "created filter id", log.Data{"filter_id": fid})
		http.Redirect(w, req, "/filters/"+fid+"/dimensions", http.StatusMovedPermanently)
	})
}

// LegacyLanding will load a zebedee landing page
func LegacyLanding(zc ZebedeeClient, dc DatasetClient, rend RenderClient, cfg config.Config) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		legacyLanding(w, req, zc, dc, rend, collectionID, lang, userAccessToken)
	})
}

// FilterableLanding will load a filterable landing page
func FilterableLanding(dc DatasetClient, rend RenderClient, zc ZebedeeClient, cfg config.Config, apiRouterVersion string) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		filterableLanding(w, req, dc, rend, zc, cfg, collectionID, lang, apiRouterVersion, userAccessToken)
	})
}

// EditionsList will load a list of editions for a filterable dataset
func EditionsList(dc DatasetClient, zc ZebedeeClient, rend RenderClient, cfg config.Config, apiRouterVersion string) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		editionsList(w, req, dc, zc, rend, collectionID, lang, apiRouterVersion, userAccessToken)
	})
}

// VersionsList will load a list of versions for a filterable dataset
func VersionsList(dc DatasetClient, rend RenderClient, cfg config.Config) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		versionsList(w, req, dc, rend, collectionID, userAccessToken)
	})
}

func censusLanding(ctx context.Context, w http.ResponseWriter, req *http.Request, dc DatasetClient, datasetModel dataset.DatasetDetails, rend RenderClient, edition string, version dataset.Version, hasOtherVersions bool, allVersions []dataset.Version, latestVersionNumber int, latestVersionURL string, collectionID, lang, userAccessToken string) {
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
		setStatusCode(req, w, err)
		return
	}

	dims := dataset.VersionDimensions{Items: nil}
	dims, err = dc.GetVersionDimensions(ctx, userAccessToken, "", collectionID, datasetModel.ID, edition, fmt.Sprint(version.Version))
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	opts, err := getOptionsSummary(ctx, dc, userAccessToken, collectionID, datasetModel.ID, edition, fmt.Sprint(version.Version), dims, numOptsSummary)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	if version.Downloads == nil {
		version.Downloads = make(map[string]dataset.Download)
	}

	if form == "get-data" && format == "" {
		isValidationError = true
	}
	if form == "get-data" && format != "" {
		getDownloadFile(version, format, w, req)
	}

	basePage := rend.NewBasePageModel()
	m := mapper.CreateCensusDatasetLandingPage(ctx, req, basePage, datasetModel, version, opts, dims, initialVersionReleaseDate, hasOtherVersions, allVersions, latestVersionNumber, latestVersionURL, lang, numOptsSummary, isValidationError)
	rend.BuildPage(w, m, "census-landing")
}

func getDownloadFile(version dataset.Version, format string, w http.ResponseWriter, req *http.Request) {
	for ext, download := range version.Downloads {
		if strings.EqualFold(ext, format) {
			http.Redirect(w, req, download.URL, http.StatusFound)
		}
	}
}

func versionsList(w http.ResponseWriter, req *http.Request, dc DatasetClient, rend RenderClient, collectionID, userAccessToken string) {
	vars := mux.Vars(req)
	datasetID := vars["datasetID"]
	edition := vars["edition"]
	ctx := req.Context()

	d, err := dc.Get(ctx, userAccessToken, "", collectionID, datasetID)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	versions, err := dc.GetVersions(ctx, userAccessToken, "", "", collectionID, datasetID, edition)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	e, err := dc.GetEdition(ctx, userAccessToken, "", collectionID, datasetID, edition)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	basePage := rend.NewBasePageModel()
	m := mapper.CreateVersionsList(basePage, req, d, e, versions)
	rend.BuildPage(w, m, "version-list")
}

// getOptionsSummary requests a maximum of numOpts for each dimension, and returns the array of Options structs for each dimension, each one containing up to numOpts options.
func getOptionsSummary(ctx context.Context, dc DatasetClient, userAccessToken, collectionID, datasetID, edition, version string, dimensions dataset.VersionDimensions, numOpts int) (opts []dataset.Options, err error) {
	for _, dim := range dimensions.Items {

		// for time and age, request all the options (assumed less than maxAgeAndTimeOptions)
		if dim.Name == mapper.DimensionTime || dim.Name == mapper.DimensionAge {

			// query with limit maxAgeAndTimeOptions
			q := dataset.QueryParams{Offset: 0, Limit: maxAgeAndTimeOptions}
			opt, err := dc.GetOptions(ctx, userAccessToken, "", collectionID, datasetID, edition, version, dim.Name, &q)
			if err != nil {
				return opts, err
			}

			if opt.TotalCount > maxAgeAndTimeOptions {
				log.Warn(ctx, "total number of options is greater than the requested number", log.Data{"max_age_and_time_options": maxAgeAndTimeOptions, "total_count": opt.TotalCount})
			}

			opts = append(opts, opt)
			continue
		}

		// for other dimensions, cap the number of options to numOpts
		q := dataset.QueryParams{Offset: 0, Limit: numOpts}
		opt, err := dc.GetOptions(ctx, userAccessToken, "", collectionID, datasetID, edition, version, dim.Name, &q)
		if err != nil {
			return opts, err
		}
		opts = append(opts, opt)
	}
	return opts, nil
}

func filterableLanding(w http.ResponseWriter, req *http.Request, dc DatasetClient, rend RenderClient, zc ZebedeeClient, cfg config.Config, collectionID, lang, apiRouterVersion, userAccessToken string) {
	vars := mux.Vars(req)
	datasetID := vars["datasetID"]
	edition := vars["editionID"]
	version := vars["versionID"]
	ctx := req.Context()

	datasetModel, err := dc.Get(ctx, userAccessToken, "", collectionID, datasetID)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	if len(edition) == 0 {
		latestVersionURL, err := url.Parse(datasetModel.Links.LatestVersion.URL)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}

		_, edition, version, err = helpers.ExtractDatasetInfoFromPath(latestVersionURL.Path)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}
	}

	allVers, err := dc.GetVersions(ctx, userAccessToken, "", "", collectionID, datasetID, edition)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	var displayOtherVersionsLink bool
	if len(allVers) > 1 {
		displayOtherVersionsLink = true
	}

	latestVersionNumber := 1
	for _, singleVersion := range allVers {
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
		setStatusCode(req, w, err)
		return
	}

	if cfg.EnableCensusPages && strings.Contains(datasetModel.Type, "cantabular") {
		censusLanding(ctx, w, req, dc, datasetModel, rend, edition, ver, displayOtherVersionsLink, allVers, latestVersionNumber, latestVersionURL, collectionID, lang, userAccessToken)
		return
	}

	dims := dataset.VersionDimensions{Items: nil}
	if datasetModel.Type != "nomis" {
		dims, err = dc.GetVersionDimensions(ctx, userAccessToken, "", collectionID, datasetID, edition, version)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}
	}

	opts, err := getOptionsSummary(ctx, dc, userAccessToken, collectionID, datasetID, edition, version, dims, numOptsSummary)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	metadata, err := dc.GetVersionMetadata(ctx, userAccessToken, "", collectionID, datasetID, edition, version)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	// get metadata file content. If a dimension has too many options, ignore the error and a size 0 will be shown to the user
	textBytes, err := getText(dc, userAccessToken, collectionID, datasetID, edition, version, metadata, dims, req)
	if err != nil {
		if err != errTooManyOptions {
			setStatusCode(req, w, err)
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
	m := mapper.CreateFilterableLandingPage(basePage, ctx, req, datasetModel, ver, datasetID, opts, dims, displayOtherVersionsLink, bc, latestVersionNumber, latestVersionURL, lang, apiRouterVersion, numOptsSummary)

	for i, d := range m.DatasetLandingPage.Version.Downloads {
		if len(cfg.DownloadServiceURL) > 0 {
			downloadURL, err := url.Parse(d.URI)
			if err != nil {
				setStatusCode(req, w, err)
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

func editionsList(w http.ResponseWriter, req *http.Request, dc DatasetClient, zc ZebedeeClient, rend RenderClient, collectionID, lang, apiRouterVersion, userAccessToken string) {
	vars := mux.Vars(req)
	datasetID := vars["datasetID"]
	ctx := req.Context()

	datasetModel, err := dc.Get(ctx, userAccessToken, "", collectionID, datasetID)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	datasetEditions, err := dc.GetEditions(ctx, userAccessToken, "", collectionID, datasetID)
	if err != nil {
		if err, ok := err.(ClientError); ok {
			if err.Code() != http.StatusNotFound {
				setStatusCode(req, w, err)
				return
			}
		}
	}

	bc, err := zc.GetBreadcrumb(ctx, userAccessToken, userAccessToken, collectionID, datasetModel.Links.Taxonomy.URL)
	if err != nil {
		log.Warn(ctx, "unable to get breadcrumb for dataset uri", log.FormatErrors([]error{err}), log.Data{"taxonomy_url": datasetModel.Links.Taxonomy.URL})
	}

	numberOfEditions := len(datasetEditions)
	if numberOfEditions == 1 {
		latestVersionPath := helpers.DatasetVersionUrl(datasetID, datasetEditions[0].Edition, datasetEditions[0].Links.LatestVersion.ID)
		log.Info(ctx, "only one edition, therefore redirecting to latest version", log.Data{"latestVersionPath": latestVersionPath})
		http.Redirect(w, req, latestVersionPath, http.StatusFound)
	}

	basePage := rend.NewBasePageModel()
	m := mapper.CreateEditionsList(basePage, ctx, req, datasetModel, datasetEditions, datasetID, bc, lang, apiRouterVersion)
	rend.BuildPage(w, m, "edition-list")
}

func legacyLanding(w http.ResponseWriter, req *http.Request, zc ZebedeeClient, dc DatasetClient, rend RenderClient, collectionID, lang, userAccessToken string) {
	path := req.URL.Path
	ctx := req.Context()

	// Since MatchString will only error if the regex is invalid, and the regex is
	// constant, don't capture the error
	if ok, _ := regexp.MatchString(dataEndpoint, path); ok {
		b, err := zc.Get(ctx, userAccessToken, "/data?uri="+path)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}
		_, err = w.Write(b)
		if err != nil {
			setStatusCode(req, w, errors.Wrap(err, "failed to write zebedee client get response"))

		}

		return
	}

	dlp, err := zc.GetDatasetLandingPage(ctx, userAccessToken, collectionID, lang, path)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	bc, err := zc.GetBreadcrumb(ctx, userAccessToken, collectionID, lang, dlp.URI)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	var ds []zebedee.Dataset
	for _, v := range dlp.Datasets {
		d, err := zc.GetDataset(ctx, userAccessToken, collectionID, lang, v.URI)
		if err != nil {
			setStatusCode(req, w, errors.Wrap(err, "zebedee client legacy dataset returned an error"))
			return
		}
		ds = append(ds, d)
	}

	// Check for filterable datasets and fetch details
	if len(dlp.RelatedFilterableDatasets) > 0 {
		var relatedFilterableDatasets []zebedee.Related
		var wg sync.WaitGroup
		var mutex = &sync.Mutex{}
		for _, relatedFilterableDataset := range dlp.RelatedFilterableDatasets {
			wg.Add(1)
			go func(ctx context.Context, dc DatasetClient, relatedFilterableDataset zebedee.Related) {
				defer wg.Done()
				d, err := dc.GetByPath(ctx, userAccessToken, "", collectionID, relatedFilterableDataset.URI)
				if err != nil {
					// log error but continue to map data. any datasets that fail won't get mapped and won't be displayed on frontend
					log.Error(req.Context(), "error fetching dataset details", err, log.Data{
						"dataset": relatedFilterableDataset.URI,
					})
					return
				}
				mutex.Lock()
				defer mutex.Unlock()
				relatedFilterableDatasets = append(relatedFilterableDatasets, zebedee.Related{Title: d.Title, URI: relatedFilterableDataset.URI})
			}(req.Context(), dc, relatedFilterableDataset)
		}
		wg.Wait()
		dlp.RelatedFilterableDatasets = relatedFilterableDatasets
	}

	basePage := rend.NewBasePageModel()
	m := mapper.CreateLegacyDatasetLanding(basePage, ctx, req, dlp, bc, ds, lang)
	rend.BuildPage(w, m, "static")
}

// MetadataText generates a metadata text file
func MetadataText(dc DatasetClient, cfg config.Config) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		metadataText(w, req, dc, cfg, userAccessToken, collectionID)
	})
}

func metadataText(w http.ResponseWriter, req *http.Request, dc DatasetClient, cfg config.Config, userAccessToken, collectionID string) {
	vars := mux.Vars(req)
	datasetID := vars["datasetID"]
	edition := vars["edition"]
	version := vars["version"]
	ctx := req.Context()

	metadata, err := dc.GetVersionMetadata(ctx, userAccessToken, "", collectionID, datasetID, edition, version)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	dimensions, err := dc.GetVersionDimensions(ctx, userAccessToken, "", collectionID, datasetID, edition, version)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	b, err := getText(dc, userAccessToken, collectionID, datasetID, edition, version, metadata, dimensions, req)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	w.Header().Set("Content-Type", "plain/text")
	_, err = w.Write(b)
	if err != nil {
		setStatusCode(req, w, errors.Wrap(err, "failed to write metadata text response"))
	}
}

// getText gets a byte array containing the metadata content, based on options returned by dataset API.
// If a dimension has more than maxMetadataOptions, an error will be returned
func getText(dc DatasetClient, userAccessToken, collectionID, datasetID, edition, version string, metadata dataset.Metadata, dimensions dataset.VersionDimensions, req *http.Request) ([]byte, error) {
	var b bytes.Buffer

	b.WriteString(metadata.ToString())
	b.WriteString("Dimensions:\n")

	for _, dimension := range dimensions.Items {
		q := dataset.QueryParams{Offset: 0, Limit: maxMetadataOptions}
		options, err := dc.GetOptions(req.Context(), userAccessToken, "", collectionID, datasetID, edition, version, dimension.Name, &q)
		if err != nil {
			return nil, err
		}
		if options.TotalCount > maxMetadataOptions {
			return []byte{}, errTooManyOptions
		}

		b.WriteString(options.String())
	}

	return b.Bytes(), nil
}
