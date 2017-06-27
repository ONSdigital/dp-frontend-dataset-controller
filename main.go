package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/zebedeeModels"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/datasetLandingPageStatic"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/pat"
)

var cfg = config.Get()
var client = http.Client{
	Timeout: time.Duration(time.Second * 5),
}

func main() {
	log.Namespace = "frontend-dataset-controller"

	router := pat.New()

	router.HandleFunc("/{uri:.*}", handler)

	log.Debug("Starting server", log.Data{
		"bind_addr":    cfg.BindAddr,
		"zebedee_url":  cfg.ZebedeeURL,
		"renderer_url": cfg.RendererURL,
	})

	s := server.New(cfg.BindAddr, router)

	if err := s.ListenAndServe(); err != nil {
		log.Error(err, nil)
		os.Exit(2)
	}
}

func handler(w http.ResponseWriter, req *http.Request) {
	//FIXME need to do NewRequest and RawQuery
	res, err := http.Get(cfg.ZebedeeURL + "/data?uri=" + req.URL.Path)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(500)
		return
	}

	defer res.Body.Close()

	JSONBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(500)
		return
	}

	var pageJSON zebedeeModels.DatasetLandingPage
	err = json.Unmarshal(JSONBody, &pageJSON)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(500)
		return
	}

	//Concurrently resolve any URIs where we need more data from another page
	var wg sync.WaitGroup
	sem := make(chan int, 10)
	related := [][]model.Related{
		pageJSON.RelatedDatasets,
		pageJSON.RelatedDocuments,
		pageJSON.RelatedMethodology,
		pageJSON.RelatedMethodologyArticle,
	}

	for _, element := range related {
		for i, e := range element {
			sem <- 1
			wg.Add(1)
			go func(i int, e model.Related, element []model.Related) {
				defer func() {
					<-sem
					wg.Done()
				}()
				element[i].Title = getPageTitle(req, e.URI)
			}(i, e, element)
		}
	}
	wg.Wait()

	//Map Zebedee response data to new page model
	var templateData datasetLandingPageStatic.Page
	templateData.Type = pageJSON.Type
	templateData.URI = pageJSON.URI
	templateData.Metadata.Title = pageJSON.Description.Title
	templateData.Metadata.Description = pageJSON.Description.Summary
	templateData.DatasetLandingPage.Related.Datasets = pageJSON.RelatedDatasets
	templateData.DatasetLandingPage.Related.Publications = pageJSON.RelatedDocuments
	templateData.DatasetLandingPage.Related.Methodology = append(pageJSON.RelatedMethodology, pageJSON.RelatedMethodologyArticle...)
	templateData.DatasetLandingPage.Related.Links = pageJSON.RelatedLinks
	templateData.DatasetLandingPage.IsNationalStatistic = pageJSON.Description.NationalStatistic
	templateData.DatasetLandingPage.IsTimeseries = pageJSON.Timeseries
	templateData.ContactDetails.Email = pageJSON.Description.Contact.Email
	templateData.ContactDetails.Telephone = pageJSON.Description.Contact.Telephone
	templateData.ContactDetails.Name = pageJSON.Description.Contact.Name
	templateData.DatasetLandingPage.ReleaseDate = pageJSON.Description.ReleaseDate
	templateData.DatasetLandingPage.NextRelease = pageJSON.Description.NextRelease
	templateData.DatasetLandingPage.Notes = pageJSON.Section.Markdown

	for index, value := range pageJSON.Datasets {
		dataset := getDatasetDetails(req, value.URI)
		dataset.IsLast = index+1 == len(pageJSON.Datasets)

		templateData.DatasetLandingPage.Datasets = append(templateData.DatasetLandingPage.Datasets, dataset)
	}

	for _, value := range pageJSON.Alerts {
		switch value.Type {
		default:
			log.Debug("Unrecognised alert type", log.Data{"alert": value})
			fallthrough
		case "alert":
			templateData.DatasetLandingPage.Notices = append(templateData.DatasetLandingPage.Notices, datasetLandingPageStatic.Message{
				Date:     value.Date,
				Markdown: value.Markdown,
			})
		case "correction":
			templateData.DatasetLandingPage.Corrections = append(templateData.DatasetLandingPage.Corrections, datasetLandingPageStatic.Message{
				Date:     value.Date,
				Markdown: value.Markdown,
			})
		}
	}

	//Marshal template data to JSON
	templateJSON, err := json.Marshal(templateData)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(500)
		return
	}

	rdr := bytes.NewReader(templateJSON)

	var rendererReq *http.Request
	var datasetType string

	if pageJSON.FilterID == "" {
		datasetType = "static"
		rendererReq, err = http.NewRequest("POST", cfg.RendererURL+"/dataset-landing-page-static", rdr)
	} else {
		datasetType = "filterable"
		rendererReq, err = http.NewRequest("POST", cfg.RendererURL+"/dataset-landing-page-filterable", rdr)
	}

	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(500)
		return
	}

	rendererRes, err := http.DefaultClient.Do(rendererReq)
	if err != nil {
		log.ErrorR(req, err, nil)
	}

	HTMLBody, err := ioutil.ReadAll(rendererRes.Body)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(500)
		return
	}

	log.Debug("Returning "+datasetType+" dataset landing page", nil)
	w.WriteHeader(res.StatusCode)
	w.Write(HTMLBody)
}

type pageTitle struct {
	Title   string `json:"title"`
	Edition string `json:"edition"`
}

func getPageTitle(req *http.Request, uri string) string {
	res, err := client.Get(cfg.ZebedeeURL + "/data?uri=" + uri + "&title")
	if err != nil {
		log.ErrorR(req, err, nil)
		return ""
	}

	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.ErrorR(req, err, nil)
		return ""
	}

	var pageTitle pageTitle
	err = json.Unmarshal(b, &pageTitle)
	if err != nil {
		log.ErrorR(req, err, nil)
		return ""
	}

	title := pageTitle.Title

	if len(pageTitle.Edition) > 0 && len(pageTitle.Title) > 0 {
		title = title + ": " + pageTitle.Edition
	}

	return title
}

func getDatasetDetails(req *http.Request, uri string) datasetLandingPageStatic.Dataset {
	res, err := client.Get(cfg.ZebedeeURL + "/data?uri=" + uri)
	if err != nil {
		log.ErrorR(req, err, nil)
		return datasetLandingPageStatic.Dataset{
			URI: uri,
		}
	}

	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.ErrorR(req, err, nil)
		return datasetLandingPageStatic.Dataset{
			URI: uri,
		}
	}

	var pageJSON zebedeeModels.Dataset
	err = json.Unmarshal(b, &pageJSON)
	if err != nil {
		log.ErrorR(req, err, nil)
		return datasetLandingPageStatic.Dataset{
			URI: uri,
		}
	}

	var dataset datasetLandingPageStatic.Dataset
	for _, value := range pageJSON.Downloads {
		dataset.Downloads = append(dataset.Downloads, datasetLandingPageStatic.Download{
			URI:       value.File,
			Extension: filepath.Ext(value.File),
			Size:      getFileSize(req, uri+"/"+value.File),
		})
	}
	dataset.Title = pageJSON.Description.Edition
	if len(pageJSON.SupplementaryFiles) > 0 {
		dataset.HasSupplementaryFiles = true
	}
	if len(pageJSON.Versions) > 0 {
		dataset.HasVersions = true
	}

	dataset.URI = pageJSON.URI

	return dataset
}

type fileSize struct {
	Size int `json:"fileSize"`
}

func getFileSize(req *http.Request, uri string) string {
	res, err := client.Get(cfg.ZebedeeURL + "/filesize?uri=" + uri)
	if err != nil {
		log.ErrorR(req, err, nil)
		return ""
	}

	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.ErrorR(req, err, nil)
		return ""
	}

	var fileSizeJSON fileSize
	err = json.Unmarshal(b, &fileSizeJSON)
	if err != nil {
		log.ErrorR(req, err, nil)
		return ""
	}

	var size string
	if fileSizeJSON.Size < 1000000 {
		size = strconv.Itoa(fileSizeJSON.Size/1000) + " kb"
	}

	return size
}
