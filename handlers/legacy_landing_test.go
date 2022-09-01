package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/files"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/cache"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	coreModel "github.com/ONSdigital/dp-renderer/model"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLegacyLanding(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()
	cfg := initialiseMockConfig()

	mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
	mockDatasetClient := NewMockDatasetClient(mockCtrl)
	mockFilesAPIClient := NewMockFilesAPIClient(mockCtrl)

	Convey("test /path/to/something/data endpoint", t, func() {
		path := "/path/to/something"
		zebedeePath := "/data?uri=" + path

		Convey("test successful json response", func() {
			mockZebedeeClient.EXPECT().Get(ctx, "12345", zebedeePath).Return([]byte(`{"some_json":true}`), nil)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, path+"/data", nil)
			req.AddCookie(&http.Cookie{Name: "access_token", Value: "12345"})

			ctxOther := context.Background()
			mockCacheList, err := cache.GetMockCacheList(ctxOther, cfg.SupportedLanguages)
			So(err, ShouldBeNil)

			LegacyLanding(mockZebedeeClient, mockDatasetClient, mockFilesAPIClient, nil, mockCacheList)(w, req)

			So(w.Body.String(), ShouldEqual, `{"some_json":true}`)
		})
	})

	Convey("test legacy landing handler with non /data endpoint", t, func() {
		Convey("test successful data retrieval and rendering", func() {
			dlp := zebedee.DatasetLandingPage{URI: "https://helloworld.com"}
			dlp.Datasets = append(dlp.Datasets, zebedee.Link{Title: "A dataset!", URI: "dataset.com"})

			mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, userAuthToken, collectionID, locale, "/somelegacypage").Return(dlp, nil)
			mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthToken, collectionID, locale, dlp.URI)
			mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthToken, collectionID, locale, "dataset.com")
			mockZebedeeClient.EXPECT().GetHomepageContent(ctx, userAuthToken, collectionID, locale, "/")

			mockRend := NewMockRenderClient(mockCtrl)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "static")

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/somelegacypage", nil)

			ctxOther := context.Background()
			mockCacheList, err := cache.GetMockCacheList(ctxOther, cfg.SupportedLanguages)
			So(err, ShouldBeNil)

			LegacyLanding(mockZebedeeClient, mockDatasetClient, mockFilesAPIClient, mockRend, mockCacheList).ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("test status 500 returned when zebedee client returns error retrieving landing page", func() {
			dlp := zebedee.DatasetLandingPage{}
			mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, userAuthToken, collectionID, locale, "/somelegacypage").Return(dlp, errors.New("something went wrong :("))

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/somelegacypage", nil)

			ctxOther := context.Background()
			mockCacheList, err := cache.GetMockCacheList(ctxOther, cfg.SupportedLanguages)
			So(err, ShouldBeNil)

			LegacyLanding(mockZebedeeClient, mockDatasetClient, mockFilesAPIClient, nil, mockCacheList).ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("test status 500 returned when zebedee client returns error retrieving breadcrumb", func() {
			dlp := zebedee.DatasetLandingPage{URI: "https://helloworld.com"}
			mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, userAuthToken, collectionID, locale, "/somelegacypage").Return(dlp, nil)
			mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthToken, collectionID, locale, dlp.URI).Return(nil, errors.New("something went wrong"))

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/somelegacypage", nil)
			So(err, ShouldBeNil)

			ctxOther := context.Background()
			mockCacheList, err := cache.GetMockCacheList(ctxOther, cfg.SupportedLanguages)
			So(err, ShouldBeNil)

			LegacyLanding(mockZebedeeClient, mockDatasetClient, mockFilesAPIClient, nil, mockCacheList).ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func setupMockClients(ctx gomock.Matcher, mockZebedeeClient *MockZebedeeClient, mockRend *MockRenderClient, legacyURL string, dlp zebedee.DatasetLandingPage, authToken string, cfg config.Config) {
	mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, authToken, collectionID, locale, legacyURL).Return(dlp, nil)
	mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, authToken, collectionID, locale, dlp.URI)
	mockZebedeeClient.EXPECT().GetHomepageContent(ctx, authToken, collectionID, locale, "/")

	mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
}

func TestHandlersFilesAPI(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()
	cfg := initialiseMockConfig()

	mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
	mockDatasetClient := NewMockDatasetClient(mockCtrl)
	mockRend := NewMockRenderClient(mockCtrl)
	mockFilesAPIClient := NewMockFilesAPIClient(mockCtrl)

	landingPageURI := "https://helloworld.com"
	dataSetURI := "dataset.com"
	legacyURL := "/some_legacy_page"
	expectedAuthToken := "auth-token"
	authHeaderKey := "X-Florence-Token"

	dlp := zebedee.DatasetLandingPage{
		URI:      landingPageURI,
		Datasets: []zebedee.Link{{Title: "Dataset", URI: dataSetURI}},
	}

	Convey("LegacyLanding handler file storage", t, func() {
		Convey("File stored in Zebedee", func() {
			expectedDownloadFileSize := "100"
			expectedSupplementaryFileSize := "101"
			downloadURI := "download_file_from_zebedee"
			supplementaryURI := "supplementary_file_from_zebedee"

			zebedeeDataset := zebedee.Dataset{
				Downloads:          []zebedee.Download{{File: downloadURI, Size: expectedDownloadFileSize}},
				SupplementaryFiles: []zebedee.SupplementaryFile{{File: supplementaryURI, Size: expectedSupplementaryFileSize}},
			}

			setupMockClients(ctx, mockZebedeeClient, mockRend, legacyURL, dlp, expectedAuthToken, cfg)

			mockZebedeeClient.EXPECT().GetDataset(ctx, expectedAuthToken, collectionID, locale, dataSetURI).Return(zebedeeDataset, nil)

			var actualPageModel mapper.StaticDatasetLandingPage

			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "static").Do(func(w io.Writer, pageModel interface{}, templateName string) {
				actualPageModel = pageModel.(mapper.StaticDatasetLandingPage)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", legacyURL, nil)
			req.Header.Set(authHeaderKey, expectedAuthToken)

			ctxOther := context.Background()
			mockCacheList, err := cache.GetMockCacheList(ctxOther, cfg.SupportedLanguages)
			So(err, ShouldBeNil)
			handler := LegacyLanding(mockZebedeeClient, mockDatasetClient, mockFilesAPIClient, mockRend, mockCacheList)
			handler(w, req)

			actualDownloadSize := actualPageModel.DatasetLandingPage.Datasets[0].Downloads[0].Size
			actualSupplementaryFileSize := actualPageModel.DatasetLandingPage.Datasets[0].SupplementaryFiles[0].Size

			So(w.Code, ShouldEqual, http.StatusOK)
			So(actualDownloadSize, ShouldEqual, expectedDownloadFileSize)
			So(actualSupplementaryFileSize, ShouldEqual, expectedSupplementaryFileSize)
		})

		Convey("Files stored in Files API", func() {
			expectedDownloadFileSize := "100"
			expectedSupplementaryFileSize := "101"
			downloadURI := "download_file_from_zebedee"
			supplementaryURI := "supplementary_file_from_zebedee"
			expectedDownloadFileSizeInt, _ := strconv.Atoi(expectedDownloadFileSize)
			expectedSupplementaryFileSizeInt, _ := strconv.Atoi(expectedSupplementaryFileSize)
			actualVersion := "v2"
			zebedeeDownloadFileSize := ""
			zebedeeSupplementaryFileSize := ""

			fmdd := files.FileMetaData{SizeInBytes: uint64(expectedDownloadFileSizeInt)}
			fmds := files.FileMetaData{SizeInBytes: uint64(expectedSupplementaryFileSizeInt)}

			zebedeeDataset := zebedee.Dataset{
				Downloads:          []zebedee.Download{{URI: downloadURI, Size: zebedeeDownloadFileSize, Version: actualVersion}},
				SupplementaryFiles: []zebedee.SupplementaryFile{{URI: supplementaryURI, Size: zebedeeSupplementaryFileSize}},
			}

			setupMockClients(ctx, mockZebedeeClient, mockRend, legacyURL, dlp, expectedAuthToken, cfg)

			mockZebedeeClient.EXPECT().GetDataset(ctx, expectedAuthToken, collectionID, locale, dataSetURI).Return(zebedeeDataset, nil)
			mockFilesAPIClient.EXPECT().GetFile(ctx, gomock.Any(), expectedAuthToken).Return(fmdd, nil)
			mockFilesAPIClient.EXPECT().GetFile(ctx, gomock.Any(), expectedAuthToken).Return(fmds, nil)

			var actualPageModel mapper.StaticDatasetLandingPage

			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "static").Do(func(w io.Writer, pageModel interface{}, templateName string) {
				actualPageModel = pageModel.(mapper.StaticDatasetLandingPage)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", legacyURL, nil)
			req.Header.Set(authHeaderKey, expectedAuthToken)

			ctxOther := context.Background()
			mockCacheList, err := cache.GetMockCacheList(ctxOther, cfg.SupportedLanguages)
			So(err, ShouldBeNil)
			handler := LegacyLanding(mockZebedeeClient, mockDatasetClient, mockFilesAPIClient, mockRend, mockCacheList)
			handler(w, req)

			actualDownloadFileSize := actualPageModel.DatasetLandingPage.Datasets[0].Downloads[0].Size
			actualSupplementaryFileSize := actualPageModel.DatasetLandingPage.Datasets[0].SupplementaryFiles[0].Size

			So(w.Code, ShouldEqual, http.StatusOK)
			So(actualDownloadFileSize, ShouldEqual, expectedDownloadFileSize)
			So(actualSupplementaryFileSize, ShouldEqual, expectedSupplementaryFileSize)
		})
	})
}
