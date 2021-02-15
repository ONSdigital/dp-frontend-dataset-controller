package main

import (
	"context"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"os/signal"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-api-clients-go/renderer"
	"github.com/ONSdigital/dp-api-clients-go/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/handlers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/pkg/errors"

	dpnethandlers "github.com/ONSdigital/dp-net/handlers"
	dpnethttp "github.com/ONSdigital/dp-net/http"

	_ "net/http/pprof"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/health"
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
	ctx := context.Background()
	log.Namespace = "dp-frontend-dataset-controller"

	if err := run(ctx); err != nil {
		log.Event(ctx, "application unexpectedly failed", log.ERROR, log.Error(err))
		os.Exit(1)
	}

	os.Exit(0)
}

func run(ctx context.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	// Get config
	cfg, err := config.Get()
	if err != nil {
		log.Event(ctx, "unable to retrieve service configuration", log.ERROR, log.Error(err))
		return err
	}
	log.Event(ctx, "got service configuration", log.INFO, log.Data{"config": cfg})

	// Get API version from its URL
	apiRouterVersion, err := helpers.GetAPIRouterVersion(cfg.APIRouterURL)
	if err != nil {
		return err
	}

	// Healthcheck version Info
	versionInfo, err := health.NewVersionInfo(
		BuildTime,
		GitCommit,
		Version,
	)
	if err != nil {
		log.Event(ctx, "failed to create service version information", log.ERROR, log.Error(err))
		return err
	}

	router := mux.NewRouter()

	apiRouterCli := healthcheck.NewClient("api-router", cfg.APIRouterURL)

	f := filter.NewWithHealthClient(apiRouterCli)
	zc := zebedee.NewWithHealthClient(apiRouterCli)
	dc := dataset.NewWithHealthClient(apiRouterCli)
	rend := renderer.New(cfg.RendererURL)

	healthcheck := health.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)

	if err = registerCheckers(ctx, &healthcheck, rend, apiRouterCli); err != nil {
		os.Exit(1)
	}

	// Enable profiling endpoint for authorised users
	if cfg.EnableProfiler {
		middlewareChain := alice.New(profileMiddleware(cfg.PprofToken)).Then(http.DefaultServeMux)
		router.PathPrefix("/debug").Handler(middlewareChain)
	}

	router.StrictSlash(true).Path("/health").HandlerFunc(healthcheck.Handler)

	router.StrictSlash(true).Path("/datasets/{datasetID}").Methods("GET").HandlerFunc(handlers.EditionsList(dc, zc, rend, *cfg, apiRouterVersion))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions").Methods("GET").HandlerFunc(handlers.EditionsList(dc, zc, rend, *cfg, apiRouterVersion))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{editionID}").Methods("GET").HandlerFunc(handlers.FilterableLanding(dc, rend, zc, *cfg, apiRouterVersion))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{edition}/versions").Methods("GET").HandlerFunc(handlers.VersionsList(dc, rend, *cfg))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}").Methods("GET").HandlerFunc(handlers.FilterableLanding(dc, rend, zc, *cfg, apiRouterVersion))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{edition}/versions/{version}/metadata.txt").Methods("GET").HandlerFunc(handlers.MetadataText(dc, *cfg))

	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter").Methods("POST").HandlerFunc(handlers.CreateFilterID(f, dc, *cfg))

	// Nomis dataset landing page
	//router.StrictSlash(true).Path("/datasets/nomis/{datasetID}/editions/{editionID}").Methods("GET").HandlerFunc(handlers.NomisLanding(dc, rend, zc, *cfg))
	//router.StrictSlash(true).Path("/datasets/nomis/{datasetID}/editions/{editionID}/versions/{versionID}").Methods("GET").HandlerFunc(handlers.NomisLanding(dc, rend, zc, *cfg))

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

		log.Event(ctx, "adding feedback routes", log.INFO)
		router.StrictSlash(true).Path("/feedback").Methods("POST").HandlerFunc(handlers.AddFeedback(auth, mailAddr, cfg.FeedbackTo, cfg.FeedbackFrom, rend, false))
		router.StrictSlash(true).Path("/feedback/positive").Methods("POST").HandlerFunc(handlers.AddFeedback(auth, mailAddr, cfg.FeedbackTo, cfg.FeedbackFrom, rend, false))
		router.StrictSlash(true).Path("/feedback").Methods("GET").HandlerFunc(handlers.GetFeedback(rend))
		router.StrictSlash(true).Path("/feedback/thanks").Methods("GET").HandlerFunc(handlers.FeedbackThanks(rend))
	}
	//router.StrictSlash(true).Path("/datasets/nomis/editions/{editionID}/versions/{versionID}/filter").Methods("POST").HandlerFunc(handlers.CreateFilterID(f, dc, *cfg))
	router.StrictSlash(true).HandleFunc("/{uri:.*}", handlers.LegacyLanding(zc, dc, rend, *cfg))

	log.Event(ctx, "Starting server", log.INFO, log.Data{"config": cfg})

	// Start healthcheck tickers
	healthcheck.Start(ctx)

	collectionIDMiddleware := dpnethandlers.CheckCookie(dpnethandlers.CollectionID)
	accessTokenMiddleware := dpnethandlers.CheckCookie(dpnethandlers.UserAccess)
	localeMiddleware := dpnethandlers.CheckHeader(dpnethandlers.Locale)
	middlewareChain := alice.New(collectionIDMiddleware, accessTokenMiddleware, localeMiddleware).Then(router)

	s := dpnethttp.NewServer(cfg.BindAddr, middlewareChain)
	s.HandleOSSignals = false

	svcErrors := make(chan error, 1)
	go func() {
		if err := s.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()

	// Block until a signal is called to shutdown application
	select {
	case err := <-svcErrors:
		log.Event(ctx, "service error received", log.ERROR, log.Error(err))
	case signal := <-signals:
		log.Event(ctx, "quitting after os signal received", log.INFO, log.Data{"signal": signal})
	}

	log.Event(ctx, fmt.Sprintf("shutdown with timeout: %s", cfg.GracefulShutdownTimeout), log.INFO)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)

	var gracefulShutdown bool

	go func() {
		defer cancel()
		var hasShutdownErrs bool

		log.Event(ctx, "stop health checkers", log.INFO)
		healthcheck.Stop()

		if err := s.Shutdown(ctx); err != nil {
			log.Event(ctx, "failed to gracefully shutdown http server", log.ERROR, log.Error(err))
			hasShutdownErrs = true
		}

		if !hasShutdownErrs {
			gracefulShutdown = true
		}
	}()

	// wait for timeout or success (via cancel)
	<-ctx.Done()
	if ctx.Err() == context.DeadlineExceeded {
		log.Event(ctx, "context deadline exceeded", log.WARN, log.Error(ctx.Err()))
		return err
	}

	if !gracefulShutdown {
		err = errors.New("failed to shutdown gracefully")
		log.Event(ctx, "failed to shutdown gracefully ", log.ERROR, log.Error(err))
		return err
	}

	log.Event(ctx, "graceful shutdown complete", log.INFO, log.Data{"context": ctx.Err()})

	return nil
}

func registerCheckers(ctx context.Context, h *health.HealthCheck, r *renderer.Renderer, apiRouterCli *healthcheck.Client) (err error) {

	hasErrors := false

	if err = h.AddCheck("frontend renderer", r.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add frontend renderer checker", log.ERROR, log.Error(err))
	}

	if err = h.AddCheck("API router", apiRouterCli.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add API router health checker", log.ERROR, log.Error(err))
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}

	return nil
}

// profileMiddleware to validate auth token before accessing endpoint
func profileMiddleware(token string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()

			pprofToken := req.Header.Get("Authorization")
			if pprofToken == "Bearer " || pprofToken != "Bearer "+token {
				log.Event(ctx, "invalid auth token", log.ERROR, log.Error(errors.New("invalid auth token")))
				w.WriteHeader(404)
				return
			}

			log.Event(ctx, "accessing profiling endpoint", log.INFO)
			h.ServeHTTP(w, req)
		})
	}
}
