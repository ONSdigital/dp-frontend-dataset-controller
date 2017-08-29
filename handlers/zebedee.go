package handlers

import (
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/zebedee/data"
)

// ZebedeeClient is an interface for zebedee client
type ZebedeeClient interface {
	SetAccessToken(string)
	GetBreadcrumb(string) ([]data.Breadcrumb, error)
	Get(string) ([]byte, error)
	GetDatasetLandingPage(string) (data.DatasetLandingPage, error)
	GetDataset(string) (data.Dataset, error)
	healthcheck.Client
}
