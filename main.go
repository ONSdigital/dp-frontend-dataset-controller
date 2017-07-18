package main

import (
	"os"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/handlers"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Get()
	log.Namespace = "frontend-dataset-controller"

	router := mux.NewRouter()

	router.Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}").Methods("GET").HandlerFunc(handlers.FilterableLanding)
	router.Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter").Methods("POST").HandlerFunc(handlers.CreateFilterID)

	router.HandleFunc("/{uri:.*}", handlers.LegacyLanding)

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
