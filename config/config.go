package config

import "github.com/ian-kent/gofigure"

var cfg Config

type Config struct {
	BindAddr            string `env:"BIND_ADDR"`
	ZebedeeURL          string `env:"ZEBEDEE_URL"`
	RendererURL         string `env:"RENDERER_URL"`
	FilterAPIURL        string `env:"FILTER_API_URL"`
	DatasetAPIURL       string `env:"DATASET_API_URL"`
	DatasetAPIAuthToken string `env:"DATASET_API_AUTH_TOKEN"`
	MailHost            string `env:"MAIL_HOST"`
	MailUser            string `env:"MAIL_USER"`
	MailPassword        string `env:"MAIL_PASSWORD"`
	MailPort            string `env:"MAIL_PORT"`
	FeedbackTo          string `env:"FEEDBACK_TO"`
	FeedbackFrom        string `env:"FEEDBACK_FROM"`
}

func init() {
	cfg = Config{
		BindAddr:            ":20200",
		ZebedeeURL:          "http://localhost:8082",
		RendererURL:         "http://localhost:20010",
		FilterAPIURL:        "http://localhost:22100",
		DatasetAPIURL:       "http://localhost:22000",
		DatasetAPIAuthToken: "FD0108EA-825D-411C-9B1D-41EF7727F465",
		MailHost:            "",
		MailPort:            "",
		MailUser:            "",
		MailPassword:        "",
		FeedbackTo:          "",
		FeedbackFrom:        "",
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
