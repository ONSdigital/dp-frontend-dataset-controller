package clients

import (
	"context"

	topicAPIModels "github.com/ONSdigital/dp-topic-api/models"
	topicAPISDK "github.com/ONSdigital/dp-topic-api/sdk"
	topicAPIErrors "github.com/ONSdigital/dp-topic-api/sdk/errors"
)

// TopicClient is an interface with methods required for a topic client
type TopicAPIClient interface {
	GetTopicPublic(ctx context.Context, headers topicAPISDK.Headers, id string) (*topicAPIModels.Topic, topicAPIErrors.Error)
	GetTopicPrivate(ctx context.Context, headers topicAPISDK.Headers, id string) (*topicAPIModels.TopicResponse, topicAPIErrors.Error)
}
