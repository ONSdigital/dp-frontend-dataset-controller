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
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/go-ns/handlers/accessToken"
	"github.com/ONSdigital/go-ns/handlers/collectionID"
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

// App version informaton retrieved on runtime
var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string
)

func main() {
	log.Namespace = "dp-frontend-dataset-controller"

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	ctx := context.Background()

	cfg, err := config.Get()
	if err != nil {
		log.Event(ctx, "unable to retrieve service configuration", log.Error(err))
		os.Exit(1)
	}

	log.Event(ctx, "got service configuration", log.Data{"config": cfg})

	versionInfo, err := health.NewVersionInfo(
		BuildTime,
		GitCommit,
		Version,
	)
	if err != nil {
		log.Event(ctx, "failed to create service version information", log.Error(err))
		os.Exit(1)
	}

	router := mux.NewRouter()

	f := filter.New(cfg.FilterAPIURL)
	zc := zebedee.New(cfg.ZebedeeURL)
	dc := dataset.NewAPIClient(cfg.DatasetAPIURL)
	rend := renderer.New(cfg.RendererURL)

	healthcheck := health.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)

	if err = registerCheckers(ctx, &healthcheck, f, zc, dc, rend); err != nil {
		log.Event(ctx, "failed to add checkers", log.Error(err))
		os.Exit(1)
	}

	router.StrictSlash(true).Path("/health").HandlerFunc(healthcheck.Handler)

	router.StrictSlash(true).Path("/datasets/{datasetID}").Methods("GET").HandlerFunc(handlers.EditionsList(dc, zc, rend, *cfg))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions").Methods("GET").HandlerFunc(handlers.EditionsList(dc, zc, rend, *cfg))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{editionID}").Methods("GET").HandlerFunc(handlers.FilterableLanding(dc, rend, zc, *cfg))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{edition}/versions").Methods("GET").HandlerFunc(handlers.VersionsList(dc, rend, *cfg))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}").Methods("GET").HandlerFunc(handlers.FilterableLanding(dc, rend, zc, *cfg))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{edition}/versions/{version}/metadata.txt").Methods("GET").HandlerFunc(handlers.MetadataText(dc, *cfg))

	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter").Methods("POST").HandlerFunc(handlers.CreateFilterID(f, dc, *cfg))

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
		router.StrictSlash(true).Path("/feedback").Methods("POST").HandlerFunc(handlers.AddFeedback(auth, mailAddr, cfg.FeedbackTo, cfg.FeedbackFrom, cfg.RendererURL, false))
		router.StrictSlash(true).Path("/feedback/positive").Methods("POST").HandlerFunc(handlers.AddFeedback(auth, mailAddr, cfg.FeedbackTo, cfg.FeedbackFrom, cfg.RendererURL, false))
		router.StrictSlash(true).Path("/feedback").Methods("GET").HandlerFunc(handlers.GetFeedback(cfg.RendererURL))
		router.StrictSlash(true).Path("/feedback/thanks").Methods("GET").HandlerFunc(handlers.FeedbackThanks(cfg.RendererURL))
	}

	router.StrictSlash(true).HandleFunc("/{uri:.*}", handlers.LegacyLanding(zc, dc, rend, *cfg))

	log.Event(ctx, "Starting server", log.Data{"config": cfg})

	// Start healthcheck tickers
	healthcheck.Start(ctx)

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

	<-signals

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	log.Event(ctx, "shutting service down gracefully")
	defer cancel()

	// Stop healthcheck tickers
	healthcheck.Stop()
	if err := s.Server.Shutdown(ctx); err != nil {
		log.Event(ctx, "failed to shutdown http server", log.Error(err))
	}
}

func registerCheckers(ctx context.Context, h *health.HealthCheck, f *filter.Client, z *zebedee.Client, d *dataset.Client, r *renderer.Renderer) (err error) {
	if err = h.AddCheck("filter API", f.Checker); err != nil {
		log.Event(ctx, "failed to add filter API checker", log.Error(err))
	}

	if err = h.AddCheck("zebedee", z.Checker); err != nil {
		log.Event(ctx, "failed to add zebedee checker", log.Error(err))
	}

	if err = h.AddCheck("dataset API", d.Checker); err != nil {
		log.Event(ctx, "failed to add dataset API checker", log.Error(err))
	}

	if err = h.AddCheck("frontend renderer", r.Checker); err != nil {
		log.Event(ctx, "failed to add frontend renderer checker", log.Error(err))
	}

	return
}
