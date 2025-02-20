package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

// Tests for `FilterableLanding` for any dataset type
func TestLatestVersionRedirect(t *testing.T) {
	// Mocks
	mockContext := gomock.Any()
	mockController := gomock.NewController(t)
	mockDatasetClient := NewMockDatasetClient(mockController)

	// Default test values
	datasetId := "12345"
	downloadServiceAuthToken := ""
	editionId := "67890"
	getVersionsQueryParams := dataset.QueryParams{Offset: 0, Limit: 1000}
	mockGetVersionsResponse := dataset.VersionsList{
		Items: []dataset.Version{
			{
				Version: 1,
			},
			{
				Version: 2,
			},
		},
		Count: 2,
	}

	Convey("Test latest version redirect", t, func() {
		Convey("Test latestVersionRedirect returns 500 error if dataset versions are not found", func() {
			// Dataset client `GetVersions()` will return an error if dataset versions are not found
			mockDatasetClient.EXPECT().GetVersions(
				mockContext, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetId,
				editionId, &getVersionsQueryParams,
			).Return(
				mockGetVersionsResponse, errors.New("sorry"),
			)

			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest("GET", fmt.Sprintf("/datasets/%s/editions/%s", datasetId, editionId), nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}", LatestVersionRedirect(mockDatasetClient))

			router.ServeHTTP(mockRequestWriter, mockRequest)

			So(mockRequestWriter.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Test latestVersionRedirect returns 302 and redirects to the correct url for edition level requests without version", func() {
			// Dataset client `GetVersions()` will return a list of versions if dataset versions are found
			mockDatasetClient.EXPECT().GetVersions(
				mockContext, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetId,
				editionId, &getVersionsQueryParams,
			).Return(
				mockGetVersionsResponse, nil,
			)
			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest("GET", fmt.Sprintf("/datasets/%s/editions/%s", datasetId, editionId), nil)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}", LatestVersionRedirect(mockDatasetClient))

			router.ServeHTTP(mockRequestWriter, mockRequest)

			So(mockRequestWriter.Code, ShouldEqual, http.StatusFound)
			// There are two versions so should redirect to version 2
			So(mockRequestWriter.Body.String(), ShouldEqual, fmt.Sprintf("<a href=\"/datasets/%s/editions/%s/versions/2\">Found</a>.\n\n", datasetId, editionId))
		})
	})
}
