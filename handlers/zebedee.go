package handlers

import (
	"context"

	zebedee "github.com/ONSdigital/dp-api-clients-go/zebedee"
)

// To mock interfaces in this file
//go:generate mockgen -source=zebedee.go -destination=mock_zebedee.go -package=handlers github.com/ONSdigital/dp-frontend-dataset-controller/handlers ZebedeeClient

// ZebedeeClient is an interface for zebedee client
type ZebedeeClient interface {
	GetBreadcrumb(ctx context.Context, userAccessToken, collectionID, lang, path string) ([]zebedee.Breadcrumb, error)
	Get(ctx context.Context, userAccessToken, path string) ([]byte, error)
	GetDatasetLandingPage(ctx context.Context, userAccessToken, collectionID, lang, path string) (zebedee.DatasetLandingPage, error)
	GetDataset(ctx context.Context, userAccessToken, collectionID, lang, path string) (zebedee.Dataset, error)
}
