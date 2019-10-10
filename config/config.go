package config

import (
	"encoding/json"

	"github.com/ian-kent/gofigure"
)

var cfg Config

type Config struct {
	BindAddr           string `env:"BIND_ADDR"`
	ZebedeeURL         string `env:"ZEBEDEE_URL"`
	RendererURL        string `env:"RENDERER_URL"`
	FilterAPIURL       string `env:"FILTER_API_URL"`
	DatasetAPIURL      string `env:"DATASET_API_URL"`
	MailHost           string `env:"MAIL_HOST"`
	MailUser           string `env:"MAIL_USER"`
	MailPassword       string `env:"MAIL_PASSWORD" json:"-"`
	MailPort           string `env:"MAIL_PORT"`
	FeedbackTo         string `env:"FEEDBACK_TO"`
	FeedbackFrom       string `env:"FEEDBACK_FROM"`
	DownloadServiceURL string `env:"DOWNLOAD_SERVICE_URL"`
	ServiceToken       string `env:"SERVICE_TOKEN"`
}

func init() {
	cfg = Config{
		BindAddr:           ":20200",
		ZebedeeURL:         "http://localhost:8082",
		RendererURL:        "http://localhost:20010",
		FilterAPIURL:       "http://localhost:22100",
		DatasetAPIURL:      "http://localhost:22000",
		DownloadServiceURL: "http://localhost:23600",
		MailHost:           "localhost",
		MailPort:           "1025",
		MailUser:           "",
		MailPassword:       "",
		FeedbackTo:         "",
		FeedbackFrom:       "",
		ServiceToken:       "",
	}
	err := gofigure.Gofigure(&cfg)
	if err != nil {
		panic(err)
	}
}

//Get ...
func Get() Config {
	return cfg
}

// String is implemented to prevent sensitive fields being logged.
// The config is returned as JSON with sensitive fields omitted.
func (config Config) String() string {
	json, _ := json.Marshal(config)
	return string(json)
}
