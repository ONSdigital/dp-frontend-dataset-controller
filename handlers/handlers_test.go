package handlers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	dataset "github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/zebedee/data"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
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

	Convey("test CreateFilterID", t, func() {
		Convey("test CreateFilterID handler, creates a filter id and redirects", func() {
			mockClient := NewMockFilterClient(mockCtrl)
			mockClient.EXPECT().CreateBlueprint("87654321", []string{"aggregate", "time"}).Return("12345", nil)

			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			dims := dataset.Dimensions{
				Items: []dataset.Dimension{
					{
						ID: "aggregate",
					},
					{
						ID: "time",
					},
				},
			}
			opts := dataset.Options{
				Items: []dataset.Option{
					{
						Label: "1",
					},
					{
						Label: "2",
					},
				},
			}
			mockDatasetClient.EXPECT().GetDimensions("1234", "5678", "2017").Return(dims, nil)
			mockDatasetClient.EXPECT().GetOptions("1234", "5678", "2017", "aggregate").Return(opts, nil)
			mockDatasetClient.EXPECT().GetOptions("1234", "5678", "2017", "time").Return(opts, nil)
			mockDatasetClient.EXPECT().GetVersion("1234", "5678", "2017").Return(dataset.Version{ID: "87654321"}, nil)

			w := testResponse(301, "", "/datasets/1234/editions/5678/versions/2017/filter", mockClient, mockDatasetClient, CreateFilterID(mockClient, mockDatasetClient))

			location := w.Header().Get("Location")
			So(location, ShouldNotBeEmpty)

			So(location, ShouldEqual, "/filters/12345/dimensions")
		})

		Convey("test CreateFilterID returns 500 if unable to create filter job on filter api", func() {
			mockClient := NewMockFilterClient(mockCtrl)
			mockClient.EXPECT().CreateBlueprint(gomock.Any(), gomock.Any()).Return("", errors.New("no filter job for you"))

			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockDatasetClient.EXPECT().GetVersion("1234", "5678", "2017")
			mockDatasetClient.EXPECT().GetDimensions("1234", "5678", "2017").Return(dataset.Dimensions{}, nil)

			testResponse(500, "", "/datasets/1234/editions/5678/versions/2017/filter", mockClient, mockDatasetClient, CreateFilterID(mockClient, mockDatasetClient))
		})
	})

	Convey("test /data endpoint", t, func() {

		Convey("test successful json response", func() {
			mockClient := NewMockZebedeeClient(mockCtrl)
			mockClient.EXPECT().Get("/data?uri=/data").Return([]byte(`{"some_json":true}`), nil)
			mockClient.EXPECT().SetAccessToken("12345")

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/data", nil)
			So(err, ShouldBeNil)
			req.AddCookie(&http.Cookie{Name: "access_token", Value: "12345"})

			router := mux.NewRouter()
			router.Path("/{uri:.*}").HandlerFunc(LegacyLanding(mockClient))

			router.ServeHTTP(w, req)

			So(w.Body.String(), ShouldEqual, `{"some_json":true}`)
		})

		Convey("test status 500 returned if zedbedee get returns error", func() {
			mockClient := NewMockZebedeeClient(mockCtrl)
			mockClient.EXPECT().Get("/data?uri=/data").Return(nil, errors.New("something went wrong with zebedee"))

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/data", nil)
			So(err, ShouldBeNil)
			router := mux.NewRouter()
			router.Path("/{uri:.*}").HandlerFunc(LegacyLanding(mockClient))

			router.ServeHTTP(w, req)

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

			router := mux.NewRouter()
			router.Path("/{uri:.*}").HandlerFunc(LegacyLanding(mockClient))

			router.ServeHTTP(w, req)

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

			router := mux.NewRouter()
			router.Path("/{uri:.*}").HandlerFunc(LegacyLanding(mockClient))

			router.ServeHTTP(w, req)
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

			router := mux.NewRouter()
			router.Path("/{uri:.*}").HandlerFunc(LegacyLanding(mockClient))

			router.ServeHTTP(w, req)
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

			router := mux.NewRouter()
			router.Path("/{uri:.*}").HandlerFunc(LegacyLanding(mockClient))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})

	Convey("test filterable landing page", t, func() {
		Convey("test filterable landing page is successful, when it receives good dataset api responses", func() {
			mockClient := NewMockDatasetClient(mockCtrl)
			mockClient.EXPECT().Get("12345").Return(dataset.Model{Contacts: []dataset.Contact{{Name: "Matt"}}, Links: dataset.Links{LatestVersion: dataset.Link{URL: "/datasets/1234/editions/5678/versions/2017"}}}, nil)
			editions := []dataset.Edition{dataset.Edition{Edition: "2016"}}
			mockClient.EXPECT().GetEditions("12345").Return(editions, nil)
			versions := []dataset.Version{dataset.Version{ReleaseDate: "02-01-2005", Links: dataset.Links{Self: dataset.Link{URL: "/datasets/12345/editions/2016/versions/1"}}}}
			mockClient.EXPECT().GetVersions("12345", "2016").Return(versions, nil)
			dims := dataset.Dimensions{
				Items: []dataset.Dimension{
					{
						ID: "aggregate",
					},
				},
			}
			opts := dataset.Options{
				Items: []dataset.Option{
					{
						Label:  "1",
						Option: "abd",
					},
					{
						Label:  "2",
						Option: "fjd",
					},
				},
			}
			mockClient.EXPECT().GetDimensions("12345", "5678", "2017").Return(dims, nil)
			mockClient.EXPECT().GetOptions("12345", "5678", "2017", "aggregate").Return(opts, nil)

			cli = createMockClient([]byte(`<html><body><h1>Some HTML from renderer!</h1></body></html>`), 200)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, "<html><body><h1>Some HTML from renderer!</h1></body></html>")
		})

		Convey("test filterableLanding returns 500 if client Get() returns an error", func() {
			mockClient := NewMockDatasetClient(mockCtrl)
			mockClient.EXPECT().Get("12345").Return(dataset.Model{}, errors.New("sorry"))

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("test filterableLanding returns 500 if client GetEditions() returns error", func() {
			mockClient := NewMockDatasetClient(mockCtrl)
			mockClient.EXPECT().Get("12345").Return(dataset.Model{}, nil)
			editions := []dataset.Edition{dataset.Edition{Edition: "2016"}}
			mockClient.EXPECT().GetEditions("12345").Return(editions, errors.New("sorry"))

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("test filterableLanding returns 500 if client GetVersions() returns error", func() {
			mockClient := NewMockDatasetClient(mockCtrl)
			mockClient.EXPECT().Get("12345").Return(dataset.Model{}, nil)
			editions := []dataset.Edition{dataset.Edition{Edition: "2016"}}
			mockClient.EXPECT().GetEditions("12345").Return(editions, nil)
			versions := []dataset.Version{dataset.Version{ReleaseDate: "02-01-2005", Links: dataset.Links{Self: dataset.Link{URL: "/datasets/12345/editions/2016/versions/1"}}}}
			mockClient.EXPECT().GetVersions("12345", "2016").Return(versions, errors.New("sorry"))

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("test filterableLanding returns 500 if renderer returns error", func() {
			mockClient := NewMockDatasetClient(mockCtrl)
			mockClient.EXPECT().Get("12345").Return(dataset.Model{}, nil)
			editions := []dataset.Edition{dataset.Edition{Edition: "2016"}}
			mockClient.EXPECT().GetEditions("12345").Return(editions, nil)
			versions := []dataset.Version{dataset.Version{ReleaseDate: "02-01-2005", Links: dataset.Links{Self: dataset.Link{URL: "/datasets/12345/editions/2016/versions/1"}}}}
			mockClient.EXPECT().GetVersions("12345", "2016").Return(versions, nil)

			cli = createMockClient(nil, 500)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})

}

func testResponse(code int, respBody, url string, client FilterClient, dc DatasetClient, f http.HandlerFunc) *httptest.ResponseRecorder {
	req, err := http.NewRequest("POST", url, nil)
	So(err, ShouldBeNil)

	w := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter", CreateFilterID(client, dc))

	router.ServeHTTP(w, req)

	So(w.Code, ShouldEqual, code)

	b, err := ioutil.ReadAll(w.Body)
	So(err, ShouldBeNil)

	So(string(b), ShouldEqual, respBody)

	return w
}
