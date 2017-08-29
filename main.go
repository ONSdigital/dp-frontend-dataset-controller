package main

import (
	"os"
	"time"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/handlers"
	"github.com/ONSdigital/go-ns/clients/filter"
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/ONSdigital/go-ns/zebedee/client"
	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Get()
	log.Namespace = "frontend-dataset-controller"

	router := mux.NewRouter()

	f := filter.New(cfg.FilterAPIURL)
	zc := client.NewZebedeeClient(cfg.ZebedeeURL)

	go func() {
		for {
			timer := time.NewTimer(time.Second * 60)

			healthcheck.MonitorExternal(f, zc)

			<-timer.C
		}
	}()

	router.Path("/healthcheck").HandlerFunc(healthcheck.Do)

	router.Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}").Methods("GET").HandlerFunc(handlers.FilterableLanding(zc))
	router.Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter").Methods("POST").HandlerFunc(handlers.CreateFilterID(f))

	router.HandleFunc("/{uri:.*}", handlers.LegacyLanding(zc))

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
