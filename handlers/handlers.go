package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"sync"

	"github.com/ONSdigital/dp-net/handlers"
	"github.com/ONSdigital/dp-net/request"

	"github.com/pkg/errors"

	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-models/model/datasetLandingPageFilterable"

	"github.com/gorilla/mux"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/log.go/log"
)

const dataEndpoint = `\/data$`
const numOptsSummary = 50

// To mock interfaces in this file
//go:generate mockgen -source=handlers.go -destination=mock_handlers.go -package=handlers github.com/ONSdigital/dp-frontend-dataset-controller/handlers FilterClient,DatasetClient,RenderClient

// FilterClient is an interface with the methods required for a filter client
type FilterClient interface {
	CreateBlueprint(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version string, names []string) (string, error)
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
	GetOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension string, offset, limit int) (m dataset.Options, err error)
}

// RenderClient is an interface with methods for require for rendering a template
type RenderClient interface {
	Do(string, []byte) ([]byte, error)
}

// ClientError is an interface that can be used to retrieve the status code if a client has errored
type ClientError interface {
	Error() string
	Code() int
}

func setStatusCode(req *http.Request, w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if err, ok := err.(ClientError); ok {
		if err.Code() == http.StatusNotFound {
			status = err.Code()
		}
	}
	log.Event(req.Context(), "client error", log.ERROR, log.Error(err), log.Data{"setting-response-status": status})
	w.WriteHeader(status)
}

func forwardFlorenceTokenIfRequired(req *http.Request) *http.Request {
	if len(req.Header.Get(request.FlorenceHeaderKey)) > 0 {
		ctx := request.SetFlorenceIdentity(req.Context(), req.Header.Get(request.FlorenceHeaderKey))
		return req.WithContext(ctx)
	}
	return req
}

// CreateFilterID controls the creating of a filter idea when a new user journey is
// requested
func CreateFilterID(c FilterClient, dc DatasetClient, cfg config.Config) http.HandlerFunc {
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
			// we are only interested in the totalCount, no need to request more items.
			offset := 0
			limit := 1
			opts, err := dc.GetOptions(ctx, userAccessToken, "", collectionID, datasetID, edition, version, dim.Name, offset, limit)
			if err != nil {
				setStatusCode(req, w, err)
				return
			}

			if opts.TotalCount > 1 { // If there is only one option then it can't be filterable so don't add to filter api
				names = append(names, dim.Name)
			}
		}
		fid, err := c.CreateBlueprint(ctx, userAccessToken, "", "", collectionID, datasetID, edition, version, names)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}

		log.Event(ctx, "created filter id", log.INFO, log.Data{"filter_id": fid})
		http.Redirect(w, req, "/filters/"+fid+"/dimensions", 301)
	})
}

// LegacyLanding will load a zebedee landing page
func LegacyLanding(zc ZebedeeClient, dc DatasetClient, rend RenderClient, cfg config.Config) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		legacyLanding(w, req, zc, dc, rend, cfg, collectionID, lang, userAccessToken)
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
		editionsList(w, req, dc, zc, rend, cfg, collectionID, lang, apiRouterVersion, userAccessToken)
	})
}

// VersionsList will load a list of versions for a filterable datase
func VersionsList(dc DatasetClient, rend RenderClient, cfg config.Config) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		versionsList(w, req, dc, rend, cfg, collectionID, lang, userAccessToken)
	})
}

func versionsList(w http.ResponseWriter, req *http.Request, dc DatasetClient, rend RenderClient, cfg config.Config, collectionID, lang, userAccessToken string) {
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

	p := mapper.CreateVersionsList(ctx, req, d, e, versions)
	b, err := json.Marshal(p)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	templateHTML, err := rend.Do("dataset-version-list", b)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	w.Write(templateHTML)
}

// getOptionsSummary requests a maximum of numOpts for each dimension, and returns the array of Options structs for each dimension, each one containing up to numOpts options.
func getOptionsSummary(ctx context.Context, dc DatasetClient, userAccessToken, collectionID, datasetID, edition, version string, dimensions dataset.VersionDimensions, numOpts int) (opts []dataset.Options, err error) {
	for _, dim := range dimensions.Items {
		opt, err := dc.GetOptions(ctx, userAccessToken, "", collectionID, datasetID, edition, version, dim.Name, 0, numOpts)
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

	bc, err := zc.GetBreadcrumb(ctx, userAccessToken, collectionID, lang, datasetModel.Links.Taxonomy.URL)
	if err != nil {
		log.Event(ctx, "unable to get breadcrumb for dataset uri", log.WARN, log.Error(err), log.Data{"taxonomy_url": datasetModel.Links.Taxonomy.URL})
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

	latestVersionOfEditionURL := helpers.DatasetVersionUrl(datasetID, edition, strconv.Itoa(latestVersionNumber))

	ver, err := dc.GetVersion(ctx, userAccessToken, "", "", collectionID, datasetID, edition, version)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	dims, err := dc.GetVersionDimensions(ctx, userAccessToken, "", collectionID, datasetID, edition, version)
	if err != nil {
		setStatusCode(req, w, err)
		return
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

	textBytes, err := getText(dc, userAccessToken, collectionID, datasetID, edition, version, metadata, dims, req, cfg)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	if ver.Downloads == nil {
		ver.Downloads = make(map[string]dataset.Download)
	}

	m := mapper.CreateFilterableLandingPage(ctx, req, datasetModel, ver, datasetID, opts, dims, displayOtherVersionsLink, bc, latestVersionNumber, latestVersionOfEditionURL, lang, apiRouterVersion, numOptsSummary)

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
	m.DatasetLandingPage.Version.Downloads = append(m.DatasetLandingPage.Version.Downloads, datasetLandingPageFilterable.Download{
		Extension: "txt",
		Size:      strconv.Itoa(len(textBytes)),
		URI:       fmt.Sprintf("/datasets/%s/editions/%s/versions/%s/metadata.txt", datasetID, edition, version),
	})

	b, err := json.Marshal(m)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	templateHTML, err := rend.Do("dataset-landing-page-filterable", b)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	w.Write(templateHTML)

}

func editionsList(w http.ResponseWriter, req *http.Request, dc DatasetClient, zc ZebedeeClient, rend RenderClient, cfg config.Config, collectionID, lang, apiRouterVersion, userAccessToken string) {
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
		log.Event(ctx, "unable to get breadcrumb for dataset uri", log.WARN, log.Error(err), log.Data{"taxonomy_url": datasetModel.Links.Taxonomy.URL})
	}

	numberOfEditions := len(datasetEditions)
	if numberOfEditions == 1 {
		latestVersionPath := helpers.DatasetVersionUrl(datasetID, datasetEditions[0].Edition, datasetEditions[0].Links.LatestVersion.ID)
		log.Event(ctx, "only one edition, therefore redirecting to latest version", log.INFO, log.Data{"latestVersionPath": latestVersionPath})
		http.Redirect(w, req, latestVersionPath, 302)
	}

	m := mapper.CreateEditionsList(ctx, req, datasetModel, datasetEditions, datasetID, bc, lang, apiRouterVersion)

	b, err := json.Marshal(m)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	templateHTML, err := rend.Do("dataset-edition-list", b)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	w.Write(templateHTML)
	return

}

func legacyLanding(w http.ResponseWriter, req *http.Request, zc ZebedeeClient, dc DatasetClient, rend RenderClient, cfg config.Config, collectionID, lang, userAccessToken string) {
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
		w.Write(b)
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
					log.Event(req.Context(), "error fetching dataset details", log.ERROR, log.Error(err), log.Data{
						"dataset": relatedFilterableDataset.URI,
					})
					return
				}
				mutex.Lock()
				defer mutex.Unlock()
				relatedFilterableDatasets = append(relatedFilterableDatasets, zebedee.Related{Title: d.Title, URI: relatedFilterableDataset.URI})
				return
			}(req.Context(), dc, relatedFilterableDataset)
		}
		wg.Wait()
		dlp.RelatedFilterableDatasets = relatedFilterableDatasets
	}

	m := mapper.CreateLegacyDatasetLanding(ctx, req, dlp, bc, ds, lang)

	var templateJSON []byte
	templateJSON, err = json.Marshal(m)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	templateHTML, err := rend.Do("dataset-landing-page-static", templateJSON)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	w.Write(templateHTML)
	return

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

	b, err := getText(dc, userAccessToken, collectionID, datasetID, edition, version, metadata, dimensions, req, cfg)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	w.Header().Set("Content-Type", "plain/text")

	w.Write(b)

}

func getText(dc DatasetClient, userAccessToken, collectionID, datasetID, edition, version string, metadata dataset.Metadata, dimensions dataset.VersionDimensions, req *http.Request, cfg config.Config) ([]byte, error) {
	var b bytes.Buffer

	b.WriteString(metadata.ToString())
	b.WriteString("Dimensions:\n")

	for _, dimension := range dimensions.Items {
		options, err := dc.GetOptions(req.Context(), userAccessToken, "", collectionID, datasetID, edition, version, dimension.Name, 0, 0)
		if err != nil {
			return nil, err
		}

		b.WriteString(options.String())
	}

	return b.Bytes(), nil
}

func getUserAccessTokenFromContext(ctx context.Context) string {
	if ctx.Value(request.FlorenceIdentityKey) != nil {
		accessToken, ok := ctx.Value(request.FlorenceIdentityKey).(string)
		if !ok {
			log.Event(ctx, "error retrieving user access token", log.WARN, log.Error(errors.New("error casting access token context value to string")))
		}
		return accessToken
	}
	return ""
}

func getCollectionIDFromContext(ctx context.Context) string {
	if ctx.Value(request.CollectionIDHeaderKey) != nil {
		collectionID, ok := ctx.Value(request.CollectionIDHeaderKey).(string)
		if !ok {
			log.Event(ctx, "error retrieving collection ID", log.WARN, log.Error(errors.New("error casting collection ID context value to string")))
		}
		return collectionID
	}
	return ""
}
