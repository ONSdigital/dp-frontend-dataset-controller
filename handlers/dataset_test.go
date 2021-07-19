package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func DatasetTestUnitHandlers(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()
	Convey("test datasetPage handler with non /data endpoint", t, func() {
		Convey("test successful data retrieval and rendering", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockConfig := config.Config{}
			dlp := zebedee.DatasetLandingPage{URI: "http://dataset.com"}
			dlp.Datasets = append(dlp.Datasets, zebedee.Related{Title: "A dataset!", URI: "dataset.com/datasetpage"})
			dp := zebedee.Dataset{URI: "http://dataset.com/datasetpage"}
			bc := []zebedee.Breadcrumb{
				{
					URI: "/datasets",
				},
				{
					URI: "/datasets/somedatasetpage",
				},
			}

			mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, userAuthToken, collectionID, locale, "/datasets/somedatasetpage").Return(dlp, nil)
			mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthToken, collectionID, locale, dp.URI).Return(bc, nil)
			mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthToken, collectionID, locale, "/datasets/somedatasetpage/current").Return(dp, nil)

			mockRend := NewMockRenderClient(mockCtrl)
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "dataset-page").Return([]byte(`<html><body><h1>Some HTML from renderer!</h1></body></html>`), nil)

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/datasets/somedatasetpage/current", nil)
			So(err, ShouldBeNil)

			router := mux.NewRouter()
			router.Path("/{uri:.*}").HandlerFunc(DatasetPage(mockZebedeeClient, mockDatasetClient, mockRend, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, `<html><body><h1>Some HTML from renderer!</h1></body></html>`)
		})

		Convey("test status 500 returned when zebedee client returns error retrieving dataset page", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockConfig := config.Config{}
			dp := zebedee.Dataset{}
			mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthToken, collectionID, locale, "/datasets/somedatasetpage/current").Return(dp, errors.New("something went wrong :("))

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/datasets/somedatasetpage/current", nil)
			So(err, ShouldBeNil)

			router := mux.NewRouter()
			router.Path("/{uri:.*}").HandlerFunc(DatasetPage(mockZebedeeClient, mockDatasetClient, nil, mockConfig, "/v1"))

			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("test status 500 returned when zebedee client returns error retrieving breadcrumb", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockConfig := config.Config{}
			dp := zebedee.Dataset{URI: "http://helloworld.com"}
			mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthToken, collectionID, locale, "/datasets/somedatasetpage/current").Return(dp, nil)
			mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthToken, collectionID, locale, dp.URI).Return(nil, errors.New("something went wrong"))

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/datasets/somedatasetpage/current", nil)
			So(err, ShouldBeNil)

			router := mux.NewRouter()
			router.Path("/{uri:.*}").HandlerFunc(DatasetPage(mockZebedeeClient, mockDatasetClient, nil, mockConfig, "/v1"))

			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("test status 500 returned when zebedee client returns error retrieving parent dataset landing page", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockConfig := config.Config{}
			dp := zebedee.Dataset{URI: "http://helloworld.com"}
			dlp := zebedee.DatasetLandingPage{}
			bc := []zebedee.Breadcrumb{
				{
					URI: "/datasets",
				},
				{
					URI: "/datasets/somedatasetpage",
				},
			}

			mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, userAuthToken, collectionID, locale, "/datasets/somedatasetpage").Return(dlp, errors.New("something went wrong :("))
			mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthToken, collectionID, locale, dp.URI).Return(bc, nil)
			mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthToken, collectionID, locale, "/datasets/somedatasetpage/current").Return(dp, nil)

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/datasets/somedatasetpage/current", nil)
			So(err, ShouldBeNil)

			router := mux.NewRouter()
			router.Path("/{uri:.*}").HandlerFunc(DatasetPage(mockZebedeeClient, mockDatasetClient, nil, mockConfig, "/v1"))

			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("test status 500 returned if render client returns error", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockConfig := config.Config{}
			dlp := zebedee.DatasetLandingPage{URI: "http://dataset.com"}
			dlp.Datasets = append(dlp.Datasets, zebedee.Related{Title: "A dataset!", URI: "dataset.com/datasetpage"})
			dp := zebedee.Dataset{URI: "http://dataset.com/datasetpage"}
			bc := []zebedee.Breadcrumb{
				{
					URI: "/datasets",
				},
				{
					URI: "/datasets/somedatasetpage",
				},
			}

			mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, userAuthToken, collectionID, locale, "/datasets/somedatasetpage").Return(dlp, nil)
			mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthToken, collectionID, locale, dp.URI).Return(bc, nil)
			mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthToken, collectionID, locale, "/datasets/somedatasetpage/current").Return(dp, nil)

			mockRend := NewMockRenderClient(mockCtrl)
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "dataset-page").Return(nil, errors.New("error from renderer"))

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/datasets/somedatasetpage/current", nil)
			So(err, ShouldBeNil)

			router := mux.NewRouter()
			router.Path("/{uri:.*}").HandlerFunc(DatasetPage(mockZebedeeClient, mockDatasetClient, mockRend, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})

}
