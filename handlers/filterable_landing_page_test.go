package handlers

import (
	"errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	coreModel "github.com/ONSdigital/dp-renderer/model"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFilterableLandingPage(t *testing.T) {
	mockCtrl := gomock.NewController(t)
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
}
