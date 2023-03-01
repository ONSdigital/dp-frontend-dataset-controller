package handlers

import (
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/dp-net/v2/handlers"
	"github.com/ONSdigital/log.go/v2/log"
)

// FilterOutput will load a filtered landing page
func CreateCustomDataset(pc PopulationClient, zc ZebedeeClient, rend RenderClient, cfg config.Config, apiRouterVersion string) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		createCustomDataset(w, req, pc, zc, rend, cfg, collectionID, lang, apiRouterVersion, userAccessToken)
	})
}

func createCustomDataset(w http.ResponseWriter, req *http.Request, pc PopulationClient, zc ZebedeeClient, rend RenderClient, cfg config.Config, collectionID, lang, apiRouterVersion, userAccessToken string) {
	basePage := rend.NewBasePageModel()

	ctx := req.Context()

	homepageContent, err := zc.GetHomepageContent(ctx, userAccessToken, collectionID, lang, homepagePath)
	if err != nil {
		log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{err}), log.Data{"homepage_content": err})
	}

	populationTypes, err := pc.GetPopulationTypes(ctx, population.GetPopulationTypesInput{
		DefaultDatasets: true,
		AuthTokens: population.AuthTokens{
			UserAuthToken: userAccessToken,
		},
	})
	if err != nil {
		log.Warn(ctx, "unable to get population types", log.FormatErrors([]error{err}), log.Data{"homepage_content": err})
	}

	page := mapper.CreateCustomDatasetPage(ctx, req, basePage, populationTypes.Items, lang, homepageContent.ServiceMessage, zebedee.EmergencyBanner{})
	rend.BuildPage(w, page, "create-custom-dataset")
}
