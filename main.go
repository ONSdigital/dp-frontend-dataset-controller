package main

import (
	"context"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"os/signal"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	apihealthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/assets"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/handlers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	render "github.com/ONSdigital/dp-renderer"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/pkg/errors"

	dpnethandlers "github.com/ONSdigital/dp-net/v2/handlers"
	dpnethttp "github.com/ONSdigital/dp-net/v2/http"

	_ "net/http/pprof"
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
		log.Error(ctx, "application unexpectedly failed", err)
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
		log.Error(ctx, "unable to retrieve service configuration", err)
		return err
	}

	log.Info(ctx, "got service configuration", log.Data{"config": cfg})

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
		log.Error(ctx, "failed to create service version information", err)
		return err
	}

	router := mux.NewRouter()

	apiRouterCli := apihealthcheck.NewClient("api-router", cfg.APIRouterURL)

	f := filter.NewWithHealthClient(apiRouterCli)
	zc := zebedee.NewWithHealthClient(apiRouterCli)
	dc := dataset.NewWithHealthClient(apiRouterCli)

	healthcheck := health.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)

	if err = registerCheckers(ctx, &healthcheck, apiRouterCli); err != nil {
		os.Exit(1)
	}

	// Initialise render client, routes and initialise localisations bundles
	rend := render.NewWithDefaultClient(assets.Asset, assets.AssetNames, cfg.PatternLibraryAssetsPath, cfg.SiteDomain)

	// Enable profiling endpoint for authorised users
	if cfg.EnableProfiler {
		middlewareChain := alice.New(profileMiddleware(cfg.PprofToken)).Then(http.DefaultServeMux)
		router.PathPrefix("/debug").Handler(middlewareChain)
	}

	router.StrictSlash(true).Path("/health").HandlerFunc(healthcheck.Handler)

	router.StrictSlash(true).Path("/datasets/{datasetID}").Methods("GET").HandlerFunc(handlers.EditionsList(dc, zc, rend, *cfg, apiRouterVersion))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions").Methods("GET").HandlerFunc(handlers.EditionsList(dc, zc, rend, *cfg, apiRouterVersion))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{editionID}").Methods("GET").HandlerFunc(handlers.FilterableLanding(dc, rend, zc, *cfg, apiRouterVersion))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{edition}/versions").Methods("GET").HandlerFunc(handlers.VersionsList(dc, zc, rend, *cfg))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}").Methods("GET").HandlerFunc(handlers.FilterableLanding(dc, rend, zc, *cfg, apiRouterVersion))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{edition}/versions/{version}/metadata.txt").Methods("GET").HandlerFunc(handlers.MetadataText(dc, *cfg))

	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter").Methods("POST").HandlerFunc(handlers.CreateFilterID(f, dc))
	if cfg.EnableCensusPages {
		router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-flex").Methods("POST").HandlerFunc(handlers.CreateFilterFlexID(f, dc, *cfg))
	}

	router.StrictSlash(true).HandleFunc("/{uri:.*}", handlers.LegacyLanding(zc, dc, rend, *cfg))

	log.Info(ctx, "Starting server", log.Data{"config": cfg})

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
		log.Error(ctx, "service error received", err)
	case osSignal := <-signals:
		log.Info(ctx, "quitting after os signal received", log.Data{"signal": osSignal})
	}

	log.Info(ctx, fmt.Sprintf("shutdown with timeout: %s", cfg.GracefulShutdownTimeout))

	ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)

	var gracefulShutdown bool

	go func() {
		defer cancel()
		var hasShutdownErrs bool

		log.Info(ctx, "stop health checkers")
		healthcheck.Stop()

		if err := s.Shutdown(ctx); err != nil {
			log.Error(ctx, "failed to gracefully shutdown http server", err)
			hasShutdownErrs = true
		}

		if !hasShutdownErrs {
			gracefulShutdown = true
		}
	}()

	// wait for timeout or success (via cancel)
	<-ctx.Done()
	if ctx.Err() == context.DeadlineExceeded {
		log.Warn(ctx, "context deadline exceeded", log.FormatErrors([]error{ctx.Err()}))
		return err
	}

	if !gracefulShutdown {
		err = errors.New("failed to shutdown gracefully")
		log.Error(ctx, "failed to shutdown gracefully ", err)
		return err
	}

	log.Info(ctx, "graceful shutdown complete", log.Data{"context": ctx.Err()})

	return nil
}

func registerCheckers(ctx context.Context, h *health.HealthCheck, apiRouterCli *apihealthcheck.Client) (err error) {
	hasErrors := false

	if err = h.AddCheck("API router", apiRouterCli.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "failed to add API router health checker", err)
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
				log.Error(ctx, "invalid auth token", errors.New("invalid auth token"))
				w.WriteHeader(404)
				return
			}

			log.Info(ctx, "accessing profiling endpoint")
			h.ServeHTTP(w, req)
		})
	}
}
