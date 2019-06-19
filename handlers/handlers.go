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

	"github.com/pkg/errors"

	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-models/model/datasetLandingPageFilterable"

	"github.com/gorilla/mux"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/zebedee/data"
	"github.com/ONSdigital/go-ns/zebedee/zebedeeMapper"
)

const dataEndpoint = `\/data$`

// FilterClient is an interface with the methods required for a filter client
type FilterClient interface {
	healthcheck.Client
	CreateBlueprint(ctx context.Context, datasetID, edition, version string, names []string) (string, error)
	AddDimension(ctx context.Context, id, name string) error
	AddDimensionValue(ctx context.Context, filterID, name, value string) error
}

// DatasetClient is an interface with methods required for a dataset client
type DatasetClient interface {
	healthcheck.Client
	Get(ctx context.Context, id string) (m dataset.Model, err error)
	GetByPath(ctx context.Context, path string) (m dataset.Model, err error)
	GetEditions(ctx context.Context, id string) (m []dataset.Edition, err error)
	GetEdition(ctx context.Context, id, edition string) (dataset.Edition, error)
	GetVersions(ctx context.Context, id, edition string) (m []dataset.Version, err error)
	GetVersion(ctx context.Context, id, edition, version string) (m dataset.Version, err error)
	GetVersionMetadata(ctx context.Context, id, edition, version string) (m dataset.Metadata, err error)
	GetDimensions(ctx context.Context, id, edition, version string) (m dataset.Dimensions, err error)
	GetOptions(ctx context.Context, id, edition, version, dimension string) (m dataset.Options, err error)
}

// RenderClient is an interface with methods for require for rendering a template
type RenderClient interface {
	healthcheck.Client
	Do(string, []byte) ([]byte, error)
}

// ClientError is an interface that can be used to retrieve the status code if a client has errored
type ClientError interface {
	error
	Code() int
}

func setStatusCode(req *http.Request, w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if err, ok := err.(ClientError); ok {
		if err.Code() == http.StatusNotFound {
			status = err.Code()
		}
	}
	log.ErrorCtx(req.Context(), err, log.Data{"setting-response-status": status})
	w.WriteHeader(status)
}

func forwardFlorenceTokenIfRequired(req *http.Request) *http.Request {
	if len(req.Header.Get(common.FlorenceHeaderKey)) > 0 {
		ctx := common.SetFlorenceIdentity(req.Context(), req.Header.Get(common.FlorenceHeaderKey))
		return req.WithContext(ctx)
	}
	return req
}

// CreateFilterID controls the creating of a filter idea when a new user journey is
// requested
func CreateFilterID(c FilterClient, dc DatasetClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		datasetID := vars["datasetID"]
		edition := vars["editionID"]
		version := vars["versionID"]

		req = forwardFlorenceTokenIfRequired(req)

		dimensions, err := dc.GetDimensions(req.Context(), datasetID, edition, version)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}

		var names []string
		for _, dim := range dimensions.Items {
			opts, err := dc.GetOptions(req.Context(), datasetID, edition, version, dim.Name)
			if err != nil {
				setStatusCode(req, w, err)
				return
			}

			if len(opts.Items) > 1 { // If there is only one option then it can't be filterable so don't add to filter api
				names = append(names, dim.Name)
			}
		}

		fid, err := c.CreateBlueprint(req.Context(), datasetID, edition, version, names)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}

		log.InfoCtx(req.Context(), "created filter id", log.Data{"filter_id": fid})
		http.Redirect(w, req, "/filters/"+fid+"/dimensions", 301)
	}
}

// LegacyLanding will load a zebedee landing page
func LegacyLanding(zc ZebedeeClient, dc DatasetClient, rend RenderClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		cfg := config.Get()
		legacyLanding(w, req, zc, dc, rend, cfg)
	}
}

// FilterableLanding will load a filterable landing page
func FilterableLanding(dc DatasetClient, rend RenderClient, zc ZebedeeClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		cfg := config.Get()
		filterableLanding(w, req, dc, rend, zc, cfg)
	}
}

// EditionsList will load a list of editions for a filterable dataset
func EditionsList(dc DatasetClient, rend RenderClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		cfg := config.Get()
		editionsList(w, req, dc, rend, cfg)
	}
}

// VersionsList will load a list of versions for a filterable datase
func VersionsList(dc DatasetClient, rend RenderClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		cfg := config.Get()
		versionsList(w, req, dc, rend, cfg)
	}
}

func versionsList(w http.ResponseWriter, req *http.Request, dc DatasetClient, rend RenderClient, cfg config.Config) {
	vars := mux.Vars(req)
	datasetID := vars["datasetID"]
	edition := vars["edition"]

	req = forwardFlorenceTokenIfRequired(req)

	d, err := dc.Get(req.Context(), datasetID)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	versions, err := dc.GetVersions(req.Context(), datasetID, edition)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	e, err := dc.GetEdition(req.Context(), datasetID, edition)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	p := mapper.CreateVersionsList(req.Context(), d, e, versions)
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

func filterableLanding(w http.ResponseWriter, req *http.Request, dc DatasetClient, rend RenderClient, zc ZebedeeClient, cfg config.Config) {
	vars := mux.Vars(req)
	datasetID := vars["datasetID"]
	edition := vars["editionID"]
	version := vars["versionID"]

	req = forwardFlorenceTokenIfRequired(req)

	datasetModel, err := dc.Get(req.Context(), datasetID)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	if c, err := req.Cookie("access_token"); err == nil && len(c.Value) > 0 {
		zc.SetAccessToken(c.Value)
	}
	bc, err := zc.GetBreadcrumb(datasetModel.URI)
	if err != nil {
		log.ErrorCtx(req.Context(), err, log.Data{"Getting breadcrumb for dataset URI": datasetModel.URI})
	}

	if len(bc) > 0 {
		bc = append(bc, data.Breadcrumb{
			Description: data.NodeDescription{
				Title: datasetModel.Title,
			},
			URI: datasetModel.Links.Self.URL,
		})
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

	allVers, err := dc.GetVersions(req.Context(), datasetID, edition)
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

	latestVersionOfEditionURL := fmt.Sprintf("/datasets/%s/editions/%s/versions/%s", datasetID, edition, strconv.Itoa(latestVersionNumber))

	ver, err := dc.GetVersion(req.Context(), datasetID, edition, version)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	dims, err := dc.GetDimensions(req.Context(), datasetID, edition, version)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	var opts []dataset.Options
	for _, dim := range dims.Items {
		opt, err := dc.GetOptions(req.Context(), datasetID, edition, version, dim.Name)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}

		opts = append(opts, opt)
	}

	metadata, err := dc.GetVersionMetadata(req.Context(), datasetID, edition, version)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	textBytes, err := getText(dc, datasetID, edition, version, metadata, dims, req)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	if ver.Downloads == nil {
		ver.Downloads = make(map[string]dataset.Download)
	}

	m := mapper.CreateFilterableLandingPage(req.Context(), datasetModel, ver, datasetID, opts, dims, displayOtherVersionsLink, bc, latestVersionNumber, latestVersionOfEditionURL)

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

func editionsList(w http.ResponseWriter, req *http.Request, dc DatasetClient, rend RenderClient, cfg config.Config) {
	vars := mux.Vars(req)
	datasetID := vars["datasetID"]

	req = forwardFlorenceTokenIfRequired(req)

	datasetModel, err := dc.Get(req.Context(), datasetID)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	datasetEditions, err := dc.GetEditions(req.Context(), datasetID)
	if err != nil {
		if err, ok := err.(ClientError); ok {
			if err.Code() != http.StatusNotFound {
				setStatusCode(req, w, err)
				return
			}
		}
	}

	numberOfEditions := len(datasetEditions)
	if numberOfEditions == 1 {
		var latestVersionURL, err = url.Parse(datasetEditions[0].Links.LatestVersion.URL)
		if err != nil {
			log.Error(err, nil)
		} else {
			log.Info("only one edition, therefore redirecting to latest version", log.Data{"latestVersionPath": latestVersionURL.Path})
			http.Redirect(w, req, latestVersionURL.Path, 302)
		}
	}

	m := mapper.CreateEditionsList(req.Context(), datasetModel, datasetEditions, datasetID)

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

func legacyLanding(w http.ResponseWriter, req *http.Request, zc ZebedeeClient, dc DatasetClient, rend RenderClient, cfg config.Config) {
	if c, err := req.Cookie("access_token"); err == nil && len(c.Value) > 0 {
		zc.SetAccessToken(c.Value)
	}

	path := req.URL.Path

	// Since MatchString will only error if the regex is invalid, and the regex is
	// constant, don't capture the error
	if ok, _ := regexp.MatchString(dataEndpoint, path); ok {
		b, err := zc.Get("/data?uri=" + path)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}
		w.Write(b)
		return
	}

	dlp, err := zc.GetDatasetLandingPage("/data?uri=" + path)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	bc, err := zc.GetBreadcrumb(dlp.URI)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	var ds []data.Dataset
	for _, v := range dlp.Datasets {
		d, err := zc.GetDataset(v.URI)
		if err != nil {
			setStatusCode(req, w, errors.Wrap(err, "zebedee client legacy dataset returned an error"))
			return
		}
		ds = append(ds, d)
	}

	// Check for filterable datasets and fetch details
	if len(dlp.RelatedFilterableDatasets) > 0 {
		var relatedFilterableDatasets []data.Related
		var wg sync.WaitGroup
		var mutex = &sync.Mutex{}
		for _, relatedFilterableDataset := range dlp.RelatedFilterableDatasets {
			wg.Add(1)
			go func(ctx context.Context, dc DatasetClient, relatedFilterableDataset data.Related) {
				defer wg.Done()
				d, err := dc.GetByPath(ctx, relatedFilterableDataset.URI)
				if err != nil {
					// log error but continue to map data. any datasets that fail won't get mapped and won't be displayed on frontend
					log.ErrorCtx(req.Context(), errors.WithMessage(err, "error fetching dataset details"), log.Data{
						"dataset": relatedFilterableDataset.URI,
					})
					return
				}
				mutex.Lock()
				defer mutex.Unlock()
				relatedFilterableDatasets = append(relatedFilterableDatasets, data.Related{Title: d.Title, URI: relatedFilterableDataset.URI})
				return
			}(req.Context(), dc, relatedFilterableDataset)
		}
		wg.Wait()
		dlp.RelatedFilterableDatasets = relatedFilterableDatasets
	}

	m := zebedeeMapper.MapZebedeeDatasetLandingPageToFrontendModel(dlp, bc, ds)

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
func MetadataText(dc DatasetClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		metadataText(w, req, dc)
	}
}

func metadataText(w http.ResponseWriter, req *http.Request, dc DatasetClient) {
	vars := mux.Vars(req)
	datasetID := vars["datasetID"]
	edition := vars["edition"]
	version := vars["version"]

	req = forwardFlorenceTokenIfRequired(req)

	metadata, err := dc.GetVersionMetadata(req.Context(), datasetID, edition, version)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	dimensions, err := dc.GetDimensions(req.Context(), datasetID, edition, version)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	b, err := getText(dc, datasetID, edition, version, metadata, dimensions, req)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	w.Header().Set("Content-Type", "plain/text")

	w.Write(b)

}

func getText(dc DatasetClient, datasetID, edition, version string, metadata dataset.Metadata, dimensions dataset.Dimensions, req *http.Request) ([]byte, error) {
	var b bytes.Buffer

	b.WriteString(metadata.ToString())
	b.WriteString("Dimensions:\n")
	for _, dimension := range dimensions.Items {
		options, err := dc.GetOptions(req.Context(), datasetID, edition, version, dimension.Name)
		if err != nil {
			return nil, err
		}

		b.WriteString(options.String())
	}
	return b.Bytes(), nil
}
