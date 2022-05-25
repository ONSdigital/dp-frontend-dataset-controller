package handlers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
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

}

func testResponse(code int, body *strings.Reader, url string, fc FilterClient, dc DatasetClient, filterFlexRoute bool, cfg config.Config) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()

	router := mux.NewRouter()
	if filterFlexRoute {
		router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-flex", CreateFilterFlexID(fc, dc, cfg))
	} else {
		router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter", CreateFilterID(fc, dc))
	}

	router.ServeHTTP(w, req)

	So(w.Code, ShouldEqual, code)

	b, err := ioutil.ReadAll(w.Body)
	So(err, ShouldBeNil)
	// Writer body should be empty, we don't write a response
	So(b, ShouldBeEmpty)

	return w
}

func initialiseMockConfig() config.Config {
	return config.Config{
		PatternLibraryAssetsPath: "http://localhost:9000/dist",
		SiteDomain:               "ons",
		SupportedLanguages:       []string{"en", "cy"},
	}
}
