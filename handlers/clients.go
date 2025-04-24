package handlers

import (
	"context"
	io "io"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/files"
	"github.com/ONSdigital/dp-api-clients-go/v2/population"

	"github.com/ONSdigital/dp-api-clients-go/v2/filter"

	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
)

// To mock interfaces in this file
//go:generate mockgen -source=clients.go -destination=mock_clients.go -package=handlers github.com/ONSdigital/dp-frontend-dataset-controller/handlers FilterClient,ApiClientsGoDatasetClient,RenderClient

// FilterClient is an interface with the methods required for a filter client
type FilterClient interface {
	CreateBlueprint(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version string, names []string) (filterID, eTag string, err error)
	CreateCustomFilter(ctx context.Context, userAuthToken, serviceAuthToken, populationType string) (filterID string, err error)
	CreateFlexibleBlueprint(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version string, dimensions []filter.ModelDimension, populationType string) (filterID, eTag string, err error)
	GetOutput(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterOutputID string) (m filter.Model, err error)
	GetDimensionOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, q *filter.QueryParams) (opts filter.DimensionOptions, eTag string, err error)
	CreateFlexibleBlueprintCustom(ctx context.Context, uAuthToken, svcAuthToken, dlServiceToken string, req filter.CreateFlexBlueprintCustomRequest) (filterID, eTag string, err error)
}

// Interface with methods required for a dp-api-clients-go dataset client
type ApiClientsGoDatasetClient interface {
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
type DatasetApiSdkClient interface {
	GetDataset(ctx context.Context, headers dpDatasetApiSdk.Headers, collectionID, datasetID string) (m dpDatasetApiModels.Dataset, err error)
	GetDatasetByPath(ctx context.Context, headers dpDatasetApiSdk.Headers, path string) (m dpDatasetApiModels.Dataset, err error)
	GetEditions(ctx context.Context, headers dpDatasetApiSdk.Headers, datasetID string, q *dpDatasetApiSdk.QueryParams) (m dpDatasetApiSdk.EditionsList, err error)
	GetEdition(ctx context.Context, headers dpDatasetApiSdk.Headers, datasetID, edition string) (dpDatasetApiModels.Edition, error)
	GetVersions(ctx context.Context, headers dpDatasetApiSdk.Headers, datasetID, editionID string, q *dpDatasetApiSdk.QueryParams) (m dpDatasetApiSdk.VersionsList, err error)
	GetVersion(ctx context.Context, headers dpDatasetApiSdk.Headers, datasetID, editionID, versionID string) (m dpDatasetApiModels.Version, err error)
	GetVersionMetadata(ctx context.Context, headers dpDatasetApiSdk.Headers, datasetID, editionID, versionID string) (m dpDatasetApiModels.Metadata, err error)
	GetVersionDimensions(ctx context.Context, headers dpDatasetApiSdk.Headers, datasetID, editionID, versionID string) (m dpDatasetApiSdk.VersionDimensionsList, err error)
	GetVersionDimensionOptions(ctx context.Context, headers dpDatasetApiSdk.Headers, datasetID, editionID, versionID, dimensionID string, q *dpDatasetApiSdk.QueryParams) (m dpDatasetApiSdk.VersionDimensionOptionsList, err error)
}

// RenderClient is an interface with methods for require for rendering a template
type RenderClient interface {
	BuildPage(w io.Writer, pageModel interface{}, templateName string)
	NewBasePageModel() coreModel.Page
}

// FilesAPIClient is an interface with methods required for getting metadata from Files API
type FilesAPIClient interface {
	GetFile(ctx context.Context, path string, authToken string) (files.FileMetaData, error)
}

// PopulationClient is an interface with methods required for a population client
type PopulationClient interface {
	GetArea(ctx context.Context, input population.GetAreaInput) (population.GetAreaResponse, error)
	GetAreas(ctx context.Context, input population.GetAreasInput) (population.GetAreasResponse, error)
	GetBlockedAreaCount(ctx context.Context, input population.GetBlockedAreaCountInput) (*cantabular.GetBlockedAreaCountResult, error)
	GetDimensionCategories(ctx context.Context, input population.GetDimensionCategoryInput) (population.GetDimensionCategoriesResponse, error)
	GetDimensionsDescription(ctx context.Context, input population.GetDimensionsDescriptionInput) (population.GetDimensionsResponse, error)
	GetCategorisations(ctx context.Context, input population.GetCategorisationsInput) (population.GetCategorisationsResponse, error)
	GetPopulationType(ctx context.Context, input population.GetPopulationTypeInput) (population.GetPopulationTypeResponse, error)
	GetPopulationTypes(ctx context.Context, input population.GetPopulationTypesInput) (population.GetPopulationTypesResponse, error)
	GetPopulationTypeMetadata(ctx context.Context, input population.GetPopulationTypeMetadataInput) (population.GetPopulationTypeMetadataResponse, error)
}

// ClientError is an interface that can be used to retrieve the status code if a client has errored
type ClientError interface {
	Error() string
	Code() int
}
