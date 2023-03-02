package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

var cfg *Config

// Config represents service configuration for dp-frontend-dataset-controller
type Config struct {
	BindAddr                      string        `envconfig:"BIND_ADDR"`
	Debug                         bool          `envconfig:"DEBUG"`
	EnableMultivariate            bool          `envconfig:"ENABLE_MULTIVARIATE"`
	APIRouterURL                  string        `envconfig:"API_ROUTER_URL"`
	SiteDomain                    string        `envconfig:"SITE_DOMAIN"`
	PatternLibraryAssetsPath      string        `envconfig:"PATTERN_LIBRARY_ASSETS_PATH"`
	SupportedLanguages            []string      `envconfig:"SUPPORTED_LANGUAGES"`
	DownloadServiceURL            string        `envconfig:"DOWNLOAD_SERVICE_URL"`
	GracefulShutdownTimeout       time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval           time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout    time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	EnableProfiler                bool          `envconfig:"ENABLE_PROFILER"`
	PprofToken                    string        `envconfig:"PPROF_TOKEN" json:"-"`
	CacheNavigationUpdateInterval time.Duration `envconfig:"CACHE_NAVIGATION_UPDATE_INTERVAL"`
	EnableNewNavBar               bool          `envconfig:"ENABLE_NEW_NAV_BAR"`
}

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	cfg, err := get()
	if err != nil {
		return nil, err
	}

	if cfg.Debug {
		cfg.PatternLibraryAssetsPath = "http://localhost:9002/dist/assets"
	} else {
		cfg.PatternLibraryAssetsPath = "//cdn.ons.gov.uk/dp-design-system/db32164"
	}

	return cfg, nil
}

func get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                      "localhost:20200",
		Debug:                         false,
		EnableMultivariate:            false,
		APIRouterURL:                  "http://localhost:23200/v1",
		DownloadServiceURL:            "http://localhost:23600",
		SiteDomain:                    "localhost",
		SupportedLanguages:            []string{"en", "cy"},
		GracefulShutdownTimeout:       5 * time.Second,
		HealthCheckInterval:           30 * time.Second,
		HealthCheckCriticalTimeout:    90 * time.Second,
		EnableProfiler:                false,
		CacheNavigationUpdateInterval: 10 * time.Second,
		EnableNewNavBar:               false,
	}

	return cfg, envconfig.Process("", cfg)
}
