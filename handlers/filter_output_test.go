package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	coreModel "github.com/ONSdigital/dp-renderer/model"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFilterOutputHandler(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()
	cfg := config.Config{EnableCensusPages: true}
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

	Convey("Given the FilterOutput handler", t, func() {
		Convey("When it receives good dataset api responses", func() {
			hp := zebedee.HomepageContent{}
			mockZebedeeClient.
				EXPECT().
				GetHomepageContent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(hp, nil).AnyTimes()

			mockDc := NewMockDatasetClient(mockCtrl)
			mockDc.
				EXPECT().
				Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").
				Return(dataset.DatasetDetails{
					Contacts: &[]dataset.Contact{{Name: "Nick"}},
					Type:     "flexible",
					URI:      "/economy/grossdomesticproduct/datasets/gdpjanuary2018",
					Links: dataset.Links{
						LatestVersion: dataset.Link{
							URL: "/datasets/12345/editions/2021/versions/1",
						},
					},
					ID: "12345",
				}, nil)
			mockDc.
				EXPECT().
				GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, nil)
			mockDc.
				EXPECT().
				GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").
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

			mockRend := NewMockRenderClient(mockCtrl)
			mockRend.
				EXPECT().
				NewBasePageModel().
				Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.
				EXPECT().
				BuildPage(gomock.Any(), gomock.Any(), "census-landing")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)

			Convey("Then the status code is 200", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("When downloads are nil", func() {
			mockDc := NewMockDatasetClient(mockCtrl)
			mockDc.
				EXPECT().
				Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").
				Return(dataset.DatasetDetails{
					Contacts: &[]dataset.Contact{{Name: "Nick"}},
					Type:     "flexible",
					URI:      "/economy/grossdomesticproduct/datasets/gdpjanuary2018",
					Links: dataset.Links{
						LatestVersion: dataset.Link{
							URL: "/datasets/12345/editions/2021/versions/1",
						},
					},
					ID: "12345",
				}, nil)

			mockDc.
				EXPECT().
				GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, nil)
			mockDc.
				EXPECT().
				GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").
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

			mockRend := NewMockRenderClient(mockCtrl)
			mockRend.
				EXPECT().NewBasePageModel().
				Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.
				EXPECT().
				BuildPage(gomock.Any(), gomock.Any(), "census-landing")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)

			Convey("Then the status code is 200", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("When valid download chosen", func() {
			mockDc := NewMockDatasetClient(mockCtrl)
			mockDc.
				EXPECT().
				Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").
				Return(dataset.DatasetDetails{
					Contacts: &[]dataset.Contact{{Name: "Nick"}},
					Type:     "flexible",
					URI:      "/economy/grossdomesticproduct/datasets/gdpjanuary2018",
					Links: dataset.Links{
						LatestVersion: dataset.Link{
							URL: "/datasets/12345/editions/2021/versions/1",
						},
					},
					ID: "12345",
				}, nil)
			mockDc.
				EXPECT().
				GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, nil)
			mockDc.
				EXPECT().
				GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").
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

			mockRend := NewMockRenderClient(mockCtrl)
			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890?f=get-data&format=csv", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)

			Convey("Then the status code is 302", func() {
				So(w.Code, ShouldEqual, http.StatusFound)
			})
		})

		Convey("When invalid download option chosen", func() {
			mockDc := NewMockDatasetClient(mockCtrl)
			mockDc.
				EXPECT().
				Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").
				Return(dataset.DatasetDetails{
					Contacts: &[]dataset.Contact{{Name: "Nick"}},
					Type:     "flexible",
					URI:      "/economy/grossdomesticproduct/datasets/gdpjanuary2018",
					Links: dataset.Links{
						LatestVersion: dataset.Link{
							URL: "/datasets/12345/editions/2021/versions/1",
						},
					},
					ID: "12345",
				}, nil)
			mockDc.
				EXPECT().
				GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, nil)
			mockDc.
				EXPECT().
				GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").
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

			mockRend := NewMockRenderClient(mockCtrl)
			mockRend.
				EXPECT().
				NewBasePageModel().
				Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.
				EXPECT().
				BuildPage(gomock.Any(), gomock.Any(), "census-landing")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890?f=get-data&format=doc", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)

			Convey("Then the status code is 200", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("When unknown query made", func() {
			mockDc := NewMockDatasetClient(mockCtrl)
			mockDc.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").
				Return(dataset.DatasetDetails{
					Contacts: &[]dataset.Contact{{Name: "Nick"}},
					Type:     "flexible",
					URI:      "/economy/grossdomesticproduct/datasets/gdpjanuary2018",
					Links: dataset.Links{
						LatestVersion: dataset.Link{
							URL: "/datasets/12345/editions/2021/versions/1",
						},
					},
					ID: "12345",
				}, nil)

			mockDc.
				EXPECT().
				GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, nil)
			mockDc.
				EXPECT().
				GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").
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

			mockRend := NewMockRenderClient(mockCtrl)
			mockRend.
				EXPECT().
				NewBasePageModel().
				Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.
				EXPECT().
				BuildPage(gomock.Any(), gomock.Any(), "census-landing")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890?f=bob", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)

			Convey("Then the status code is 200", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("Given a dimension is not an area type", func() {
			Convey("When the dc.GetOptions is called", func() {
				mockDc := NewMockDatasetClient(mockCtrl)
				mockDc.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").
					Return(dataset.DatasetDetails{
						Contacts: &[]dataset.Contact{{Name: "Nick"}},
						Type:     "flexible",
						URI:      "/economy/grossdomesticproduct/datasets/gdpjanuary2018",
						Links: dataset.Links{
							LatestVersion: dataset.Link{
								URL: "/datasets/12345/editions/2021/versions/1",
							},
						},
						ID: "12345",
					}, nil)
				mockDc.EXPECT().GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).Return(versions, nil)
				mockDc.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").Return(versions.Items[0], nil)
				mockDc.EXPECT().GetOptions(ctx, gomock.Any(), gomock.Any(), gomock.Any(), "12345", "2021", "1", filterModels[0].Dimensions[0].Name,
					&dataset.QueryParams{Offset: 0, Limit: 1000}).Return(mockOpts[0], nil)

				mockFc := NewMockFilterClient(mockCtrl)
				mockFc.EXPECT().GetOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(filterModels[0], nil)

				mockRend := NewMockRenderClient(mockCtrl)
				mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
				mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "census-landing")

				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", nil)

				router := mux.NewRouter()
				router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, NewMockPopulationClient(mockCtrl), mockDc, mockRend, cfg, ""))

				router.ServeHTTP(w, req)
				Convey("Then the status code is 200", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
				})
			})
		})

		Convey("Given a dimension is an area type", func() {
			Convey("When the pc.GetAreas is called", func() {
				mockDc := NewMockDatasetClient(mockCtrl)
				mockDc.
					EXPECT().
					Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").
					Return(dataset.DatasetDetails{
						Contacts: &[]dataset.Contact{{Name: "Nick"}},
						Type:     "flexible",
						URI:      "/economy/grossdomesticproduct/datasets/gdpjanuary2018",
						Links: dataset.Links{
							LatestVersion: dataset.Link{
								URL: "/datasets/12345/editions/2021/versions/1",
							},
						},
						ID: "12345",
					}, nil)
				mockDc.
					EXPECT().
					GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).
					Return(versions, nil)
				mockDc.
					EXPECT().
					GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").
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

				mockRend := NewMockRenderClient(mockCtrl)
				mockRend.
					EXPECT().
					NewBasePageModel().
					Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
				mockRend.
					EXPECT().
					BuildPage(gomock.Any(), gomock.Any(), "census-landing")

				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", nil)

				router := mux.NewRouter()
				router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

				router.ServeHTTP(w, req)
				Convey("Then the status code is 200", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
				})
			})

			Convey("When the fc.GetDimensionOptions is called", func() {
				Convey("and an additional call to pc.GetArea is made", func() {
					mockDc := NewMockDatasetClient(mockCtrl)
					mockDc.
						EXPECT().
						Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").
						Return(dataset.DatasetDetails{
							Contacts: &[]dataset.Contact{{Name: "Nick"}},
							Type:     "flexible",
							URI:      "/economy/grossdomesticproduct/datasets/gdpjanuary2018",
							Links: dataset.Links{
								LatestVersion: dataset.Link{
									URL: "/datasets/12345/editions/2021/versions/1",
								},
							},
							ID: "12345",
						}, nil)
					mockDc.
						EXPECT().
						GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).
						Return(versions, nil)
					mockDc.
						EXPECT().
						GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").
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

					mockRend := NewMockRenderClient(mockCtrl)
					mockRend.
						EXPECT().
						NewBasePageModel().
						Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
					mockRend.
						EXPECT().
						BuildPage(gomock.Any(), gomock.Any(), "census-landing")

					w := httptest.NewRecorder()
					req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", nil)

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
					mockDc := NewMockDatasetClient(mockCtrl)
					mockDc.
						EXPECT().
						Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").
						Return(dataset.DatasetDetails{
							Contacts: &[]dataset.Contact{{Name: "Nick"}},
							Type:     "flexible",
							URI:      "/economy/grossdomesticproduct/datasets/gdpjanuary2018",
							Links: dataset.Links{
								LatestVersion: dataset.Link{
									URL: "/datasets/12345/editions/2021/versions/1",
								},
							},
							ID: "12345",
						}, nil)
					mockDc.
						EXPECT().
						GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).
						Return(versions, nil)
					mockDc.
						EXPECT().
						GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").
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
					// TODO: Hotfix to remove api call due to graphQL error
					// mockPc.
					// 	EXPECT().
					// 	GetParentAreaCount(gomock.Any(), gomock.Any()).
					// 	Return(0, nil)

					mockRend := NewMockRenderClient(mockCtrl)
					mockRend.
						EXPECT().
						NewBasePageModel().
						Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
					mockRend.
						EXPECT().
						BuildPage(gomock.Any(), gomock.Any(), "census-landing")

					w := httptest.NewRecorder()
					req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", nil)

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
			mockDc := NewMockDatasetClient(mockCtrl)
			mockFc := NewMockFilterClient(mockCtrl)
			mockPc := NewMockPopulationClient(mockCtrl)
			mockRend := NewMockRenderClient(mockCtrl)

			mockDc.
				EXPECT().
				Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").
				Return(dataset.DatasetDetails{}, errors.New("dataset client error"))
			mockDc.
				EXPECT().
				GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, nil)
			mockDc.
				EXPECT().
				GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").
				Return(versions.Items[0], nil)

			mockFc.
				EXPECT().
				GetOutput(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filter.Model{}, nil)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)
			Convey("Then the status code is 500", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When the dc.GetVersions fails", func() {
			mockDc := NewMockDatasetClient(mockCtrl)
			mockFc := NewMockFilterClient(mockCtrl)
			mockPc := NewMockPopulationClient(mockCtrl)
			mockRend := NewMockRenderClient(mockCtrl)

			mockDc.
				EXPECT().
				Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").
				Return(dataset.DatasetDetails{}, nil)
			mockDc.
				EXPECT().
				GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, errors.New("dataset client error"))
			mockDc.
				EXPECT().
				GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").
				Return(versions.Items[0], nil)

			mockFc.
				EXPECT().
				GetOutput(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filter.Model{}, nil)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)
			Convey("Then the status code is 500", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When the dc.GetVersion fails", func() {
			mockDc := NewMockDatasetClient(mockCtrl)
			mockFc := NewMockFilterClient(mockCtrl)
			mockPc := NewMockPopulationClient(mockCtrl)
			mockRend := NewMockRenderClient(mockCtrl)

			mockDc.
				EXPECT().
				Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").
				Return(dataset.DatasetDetails{}, nil)
			mockDc.
				EXPECT().
				GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, nil)
			mockDc.
				EXPECT().
				GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").
				Return(versions.Items[0], errors.New("dataset client error"))

			mockFc.
				EXPECT().
				GetOutput(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filter.Model{}, nil)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)
			Convey("Then the status code is 500", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When the fc.GetOutput fails", func() {
			mockDc := NewMockDatasetClient(mockCtrl)
			mockFc := NewMockFilterClient(mockCtrl)
			mockPc := NewMockPopulationClient(mockCtrl)
			mockRend := NewMockRenderClient(mockCtrl)

			mockDc.
				EXPECT().
				Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").
				Return(dataset.DatasetDetails{}, nil)
			mockDc.
				EXPECT().
				GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).
				Return(versions, nil)
			mockDc.
				EXPECT().
				GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").
				Return(versions.Items[0], nil)

			mockFc.
				EXPECT().
				GetOutput(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filter.Model{}, errors.New("filter client error"))

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, mockRend, cfg, ""))

			router.ServeHTTP(w, req)
			Convey("Then the status code is 500", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When the fc.GetDimensionOptions is called", func() {
			Convey("and the additional call to pc.GetArea fails", func() {
				mockDc := NewMockDatasetClient(mockCtrl)
				mockDc.
					EXPECT().
					Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").
					Return(dataset.DatasetDetails{
						Contacts: &[]dataset.Contact{{Name: "Nick"}},
						Type:     "flexible",
						URI:      "/economy/grossdomesticproduct/datasets/gdpjanuary2018",
						Links: dataset.Links{
							LatestVersion: dataset.Link{
								URL: "/datasets/12345/editions/2021/versions/1",
							},
						},
						ID: "12345",
					}, nil)
				mockDc.
					EXPECT().
					GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).
					Return(versions, nil)
				mockDc.
					EXPECT().
					GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").
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

				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", nil)

				router := mux.NewRouter()
				router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, NewMockRenderClient(mockCtrl), cfg, ""))

				router.ServeHTTP(w, req)
				Convey("Then the status code is 500", func() {
					So(w.Code, ShouldEqual, http.StatusInternalServerError)
				})
			})
			// TODO: Hotfix to remove api call due to graphQL error
			// Convey("and the additional call to pc.GetParentAreaCount fails", func() {
			// 	mockDc := NewMockDatasetClient(mockCtrl)
			// 	mockDc.
			// 		EXPECT().
			// 		Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").
			// 		Return(dataset.DatasetDetails{
			// 			Contacts: &[]dataset.Contact{{Name: "Nick"}},
			// 			Type:     "flexible",
			// 			URI:      "/economy/grossdomesticproduct/datasets/gdpjanuary2018",
			// 			Links: dataset.Links{
			// 				LatestVersion: dataset.Link{
			// 					URL: "/datasets/12345/editions/2021/versions/1",
			// 				},
			// 			},
			// 			ID: "12345",
			// 		}, nil)
			// 	mockDc.
			// 		EXPECT().
			// 		GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).
			// 		Return(versions, nil)
			// 	mockDc.
			// 		EXPECT().
			// 		GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").
			// 		Return(versions.Items[0], nil)

			// TODO: Hotfix to remove api call due to graphQL error
			// Convey("and the additional call to pc.GetParentAreaCount fails", func() {
			// 	mockDc := NewMockDatasetClient(mockCtrl)
			// 	mockDc.
			// 		EXPECT().
			// 		Get(ctx, userAuthToken, serviceAuthToken, collectionID, "12345").
			// 		Return(dataset.DatasetDetails{
			// 			Contacts: &[]dataset.Contact{{Name: "Nick"}},
			// 			Type:     "flexible",
			// 			URI:      "/economy/grossdomesticproduct/datasets/gdpjanuary2018",
			// 			Links: dataset.Links{
			// 				LatestVersion: dataset.Link{
			// 					URL: "/datasets/12345/editions/2021/versions/1",
			// 				},
			// 			},
			// 			ID: "12345",
			// 		}, nil)
			// 	mockDc.
			// 		EXPECT().
			// 		GetVersions(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", &dataset.QueryParams{Offset: 0, Limit: 1000}).
			// 		Return(versions, nil)
			// 	mockDc.
			// 		EXPECT().
			// 		GetVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "", "12345", "2021", "1").
			// 		Return(versions.Items[0], nil)

			// 	mockFc := NewMockFilterClient(mockCtrl)
			// 	mockFc.
			// 		EXPECT().
			// 		GetOutput(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			// 		Return(filterModels[3], nil)
			// 	mockFc.
			// 		EXPECT().
			// 		GetDimensionOptions(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			// 		Return(filter.DimensionOptions{
			// 			Items: []filter.DimensionOption{
			// 				{
			// 					Option: "area 1",
			// 				},
			// 			},
			// 			TotalCount: 1,
			// 		}, "", nil)

			// 	mockPc := NewMockPopulationClient(mockCtrl)
			// 	mockPc.
			// 		EXPECT().
			// 		GetArea(gomock.Any(), gomock.Any()).
			// 		Return(population.GetAreaResponse{}, nil)
			// 	mockPc.
			// 		EXPECT().
			// 		GetParentAreaCount(gomock.Any(), gomock.Any()).
			// 		Return(0, errors.New("area parent count client error"))

			// 	w := httptest.NewRecorder()
			// 	req := httptest.NewRequest("GET", "/datasets/12345/editions/2021/versions/1/filter-outputs/67890", nil)

			// 	router := mux.NewRouter()
			// 	router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", FilterOutput(mockZebedeeClient, mockFc, mockPc, mockDc, NewMockRenderClient(mockCtrl), cfg, ""))

			// 	router.ServeHTTP(w, req)
			// 	Convey("Then the status code is 500", func() {
			// 		So(w.Code, ShouldEqual, http.StatusInternalServerError)
			// 	})
			// })
		})
	})
}

// toBoolPtr converts a boolean to a pointer
func toBoolPtr(val bool) *bool {
	return &val
}
