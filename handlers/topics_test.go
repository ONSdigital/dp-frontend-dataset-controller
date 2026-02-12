package handlers

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"

	topicAPIModels "github.com/ONSdigital/dp-topic-api/models"
	topicAPISDK "github.com/ONSdigital/dp-topic-api/sdk"
	topicAPISDKErrors "github.com/ONSdigital/dp-topic-api/sdk/errors"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testTopic1 = topicAPIModels.Topic{ID: "topic1"}
	testTopic2 = topicAPIModels.Topic{ID: "topic2"}
)

func TestFetchTopics(t *testing.T) {
	Convey("Given a list of topic IDs, topicAPIClient, config and userAccessToken", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		mockTopicAPIClient := NewMockTopicAPIClient(ctrl)
		cfg := config.Config{IsPublishing: true}
		userAccessToken := "testAccessToken"

		Convey("When fetchTopics is called successfully for all topic IDs", func() {
			mockTopicAPIClient.EXPECT().GetTopicPrivate(ctx, topicAPISDK.Headers{UserAuthToken: userAccessToken}, "topic1").
				Return(&topicAPIModels.TopicResponse{Current: &testTopic1}, nil)
			mockTopicAPIClient.EXPECT().GetTopicPrivate(ctx, topicAPISDK.Headers{UserAuthToken: userAccessToken}, "topic2").
				Return(&topicAPIModels.TopicResponse{Current: &testTopic2}, nil)

			topics := fetchTopics(ctx, cfg, mockTopicAPIClient, []string{"topic1", "topic2"}, userAccessToken)

			Convey("Then the correct topics are returned", func() {
				So(topics, ShouldResemble, []topicAPIModels.Topic{testTopic1, testTopic2})
			})
		})

		Convey("When fetchTopics fails for any topic ID", func() {
			mockTopicAPIClient.EXPECT().GetTopicPrivate(ctx, topicAPISDK.Headers{UserAuthToken: userAccessToken}, "topic1").
				Return(nil, topicAPISDKErrors.StatusError{Code: http.StatusInternalServerError, Err: errors.New("something failed")})

			topics := fetchTopics(ctx, cfg, mockTopicAPIClient, []string{"topic1", "topic2"}, userAccessToken)

			Convey("Then an empty slice is returned", func() {
				So(topics, ShouldBeEmpty)
			})
		})

		Convey("When fetchTopics is called with an nil topics list", func() {
			topics := fetchTopics(ctx, cfg, mockTopicAPIClient, nil, userAccessToken)

			Convey("Then an empty slice is returned", func() {
				So(topics, ShouldBeEmpty)
			})
		})
	})
}
