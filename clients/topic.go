package clients

import (
	"context"
	"fmt"

	topicAPIModels "github.com/ONSdigital/dp-topic-api/models"
	topicAPISDK "github.com/ONSdigital/dp-topic-api/sdk"
	topicAPIErrors "github.com/ONSdigital/dp-topic-api/sdk/errors"
)

// TopicClient is an interface with methods required for a topic client
type TopicAPIClient interface {
	GetTopicPublic(ctx context.Context, headers topicAPISDK.Headers, id string) (*topicAPIModels.Topic, topicAPIErrors.Error)
	GetTopicPrivate(ctx context.Context, headers topicAPISDK.Headers, id string) (*topicAPIModels.TopicResponse, topicAPIErrors.Error)
}

// FetchTopic retrieves a single topic from the topic API.
// The userAccessToken is only required when isPublishing is true.
func FetchTopic(ctx context.Context, topicAPIClient TopicAPIClient, topicID string, isPublishing bool, userAccessToken string) (*topicAPIModels.Topic, error) {
	headers := topicAPISDK.Headers{
		UserAuthToken: userAccessToken,
	}

	if isPublishing {
		resp, err := topicAPIClient.GetTopicPrivate(ctx, headers, topicID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch topic with id %s: %w", topicID, err)
		}
		return resp.Current, nil
	}

	topic, err := topicAPIClient.GetTopicPublic(ctx, headers, topicID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch topic with id %s: %w", topicID, err)
	}
	return topic, nil
}

// FetchTopics retrieves a list of topics from the topic API for the given topic IDs.
// The userAccessToken is only required when isPublishing is true.
func FetchTopics(ctx context.Context, topicAPIClient TopicAPIClient, topicIDs []string, isPublishing bool, userAccessToken string) ([]*topicAPIModels.Topic, error) {
	fetchedTopics := make([]*topicAPIModels.Topic, 0, len(topicIDs))

	for _, topicID := range topicIDs {
		topic, err := FetchTopic(ctx, topicAPIClient, topicID, isPublishing, userAccessToken)
		if err != nil {
			return nil, err
		}
		fetchedTopics = append(fetchedTopics, topic)
	}

	return fetchedTopics, nil
}
