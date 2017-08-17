package config

import "github.com/ian-kent/gofigure"

var cfg Config

type Config struct {
	BindAddr     string `env:"BIND_ADDR"`
	ZebedeeURL   string `env:"ZEBEDEE_URL"`
	RendererURL  string `env:"RENDERER_URL"`
	FilterAPIURL string `env:"FILTER_API_URL"`
}

func init() {
	cfg = Config{
		BindAddr:     ":20200",
		ZebedeeURL:   "http://localhost:8082",
		RendererURL:  "http://localhost:20010",
		FilterAPIURL: "http://localhost:22100",
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
