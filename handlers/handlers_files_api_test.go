package handlers

import (
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	coreModel "github.com/ONSdigital/dp-renderer/model"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"net/http"
	"net/http/httptest"
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

			mockConfig := config.Config{}

			landingPageURI := "https://helloworld.com"
			dataSetURI := "dataset.com"
			legacyURL := "/some_legacy_page"
			downloadFileSize := "100"
			downloadURI := "file_from_zebedee"

			dlp := zebedee.DatasetLandingPage{
				URI:      landingPageURI,
				Datasets: []zebedee.Link{{Title: "Dataset", URI: dataSetURI}},
			}

			zebedeeDataset := zebedee.Dataset{Downloads: []zebedee.Download{{File: downloadURI, Size: downloadFileSize}}}

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

			handler := LegacyLanding(mockZebedeeClient, mockDatasetClient, mockRend, mockConfig)
			handler(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
			So(actualPageModel.DatasetLandingPage.Datasets[0].Downloads[0].Size, ShouldEqual, downloadFileSize)
		})
	})
}
