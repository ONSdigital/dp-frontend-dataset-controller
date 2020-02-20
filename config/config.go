package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

var cfg *Config

// Config represents service configuration for dp-frontend-dataset-controller
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	ZebedeeURL                 string        `envconfig:"ZEBEDEE_URL"`
	RendererURL                string        `envconfig:"RENDERER_URL"`
	FilterAPIURL               string        `envconfig:"FILTER_API_URL"`
	DatasetAPIURL              string        `envconfig:"DATASET_API_URL"`
	MailHost                   string        `envconfig:"MAIL_HOST"`
	MailUser                   string        `envconfig:"MAIL_USER"`
	MailPassword               string        `envconfig:"MAIL_PASSWORD" json:"-"`
	MailPort                   string        `envconfig:"MAIL_PORT"`
	FeedbackTo                 string        `envconfig:"FEEDBACK_TO"`
	FeedbackFrom               string        `envconfig:"FEEDBACK_FROM"`
	DownloadServiceURL         string        `envconfig:"DOWNLOAD_SERVICE_URL"`
	ServiceToken               string        `envconfig:"SERVICE_TOKEN" json:"-"`
	EnableLoop11               bool          `envconfig:"ENABLE_LOOP11"`
	EnableCookiesControl       bool          `envconfig:"ENABLE_COOKIES_CONTROL"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
}

// Get returns the default config with any modifications through environment
// variables
func Get() (cfg *Config, err error) {
	cfg = &Config{
		BindAddr:                   ":20200",
		ZebedeeURL:                 "http://localhost:8082",
		RendererURL:                "http://localhost:20010",
		FilterAPIURL:               "http://localhost:22100",
		DatasetAPIURL:              "http://localhost:22000",
		DownloadServiceURL:         "http://localhost:23600",
		MailHost:                   "localhost",
		MailPort:                   "1025",
		MailUser:                   "",
		MailPassword:               "",
		FeedbackTo:                 "",
		FeedbackFrom:               "",
		ServiceToken:               "",
		EnableLoop11:               false,
		EnableCookiesControl:       false,
		HealthCheckInterval:        10 * time.Second,
		HealthCheckCriticalTimeout: time.Minute,
	}

	return cfg, envconfig.Process("", cfg)
}
