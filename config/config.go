package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

var cfg *Config

// Config represents service configuration for dp-frontend-dataset-controller
type Config struct {
	APIRouterURL                       string        `envconfig:"API_ROUTER_URL"`
	BindAddr                           string        `envconfig:"BIND_ADDR"`
	CacheNavigationUpdateInterval      time.Duration `envconfig:"CACHE_NAVIGATION_UPDATE_INTERVAL"`
	Debug                              bool          `envconfig:"DEBUG"`
	DownloadServiceURL                 string        `envconfig:"DOWNLOAD_SERVICE_URL"`
	EnableMultivariate                 bool          `envconfig:"ENABLE_MULTIVARIATE"`
	EnableNewNavBar                    bool          `envconfig:"ENABLE_NEW_NAV_BAR"`
	EnableProfiler                     bool          `envconfig:"ENABLE_PROFILER"`
	FeedbackAPIURL                     string        `envconfig:"FEEDBACK_API_URL"`
	FilterFlexDatasetServiceURL        string        `envconfig:"FILTER_FLEX_DATASET_SERVICE_URL"`
	FrontendFilterDatasetControllerURL string        `envconfig:"FRONTEND_FILTER_DATASET_CONTROLLER_URL"`
	GracefulShutdownTimeout            time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckCriticalTimeout         time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	HealthCheckInterval                time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	IsPublishing                       bool          `envconfig:"IS_PUBLISHING"`
	OTBatchTimeout                     time.Duration `encconfig:"OTEL_BATCH_TIMEOUT"`
	OTServiceName                      string        `envconfig:"OTEL_SERVICE_NAME"`
	OTExporterOTLPEndpoint             string        `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OtelEnabled                        bool          `envconfig:"OTEL_ENABLED"`
	PatternLibraryAssetsPath           string        `envconfig:"PATTERN_LIBRARY_ASSETS_PATH"`
	PprofToken                         string        `envconfig:"PPROF_TOKEN" json:"-"`
	SiteDomain                         string        `envconfig:"SITE_DOMAIN"`
	SupportedLanguages                 []string      `envconfig:"SUPPORTED_LANGUAGES"`
}

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	config, err := get()
	if err != nil {
		return nil, err
	}

	if config.Debug {
		config.PatternLibraryAssetsPath = "http://localhost:9002/dist/assets"
	} else {
		config.PatternLibraryAssetsPath = "//cdn.ons.gov.uk/dp-design-system/f3e1909"
	}

	return config, nil
}

func get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		APIRouterURL:                       "http://localhost:23200/v1",
		BindAddr:                           "localhost:20200",
		CacheNavigationUpdateInterval:      10 * time.Second,
		Debug:                              false,
		DownloadServiceURL:                 "http://localhost:23600",
		EnableMultivariate:                 false,
		EnableNewNavBar:                    false,
		EnableProfiler:                     false,
		FeedbackAPIURL:                     "http://localhost:23200/v1/feedback",
		FilterFlexDatasetServiceURL:        "http://localhost:27100",
		FrontendFilterDatasetControllerURL: "http://localhost:20001",
		GracefulShutdownTimeout:            5 * time.Second,
		HealthCheckCriticalTimeout:         90 * time.Second,
		HealthCheckInterval:                30 * time.Second,
		IsPublishing:                       false,
		OTBatchTimeout:                     5 * time.Second,
		OTExporterOTLPEndpoint:             "localhost:4317",
		OTServiceName:                      "dp-frontend-dataset-controller",
		OtelEnabled:                        false,
		SiteDomain:                         "localhost",
		SupportedLanguages:                 []string{"en", "cy"},
	}

	return cfg, envconfig.Process("", cfg)
}
