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

func TestVersionList(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()
	cfg := initialiseMockConfig()

	Convey("test versions list", t, func() {
		Convey("test versions list returns 200 when rendered successfully", func() {
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
}
