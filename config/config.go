package config

import "github.com/ian-kent/gofigure"

var cfg Config

type Config struct {
	BindAddr      string `env:"BIND_ADDR"`
	ZebedeeURL    string `env:"ZEBEDEE_URL"`
	RendererURL   string `env:"RENDERER_URL"`
	FilterAPIURL  string `env:"FILTER_API_URL"`
	DatasetAPIURL string `env:"DATASET_API_URL"`
	SlackToken    string `env:"SLACK_TOKEN"`
}

func init() {
	cfg = Config{
		BindAddr:      ":20200",
		ZebedeeURL:    "http://localhost:8082",
		RendererURL:   "http://localhost:20010",
		FilterAPIURL:  "http://localhost:22100",
		DatasetAPIURL: "http://localhost:22000",
		SlackToken:    "",
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
