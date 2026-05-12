package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/clients"
	topicAPIModels "github.com/ONSdigital/dp-topic-api/models"
	topicAPISDK "github.com/ONSdigital/dp-topic-api/sdk"
	topicAPISDKErrors "github.com/ONSdigital/dp-topic-api/sdk/errors"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDatasetData(t *testing.T) {
	ctx := gomock.Any()
	accessToken := "test-access-token"
	headers := datasetAPISDK.Headers{AccessToken: accessToken}

	datasetID := "dataset-123"

	topicSlug := "economy"
	topicSlug2 := "inflation"
	topicID := "topic-economy"
	topicID2 := "topic-inflation"

	requestPath := "/economy/datasets/dataset-123/data"
	urlVars := map[string]string{"topic": topicSlug, "datasetID": datasetID}

	dataset := datasetAPIModels.Dataset{
		ID:          datasetID,
		Type:        DatasetTypeStatic,
		Title:       "Producer price inflation (MM22)",
		Description: "Dataset summary",
		Keywords:    []string{"producer prices", "input prices"},
		Topics:      []string{topicID, topicID2},
		Contacts: []datasetAPIModels.ContactDetails{
			{
				Name:      "Business Prices team",
				Email:     "business.prices@ons.gov.uk",
				Telephone: "+44 1633 456907",
			},
		},
	}

	expectedResponseBody := &zebedee.DatasetLandingPage{
		Type: zebedee.PageTypeDatasetLandingPage,
		URI:  "/economy/datasets/dataset-123",
		Description: zebedee.Description{
			DatasetID:       "dataset-123",
			Title:           "Producer price inflation (MM22)",
			Summary:         "Dataset summary",
			MetaDescription: "Dataset summary",
			Keywords:        []string{"producer prices", "input prices"},
			Topics:          []string{topicSlug, topicSlug2},
			Contact: zebedee.Contact{
				Name:      "Business Prices team",
				Email:     "business.prices@ons.gov.uk",
				Telephone: "+44 1633 456907",
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDatasetClient := clients.NewMockDatasetAPISdkClient(ctrl)
	mockTopicClient := clients.NewMockTopicAPIClient(ctrl)

	Convey("Given datasetData handler", t, func() {
		Convey("When dataset is static and topic matches", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, headers, datasetID).
				Return(dataset, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, topicAPISDK.Headers{}, topicID).
				Return(&topicAPIModels.Topic{ID: topicID, Slug: topicSlug}, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, topicAPISDK.Headers{}, topicID).
				Return(&topicAPIModels.Topic{ID: topicID, Slug: topicSlug}, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, topicAPISDK.Headers{}, topicID2).
				Return(&topicAPIModels.Topic{ID: topicID2, Slug: topicSlug2}, nil)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			datasetData(r, w, mockDatasetClient, mockTopicClient, accessToken)

			Convey("Then the response status code should be 200 with the expected JSON body", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")

				var resp zebedee.DatasetLandingPage
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				So(err, ShouldBeNil)
				So(&resp, ShouldResemble, expectedResponseBody)
			})
		})

		Convey("When GetDataset fails", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, headers, datasetID).
				Return(datasetAPIModels.Dataset{}, errors.New("failed to fetch dataset"))

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			datasetData(r, w, mockDatasetClient, mockTopicClient, accessToken)

			Convey("Then the response status code should be 500 Internal Server Error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When dataset type is not static", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, headers, datasetID).
				Return(datasetAPIModels.Dataset{ID: datasetID, Type: "filterable"}, nil)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			datasetData(r, w, mockDatasetClient, mockTopicClient, accessToken)

			Convey("Then the response status code should be 404 Not Found", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("When canonical topic does not match topic slug in URL", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, headers, datasetID).
				Return(dataset, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, topicAPISDK.Headers{}, topicID).
				Return(&topicAPIModels.Topic{ID: topicID, Slug: "different-topic"}, nil)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			datasetData(r, w, mockDatasetClient, mockTopicClient, accessToken)

			Convey("Then the response status code should be 404 Not Found", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("When topic client fails", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, headers, datasetID).
				Return(dataset, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, topicAPISDK.Headers{}, topicID).
				Return(nil, topicAPISDKErrors.StatusError{Code: http.StatusInternalServerError, Err: errors.New("topic API error")})

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			datasetData(r, w, mockDatasetClient, mockTopicClient, accessToken)

			Convey("Then the response status code should be 500 Internal Server Error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When mapper fails due to invalid QMI URL", func() {
			datasetWithInvalidQMI := dataset
			datasetWithInvalidQMI.QMI = &datasetAPIModels.GeneralDetails{
				Title:       "QMI",
				Description: "Invalid QMI URL",
				HRef:        "https://[invalid:url",
			}

			mockDatasetClient.EXPECT().GetDataset(ctx, headers, datasetID).
				Return(datasetWithInvalidQMI, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, topicAPISDK.Headers{}, topicID).
				Return(&topicAPIModels.Topic{ID: topicID, Slug: topicSlug}, nil)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			datasetData(r, w, mockDatasetClient, mockTopicClient, accessToken)

			Convey("Then the response status code should be 500 Internal Server Error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}
