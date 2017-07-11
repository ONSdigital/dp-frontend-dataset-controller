package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/go-ns/zebedee/data"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func createMockClient(expectedResponse []byte, expectedCode int) *http.Client {
	mockStreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(expectedCode)
		w.Write(expectedResponse)
	}))
	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(mockStreamServer.URL)
		},
	}
	return &http.Client{Transport: transport}
}

func TestUnitHandlers(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	Convey("test /data endpoint", t, func() {

		Convey("test successful json response", func() {
			mockClient := NewMockZebedeeClient(mockCtrl)
			mockClient.EXPECT().Get("/data?uri=/data").Return([]byte(`{"some_json":true}`), nil)
			mockClient.EXPECT().SetAccessToken("12345")

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/data", nil)
			So(err, ShouldBeNil)
			cfg := config.Get()
			req.AddCookie(&http.Cookie{Name: "access_token", Value: "12345"})

			legacyLanding(w, req, mockClient, cfg)

			So(w.Body.String(), ShouldEqual, `{"some_json":true}`)
		})

		Convey("test status 500 returned if zedbedee get returns error", func() {
			mockClient := NewMockZebedeeClient(mockCtrl)
			mockClient.EXPECT().Get("/data?uri=/data").Return(nil, errors.New("something went wrong with zebedee"))

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/data", nil)
			So(err, ShouldBeNil)
			cfg := config.Get()

			legacyLanding(w, req, mockClient, cfg)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

	})

	Convey("test legacylanding handler with non /data endpoint", t, func() {
		Convey("test sucessful data retrieval and rendering", func() {
			mockClient := NewMockZebedeeClient(mockCtrl)
			dlp := data.DatasetLandingPage{URI: "http://helloworld.com"}
			dlp.Datasets = append(dlp.Datasets, data.Related{Title: "A dataset!", URI: "dataset.com"})

			mockClient.EXPECT().GetDatasetLandingPage("/data?uri=/somelegacypage").Return(dlp, nil)
			mockClient.EXPECT().GetBreadcrumb(dlp.URI)
			mockClient.EXPECT().GetDataset("dataset.com")

			cli = createMockClient([]byte(`<html><body><h1>Some HTML from renderer!</h1></body></html>`), 200)

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/somelegacypage", nil)
			So(err, ShouldBeNil)
			cfg := config.Get()

			legacyLanding(w, req, mockClient, cfg)

			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, `<html><body><h1>Some HTML from renderer!</h1></body></html>`)
		})

		Convey("test status 500 returned when zebedee client returns error retrieving landing page", func() {
			mockClient := NewMockZebedeeClient(mockCtrl)
			dlp := data.DatasetLandingPage{}
			mockClient.EXPECT().GetDatasetLandingPage("/data?uri=/somelegacypage").Return(dlp, errors.New("something went wrong :("))

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/somelegacypage", nil)
			So(err, ShouldBeNil)
			cfg := config.Get()

			legacyLanding(w, req, mockClient, cfg)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("test status 500 returned when zebedee client returns error retrieving breadcrumb", func() {
			mockClient := NewMockZebedeeClient(mockCtrl)
			dlp := data.DatasetLandingPage{URI: "http://helloworld.com"}
			mockClient.EXPECT().GetDatasetLandingPage("/data?uri=/somelegacypage").Return(dlp, nil)
			mockClient.EXPECT().GetBreadcrumb(dlp.URI).Return(nil, errors.New("something went wrong"))

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/somelegacypage", nil)
			So(err, ShouldBeNil)
			cfg := config.Get()

			legacyLanding(w, req, mockClient, cfg)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("test status 500 returned if render client returns error", func() {
			mockClient := NewMockZebedeeClient(mockCtrl)
			dlp := data.DatasetLandingPage{URI: "http://helloworld.com"}
			dlp.Datasets = append(dlp.Datasets, data.Related{Title: "A dataset!", URI: "dataset.com"})

			mockClient.EXPECT().GetDatasetLandingPage("/data?uri=/somelegacypage").Return(dlp, nil)
			mockClient.EXPECT().GetBreadcrumb(dlp.URI)
			mockClient.EXPECT().GetDataset("dataset.com")

			cli = createMockClient(nil, 500)

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/somelegacypage", nil)
			So(err, ShouldBeNil)
			cfg := config.Get()

			legacyLanding(w, req, mockClient, cfg)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})

}
