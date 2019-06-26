package handlers

import (
	"context"

	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/zebedee/data"
)

// To mock interfaces in this file
// mockgen -source=handlers/zebedee.go -destination=handlers/mock_zebedee.go -imports=handlers=github.com/ONSdigital/dp-frontend-dataset-controller/handlers -package=handlers

// ZebedeeClient is an interface for zebedee client
type ZebedeeClient interface {
	GetBreadcrumb(context.Context, string) ([]data.Breadcrumb, error)
	Get(context.Context, string) ([]byte, error)
	GetDatasetLandingPage(context.Context, string) (data.DatasetLandingPage, error)
	GetDataset(context.Context, string) (data.Dataset, error)
	healthcheck.Client
}
