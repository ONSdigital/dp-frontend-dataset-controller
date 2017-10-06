package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/handlers"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/clients/filter"
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/ONSdigital/go-ns/zebedee/client"
	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Get()
	log.Debug("got service configuration", log.Data{"config": cfg})

	log.Namespace = "frontend-dataset-controller"

	router := mux.NewRouter()

	f := filter.New(cfg.FilterAPIURL)
	zc := client.NewZebedeeClient(cfg.ZebedeeURL)
	dc := dataset.New(cfg.DatasetAPIURL)

	router.Path("/healthcheck").HandlerFunc(healthcheck.Do)

	router.Path("/datasets/{datasetID}").Methods("GET").HandlerFunc(handlers.FilterableLanding(dc))
	router.Path("/datasets/{datasetID}/{uri:.*}").Methods("GET").HandlerFunc(handlers.FilterableLanding(dc))
	router.Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter").Methods("POST").HandlerFunc(handlers.CreateFilterID(f, dc))

	router.HandleFunc("/{uri:.*}", handlers.LegacyLanding(zc))

	log.Debug("Starting server", log.Data{
		"bind_addr":    cfg.BindAddr,
		"zebedee_url":  cfg.ZebedeeURL,
		"renderer_url": cfg.RendererURL,
	})

	s := server.New(cfg.BindAddr, router)
	s.HandleOSSignals = false

	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Error(err, nil)
			os.Exit(2)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	for {
		log.Debug("conducting service healthcheck", log.Data{
			"services": []string{
				"filter-api",
				"dataset-api",
				"zebedee",
			},
		})

		healthcheck.MonitorExternal(f, zc, dc)

		timer := time.NewTimer(time.Second * 60)

		select {
		case <-timer.C:
			continue
		case <-stop:
			log.Info("shutting service down gracefully", nil)
			timer.Stop()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := s.Server.Shutdown(ctx); err != nil {
				log.Error(err, nil)
			}
			return
		}
	}

}
