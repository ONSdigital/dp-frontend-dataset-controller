package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	core "github.com/ONSdigital/dis-design-system-go/model"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	authMock "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
	topicAPIModels "github.com/ONSdigital/dp-topic-api/models"
	topicAPISDK "github.com/ONSdigital/dp-topic-api/sdk"
	"github.com/gorilla/mux"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testUserAccessToken  = "user-token"
	testAdminAccessToken = "admin-token"

	testAdminDatasetSDKHeaders = datasetAPISDK.Headers{AccessToken: testAdminAccessToken}
	testUserDatasetSDKHeaders  = datasetAPISDK.Headers{AccessToken: testUserAccessToken}
)

func TestStaticLanding(t *testing.T) {
	cfg := initialiseMockConfig()
	cfg.IsPublishing = true
	ctx := gomock.Any()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDatasetClient := NewMockDatasetAPISdkClient(ctrl)
	mockRenderClient := NewMockRenderClient(ctrl)
	mockZebedeeClient := NewMockZebedeeClient(ctrl)
	mockTopicAPIClient := NewMockTopicAPIClient(ctrl)
	mockAuthMiddleware := &authMock.MiddlewareMock{
		ParseFunc: func(token string) (*permissionsAPISDK.EntityData, error) {
			switch token {
			case testAdminAccessToken:
				return &permissionsAPISDK.EntityData{
					Groups: []string{"role-admin"},
				}, nil
			case testUserAccessToken:
				return &permissionsAPISDK.EntityData{}, nil
			}
			return nil, errors.New("failed to parse token")
		},
	}

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
	editionID := "2025"
	versionID := "1"
	version := datasetAPIModels.Version{
		Distributions: &[]datasetAPIModels.Distribution{
			{
				Format:      "csv",
				DownloadURL: "/path/to/download.csv",
			},
		},
	}
	versionList := datasetAPISDK.VersionsList{}
	collectionID := ""
	lang := "en"

	Convey("Given a successful request to the static landing page handler as an admin", t, func() {
		mockDatasetClient.EXPECT().GetDataset(ctx, testAdminDatasetSDKHeaders, datasetID).
			Return(dataset, nil)
		mockDatasetClient.EXPECT().GetVersionV2(ctx, testAdminDatasetSDKHeaders, datasetID, editionID, versionID).
			Return(version, nil)
		mockDatasetClient.EXPECT().GetVersions(ctx, testAdminDatasetSDKHeaders, datasetID, editionID, &datasetAPISDK.QueryParams{Limit: 1000}).
			Return(versionList, nil)

		mockTopicAPIClient.EXPECT().GetTopicPrivate(ctx, topicAPISDK.Headers{UserAuthToken: testAdminAccessToken}, "topic1").
			Return(&topicAPIModels.TopicResponse{Current: &testTopic1}, nil)
		mockTopicAPIClient.EXPECT().GetTopicPrivate(ctx, topicAPISDK.Headers{UserAuthToken: testAdminAccessToken}, "topic2").
			Return(&topicAPIModels.TopicResponse{Current: &testTopic2}, nil)

		mockZebedeeClient.EXPECT().GetHomepageContent(ctx, testAdminAccessToken, collectionID, lang, homepagePath).
			Return(zebedee.HomepageContent{}, nil)

		mockRenderClient.EXPECT().NewBasePageModel().
			Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
		mockRenderClient.EXPECT().BuildPage(ctx, gomock.Any(), templateNameStatic)

		Convey("When the StaticLanding handler is called", func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/datasets/%s/editions/%s/versions/%s", "topic1", datasetID, editionID, versionID), http.NoBody)
			r = mux.SetURLVars(r, map[string]string{
				"topic":     "topic1",
				"datasetID": datasetID,
				"editionID": editionID,
				"versionID": versionID,
			})
			staticLanding(r, w, mockDatasetClient, mockRenderClient, mockZebedeeClient, mockTopicAPIClient, cfg, mockAuthMiddleware, testAdminAccessToken, lang, collectionID)

			Convey("Then the response status code should be 200 OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})
	})

	Convey("When GetDataset fails", t, func() {
		mockDatasetClient.EXPECT().GetDataset(ctx, testUserDatasetSDKHeaders, datasetID).
			Return(datasetAPIModels.Dataset{}, errors.New("GetDataset failed"))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/datasets/%s/editions/%s/versions/%s", "topic1", datasetID, editionID, versionID), http.NoBody)
		r = mux.SetURLVars(r, map[string]string{
			"topic":     "topic1",
			"datasetID": datasetID,
			"editionID": editionID,
			"versionID": versionID,
		})

		staticLanding(r, w, mockDatasetClient, mockRenderClient, mockZebedeeClient, mockTopicAPIClient, cfg, mockAuthMiddleware, testUserAccessToken, lang, collectionID)

		Convey("Then the response status code should be 500 Internal Server Error", func() {
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})

	Convey("When the dataset is not of type static", t, func() {
		mockDatasetClient.EXPECT().GetDataset(ctx, testUserDatasetSDKHeaders, datasetID).
			Return(datasetAPIModels.Dataset{Type: "filterable"}, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/datasets/%s/editions/%s/versions/%s", "topic1", datasetID, editionID, versionID), http.NoBody)
		r = mux.SetURLVars(r, map[string]string{
			"topic":     "topic1",
			"datasetID": datasetID,
			"editionID": editionID,
			"versionID": versionID,
		})

		staticLanding(r, w, mockDatasetClient, mockRenderClient, mockZebedeeClient, mockTopicAPIClient, cfg, mockAuthMiddleware, testUserAccessToken, lang, collectionID)

		Convey("Then the response status code should be 404 Not Found", func() {
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
	})

	Convey("When the first topic in the dataset does not match the topic in the URL", t, func() {
		mockDatasetClient.EXPECT().GetDataset(ctx, testUserDatasetSDKHeaders, datasetID).
			Return(dataset, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/datasets/%s/editions/%s/versions/%s", "different-topic", datasetID, editionID, versionID), http.NoBody)
		r = mux.SetURLVars(r, map[string]string{
			"topic":     "different-topic",
			"datasetID": datasetID,
			"editionID": editionID,
			"versionID": versionID,
		})

		staticLanding(r, w, mockDatasetClient, mockRenderClient, mockZebedeeClient, mockTopicAPIClient, cfg, mockAuthMiddleware, testUserAccessToken, lang, collectionID)

		Convey("Then the response status code should be 404 Not Found", func() {
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
	})

	// Request from /{topic}/datasets/{datasetID}/editions/{editionID}
	Convey("When versionID not provided in the URL, redirect to latest version", t, func() {
		mockDatasetClient.EXPECT().GetDataset(ctx, testUserDatasetSDKHeaders, datasetID).
			Return(dataset, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/datasets/%s/editions/%s", "topic1", datasetID, editionID), http.NoBody)
		r = mux.SetURLVars(r, map[string]string{
			"topic":     "topic1",
			"datasetID": datasetID,
			"editionID": editionID,
		})

		staticLanding(r, w, mockDatasetClient, mockRenderClient, mockZebedeeClient, mockTopicAPIClient, cfg, mockAuthMiddleware, testUserAccessToken, lang, collectionID)

		Convey("Then the response should be a redirect to the latest version", func() {
			So(w.Code, ShouldEqual, http.StatusFound)
			expectedLocation := fmt.Sprintf("/%s/datasets/%s/editions/%s/versions/%s", "topic1", datasetID, editionID, dataset.Links.LatestVersion.ID)
			So(w.Header().Get("Location"), ShouldEqual, expectedLocation)
		})
	})

	// Request from /{topic}/datasets/{datasetID}/editions/{editionID}/versions
	Convey("When versionID not provided in the URL, redirect to latest version", t, func() {
		mockDatasetClient.EXPECT().GetDataset(ctx, testUserDatasetSDKHeaders, datasetID).
			Return(dataset, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/datasets/%s/editions/%s/versions", "topic1", datasetID, editionID), http.NoBody)
		r = mux.SetURLVars(r, map[string]string{
			"topic":     "topic1",
			"datasetID": datasetID,
			"editionID": editionID,
		})

		staticLanding(r, w, mockDatasetClient, mockRenderClient, mockZebedeeClient, mockTopicAPIClient, cfg, mockAuthMiddleware, testUserAccessToken, lang, collectionID)

		Convey("Then the response should be a redirect to the latest version", func() {
			So(w.Code, ShouldEqual, http.StatusFound)
			expectedLocation := fmt.Sprintf("/%s/datasets/%s/editions/%s/versions/%s", "topic1", datasetID, editionID, dataset.Links.LatestVersion.ID)
			So(w.Header().Get("Location"), ShouldEqual, expectedLocation)
		})
	})

	Convey("When GetVersionV2 fails", t, func() {
		mockDatasetClient.EXPECT().GetDataset(ctx, testUserDatasetSDKHeaders, datasetID).
			Return(dataset, nil)

		mockDatasetClient.EXPECT().GetVersionV2(ctx, testUserDatasetSDKHeaders, datasetID, editionID, versionID).
			Return(datasetAPIModels.Version{}, errors.New("GetVersionV2 failed"))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/datasets/%s/editions/%s/versions/%s", "topic1", datasetID, editionID, versionID), http.NoBody)

		r = mux.SetURLVars(r, map[string]string{
			"topic":     "topic1",
			"datasetID": datasetID,
			"editionID": editionID,
			"versionID": versionID,
		})

		staticLanding(r, w, mockDatasetClient, mockRenderClient, mockZebedeeClient, mockTopicAPIClient, cfg, mockAuthMiddleware, testUserAccessToken, lang, collectionID)

		Convey("Then the response status code should be 500 Internal Server Error", func() {
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})

	// In this event a log is created and isValidationError is set to true
	Convey("When Download request has no format query parameter or format doesn't match any distribution format then request is successful", t, func() {
		mockDatasetClient.EXPECT().GetDataset(ctx, testUserDatasetSDKHeaders, datasetID).
			Return(dataset, nil)

		mockDatasetClient.EXPECT().GetVersionV2(ctx, testUserDatasetSDKHeaders, datasetID, editionID, versionID).
			Return(version, nil)

		mockDatasetClient.EXPECT().GetVersions(ctx, testUserDatasetSDKHeaders, datasetID, editionID, &datasetAPISDK.QueryParams{Limit: 1000}).
			Return(versionList, nil)

		mockTopicAPIClient.EXPECT().GetTopicPrivate(ctx, topicAPISDK.Headers{UserAuthToken: testUserAccessToken}, "topic1").
			Return(&topicAPIModels.TopicResponse{Current: &testTopic1}, nil)
		mockTopicAPIClient.EXPECT().GetTopicPrivate(ctx, topicAPISDK.Headers{UserAuthToken: testUserAccessToken}, "topic2").
			Return(&topicAPIModels.TopicResponse{Current: &testTopic2}, nil)

		mockZebedeeClient.EXPECT().GetHomepageContent(ctx, testUserAccessToken, collectionID, lang, homepagePath).
			Return(zebedee.HomepageContent{}, nil)

		mockRenderClient.EXPECT().NewBasePageModel().
			Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
		mockRenderClient.EXPECT().BuildPage(ctx, gomock.Any(), templateNameStatic)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/datasets/%s/editions/%s/versions/%s?f=get-data", "topic1", datasetID, editionID, versionID), http.NoBody)
		r = mux.SetURLVars(r, map[string]string{
			"topic":     "topic1",
			"datasetID": datasetID,
			"editionID": editionID,
			"versionID": versionID,
		})

		staticLanding(r, w, mockDatasetClient, mockRenderClient, mockZebedeeClient, mockTopicAPIClient, cfg, mockAuthMiddleware, testUserAccessToken, lang, collectionID)

		Convey("Then the response status code should be 200 OK", func() {
			So(w.Code, ShouldEqual, http.StatusOK)
		})
	})

	Convey("Download request gets redirected to the download URL provided in the version", t, func() {
		mockDatasetClient.EXPECT().GetDataset(ctx, testUserDatasetSDKHeaders, datasetID).
			Return(dataset, nil)

		mockDatasetClient.EXPECT().GetVersionV2(ctx, testUserDatasetSDKHeaders, datasetID, editionID, versionID).
			Return(version, nil)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/datasets/%s/editions/%s/versions/%s?f=get-data&format=csv", "topic1", datasetID, editionID, versionID), http.NoBody)
		r = mux.SetURLVars(r, map[string]string{
			"topic":     "topic1",
			"datasetID": datasetID,
			"editionID": editionID,
			"versionID": versionID,
		})

		staticLanding(r, w, mockDatasetClient, mockRenderClient, mockZebedeeClient, mockTopicAPIClient, cfg, mockAuthMiddleware, testUserAccessToken, lang, collectionID)

		Convey("Then the response should be a redirect to the download URL", func() {
			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Header().Get("Location"), ShouldEqual, "/path/to/download.csv")
		})
	})

	Convey("When CheckIsAdmin fails due to auth middleware error", t, func() {
		mockDatasetClient.EXPECT().GetDataset(ctx, testUserDatasetSDKHeaders, datasetID).
			Return(dataset, nil)

		mockDatasetClient.EXPECT().GetVersionV2(ctx, testUserDatasetSDKHeaders, datasetID, editionID, versionID).
			Return(version, nil)

		failingAuthMiddleware := &authMock.MiddlewareMock{
			ParseFunc: func(token string) (*permissionsAPISDK.EntityData, error) {
				return nil, errors.New("failed to parse token")
			},
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/datasets/%s/editions/%s/versions/%s", "topic1", datasetID, editionID, versionID), http.NoBody)
		r = mux.SetURLVars(r, map[string]string{
			"topic":     "topic1",
			"datasetID": datasetID,
			"editionID": editionID,
			"versionID": versionID,
		})

		staticLanding(r, w, mockDatasetClient, mockRenderClient, mockZebedeeClient, mockTopicAPIClient, cfg, failingAuthMiddleware, testUserAccessToken, lang, collectionID)

		Convey("Then the response status code should be 500 Internal Server Error", func() {
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})

	Convey("When GetVersions fails", t, func() {
		mockDatasetClient.EXPECT().GetDataset(ctx, testUserDatasetSDKHeaders, datasetID).
			Return(dataset, nil)

		mockDatasetClient.EXPECT().GetVersionV2(ctx, testUserDatasetSDKHeaders, datasetID, editionID, versionID).
			Return(version, nil)

		mockDatasetClient.EXPECT().GetVersions(ctx, testUserDatasetSDKHeaders, datasetID, editionID, &datasetAPISDK.QueryParams{Limit: 1000}).
			Return(datasetAPISDK.VersionsList{}, errors.New("GetVersions failed"))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/datasets/%s/editions/%s/versions/%s", "topic1", datasetID, editionID, versionID), http.NoBody)
		r = mux.SetURLVars(r, map[string]string{
			"topic":     "topic1",
			"datasetID": datasetID,
			"editionID": editionID,
			"versionID": versionID,
		})

		staticLanding(r, w, mockDatasetClient, mockRenderClient, mockZebedeeClient, mockTopicAPIClient, cfg, mockAuthMiddleware, testUserAccessToken, lang, collectionID)

		Convey("Then the response status code should be 500 Internal Server Error", func() {
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})

	Convey("When GetHomepageContent fails then request is still successful", t, func() {
		mockDatasetClient.EXPECT().GetDataset(ctx, testUserDatasetSDKHeaders, datasetID).
			Return(dataset, nil)

		mockDatasetClient.EXPECT().GetVersionV2(ctx, testUserDatasetSDKHeaders, datasetID, editionID, versionID).
			Return(version, nil)

		mockDatasetClient.EXPECT().GetVersions(ctx, testUserDatasetSDKHeaders, datasetID, editionID, &datasetAPISDK.QueryParams{Limit: 1000}).
			Return(versionList, nil)

		mockTopicAPIClient.EXPECT().GetTopicPrivate(ctx, topicAPISDK.Headers{UserAuthToken: testUserAccessToken}, "topic1").
			Return(&topicAPIModels.TopicResponse{Current: &testTopic1}, nil)

		mockTopicAPIClient.EXPECT().GetTopicPrivate(ctx, topicAPISDK.Headers{UserAuthToken: testUserAccessToken}, "topic2").
			Return(&topicAPIModels.TopicResponse{Current: &testTopic2}, nil)

		mockZebedeeClient.EXPECT().GetHomepageContent(ctx, testUserAccessToken, collectionID, lang, homepagePath).
			Return(zebedee.HomepageContent{}, errors.New("GetHomepageContent failed"))

		mockRenderClient.EXPECT().NewBasePageModel().
			Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
		mockRenderClient.EXPECT().BuildPage(ctx, gomock.Any(), templateNameStatic)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/datasets/%s/editions/%s/versions/%s", "topic1", datasetID, editionID, versionID), http.NoBody)

		r = mux.SetURLVars(r, map[string]string{
			"topic":     "topic1",
			"datasetID": datasetID,
			"editionID": editionID,
			"versionID": versionID,
		})

		staticLanding(r, w, mockDatasetClient, mockRenderClient, mockZebedeeClient, mockTopicAPIClient, cfg, mockAuthMiddleware, testUserAccessToken, lang, collectionID)

		Convey("Then the response status code should be 200 OK", func() {
			So(w.Code, ShouldEqual, http.StatusOK)
		})
	})
}
