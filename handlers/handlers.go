package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"

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

var cli *http.Client

const dataEndpoint = `\/data$`

func init() {
	if cli == nil {
		cli = &http.Client{Timeout: 5 * time.Second}
	}
}

// FilterClient is an interface with the methods required for a filter client
type FilterClient interface {
	healthcheck.Client
	CreateJob(datasetFilterID string) (string, error)
	AddDimension(id, name string) error
	AddDimensionValue(filterID, name, value string) error
}

// DatasetClient is an interface with methods required for a dataset client
type DatasetClient interface {
	healthcheck.Client
	Get(id string) (m dataset.Model, err error)
	GetEditions(id string) (m []dataset.Edition, err error)
	GetVersions(id, edition string) (m []dataset.Version, err error)
	GetVersion(id, edition, version string) (m dataset.Version, err error)
	GetDimensions(id, edition, version string) (m dataset.Dimensions, err error)
	GetOptions(id, edition, version, dimension string) (m dataset.Options, err error)
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
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fid, err := c.CreateJob(datasetModel.InstanceID)
		if err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		dimensions, err := dc.GetDimensions(datasetID, edition, version)
		if err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for _, dim := range dimensions.Items {
			opts, err := dc.GetOptions(datasetID, edition, version, dim.ID)
			if err != nil {
				log.ErrorR(req, err, nil)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if len(opts.Items) > 1 { // If there is only one option then it can't be filterable so don't add to filter api
				if err = c.AddDimension(fid, dim.ID); err != nil {
					log.ErrorR(req, err, nil)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}

		}

		log.Trace("created filter id", log.Data{"filter_id": fid})
		http.Redirect(w, req, "/filters/"+fid+"/dimensions", 301)
	}
}

// LegacyLanding will load a zebedee landing page
func LegacyLanding(zc ZebedeeClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		cfg := config.Get()
		legacyLanding(w, req, zc, cfg)
	}
}

// FilterableLanding will load a filterable landing page
func FilterableLanding(dc DatasetClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		cfg := config.Get()
		filterableLanding(w, req, dc, cfg)
	}
}

func filterableLanding(w http.ResponseWriter, req *http.Request, dc DatasetClient, cfg config.Config) {
	vars := mux.Vars(req)
	datasetID := vars["datasetID"]

	datasetModel, err := dc.Get(datasetID)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	datasetEditions, err := dc.GetEditions(datasetID)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var datasetVersions []dataset.Version
	for _, ed := range datasetEditions {
		editionVersions, err := dc.GetVersions(datasetID, ed.Edition)
		if err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		datasetVersions = append(datasetVersions, editionVersions...)
	}

	latestVersionURL, err := url.Parse(datasetModel.Links.LatestVersion.URL)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, edition, version, err := helpers.ExtractDatasetInfoFromPath(latestVersionURL.Path)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dims, err := dc.GetDimensions(datasetID, edition, version)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var opts []dataset.Options
	for _, dim := range dims.Items {
		opt, err := dc.GetOptions(datasetID, edition, version, dim.ID)
		if err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		opts = append(opts, opt)
	}

	m := mapper.CreateFilterableLandingPage(datasetModel, datasetVersions, datasetID, opts)

	b, err := json.Marshal(m)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateHTML, err := render(b, "filterable", cfg)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateHTML)

}

func legacyLanding(w http.ResponseWriter, req *http.Request, zc ZebedeeClient, cfg config.Config) {
	if c, err := req.Cookie("access_token"); err == nil && len(c.Value) > 0 {
		zc.SetAccessToken(c.Value)
	}

	path := req.URL.Path

	// Since MatchString will only error if the regex is invalid, and the regex is
	// constant, don't capture the error
	if ok, _ := regexp.MatchString(dataEndpoint, path); ok {
		b, err := zc.Get("/data?uri=" + path)
		if err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(b)
		return
	}

	dlp, err := zc.GetDatasetLandingPage("/data?uri=" + path)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bc, err := zc.GetBreadcrumb(dlp.URI)
	if err != nil {
		log.ErrorR(req, err, nil)
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
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templateHTML, err := render(templateJSON, m.FilterID, cfg)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateHTML)
	return

}

func render(data []byte, filterID string, cfg config.Config) ([]byte, error) {
	rdr := bytes.NewReader(data)

	var rendererReq *http.Request
	var err error
	if filterID == "" {
		rendererReq, err = http.NewRequest("POST", cfg.RendererURL+"/dataset-landing-page-static", rdr)
	} else {
		rendererReq, err = http.NewRequest("POST", cfg.RendererURL+"/dataset-landing-page-filterable", rdr)
	}
	if err != nil {
		return nil, err
	}

	rendererRes, err := cli.Do(rendererReq)
	if err != nil {
		return nil, err
	}
	defer rendererRes.Body.Close()

	if rendererRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response from renderer service: %d", rendererRes.StatusCode)
	}

	return ioutil.ReadAll(rendererRes.Body)
}
