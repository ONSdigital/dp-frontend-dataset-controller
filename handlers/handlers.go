package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/zebedee"
)

var client *http.Client

const dataEndpoint = `\/data$`

func init() {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
}

// LegacyLanding ...
func LegacyLanding(w http.ResponseWriter, req *http.Request) {
	cfg := config.Get()
	zc := zebedee.NewClient(cfg.ZebedeeURL)
	legacyLanding(w, req, zc, cfg)
}

func legacyLanding(w http.ResponseWriter, req *http.Request, zc zebedee.Client, cfg config.Config) {
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

	m, err := zc.GetLanding("/data?uri=" + path)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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

	rendererRes, err := client.Do(rendererReq)
	if err != nil {
		return nil, err
	}
	defer rendererRes.Body.Close()

	if rendererRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response from renderer service: %d", rendererRes.StatusCode)
	}

	return ioutil.ReadAll(rendererRes.Body)
}
