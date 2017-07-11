package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/zebedee/client"
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

// CreateJobID controls the creating of a job idea when a new user journey is
// requested
func CreateJobID(w http.ResponseWriter, req *http.Request) {
	// TODO: This is a stubbed job id - replace with real job id from api once
	// code has been written
	jobID := rand.Intn(100000000)
	jid := strconv.Itoa(jobID)

	log.Trace("created job id", log.Data{"job_id": jid})
	http.Redirect(w, req, "/jobs/"+jid+"/dimensions", 301)
}

// LegacyLanding will load a zebedee landing page
func LegacyLanding(w http.ResponseWriter, req *http.Request) {
	cfg := config.Get()
	zc := client.NewZebedeeClient(cfg.ZebedeeURL)
	legacyLanding(w, req, zc, cfg)
}

// FilterableLanding ..
func FilterableLanding(w http.ResponseWriter, req *http.Request) {
	cfg := config.Get()

	b, err := render([]byte(`{}`), "filterable", cfg)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
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

	//Marshal template data to JSON
	templateJSON, err := json.Marshal(m)
	if err != nil {
		log.Error(err, nil)
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
