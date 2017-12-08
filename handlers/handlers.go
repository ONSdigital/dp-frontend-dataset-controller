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
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/zebedee/data"
	"github.com/ONSdigital/go-ns/zebedee/zebedeeMapper"
)

const dataEndpoint = `\/data$`

// FilterClient is an interface with the methods required for a filter client
type FilterClient interface {
	healthcheck.Client
	CreateBlueprint(datasetFilterID string, names []string) (string, error)
	AddDimension(id, name string) error
	AddDimensionValue(filterID, name, value string) error
}

// DatasetClient is an interface with methods required for a dataset client
type DatasetClient interface {
	healthcheck.Client
	Get(id string) (m dataset.Model, err error)
	GetEditions(id string) (m []dataset.Edition, err error)
	GetEdition(id, edition string) (dataset.Edition, error)
	GetVersions(id, edition string) (m []dataset.Version, err error)
	GetVersion(id, edition, version string) (m dataset.Version, err error)
	GetVersionMetadata(id, edition, version string) (m dataset.Metadata, err error)
	GetDimensions(id, edition, version string) (m dataset.Dimensions, err error)
	GetOptions(id, edition, version, dimension string) (m dataset.Options, err error)
}

// RenderClient is an interface with methods for require for rendering a template
type RenderClient interface {
	healthcheck.Client
	Do(string, []byte) ([]byte, error)
}

// CreateFilterID controls the creating of a filter idea when a new user journey is
// requested
func CreateFilterID(c FilterClient, dc DatasetClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		datasetID := vars["datasetID"]
		edition := vars["editionID"]
		version := vars["versionID"]

		datasetModel, err := dc.GetVersion(datasetID, edition, version)
		if err != nil {
			log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		dimensions, err := dc.GetDimensions(datasetID, edition, version)
		if err != nil {
			log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var names []string
		for _, dim := range dimensions.Items {
			opts, err := dc.GetOptions(datasetID, edition, version, dim.ID)
			if err != nil {
				log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if len(opts.Items) > 1 { // If there is only one option then it can't be filterable so don't add to filter api
				names = append(names, dim.ID)
			}
		}

		fid, err := c.CreateBlueprint(datasetModel.ID, names)
		if err != nil {
			log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
			w.WriteHeader(http.StatusInternalServerError)
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
func FilterableLanding(dc DatasetClient, rend RenderClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		cfg := config.Get()
		filterableLanding(w, req, dc, rend, cfg)
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

	d, err := dc.Get(datasetID)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	versions, err := dc.GetVersions(datasetID, edition)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	e, err := dc.GetEdition(datasetID, edition)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	p := mapper.CreateVersionsList(d, e, versions)
	b, err := json.Marshal(p)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateHTML, err := rend.Do("dataset-version-list", b)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateHTML)
}

func filterableLanding(w http.ResponseWriter, req *http.Request, dc DatasetClient, rend RenderClient, cfg config.Config) {
	vars := mux.Vars(req)
	datasetID := vars["datasetID"]
	edition := vars["editionID"]
	version := vars["versionID"]

	datasetModel, err := dc.Get(datasetID)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(edition) == 0 {
		latestVersionURL, err := url.Parse(datasetModel.Links.LatestVersion.URL)
		if err != nil {
			log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, edition, version, err = helpers.ExtractDatasetInfoFromPath(latestVersionURL.Path)
		if err != nil {
			log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	allVers, err := dc.GetVersions(datasetID, edition)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var displayOtherVersionsLink bool
	if len(allVers) > 1 {
		displayOtherVersionsLink = true
	}

	ver, err := dc.GetVersion(datasetID, edition, version)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dims, err := dc.GetDimensions(datasetID, edition, version)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var opts []dataset.Options
	for _, dim := range dims.Items {
		opt, err := dc.GetOptions(datasetID, edition, version, dim.ID)
		if err != nil {
			log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		opts = append(opts, opt)
	}

	metadata, err := dc.GetVersionMetadata(datasetID, edition, version)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	textBytes, err := getText(dc, datasetID, edition, version, metadata, dims)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if ver.Downloads == nil {
		ver.Downloads = make(map[string]dataset.Download)
	}

	ver.Downloads["Text"] = dataset.Download{
		Size: strconv.Itoa(len(textBytes)),
		URL:  fmt.Sprintf("/datasets/%s/editions/%s/versions/%s/metadata.txt", datasetID, edition, version),
	}

	m := mapper.CreateFilterableLandingPage(datasetModel, ver, datasetID, opts, dims, displayOtherVersionsLink)

	b, err := json.Marshal(m)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateHTML, err := rend.Do("dataset-landing-page-filterable", b)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateHTML)

}

func editionsList(w http.ResponseWriter, req *http.Request, dc DatasetClient, rend RenderClient, cfg config.Config) {
	vars := mux.Vars(req)
	datasetID := vars["datasetID"]

	datasetModel, err := dc.Get(datasetID)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	datasetEditions, err := dc.GetEditions(datasetID)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	m := mapper.CreateEditionsList(datasetModel, datasetEditions, datasetID)

	b, err := json.Marshal(m)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateHTML, err := rend.Do("dataset-edition-list", b)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
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
			log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(b)
		return
	}

	dlp, err := zc.GetDatasetLandingPage("/data?uri=" + path)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bc, err := zc.GetBreadcrumb(dlp.URI)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
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
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateHTML, err := rend.Do("dataset-landing-page-static", templateJSON)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
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

	metadata, err := dc.GetVersionMetadata(datasetID, edition, version)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dimensions, err := dc.GetDimensions(datasetID, edition, version)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := getText(dc, datasetID, edition, version, metadata, dimensions)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "plain/text")

	w.Write(b)

}

func getText(dc DatasetClient, datasetID, edition, version string, metadata dataset.Metadata, dimensions dataset.Dimensions) ([]byte, error) {
	var b bytes.Buffer

	b.WriteString(metadata.String())
	b.WriteString("Dimensions:\n")
	for _, dimension := range dimensions.Items {
		options, err := dc.GetOptions(datasetID, edition, version, dimension.ID)
		if err != nil {
			return nil, err
		}

		b.WriteString(options.String())
	}
	return b.Bytes(), nil
}
