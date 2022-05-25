package handlers

import (
	"errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	coreModel "github.com/ONSdigital/dp-renderer/model"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLegacyLanding(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()
	cfg := initialiseMockConfig()

	Convey("test /data endpoint", t, func() {
		Convey("test successful json response", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockFilesAPIClient := NewMockFilesAPIClient(mockCtrl)
			mockZebedeeClient.EXPECT().Get(ctx, "12345", "/data?uri=/data").Return([]byte(`{"some_json":true}`), nil)
			mockConfig := config.Config{}

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/data", nil)
			So(err, ShouldBeNil)
			req.AddCookie(&http.Cookie{Name: "access_token", Value: "12345"})

			router := mux.NewRouter()
			router.Path("/{uri:.*}").HandlerFunc(LegacyLanding(mockZebedeeClient, mockDatasetClient, mockFilesAPIClient, nil, mockConfig))

			router.ServeHTTP(w, req)

			So(w.Body.String(), ShouldEqual, `{"some_json":true}`)
		})

		Convey("test status 500 returned if zedbedee get returns error", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockFilesAPIClient := NewMockFilesAPIClient(mockCtrl)
			mockZebedeeClient.EXPECT().Get(ctx, userAuthToken, "/data?uri=/data").Return(nil, errors.New("something went wrong with zebedee"))
			mockConfig := config.Config{}

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/data", nil)
			So(err, ShouldBeNil)
			router := mux.NewRouter()
			router.Path("/{uri:.*}").HandlerFunc(LegacyLanding(mockZebedeeClient, mockDatasetClient, mockFilesAPIClient, nil, mockConfig))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})

	Convey("test legacylanding handler with non /data endpoint", t, func() {
		Convey("test successful data retrieval and rendering", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockFilesAPIClient := NewMockFilesAPIClient(mockCtrl)
			mockConfig := config.Config{}
			dlp := zebedee.DatasetLandingPage{URI: "https://helloworld.com"}
			dlp.Datasets = append(dlp.Datasets, zebedee.Related{Title: "A dataset!", URI: "dataset.com"})

			mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, userAuthToken, collectionID, locale, "/somelegacypage").Return(dlp, nil)
			mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthToken, collectionID, locale, dlp.URI)
			mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthToken, collectionID, locale, "dataset.com")
			mockZebedeeClient.EXPECT().GetHomepageContent(ctx, userAuthToken, collectionID, locale, "/")

			mockRend := NewMockRenderClient(mockCtrl)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "static")

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/somelegacypage", nil)
			So(err, ShouldBeNil)

			router := mux.NewRouter()
			router.Path("/{uri:.*}").HandlerFunc(LegacyLanding(mockZebedeeClient, mockDatasetClient, mockFilesAPIClient, mockRend, mockConfig))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("test status 500 returned when zebedee client returns error retrieving landing page", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockFilesAPIClient := NewMockFilesAPIClient(mockCtrl)
			mockConfig := config.Config{}
			dlp := zebedee.DatasetLandingPage{}
			mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, userAuthToken, collectionID, locale, "/somelegacypage").Return(dlp, errors.New("something went wrong :("))

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/somelegacypage", nil)
			So(err, ShouldBeNil)

			router := mux.NewRouter()
			router.Path("/{uri:.*}").HandlerFunc(LegacyLanding(mockZebedeeClient, mockDatasetClient, mockFilesAPIClient, nil, mockConfig))

			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("test status 500 returned when zebedee client returns error retrieving breadcrumb", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockFilesAPIClient := NewMockFilesAPIClient(mockCtrl)
			mockConfig := config.Config{}
			dlp := zebedee.DatasetLandingPage{URI: "https://helloworld.com"}
			mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, userAuthToken, collectionID, locale, "/somelegacypage").Return(dlp, nil)
			mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthToken, collectionID, locale, dlp.URI).Return(nil, errors.New("something went wrong"))

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/somelegacypage", nil)
			So(err, ShouldBeNil)

			router := mux.NewRouter()
			router.Path("/{uri:.*}").HandlerFunc(LegacyLanding(mockZebedeeClient, mockDatasetClient, mockFilesAPIClient, nil, mockConfig))

			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
