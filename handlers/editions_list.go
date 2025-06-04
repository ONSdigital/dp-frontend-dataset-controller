package handlers

import (
	"fmt"
	"net/http"

	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/dp-net/v3/handlers"
	dpTopicApiModels "github.com/ONSdigital/dp-topic-api/models"
	dpTopicApiSdk "github.com/ONSdigital/dp-topic-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// EditionsList will load a list of editions for a filterable dataset
func EditionsList(dc DatasetAPISdkClient, zc ZebedeeClient, tc TopicAPIClient, rend RenderClient, cfg config.Config, apiRouterVersion string) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		editionsList(w, req, dc, zc, tc, rend, cfg, collectionID, lang, apiRouterVersion, userAccessToken)
	})
}

func editionsList(w http.ResponseWriter, req *http.Request, dc DatasetAPISdkClient, zc ZebedeeClient, tc TopicAPIClient, rend RenderClient, cfg config.Config, collectionID, lang, apiRouterVersion, userAccessToken string) {
	vars := mux.Vars(req)
	datasetID := vars["datasetID"]
	ctx := req.Context()

	serviceAuthToken := ""

	headers := dpDatasetApiSdk.Headers{
		UserAccessToken: userAccessToken,
		CollectionID:    collectionID,
	}

	topicHeaders := dpTopicApiSdk.Headers{
		ServiceAuthToken: serviceAuthToken,
		UserAuthToken:    userAccessToken,
	}

	datasetDetails, err := dc.GetDataset(ctx, headers, collectionID, datasetID)
	if err != nil {
		setStatusCode(ctx, w, err)
		return
	}

	queryParams := dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}
	datasetEditions, err := dc.GetEditions(ctx, headers, datasetID, &queryParams)
	if err != nil {
		if err, ok := err.(ClientError); ok {
			if err.Code() != http.StatusNotFound {
				setStatusCode(ctx, w, err)
				return
			}
		}
	}

	// redirect to latest version if number of editions is one or less
	numberOfEditions := len(datasetEditions.Items)
	if numberOfEditions == 1 {
		latestVersionPath := helpers.DatasetVersionURL(datasetID, datasetEditions.Items[0].Edition, datasetEditions.Items[0].Links.LatestVersion.ID)
		log.Info(ctx, "only one edition, therefore redirecting to latest version", log.Data{"latestVersionPath": latestVersionPath})
		http.Redirect(w, req, latestVersionPath, http.StatusFound)
	}

	// Fetch homepage content
	homepageContent, err := zc.GetHomepageContent(ctx, userAccessToken, collectionID, lang, homepagePath)
	if err != nil {
		log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{err}), log.Data{"homepage_content": err})
	}

	// Build page context
	basePage := rend.NewBasePageModel()
	// Update basePage common parameters
	mapper.UpdateBasePage(&basePage, datasetDetails, homepageContent, false, lang, req)

	if datasetDetails.Type == DatasetTypeStatic {
		// BREADCRUMB
		topicsIDList := datasetDetails.Topics
		topicObjectList := []dpTopicApiModels.Topic{}
		someTopicAPIFetchesFailed := false

		for _, topicID := range topicsIDList {
			topicObject, err := GetPublicOrPrivateTopics(tc, cfg, ctx, topicHeaders, topicID)
			if err != nil {
				someTopicAPIFetchesFailed = true
				log.Warn(
					ctx,
					fmt.Sprintf("unable to get topic data for topic ID: %s", topicID),
					log.FormatErrors([]error{err}),
				)
				continue
			}
			topicObjectList = append(topicObjectList, *topicObject)
		}
		// We can't construct breadcrumbs with only part of data
		if someTopicAPIFetchesFailed {
			topicObjectList = []dpTopicApiModels.Topic{}
		}

		m := mapper.CreateEditionsListForStaticDatasetType(ctx, basePage, req, datasetDetails, datasetEditions, datasetID, apiRouterVersion, topicObjectList)
		rend.BuildPage(w, m, "edition-list-static")
	} else {
		bc, err := zc.GetBreadcrumb(ctx, userAccessToken, userAccessToken, collectionID, datasetDetails.Links.Taxonomy.HRef)
		if err != nil {
			log.Warn(ctx, "unable to get breadcrumb for dataset uri", log.FormatErrors([]error{err}), log.Data{"taxonomy_url": datasetDetails.Links.Taxonomy.HRef})
		}

		m := mapper.CreateEditionsList(ctx, basePage, req, datasetDetails, datasetEditions, datasetID, bc, apiRouterVersion)
		rend.BuildPage(w, m, "edition-list")
	}
}
