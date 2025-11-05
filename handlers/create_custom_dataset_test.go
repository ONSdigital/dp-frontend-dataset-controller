package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dis-design-system-go/helper"
	core "github.com/ONSdigital/dis-design-system-go/model"
	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper/mocks"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateCustomDatasetHandlers(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()
	cfg := initialiseMockConfig()
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)

	zc := NewMockZebedeeClient(mockCtrl)
	pc := NewMockPopulationClient(mockCtrl)
	rend := NewMockRenderClient(mockCtrl)

	mockPopulationTypes := population.GetPopulationTypesResponse{
		Items: []population.PopulationType{
			{
				Name:        "name",
				Label:       "label",
				Description: "Description",
			},
		},
	}

	Convey("Given the expected calls to render a Create Custom Dataset page", t, func() {
		rend.EXPECT().NewBasePageModel().Return(core.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))
		rend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "create-custom-dataset")

		zc.EXPECT().GetHomepageContent(ctx, userAuthToken, collectionID, locale, "/")
		pc.EXPECT().GetPopulationTypes(ctx, gomock.Any()).Return(mockPopulationTypes, nil)

		Convey("When the page is rendered", func() {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/create", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/create", CreateCustomDataset(pc, zc, rend, cfg, ""))
			router.ServeHTTP(w, req)

			Convey("Then it returns StatusOK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}
