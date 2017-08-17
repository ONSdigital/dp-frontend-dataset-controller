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
	filterdata "github.com/ONSdigital/dp-frontend-dataset-controller/data"
	"github.com/ONSdigital/dp-frontend-dataset-controller/filter"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
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

// CreateFilterID controls the creating of a filter idea when a new user journey is
// requested
func CreateFilterID(fil *filter.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		fid, err := fil.CreateJob("ds83jks0-23euufr89-8i3") //use a real dataset filter id
		if err != nil {
			log.ErrorR(req, err, nil)
			return
		}

		fil.AddDimension(fid, "time")
		fil.AddDimension(fid, "goods-and-services")

		log.Trace("created filter id", log.Data{"filter_id": fid})
		http.Redirect(w, req, "/filters/"+fid+"/dimensions", 301)
	}
}

// LegacyLanding will load a zebedee landing page
func LegacyLanding(w http.ResponseWriter, req *http.Request) {
	cfg := config.Get()
	zc := client.NewZebedeeClient(cfg.ZebedeeURL)
	landing(w, req, zc, cfg)
}

// FilterableLanding ..
func FilterableLanding(w http.ResponseWriter, req *http.Request) {
	cfg := config.Get()
	zc := client.NewZebedeeClient(cfg.ZebedeeURL)
	landing(w, req, zc, cfg)
}

func landing(w http.ResponseWriter, req *http.Request, zc ZebedeeClient, cfg config.Config) {
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

	m.FilterID = "filterable" // TODO: This information should be coming from zebedee but isn't implemented as of yet
	m.DatasetLandingPage.DatasetID = "12345"

	var templateJSON []byte

	if m.FilterID == "filterable" {

		// TODO: This information will be coming from the dataset api when it is implemented
		datasets := []filterdata.Dataset{
			{
				ID:          "12345",
				Title:       "",
				URL:         "",
				ReleaseDate: "11 November 2017",
				NextRelease: "11 November 2018",
				Edition:     "2017",
				Version:     "1",
				Contact: filterdata.Contact{
					Name:      "Matt Rout",
					Telephone: "012346 382012",
					Email:     "matt@gmail.com",
				},
			},
		}

		dimensions := []filterdata.Dimension{
			{
				CodeListID: "ABDCSKA",
				ID:         "siojxuidhc",
				Name:       "Geography",
				Type:       "Hierarchy",
				Values:     []string{"Region", "County"},
			},
			{
				CodeListID: "AHDHSID",
				ID:         "eorihfieorf",
				Name:       "Age List",
				Type:       "List",
				Values:     []string{"0", "1", "2"},
			},
		}

		fp := mapper.CreateFilterableLandingPage(datasets, dimensions, m)

		templateJSON, err = json.Marshal(fp)
		if err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {

		//Marshal template data to JSON
		templateJSON, err = json.Marshal(m)
		if err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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
