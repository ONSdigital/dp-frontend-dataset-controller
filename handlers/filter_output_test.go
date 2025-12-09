package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	core "github.com/ONSdigital/dis-design-system-go/model"
	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFilterOutputHandler(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()
	cfg := config.Config{}
	versions := dpDatasetApiSdk.VersionsList{
		Items: []dpDatasetApiModels.Version{
			{
				Downloads: &dpDatasetApiModels.DownloadList{
					XLS: &dpDatasetApiModels.DownloadObject{
						Size: "78600",
						HRef: "https://www.my-url.com/file.xls",
					},
				},
				ReleaseDate: "02-01-2005",
				Version:     1,
				Links: &dpDatasetApiModels.VersionLinks{
					Self: &dpDatasetApiModels.LinkObject{
						HRef: "/datasets/12345/editions/2021/versions/1",
					},
				},
			},
		},
	}
	mockDimensionCategories := []population.DimensionCategory{
		{
			Categories: []population.DimensionCategoryItem{
				{
					Label: "an option",
				},
			},
		},
		{
			Categories: []population.DimensionCategoryItem{},
		},
	}

	filterModels := []filter.Model{
		{
			Dimensions: []filter.ModelDimension{
				{
					Name:       "Dim 1",
					IsAreaType: toBoolPtr(false),
				},
			},
		},
		{
			Dimensions: []filter.ModelDimension{
				{
					Name:       "Dim 2",
					IsAreaType: toBoolPtr(true),
				},
			},
			Downloads: nil,
		},
		{
			Dimensions: []filter.ModelDimension{
				{
					Name:       "Dim 3",
					IsAreaType: toBoolPtr(true),
				},
			},
			Downloads: map[string]filter.Download{
				"CSV": {
					Size: "1234",
					URL:  "https://a.domain.com/a-file.csv",
				},
			},
		},
		{
			Dimensions: []filter.ModelDimension{
				{
					Name:           "Dim 4",
					IsAreaType:     toBoolPtr(true),
					FilterByParent: "country",
				},
			},
			Downloads: nil,
		},
	}
	mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
	headers := dpDatasetApiSdk.Headers{
		CollectionID:         collectionID,
		DownloadServiceToken: "",
		AccessToken:          serviceAuthToken,
	}
	mockGetDatsetResponse := dpDatasetApiModels.Dataset{
		Contacts: []dpDatasetApiModels.ContactDetails{
			{Name: "Nick"},
		},
		Type: "flexible",
		URI:  "/economy/grossdomesticproduct/datasets/gdpjanuary2018",
		Links: &dpDatasetApiModels.DatasetLinks{
			LatestVersion: &dpDatasetApiModels.LinkObject{
				HRef: "/datasets/12345/editions/2021/versions/1",
			},
		},
		ID: "12345",
	}

	Convey("Given the FilterOutput handler", t, func() {
		Convey("When it receives good dataset api responses", func() {
			hp := zebedee.HomepageContent{}
			mockZebedeeClient.
				EXPECT().
				GetHomepageContent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(hp, nil).AnyTimes()

			mockDc := NewMockDatasetAPISdkClient(mockCtrl)
			mockDc.
				EXPECT().
				GetDataset(ctx, headers, "12345").
				Return(mockGetDatsetResponse, nil)
			mockDc.
				EXPECT().
				GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, nil)
			mockDc.
				EXPECT().
				GetVersion(ctx, headers, "12345", "2021", "1").
				Return(versions.Items[0], nil)

			mockFc := NewMockFilterClient(mockCtrl)
			mockFc.
				EXPECT().
				GetOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filterModels[1], nil)
			mockFc.
				EXPECT().
				GetDimensionOptions(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filter.DimensionOptions{}, "", nil)

			mockPc := NewMockPopulationClient(mockCtrl)
			mockPc.
				EXPECT().
				GetAreas(ctx, gomock.Any()).
				Return(population.GetAreasResponse{}, nil)
			mockPc.
				EXPECT().
				GetDimensionsDescription(ctx, gomock.Any()).
				Return(population.GetDimensionsResponse{}, nil)
			mockPc.EXPECT().GetDimensionCategories(ctx, gomock.Any()).
				Return(population.GetDimensionCategoriesResponse{
					PaginationResponse: population.PaginationResponse{TotalCount: 1},
					Categories:         mockDimensionCategories,
				}, nil).AnyTimes()
			mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
				PaginationResponse: population.PaginationResponse{
					TotalCount: 2,
				},
			}, nil).AnyTimes()
			mockPc.
				EXPECT().
				GetPopulationType(gomock.Any(), gomock.Any()).
				Return(population.GetPopulationTypeResponse{}, nil)

			mockRend := NewMockRenderClient(mockCtrl)
			mockRend.
				EXPECT().
				NewBasePageModel().
				Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.
				EXPECT().
				BuildPage(gomock.Any(), gomock.Any(), "census-landing")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)

			Convey("Then the status code is 200", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("When downloads are nil", func() {
			mockDc := NewMockDatasetAPISdkClient(mockCtrl)
			mockDc.
				EXPECT().
				GetDataset(ctx, headers, "12345").
				Return(mockGetDatsetResponse, nil)

			mockDc.
				EXPECT().
				GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, nil)
			mockDc.
				EXPECT().
				GetVersion(ctx, headers, "12345", "2021", "1").
				Return(versions.Items[0], nil)

			mockFc := NewMockFilterClient(mockCtrl)
			mockFc.
				EXPECT().
				GetOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filterModels[1], nil)
			mockFc.
				EXPECT().
				GetDimensionOptions(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filter.DimensionOptions{}, "", nil)

			mockPc := NewMockPopulationClient(mockCtrl)
			mockPc.
				EXPECT().
				GetAreas(ctx, gomock.Any()).
				Return(population.GetAreasResponse{}, nil)
			mockPc.
				EXPECT().
				GetDimensionsDescription(ctx, gomock.Any()).
				Return(population.GetDimensionsResponse{}, nil)
			mockPc.EXPECT().GetDimensionCategories(ctx, gomock.Any()).
				Return(population.GetDimensionCategoriesResponse{
					PaginationResponse: population.PaginationResponse{TotalCount: 1},
					Categories:         mockDimensionCategories,
				}, nil).AnyTimes()
			mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
				PaginationResponse: population.PaginationResponse{
					TotalCount: 2,
				},
			}, nil).AnyTimes()
			mockPc.
				EXPECT().
				GetPopulationType(gomock.Any(), gomock.Any()).
				Return(population.GetPopulationTypeResponse{}, nil)

			mockRend := NewMockRenderClient(mockCtrl)
			mockRend.
				EXPECT().NewBasePageModel().
				Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.
				EXPECT().
				BuildPage(gomock.Any(), gomock.Any(), "census-landing")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)

			Convey("Then the status code is 200", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("When valid download chosen", func() {
			mockDc := NewMockDatasetAPISdkClient(mockCtrl)
			mockDc.
				EXPECT().
				GetDataset(ctx, headers, "12345").
				Return(mockGetDatsetResponse, nil)
			mockDc.
				EXPECT().
				GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, nil)
			mockDc.
				EXPECT().
				GetVersion(ctx, headers, "12345", "2021", "1").
				Return(versions.Items[0], nil)

			mockFc := NewMockFilterClient(mockCtrl)
			mockFc.
				EXPECT().
				GetOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filterModels[2], nil)
			mockFc.
				EXPECT().
				GetDimensionOptions(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filter.DimensionOptions{}, "", nil)

			mockPc := NewMockPopulationClient(mockCtrl)
			mockPc.
				EXPECT().
				GetAreas(ctx, gomock.Any()).
				Return(population.GetAreasResponse{}, nil)
			mockPc.
				EXPECT().
				GetDimensionsDescription(ctx, gomock.Any()).
				Return(population.GetDimensionsResponse{}, nil)
			mockPc.EXPECT().GetDimensionCategories(ctx, gomock.Any()).
				Return(population.GetDimensionCategoriesResponse{
					PaginationResponse: population.PaginationResponse{TotalCount: 1},
					Categories:         mockDimensionCategories,
				}, nil).AnyTimes()
			mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
				PaginationResponse: population.PaginationResponse{
					TotalCount: 2,
				},
			}, nil).AnyTimes()
			mockPc.
				EXPECT().
				GetPopulationType(gomock.Any(), gomock.Any()).
				Return(population.GetPopulationTypeResponse{}, nil)

			mockRend := NewMockRenderClient(mockCtrl)
			mockRend.EXPECT().NewBasePageModel().Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890?f=get-data&format=csv", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)

			Convey("Then the status code is 302", func() {
				So(w.Code, ShouldEqual, http.StatusFound)
			})
		})

		Convey("When invalid download option chosen", func() {
			mockDc := NewMockDatasetAPISdkClient(mockCtrl)
			mockDc.
				EXPECT().
				GetDataset(ctx, headers, "12345").
				Return(mockGetDatsetResponse, nil)
			mockDc.
				EXPECT().
				GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, nil)
			mockDc.
				EXPECT().
				GetVersion(ctx, headers, "12345", "2021", "1").
				Return(versions.Items[0], nil)

			mockFc := NewMockFilterClient(mockCtrl)
			mockFc.
				EXPECT().
				GetOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filterModels[2], nil)
			mockFc.
				EXPECT().
				GetDimensionOptions(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filter.DimensionOptions{}, "", nil)

			mockPc := NewMockPopulationClient(mockCtrl)
			mockPc.
				EXPECT().
				GetAreas(ctx, gomock.Any()).
				Return(population.GetAreasResponse{}, nil)
			mockPc.
				EXPECT().
				GetDimensionsDescription(ctx, gomock.Any()).
				Return(population.GetDimensionsResponse{}, nil)
			mockPc.EXPECT().GetDimensionCategories(ctx, gomock.Any()).
				Return(population.GetDimensionCategoriesResponse{
					PaginationResponse: population.PaginationResponse{TotalCount: 1},
					Categories:         mockDimensionCategories,
				}, nil).AnyTimes()
			mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
				PaginationResponse: population.PaginationResponse{
					TotalCount: 2,
				},
			}, nil).AnyTimes()
			mockPc.
				EXPECT().
				GetPopulationType(gomock.Any(), gomock.Any()).
				Return(population.GetPopulationTypeResponse{}, nil)

			mockRend := NewMockRenderClient(mockCtrl)
			mockRend.
				EXPECT().
				NewBasePageModel().
				Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.
				EXPECT().
				BuildPage(gomock.Any(), gomock.Any(), "census-landing")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890?f=get-data&format=doc", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)

			Convey("Then the status code is 200", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("When unknown query made", func() {
			mockDc := NewMockDatasetAPISdkClient(mockCtrl)
			mockDc.EXPECT().GetDataset(ctx, headers, "12345").
				Return(mockGetDatsetResponse, nil)

			mockDc.
				EXPECT().
				GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, nil)
			mockDc.
				EXPECT().
				GetVersion(ctx, headers, "12345", "2021", "1").
				Return(versions.Items[0], nil)

			mockFc := NewMockFilterClient(mockCtrl)
			mockFc.
				EXPECT().
				GetOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filterModels[2], nil)
			mockFc.
				EXPECT().
				GetDimensionOptions(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filter.DimensionOptions{}, "", nil)

			mockPc := NewMockPopulationClient(mockCtrl)
			mockPc.
				EXPECT().
				GetAreas(ctx, gomock.Any()).
				Return(population.GetAreasResponse{}, nil)
			mockPc.
				EXPECT().
				GetDimensionsDescription(ctx, gomock.Any()).
				Return(population.GetDimensionsResponse{}, nil)
			mockPc.EXPECT().GetDimensionCategories(ctx, gomock.Any()).
				Return(population.GetDimensionCategoriesResponse{
					PaginationResponse: population.PaginationResponse{TotalCount: 1},
					Categories:         mockDimensionCategories,
				}, nil).AnyTimes()
			mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
				PaginationResponse: population.PaginationResponse{
					TotalCount: 2,
				},
			}, nil).AnyTimes()
			mockPc.
				EXPECT().
				GetPopulationType(gomock.Any(), gomock.Any()).
				Return(population.GetPopulationTypeResponse{}, nil)

			mockRend := NewMockRenderClient(mockCtrl)
			mockRend.
				EXPECT().
				NewBasePageModel().
				Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.
				EXPECT().
				BuildPage(gomock.Any(), gomock.Any(), "census-landing")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890?f=bob", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)

			Convey("Then the status code is 200", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("Given a dimension is not an area type", func() {
			Convey("When the dc.GetOptions is called", func() {
				mockDc := NewMockDatasetAPISdkClient(mockCtrl)
				mockDc.EXPECT().GetDataset(ctx, headers, "12345").
					Return(mockGetDatsetResponse, nil)
				mockDc.EXPECT().GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
				mockDc.EXPECT().GetVersion(ctx, headers, "12345", "2021", "1").Return(versions.Items[0], nil)

				mockFc := NewMockFilterClient(mockCtrl)
				mockFc.EXPECT().GetOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(filterModels[0], nil)

				mockRend := NewMockRenderClient(mockCtrl)
				mockRend.EXPECT().NewBasePageModel().Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
				mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")

				mockPc := NewMockPopulationClient(mockCtrl)
				mockPc.
					EXPECT().
					GetDimensionsDescription(ctx, gomock.Any()).
					Return(population.GetDimensionsResponse{}, nil)
				mockPc.EXPECT().GetDimensionCategories(ctx, gomock.Any()).
					Return(population.GetDimensionCategoriesResponse{
						PaginationResponse: population.PaginationResponse{TotalCount: 1},
						Categories:         mockDimensionCategories,
					}, nil).AnyTimes()
				mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
					PaginationResponse: population.PaginationResponse{
						TotalCount: 2,
					},
				}, nil).AnyTimes()
				mockPc.
					EXPECT().
					GetPopulationType(gomock.Any(), gomock.Any()).
					Return(population.GetPopulationTypeResponse{}, nil)

				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", http.NoBody)

				router := mux.NewRouter()
				router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

				router.ServeHTTP(w, req)
				Convey("Then the status code is 200", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
				})
			})
		})

		Convey("Given a dimension is an area type", func() {
			Convey("When the pc.GetAreas is called", func() {
				mockDc := NewMockDatasetAPISdkClient(mockCtrl)
				mockDc.
					EXPECT().
					GetDataset(ctx, headers, "12345").
					Return(mockGetDatsetResponse, nil)
				mockDc.
					EXPECT().
					GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).
					Return(versions, nil)
				mockDc.
					EXPECT().
					GetVersion(ctx, headers, "12345", "2021", "1").
					Return(versions.Items[0], nil)

				mockFc := NewMockFilterClient(mockCtrl)
				mockFc.
					EXPECT().
					GetOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(filterModels[1], nil)
				mockFc.
					EXPECT().
					GetDimensionOptions(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(filter.DimensionOptions{}, "", nil)

				mockPc := NewMockPopulationClient(mockCtrl)
				mockPc.
					EXPECT().
					GetAreas(gomock.Any(), gomock.Any()).
					Return(population.GetAreasResponse{}, nil)
				mockPc.
					EXPECT().
					GetDimensionsDescription(ctx, gomock.Any()).
					Return(population.GetDimensionsResponse{}, nil)
				mockPc.
					EXPECT().
					GetDimensionCategories(ctx, gomock.Any()).
					Return(population.GetDimensionCategoriesResponse{
						PaginationResponse: population.PaginationResponse{TotalCount: 1},
						Categories:         mockDimensionCategories,
					}, nil).AnyTimes()
				mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
					PaginationResponse: population.PaginationResponse{
						TotalCount: 2,
					},
				}, nil).AnyTimes()
				mockPc.
					EXPECT().
					GetPopulationType(gomock.Any(), gomock.Any()).
					Return(population.GetPopulationTypeResponse{}, nil)

				mockRend := NewMockRenderClient(mockCtrl)
				mockRend.
					EXPECT().
					NewBasePageModel().
					Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
				mockRend.
					EXPECT().
					BuildPage(gomock.Any(), gomock.Any(), "census-landing")

				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", http.NoBody)

				router := mux.NewRouter()
				router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

				router.ServeHTTP(w, req)
				Convey("Then the status code is 200", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
				})
			})

			Convey("When the fc.GetDimensionOptions is called", func() {
				Convey("and an additional call to pc.GetArea is made", func() {
					mockDc := NewMockDatasetAPISdkClient(mockCtrl)
					mockDc.
						EXPECT().
						GetDataset(ctx, headers, "12345").
						Return(mockGetDatsetResponse, nil)
					mockDc.
						EXPECT().
						GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).
						Return(versions, nil)
					mockDc.
						EXPECT().
						GetVersion(ctx, headers, "12345", "2021", "1").
						Return(versions.Items[0], nil)

					mockFc := NewMockFilterClient(mockCtrl)
					mockFc.
						EXPECT().
						GetOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(filterModels[1], nil)
					mockFc.
						EXPECT().
						GetDimensionOptions(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(filter.DimensionOptions{
							Items: []filter.DimensionOption{
								{
									Option: "area 1",
								},
							},
							TotalCount: 1,
						}, "", nil)

					mockPc := NewMockPopulationClient(mockCtrl)
					mockPc.
						EXPECT().
						GetArea(gomock.Any(), gomock.Any()).
						Return(population.GetAreaResponse{}, nil)
					mockPc.
						EXPECT().
						GetDimensionsDescription(ctx, gomock.Any()).
						Return(population.GetDimensionsResponse{}, nil)
					mockPc.EXPECT().GetDimensionCategories(ctx, gomock.Any()).
						Return(population.GetDimensionCategoriesResponse{
							PaginationResponse: population.PaginationResponse{TotalCount: 1},
							Categories:         mockDimensionCategories,
						}, nil).AnyTimes()
					mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
						PaginationResponse: population.PaginationResponse{
							TotalCount: 2,
						},
					}, nil).AnyTimes()
					mockPc.
						EXPECT().
						GetPopulationType(gomock.Any(), gomock.Any()).
						Return(population.GetPopulationTypeResponse{}, nil)

					mockRend := NewMockRenderClient(mockCtrl)
					mockRend.
						EXPECT().
						NewBasePageModel().
						Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
					mockRend.
						EXPECT().
						BuildPage(gomock.Any(), gomock.Any(), "census-landing")

					w := httptest.NewRecorder()
					req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", http.NoBody)

					router := mux.NewRouter()
					router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

					router.ServeHTTP(w, req)
					Convey("Then the status code is 200", func() {
						So(w.Code, ShouldEqual, http.StatusOK)
					})
				})
			})

			Convey("When the fc.GetDimensionOptions is called with parent options", func() {
				Convey("and an additional call to pc.GetArea is made", func() {
					mockDc := NewMockDatasetAPISdkClient(mockCtrl)
					mockDc.
						EXPECT().
						GetDataset(ctx, headers, "12345").
						Return(mockGetDatsetResponse, nil)
					mockDc.
						EXPECT().
						GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).
						Return(versions, nil)
					mockDc.
						EXPECT().
						GetVersion(ctx, headers, "12345", "2021", "1").
						Return(versions.Items[0], nil)

					mockFc := NewMockFilterClient(mockCtrl)
					mockFc.
						EXPECT().
						GetOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(filterModels[3], nil)
					mockFc.
						EXPECT().
						GetDimensionOptions(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(filter.DimensionOptions{
							Items: []filter.DimensionOption{
								{
									Option: "area 1",
								},
							},
							TotalCount: 1,
						}, "", nil)

					mockPc := NewMockPopulationClient(mockCtrl)
					mockPc.
						EXPECT().
						GetArea(gomock.Any(), gomock.Any()).
						Return(population.GetAreaResponse{}, nil)
					mockPc.
						EXPECT().
						GetDimensionsDescription(ctx, gomock.Any()).
						Return(population.GetDimensionsResponse{}, nil)
					mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
						PaginationResponse: population.PaginationResponse{
							TotalCount: 2,
						},
					}, nil).AnyTimes()
					mockPc.
						EXPECT().
						GetPopulationType(gomock.Any(), gomock.Any()).
						Return(population.GetPopulationTypeResponse{}, nil)
					mockPc.EXPECT().GetDimensionCategories(ctx, gomock.Any()).
						Return(population.GetDimensionCategoriesResponse{
							PaginationResponse: population.PaginationResponse{TotalCount: 1},
							Categories:         mockDimensionCategories,
						}, nil).AnyTimes()

					mockRend := NewMockRenderClient(mockCtrl)
					mockRend.
						EXPECT().
						NewBasePageModel().
						Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
					mockRend.
						EXPECT().
						BuildPage(gomock.Any(), gomock.Any(), "census-landing")

					w := httptest.NewRecorder()
					req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", http.NoBody)

					router := mux.NewRouter()
					router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

					router.ServeHTTP(w, req)
					Convey("Then the status code is 200", func() {
						So(w.Code, ShouldEqual, http.StatusOK)
					})
				})
			})

			Convey("When the dataset is a multivariate", func() {
				Convey("Then an additional call to pc.GetBlockedAreaCount is made", func() {
					mockDc := NewMockDatasetAPISdkClient(mockCtrl)
					mockDc.
						EXPECT().
						GetDataset(ctx, headers, "12345").
						Return(dpDatasetApiModels.Dataset{
							Type: "multivariate",
						}, nil)
					mockDc.
						EXPECT().
						GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).
						Return(versions, nil)
					mockDc.
						EXPECT().
						GetVersion(ctx, headers, "12345", "2021", "1").
						Return(versions.Items[0], nil)

					mockFc := NewMockFilterClient(mockCtrl)
					mockFc.
						EXPECT().
						GetOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(filterModels[3], nil)
					mockFc.
						EXPECT().
						GetDimensionOptions(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(filter.DimensionOptions{
							Items: []filter.DimensionOption{
								{
									Option: "area 1",
								},
							},
							TotalCount: 1,
						}, "", nil)

					mockPc := NewMockPopulationClient(mockCtrl)
					mockPc.
						EXPECT().
						GetArea(gomock.Any(), gomock.Any()).
						Return(population.GetAreaResponse{}, nil)
					mockPc.
						EXPECT().
						GetDimensionsDescription(ctx, gomock.Any()).
						Return(population.GetDimensionsResponse{}, nil)
					mockPc.EXPECT().GetDimensionCategories(ctx, gomock.Any()).
						Return(population.GetDimensionCategoriesResponse{
							PaginationResponse: population.PaginationResponse{TotalCount: 1},
							Categories:         mockDimensionCategories,
						}, nil).AnyTimes()
					mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
						PaginationResponse: population.PaginationResponse{
							TotalCount: 2,
						},
					}, nil).AnyTimes()
					mockPc.
						EXPECT().
						GetBlockedAreaCount(ctx, gomock.Any()).
						Return(&cantabular.GetBlockedAreaCountResult{}, nil)
					mockPc.
						EXPECT().
						GetPopulationType(gomock.Any(), gomock.Any()).
						Return(population.GetPopulationTypeResponse{}, nil)

					mockRend := NewMockRenderClient(mockCtrl)
					mockRend.
						EXPECT().
						NewBasePageModel().
						Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
					mockRend.
						EXPECT().
						BuildPage(gomock.Any(), gomock.Any(), "census-landing")

					w := httptest.NewRecorder()
					req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", http.NoBody)

					router := mux.NewRouter()
					router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

					router.ServeHTTP(w, req)
					Convey("Then the status code is 200", func() {
						So(w.Code, ShouldEqual, http.StatusOK)
					})
				})
			})
		})

		Convey("When the dc.Get fails", func() {
			mockDc := NewMockDatasetAPISdkClient(mockCtrl)
			mockFc := NewMockFilterClient(mockCtrl)
			mockPc := NewMockPopulationClient(mockCtrl)
			mockRend := NewMockRenderClient(mockCtrl)

			mockDc.
				EXPECT().
				GetDataset(ctx, headers, "12345").
				Return(dpDatasetApiModels.Dataset{}, errors.New("dataset client error"))
			mockDc.
				EXPECT().
				GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, nil)
			mockDc.
				EXPECT().
				GetVersion(ctx, headers, "12345", "2021", "1").
				Return(versions.Items[0], nil)

			mockFc.
				EXPECT().
				GetOutput(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filter.Model{}, nil)

			mockPc.EXPECT().GetDimensionCategories(ctx, gomock.Any()).
				Return(population.GetDimensionCategoriesResponse{
					PaginationResponse: population.PaginationResponse{TotalCount: 1},
					Categories:         mockDimensionCategories,
				}, nil).AnyTimes()
			mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
				PaginationResponse: population.PaginationResponse{
					TotalCount: 2,
				},
			}, nil).AnyTimes()

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)
			Convey("Then the status code is 500", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When the dc.GetVersions fails", func() {
			mockDc := NewMockDatasetAPISdkClient(mockCtrl)
			mockFc := NewMockFilterClient(mockCtrl)
			mockPc := NewMockPopulationClient(mockCtrl)
			mockRend := NewMockRenderClient(mockCtrl)

			mockDc.
				EXPECT().
				GetDataset(ctx, headers, "12345").
				Return(dpDatasetApiModels.Dataset{}, nil)
			mockDc.
				EXPECT().
				GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, errors.New("dataset client error"))
			mockDc.
				EXPECT().
				GetVersion(ctx, headers, "12345", "2021", "1").
				Return(versions.Items[0], nil)

			mockFc.
				EXPECT().
				GetOutput(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filter.Model{}, nil)

			mockPc.EXPECT().GetDimensionCategories(ctx, gomock.Any()).
				Return(population.GetDimensionCategoriesResponse{
					PaginationResponse: population.PaginationResponse{TotalCount: 1},
					Categories:         mockDimensionCategories,
				}, nil).AnyTimes()
			mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
				PaginationResponse: population.PaginationResponse{
					TotalCount: 2,
				},
			}, nil).AnyTimes()

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)
			Convey("Then the status code is 500", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When the dc.GetVersion fails", func() {
			mockDc := NewMockDatasetAPISdkClient(mockCtrl)
			mockFc := NewMockFilterClient(mockCtrl)
			mockPc := NewMockPopulationClient(mockCtrl)
			mockRend := NewMockRenderClient(mockCtrl)

			mockDc.
				EXPECT().
				GetDataset(ctx, headers, "12345").
				Return(dpDatasetApiModels.Dataset{}, nil)
			mockDc.
				EXPECT().
				GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, nil)
			mockDc.
				EXPECT().
				GetVersion(ctx, headers, "12345", "2021", "1").
				Return(versions.Items[0], errors.New("dataset client error"))

			mockFc.
				EXPECT().
				GetOutput(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filter.Model{}, nil)

			mockPc.EXPECT().GetDimensionCategories(ctx, gomock.Any()).
				Return(population.GetDimensionCategoriesResponse{
					PaginationResponse: population.PaginationResponse{TotalCount: 1},
					Categories:         mockDimensionCategories,
				}, nil).AnyTimes()
			mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
				PaginationResponse: population.PaginationResponse{
					TotalCount: 2,
				},
			}, nil).AnyTimes()

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)
			Convey("Then the status code is 500", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When the fc.GetOutput fails", func() {
			mockDc := NewMockDatasetAPISdkClient(mockCtrl)
			mockFc := NewMockFilterClient(mockCtrl)
			mockPc := NewMockPopulationClient(mockCtrl)
			mockRend := NewMockRenderClient(mockCtrl)

			mockDc.
				EXPECT().
				GetDataset(ctx, headers, "12345").
				Return(dpDatasetApiModels.Dataset{}, nil)
			mockDc.
				EXPECT().
				GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, nil)
			mockDc.
				EXPECT().
				GetVersion(ctx, headers, "12345", "2021", "1").
				Return(versions.Items[0], nil)

			mockFc.
				EXPECT().
				GetOutput(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filter.Model{}, errors.New("filter client error"))

			mockPc.EXPECT().GetDimensionCategories(ctx, gomock.Any()).
				Return(population.GetDimensionCategoriesResponse{
					PaginationResponse: population.PaginationResponse{TotalCount: 1},
					Categories:         mockDimensionCategories,
				}, nil).AnyTimes()
			mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
				PaginationResponse: population.PaginationResponse{
					TotalCount: 2,
				},
			}, nil).AnyTimes()

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)
			Convey("Then the status code is 500", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When the pc.GetDimensionsDescription fails", func() {
			mockDc := NewMockDatasetAPISdkClient(mockCtrl)
			mockFc := NewMockFilterClient(mockCtrl)
			mockPc := NewMockPopulationClient(mockCtrl)
			mockRend := NewMockRenderClient(mockCtrl)

			mockDc.
				EXPECT().
				GetDataset(ctx, headers, "12345").
				Return(dpDatasetApiModels.Dataset{}, nil)
			mockDc.
				EXPECT().
				GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, nil)
			mockDc.
				EXPECT().
				GetVersion(ctx, headers, "12345", "2021", "1").
				Return(versions.Items[0], nil)

			mockFc.
				EXPECT().
				GetOutput(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filter.Model{}, nil)

			mockPc.
				EXPECT().
				GetDimensionsDescription(ctx, gomock.Any()).
				Return(population.GetDimensionsResponse{}, errors.New("Internal error"))
			mockPc.EXPECT().GetDimensionCategories(ctx, gomock.Any()).
				Return(population.GetDimensionCategoriesResponse{
					PaginationResponse: population.PaginationResponse{TotalCount: 1},
					Categories:         mockDimensionCategories,
				}, nil).AnyTimes()
			mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
				PaginationResponse: population.PaginationResponse{
					TotalCount: 2,
				},
			}, nil).AnyTimes()
			mockPc.
				EXPECT().
				GetPopulationType(gomock.Any(), gomock.Any()).
				Return(population.GetPopulationTypeResponse{}, nil)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)
			Convey("Then the status code is 500", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When the pc.GetBlockedAreaCount fails", func() {
			mockDc := NewMockDatasetAPISdkClient(mockCtrl)
			mockFc := NewMockFilterClient(mockCtrl)
			mockPc := NewMockPopulationClient(mockCtrl)
			mockRend := NewMockRenderClient(mockCtrl)

			mockDc.
				EXPECT().
				GetDataset(ctx, headers, "12345").
				Return(dpDatasetApiModels.Dataset{
					Type: "multivariate",
				}, nil)
			mockDc.
				EXPECT().
				GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, nil)
			mockDc.
				EXPECT().
				GetVersion(ctx, headers, "12345", "2021", "1").
				Return(versions.Items[0], nil)

			mockFc.
				EXPECT().
				GetOutput(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filter.Model{}, nil)

			mockPc.
				EXPECT().
				GetDimensionsDescription(ctx, gomock.Any()).
				Return(population.GetDimensionsResponse{}, nil)
			mockPc.
				EXPECT().
				GetDimensionCategories(ctx, gomock.Any()).
				Return(population.GetDimensionCategoriesResponse{
					PaginationResponse: population.PaginationResponse{TotalCount: 1},
					Categories:         mockDimensionCategories,
				}, nil).AnyTimes()
			mockPc.
				EXPECT().
				GetBlockedAreaCount(ctx, gomock.Any()).
				Return(&cantabular.GetBlockedAreaCountResult{}, errors.New("Internal error"))
			mockPc.
				EXPECT().
				GetPopulationType(gomock.Any(), gomock.Any()).
				Return(population.GetPopulationTypeResponse{}, nil)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)
			Convey("Then the status code is 500", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When the pc.GetPopulationTypeResponse fails", func() {
			mockDc := NewMockDatasetAPISdkClient(mockCtrl)
			mockFc := NewMockFilterClient(mockCtrl)
			mockPc := NewMockPopulationClient(mockCtrl)
			mockRend := NewMockRenderClient(mockCtrl)

			mockDc.
				EXPECT().
				GetDataset(ctx, headers, "12345").
				Return(dpDatasetApiModels.Dataset{
					Type: "multivariate",
				}, nil)
			mockDc.
				EXPECT().
				GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, nil)
			mockDc.
				EXPECT().
				GetVersion(ctx, headers, "12345", "2021", "1").
				Return(versions.Items[0], nil)

			mockFc.
				EXPECT().
				GetOutput(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filter.Model{}, nil)
			mockPc.
				EXPECT().
				GetDimensionCategories(ctx, gomock.Any()).
				Return(population.GetDimensionCategoriesResponse{
					PaginationResponse: population.PaginationResponse{TotalCount: 1},
					Categories:         mockDimensionCategories,
				}, nil).AnyTimes()
			mockPc.
				EXPECT().
				GetDimensionsDescription(gomock.Any(), gomock.Any()).
				Return(population.GetDimensionsResponse{}, nil)
			mockPc.
				EXPECT().
				GetPopulationType(gomock.Any(), gomock.Any()).
				Return(population.GetPopulationTypeResponse{}, errors.New("Internal error"))

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)
			Convey("Then the status code is 500", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When the fc.GetDimensionOptions is called", func() {
			Convey("and the additional call to pc.GetArea fails", func() {
				mockDc := NewMockDatasetAPISdkClient(mockCtrl)
				mockDc.
					EXPECT().
					GetDataset(ctx, headers, "12345").
					Return(mockGetDatsetResponse, nil)
				mockDc.
					EXPECT().
					GetVersions(ctx, headers, "12345", "2021", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).
					Return(versions, nil)
				mockDc.
					EXPECT().
					GetVersion(ctx, headers, "12345", "2021", "1").
					Return(versions.Items[0], nil)

				mockFc := NewMockFilterClient(mockCtrl)
				mockFc.
					EXPECT().
					GetOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(filterModels[1], nil)
				mockFc.
					EXPECT().
					GetDimensionOptions(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(filter.DimensionOptions{
						Items: []filter.DimensionOption{
							{
								Option: "area 1",
							},
						},
						TotalCount: 1,
					}, "", nil)

				mockPc := NewMockPopulationClient(mockCtrl)
				mockPc.
					EXPECT().
					GetArea(gomock.Any(), gomock.Any()).
					Return(population.GetAreaResponse{}, errors.New("area client error"))
				mockPc.
					EXPECT().
					GetDimensionsDescription(ctx, gomock.Any()).
					Return(population.GetDimensionsResponse{}, nil)
				mockPc.EXPECT().GetDimensionCategories(ctx, gomock.Any()).
					Return(population.GetDimensionCategoriesResponse{
						PaginationResponse: population.PaginationResponse{TotalCount: 1},
						Categories:         mockDimensionCategories,
					}, nil).AnyTimes()
				mockPc.EXPECT().GetCategorisations(ctx, gomock.Any()).Return(population.GetCategorisationsResponse{
					PaginationResponse: population.PaginationResponse{
						TotalCount: 2,
					},
				}, nil).AnyTimes()
				mockPc.
					EXPECT().
					GetPopulationType(gomock.Any(), gomock.Any()).
					Return(population.GetPopulationTypeResponse{}, nil)

				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", http.NoBody)

				router := mux.NewRouter()
				router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, NewMockRenderClient(mockCtrl), cfg, ""))

				router.ServeHTTP(w, req)
				Convey("Then the status code is 500", func() {
					So(w.Code, ShouldEqual, http.StatusInternalServerError)
				})
			})
		})
	})
}

// toBoolPtr converts a boolean to a pointer
func toBoolPtr(val bool) *bool {
	return &val
}
