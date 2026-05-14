package clients

import (
	"context"
	"errors"
	"net/http"
	"testing"

	topicAPIModels "github.com/ONSdigital/dp-topic-api/models"
	topicAPISDK "github.com/ONSdigital/dp-topic-api/sdk"
	topicAPISDKErrors "github.com/ONSdigital/dp-topic-api/sdk/errors"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testTopic1 = topicAPIModels.Topic{ID: "topic1", Slug: "topic1-slug"}
	testTopic2 = topicAPIModels.Topic{ID: "topic2", Slug: "topic2-slug"}
)

func TestFetchTopics(t *testing.T) {
	Convey("Given a list of topic IDs, topicAPIClient, and userAccessToken", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		mockTopicAPIClient := NewMockTopicAPIClient(ctrl)
		userAccessToken := "testAccessToken"

		Convey("When FetchTopics is called successfully for all topic IDs in publishing mode", func() {
			mockTopicAPIClient.EXPECT().GetTopicPrivate(ctx, topicAPISDK.Headers{UserAuthToken: userAccessToken}, "topic1").
				Return(&topicAPIModels.TopicResponse{Current: &testTopic1}, nil)
			mockTopicAPIClient.EXPECT().GetTopicPrivate(ctx, topicAPISDK.Headers{UserAuthToken: userAccessToken}, "topic2").
				Return(&topicAPIModels.TopicResponse{Current: &testTopic2}, nil)

			topics, err := FetchTopics(ctx, mockTopicAPIClient, []string{"topic1", "topic2"}, true, userAccessToken)

			Convey("Then the correct topics are returned and error is nil", func() {
				So(err, ShouldBeNil)
				So(topics, ShouldResemble, []*topicAPIModels.Topic{&testTopic1, &testTopic2})
			})
		})

		Convey("When FetchTopics is called successfully for all topic IDs in web mode", func() {
			mockTopicAPIClient.EXPECT().GetTopicPublic(ctx, topicAPISDK.Headers{UserAuthToken: userAccessToken}, "topic1").
				Return(&testTopic1, nil)
			mockTopicAPIClient.EXPECT().GetTopicPublic(ctx, topicAPISDK.Headers{UserAuthToken: userAccessToken}, "topic2").
				Return(&testTopic2, nil)

			topics, err := FetchTopics(ctx, mockTopicAPIClient, []string{"topic1", "topic2"}, false, userAccessToken)

			Convey("Then the correct topics are returned and error is nil", func() {
				So(err, ShouldBeNil)
				So(topics, ShouldResemble, []*topicAPIModels.Topic{&testTopic1, &testTopic2})
			})
		})

		Convey("When FetchTopics fails for any topic ID in publishing mode", func() {
			mockTopicAPIClient.EXPECT().GetTopicPrivate(ctx, topicAPISDK.Headers{UserAuthToken: userAccessToken}, "topic1").
				Return(nil, topicAPISDKErrors.StatusError{Code: http.StatusInternalServerError, Err: errors.New("something failed")})

			topics, err := FetchTopics(ctx, mockTopicAPIClient, []string{"topic1", "topic2"}, true, userAccessToken)

			Convey("Then an error is returned and topics is nil", func() {
				So(err, ShouldNotBeNil)
				So(topics, ShouldBeNil)
			})
		})

		Convey("When FetchTopics fails for any topic ID in web mode", func() {
			mockTopicAPIClient.EXPECT().GetTopicPublic(ctx, topicAPISDK.Headers{UserAuthToken: userAccessToken}, "topic1").
				Return(nil, topicAPISDKErrors.StatusError{Code: http.StatusInternalServerError, Err: errors.New("something failed")})

			topics, err := FetchTopics(ctx, mockTopicAPIClient, []string{"topic1", "topic2"}, false, userAccessToken)

			Convey("Then an error is returned and topics is nil", func() {
				So(err, ShouldNotBeNil)
				So(topics, ShouldBeNil)
			})
		})

		Convey("When FetchTopics is called with a nil topics list", func() {
			topics, err := FetchTopics(ctx, mockTopicAPIClient, nil, true, userAccessToken)

			Convey("Then an empty slice is returned and error is nil", func() {
				So(err, ShouldBeNil)
				So(topics, ShouldBeEmpty)
			})
		})

		Convey("When FetchTopics is called with an empty topics list", func() {
			topics, err := FetchTopics(ctx, mockTopicAPIClient, []string{}, false, userAccessToken)

			Convey("Then an empty slice is returned and error is nil", func() {
				So(err, ShouldBeNil)
				So(topics, ShouldBeEmpty)
			})
		})
	})
}
