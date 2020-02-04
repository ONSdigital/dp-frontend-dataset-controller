package main

import (
	"context"
	"fmt"
	"net/smtp"
	"os"
	"os/signal"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-api-clients-go/renderer"
	zebedee "github.com/ONSdigital/dp-api-clients-go/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/handlers"
	"github.com/ONSdigital/go-ns/handlers/accessToken"
	"github.com/ONSdigital/go-ns/handlers/collectionID"
	"github.com/ONSdigital/go-ns/handlers/healthcheck"
	"github.com/ONSdigital/go-ns/handlers/localeCode"
	"github.com/ONSdigital/go-ns/server"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
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
	ctx := context.Background()

	log.Namespace = "dp-frontend-dataset-controller"

	router := mux.NewRouter()

	f := filter.New(cfg.FilterAPIURL)
	zc := zebedee.New(cfg.ZebedeeURL)
	dc := dataset.NewAPIClient(cfg.DatasetAPIURL)
	rend := renderer.New(cfg.RendererURL)

	router.StrictSlash(true).Path("/healthcheck").HandlerFunc(healthcheck.Handler)

	router.StrictSlash(true).Path("/datasets/{datasetID}").Methods("GET").HandlerFunc(handlers.EditionsList(dc, zc, rend, cfg))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions").Methods("GET").HandlerFunc(handlers.EditionsList(dc, zc, rend, cfg))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{editionID}").Methods("GET").HandlerFunc(handlers.FilterableLanding(dc, rend, zc, cfg))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{edition}/versions").Methods("GET").HandlerFunc(handlers.VersionsList(dc, rend, cfg))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}").Methods("GET").HandlerFunc(handlers.FilterableLanding(dc, rend, zc, cfg))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{edition}/versions/{version}/metadata.txt").Methods("GET").HandlerFunc(handlers.MetadataText(dc, cfg))

	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter").Methods("POST").HandlerFunc(handlers.CreateFilterID(f, dc, cfg))

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

		log.Event(ctx, "adding feedback routes")
		router.StrictSlash(true).Path("/feedback").Methods("POST").HandlerFunc(handlers.AddFeedback(auth, mailAddr, cfg.FeedbackTo, cfg.FeedbackFrom, false))
		router.StrictSlash(true).Path("/feedback/positive").Methods("POST").HandlerFunc(handlers.AddFeedback(auth, mailAddr, cfg.FeedbackTo, cfg.FeedbackFrom, true))
		router.StrictSlash(true).Path("/feedback").Methods("GET").HandlerFunc(handlers.GetFeedback)
		router.StrictSlash(true).Path("/feedback/thanks").Methods("GET").HandlerFunc(handlers.FeedbackThanks)
	}

	router.StrictSlash(true).HandleFunc("/{uri:.*}", handlers.LegacyLanding(zc, dc, rend, cfg))

	log.Event(ctx, "Starting server", log.Data{"config": cfg})

	s := server.New(cfg.BindAddr, router)
	s.HandleOSSignals = false

	s.Middleware["CollectionID"] = collectionID.CheckCookie
	s.MiddlewareOrder = append(s.MiddlewareOrder, "CollectionID")

	s.Middleware["AccessToken"] = accessToken.CheckCookieValueAndForwardWithRequestContext
	s.MiddlewareOrder = append(s.MiddlewareOrder, "AccessToken")

	s.Middleware["LocaleCode"] = localeCode.CheckHeaderValueAndForwardWithRequestContext
	s.MiddlewareOrder = append(s.MiddlewareOrder, "LocaleCode")

	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Event(ctx, "failed to start http listen and serve", log.Error(err))
			os.Exit(2)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	log.Event(ctx, "shutting service down gracefully")
	defer cancel()
	if err := s.Server.Shutdown(ctx); err != nil {
		log.Event(ctx, "failed to shutdown http server", log.Error(err))
	}
}
