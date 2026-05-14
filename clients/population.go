package clients

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/population"
)

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
