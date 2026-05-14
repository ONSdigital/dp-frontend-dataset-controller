package clients

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
)

// FilterClient is an interface with the methods required for a filter client
type FilterClient interface {
	CreateBlueprint(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version string, names []string) (filterID, eTag string, err error)
	CreateCustomFilter(ctx context.Context, userAuthToken, serviceAuthToken, populationType string) (filterID string, err error)
	CreateFlexibleBlueprint(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version string, dimensions []filter.ModelDimension, populationType string) (filterID, eTag string, err error)
	GetOutput(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterOutputID string) (m filter.Model, err error)
	GetDimensionOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, q *filter.QueryParams) (opts filter.DimensionOptions, eTag string, err error)
	CreateFlexibleBlueprintCustom(ctx context.Context, uAuthToken, svcAuthToken, dlServiceToken string, req filter.CreateFlexBlueprintCustomRequest) (filterID, eTag string, err error)
}
