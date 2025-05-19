package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
	dpTopicApiModels "github.com/ONSdigital/dp-topic-api/models"
	dpTopicApiSdk "github.com/ONSdigital/dp-topic-api/sdk"
	dpTopicApiErrors "github.com/ONSdigital/dp-topic-api/sdk/errors"
	dpLogger "github.com/ONSdigital/log.go/v2/log"
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
	mockDatasetClient := NewMockDatasetAPISdkClient(mockController)
	mockPopulationClient := NewMockPopulationClient(mockController)
	mockRenderClient := NewMockRenderClient(mockController)
	mockTopicClient := NewMockTopicAPIClient(mockController)
	mockZebedeeClient := NewMockZebedeeClient(mockController)

	// Default test values
	apiRouterVersion := "/v1"
	datasetID := "12345"
	downloadServiceAuthToken := ""
	editionID := "67890"
	getVersionsQueryParams := dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}
	mockGetResponse := dpDatasetApiModels.Dataset{}
	mockGetVersionsResponse := dpDatasetApiSdk.VersionsList{}
	headers := dpDatasetApiSdk.Headers{
		CollectionID:         collectionID,
		DownloadServiceToken: downloadServiceAuthToken,
		ServiceToken:         serviceAuthToken,
		UserAccessToken:      "",
	}

	Convey("Test filterable landing page", t, func() {
		Convey("Test filterableLanding returns 500 error if dataset is not found", func() {
			// Dataset client `GetDataset()` will return an error if dataset is not found
			mockDatasetClient.EXPECT().GetDataset(
				mockContext, headers, collectionID, datasetID,
			).Return(
				mockGetResponse, errors.New("sorry"),
			)

			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest("GET", fmt.Sprintf("/datasets/%s", datasetID), http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockDatasetClient, mockPopulationClient, mockRenderClient, mockZebedeeClient, mockTopicClient, mockConfig, apiRouterVersion))

			router.ServeHTTP(mockRequestWriter, mockRequest)

			So(mockRequestWriter.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Test filterableLanding returns 500 if dataset versions are not found", func() {
			// Dataset client `GetDataset()` will return valid response if dataset found
			mockDatasetClient.EXPECT().GetDataset(
				mockContext, headers, collectionID, datasetID,
			).Return(
				mockGetResponse, nil,
			)
			// Dataset client `GetVersions()` will return an error if dataset versions are not found
			mockDatasetClient.EXPECT().GetVersions(
				mockContext, headers, datasetID, editionID, &getVersionsQueryParams,
			).Return(
				mockGetVersionsResponse, errors.New("sorry"),
			)

			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest("GET", fmt.Sprintf("/datasets/%s/editions/%s", datasetID, editionID), http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}", FilterableLanding(mockDatasetClient, mockPopulationClient, mockRenderClient, mockZebedeeClient, mockTopicClient, mockConfig, apiRouterVersion))

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
	mockDatasetClient := NewMockDatasetAPISdkClient(mockController)
	mockPopulationClient := NewMockPopulationClient(mockController)
	mockRenderClient := NewMockRenderClient(mockController)
	mockTopicClient := NewMockTopicAPIClient(mockController)
	mockZebedeeClient := NewMockZebedeeClient(mockController)

	datasetID := "12345"
	datasetType := "filterable"
	downloadServiceAuthToken := ""
	getVersionsQueryParams := dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}
	editionID := "5678"
	versionID := "2017"

	mockGetResponse := dpDatasetApiModels.Dataset{
		Contacts: []dpDatasetApiModels.ContactDetails{
			{Name: "Matt"}},
		URI: "/economy/grossdomesticproduct/datasets/gdpjanuary2018",
		Links: &dpDatasetApiModels.DatasetLinks{
			LatestVersion: &dpDatasetApiModels.LinkObject{
				HRef: "/datasets/1234/editions/5678/versions/2017",
			},
		},
		Type: datasetType,
		ID:   datasetID,
	}
	mockGetVersionsResponse := dpDatasetApiSdk.VersionsList{
		Items: []dpDatasetApiModels.Version{
			{
				Links: &dpDatasetApiModels.VersionLinks{
					Self: &dpDatasetApiModels.LinkObject{
						HRef: "/datasets/12345/editions/2016/versions/1",
					},
				},
				ReleaseDate: "02-01-2005",
			},
		},
	}
	mockGetVersionDimsResponse := dpDatasetApiSdk.VersionDimensionsList{
		Items: []dpDatasetApiModels.Dimension{
			{
				Name: "aggregate",
			},
		},
	}
	mockGetVersionDimsOptsResponse := dpDatasetApiSdk.VersionDimensionOptionsList{
		Items: []dpDatasetApiModels.PublicDimensionOption{
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
	headers := dpDatasetApiSdk.Headers{
		CollectionID:         collectionID,
		DownloadServiceToken: downloadServiceAuthToken,
		ServiceToken:         serviceAuthToken,
		UserAccessToken:      "",
	}

	Convey("test filterable landing page", t, func() {
		Convey("test filterable landing page is successful, when it receives good dataset api responses", func() {
			mockDatasetClient.EXPECT().GetDataset(
				mockContext, headers, collectionID, datasetID,
			).Return(
				mockGetResponse, nil,
			)
			mockDatasetClient.EXPECT().GetVersions(
				mockContext, headers, datasetID, editionID, &getVersionsQueryParams,
			).Return(
				mockGetVersionsResponse, nil,
			)
			mockDatasetClient.EXPECT().GetVersion(
				mockContext, headers, datasetID, editionID, versionID,
			).Return(
				mockGetVersionsResponse.Items[0], nil,
			)
			mockDatasetClient.EXPECT().GetVersionDimensions(
				mockContext, headers, datasetID, editionID, versionID,
			).Return(
				mockGetVersionDimsResponse, nil,
			)
			mockDatasetClient.EXPECT().GetVersionDimensionOptions(
				mockContext, headers, datasetID, editionID, versionID, "aggregate",
				&dpDatasetApiSdk.QueryParams{Offset: 0, Limit: numOptsSummary},
			).Return(
				mockGetVersionDimsOptsResponse, nil,
			)
			mockZebedeeClient.EXPECT().GetHomepageContent(mockContext, userAuthToken, collectionID, locale, "/")
			mockZebedeeClient.EXPECT().GetBreadcrumb(mockContext, userAuthToken, collectionID, locale, "")
			mockRenderClient.EXPECT().NewBasePageModel().Return(
				coreModel.NewPage(mockConfig.PatternLibraryAssetsPath, mockConfig.SiteDomain),
			)
			// `BuildPage` should be called with the `dataset.DatasetDetails.Type` defining the template to be used
			mockRenderClient.EXPECT().BuildPage(gomock.Any(), gomock.Any(), datasetType)

			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest("GET", "/datasets/12345", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockDatasetClient, mockPopulationClient, mockRenderClient, mockZebedeeClient, mockTopicClient, mockConfig, ""))

			router.ServeHTTP(mockRequestWriter, mockRequest)

			So(mockRequestWriter.Code, ShouldEqual, http.StatusOK)
		})

		Convey("test filterableLanding returns 302 and redirects to the correct url for edition level requests without version", func() {
			mockZebedeeClient := NewMockZebedeeClient(mockController)
			mockDatasetClient.EXPECT().GetDataset(mockContext, headers, collectionID, "12345").Return(mockGetResponse, nil)
			versions := dpDatasetApiSdk.VersionsList{
				Items: []dpDatasetApiModels.Version{
					{
						Links: &dpDatasetApiModels.VersionLinks{
							Dataset: &dpDatasetApiModels.LinkObject{},
							Self: &dpDatasetApiModels.LinkObject{
								HRef: "/datasets/12345/editions/2016/versions/1",
							},
						},
						ReleaseDate: "02-01-2005",
					},
				},
			}
			mockDatasetClient.EXPECT().GetVersions(mockContext, headers, "12345", "5678", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/5678", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}", FilterableLanding(mockDatasetClient, mockPopulationClient, mockRenderClient, mockZebedeeClient, mockTopicClient, mockConfig, ""))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/datasets/12345/editions/5678/versions/1\">Found</a>.\n\n")
		})
	})
}

func TestFilterableLandingPageCantabularDataTypes(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockPc := NewMockPopulationClient(mockCtrl)
	mockTopicClient := NewMockTopicAPIClient(mockCtrl)

	defer mockCtrl.Finish()
	ctx := gomock.Any()
	cfg := initialiseMockConfig()
	headers := dpDatasetApiSdk.Headers{
		CollectionID:         collectionID,
		DownloadServiceToken: "",
		ServiceToken:         serviceAuthToken,
		UserAccessToken:      "",
	}
	mockGetDatasetResponse := dpDatasetApiModels.Dataset{
		Contacts: []dpDatasetApiModels.ContactDetails{
			{Name: "Nick"},
		},
		Type: "cantabular-table",
		URI:  "/economy/grossdomesticproduct/datasets/gdpjanuary2018",
		Links: &dpDatasetApiModels.DatasetLinks{
			LatestVersion: &dpDatasetApiModels.LinkObject{
				HRef: "/datasets/12345/editions/2021/versions/1",
			},
		},
		ID: "12345",
	}

	Convey("test census landing page", t, func() {
		mockClient := NewMockDatasetAPISdkClient(mockCtrl)
		mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
		mockZebedeeClient.EXPECT().GetHomepageContent(ctx, userAuthToken, collectionID, locale, "/")
		mockRend := NewMockRenderClient(mockCtrl)
		mockGetVersionDimensionOptionsResponse := dpDatasetApiSdk.VersionDimensionOptionsList{
			Items: []dpDatasetApiModels.PublicDimensionOption{
				{
					Label: "an option",
				},
			},
		}
		Convey("filterable landing handler returns census landing template for cantabular types", func() {
			mockConfig := config.Config{}
			mockClient.EXPECT().GetDataset(ctx, headers, collectionID, "12345").Return(mockGetDatasetResponse, nil)
			versions := dpDatasetApiSdk.VersionsList{
				Items: []dpDatasetApiModels.Version{
					{
						Dimensions: []dpDatasetApiModels.Dimension{
							{
								Name: "Dim name",
							},
						},
						Downloads: &dpDatasetApiModels.DownloadList{
							XLS: &dpDatasetApiModels.DownloadObject{
								Size: "78600",
								HRef: "https://www.my-url.com/file.xls",
							},
						},
						ReleaseDate: "02-01-2005",
						Version:     1,
						Links: &dpDatasetApiModels.VersionLinks{
							Dataset: &dpDatasetApiModels.LinkObject{},
							Self: &dpDatasetApiModels.LinkObject{
								HRef: "/datasets/12345/editions/2021/versions/1",
							},
						},
						IsBasedOn: &dpDatasetApiModels.IsBasedOn{ID: "UR"},
					},
				},
			}
			mockGetVersionDimensionsResponse := dpDatasetApiSdk.VersionDimensionsList{
				Items: versions.Items[0].Dimensions,
			}
			mockClient.EXPECT().GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
			mockClient.EXPECT().GetVersion(ctx, headers, "12345", "2021", "1").Return(versions.Items[0], nil)
			mockClient.EXPECT().GetVersionDimensionOptions(ctx, headers, "12345", "2021", "1", versions.Items[0].Dimensions[0].Name,
				&dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).Return(mockGetVersionDimensionOptionsResponse, nil)
			mockClient.EXPECT().GetVersionDimensions(ctx, headers, "12345", "2021", "1").Return(mockGetVersionDimensionsResponse, nil)
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
			req := httptest.NewRequest("GET", "/datasets/12345", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockPc, mockRend, mockZebedeeClient, mockTopicClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("census dataset landing page correctly fetches version 1 data for initial release date field, when loading a later version", func() {
			// Creating a custom response here as we want the latest version url to be version 2,
			// as there are 2 versions associated with the dataset
			mockConfig := config.Config{}
			mockGetDatasetResponseNew := dpDatasetApiModels.Dataset{
				Contacts: []dpDatasetApiModels.ContactDetails{
					{Name: "Nick"},
				},
				Type: "cantabular-table",
				URI:  "/economy/grossdomesticproduct/datasets/gdpjanuary2018",
				Links: &dpDatasetApiModels.DatasetLinks{
					LatestVersion: &dpDatasetApiModels.LinkObject{
						HRef: "/datasets/12345/editions/2021/versions/2",
					},
				},
				ID: "12345",
			}
			mockClient.EXPECT().GetDataset(ctx, headers, collectionID, "12345").Return(mockGetDatasetResponseNew, nil)
			mockGetVersionsResponse := dpDatasetApiSdk.VersionsList{
				Items: []dpDatasetApiModels.Version{
					{
						ReleaseDate: "02-01-2005",
						Version:     1,
						Links: &dpDatasetApiModels.VersionLinks{
							Dataset: &dpDatasetApiModels.LinkObject{},
							Self: &dpDatasetApiModels.LinkObject{
								HRef: "/datasets/12345/editions/2021/versions/1",
							},
						},
						Dimensions: []dpDatasetApiModels.Dimension{
							{
								Name: "Dim 1",
							},
						},
						IsBasedOn: &dpDatasetApiModels.IsBasedOn{
							ID: "UR",
						},
					},
					{
						ReleaseDate: "05-01-2005",
						Version:     2,
						Links: &dpDatasetApiModels.VersionLinks{
							Dataset: &dpDatasetApiModels.LinkObject{},
							Self: &dpDatasetApiModels.LinkObject{
								HRef: "/datasets/12345/editions/2021/versions/2",
							},
						},
						IsBasedOn: &dpDatasetApiModels.IsBasedOn{
							ID: "UR",
						},
					},
				},
			}
			// Version requested doesn't have any dimensions
			mockGetVersionDimensionsResponse := dpDatasetApiSdk.VersionDimensionsList{}
			mockClient.EXPECT().GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).Return(mockGetVersionsResponse, nil)
			mockClient.EXPECT().GetVersion(ctx, headers, "12345", "2021", "2").Return(mockGetVersionsResponse.Items[1], nil)
			mockClient.EXPECT().GetVersionDimensions(ctx, headers, "12345", "2021", "2").Return(mockGetVersionDimensionsResponse, nil)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockPc, mockRend, mockZebedeeClient, mockTopicClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("census dataset landing page returns 200 when no downloadable files provided", func() {
			mockConfig := config.Config{}
			mockClient.EXPECT().GetDataset(ctx, headers, collectionID, "12345").Return(mockGetDatasetResponse, nil)
			versions := dpDatasetApiSdk.VersionsList{
				Items: []dpDatasetApiModels.Version{
					{
						Downloads:   nil,
						ReleaseDate: "02-01-2005",
						Version:     1,
						Links: &dpDatasetApiModels.VersionLinks{
							Dataset: &dpDatasetApiModels.LinkObject{},
							Self: &dpDatasetApiModels.LinkObject{
								HRef: "/datasets/12345/editions/2021/versions/1",
							},
						},
						IsBasedOn: &dpDatasetApiModels.IsBasedOn{
							ID: "UR",
						},
					},
				},
			}
			// Version requested doesn't have any dimensions
			mockGetVersionDimensionsResponse := dpDatasetApiSdk.VersionDimensionsList{}
			mockClient.EXPECT().GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
			mockClient.EXPECT().GetVersion(ctx, headers, "12345", "2021", "1").Return(versions.Items[0], nil)
			mockClient.EXPECT().GetVersionDimensions(ctx, headers, "12345", "2021", "1").Return(mockGetVersionDimensionsResponse, nil)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockPc, mockRend, mockZebedeeClient, mockTopicClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("census dataset landing page returns 302 when valid download option chosen", func() {
			mockConfig := config.Config{}
			mockClient.EXPECT().GetDataset(ctx, headers, collectionID, "12345").Return(mockGetDatasetResponse, nil)
			versions := dpDatasetApiSdk.VersionsList{
				Items: []dpDatasetApiModels.Version{
					{
						Downloads: &dpDatasetApiModels.DownloadList{
							CSV: &dpDatasetApiModels.DownloadObject{
								Size: "1234",
								HRef: "https://a.domain.com/a-file.csv",
							},
						},
						ReleaseDate: "02-01-2005",
						Version:     1,
						Links: &dpDatasetApiModels.VersionLinks{
							Dataset: &dpDatasetApiModels.LinkObject{},
							Self: &dpDatasetApiModels.LinkObject{
								HRef: "/datasets/12345/editions/2021/versions/1",
							},
						},
						IsBasedOn: &dpDatasetApiModels.IsBasedOn{
							ID: "UR",
						},
					},
				},
			}
			// Version requested doesn't have any dimensions
			mockGetVersionDimensionsResponse := dpDatasetApiSdk.VersionDimensionsList{}
			mockClient.EXPECT().GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
			mockClient.EXPECT().GetVersion(ctx, headers, "12345", "2021", "1").Return(versions.Items[0], nil)
			mockClient.EXPECT().GetVersionDimensions(ctx, headers, "12345", "2021", "1").Return(mockGetVersionDimensionsResponse, nil)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")
			mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
				PaginationResponse: population.PaginationResponse{
					TotalCount: 2,
				},
			}, nil).AnyTimes()

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345?f=get-data&format=csv", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockPc, mockRend, mockZebedeeClient, mockTopicClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusFound)
		})

		Convey("census dataset landing page returns 200 when invalid download option chosen", func() {
			mockConfig := config.Config{}
			mockClient.EXPECT().GetDataset(ctx, headers, collectionID, "12345").Return(mockGetDatasetResponse, nil)
			versions := dpDatasetApiSdk.VersionsList{
				Items: []dpDatasetApiModels.Version{
					{
						Downloads:   nil,
						ReleaseDate: "02-01-2005",
						Version:     1,
						Links: &dpDatasetApiModels.VersionLinks{
							Dataset: &dpDatasetApiModels.LinkObject{},
							Self: &dpDatasetApiModels.LinkObject{
								HRef: "/datasets/12345/editions/2021/versions/1",
							},
						},
						IsBasedOn: &dpDatasetApiModels.IsBasedOn{
							ID: "UR",
						},
					},
				},
			}
			// Version requested doesn't have any dimensions
			mockGetVersionDimensionsResponse := dpDatasetApiSdk.VersionDimensionsList{}
			mockClient.EXPECT().GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
			mockClient.EXPECT().GetVersion(ctx, headers, "12345", "2021", "1").Return(versions.Items[0], nil)
			mockClient.EXPECT().GetVersionDimensions(ctx, headers, "12345", "2021", "1").Return(mockGetVersionDimensionsResponse, nil)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")
			mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
				PaginationResponse: population.PaginationResponse{
					TotalCount: 2,
				},
			}, nil).AnyTimes()

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345?f=get-data&format=aFormat", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockPc, mockRend, mockZebedeeClient, mockTopicClient, mockConfig, "/v1"))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("census dataset landing page returns 200 when unknown get query request made", func() {
			mockConfig := config.Config{}
			mockClient.EXPECT().GetDataset(ctx, headers, collectionID, "12345").Return(mockGetDatasetResponse, nil)
			versions := dpDatasetApiSdk.VersionsList{
				Items: []dpDatasetApiModels.Version{
					{
						Downloads:   nil,
						ReleaseDate: "02-01-2005",
						Version:     1,
						Links: &dpDatasetApiModels.VersionLinks{
							Dataset: &dpDatasetApiModels.LinkObject{},
							Self: &dpDatasetApiModels.LinkObject{
								HRef: "/datasets/12345/editions/2021/versions/1",
							},
						},
						IsBasedOn: &dpDatasetApiModels.IsBasedOn{
							ID: "UR",
						},
					},
				},
			}
			// Version requested doesn't have any dimensions
			mockGetVersionDimensionsResponse := dpDatasetApiSdk.VersionDimensionsList{}
			mockClient.EXPECT().GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
			mockClient.EXPECT().GetVersion(ctx, headers, "12345", "2021", "1").Return(versions.Items[0], nil)
			mockClient.EXPECT().GetVersionDimensions(ctx, headers, "12345", "2021", "1").Return(mockGetVersionDimensionsResponse, nil)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")
			mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
				PaginationResponse: population.PaginationResponse{
					TotalCount: 2,
				},
			}, nil).AnyTimes()

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345?f=blah-blah&format=bob", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}", FilterableLanding(mockClient, mockPc, mockRend, mockZebedeeClient, mockTopicClient, mockConfig, "/v1"))

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
	mockDatasetClient := NewMockDatasetAPISdkClient(mockController)
	mockPopulationClient := NewMockPopulationClient(mockController)
	mockRenderClient := NewMockRenderClient(mockController)
	mockTopicClient := NewMockTopicAPIClient(mockController)
	mockZebedeeClient := NewMockZebedeeClient(mockController)

	datasetID := "12345"
	datasetType := "static"
	downloadServiceAuthToken := ""
	getVersionsQueryParams := dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}
	editionID := "5678"
	versionID := "1"
	mockGetResponse := dpDatasetApiModels.Dataset{
		Contacts: []dpDatasetApiModels.ContactDetails{
			{Name: "Matt"},
		},
		URI: "/economy/grossdomesticproduct/datasets/gdpjanuary2018",
		Links: &dpDatasetApiModels.DatasetLinks{
			LatestVersion: &dpDatasetApiModels.LinkObject{
				HRef: "/datasets/1234/editions/5678/versions/2017",
			},
		},
		Type: datasetType,
		ID:   datasetID,
	}
	mockGetVersionsResponse := dpDatasetApiSdk.VersionsList{
		Items: []dpDatasetApiModels.Version{
			{
				Links: &dpDatasetApiModels.VersionLinks{
					Self: &dpDatasetApiModels.LinkObject{
						HRef: "/datasets/12345/editions/2016/versions/1",
					},
				},
				ReleaseDate: "02-01-2005",
				Version:     1,
			},
		},
	}
	mockGetVersionMetadataResponse := dpDatasetApiModels.Metadata{
		EditableMetadata: dpDatasetApiModels.EditableMetadata{
			Subtopics: []string{"123"},
		},
	}
	mockGetTopicPublicResponse := dpTopicApiModels.Topic{
		ID:          "123",
		Description: "Root Topic 1",
		Title:       "Root Topic 1",
		Slug:        "roottopic2",
		Keywords:    &[]string{"test"},
		State:       "published",
	}
	mockGetTopicPrivateResponse := dpTopicApiModels.TopicResponse{
		ID: "123",
		Current: &dpTopicApiModels.Topic{
			ID: "123",
		},
		Next: &dpTopicApiModels.Topic{},
	}
	headers := dpDatasetApiSdk.Headers{
		CollectionID:         collectionID,
		DownloadServiceToken: downloadServiceAuthToken,
		ServiceToken:         serviceAuthToken,
		UserAccessToken:      userAuthToken,
	}
	topicHeaders := dpTopicApiSdk.Headers{
		ServiceAuthToken: serviceAuthToken,
		UserAuthToken:    "",
	}

	Convey("test filterable landing page", t, func() {
		Convey("test filterable landing page is successful, when it receives good dataset api responses", func() {
			mockDatasetClient.EXPECT().GetDataset(
				mockContext, headers, collectionID, datasetID,
			).Return(
				mockGetResponse, nil,
			)
			mockDatasetClient.EXPECT().GetVersions(
				mockContext, headers, datasetID, editionID, &getVersionsQueryParams,
			).Return(
				mockGetVersionsResponse, nil,
			)
			mockDatasetClient.EXPECT().GetVersion(
				mockContext, headers, datasetID, editionID, versionID,
			).Return(
				mockGetVersionsResponse.Items[0], nil,
			)
			mockDatasetClient.EXPECT().GetVersionMetadata(mockContext, headers, datasetID, editionID, versionID).Return(
				mockGetVersionMetadataResponse, nil,
			)
			mockTopicClient.EXPECT().GetTopicPublic(mockContext, topicHeaders, "123").Return(
				&mockGetTopicPublicResponse, nil,
			)
			mockZebedeeClient.EXPECT().GetHomepageContent(mockContext, userAuthToken, collectionID, locale, "/")
			mockRenderClient.EXPECT().NewBasePageModel().Return(
				coreModel.NewPage(mockConfig.PatternLibraryAssetsPath, mockConfig.SiteDomain),
			)
			// `BuildPage` should be called with the `dataset.DatasetDetails.Type` defining the template to be used
			mockRenderClient.EXPECT().BuildPage(gomock.Any(), gomock.Any(), datasetType)

			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest("GET", fmt.Sprintf("/datasets/%s/editions/%s/versions/%s", datasetID, editionID, versionID), http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}", FilterableLanding(mockDatasetClient, mockPopulationClient, mockRenderClient, mockZebedeeClient, mockTopicClient, mockConfig, ""))

			router.ServeHTTP(mockRequestWriter, mockRequest)

			So(mockRequestWriter.Code, ShouldEqual, http.StatusOK)
		})

		Convey("test filterableLanding returns 302 and redirects to the correct url for edition level requests without version", func() {
			mockDatasetClient.EXPECT().GetDataset(mockContext, headers, collectionID, datasetID).Return(mockGetResponse, nil)
			mockDatasetClient.EXPECT().GetVersions(mockContext, headers, datasetID, editionID, &getVersionsQueryParams).Return(mockGetVersionsResponse, nil)

			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest("GET", fmt.Sprintf("/datasets/%s/editions/%s", datasetID, editionID), http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}", FilterableLanding(mockDatasetClient, mockPopulationClient, mockRenderClient, mockZebedeeClient, mockTopicClient, mockConfig, ""))

			router.ServeHTTP(mockRequestWriter, mockRequest)

			So(mockRequestWriter.Code, ShouldEqual, http.StatusFound)
			So(mockRequestWriter.Body.String(), ShouldEqual, "<a href=\"/datasets/12345/editions/5678/versions/1\">Found</a>.\n\n")
		})

		Convey("test filterable landing page is successful, when GetVersionMetadata retrieved is an empty object", func() {
			mockDatasetClient.EXPECT().GetDataset(
				mockContext, headers, collectionID, datasetID,
			).Return(
				mockGetResponse, nil,
			)
			mockDatasetClient.EXPECT().GetVersions(
				mockContext, headers, datasetID, editionID, &getVersionsQueryParams,
			).Return(
				mockGetVersionsResponse, nil,
			)
			mockDatasetClient.EXPECT().GetVersion(
				mockContext, headers, datasetID, editionID, versionID,
			).Return(
				mockGetVersionsResponse.Items[0], nil,
			)

			// Returns empty metadata object with no subtopics
			mockDatasetClient.EXPECT().GetVersionMetadata(mockContext, headers, datasetID, editionID, versionID).Return(
				dpDatasetApiModels.Metadata{}, nil,
			)
			mockZebedeeClient.EXPECT().GetHomepageContent(mockContext, userAuthToken, collectionID, locale, "/")
			mockRenderClient.EXPECT().NewBasePageModel().Return(
				coreModel.NewPage(mockConfig.PatternLibraryAssetsPath, mockConfig.SiteDomain),
			)
			// `BuildPage` should be called with the `dataset.DatasetDetails.Type` defining the template to be used
			mockRenderClient.EXPECT().BuildPage(gomock.Any(), gomock.Any(), datasetType)

			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest("GET", fmt.Sprintf("/datasets/%s/editions/%s/versions/%s", datasetID, editionID, versionID), http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}", FilterableLanding(mockDatasetClient, mockPopulationClient, mockRenderClient, mockZebedeeClient, mockTopicClient, mockConfig, ""))

			router.ServeHTTP(mockRequestWriter, mockRequest)

			So(mockRequestWriter.Code, ShouldEqual, http.StatusOK)
		})

		Convey("test filterable landing page logs correct warning when GetTopicPublic is called with invalid topicID ", func() {
			var buf bytes.Buffer
			var fbBuf bytes.Buffer
			dpLogger.SetDestination(&buf, &fbBuf)

			testTopicAPIError := dpTopicApiErrors.StatusError{
				Code: 404,
				Err:  errors.New("test error"),
			}

			mockDatasetClient.EXPECT().GetDataset(
				mockContext, headers, collectionID, datasetID,
			).Return(
				mockGetResponse, nil,
			)
			mockDatasetClient.EXPECT().GetVersions(
				mockContext, headers, datasetID, editionID, &getVersionsQueryParams,
			).Return(
				mockGetVersionsResponse, nil,
			)
			mockDatasetClient.EXPECT().GetVersion(
				mockContext, headers, datasetID, editionID, versionID,
			).Return(
				mockGetVersionsResponse.Items[0], nil,
			)
			mockDatasetClient.EXPECT().GetVersionMetadata(mockContext, headers, datasetID, editionID, versionID).Return(
				mockGetVersionMetadataResponse, nil,
			)
			mockTopicClient.EXPECT().GetTopicPublic(mockContext, topicHeaders, "123").Return(
				nil, testTopicAPIError,
			)
			mockZebedeeClient.EXPECT().GetHomepageContent(mockContext, userAuthToken, collectionID, locale, "/")
			mockRenderClient.EXPECT().NewBasePageModel().Return(
				coreModel.NewPage(mockConfig.PatternLibraryAssetsPath, mockConfig.SiteDomain),
			)
			// `BuildPage` should be called with the `dataset.DatasetDetails.Type` defining the template to be used
			mockRenderClient.EXPECT().BuildPage(gomock.Any(), gomock.Any(), datasetType)

			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest("GET", fmt.Sprintf("/datasets/%s/editions/%s/versions/%s", datasetID, editionID, versionID), http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}", FilterableLanding(mockDatasetClient, mockPopulationClient, mockRenderClient, mockZebedeeClient, mockTopicClient, mockConfig, ""))

			router.ServeHTTP(mockRequestWriter, mockRequest)
			So(buf.String(), ShouldContainSubstring, "unable to get topic data for topic ID: 123")
		})

		Convey("test filterable landing page successfully builds when one of two topic-api calls fails", func() {
			testTopicAPIError := dpTopicApiErrors.StatusError{
				Code: 404,
				Err:  errors.New("test error"),
			}

			mockGetVersionMetadataResponse := dpDatasetApiModels.Metadata{
				EditableMetadata: dpDatasetApiModels.EditableMetadata{
					Subtopics: []string{"123", "456"},
				},
			}
			mockDatasetClient.EXPECT().GetDataset(
				mockContext, headers, collectionID, datasetID,
			).Return(
				mockGetResponse, nil,
			)
			mockDatasetClient.EXPECT().GetVersions(
				mockContext, headers, datasetID, editionID, &getVersionsQueryParams,
			).Return(
				mockGetVersionsResponse, nil,
			)
			mockDatasetClient.EXPECT().GetVersion(
				mockContext, headers, datasetID, editionID, versionID,
			).Return(
				mockGetVersionsResponse.Items[0], nil,
			)
			mockDatasetClient.EXPECT().GetVersionMetadata(mockContext, headers, datasetID, editionID, versionID).Return(
				mockGetVersionMetadataResponse, nil,
			)
			mockTopicClient.EXPECT().GetTopicPublic(mockContext, topicHeaders, "123").Return(
				nil, testTopicAPIError,
			)
			mockTopicClient.EXPECT().GetTopicPublic(mockContext, topicHeaders, "456").Return(
				&mockGetTopicPublicResponse, nil,
			)
			mockZebedeeClient.EXPECT().GetHomepageContent(mockContext, userAuthToken, collectionID, locale, "/")
			mockRenderClient.EXPECT().NewBasePageModel().Return(
				coreModel.NewPage(mockConfig.PatternLibraryAssetsPath, mockConfig.SiteDomain),
			)
			// `BuildPage` should be called with the `dataset.DatasetDetails.Type` defining the template to be used
			mockRenderClient.EXPECT().BuildPage(gomock.Any(), gomock.Any(), datasetType)

			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest("GET", fmt.Sprintf("/datasets/%s/editions/%s/versions/%s", datasetID, editionID, versionID), http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}", FilterableLanding(mockDatasetClient, mockPopulationClient, mockRenderClient, mockZebedeeClient, mockTopicClient, mockConfig, ""))

			router.ServeHTTP(mockRequestWriter, mockRequest)
			So(mockRequestWriter.Code, ShouldEqual, http.StatusOK)
		})
		Convey("Given IsPublishing flag is enabled", func() {
			mockConfig.IsPublishing = true

			Convey("When GetTopicPrivate is called with invalid topicID", func() {
				var buf bytes.Buffer
				var fbBuf bytes.Buffer
				dpLogger.SetDestination(&buf, &fbBuf)

				testTopicAPIError := dpTopicApiErrors.StatusError{
					Code: 404,
					Err:  errors.New("test error"),
				}

				mockDatasetClient.EXPECT().GetDataset(
					mockContext, headers, collectionID, datasetID,
				).Return(
					mockGetResponse, nil,
				)
				mockDatasetClient.EXPECT().GetVersions(
					mockContext, headers, datasetID, editionID, &getVersionsQueryParams,
				).Return(
					mockGetVersionsResponse, nil,
				)
				mockDatasetClient.EXPECT().GetVersion(
					mockContext, headers, datasetID, editionID, versionID,
				).Return(
					mockGetVersionsResponse.Items[0], nil,
				)
				mockDatasetClient.EXPECT().GetVersionMetadata(mockContext, headers, datasetID, editionID, versionID).Return(
					mockGetVersionMetadataResponse, nil,
				)
				mockTopicClient.EXPECT().GetTopicPrivate(mockContext, topicHeaders, "123").Return(
					nil, testTopicAPIError,
				)
				mockZebedeeClient.EXPECT().GetHomepageContent(mockContext, userAuthToken, collectionID, locale, "/")
				mockRenderClient.EXPECT().NewBasePageModel().Return(
					coreModel.NewPage(mockConfig.PatternLibraryAssetsPath, mockConfig.SiteDomain),
				)
				// `BuildPage` should be called with the `dataset.DatasetDetails.Type` defining the template to be used
				mockRenderClient.EXPECT().BuildPage(gomock.Any(), gomock.Any(), datasetType)

				mockRequestWriter := httptest.NewRecorder()
				mockRequest := httptest.NewRequest("GET", fmt.Sprintf("/datasets/%s/editions/%s/versions/%s", datasetID, editionID, versionID), http.NoBody)

				router := mux.NewRouter()
				router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}", FilterableLanding(mockDatasetClient, mockPopulationClient, mockRenderClient, mockZebedeeClient, mockTopicClient, mockConfig, ""))

				router.ServeHTTP(mockRequestWriter, mockRequest)
				Convey("Then filterable landing page logs correct warning", func() {
					So(buf.String(), ShouldContainSubstring, "unable to get topic data for topic ID: 123")
				})
			})

			Convey("When one of two topic api calls fails", func() {
				testTopicAPIError := dpTopicApiErrors.StatusError{
					Code: 404,
					Err:  errors.New("test error"),
				}

				mockGetVersionMetadataResponse := dpDatasetApiModels.Metadata{
					EditableMetadata: dpDatasetApiModels.EditableMetadata{
						Subtopics: []string{"123", "456"},
					},
				}
				mockDatasetClient.EXPECT().GetDataset(
					mockContext, headers, collectionID, datasetID,
				).Return(
					mockGetResponse, nil,
				)
				mockDatasetClient.EXPECT().GetVersions(
					mockContext, headers, datasetID, editionID, &getVersionsQueryParams,
				).Return(
					mockGetVersionsResponse, nil,
				)
				mockDatasetClient.EXPECT().GetVersion(
					mockContext, headers, datasetID, editionID, versionID,
				).Return(
					mockGetVersionsResponse.Items[0], nil,
				)
				mockDatasetClient.EXPECT().GetVersionMetadata(mockContext, headers, datasetID, editionID, versionID).Return(
					mockGetVersionMetadataResponse, nil,
				)
				mockTopicClient.EXPECT().GetTopicPrivate(mockContext, topicHeaders, "123").Return(
					nil, testTopicAPIError,
				)
				mockTopicClient.EXPECT().GetTopicPrivate(mockContext, topicHeaders, "456").Return(
					&mockGetTopicPrivateResponse, nil,
				)
				mockZebedeeClient.EXPECT().GetHomepageContent(mockContext, userAuthToken, collectionID, locale, "/")
				mockRenderClient.EXPECT().NewBasePageModel().Return(
					coreModel.NewPage(mockConfig.PatternLibraryAssetsPath, mockConfig.SiteDomain),
				)
				// `BuildPage` should be called with the `dataset.DatasetDetails.Type` defining the template to be used
				mockRenderClient.EXPECT().BuildPage(gomock.Any(), gomock.Any(), datasetType)

				mockRequestWriter := httptest.NewRecorder()
				mockRequest := httptest.NewRequest("GET", fmt.Sprintf("/datasets/%s/editions/%s/versions/%s", datasetID, editionID, versionID), http.NoBody)

				router := mux.NewRouter()
				router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}", FilterableLanding(mockDatasetClient, mockPopulationClient, mockRenderClient, mockZebedeeClient, mockTopicClient, mockConfig, ""))

				router.ServeHTTP(mockRequestWriter, mockRequest)

				Convey("Then the filterable landing page builds correctly, with status 200", func() {
					So(mockRequestWriter.Code, ShouldEqual, http.StatusOK)
				})
			})
		})
	})
}
