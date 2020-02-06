package config

import (
	"encoding/json"
	"time"

	"github.com/kelseyhightower/envconfig"
)

var cfg *Config

// Config represents service configuration for dp-frontend-dataset-controller
type Config struct {
	BindAddr                   string        `env:"BIND_ADDR"`
	ZebedeeURL                 string        `env:"ZEBEDEE_URL"`
	RendererURL                string        `env:"RENDERER_URL"`
	FilterAPIURL               string        `env:"FILTER_API_URL"`
	DatasetAPIURL              string        `env:"DATASET_API_URL"`
	MailHost                   string        `env:"MAIL_HOST"`
	MailUser                   string        `env:"MAIL_USER"`
	MailPassword               string        `env:"MAIL_PASSWORD" json:"-"`
	MailPort                   string        `env:"MAIL_PORT"`
	FeedbackTo                 string        `env:"FEEDBACK_TO"`
	FeedbackFrom               string        `env:"FEEDBACK_FROM"`
	DownloadServiceURL         string        `env:"DOWNLOAD_SERVICE_URL"`
	ServiceToken               string        `env:"SERVICE_TOKEN"`
	EnableLoop11               bool          `env:"ENABLE_LOOP11"`
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
		HealthCheckInterval:        10 * time.Second,
		HealthCheckCriticalTimeout: time.Minute,
	}

	return cfg, envconfig.Process("", cfg)
}

// String is implemented to prevent sensitive fields being logged.
// The config is returned as JSON with sensitive fields omitted.
func (config Config) String() string {
	json, _ := json.Marshal(config)
	return string(json)
}
