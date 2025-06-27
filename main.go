package main

import (
	"context"
	goErrors "errors"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"os/signal"

	"github.com/ONSdigital/dp-api-clients-go/v2/files"
	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	"github.com/ONSdigital/dp-frontend-dataset-controller/cache"
	cachePublic "github.com/ONSdigital/dp-frontend-dataset-controller/cache/public"
	topic "github.com/ONSdigital/dp-topic-api/sdk"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	apihealthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/assets"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/handlers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	render "github.com/ONSdigital/dp-renderer/v2"
	"github.com/ONSdigital/dp-renderer/v2/middleware/renderror"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/pkg/errors"

	dpnethandlers "github.com/ONSdigital/dp-net/v3/handlers"
	dpnethttp "github.com/ONSdigital/dp-net/v3/http"
	dpotelgo "github.com/ONSdigital/dp-otel-go"
)

// nolint:unused // ignoring unused type
type unencryptedAuth struct {
	smtp.Auth
}

// nolint:unused // ignoring unused function
func (a unencryptedAuth) Start(server *smtp.ServerInfo) (string, []byte, error) { // nolint:gocritic // unnamedResult: consider giving a name to these results
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

// nolint:gocyclo // ignoring complexity warning
func run(ctx context.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill) // nolint:staticcheck // SA1016: os.Kill cannot be trapped (did you mean syscall.SIGTERM?)

	// Get config
	cfg, err := config.Get()
	if err != nil {
		log.Error(ctx, "unable to retrieve service configuration", err)
		return err
	}

	log.Info(ctx, "got service configuration", log.Data{"config": cfg})

	if cfg.OtelEnabled {
		otelConfig := dpotelgo.Config{
			OtelServiceName:          cfg.OTServiceName,
			OtelExporterOtlpEndpoint: cfg.OTExporterOTLPEndpoint,
			OtelBatchTimeout:         cfg.OTBatchTimeout,
		}

		otelShutdown, oErr := dpotelgo.SetupOTelSDK(ctx, otelConfig)
		if oErr != nil {
			log.Fatal(ctx, "error setting up OpenTelemetry - hint: ensure OTEL_EXPORTER_OTLP_ENDPOINT is set", oErr)
			return oErr
		}
		// Handle shutdown properly so nothing leaks.
		defer func() {
			err = goErrors.Join(err, otelShutdown(context.Background()))
		}()
	}

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

	router := mux.NewRouter().StrictSlash(true)

	if cfg.OtelEnabled {
		router.Use(otelmux.Middleware(cfg.OTServiceName))
	}

	apiRouterCli := apihealthcheck.NewClient("api-router", cfg.APIRouterURL)

	populationClient, err := population.NewWithHealthClient(apiRouterCli)
	if err != nil {
		return fmt.Errorf("failed to create population API client: %w", err)
	}

	f := filter.NewWithHealthClient(apiRouterCli)
	zc := zebedee.NewWithHealthClient(apiRouterCli)
	apiClientsGoDatasetClient := dataset.NewWithHealthClient(apiRouterCli)
	datasetAPISdkClient := dpDatasetApiSdk.NewWithHealthClient(apiRouterCli)
	fc := files.NewWithHealthClient(apiRouterCli)
	tc := topic.NewWithHealthClient(apiRouterCli)
	pc := populationClient
	fc.Version = "v1"

	healthcheck := health.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)

	if err = registerCheckers(ctx, &healthcheck, apiRouterCli); err != nil {
		os.Exit(1) // nolint:gocritic // ignoring exitAfterDefer: os.Exit will exit, and `defer func(){...}(...)` will not run
	}

	// Initialise render client, routes and initialise localisations bundles
	rend := render.NewWithDefaultClient(assets.Asset, assets.AssetNames, cfg.PatternLibraryAssetsPath, cfg.SiteDomain)

	// Enable profiling endpoint for authorised users
	if cfg.EnableProfiler {
		middlewareChain := alice.New(profileMiddleware(cfg.PprofToken)).Then(http.DefaultServeMux)
		router.PathPrefix("/debug").Handler(middlewareChain)
	}

	// Initialise caching
	cacheList := &cache.List{}
	cacheList.Navigation, err = cache.NewNavigationCache(ctx, &cfg.CacheNavigationUpdateInterval)
	if err != nil {
		log.Error(ctx, "failed to create navigation cache", err, log.Data{"update_interval": cfg.CacheNavigationUpdateInterval})
		return err
	}
	for _, lang := range cfg.SupportedLanguages {
		navigationlangKey := cacheList.Navigation.GetCachingKeyForNavigationLanguage(lang)
		cacheList.Navigation.AddUpdateFunc(navigationlangKey, cachePublic.UpdateNavigationData(ctx, cfg, lang, tc))
	}

	router.Path("/health").HandlerFunc(healthcheck.Handler)

	filterDatasetControllerURL, _ := helpers.ParseURL(ctx, cfg.FrontendFilterDatasetControllerURL, "FilterDatasetControllerURL")
	filterFlexDatasetServiceURL, _ := helpers.ParseURL(ctx, cfg.FilterFlexDatasetServiceURL, "FilterFlexDatasetServiceURL")
	filterHandler := helpers.CreateReverseProxy("filters", filterDatasetControllerURL) // CMD
	filterFlexHandler := helpers.CreateReverseProxy("flex", filterFlexDatasetServiceURL) // Cantabular


	if cfg.EnableMultivariate {
		router.Path("/datasets/create").Methods("GET").HandlerFunc(handlers.CreateCustomDataset(pc, zc, rend, *cfg, apiRouterVersion))
		router.Path("/datasets/create").Methods("POST").HandlerFunc(handlers.PostCreateCustomDataset(f))
		router.Path("/datasets/create/filter-outputs/{filterOutputID}").Methods("GET").HandlerFunc(handlers.FilterOutput(zc, f, pc, datasetAPISdkClient, rend, *cfg, apiRouterVersion))
		router.Path("/datasets/create/filter-outputs/{filterOutputID}").Methods("POST").HandlerFunc(handlers.CreateFilterFlexIDFromOutput(f))
	}

	router.Path("/datasets/{datasetID}").Methods("GET").HandlerFunc(handlers.EditionsList(datasetAPISdkClient, zc, tc, rend, *cfg, apiRouterVersion))
	router.Path("/datasets/{datasetID}/editions").Methods("GET").HandlerFunc(handlers.EditionsList(datasetAPISdkClient, zc, tc, rend, *cfg, apiRouterVersion))
	router.Path("/datasets/{datasetID}/editions/{editionID}").Methods("GET").HandlerFunc(handlers.FilterableLanding(datasetAPISdkClient, pc, rend, zc, tc, *cfg, apiRouterVersion))
	router.Path("/datasets/{datasetID}/editions/{edition}/versions").Methods("GET").HandlerFunc(handlers.VersionsList(datasetAPISdkClient, zc, rend, *cfg))
	router.Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}").Methods("GET").HandlerFunc(handlers.FilterableLanding(datasetAPISdkClient, pc, rend, zc, tc, *cfg, apiRouterVersion))
	router.Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}").Methods("POST").HandlerFunc(handlers.CreateFilterFlexID(f, apiClientsGoDatasetClient))
	router.Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter").Methods("POST").HandlerFunc(handlers.CreateFilterID(f, apiClientsGoDatasetClient))
	router.Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}").Methods("GET").HandlerFunc(handlers.FilterOutput(zc, f, pc, datasetAPISdkClient, rend, *cfg, apiRouterVersion))
	router.Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}").Methods("POST").HandlerFunc(handlers.CreateFilterFlexIDFromOutput(f))

	router.Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/metadata.txt").Methods("GET").HandlerFunc(handlers.MetadataText(datasetAPISdkClient, *cfg))

	router.PathPrefix("/dataset/").Methods("GET").Handler(http.StripPrefix("/dataset/", handlers.DatasetPage(zc, rend, fc, cacheList)))
	router.HandleFunc("/filters/{uri:.*}", handlers.FilterPageHandler(f, datasetAPISdkClient, filterHandler, filterFlexHandler))
	router.HandleFunc("/{uri:.*}", handlers.LegacyLanding(zc, apiClientsGoDatasetClient, fc, rend, cacheList, *cfg))

	log.Info(ctx, "Starting server", log.Data{"config": cfg})

	// Start healthcheck tickers
	healthcheck.Start(ctx)

	// Start caching
	go cacheList.Navigation.StartUpdates(ctx, make(chan error))

	collectionIDMiddleware := dpnethandlers.CheckCookie(dpnethandlers.CollectionID)
	accessTokenMiddleware := dpnethandlers.CheckCookie(dpnethandlers.UserAccess)
	localeMiddleware := dpnethandlers.CheckHeader(dpnethandlers.Locale)
	renderrorMiddleware := renderror.Handler(rend)

	var middlewareChain http.Handler
	if cfg.OtelEnabled {
		otelMiddleware := otelhttp.NewMiddleware(cfg.OTServiceName)
		middlewareChain = alice.New(collectionIDMiddleware, accessTokenMiddleware, localeMiddleware, renderrorMiddleware, otelMiddleware).Then(router)
	} else {
		middlewareChain = alice.New(collectionIDMiddleware, accessTokenMiddleware, localeMiddleware, renderrorMiddleware).Then(router)
	}

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

		// Stop caching
		cacheList.Navigation.Close()

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
