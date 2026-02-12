package handlers

import (
	"fmt"
	"net/http"

	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	dpHandlers "github.com/ONSdigital/dp-net/v3/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// StaticEditionsList handles requests for the editions list page of static datasets
func StaticEditionsList(datasetAPIClient DatasetAPISdkClient, renderClient RenderClient, zebedeeClient ZebedeeClient, topicAPIClient TopicAPIClient, cfg config.Config, apiRouterVersion string) http.HandlerFunc {
	return dpHandlers.ControllerHandler(func(w http.ResponseWriter, r *http.Request, lang, collectionID, userAccessToken string) {
		staticEditionsList(r, w, datasetAPIClient, renderClient, zebedeeClient, topicAPIClient, cfg, apiRouterVersion, userAccessToken, lang, collectionID)
	})
}

func staticEditionsList(r *http.Request, w http.ResponseWriter, datasetAPIClient DatasetAPISdkClient, renderClient RenderClient, zebedeeClient ZebedeeClient, topicAPIClient TopicAPIClient, cfg config.Config, apiRouterVersion, userAccessToken, lang, collectionID string) {
	ctx := r.Context()

	vars := mux.Vars(r)
	topic := vars["topic"]
	datasetID := vars["datasetID"]
	editionID := vars["editionID"]

	logData := log.Data{
		"topicID":   topic,
		"datasetID": datasetID,
		"editionID": editionID,
	}

	datasetAPIClientHeaders := datasetAPISDK.Headers{AccessToken: userAccessToken}

	dataset, err := datasetAPIClient.GetDataset(ctx, datasetAPIClientHeaders, datasetID)
	if err != nil {
		log.Error(ctx, "failed to fetch dataset", err, logData)
		setStatusCode(ctx, w, err)
		return
	}

	if dataset.Type != DatasetTypeStatic {
		log.Error(ctx, "dataset is not of type static", errDatasetNotStatic, logData)
		setStatusCode(ctx, w, errDatasetNotStatic)
		return
	}

	// Topics is a mandatory field but nil check is added to prevent potential panics
	if len(dataset.Topics) == 0 || dataset.Topics[0] != topic {
		log.Error(ctx, "dataset topic does not match URL topic", errDatasetTopicMismatch, logData)
		setStatusCode(ctx, w, errDatasetTopicMismatch)
		return
	}

	// If editionID is provided then request came from /{topic}/datasets/{datasetID}/editions/{editionID}.
	// Redirect to the latest version of the dataset.
	if editionID != "" {
		log.Info(ctx, "editionID provided in URL, redirecting to latest version", logData)
		redirectPath := fmt.Sprintf("/%s/datasets/%s/editions/%s/versions/%s", topic, datasetID, editionID, dataset.Links.LatestVersion.ID)
		http.Redirect(w, r, redirectPath, http.StatusFound)
		return
	}

	editions, err := datasetAPIClient.GetEditions(ctx, datasetAPIClientHeaders, datasetID, &datasetAPISDK.QueryParams{Limit: 1000})
	if err != nil {
		log.Error(ctx, "failed to fetch editions list", err, logData)
		setStatusCode(ctx, w, err)
		return
	}

	// Redirect to latest version if number of editions <= 1
	if len(editions.Items) <= 1 {
		log.Info(ctx, "only one edition exists, redirecting to latest version", logData)
		redirectPath := fmt.Sprintf("/%s/datasets/%s/editions/%s/versions/%s", topic, datasetID, editions.Items[0].Edition, editions.Items[0].Links.LatestVersion.ID)
		http.Redirect(w, r, redirectPath, http.StatusFound)
		return
	}

	topicList := fetchTopics(ctx, cfg, topicAPIClient, dataset.Topics, userAccessToken)

	// Fetch homepage content
	homepageContent, err := zebedeeClient.GetHomepageContent(ctx, userAccessToken, collectionID, lang, homepagePath)
	if err != nil {
		logData["homepageContentError"] = err
		log.Warn(ctx, "failed to get homepage content", logData)
	}

	// Build and render the page
	basePage := renderClient.NewBasePageModel()
	mapper.UpdateBasePage(&basePage, dataset, homepageContent, false, lang, r)
	pageModel := mapper.CreateEditionsListForStaticDatasetType(ctx, basePage, r, dataset, editions, datasetID, apiRouterVersion, topicList)
	renderClient.BuildPage(w, pageModel, templateNameStaticEditionsList)
}
