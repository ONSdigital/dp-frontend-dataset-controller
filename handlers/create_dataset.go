package handlers

import (
	"net/http"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-net/v2/handlers"
)

// FilterOutput will load a filtered landing page
func CreateDataset(pc PopulationClient, rend RenderClient, cfg config.Config, apiRouterVersion string) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		createDataset(w, req, pc, rend, cfg, collectionID, lang, apiRouterVersion, userAccessToken)
	})
}

func createDataset(w http.ResponseWriter, req *http.Request, pc PopulationClient, rend RenderClient, cfg config.Config, collectionID, lang, apiRouterVersion, userAccessToken string) {
	p := rend.NewBasePageModel()
	rend.BuildPage(w, p, "create-dataset")
}
