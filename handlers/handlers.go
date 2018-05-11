package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"

	"github.com/gorilla/mux"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/clients/filter"
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/zebedee/data"
	"github.com/ONSdigital/go-ns/zebedee/zebedeeMapper"
)

const dataEndpoint = `\/data$`

// FilterClient is an interface with the methods required for a filter client
type FilterClient interface {
	healthcheck.Client
	CreateBlueprint(datasetID, edition, version string, names []string, cfg ...filter.Config) (string, error)
	AddDimension(id, name string, cfg ...filter.Config) error
	AddDimensionValue(filterID, name, value string, cfg ...filter.Config) error
}

// DatasetClient is an interface with methods required for a dataset client
type DatasetClient interface {
	healthcheck.Client
	Get(id string, cfg ...dataset.Config) (m dataset.Model, err error)
	GetEditions(id string, cfg ...dataset.Config) (m []dataset.Edition, err error)
	GetEdition(id, edition string, cfg ...dataset.Config) (dataset.Edition, error)
	GetVersions(id, edition string, cfg ...dataset.Config) (m []dataset.Version, err error)
	GetVersion(id, edition, version string, cfg ...dataset.Config) (m dataset.Version, err error)
	GetVersionMetadata(id, edition, version string, cfg ...dataset.Config) (m dataset.Metadata, err error)
	GetDimensions(id, edition, version string, cfg ...dataset.Config) (m dataset.Dimensions, err error)
	GetOptions(id, edition, version, dimension string, cfg ...dataset.Config) (m dataset.Options, err error)
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
	log.ErrorR(req, err, log.Data{"setting-response-status": status})
	w.WriteHeader(status)
}

func setAuthTokenIfRequired(req *http.Request) ([]dataset.Config, []filter.Config) {
	var datasetConfig []dataset.Config
	var filterConfig []filter.Config
	florenceToken := req.Header.Get("X-Florence-Token")
	if len(florenceToken) > 0 {
		cfg := config.Get()
		datasetConfig = append(datasetConfig, dataset.Config{InternalToken: cfg.DatasetAPIAuthToken, FlorenceToken: florenceToken})
		filterConfig = append(filterConfig, filter.Config{InternalToken: cfg.FilterAPIAuthToken, FlorenceToken: florenceToken})
	}
	return datasetConfig, filterConfig
}

// CreateFilterID controls the creating of a filter idea when a new user journey is
// requested
func CreateFilterID(c FilterClient, dc DatasetClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		datasetID := vars["datasetID"]
		edition := vars["editionID"]
		version := vars["versionID"]

		datasetCfg, filterConfig := setAuthTokenIfRequired(req)

		dimensions, err := dc.GetDimensions(datasetID, edition, version, datasetCfg...)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}

		var names []string
		for _, dim := range dimensions.Items {
			opts, err := dc.GetOptions(datasetID, edition, version, dim.ID, datasetCfg...)
			if err != nil {
				setStatusCode(req, w, err)
				return
			}

			if len(opts.Items) > 1 { // If there is only one option then it can't be filterable so don't add to filter api
				names = append(names, dim.ID)
			}
		}

		fid, err := c.CreateBlueprint(datasetID, edition, version, names, filterConfig...)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}

		log.Trace("created filter id", log.Data{"filter_id": fid})
		http.Redirect(w, req, "/filters/"+fid+"/dimensions", 301)
	}
}

// LegacyLanding will load a zebedee landing page
func LegacyLanding(zc ZebedeeClient, rend RenderClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		cfg := config.Get()
		legacyLanding(w, req, zc, rend, cfg)
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

	datasetCfg, _ := setAuthTokenIfRequired(req)

	d, err := dc.Get(datasetID, datasetCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	versions, err := dc.GetVersions(datasetID, edition, datasetCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	e, err := dc.GetEdition(datasetID, edition, datasetCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	p := mapper.CreateVersionsList(d, e, versions)
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

	datasetCfg, _ := setAuthTokenIfRequired(req)

	datasetModel, err := dc.Get(datasetID, datasetCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	if c, err := req.Cookie("access_token"); err == nil && len(c.Value) > 0 {
		zc.SetAccessToken(c.Value)
	}
	bc, err := zc.GetBreadcrumb(datasetModel.URI)
	if err != nil {
		log.ErrorR(req, err, log.Data{"Getting breadcrumb for dataset URI": datasetModel.URI})
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

	allVers, err := dc.GetVersions(datasetID, edition, datasetCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	var displayOtherVersionsLink bool
	if len(allVers) > 1 {
		displayOtherVersionsLink = true
	}

	ver, err := dc.GetVersion(datasetID, edition, version, datasetCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	dims, err := dc.GetDimensions(datasetID, edition, version, datasetCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	var opts []dataset.Options
	for _, dim := range dims.Items {
		opt, err := dc.GetOptions(datasetID, edition, version, dim.ID, datasetCfg...)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}

		opts = append(opts, opt)
	}

	metadata, err := dc.GetVersionMetadata(datasetID, edition, version, datasetCfg...)
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

	ver.Downloads["Text"] = dataset.Download{
		Size: strconv.Itoa(len(textBytes)),
		URL:  fmt.Sprintf("/datasets/%s/editions/%s/versions/%s/metadata.txt", datasetID, edition, version),
	}

	m := mapper.CreateFilterableLandingPage(datasetModel, ver, datasetID, opts, dims, displayOtherVersionsLink, bc)

	for _, d := range m.DatasetLandingPage.Version.Downloads {
		if len(cfg.DownloadServiceURL) > 0 {
			downloadURL, err := url.Parse(d.URI)
			if err != nil {
				setStatusCode(req, w, err)
				return
			}

			d.URI = cfg.DownloadServiceURL + downloadURL.Path
		}
	}

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

	datasetCfg, _ := setAuthTokenIfRequired(req)

	datasetModel, err := dc.Get(datasetID, datasetCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	datasetEditions, err := dc.GetEditions(datasetID, datasetCfg...)
	if err != nil {
		if err, ok := err.(ClientError); ok {
			if err.Code() != http.StatusNotFound {
				setStatusCode(req, w, err)
				return
			}
		}
	}

	m := mapper.CreateEditionsList(datasetModel, datasetEditions, datasetID)

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

func legacyLanding(w http.ResponseWriter, req *http.Request, zc ZebedeeClient, rend RenderClient, cfg config.Config) {
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
		d, _ := zc.GetDataset(v.URI)
		ds = append(ds, d)
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

	datasetCfg, _ := setAuthTokenIfRequired(req)

	metadata, err := dc.GetVersionMetadata(datasetID, edition, version, datasetCfg...)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	dimensions, err := dc.GetDimensions(datasetID, edition, version)
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

	datasetCfg, _ := setAuthTokenIfRequired(req)

	b.WriteString(metadata.String())
	b.WriteString("Dimensions:\n")
	for _, dimension := range dimensions.Items {
		options, err := dc.GetOptions(datasetID, edition, version, dimension.ID, datasetCfg...)
		if err != nil {
			return nil, err
		}

		b.WriteString(options.String())
	}
	return b.Bytes(), nil
}
