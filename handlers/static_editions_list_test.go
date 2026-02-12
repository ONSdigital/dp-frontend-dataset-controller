package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	core "github.com/ONSdigital/dis-design-system-go/model"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	topicAPIModels "github.com/ONSdigital/dp-topic-api/models"
	topicAPISDK "github.com/ONSdigital/dp-topic-api/sdk"
	"github.com/gorilla/mux"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestStaticEditionsList(t *testing.T) {
	cfg := initialiseMockConfig()
	cfg.IsPublishing = true
	ctx := gomock.Any()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDatasetClient := NewMockDatasetAPISdkClient(ctrl)
	mockRenderClient := NewMockRenderClient(ctrl)
	mockZebedeeClient := NewMockZebedeeClient(ctrl)
	mockTopicAPIClient := NewMockTopicAPIClient(ctrl)

	datasetID := "static-dataset"
	dataset := datasetAPIModels.Dataset{
		ID:     datasetID,
		Type:   DatasetTypeStatic,
		Topics: []string{"topic1", "topic2"},
		Links: &datasetAPIModels.DatasetLinks{
			LatestVersion: &datasetAPIModels.LinkObject{
				ID: "2",
			},
		},
	}
	editionID := "2024"
	edition := datasetAPIModels.Edition{
		Edition: "2024",
		Links: &datasetAPIModels.EditionUpdateLinks{
			LatestVersion: dataset.Links.LatestVersion,
		},
	}
	edition2 := datasetAPIModels.Edition{
		Edition: "2025",
		Links: &datasetAPIModels.EditionUpdateLinks{
			LatestVersion: dataset.Links.LatestVersion,
		},
	}
	editionList := datasetAPISDK.EditionsList{
		Items: []datasetAPIModels.Edition{edition, edition2},
	}
	collectionID := ""
	lang := "en"
	apiRouterVersion := "/v1"

	Convey("Given a successful request to the static editions list handler", t, func() {
		mockDatasetClient.EXPECT().GetDataset(ctx, testUserDatasetSDKHeaders, datasetID).
			Return(dataset, nil)
		mockDatasetClient.EXPECT().GetEditions(ctx, testUserDatasetSDKHeaders, datasetID, &datasetAPISDK.QueryParams{Limit: 1000}).
			Return(editionList, nil)

		mockTopicAPIClient.EXPECT().GetTopicPrivate(ctx, topicAPISDK.Headers{UserAuthToken: testUserAccessToken}, "topic1").
			Return(&topicAPIModels.TopicResponse{Current: &testTopic1}, nil)
		mockTopicAPIClient.EXPECT().GetTopicPrivate(ctx, topicAPISDK.Headers{UserAuthToken: testUserAccessToken}, "topic2").
			Return(&topicAPIModels.TopicResponse{Current: &testTopic2}, nil)

		mockZebedeeClient.EXPECT().GetHomepageContent(ctx, testUserAccessToken, collectionID, lang, homepagePath).
			Return(zebedee.HomepageContent{}, nil)

		mockRenderClient.EXPECT().NewBasePageModel().
			Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
		mockRenderClient.EXPECT().BuildPage(ctx, gomock.Any(), templateNameStaticEditionsList)

		Convey("When the StaticEditionsList handler is called", func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/datasets/%s", "topic1", datasetID), http.NoBody)
			r = mux.SetURLVars(r, map[string]string{
				"topic":     "topic1",
				"datasetID": datasetID,
			})

			staticEditionsList(r, w, mockDatasetClient, mockRenderClient, mockZebedeeClient, mockTopicAPIClient, cfg, apiRouterVersion, testUserAccessToken, lang, collectionID)

			Convey("Then the response status code should be 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})
	})

	Convey("When GetDataset fails", t, func() {
		mockDatasetClient.EXPECT().GetDataset(ctx, testUserDatasetSDKHeaders, datasetID).
			Return(datasetAPIModels.Dataset{}, errors.New("GetDataset failed"))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/datasets/%s", "topic1", datasetID), http.NoBody)
		r = mux.SetURLVars(r, map[string]string{
			"topic":     "topic1",
			"datasetID": datasetID,
		})

		staticEditionsList(r, w, mockDatasetClient, mockRenderClient, mockZebedeeClient, mockTopicAPIClient, cfg, apiRouterVersion, testUserAccessToken, lang, collectionID)

		Convey("Then the response status code should be 500 Internal Server Error", func() {
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})

	Convey("When the dataset is not of type static", t, func() {
		mockDatasetClient.EXPECT().GetDataset(ctx, testUserDatasetSDKHeaders, datasetID).
			Return(datasetAPIModels.Dataset{ID: datasetID, Type: "filterable"}, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/datasets/%s", "topic1", datasetID), http.NoBody)
		r = mux.SetURLVars(r, map[string]string{
			"topic":     "topic1",
			"datasetID": datasetID,
		})

		staticEditionsList(r, w, mockDatasetClient, mockRenderClient, mockZebedeeClient, mockTopicAPIClient, cfg, apiRouterVersion, testUserAccessToken, lang, collectionID)

		Convey("Then the response status code should be 404 Not Found", func() {
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
	})

	Convey("When the first topic in the dataset does not match the topic in the URL", t, func() {
		mockDatasetClient.EXPECT().GetDataset(ctx, testUserDatasetSDKHeaders, datasetID).
			Return(dataset, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/datasets/%s", "different-topic", datasetID), http.NoBody)
		r = mux.SetURLVars(r, map[string]string{
			"topic":     "different-topic",
			"datasetID": datasetID,
		})

		staticEditionsList(r, w, mockDatasetClient, mockRenderClient, mockZebedeeClient, mockTopicAPIClient, cfg, apiRouterVersion, testUserAccessToken, lang, collectionID)

		Convey("Then the response status code should be 404 Not Found", func() {
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
	})

	Convey("When the editionID is provided in the URL, redirect to the latest", t, func() {
		mockDatasetClient.EXPECT().GetDataset(ctx, testUserDatasetSDKHeaders, datasetID).
			Return(dataset, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/datasets/%s/editions/%s", "topic1", datasetID, editionID), http.NoBody)
		r = mux.SetURLVars(r, map[string]string{
			"topic":     "topic1",
			"datasetID": datasetID,
			"editionID": editionID,
		})

		staticEditionsList(r, w, mockDatasetClient, mockRenderClient, mockZebedeeClient, mockTopicAPIClient, cfg, apiRouterVersion, testUserAccessToken, lang, collectionID)

		Convey("Then the response status code should be 302 Found and redirect to the latest version", func() {
			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Header().Get("Location"), ShouldEqual, fmt.Sprintf("/%s/datasets/%s/editions/%s/versions/%s", "topic1", datasetID, editionID, dataset.Links.LatestVersion.ID))
		})
	})

	Convey("When GetEditions fails", t, func() {
		mockDatasetClient.EXPECT().GetDataset(ctx, testUserDatasetSDKHeaders, datasetID).
			Return(dataset, nil)

		mockDatasetClient.EXPECT().GetEditions(ctx, testUserDatasetSDKHeaders, datasetID, &datasetAPISDK.QueryParams{Limit: 1000}).
			Return(datasetAPISDK.EditionsList{}, errors.New("GetEditions failed"))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/datasets/%s", "topic1", datasetID), http.NoBody)
		r = mux.SetURLVars(r, map[string]string{
			"topic":     "topic1",
			"datasetID": datasetID,
		})

		staticEditionsList(r, w, mockDatasetClient, mockRenderClient, mockZebedeeClient, mockTopicAPIClient, cfg, apiRouterVersion, testUserAccessToken, lang, collectionID)

		Convey("Then the response status code should be 500 Internal Server Error", func() {
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})

	Convey("When the number of editions is 1, redirect to the latest verison", t, func() {
		mockDatasetClient.EXPECT().GetDataset(ctx, testUserDatasetSDKHeaders, datasetID).
			Return(dataset, nil)

		mockDatasetClient.EXPECT().GetEditions(ctx, testUserDatasetSDKHeaders, datasetID, &datasetAPISDK.QueryParams{Limit: 1000}).
			Return(datasetAPISDK.EditionsList{Items: []datasetAPIModels.Edition{edition}}, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/datasets/%s", "topic1", datasetID), http.NoBody)
		r = mux.SetURLVars(r, map[string]string{
			"topic":     "topic1",
			"datasetID": datasetID,
		})

		staticEditionsList(r, w, mockDatasetClient, mockRenderClient, mockZebedeeClient, mockTopicAPIClient, cfg, apiRouterVersion, testUserAccessToken, lang, collectionID)

		Convey("Then the response status code should be 302 Found and redirect to the latest version", func() {
			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Header().Get("Location"), ShouldEqual, fmt.Sprintf("/%s/datasets/%s/editions/%s/versions/%s", "topic1", datasetID, edition.Edition, edition.Links.LatestVersion.ID))
		})
	})

	Convey("When GetHomepageContent fail, then the request is still successful", t, func() {
		mockDatasetClient.EXPECT().GetDataset(ctx, testUserDatasetSDKHeaders, datasetID).
			Return(dataset, nil)

		mockDatasetClient.EXPECT().GetEditions(ctx, testUserDatasetSDKHeaders, datasetID, &datasetAPISDK.QueryParams{Limit: 1000}).
			Return(editionList, nil)

		mockTopicAPIClient.EXPECT().GetTopicPrivate(ctx, topicAPISDK.Headers{UserAuthToken: testUserAccessToken}, "topic1").
			Return(&topicAPIModels.TopicResponse{Current: &testTopic1}, nil)

		mockTopicAPIClient.EXPECT().GetTopicPrivate(ctx, topicAPISDK.Headers{UserAuthToken: testUserAccessToken}, "topic2").
			Return(&topicAPIModels.TopicResponse{Current: &testTopic2}, nil)

		mockZebedeeClient.EXPECT().GetHomepageContent(ctx, testUserAccessToken, collectionID, lang, homepagePath).
			Return(zebedee.HomepageContent{}, errors.New("GetHomepageContent failed"))

		mockRenderClient.EXPECT().NewBasePageModel().
			Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))

		mockRenderClient.EXPECT().BuildPage(ctx, gomock.Any(), templateNameStaticEditionsList)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/datasets/%s", "topic1", datasetID), http.NoBody)
		r = mux.SetURLVars(r, map[string]string{
			"topic":     "topic1",
			"datasetID": datasetID,
		})

		staticEditionsList(r, w, mockDatasetClient, mockRenderClient, mockZebedeeClient, mockTopicAPIClient, cfg, apiRouterVersion, testUserAccessToken, lang, collectionID)

		Convey("Then the response status code should be 200 OK", func() {
			So(w.Code, ShouldEqual, http.StatusOK)
		})
	})
}
