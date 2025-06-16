package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
)

func TestFilterPageHandler(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockContext := gomock.Any()
	mockConfig := config.Config{
		FilterFlexDatasetServiceURL:        "http://localhost:27100",
		FrontendFilterDatasetControllerURL: "http://localhost:20001",
	}

	mockFilterClient := NewMockFilterClient(mockCtrl)
	mockDatasetClient := NewMockDatasetAPISdkClient(mockCtrl)

	mockFilterModel := &filter.Model{
		Dataset: filter.Dataset{
			DatasetID: "123",
		},
	}

	headers := dpDatasetApiSdk.Headers{}

	Convey("Given a FilterPageHandler", t, func() {
		handler := FilterPageHandler(mockFilterClient, mockDatasetClient)
		router := mux.NewRouter()
		router.HandleFunc("/filters/{filterID}/dimensions", handler)
		router.HandleFunc("/filters/{filterID}/dimensions/{dimension}", handler)

		Convey("When filterID is missing", func() {
			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest(http.MethodGet, "/filters/missing-filter-id/dimensions", http.NoBody)

			// Override behaviour — simulate the missing filter id
			vars := map[string]string{"filterID": ""}
			mockRequest = mux.SetURLVars(mockRequest, vars)

			handler := FilterPageHandler(mockFilterClient, mockDatasetClient)
			handler(mockRequestWriter, mockRequest)

			Convey("Then the status code is 400", func() {
				So(mockRequestWriter.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("If GetJobState returns error", func() {
			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest(http.MethodGet, "/filters/123/dimensions", http.NoBody)

			mockFilterClient.EXPECT().
				GetJobState(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "123").
				Return(filter.Model{}, "", errors.New("some error"))

			router.ServeHTTP(mockRequestWriter, mockRequest)

			Convey("Then the status code is 500", func() {
				So(mockRequestWriter.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("If error is returned fetching dataset", func() {
			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest(http.MethodGet, "/filters/123/dimensions", http.NoBody)

			mockFilterClient.EXPECT().GetJobState(
				gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "123",
			).Return(*mockFilterModel, "", nil)

			mockDatasetClient.EXPECT().GetDataset(
				mockContext, headers, collectionID, "123",
			).Return(dpDatasetApiModels.Dataset{}, errors.New("dataset error"))

			router.ServeHTTP(mockRequestWriter, mockRequest)

			Convey("Then the status code is 500", func() {
				So(mockRequestWriter.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("If dataset is 'cantabular' type", func() {
			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest(http.MethodGet, "/filters/123/dimensions/ltla", http.NoBody)

			datasetDetails := dpDatasetApiModels.Dataset{
				Type: "cantabular",
			}

			mockFilterClient.EXPECT().GetJobState(
				gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "123",
			).Return(*mockFilterModel, "", nil)

			mockDatasetClient.EXPECT().GetDataset(
				mockContext, headers, collectionID, "123",
			).Return(datasetDetails, nil)

			router.ServeHTTP(mockRequestWriter, mockRequest)

			Convey("Then the status code is 307", func() {
				So(mockRequestWriter.Code, ShouldEqual, http.StatusTemporaryRedirect)
			})

			// TODO: look at other redirects used in this repo
			Convey("Then a redirect is made to FilterFlexDatasetServiceURL ", func() {
				So(mockRequestWriter.Header().Get("Location"), ShouldStartWith, mockConfig.FilterFlexDatasetServiceURL)
			})
		})

		Convey("If dataset is not 'cantabular' type", func() {
			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest(http.MethodGet, "/filters/123/dimensions", http.NoBody)

			datasetDetails := dpDatasetApiModels.Dataset{
				Type: "cmd",
			}

			mockFilterClient.EXPECT().GetJobState(
				gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "123",
			).Return(*mockFilterModel, "", nil)

			mockDatasetClient.EXPECT().GetDataset(
				mockContext, headers, collectionID, "123",
			).Return(datasetDetails, nil)

			router.ServeHTTP(mockRequestWriter, mockRequest)

			Convey("Then the status code is 307", func() {
				So(mockRequestWriter.Code, ShouldEqual, http.StatusTemporaryRedirect)
			})

			// TODO: look at other redirects used in this repo
			Convey("Then a redirect is made to FrontendFilterDatasetControllerURL ", func() {
				So(mockRequestWriter.Header().Get("Location"), ShouldStartWith, mockConfig.FrontendFilterDatasetControllerURL)
			})
		})
	})
}
