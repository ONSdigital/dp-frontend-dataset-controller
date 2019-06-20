package handlers

import (
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/zebedee/data"
)

// To mock interfaces in this file
// mockgen -source=handlers/zebedee.go -destination=handlers/mock_zebedee.go -imports=handlers=github.com/ONSdigital/dp-frontend-dataset-controller/handlers -package=handlers

// ZebedeeClient is an interface for zebedee client
type ZebedeeClient interface {
	SetAccessToken(string)
	GetBreadcrumb(string) ([]data.Breadcrumb, error)
	Get(string) ([]byte, error)
	GetDatasetLandingPage(string) (data.DatasetLandingPage, error)
	GetDataset(string) (data.Dataset, error)
	healthcheck.Client
}
