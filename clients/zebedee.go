package clients

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
)

// ZebedeeClient is an interface for a zebedee client
type ZebedeeClient interface {
	GetBreadcrumb(ctx context.Context, userAccessToken, collectionID, lang, path string) ([]zebedee.Breadcrumb, error)
	Get(ctx context.Context, userAccessToken, path string) ([]byte, error)
	GetDatasetLandingPage(ctx context.Context, userAccessToken, collectionID, lang, path string) (zebedee.DatasetLandingPage, error)
	GetDataset(ctx context.Context, userAccessToken, collectionID, lang, path string) (zebedee.Dataset, error)
	GetHomepageContent(ctx context.Context, userAccessToken, collectionID, lang, path string) (m zebedee.HomepageContent, err error)
}
