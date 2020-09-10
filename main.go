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
	"github.com/justinas/alice"
	"github.com/pkg/errors"

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
		log.Event(ctx, "application unexpectedly failed", log.ERROR, log.Error(err))
		os.Exit(1)
	}

	os.Exit(0)
}

func run(ctx context.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	cfg, err := config.Get()
	if err != nil {
		log.Event(ctx, "unable to retrieve service configuration", log.ERROR, log.Error(err))
		return err
	}

	log.Event(ctx, "got service configuration", log.INFO, log.Data{"config": cfg})

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

	f := filter.New(cfg.FilterAPIURL)
	zc := zebedee.New(cfg.ZebedeeURL)
	dc := dataset.NewAPIClient(cfg.DatasetAPIURL)
	rend := renderer.New(cfg.RendererURL)

	healthcheck := health.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)

	if err = registerCheckers(ctx, &healthcheck, f, zc, dc, rend); err != nil {
		os.Exit(1)
	}

	// Enable profiling endpoint for authorised users
	if cfg.EnableProfiler {
		middlewareChain := alice.New(profileMiddleware(cfg.PprofToken)).Then(http.DefaultServeMux)
		router.PathPrefix("/debug").Handler(middlewareChain)
	}

	router.StrictSlash(true).Path("/health").HandlerFunc(healthcheck.Handler)

	router.StrictSlash(true).Path("/datasets/{datasetID}").Methods("GET").HandlerFunc(handlers.EditionsList(dc, zc, rend, *cfg))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions").Methods("GET").HandlerFunc(handlers.EditionsList(dc, zc, rend, *cfg))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{editionID}").Methods("GET").HandlerFunc(handlers.FilterableLanding(dc, rend, zc, *cfg))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{edition}/versions").Methods("GET").HandlerFunc(handlers.VersionsList(dc, rend, *cfg))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}").Methods("GET").HandlerFunc(handlers.FilterableLanding(dc, rend, zc, *cfg))
	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{edition}/versions/{version}/metadata.txt").Methods("GET").HandlerFunc(handlers.MetadataText(dc, *cfg))

	router.StrictSlash(true).Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter").Methods("POST").HandlerFunc(handlers.CreateFilterID(f, dc, *cfg))

	router.StrictSlash(true).HandleFunc("/{uri:.*}", handlers.LegacyLanding(zc, dc, rend, *cfg))

	log.Event(ctx, "Starting server", log.INFO, log.Data{"config": cfg})

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

func registerCheckers(ctx context.Context, h *health.HealthCheck, f *filter.Client, z *zebedee.Client, d *dataset.Client, r *renderer.Renderer) (err error) {

	hasErrors := false

	if err = h.AddCheck("filter API", f.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add filter API checker", log.ERROR, log.Error(err))
	}

	if err = h.AddCheck("zebedee", z.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add zebedee checker", log.ERROR, log.Error(err))
	}

	if err = h.AddCheck("dataset API", d.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add dataset API checker", log.ERROR, log.Error(err))
	}

	if err = h.AddCheck("frontend renderer", r.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add frontend renderer checker", log.ERROR, log.Error(err))
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
