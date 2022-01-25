package handlers

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"

	coreModel "github.com/ONSdigital/dp-renderer/model"
)

type testCliError struct{}

func (e *testCliError) Error() string { return "client error" }
func (e *testCliError) Code() int     { return http.StatusNotFound }

const serviceAuthToken = ""
const userAuthToken = ""
const collectionID = ""
const locale = "en"

// datasetOptions returns a mocked dataset.Options struct according to the provided offset and limit
func datasetOptions(offset, limit int) dataset.Options {
	allItems := []dataset.Option{
		{
			Label:  "1",
			Option: "abd",
		},
		{
			Label:  "2",
			Option: "fjd",
		},
	}
	o := dataset.Options{
		Offset:     offset,
		Limit:      limit,
		TotalCount: len(allItems),
	}
	o.Items = slice(allItems, offset, limit)
	o.Count = len(o.Items)
	return o
}

func slice(full []dataset.Option, offset, limit int) (sliced []dataset.Option) {
	end := offset + limit
	if end > len(full) {
		end = len(full)
	}

	if offset > len(full) || limit == 0 {
		return []dataset.Option{}
	}

	return full[offset:end]
}

func TestUnitHandlers(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()
	cfg := initialiseMockConfig()

	Convey("test setStatusCode", t, func() {

		Convey("test status code handles 404 response from client", func() {
			req := httptest.NewRequest("GET", "http://localhost:20000", nil)
			w := httptest.NewRecorder()
			err := &testCliError{}

			setStatusCode(req, w, err)

			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("test status code handles internal server error", func() {
			req := httptest.NewRequest("GET", "http://localhost:20000", nil)
			w := httptest.NewRecorder()
			err := errors.New("internal server error")

			setStatusCode(req, w, err)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})

	Convey("test CreateFilterID", t, func() {
		Convey("test CreateFilterID handler, creates a filter id and redirects", func() {
			mockClient := NewMockFilterClient(mockCtrl)
			mockClient.EXPECT().CreateBlueprint(ctx, userAuthToken, serviceAuthToken, "", collectionID, "1234", "5678", "2017", []string{"aggregate", "time"}).Return("12345", "testETag", nil)

			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			dims := dataset.VersionDimensions{
				Items: []dataset.VersionDimension{
					{
						Name: "aggregate",
					},
					{
						Name: "time",
					},
				},
			}
			mockDatasetClient.EXPECT().GetVersionDimensions(ctx, userAuthToken, serviceAuthToken, collectionID, "1234", "5678", "2017").Return(dims, nil)
			mockDatasetClient.EXPECT().GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, "1234", "5678", "2017", "aggregate",
				&dataset.QueryParams{Offset: 0, Limit: 0}).Return(datasetOptions(0, 0), nil)
			mockDatasetClient.EXPECT().GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, "1234", "5678", "2017", "time",
				&dataset.QueryParams{Offset: 0, Limit: 0}).Return(datasetOptions(0, 0), nil)

			w := testResponse(301, "", "/datasets/1234/editions/5678/versions/2017/filter", mockClient, mockDatasetClient)

			location := w.Header().Get("Location")
			So(location, ShouldNotBeEmpty)

			So(location, ShouldEqual, "/filters/12345/dimensions")
		})

		Convey("test CreateFilterID returns 500 if unable to create a blueprint on filter api", func() {
			mockClient := NewMockFilterClient(mockCtrl)
			mockClient.EXPECT().CreateBlueprint(ctx, userAuthToken, serviceAuthToken, "", collectionID, "1234", "5678", "2017", gomock.Any()).Return("", "", errors.New("unable to create filter blueprint"))

			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockDatasetClient.EXPECT().GetVersionDimensions(ctx, userAuthToken, serviceAuthToken, collectionID, "1234", "5678", "2017").Return(dataset.VersionDimensions{}, nil)

			testResponse(500, "", "/datasets/1234/editions/5678/versions/2017/filter", mockClient, mockDatasetClient)
		})
	})

	Convey("test /data endpoint", t, func() {

		Convey("test successful json response", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockZebedeeClient.EXPECT().Get(ctx, "12345", "/data?uri=/data").Return([]byte(`{"some_json":true}`), nil)
			mockConfig := config.Config{}

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/data", nil)
			So(err, ShouldBeNil)
			req.AddCookie(&http.Cookie{Name: "access_token", Value: "12345"})

			router := mux.NewRouter()
			router.Path("/{uri:.*}").HandlerFunc(LegacyLanding(mockZebedeeClient, mockDatasetClient, nil, mockConfig))

			router.ServeHTTP(w, req)

			So(w.Body.String(), ShouldEqual, `{"some_json":true}`)
		})

		Convey("test status 500 returned if zedbedee get returns error", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockZebedeeClient.EXPECT().Get(ctx, userAuthToken, "/data?uri=/data").Return(nil, errors.New("something went wrong with zebedee"))
			mockConfig := config.Config{}

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/data", nil)
			So(err, ShouldBeNil)
			router := mux.NewRouter()
			router.Path("/{uri:.*}").HandlerFunc(LegacyLanding(mockZebedeeClient, mockDatasetClient, nil, mockConfig))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})

	Convey("test legacylanding handler with non /data endpoint", t, func() {
		Convey("test sucessful data retrieval and rendering", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
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
			router.Path("/{uri:.*}").HandlerFunc(LegacyLanding(mockZebedeeClient, mockDatasetClient, mockRend, mockConfig))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("test status 500 returned when zebedee client returns error retrieving landing page", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockConfig := config.Config{}
			dlp := zebedee.DatasetLandingPage{}
			mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, userAuthToken, collectionID, locale, "/somelegacypage").Return(dlp, errors.New("something went wrong :("))

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/somelegacypage", nil)
			So(err, ShouldBeNil)

			router := mux.NewRouter()
			router.Path("/{uri:.*}").HandlerFunc(LegacyLanding(mockZebedeeClient, mockDatasetClient, nil, mockConfig))

			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("test status 500 returned when zebedee client returns error retrieving breadcrumb", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockConfig := config.Config{}
			dlp := zebedee.DatasetLandingPage{URI: "https://helloworld.com"}
			mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, userAuthToken, collectionID, locale, "/somelegacypage").Return(dlp, nil)
			mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthToken, collectionID, locale, dlp.URI).Return(nil, errors.New("something went wrong"))

			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/somelegacypage", nil)
			So(err, ShouldBeNil)

			router := mux.NewRouter()
			router.Path("/{uri:.*}").HandlerFunc(LegacyLanding(mockZebedeeClient, mockDatasetClient, nil, mockConfig))

			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})

	Convey("test filterable landing page", t, func() {

		Convey("test filterable landing page is successful, when it receives good dataset api responses", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockClient := NewMockDatasetClient(mockCtrl)
			mockConfig := config.Config{}
			mockClient.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").Return(dataset.DatasetDetails{Contacts: &[]dataset.Contact{{Name: "Matt"}}, URI: "/economy/grossdomesticproduct/datasets/gdpjanuary2018", Links: dataset.Links{LatestVersion: dataset.Link{URL: "/datasets/1234/editions/5678/versions/2017"}}}, nil)
			versions := dataset.VersionsList{
				Items: []dataset.Version{
					{
						Links:       dataset.Links{Self: dataset.Link{URL: "/datasets/12345/editions/2016/versions/1"}},
						ReleaseDate: "02-01-2005",
					},
				},
			}
			mockClient.EXPECT().GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "5678", &dataset.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
			mockClient.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "5678", "2017").Return(versions.Items[0], nil)
			dims := dataset.VersionDimensions{
				Items: []dataset.VersionDimension{
					{
						Name: "aggregate",
					},
				},
			}
			mockClient.EXPECT().GetVersionDimensions(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "5678", "2017").Return(dims, nil)
			mockClient.EXPECT().GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "5678", "2017", "aggregate",
				&dataset.QueryParams{Offset: 0, Limit: numOptsSummary}).Return(datasetOptions(0, numOptsSummary), nil)
			mockClient.EXPECT().GetVersionMetadata(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "5678", "2017")
			mockClient.EXPECT().GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "5678", "2017", "aggregate",
				&dataset.QueryParams{Offset: 0, Limit: maxMetadataOptions}).Return(datasetOptions(0, maxMetadataOptions), nil)
			mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthToken, collectionID, locale, "")
			mockZebedeeClient.EXPECT().GetHomepageContent(ctx, userAuthToken, collectionID, locale, "/")

			mockRend := NewMockRenderClient(mockCtrl)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "filterable")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockRend, mockZebedeeClient, mockConfig, ""))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("test filterableLanding returns 302 and redirects to the correct url for edition level requests without version", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockClient := NewMockDatasetClient(mockCtrl)
			mockConfig := config.Config{}
			mockZebedeeClient.EXPECT().GetHomepageContent(ctx, userAuthToken, collectionID, locale, "/")
			mockClient.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").Return(dataset.DatasetDetails{Contacts: &[]dataset.Contact{{Name: "Matt"}}, URI: "/economy/grossdomesticproduct/datasets/gdpjanuary2018", Links: dataset.Links{LatestVersion: dataset.Link{URL: "/datasets/1234/editions/5678/versions/2017"}}}, nil)
			versions := dataset.VersionsList{
				Items: []dataset.Version{
					{
						Links:       dataset.Links{Self: dataset.Link{URL: "/datasets/12345/editions/2016/versions/1"}},
						ReleaseDate: "02-01-2005",
					},
				},
			}
			mockClient.EXPECT().GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "5678", &dataset.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)

			mockRend := NewMockRenderClient(mockCtrl)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/5678", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}", FilterableLanding(mockClient, mockRend, mockZebedeeClient, mockConfig, ""))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/datasets/12345/editions/5678/versions/1\">Found</a>.\n\n")
		})

		Convey("test filterableLanding returns 500 if client Get() returns an error", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockClient := NewMockDatasetClient(mockCtrl)
			mockConfig := config.Config{}
			mockClient.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").Return(dataset.DatasetDetails{}, errors.New("sorry"))

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, nil, mockZebedeeClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("test filterableLanding returns 500 if client GetVersions() returns error", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockClient := NewMockDatasetClient(mockCtrl)
			mockConfig := config.Config{}
			mockClient.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").Return(dataset.DatasetDetails{}, nil)
			versions := dataset.VersionsList{
				Items: []dataset.Version{
					{
						Links:       dataset.Links{Self: dataset.Link{URL: "/datasets/12345/editions/2016/versions/1"}},
						ReleaseDate: "02-01-2005",
					},
				},
			}
			mockClient.EXPECT().GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "5678", &dataset.QueryParams{Offset: 0, Limit: 1000}).Return(versions, errors.New("sorry"))
			mockZebedeeClient.EXPECT().GetHomepageContent(ctx, userAuthToken, collectionID, locale, "/")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/5678", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}", FilterableLanding(mockClient, nil, mockZebedeeClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

	})

	Convey("test census landing page", t, func() {
		const numOptsSummary = 1000
		mockClient := NewMockDatasetClient(mockCtrl)
		mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
		mockZebedeeClient.EXPECT().GetHomepageContent(ctx, userAuthToken, collectionID, locale, "/")
		mockRend := NewMockRenderClient(mockCtrl)
		dims := dataset.VersionDimensions{
			Items: []dataset.VersionDimension{
				{
					Name: "city",
				},
			},
		}

		Convey("filterable landing handler returns census landing template for cantabular types", func() {
			mockConfig := config.Config{EnableCensusPages: true}
			mockClient.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").Return(dataset.DatasetDetails{Contacts: &[]dataset.Contact{{Name: "Nick"}}, Type: "cantabular-table", URI: "/economy/grossdomesticproduct/datasets/gdpjanuary2018", Links: dataset.Links{LatestVersion: dataset.Link{URL: "/datasets/12345/editions/2021/versions/1"}}, ID: "12345"}, nil)
			versions := dataset.VersionsList{
				Items: []dataset.Version{
					{
						Downloads: map[string]dataset.Download{
							"XLS": {
								Size: "78600",
								URL:  "https://www.my-url.com/file.xls",
							}},
						ReleaseDate: "02-01-2005",
						Version:     1,
						Links: dataset.Links{
							Self: dataset.Link{
								URL: "/datasets/12345/editions/2021/versions/1",
							},
						},
					},
				},
			}
			mockClient.EXPECT().GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
			mockClient.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").Return(versions.Items[0], nil)
			mockClient.EXPECT().GetVersionDimensions(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "2021", "1").Return(dims, nil)
			mockClient.EXPECT().GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "2021", "1", "city",
				&dataset.QueryParams{Offset: 0, Limit: numOptsSummary}).Return(datasetOptions(0, numOptsSummary), nil)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockRend, mockZebedeeClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("census dataset landing page correctly fetches version 1 data for initial release date field, when loading a later version", func() {
			mockConfig := config.Config{EnableCensusPages: true}
			mockClient.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").Return(dataset.DatasetDetails{Contacts: &[]dataset.Contact{{Name: "Nick"}}, Type: "cantabular-table", URI: "/economy/grossdomesticproduct/datasets/gdpjanuary2018", Links: dataset.Links{LatestVersion: dataset.Link{URL: "/datasets/12345/editions/2021/versions/2"}}, ID: "12345"}, nil)
			versions := dataset.VersionsList{
				Items: []dataset.Version{
					{ReleaseDate: "02-01-2005", Version: 1, Links: dataset.Links{Self: dataset.Link{URL: "/datasets/12345/editions/2021/versions/1"}}},
					{ReleaseDate: "05-01-2005", Version: 2, Links: dataset.Links{Self: dataset.Link{URL: "/datasets/12345/editions/2021/versions/2"}}},
				},
			}
			mockClient.EXPECT().GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
			mockClient.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "2").Return(versions.Items[1], nil)
			mockClient.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").Return(versions.Items[0], nil)
			mockClient.EXPECT().GetVersionDimensions(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "2021", "2").Return(dims, nil)
			mockClient.EXPECT().GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "2021", "2", "city",
				&dataset.QueryParams{Offset: 0, Limit: numOptsSummary}).Return(datasetOptions(0, numOptsSummary), nil)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockRend, mockZebedeeClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("census dataset landing page returns 200 when no downloadable files provided", func() {
			mockConfig := config.Config{EnableCensusPages: true}
			mockClient.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").Return(dataset.DatasetDetails{Contacts: &[]dataset.Contact{{Name: "Nick"}}, Type: "cantabular-table", URI: "/economy/grossdomesticproduct/datasets/gdpjanuary2018", Links: dataset.Links{LatestVersion: dataset.Link{URL: "/datasets/12345/editions/2021/versions/1"}}, ID: "12345"}, nil)
			versions := dataset.VersionsList{
				Items: []dataset.Version{
					{
						Downloads:   nil,
						ReleaseDate: "02-01-2005",
						Version:     1,
						Links: dataset.Links{
							Self: dataset.Link{
								URL: "/datasets/12345/editions/2021/versions/1",
							},
						},
					},
				},
			}

			mockClient.EXPECT().GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
			mockClient.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").Return(versions.Items[0], nil)
			mockClient.EXPECT().GetVersionDimensions(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "2021", "1").Return(dims, nil)
			mockClient.EXPECT().GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "2021", "1", "city",
				&dataset.QueryParams{Offset: 0, Limit: numOptsSummary}).Return(datasetOptions(0, numOptsSummary), nil)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockRend, mockZebedeeClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("census dataset landing page returns 302 when valid download option chosen", func() {
			mockConfig := config.Config{EnableCensusPages: true}
			mockClient.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").Return(dataset.DatasetDetails{Contacts: &[]dataset.Contact{{Name: "Nick"}}, Type: "cantabular-table", URI: "/economy/grossdomesticproduct/datasets/gdpjanuary2018", Links: dataset.Links{LatestVersion: dataset.Link{URL: "/datasets/12345/editions/2021/versions/1"}}, ID: "12345"}, nil)
			versions := dataset.VersionsList{
				Items: []dataset.Version{
					{
						Downloads: map[string]dataset.Download{
							"CSV": {
								Size: "1234",
								URL:  "https://a.domain.com/a-file.csv",
							},
						},
						ReleaseDate: "02-01-2005",
						Version:     1,
						Links: dataset.Links{
							Self: dataset.Link{
								URL: "/datasets/12345/editions/2021/versions/1",
							},
						},
					},
				},
			}
			mockClient.EXPECT().GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
			mockClient.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").Return(versions.Items[0], nil)
			mockClient.EXPECT().GetVersionDimensions(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "2021", "1").Return(dims, nil)
			mockClient.EXPECT().GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "2021", "1", "city",
				&dataset.QueryParams{Offset: 0, Limit: numOptsSummary}).Return(datasetOptions(0, numOptsSummary), nil)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345?f=get-data&format=csv", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockRend, mockZebedeeClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusFound)
		})

		Convey("census dataset landing page returns 200 when invalid download option chosen", func() {
			mockConfig := config.Config{EnableCensusPages: true}
			mockClient.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").Return(dataset.DatasetDetails{Contacts: &[]dataset.Contact{{Name: "Nick"}}, Type: "cantabular-table", URI: "/economy/grossdomesticproduct/datasets/gdpjanuary2018", Links: dataset.Links{LatestVersion: dataset.Link{URL: "/datasets/12345/editions/2021/versions/1"}}, ID: "12345"}, nil)
			versions := dataset.VersionsList{
				Items: []dataset.Version{
					{
						Downloads:   nil,
						ReleaseDate: "02-01-2005",
						Version:     1,
						Links: dataset.Links{
							Self: dataset.Link{
								URL: "/datasets/12345/editions/2021/versions/1",
							},
						},
					},
				},
			}
			mockClient.EXPECT().GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
			mockClient.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").Return(versions.Items[0], nil)
			mockClient.EXPECT().GetVersionDimensions(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "2021", "1").Return(dims, nil)
			mockClient.EXPECT().GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "2021", "1", "city",
				&dataset.QueryParams{Offset: 0, Limit: numOptsSummary}).Return(datasetOptions(0, numOptsSummary), nil)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345?f=get-data&format=aFormat", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockRend, mockZebedeeClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("census dataset landing page returns 200 when unknown get query request made", func() {
			mockConfig := config.Config{EnableCensusPages: true}
			mockClient.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").Return(dataset.DatasetDetails{Contacts: &[]dataset.Contact{{Name: "Nick"}}, Type: "cantabular-table", URI: "/economy/grossdomesticproduct/datasets/gdpjanuary2018", Links: dataset.Links{LatestVersion: dataset.Link{URL: "/datasets/12345/editions/2021/versions/1"}}, ID: "12345"}, nil)
			versions := dataset.VersionsList{
				Items: []dataset.Version{
					{
						Downloads:   nil,
						ReleaseDate: "02-01-2005",
						Version:     1,
						Links: dataset.Links{
							Self: dataset.Link{
								URL: "/datasets/12345/editions/2021/versions/1",
							},
						},
					},
				},
			}
			mockClient.EXPECT().GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
			mockClient.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").Return(versions.Items[0], nil)
			mockClient.EXPECT().GetVersionDimensions(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "2021", "1").Return(dims, nil)
			mockClient.EXPECT().GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "2021", "1", "city",
				&dataset.QueryParams{Offset: 0, Limit: numOptsSummary}).Return(datasetOptions(0, numOptsSummary), nil)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345?f=blah-blah&format=bob", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockRend, mockZebedeeClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("filterable landing page returned if census config is false", func() {
			const numOptsSummary = 50
			mockConfig := config.Config{EnableCensusPages: false}
			mockClient.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").Return(dataset.DatasetDetails{Contacts: &[]dataset.Contact{{Name: "Nick"}}, Type: "cantabular-table", URI: "/economy/grossdomesticproduct/datasets/gdpjanuary2018", Links: dataset.Links{LatestVersion: dataset.Link{URL: "/datasets/1234/editions/5678/versions/2017"}}}, nil)
			versions := dataset.VersionsList{
				Items: []dataset.Version{
					{
						Downloads:   nil,
						ReleaseDate: "02-01-2005",
						Version:     1,
						Links: dataset.Links{
							Self: dataset.Link{
								URL: "/datasets/12345/editions/2016/versions/1",
							},
						},
					},
				},
			}
			mockClient.EXPECT().GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "5678", &dataset.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
			mockClient.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "5678", "2017").Return(versions.Items[0], nil)
			dims := dataset.VersionDimensions{
				Items: []dataset.VersionDimension{
					{
						Name: "aggregate",
					},
				},
			}
			mockClient.EXPECT().GetVersionDimensions(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "5678", "2017").Return(dims, nil)
			mockClient.EXPECT().GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "5678", "2017", "aggregate",
				&dataset.QueryParams{Offset: 0, Limit: numOptsSummary}).Return(datasetOptions(0, numOptsSummary), nil)
			mockClient.EXPECT().GetVersionMetadata(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "5678", "2017")
			mockClient.EXPECT().GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "5678", "2017", "aggregate",
				&dataset.QueryParams{Offset: 0, Limit: maxMetadataOptions}).Return(datasetOptions(0, maxMetadataOptions), nil)
			mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthToken, collectionID, locale, "")

			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "filterable")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockRend, mockZebedeeClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})
	})

	Convey("test versions list", t, func() {
		Convey("test versions list returns 200 when rendered succesfully", func() {
			mockClient := NewMockDatasetClient(mockCtrl)
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockZebedeeClient.EXPECT().GetHomepageContent(ctx, userAuthToken, collectionID, locale, "/")
			mockConfig := config.Config{}
			mockClient.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").Return(dataset.DatasetDetails{}, nil)
			mockClient.EXPECT().GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2017", &dataset.QueryParams{Offset: 0, Limit: 1000}).Return(dataset.VersionsList{Items: []dataset.Version{}}, nil)
			mockClient.EXPECT().GetEdition(ctx, userAuthToken, serviceAuthToken, collectionID, "12345", "2017").Return(dataset.Edition{}, nil)

			mockRend := NewMockRenderClient(mockCtrl)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "version-list")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2017/versions", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{edition}/versions", VersionsList(mockClient, mockZebedeeClient, mockRend, mockConfig))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("test versions list returns status 500 when dataset client returns an error", func() {
			mockClient := NewMockDatasetClient(mockCtrl)
			mockConfig := config.Config{}
			mockClient.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").Return(dataset.DatasetDetails{}, errors.New("dataset client error"))

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2017/versions", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{edition}/versions", VersionsList(mockClient, nil, nil, mockConfig))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})

	Convey("test CreateFilterFlexID", t, func() {
		Convey("test CreateFilterFlexID handler, creates a filter id and redirects", func() {
			mockClient := NewMockFilterClient(mockCtrl)
			mockConfig := config.Config{EnableCensusPages: true}
			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/flex", CreateFilterFlexID(mockClient, mockConfig))
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/datasets/12345/editions/2012/versions/1/flex", nil)
			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusMovedPermanently)
		})

		Convey("test post route fails if config is false", func() {
			mockClient := NewMockFilterClient(mockCtrl)
			mockConfig := config.Config{EnableCensusPages: false}
			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/flex", CreateFilterFlexID(mockClient, mockConfig))
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/datasets/12345/editions/2012/versions/1/flex", nil)
			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})

}

func testResponse(code int, respBody, url string, fc FilterClient, dc DatasetClient) *httptest.ResponseRecorder {
	req, err := http.NewRequest("POST", url, nil)
	So(err, ShouldBeNil)

	w := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter", CreateFilterID(fc, dc))

	router.ServeHTTP(w, req)

	So(w.Code, ShouldEqual, code)

	b, err := ioutil.ReadAll(w.Body)
	So(err, ShouldBeNil)

	So(string(b), ShouldEqual, respBody)

	return w
}

func initialiseMockConfig() config.Config {
	return config.Config{
		PatternLibraryAssetsPath: "http://localhost:9000/dist",
		SiteDomain:               "ons",
		SupportedLanguages:       []string{"en", "cy"},
	}
}
