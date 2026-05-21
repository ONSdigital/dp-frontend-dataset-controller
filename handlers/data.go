package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/clients"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	dpHandlers "github.com/ONSdigital/dp-net/v3/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// DatasetData handles requests for JSON dataset data
func DatasetData(datasetAPIClient clients.DatasetAPISdkClient, topicAPIClient clients.TopicAPIClient, isPublishing bool) http.HandlerFunc {
	return dpHandlers.ControllerHandler(func(w http.ResponseWriter, r *http.Request, lang, collectionID, accessToken string) {
		datasetData(r, w, datasetAPIClient, topicAPIClient, isPublishing, accessToken)
	})
}

func datasetData(r *http.Request, w http.ResponseWriter, datasetAPIClient clients.DatasetAPISdkClient, topicAPIClient clients.TopicAPIClient, isPublishing bool, accessToken string) {
	ctx := r.Context()

	vars := mux.Vars(r)
	topicSlug := vars["topic"]
	datasetID := vars["datasetID"]

	logData := log.Data{
		"topicSlug": topicSlug,
		"datasetID": datasetID,
	}

	datasetAPIClientHeaders := datasetAPISDK.Headers{AccessToken: accessToken}

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

	topicList, err := clients.FetchTopics(ctx, topicAPIClient, dataset.Topics, isPublishing, accessToken)
	if err != nil {
		log.Error(ctx, "failed to fetch topics", err, logData)
		setStatusCode(ctx, w, err)
		return
	}

	topicSlugs := helpers.ExtractTopicSlugs(topicList)
	if len(topicSlugs) == 0 || topicSlugs[0] != topicSlug {
		log.Error(ctx, "dataset topic does not match URL topic", errDatasetTopicMismatch, logData)
		setStatusCode(ctx, w, errDatasetTopicMismatch)
		return
	}

	mappedDataset, err := mapper.MapStaticDatasetToZebedee(ctx, dataset, topicSlugs)
	if err != nil {
		log.Error(ctx, "failed to map static dataset to zebedee format", err, logData)
		setStatusCode(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(mappedDataset); err != nil {
		log.Error(ctx, "failed to encode dataset data to JSON", err, logData)
	}
}

// EditionData handles requests for JSON edition data
func EditionData(datasetAPIClient clients.DatasetAPISdkClient, topicAPIClient clients.TopicAPIClient, isPublishing bool) http.HandlerFunc {
	return dpHandlers.ControllerHandler(func(w http.ResponseWriter, r *http.Request, lang, collectionID, accessToken string) {
		editionData(r, w, datasetAPIClient, topicAPIClient, isPublishing, accessToken)
	})
}

func editionData(r *http.Request, w http.ResponseWriter, datasetAPIClient clients.DatasetAPISdkClient, topicAPIClient clients.TopicAPIClient, isPublishing bool, accessToken string) {
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

	datasetAPIClientHeaders := datasetAPISDK.Headers{AccessToken: accessToken}

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

	topicList, err := clients.FetchTopics(ctx, topicAPIClient, dataset.Topics, isPublishing, accessToken)
	if err != nil {
		log.Error(ctx, "failed to fetch topics", err, logData)
		setStatusCode(ctx, w, err)
		return
	}

	topicSlugs := helpers.ExtractTopicSlugs(topicList)
	if len(topicSlugs) == 0 || topicSlugs[0] != topicSlug {
		log.Error(ctx, "dataset topic does not match URL topic", errDatasetTopicMismatch, logData)
		setStatusCode(ctx, w, errDatasetTopicMismatch)
		return
	}

	// TODO: Fetch versions using GetVersionsInBatches to improve performance and avoid edge case of having more than 1000 versions.
	versions, err := datasetAPIClient.GetVersions(ctx, datasetAPIClientHeaders, datasetID, editionID, &datasetAPISDK.QueryParams{Limit: 1000})
	if err != nil {
		log.Error(ctx, "failed to fetch versions", err, logData)
		setStatusCode(ctx, w, err)
		return
	}

	// WARNING: The dataset API orders versions by "last_updated" in descending order.
	// Therefore we cannot be sure the first item in the versions list is the latest version or if the versions are ordered by version number.
	//
	// TODO: This potential bug will be resolved once the dataset API supports ordering by version number.
	latestVersion := versions.Items[0]
	previousVersions := versions.Items[1:]

	mappedVersion, err := mapper.MapStaticVersionToZebedee(dataset, latestVersion, previousVersions, topicSlugs)
	if err != nil {
		log.Error(ctx, "failed to map static version to zebedee format", err, logData)
		setStatusCode(ctx, w, err)
		return
	}

	// MapStaticVersionToZebedee sets the URI to the version but this endpoint is for an edition so it needs to be set to the edition URI.
	mappedVersion.URI = fmt.Sprintf("/%s/datasets/%s/editions/%s", topicSlugs[0], datasetID, editionID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(mappedVersion); err != nil {
		log.Error(ctx, "failed to encode edition data to JSON", err, logData)
	}
}

// VersionData handles requests for JSON version data
func VersionData(datasetAPIClient clients.DatasetAPISdkClient, topicAPIClient clients.TopicAPIClient, isPublishing bool) http.HandlerFunc {
	return dpHandlers.ControllerHandler(func(w http.ResponseWriter, r *http.Request, lang, collectionID, accessToken string) {
		versionData(r, w, datasetAPIClient, topicAPIClient, isPublishing, accessToken)
	})
}

func versionData(r *http.Request, w http.ResponseWriter, datasetAPIClient clients.DatasetAPISdkClient, topicAPIClient clients.TopicAPIClient, isPublishing bool, accessToken string) {
	ctx := r.Context()

	vars := mux.Vars(r)
	topicSlug := vars["topic"]
	datasetID := vars["datasetID"]
	editionID := vars["editionID"]
	versionID := vars["versionID"]

	logData := log.Data{
		"topicSlug": topicSlug,
		"datasetID": datasetID,
		"editionID": editionID,
		"versionID": versionID,
	}

	datasetAPIClientHeaders := datasetAPISDK.Headers{AccessToken: accessToken}

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

	topicList, err := clients.FetchTopics(ctx, topicAPIClient, dataset.Topics, isPublishing, accessToken)
	if err != nil {
		log.Error(ctx, "failed to fetch topics", err, logData)
		setStatusCode(ctx, w, err)
		return
	}

	topicSlugs := helpers.ExtractTopicSlugs(topicList)
	if len(topicSlugs) == 0 || topicSlugs[0] != topicSlug {
		log.Error(ctx, "dataset topic does not match URL topic", errDatasetTopicMismatch, logData)
		setStatusCode(ctx, w, errDatasetTopicMismatch)
		return
	}

	version, err := datasetAPIClient.GetVersionV2(ctx, datasetAPIClientHeaders, datasetID, editionID, versionID)
	if err != nil {
		log.Error(ctx, "failed to fetch version", err, logData)
		setStatusCode(ctx, w, err)
		return
	}

	var previousVersions []datasetAPIModels.Version

	// TODO: Fetch versions using GetVersionsInBatches to improve performance once the dataset API SDK supports query params for GetVersionsInBatches.
	if version.Version > 1 {
		if dataset.Links == nil || dataset.Links.LatestVersion == nil || dataset.Links.LatestVersion.ID == "" {
			log.Error(ctx, "dataset does not have a latest version link", errMissingLatestVersionLink, logData)
			setStatusCode(ctx, w, errMissingLatestVersionLink)
			return
		}

		latestVersionNumber, err := strconv.Atoi(dataset.Links.LatestVersion.ID)
		if err != nil {
			log.Error(ctx, "failed to convert latest version ID to integer", err, logData)
			setStatusCode(ctx, w, err)
			return
		}

		// WARNING: The dataset API orders versions by "last_updated" in descending order.
		// Therefore we cannot be sure the versions are ordered by version number and the returned version list may not contain the expected versions.
		//
		// TODO: This potential bug will be resolved once the dataset API supports ordering by version number.
		datasetAPIQueryParams := &datasetAPISDK.QueryParams{
			Limit:  version.Version - 1,
			Offset: latestVersionNumber - version.Version + 1,
		}

		// TODO: This potential bug will be resolved once we switch from GetVersions to GetVersionsInBatches in the dataset API SDK.
		if datasetAPIQueryParams.Limit > 1000 {
			log.Warn(ctx, "number of previous versions exceeds dataset API limit, only the latest 1000 versions will be returned", logData)
			datasetAPIQueryParams.Limit = 1000
		}

		previousVersionsList, err := datasetAPIClient.GetVersions(ctx, datasetAPIClientHeaders, datasetID, editionID, datasetAPIQueryParams)
		if err != nil {
			log.Error(ctx, "failed to fetch previous versions", err, logData)
			setStatusCode(ctx, w, err)
			return
		}
		previousVersions = previousVersionsList.Items
	}

	mappedVersion, err := mapper.MapStaticVersionToZebedee(dataset, version, previousVersions, topicSlugs)
	if err != nil {
		log.Error(ctx, "failed to map static version to zebedee format", err, logData)
		setStatusCode(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(mappedVersion); err != nil {
		log.Error(ctx, "failed to encode version data to JSON", err, logData)
	}
}
