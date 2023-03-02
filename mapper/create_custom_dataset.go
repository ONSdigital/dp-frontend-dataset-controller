package mapper

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/createCustomDatasetPage"
	coreModel "github.com/ONSdigital/dp-renderer/model"
)

// CreateCustomDatasetPage builds a base datasetLandingPageCensus.Page with shared functionality between Dataset Landing Pages and Filter Output pages
func CreateCustomDatasetPage(ctx context.Context, req *http.Request, basePage coreModel.Page, populationTypes []population.PopulationType, lang string, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) createCustomDatasetPage.Page {
	p := createCustomDatasetPage.Page{
		Page: basePage,
	}

	// BANNERS
	p.BetaBannerEnabled = true
	p.ServiceMessage = serviceMessage
	p.EmergencyBanner = mapEmergencyBanner(emergencyBannerContent)

	// CENSUS BRANDING
	p.ShowCensusBranding = true

	// BREADCRUMBS
	p.Breadcrumb = []coreModel.TaxonomyNode{
		{
			Title: "Home",
			URI:   "/",
		},
		{
			Title: "Census",
			URI:   "/census",
		},
	}

	// PAGE CONTENT
	p.CreateCustomDatasetPage.PopulationTypes = mapPopulationTypes(populationTypes)

	return p
}

// mapPapulationTypes maps population.PopulationType to createCensusDatasetPage.PopulationType
func mapPopulationTypes(populationTypes []population.PopulationType) []createCustomDatasetPage.PopulationType {
	mapped := []createCustomDatasetPage.PopulationType{}
	for _, pop := range populationTypes {
		custom := createCustomDatasetPage.PopulationType{
			Name:        pop.Name,
			Label:       pop.Label,
			Description: pop.Description,
		}
		mapped = append(mapped, custom)
	}
	return mapped
}
