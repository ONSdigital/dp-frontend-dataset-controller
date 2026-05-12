package handlers

import (
	"encoding/json"
	"net/http"

	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/clients"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	dpHandlers "github.com/ONSdigital/dp-net/v3/handlers"
	topicAPISDK "github.com/ONSdigital/dp-topic-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// DatasetData handles requests for JSON dataset data
func DatasetData(datasetAPIClient clients.DatasetAPISdkClient, topicAPIClient clients.TopicAPIClient) http.HandlerFunc {
	return dpHandlers.ControllerHandler(func(w http.ResponseWriter, r *http.Request, lang, collectionID, accessToken string) {
		datasetData(r, w, datasetAPIClient, topicAPIClient, accessToken)
	})
}

func datasetData(r *http.Request, w http.ResponseWriter, datasetAPIClient clients.DatasetAPISdkClient, topicAPIClient clients.TopicAPIClient, accessToken string) {
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

	// Topics is a mandatory field so dataset.Topics[0] should always exist
	canonicalTopic, err := topicAPIClient.GetTopicPublic(ctx, topicAPISDK.Headers{}, dataset.Topics[0])
	if err != nil {
		log.Error(ctx, "failed to fetch topic", err, logData)
		setStatusCode(ctx, w, err)
		return
	}

	if canonicalTopic.Slug != topicSlug {
		log.Error(ctx, "dataset topic does not match URL topic", errDatasetTopicMismatch, logData)
		setStatusCode(ctx, w, errDatasetTopicMismatch)
		return
	}

	mappedDataset, err := mapper.MapStaticDatasetToZebedee(ctx, dataset, topicAPIClient)
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
