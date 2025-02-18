package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

// Tests for `FilterableLanding` for any dataset type
func TestFilterableLandingPage(t *testing.T) {
	// Mocks
	mockConfig := initialiseMockConfig()
	mockContext := gomock.Any()
	mockController := gomock.NewController(t)
	mockDatasetClient := NewMockDatasetClient(mockController)
	mockPopulationClient := NewMockPopulationClient(mockController)
	mockRenderClient := NewMockRenderClient(mockController)
	mockZebedeeClient := NewMockZebedeeClient(mockController)

	// Default test values
	apiRouterVersion := "/v1"
	datasetId := "12345"
	downloadServiceAuthToken := ""
	editionId := "67890"
	getVersionsQueryParams := dataset.QueryParams{Offset: 0, Limit: 1000}
	mockGetResponse := dataset.DatasetDetails{}
	mockGetVersionsResponse := dataset.VersionsList{}

	Convey("Test filterable landing page", t, func() {
		Convey("Test filterableLanding returns 500 error if dataset is not found", func() {
			// Dataset client `Get()` will return an error if dataset is not found
			mockDatasetClient.EXPECT().Get(
				mockContext, userAuthToken, serviceAuthToken, collectionID, datasetId,
			).Return(
				mockGetResponse, errors.New("sorry"),
			)

			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest("GET", fmt.Sprintf("/datasets/%s", datasetId), nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockDatasetClient, mockPopulationClient, mockRenderClient, mockZebedeeClient, mockConfig, apiRouterVersion))

			router.ServeHTTP(mockRequestWriter, mockRequest)

			So(mockRequestWriter.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Test filterableLanding returns 500 if dataset versions are not found", func() {
			// Dataset client `Get()` will return valid response if dataset found
			mockDatasetClient.EXPECT().Get(
				mockContext, userAuthToken, serviceAuthToken, collectionID, datasetId,
			).Return(
				mockGetResponse, nil,
			)
			// Dataset client `GetVersions()` will return an error if dataset versions are not found
			mockDatasetClient.EXPECT().GetVersions(
				mockContext, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetId, editionId, &getVersionsQueryParams,
			).Return(
				mockGetVersionsResponse, errors.New("sorry"),
			)

			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest("GET", fmt.Sprintf("/datasets/%s/editions/%s", datasetId, editionId), nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}", FilterableLanding(mockDatasetClient, mockPopulationClient, mockRenderClient, mockZebedeeClient, mockConfig, apiRouterVersion))

			router.ServeHTTP(mockRequestWriter, mockRequest)

			So(mockRequestWriter.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func TestFilterableLandingPageFilterableDataType(t *testing.T) {
	// Mocks
	mockConfig := initialiseMockConfig()
	mockContext := gomock.Any()
	mockController := gomock.NewController(t)
	mockDatasetClient := NewMockDatasetClient(mockController)
	mockPopulationClient := NewMockPopulationClient(mockController)
	mockRenderClient := NewMockRenderClient(mockController)
	mockZebedeeClient := NewMockZebedeeClient(mockController)

	datasetId := "12345"
	datasetType := "filterable"
	downloadServiceAuthToken := ""
	getVersionsQueryParams := dataset.QueryParams{Offset: 0, Limit: 1000}
	editionId := "5678"
	versionId := "2017"
	mockGetResponse := dataset.DatasetDetails{
		Contacts: &[]dataset.Contact{
			{Name: "Matt"}},
		URI: "/economy/grossdomesticproduct/datasets/gdpjanuary2018",
		Links: dataset.Links{
			LatestVersion: dataset.Link{
				URL: "/datasets/1234/editions/5678/versions/2017",
			},
		},
		Type: datasetType,
		ID:   datasetId,
	}
	mockGetVersionsResponse := dataset.VersionsList{
		Items: []dataset.Version{
			{
				Links:       dataset.Links{Self: dataset.Link{URL: "/datasets/12345/editions/2016/versions/1"}},
				ReleaseDate: "02-01-2005",
			},
		},
	}
	mockGetVersionDimsResponse := dataset.VersionDimensions{
		Items: []dataset.VersionDimension{
			{
				Name: "aggregate",
			},
		},
	}

	Convey("test filterable landing page", t, func() {
		Convey("test filterable landing page is successful, when it receives good dataset api responses", func() {

			mockDatasetClient.EXPECT().Get(
				mockContext, userAuthToken, serviceAuthToken, collectionID, datasetId,
			).Return(
				mockGetResponse, nil,
			)
			mockDatasetClient.EXPECT().GetVersions(
				mockContext, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetId, editionId, &getVersionsQueryParams,
			).Return(
				mockGetVersionsResponse, nil,
			)
			mockDatasetClient.EXPECT().GetVersion(
				mockContext, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetId, editionId, versionId,
			).Return(
				mockGetVersionsResponse.Items[0], nil,
			)
			mockDatasetClient.EXPECT().GetVersionDimensions(
				mockContext, userAuthToken, serviceAuthToken, collectionID, datasetId, editionId, versionId,
			).Return(
				mockGetVersionDimsResponse, nil,
			)
			mockDatasetClient.EXPECT().GetOptions(
				mockContext, userAuthToken, serviceAuthToken, collectionID, datasetId, editionId, versionId, "aggregate",
				&dataset.QueryParams{Offset: 0, Limit: numOptsSummary},
			).Return(
				datasetOptions(0, numOptsSummary), nil,
			)
			mockDatasetClient.EXPECT().GetVersionMetadata(
				mockContext, userAuthToken, serviceAuthToken, collectionID, datasetId, editionId, versionId,
			)
			mockDatasetClient.EXPECT().GetOptions(
				mockContext, userAuthToken, serviceAuthToken, collectionID, datasetId, editionId, versionId, "aggregate",
				&dataset.QueryParams{Offset: 0, Limit: maxMetadataOptions},
			).Return(
				datasetOptions(0, maxMetadataOptions), nil,
			)
			mockZebedeeClient.EXPECT().GetHomepageContent(mockContext, userAuthToken, collectionID, locale, "/")

			mockZebedeeClient.EXPECT().GetBreadcrumb(mockContext, userAuthToken, collectionID, locale, "")
			mockRenderClient.EXPECT().NewBasePageModel().Return(
				coreModel.NewPage(mockConfig.PatternLibraryAssetsPath, mockConfig.SiteDomain),
			)
			// `BuildPage` should be called with the `dataset.DatasetDetails.Type` defining the template to be used
			mockRenderClient.EXPECT().BuildPage(gomock.Any(), gomock.Any(), datasetType)

			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest("GET", "/datasets/12345", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockDatasetClient, mockPopulationClient, mockRenderClient, mockZebedeeClient, mockConfig, ""))

			router.ServeHTTP(mockRequestWriter, mockRequest)

			So(mockRequestWriter.Code, ShouldEqual, http.StatusOK)
		})

		Convey("test filterableLanding returns 302 and redirects to the correct url for edition level requests without version", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockController)
			mockDatasetClient.EXPECT().Get(mockContext, userAuthToken, serviceAuthToken, collectionID, "12345").Return(dataset.DatasetDetails{Contacts: &[]dataset.Contact{{Name: "Matt"}}, URI: "/economy/grossdomesticproduct/datasets/gdpjanuary2018", Links: dataset.Links{LatestVersion: dataset.Link{URL: "/datasets/1234/editions/5678/versions/2017"}}}, nil)
			versions := dataset.VersionsList{
				Items: []dataset.Version{
					{
						Links:       dataset.Links{Self: dataset.Link{URL: "/datasets/12345/editions/2016/versions/1"}},
						ReleaseDate: "02-01-2005",
					},
				},
			}
			mockDatasetClient.EXPECT().GetVersions(mockContext, userAuthToken, serviceAuthToken, collectionID, "", "12345", "5678", &dataset.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/5678", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}", FilterableLanding(mockDatasetClient, mockPopulationClient, mockRenderClient, mockZebedeeClient, mockConfig, ""))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/datasets/12345/editions/5678/versions/1\">Found</a>.\n\n")
		})
	})
}

func TestFilterableLandingPageCantabularDataTypes(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockPc := NewMockPopulationClient(mockCtrl)

	defer mockCtrl.Finish()
	ctx := gomock.Any()
	cfg := initialiseMockConfig()

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
				GetPopulationType(gomock.Any(), gomock.Any()).
				Return(population.GetPopulationTypeResponse{}, nil).
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

func TestFilterableLandingPageStaticDataType(t *testing.T) {
	// Mocks
	mockConfig := initialiseMockConfig()
	mockContext := gomock.Any()
	mockController := gomock.NewController(t)
	mockDatasetClient := NewMockDatasetClient(mockController)
	mockPopulationClient := NewMockPopulationClient(mockController)
	mockRenderClient := NewMockRenderClient(mockController)
	mockZebedeeClient := NewMockZebedeeClient(mockController)

	datasetId := "12345"
	datasetType := "static"
	downloadServiceAuthToken := ""
	getVersionsQueryParams := dataset.QueryParams{Offset: 0, Limit: 1000}
	editionId := "5678"
	versionId := "1"
	mockGetResponse := dataset.DatasetDetails{
		Contacts: &[]dataset.Contact{
			{Name: "Matt"}},
		URI: "/economy/grossdomesticproduct/datasets/gdpjanuary2018",
		Links: dataset.Links{
			LatestVersion: dataset.Link{
				URL: "/datasets/1234/editions/5678/versions/2017",
			},
		},
		Type: datasetType,
		ID:   datasetId,
	}
	mockGetVersionsResponse := dataset.VersionsList{
		Items: []dataset.Version{
			{
				Links:       dataset.Links{Self: dataset.Link{URL: "/datasets/12345/editions/2016/versions/1"}},
				ReleaseDate: "02-01-2005",
				Version:     1,
			},
		},
	}

	Convey("test filterable landing page", t, func() {
		Convey("test filterable landing page is successful, when it receives good dataset api responses", func() {

			mockDatasetClient.EXPECT().Get(
				mockContext, userAuthToken, serviceAuthToken, collectionID, datasetId,
			).Return(
				mockGetResponse, nil,
			)
			mockDatasetClient.EXPECT().GetVersions(
				mockContext, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetId, editionId, &getVersionsQueryParams,
			).Return(
				mockGetVersionsResponse, nil,
			)
			mockDatasetClient.EXPECT().GetVersion(
				mockContext, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetId, editionId, versionId,
			).Return(
				mockGetVersionsResponse.Items[0], nil,
			)
			mockZebedeeClient.EXPECT().GetHomepageContent(mockContext, userAuthToken, collectionID, locale, "/")
			mockDatasetClient.EXPECT().GetVersionMetadata(
				mockContext, userAuthToken, serviceAuthToken, collectionID, datasetId, editionId, versionId,
			)
			mockPopulationClient.EXPECT().GetPopulationType(gomock.Any(), gomock.Any()).Return(
				population.GetPopulationTypeResponse{}, nil,
			).AnyTimes()

			mockRenderClient.EXPECT().NewBasePageModel().Return(
				coreModel.NewPage(mockConfig.PatternLibraryAssetsPath, mockConfig.SiteDomain),
			)
			// `BuildPage` should be called with the `dataset.DatasetDetails.Type` defining the template to be used
			mockRenderClient.EXPECT().BuildPage(gomock.Any(), gomock.Any(), datasetType)

			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest("GET", fmt.Sprintf("/datasets/%s/editions/%s/versions/%s", datasetId, editionId, versionId), nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}", FilterableLanding(mockDatasetClient, mockPopulationClient, mockRenderClient, mockZebedeeClient, mockConfig, ""))

			router.ServeHTTP(mockRequestWriter, mockRequest)

			So(mockRequestWriter.Code, ShouldEqual, http.StatusOK)
		})

		Convey("test filterableLanding returns 302 and redirects to the correct url for edition level requests without version", func() {
			mockDatasetClient.EXPECT().Get(mockContext, userAuthToken, serviceAuthToken, collectionID, datasetId).Return(mockGetResponse, nil)
			mockDatasetClient.EXPECT().GetVersions(mockContext, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetId, editionId, &getVersionsQueryParams).Return(mockGetVersionsResponse, nil)

			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest("GET", fmt.Sprintf("/datasets/%s/editions/%s", datasetId, editionId), nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}", FilterableLanding(mockDatasetClient, mockPopulationClient, mockRenderClient, mockZebedeeClient, mockConfig, ""))

			router.ServeHTTP(mockRequestWriter, mockRequest)

			So(mockRequestWriter.Code, ShouldEqual, http.StatusFound)
			So(mockRequestWriter.Body.String(), ShouldEqual, "<a href=\"/datasets/12345/editions/5678/versions/1\">Found</a>.\n\n")
		})
	})
}
