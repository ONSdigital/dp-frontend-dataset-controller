package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-net/v2/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// FilterOutput will load a filtered landing page
func FilterOutputDownloads(zc ZebedeeClient, fc FilterClient, pc PopulationClient, dc DatasetClient, rend RenderClient, cfg config.Config, apiRouterVersion string) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		filterOutputDownloads(w, req, zc, dc, fc, pc, rend, cfg, collectionID, lang, apiRouterVersion, userAccessToken)
	})
}

func filterOutputDownloads(w http.ResponseWriter, req *http.Request, zc ZebedeeClient, dc DatasetClient, fc FilterClient, pc PopulationClient, rend RenderClient, cfg config.Config, collectionID, lang, apiRouterVersion, userAccessToken string) {
	vars := mux.Vars(req)
	ctx := req.Context()
	filterOutputID := vars["filterOutputID"]

	filterOutput, fErr := fc.GetOutput(ctx, userAccessToken, "", "", collectionID, filterOutputID)
	if fErr != nil {
		log.Error(ctx, "failed to get filter-output", fErr, log.Data{"filter-output": filterOutputID})
		setStatusCode(ctx, w, fErr)
		return
	}

	downloads := filterOutput.Downloads
	fileTypes := make([]string, 0, len(downloads))
	for k := range downloads {
		fileTypes = append(fileTypes, k)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(fileTypes); err != nil {
		log.Error(ctx, "failed to encode filter-output downloads", err, log.Data{"filter-output": filterOutputID})
		setStatusCode(ctx, w, err)
		return
	}
}
