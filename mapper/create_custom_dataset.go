package mapper

import (
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/custom"
	"github.com/ONSdigital/dp-renderer/v2/helper"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
)

// CreateCustomDatasetPage builds a base datasetLandingPageCensus.Page with shared functionality between Dataset Landing Pages and Filter Output pages
func CreateCustomDatasetPage(req *http.Request, basePage coreModel.Page, populationTypes []population.PopulationType, lang, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) custom.Page {
	p := custom.Page{
		Page: basePage,
	}

	// PAGE BASICS
	p.Metadata.Title = helper.Localise("CreateCustomDatasetTitle", lang, 1)
	p.Language = lang
	p.URI = req.URL.Path
	p.Metadata.Description = p.Metadata.Title

	// BANNERS
	p.BetaBannerEnabled = true
	p.ServiceMessage = serviceMessage
	p.EmergencyBanner = mapEmergencyBanner(emergencyBannerContent)

	// CENSUS BRANDING
	p.ShowCensusBranding = true

	// FEEDBACK API
	p.FeatureFlags.FeedbackAPIURL = cfg.FeedbackAPIURL

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

	// ERROR HANDLING
	errorVal := req.URL.Query().Get("error")
	if errorVal == "true" {
		p.Error = coreModel.Error{
			Title: helper.Localise("CreateCustomDatasetErrorText", lang, 1),
			ErrorItems: []coreModel.ErrorItem{
				{
					Description: coreModel.Localisation{
						LocaleKey: "CreateCustomDatasetErrorText",
						Plural:    1,
					},
					URL: "#population-type",
				},
			},
			Language: lang,
		}
	}

	return p
}

// mapPopulationTypes maps population.PopulationType to createCensusDatasetPage.PopulationType
func mapPopulationTypes(populationTypes []population.PopulationType) []custom.PopulationType {
	mapped := []custom.PopulationType{}
	for _, pop := range populationTypes {
		model := custom.PopulationType{
			Name:        pop.Name,
			Label:       pop.Label,
			Description: pop.Description,
		}
		mapped = append(mapped, model)
	}
	return mapped
}
