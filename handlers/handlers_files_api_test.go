package handlers

import (
	"github.com/ONSdigital/dp-api-clients-go/v2/files"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	coreModel "github.com/ONSdigital/dp-renderer/model"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestHandlersFilesAPI(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()
	cfg := initialiseMockConfig()

	Convey("LegacyLanding handler file storage", t, func() {
		Convey("File stored in Zebedee", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockRend := NewMockRenderClient(mockCtrl)
			mockFilesAPIClient := NewMockFilesAPIClient(mockCtrl)

			mockConfig := config.Config{}

			landingPageURI := "https://helloworld.com"
			dataSetURI := "dataset.com"
			legacyURL := "/some_legacy_page"
			expectedDownloadFileSize := "100"
			expectedSupplementaryFileSize := "101"
			downloadURI := "download_file_from_zebedee"
			supplementaryURI := "supplementary_file_from_zebedee"

			dlp := zebedee.DatasetLandingPage{
				URI:      landingPageURI,
				Datasets: []zebedee.Link{{Title: "Dataset", URI: dataSetURI}},
			}

			zebedeeDataset := zebedee.Dataset{
				Downloads:          []zebedee.Download{{File: downloadURI, Size: expectedDownloadFileSize}},
				SupplementaryFiles: []zebedee.SupplementaryFile{{File: supplementaryURI, Size: expectedSupplementaryFileSize}},
			}

			mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, userAuthToken, collectionID, locale, legacyURL).Return(dlp, nil)
			mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthToken, collectionID, locale, dlp.URI)
			mockZebedeeClient.EXPECT().GetHomepageContent(ctx, userAuthToken, collectionID, locale, "/")
			mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthToken, collectionID, locale, dataSetURI).Return(zebedeeDataset, nil)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))

			var actualPageModel mapper.StaticDatasetLandingPage

			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "static").Do(func(w io.Writer, pageModel interface{}, templateName string) {
				actualPageModel = pageModel.(mapper.StaticDatasetLandingPage)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", legacyURL, nil)

			handler := LegacyLanding(mockZebedeeClient, mockDatasetClient, mockFilesAPIClient, mockRend, mockConfig)
			handler(w, req)

			actualDownloadSize := actualPageModel.DatasetLandingPage.Datasets[0].Downloads[0].Size
			actualSupplementaryFileSize := actualPageModel.DatasetLandingPage.Datasets[0].SupplementaryFiles[0].Size

			So(w.Code, ShouldEqual, http.StatusOK)
			So(actualDownloadSize, ShouldEqual, expectedDownloadFileSize)
			So(actualSupplementaryFileSize, ShouldEqual, expectedSupplementaryFileSize)
		})

		Convey("Files stored in Files API", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockRend := NewMockRenderClient(mockCtrl)
			mockFilesAPIClient := NewMockFilesAPIClient(mockCtrl)

			mockConfig := config.Config{}

			landingPageURI := "https://helloworld.com"
			dataSetURI := "dataset.com"
			legacyURL := "/some_legacy_page"
			expectedDownloadFileSize := "100"
			expectedSupplementaryFileSize := "101"
			downloadURI := "download_file_from_zebedee"
			supplementaryURI := "supplementary_file_from_zebedee"
			expectedDownloadFileSizeInt, _ := strconv.Atoi(expectedDownloadFileSize)
			expectedSupplementaryFileSizeInt, _ := strconv.Atoi(expectedSupplementaryFileSize)
			actualVersion := "v2"
			zebedeeDownloadFileSize := ""
			zebedeeSupplementaryFileSize := ""
			expectedAuthToken := "auth-token"
			authHeaderKey := "X-Florence-Token"

			dlp := zebedee.DatasetLandingPage{
				URI:      landingPageURI,
				Datasets: []zebedee.Link{{Title: "Dataset", URI: dataSetURI}},
			}

			fmdd := files.FileMetaData{SizeInBytes: uint64(expectedDownloadFileSizeInt)}
			fmds := files.FileMetaData{SizeInBytes: uint64(expectedSupplementaryFileSizeInt)}

			zebedeeDataset := zebedee.Dataset{
				Downloads:          []zebedee.Download{{URI: downloadURI, Size: zebedeeDownloadFileSize, Version: actualVersion}},
				SupplementaryFiles: []zebedee.SupplementaryFile{{URI: supplementaryURI, Size: zebedeeSupplementaryFileSize}},
			}

			mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, expectedAuthToken, collectionID, locale, legacyURL).Return(dlp, nil)
			mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, expectedAuthToken, collectionID, locale, dlp.URI)
			mockZebedeeClient.EXPECT().GetHomepageContent(ctx, expectedAuthToken, collectionID, locale, "/")
			mockZebedeeClient.EXPECT().GetDataset(ctx, expectedAuthToken, collectionID, locale, dataSetURI).Return(zebedeeDataset, nil)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockFilesAPIClient.EXPECT().GetFile(ctx, gomock.Any(), expectedAuthToken).Return(fmdd, nil)
			mockFilesAPIClient.EXPECT().GetFile(ctx, gomock.Any(), expectedAuthToken).Return(fmds, nil)

			var actualPageModel mapper.StaticDatasetLandingPage

			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "static").Do(func(w io.Writer, pageModel interface{}, templateName string) {
				actualPageModel = pageModel.(mapper.StaticDatasetLandingPage)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", legacyURL, nil)
			req.Header.Set(authHeaderKey, expectedAuthToken)

			handler := LegacyLanding(mockZebedeeClient, mockDatasetClient, mockFilesAPIClient, mockRend, mockConfig)
			handler(w, req)

			actualDownloadFileSize := actualPageModel.DatasetLandingPage.Datasets[0].Downloads[0].Size
			actualSupplementaryFileSize := actualPageModel.DatasetLandingPage.Datasets[0].SupplementaryFiles[0].Size

			So(w.Code, ShouldEqual, http.StatusOK)
			So(actualDownloadFileSize, ShouldEqual, expectedDownloadFileSize)
			So(actualSupplementaryFileSize, ShouldEqual, expectedSupplementaryFileSize)
		})
	})
}
