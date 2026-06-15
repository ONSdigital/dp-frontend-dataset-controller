package handlers

import (
	"net/http"

	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/clients"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/dp-frontend-dataset-controller/permissions"
	dpHandlers "github.com/ONSdigital/dp-net/v3/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// StaticLanding handles requests for the landing page of static datasets
func StaticLanding(datasetAPIClient clients.DatasetAPISdkClient, renderClient clients.RenderClient, zebedeeClient clients.ZebedeeClient, topicAPIClient clients.TopicAPIClient, cfg config.Config, authMiddleware authorisation.Middleware) http.HandlerFunc {
	return dpHandlers.ControllerHandler(func(w http.ResponseWriter, r *http.Request, lang, collectionID, userAccessToken string) {
		staticLanding(r, w, datasetAPIClient, renderClient, zebedeeClient, topicAPIClient, cfg, authMiddleware, userAccessToken, lang, collectionID)
	})
}

func staticLanding(r *http.Request, w http.ResponseWriter, datasetAPIClient clients.DatasetAPISdkClient, renderClient clients.RenderClient, zebedeeClient clients.ZebedeeClient, topicAPIClient clients.TopicAPIClient, cfg config.Config, authMiddleware authorisation.Middleware, userAccessToken, lang, collectionID string) {
	ctx := r.Context()

	vars := mux.Vars(r)
	topicSlug := vars["topic"]
	datasetID := vars["datasetID"]
	editionID := vars["editionID"]
	versionID := vars["versionID"]

	formQueryParam := r.URL.Query().Get("f")
	formatQueryParam := r.URL.Query().Get("format")

	logData := log.Data{
		"topicSlug": topicSlug,
		"datasetID": datasetID,
		"editionID": editionID,
		"versionID": versionID,
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

	// If versionID is missing then request came from /{topic}/datasets/{datasetID}/editions/{editionID} or /{topic}/datasets/{datasetID}/editions/{editionID}/versions.
	// Redirect to the latest version of the dataset.
	if versionID == "" {
		log.Info(ctx, "versionID not provided in URL, redirecting to latest version", logData)
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

	version, err := datasetAPIClient.GetVersionV2(ctx, datasetAPIClientHeaders, datasetID, editionID, versionID)
	if err != nil {
		log.Error(ctx, "failed to fetch version", err, logData)
		setStatusCode(ctx, w, err)
		return
	}

	// Check if this is a download request.
	// If it is then redirect to the download URL for the requested format.
	var isValidationError bool
	if formQueryParam == formQueryGetData {
		logData["requestedFormat"] = formatQueryParam
		logData["distributions"] = version.Distributions

		downloadURL := helpers.GetDistributionFileURL(version.Distributions, formatQueryParam)
		if downloadURL == "" {
			log.Warn(ctx, "requested format not available for download", logData)
			isValidationError = true
		} else {
			log.Info(ctx, "redirecting to download URL for requested format", logData)
			//nolint:gosec // false positive as downloadURL is from a trusted source (dataset API)
			http.Redirect(w, r, downloadURL, http.StatusFound)
			return
		}
	}

	// enable approval button if user has permissions and environment is publishing
	var enableApprovalButton bool
	if cfg.IsPublishing {
		enableApprovalButton, err = permissions.CheckIsAdmin(ctx, userAccessToken, authMiddleware)
		if err != nil {
			log.Error(ctx, "error checking user permissions for approval button", err, logData)
			setStatusCode(ctx, w, err)
			return
		}
	}

	fullVersionsList, err := datasetAPIClient.GetVersions(ctx, datasetAPIClientHeaders, datasetID, editionID, &datasetAPISDK.QueryParams{Limit: 1000})
	if err != nil {
		log.Error(ctx, "failed to fetch versions list", err, logData)
		setStatusCode(ctx, w, err)
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
	mapper.UpdateBasePage(&basePage, dataset, homepageContent, isValidationError, lang, r)
	pageModel := mapper.CreateStaticOverviewPage(basePage, dataset, version, fullVersionsList.Items, cfg.EnableMultivariate, topicList, cfg.IsPublishing, enableApprovalButton)
	renderClient.BuildPage(w, pageModel, templateNameStatic)
}
