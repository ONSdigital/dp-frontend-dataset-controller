package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestApproveDatasetVersion(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()

	Convey("test ApproveDatasetVersion", t, func() {
		Convey("approves version and redirects to version page", func() {
			mockClient := NewMockDatasetAPISdkClient(mockCtrl)

			headers := dpDatasetApiSdk.Headers{
				CollectionID:         collectionID,
				DownloadServiceToken: "",
				AccessToken:          userAuthToken,
			}

			mockClient.
				EXPECT().
				PutVersionState(ctx, headers, "12345", "2017", "1", "approved").
				Return(nil)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2017/versions/1/approve", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/approve", ApproveDatasetVersion(mockClient, config.Config{}))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusTemporaryRedirect)
			So(w.Header().Get("Location"), ShouldEqual, "/datasets/12345/editions/2017/versions/1")
		})

		Convey("logs error from dataset client but still redirects", func() {
			mockClient := NewMockDatasetAPISdkClient(mockCtrl)

			headers := dpDatasetApiSdk.Headers{
				CollectionID:         collectionID,
				DownloadServiceToken: "",
				AccessToken:          userAuthToken,
			}

			mockClient.
				EXPECT().
				PutVersionState(ctx, headers, "12345", "2017", "1", "approved").
				Return(errors.New("approval failed"))

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/datasets/12345/editions/2017/versions/1/approve", http.NoBody)

			router := mux.NewRouter()
			router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/approve", ApproveDatasetVersion(mockClient, config.Config{}))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusTemporaryRedirect)
			So(w.Header().Get("Location"), ShouldEqual, "/datasets/12345/editions/2017/versions/1")
		})
	})
}
