package handlers

import (
	"context"
	io "io"

	"github.com/ONSdigital/dp-api-clients-go/v2/files"
	"github.com/ONSdigital/dp-api-clients-go/v2/population"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"

	coreModel "github.com/ONSdigital/dp-renderer/model"
)

// To mock interfaces in this file
//go:generate mockgen -source=clients.go -destination=mock_clients.go -package=handlers github.com/ONSdigital/dp-frontend-dataset-controller/handlers FilterClient,DatasetClient,RenderClient

// FilterClient is an interface with the methods required for a filter client
type FilterClient interface {
	CreateBlueprint(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version string, names []string) (filterID, eTag string, err error)
	CreateFlexibleBlueprint(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version string, dimensions []filter.ModelDimension, population_type string) (filterID, eTag string, err error)
	GetOutput(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterOutputID string) (m filter.Model, err error)
	GetDimensionOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, q *filter.QueryParams) (opts filter.DimensionOptions, eTag string, err error)
}

// DatasetClient is an interface with methods required for a dataset client
type DatasetClient interface {
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
	GetAreas(ctx context.Context, input population.GetAreasInput) (population.GetAreasResponse, error)
	GetArea(ctx context.Context, input population.GetAreaInput) (population.GetAreaResponse, error)
	GetDimensionsDescription(ctx context.Context, input population.GetDimensionsDescriptionInput) (population.GetDimensionsResponse, error)
	GetParentAreaCount(ctx context.Context, input population.GetParentAreaCountInput) (int, error)
	GetCategorisations(ctx context.Context, input population.GetCategorisationsInput) (population.GetCategorisationsResponse, error)
}

// ClientError is an interface that can be used to retrieve the status code if a client has errored
type ClientError interface {
	Error() string
	Code() int
}
