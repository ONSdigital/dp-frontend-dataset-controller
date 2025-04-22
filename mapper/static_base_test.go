package mapper

import (
	"bytes"
	"context"
	"encoding/json"
	io "io"
	"net/http"
	"testing"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/dp-topic-api/models"
	"github.com/ONSdigital/dp-topic-api/sdk"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	testPublicTopic1 = models.Topic{
		ID:          "1234",
		Description: "Root Topic 1",
		Title:       "Root Topic 1",
		Slug:        "roottopic2",
		Keywords:    &[]string{"test"},
		State:       "published",
	}
)

const (
	service = "dp-topic-api"
	testHost = "http://localhost:25700"
)

func newMockHTTPClient(r *http.Response, err error) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		SetPathsWithNoRetriesFunc: func(paths []string) {},
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return r, err
		},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{"/healthcheck"}
		},
	}
}

func newTopicAPIClient(_ *testing.T, httpClient *dphttp.ClienterMock) *sdk.Client {
	healthClient := healthcheck.NewClientWithClienter(service, testHost, httpClient)
	return sdk.NewWithHealthClient(healthClient)
}

func TestGetTopicPublic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("Given public root topic is returned successfully", t, func() {
		body, err := json.Marshal(testPublicTopic1)
		if err != nil {
			t.Errorf("failed to setup test data, error: %v", err)
		}

		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(body)),
			},
			nil)

		topicAPIClient := newTopicAPIClient(t, httpClient)

		Convey("When GetTopicPublic is called", func() {
			respTopic, err := topicAPIClient.GetTopicPublic(ctx, sdk.Headers{}, "1234")

			Convey("Then the expected public root topics is returned", func() {
				So(*respTopic, ShouldResemble, testPublicTopic1)

				Convey("And no error is returned", func() {
					So(err, ShouldBeNil)

					Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						So(doCalls, ShouldHaveLength, 1)
						So(doCalls[0].Req.URL.Path, ShouldEqual, "/topics/1234")
					})
				})
			})
		})
	})
}