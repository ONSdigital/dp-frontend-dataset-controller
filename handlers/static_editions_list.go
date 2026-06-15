package handlers

import (
	"net/http"

	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/clients"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	dpHandlers "github.com/ONSdigital/dp-net/v3/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// StaticEditionsList handles requests for the editions list page of static datasets
func StaticEditionsList(datasetAPIClient clients.DatasetAPISdkClient, renderClient clients.RenderClient, zebedeeClient clients.ZebedeeClient, topicAPIClient clients.TopicAPIClient, cfg config.Config, apiRouterVersion string) http.HandlerFunc {
	return dpHandlers.ControllerHandler(func(w http.ResponseWriter, r *http.Request, lang, collectionID, userAccessToken string) {
		staticEditionsList(r, w, datasetAPIClient, renderClient, zebedeeClient, topicAPIClient, cfg, apiRouterVersion, userAccessToken, lang, collectionID)
	})
}

func staticEditionsList(r *http.Request, w http.ResponseWriter, datasetAPIClient clients.DatasetAPISdkClient, renderClient clients.RenderClient, zebedeeClient clients.ZebedeeClient, topicAPIClient clients.TopicAPIClient, cfg config.Config, apiRouterVersion, userAccessToken, lang, collectionID string) {
	ctx := r.Context()

	vars := mux.Vars(r)
	topicSlug := vars["topic"]
	datasetID := vars["datasetID"]
	editionID := vars["editionID"]

	logData := log.Data{
		"topicSlug": topicSlug,
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
		log.Error(ctx, "dataset is not of type static", errDatasetTypeNotSupported, logData)
		setStatusCode(ctx, w, errDatasetTypeNotSupported)
		return
	}

	topicList, err := clients.FetchTopics(ctx, topicAPIClient, dataset.Topics, cfg.IsPublishing, userAccessToken)
	if err != nil {
		log.Error(ctx, "failed to fetch topics", err, logData)
		setStatusCode(ctx, w, err)
		return
	}

	if len(topicList) == 0 {
		log.Error(ctx, "no topics found for dataset", errDatasetHasNoTopics, logData)
		setStatusCode(ctx, w, errDatasetHasNoTopics)
		return
	}

	expectedTopicSlug := topicList[0].Slug

	// If the URL topic slug doesn't match the dataset's primary topic slug, redirect to the expected one
	if expectedTopicSlug != topicSlug {
		logData["providedTopicSlug"] = topicSlug
		logData["expectedTopicSlug"] = expectedTopicSlug
		log.Info(ctx, "incorrect topic slug provided, redirecting to correct topic", logData)

		// Reconstruct the request path with the expected topic slug
		redirectPath := helpers.ReplaceFirstPathSegment(r.URL.Path, expectedTopicSlug)

		//nolint:gosec // false positive as this is a relative URL which can only redirect to the same host
		http.Redirect(w, r, redirectPath, http.StatusFound)
		return
	}

	// If editionID is provided then request came from /{topic}/datasets/{datasetID}/editions/{editionID}.
	// Redirect to the latest version of the dataset.
	if editionID != "" {
		log.Info(ctx, "editionID provided in URL, redirecting to latest version", logData)
		redirectPath, err := helpers.PrefixPathWithTopic(topicSlug, dataset.Links.LatestVersion.HRef)
		if err != nil {
			log.Error(ctx, "failed to create redirect path for latest version", err, logData)
			setStatusCode(ctx, w, err)
			return
		}
		//nolint:gosec // false positive as this is a relative URL which can only redirect to the same host
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
		redirectPath, err := helpers.PrefixPathWithTopic(topicSlug, dataset.Links.LatestVersion.HRef)
		if err != nil {
			log.Error(ctx, "failed to create redirect path for latest version", err, logData)
			setStatusCode(ctx, w, err)
			return
		}
		//nolint:gosec // false positive as this is a relative URL which can only redirect to the same host
		http.Redirect(w, r, redirectPath, http.StatusFound)
		return
	}

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
