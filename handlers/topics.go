package handlers

import (
	"context"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	topicAPIModels "github.com/ONSdigital/dp-topic-api/models"
	topicAPISDK "github.com/ONSdigital/dp-topic-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

// fetchTopics retrieves the topic from the topic API for each topic ID provided and returns a slice of topics.
// If any request to the topic API fails, an empty slice is returned and the error is logged.
// This is because breadcrumbs cannot be constructed without all topics.
func fetchTopics(ctx context.Context, cfg config.Config, topicAPIClient TopicAPIClient, topics []string, userAccessToken string) []topicAPIModels.Topic {
	fetchedTopics := make([]topicAPIModels.Topic, 0, len(topics))

	for _, topicID := range topics {
		topic, err := GetPublicOrPrivateTopics(topicAPIClient, cfg, ctx, topicAPISDK.Headers{UserAuthToken: userAccessToken}, topicID)
		if err != nil {
			log.Warn(ctx, "failed to fetch topic, returning empty topics list", log.Data{"topicID": topicID, "error": err})
			return nil
		}
		fetchedTopics = append(fetchedTopics, *topic)
	}

	return fetchedTopics
}
