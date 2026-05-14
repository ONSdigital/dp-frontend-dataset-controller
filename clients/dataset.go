package clients

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"

	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
)

// Interface with methods required for a dp-api-clients-go dataset client
type APIClientsGoDatasetClient interface {
	Get(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID string) (m dataset.DatasetDetails, err error)
	GetByPath(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, path string) (m dataset.DatasetDetails, err error)
	GetEditions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID string) (m []dataset.Edition, err error)
	GetEdition(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID, edition string) (dataset.Edition, error)
	GetVersions(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition string, q *dataset.QueryParams) (m dataset.VersionsList, err error)
	GetVersion(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, version string) (m dataset.Version, err error)
	GetVersionMetadata(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version string) (m dataset.Metadata, err error)
	GetVersionDimensions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version string) (m dataset.VersionDimensions, err error)
	GetOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension string, q *dataset.QueryParams) (m dataset.Options, err error)
}

// Interface with methods required for a dp-dataset-api/sdk dataset client
type DatasetAPISdkClient interface {
	GetDataset(ctx context.Context, headers datasetAPISDK.Headers, datasetID string) (m datasetAPIModels.Dataset, err error)
	GetDatasetByPath(ctx context.Context, headers datasetAPISDK.Headers, path string) (m datasetAPIModels.Dataset, err error)
	GetEditions(ctx context.Context, headers datasetAPISDK.Headers, datasetID string, q *datasetAPISDK.QueryParams) (m datasetAPISDK.EditionsList, err error)
	GetEdition(ctx context.Context, headers datasetAPISDK.Headers, datasetID, edition string) (datasetAPIModels.Edition, error)
	GetVersions(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID string, q *datasetAPISDK.QueryParams) (m datasetAPISDK.VersionsList, err error)
	GetVersion(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID string) (m datasetAPIModels.Version, err error)
	GetVersionV2(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID string) (m datasetAPIModels.Version, err error)
	GetVersionMetadata(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID string) (m datasetAPIModels.Metadata, err error)
	GetVersionDimensions(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID string) (m datasetAPISDK.VersionDimensionsList, err error)
	GetVersionDimensionOptions(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID, dimensionID string, q *datasetAPISDK.QueryParams) (m datasetAPISDK.VersionDimensionOptionsList, err error)
	PutVersionState(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID, state string) (err error)
}
