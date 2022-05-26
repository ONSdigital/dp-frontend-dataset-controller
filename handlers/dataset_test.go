package handlers

import (
	"errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/files"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	coreModel "github.com/ONSdigital/dp-renderer/model"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	datasetLandingPageURI = "/peoplepopulationandcommunity/birthsdeathsandmarriages/deaths/datasets/weeklyprovisionalfiguresondeathsregisteredinenglandandwales"
	editionPageURI        = datasetLandingPageURI + "/2022"
)

func TestDatasetHandlers(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()

	Convey("test datasetPage handler with non /data endpoint", t, func() {

		expectedDownloadFilename := "download_filename.csv"
		expectedVersionURI := editionPageURI + "/previous/v1"

		mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
		mockFilesAPIClient := NewMockFilesAPIClient(mockCtrl)
		mockRend := NewMockRenderClient(mockCtrl)
		cfg := initialiseMockConfig()
		hp := zebedee.HomepageContent{}
		bc := []zebedee.Breadcrumb{{URI: editionPageURI}, {URI: datasetLandingPageURI}}

		Convey("test successful data retrieval and rendering when files exist in Zebedee", func() {
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))

			expectedDownloadFileSize := "100"

			dlp := zebedee.DatasetLandingPage{
				URI:      datasetLandingPageURI,
				Datasets: []zebedee.Link{{URI: editionPageURI}},
			}
			editionDataSet := zebedee.Dataset{
				URI:      editionPageURI,
				Versions: []zebedee.Version{{URI: expectedVersionURI}},
			}
			versionDataSet := zebedee.Dataset{
				URI:       expectedVersionURI,
				Downloads: []zebedee.Download{{File: expectedDownloadFilename, Size: expectedDownloadFileSize}},
			}

			mockZebedeeClient.EXPECT().GetHomepageContent(ctx, userAuthToken, collectionID, locale, homepagePath).Return(hp, nil)
			mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, userAuthToken, collectionID, locale, datasetLandingPageURI).Return(dlp, nil)
			mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthToken, collectionID, locale, editionDataSet.URI).Return(bc, nil)
			mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthToken, collectionID, locale, editionPageURI).Return(editionDataSet, nil)
			mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthToken, collectionID, locale, expectedVersionURI).Return(versionDataSet, nil)

			var actualPageModel mapper.DatasetPage

			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "dataset").Do(func(w io.Writer, pageModel interface{}, templateName string) {
				actualPageModel = pageModel.(mapper.DatasetPage)
			})

			w, req := generateRecorderAndRequest()
			DatasetPage(mockZebedeeClient, mockRend, mockFilesAPIClient)(w, req)

			actualDownloadSize := actualPageModel.DatasetPage.Versions[0].Downloads[0].Size

			So(w.Code, ShouldEqual, http.StatusOK)
			So(actualDownloadSize, ShouldEqual, expectedDownloadFileSize)
		})

		Convey("test successful data retrieval and rendering Files stored in Files API", func() {
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))

			expectedDownloadFileSize := "100"
			expectedDownloadFileSizeInt, _ := strconv.Atoi(expectedDownloadFileSize)

			dlp := zebedee.DatasetLandingPage{
				URI:      datasetLandingPageURI,
				Datasets: []zebedee.Link{{URI: editionPageURI}},
			}
			editionDataSet := zebedee.Dataset{
				URI:      editionPageURI,
				Versions: []zebedee.Version{{URI: expectedVersionURI}},
			}
			versionDataSet := zebedee.Dataset{
				URI:       expectedVersionURI,
				Downloads: []zebedee.Download{{URI: expectedDownloadFilename}},
			}

			fmd := files.FileMetaData{SizeInBytes: uint64(expectedDownloadFileSizeInt)}

			mockZebedeeClient.EXPECT().GetHomepageContent(ctx, userAuthToken, collectionID, locale, homepagePath).Return(hp, nil)
			mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, userAuthToken, collectionID, locale, datasetLandingPageURI).Return(dlp, nil)
			mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthToken, collectionID, locale, editionDataSet.URI).Return(bc, nil)
			mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthToken, collectionID, locale, editionPageURI).Return(editionDataSet, nil)
			mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthToken, collectionID, locale, expectedVersionURI).Return(versionDataSet, nil)
			mockFilesAPIClient.EXPECT().GetFile(ctx, expectedDownloadFilename, userAuthToken).Return(fmd, nil)

			var actualPageModel mapper.DatasetPage

			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "dataset").Do(func(w io.Writer, pageModel interface{}, templateName string) {
				actualPageModel = pageModel.(mapper.DatasetPage)
			})

			w, req := generateRecorderAndRequest()
			DatasetPage(mockZebedeeClient, mockRend, mockFilesAPIClient)(w, req)

			actualDownloadSize := actualPageModel.DatasetPage.Versions[0].Downloads[0].Size

			So(w.Code, ShouldEqual, http.StatusOK)
			So(actualDownloadSize, ShouldEqual, expectedDownloadFileSize)
		})

		Convey("test status 500 returned when zebedee client returns error retrieving dataset page", func() {
			mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthToken, collectionID, locale, gomock.Any()).Return(zebedee.Dataset{}, errors.New("something went wrong :("))

			w, req := generateRecorderAndRequest()
			DatasetPage(mockZebedeeClient, mockRend, mockFilesAPIClient)(w, req)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("test status 500 returned when zebedee client returns error retrieving breadcrumb", func() {
			mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthToken, collectionID, locale, gomock.Any()).Return(zebedee.Dataset{}, nil)
			mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthToken, collectionID, locale, gomock.Any()).Return(nil, errors.New("something went wrong"))

			w, req := generateRecorderAndRequest()
			DatasetPage(mockZebedeeClient, mockRend, mockFilesAPIClient)(w, req)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("test status 500 returned when zebedee client returns error retrieving parent dataset landing page", func() {
			mockZebedeeClient.EXPECT().GetHomepageContent(ctx, userAuthToken, collectionID, locale, gomock.Any()).Return(hp, nil)
			mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, userAuthToken, collectionID, locale, gomock.Any()).Return(zebedee.DatasetLandingPage{}, errors.New("something went wrong :("))
			mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthToken, collectionID, locale, gomock.Any()).Return(bc, nil)
			mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthToken, collectionID, locale, gomock.Any()).Return(zebedee.Dataset{}, nil)

			w, req := generateRecorderAndRequest()
			DatasetPage(mockZebedeeClient, mockRend, mockFilesAPIClient)(w, req)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func generateRecorderAndRequest() (*httptest.ResponseRecorder, *http.Request) {
	req, _ := http.NewRequest(http.MethodGet, editionPageURI, nil)
	return httptest.NewRecorder(), req
}
