package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	coreModel "github.com/ONSdigital/dp-renderer/model"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFilterableLandingPage(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockPc := NewMockPopulationClient(mockCtrl)

	defer mockCtrl.Finish()
	ctx := gomock.Any()
	cfg := initialiseMockConfig()

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
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockPc, mockRend, mockZebedeeClient, mockConfig, ""))

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
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}", FilterableLanding(mockClient, mockPc, mockRend, mockZebedeeClient, mockConfig, ""))

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
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockPc, nil, mockZebedeeClient, mockConfig, "/v1"))

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
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}", FilterableLanding(mockClient, mockPc, nil, mockZebedeeClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

	})

	Convey("test census landing page", t, func() {
		mockOpts := []dataset.Options{
			{
				Items: []dataset.Option{
					{
						Label: "an option",
					},
				},
			},
			{
				Items: []dataset.Option{},
			},
		}
		mockClient := NewMockDatasetClient(mockCtrl)
		mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
		mockZebedeeClient.EXPECT().GetHomepageContent(ctx, userAuthToken, collectionID, locale, "/")
		mockRend := NewMockRenderClient(mockCtrl)
		Convey("filterable landing handler returns census landing template for cantabular types", func() {
			mockConfig := config.Config{}
			mockClient.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").Return(dataset.DatasetDetails{Contacts: &[]dataset.Contact{{Name: "Nick"}}, Type: "cantabular-table", URI: "/economy/grossdomesticproduct/datasets/gdpjanuary2018", Links: dataset.Links{LatestVersion: dataset.Link{URL: "/datasets/12345/editions/2021/versions/1"}}, ID: "12345"}, nil)
			versions := dataset.VersionsList{
				Items: []dataset.Version{
					{
						Dimensions: []dataset.VersionDimension{
							{
								Name: "Dim name",
							},
						},
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
						IsBasedOn: &dataset.IsBasedOn{ID: "UR"},
					},
				},
			}
			mockClient.EXPECT().GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
			mockClient.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").Return(versions.Items[0], nil)
			mockClient.EXPECT().GetOptions(ctx, gomock.Any(), gomock.Any(), gomock.Any(), "12345", "2021", "1", versions.Items[0].Dimensions[0].Name,
				&dataset.QueryParams{Offset: 0, Limit: 1000}).Return(mockOpts[0], nil)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")
			mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
				PaginationResponse: population.PaginationResponse{
					TotalCount: 2,
				},
			}, nil).AnyTimes()
			mockPc.
				EXPECT().
				GetPopulationTypes(gomock.Any(), gomock.Any()).
				Return(population.GetPopulationTypesResponse{}, nil).
				AnyTimes()

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockPc, mockRend, mockZebedeeClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("census dataset landing page correctly fetches version 1 data for initial release date field, when loading a later version", func() {
			mockConfig := config.Config{}
			mockClient.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").Return(dataset.DatasetDetails{Contacts: &[]dataset.Contact{{Name: "Nick"}}, Type: "cantabular-table", URI: "/economy/grossdomesticproduct/datasets/gdpjanuary2018", Links: dataset.Links{LatestVersion: dataset.Link{URL: "/datasets/12345/editions/2021/versions/2"}}, ID: "12345"}, nil)
			versions := dataset.VersionsList{
				Items: []dataset.Version{
					{ReleaseDate: "02-01-2005", Version: 1, Links: dataset.Links{Self: dataset.Link{URL: "/datasets/12345/editions/2021/versions/1"}}, Dimensions: []dataset.VersionDimension{{Name: "Dim 1"}}, IsBasedOn: &dataset.IsBasedOn{ID: "UR"}},
					{ReleaseDate: "05-01-2005", Version: 2, Links: dataset.Links{Self: dataset.Link{URL: "/datasets/12345/editions/2021/versions/2"}}, IsBasedOn: &dataset.IsBasedOn{ID: "UR"}},
				},
			}
			mockClient.EXPECT().GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
			mockClient.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "2").Return(versions.Items[1], nil)
			mockClient.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").Return(versions.Items[0], nil)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockPc, mockRend, mockZebedeeClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("census dataset landing page returns 200 when no downloadable files provided", func() {
			mockConfig := config.Config{}
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
						IsBasedOn: &dataset.IsBasedOn{ID: "UR"},
					},
				},
			}

			mockClient.EXPECT().GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
			mockClient.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").Return(versions.Items[0], nil)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockPc, mockRend, mockZebedeeClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("census dataset landing page returns 302 when valid download option chosen", func() {
			mockConfig := config.Config{}
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
						IsBasedOn: &dataset.IsBasedOn{ID: "UR"},
					},
				},
			}
			mockClient.EXPECT().GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
			mockClient.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").Return(versions.Items[0], nil)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")
			mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
				PaginationResponse: population.PaginationResponse{
					TotalCount: 2,
				},
			}, nil).AnyTimes()

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345?f=get-data&format=csv", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockPc, mockRend, mockZebedeeClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusFound)
		})

		Convey("census dataset landing page returns 200 when invalid download option chosen", func() {
			mockConfig := config.Config{}
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
						IsBasedOn: &dataset.IsBasedOn{ID: "UR"},
					},
				},
			}
			mockClient.EXPECT().GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
			mockClient.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").Return(versions.Items[0], nil)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")
			mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
				PaginationResponse: population.PaginationResponse{
					TotalCount: 2,
				},
			}, nil).AnyTimes()

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345?f=get-data&format=aFormat", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockPc, mockRend, mockZebedeeClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("census dataset landing page returns 200 when unknown get query request made", func() {
			mockConfig := config.Config{}
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
						IsBasedOn: &dataset.IsBasedOn{ID: "UR"},
					},
				},
			}
			mockClient.EXPECT().GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
			mockClient.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").Return(versions.Items[0], nil)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")
			mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
				PaginationResponse: population.PaginationResponse{
					TotalCount: 2,
				},
			}, nil).AnyTimes()

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345?f=blah-blah&format=bob", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockPc, mockRend, mockZebedeeClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})
	})
}
