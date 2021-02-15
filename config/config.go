package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

var cfg *Config

// Config represents service configuration for dp-frontend-dataset-controller
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	RendererURL                string        `envconfig:"RENDERER_URL"`
	APIRouterURL               string        `envconfig:"API_ROUTER_URL"`
	MailHost                   string        `envconfig:"MAIL_HOST"`
	MailUser                   string        `envconfig:"MAIL_USER"`
	MailPassword               string        `envconfig:"MAIL_PASSWORD" json:"-"`
	MailPort                   string        `envconfig:"MAIL_PORT"`
	FeedbackTo                 string        `envconfig:"FEEDBACK_TO"`
	FeedbackFrom               string        `envconfig:"FEEDBACK_FROM"`
	DownloadServiceURL         string        `envconfig:"DOWNLOAD_SERVICE_URL"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	EnableProfiler             bool          `envconfig:"ENABLE_PROFILER"`
	PprofToken                 string        `envconfig:"PPROF_TOKEN" json:"-"`
}

// Get returns the default config with any modifications through environment
// variables
func Get() (cfg *Config, err error) {
	cfg = &Config{
		BindAddr:                   ":20200",
		RendererURL:                "http://localhost:20010",
		APIRouterURL:               "http://localhost:23200/v1",
		DownloadServiceURL:         "http://localhost:23600",
		MailHost:                   "localhost",
		MailPort:                   "1025",
		MailUser:                   "",
		MailPassword:               "",
		FeedbackTo:                 "",
		FeedbackFrom:               "",
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		EnableProfiler:             false,
	}

	return cfg, envconfig.Process("", cfg)
}
