package handlers

import (
	"context"

	zebedee "github.com/ONSdigital/dp-api-clients-go/zebedee"
)

// To mock interfaces in this file
// mockgen -source=handlers/zebedee.go -destination=handlers/mock_zebedee.go -imports=handlers=github.com/ONSdigital/dp-frontend-dataset-controller/handlers -package=handlers

// ZebedeeClient is an interface for zebedee client
type ZebedeeClient interface {
	GetBreadcrumb(ctx context.Context, userAccessToken, path string) ([]zebedee.Breadcrumb, error)
	Get(ctx context.Context, userAccessToken, path string) ([]byte, error)
	GetDatasetLandingPage(ctx context.Context, userAccessToken, path string) (zebedee.DatasetLandingPage, error)
	GetDataset(ctx context.Context, userAccessToken, path string) (zebedee.Dataset, error)
}
