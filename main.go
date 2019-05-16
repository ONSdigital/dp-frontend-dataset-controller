package main

import (
	"context"
	"fmt"
	"net/smtp"
	"os"
	"os/signal"
	"time"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/handlers"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/clients/filter"
	"github.com/ONSdigital/go-ns/clients/renderer"
	"github.com/ONSdigital/go-ns/handlers/healthcheck"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/ONSdigital/go-ns/zebedee/client"
	"github.com/gorilla/mux"
	"github.com/ONSdigital/go-ns/handlers/collectionID"
)

type unencryptedAuth struct {
	smtp.Auth
}

func (a unencryptedAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	s := *server
	s.TLS = true
	return a.Auth.Start(&s)
}

func main() {
	cfg := config.Get()

	log.Namespace = "dp-frontend-dataset-controller"

	router := mux.NewRouter()

	f := filter.New(cfg.FilterAPIURL, "", "")
	zc := client.NewZebedeeClient(cfg.ZebedeeURL)
	dc := dataset.NewAPIClient(cfg.DatasetAPIURL, "", "")
	rend := renderer.New(cfg.RendererURL)

	router.StrictSlash(true).Path("/healthcheck").HandlerFunc(healthcheck.Handler)

	router.StrictSlash(true).Path("/datasets/{datasetID}").Methods("GET").HandlerFunc(handlers.EditionsList(dc, rend))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions").Methods("GET").HandlerFunc(handlers.EditionsList(dc, rend))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{editionID}").Methods("GET").HandlerFunc(handlers.FilterableLanding(dc, rend, zc))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{edition}/versions").Methods("GET").HandlerFunc(handlers.VersionsList(dc, rend))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}").Methods("GET").HandlerFunc(handlers.FilterableLanding(dc, rend, zc))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{edition}/versions/{version}/metadata.txt").Methods("GET").HandlerFunc(handlers.MetadataText(dc))

	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter").Methods("POST").HandlerFunc(handlers.CreateFilterID(f, dc))

	if len(cfg.MailHost) > 0 {

		auth := smtp.PlainAuth(
			"",
			cfg.MailUser,
			cfg.MailPassword,
			cfg.MailHost,
		)

		if cfg.MailHost == "localhost" {
			auth = unencryptedAuth{auth}
		}

		mailAddr := fmt.Sprintf("%s:%s", cfg.MailHost, cfg.MailPort)

		log.Info("adding feedback routes", nil)
		router.StrictSlash(true).Path("/feedback").Methods("POST").HandlerFunc(handlers.AddFeedback(auth, mailAddr, cfg.FeedbackTo, cfg.FeedbackFrom, false))
		router.StrictSlash(true).Path("/feedback/positive").Methods("POST").HandlerFunc(handlers.AddFeedback(auth, mailAddr, cfg.FeedbackTo, cfg.FeedbackFrom, true))
		router.StrictSlash(true).Path("/feedback").Methods("GET").HandlerFunc(handlers.GetFeedback)
		router.StrictSlash(true).Path("/feedback/thanks").Methods("GET").HandlerFunc(handlers.FeedbackThanks)
	}

	router.StrictSlash(true).HandleFunc("/{uri:.*}", handlers.LegacyLanding(zc, rend))

	log.Info("Starting server", log.Data{"config": cfg})

	s := server.New(cfg.BindAddr, router)
	s.HandleOSSignals = false

	s.Middleware["CollectionID"] = collectionID.Handler
	s.MiddlewareOrder = append(s.MiddlewareOrder, "CollectionID")

	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Error(err, nil)
			os.Exit(2)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	log.InfoCtx(ctx, "shutting service down gracefully", nil)
	defer cancel()
	if err := s.Server.Shutdown(ctx); err != nil {
		log.ErrorCtx(ctx, err, nil)
	}
}
