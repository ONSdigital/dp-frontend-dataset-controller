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

	mockConfig := config.Config{}

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

			handler := LegacyLanding(mockZebedeeClient, mockDatasetClient, mockFilesAPIClient, mockRend, mockConfig, "v1")
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

			handler := LegacyLanding(mockZebedeeClient, mockDatasetClient, mockFilesAPIClient, mockRend, mockConfig, "v1")
			handler(w, req)

			actualDownloadFileSize := actualPageModel.DatasetLandingPage.Datasets[0].Downloads[0].Size
			actualSupplementaryFileSize := actualPageModel.DatasetLandingPage.Datasets[0].SupplementaryFiles[0].Size

			So(w.Code, ShouldEqual, http.StatusOK)
			So(actualDownloadFileSize, ShouldEqual, expectedDownloadFileSize)
			So(actualSupplementaryFileSize, ShouldEqual, expectedSupplementaryFileSize)
		})
	})
}
