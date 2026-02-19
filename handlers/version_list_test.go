package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	core "github.com/ONSdigital/dis-design-system-go/model"
	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestVersionList(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()
	cfg := initialiseMockConfig()
	headers := dpDatasetApiSdk.Headers{
		CollectionID:         collectionID,
		DownloadServiceToken: "",
		AccessToken:          serviceAuthToken,
	}

	Convey("test versions list", t, func() {
		Convey("test versions list returns 200 when rendered successfully", func() {
			mockClient := NewMockDatasetAPISdkClient(mockCtrl)
			mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
			mockZebedeeClient.EXPECT().GetHomepageContent(ctx, userAuthToken, collectionID, locale, "/")
			mockClient.EXPECT().GetDataset(ctx, headers, "12345").Return(dpDatasetApiModels.Dataset{}, nil)
			mockClient.EXPECT().GetVersions(ctx, headers, "12345", "2017", &dpDatasetApiSdk.QueryParams{Offset: 0, Limit: 1000}).Return(dpDatasetApiSdk.VersionsList{Items: []dpDatasetApiModels.Version{}}, nil)
			mockClient.EXPECT().GetEdition(ctx, headers, "12345", "2017").Return(dpDatasetApiModels.Edition{}, nil)

			mockRend := NewMockRenderClient(mockCtrl)
			mockRend.EXPECT().NewBasePageModel().Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "version-list")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2017/versions", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{edition}/versions", VersionsList(mockClient, mockZebedeeClient, mockRend))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("test versions list returns status 404 when dataset type is static", func() {
			mockClient := NewMockDatasetAPISdkClient(mockCtrl)
			mockClient.EXPECT().GetDataset(ctx, headers, "12345").Return(dpDatasetApiModels.Dataset{
				Type: DatasetTypeStatic,
			}, nil)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2017/versions", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{edition}/versions", VersionsList(mockClient, nil, nil))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("test versions list returns status 500 when dataset client returns an error", func() {
			mockClient := NewMockDatasetAPISdkClient(mockCtrl)
			mockClient.EXPECT().GetDataset(ctx, headers, "12345").Return(dpDatasetApiModels.Dataset{}, errors.New("dataset client error"))

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2017/versions", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{edition}/versions", VersionsList(mockClient, nil, nil))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
