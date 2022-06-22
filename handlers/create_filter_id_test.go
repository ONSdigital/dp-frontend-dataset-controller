package handlers

import (
	"errors"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateFilterID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()

	Convey("test CreateFilterID", t, func() {
		mockCfg := config.Config{}

		Convey("test CreateFilterID handler, creates a filter id and redirects", func() {
			mockClient := NewMockFilterClient(mockCtrl)
			mockClient.EXPECT().CreateBlueprint(ctx, userAuthToken, serviceAuthToken, "", collectionID, "1234", "5678", "2017", []string{"aggregate", "time"}).Return("12345", "testETag", nil)

			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			dims := dataset.VersionDimensions{
				Items: []dataset.VersionDimension{
					{
						Name: "aggregate",
					},
					{
						Name: "time",
					},
				},
			}
			mockDatasetClient.EXPECT().GetVersionDimensions(ctx, userAuthToken, serviceAuthToken, collectionID, "1234", "5678", "2017").Return(dims, nil)
			mockDatasetClient.EXPECT().GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, "1234", "5678", "2017", "aggregate",
				&dataset.QueryParams{Offset: 0, Limit: 0}).Return(datasetOptions(0, 0), nil)
			mockDatasetClient.EXPECT().GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, "1234", "5678", "2017", "time",
				&dataset.QueryParams{Offset: 0, Limit: 0}).Return(datasetOptions(0, 0), nil)

			body := strings.NewReader("")
			w := testResponse(301, body, "/datasets/1234/editions/5678/versions/2017/filter", mockClient, mockDatasetClient, false, mockCfg)

			location := w.Header().Get("Location")
			So(location, ShouldNotBeEmpty)

			So(location, ShouldEqual, "/filters/12345/dimensions")
		})

		Convey("test CreateFilterID returns 500 if unable to create a blueprint on filter api", func() {
			mockClient := NewMockFilterClient(mockCtrl)
			mockClient.EXPECT().CreateBlueprint(ctx, userAuthToken, serviceAuthToken, "", collectionID, "1234", "5678", "2017", gomock.Any()).Return("", "", errors.New("unable to create filter blueprint"))

			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockDatasetClient.EXPECT().GetVersionDimensions(ctx, userAuthToken, serviceAuthToken, collectionID, "1234", "5678", "2017").Return(dataset.VersionDimensions{}, nil)

			body := strings.NewReader("")

			testResponse(500, body, "/datasets/1234/editions/5678/versions/2017/filter", mockClient, mockDatasetClient, false, mockCfg)
		})
	})

	Convey("test CreateFilterFlexID", t, func() {
		mockCfg := config.Config{EnableCensusPages: true}
		mockVersions := dataset.VersionsList{
			Items: []dataset.Version{
				{}, // deliberately empty
				{
					Dimensions: []dataset.VersionDimension{
						{
							Name: "geography",
						},
						{
							Name: "age",
						},
					},
				},
			},
		}

		Convey("test CreateFilterFlexID handler, creates a filter id and redirect includes dimension name", func() {
			mockDims := []filter.ModelDimension{
				{
					Name: "geography",
				},
				{
					Name: "age",
				},
			}
			mockClient := NewMockFilterClient(mockCtrl)
			mockClient.EXPECT().CreateFlexibleBlueprint(ctx, userAuthToken, serviceAuthToken, "", collectionID, "1234", "2021", "1", mockDims, "Example").
				Return("12345", "testETag", nil)

			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockDatasetClient.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, "", collectionID, "1234", "2021", "1").
				Return(mockVersions.Items[1], nil)
			mockDatasetClient.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "1234").
				Return(dataset.DatasetDetails{IsBasedOn: &dataset.IsBasedOn{ID: "Example"}}, nil)

			body := strings.NewReader("dimension=geography")
			w := testResponse(301, body, "/datasets/1234/editions/2021/versions/1/filter-flex", mockClient, mockDatasetClient, true, mockCfg)

			location := w.Header().Get("Location")
			So(location, ShouldNotBeEmpty)

			So(location, ShouldEqual, "/filters/12345/dimensions/geography")
		})

		Convey("test CreateFilterFlexID handler, creates a filter id and redirect for coverage appends to geography", func() {
			mockDims := []filter.ModelDimension{
				{
					Name: "geography",
				},
				{
					Name: "age",
				},
			}
			mockClient := NewMockFilterClient(mockCtrl)
			mockClient.EXPECT().CreateFlexibleBlueprint(ctx, userAuthToken, serviceAuthToken, "", collectionID, "1234", "2021", "1", mockDims, "Example").
				Return("12345", "testETag", nil)

			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockDatasetClient.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, "", collectionID, "1234", "2021", "1").
				Return(mockVersions.Items[1], nil)
			mockDatasetClient.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "1234").
				Return(dataset.DatasetDetails{IsBasedOn: &dataset.IsBasedOn{ID: "Example"}}, nil)

			body := strings.NewReader("dimension=coverage")
			w := testResponse(301, body, "/datasets/1234/editions/2021/versions/1/filter-flex", mockClient, mockDatasetClient, true, mockCfg)

			location := w.Header().Get("Location")
			So(location, ShouldNotBeEmpty)

			So(location, ShouldEqual, "/filters/12345/dimensions/geography/coverage")
		})

		Convey("test post route fails if config is false", func() {
			mockCfg := config.Config{EnableCensusPages: false}
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockFilterClient := NewMockFilterClient(mockCtrl)
			body := strings.NewReader("")

			testResponse(500, body, "/datasets/1234/editions/2021/versions/1/filter-flex", mockFilterClient, mockDatasetClient, true, mockCfg)
		})

		Convey("test CreateFilterFlexID returns 500 if unable to create a blueprint on filter api", func() {
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockDatasetClient.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, "", collectionID, "1234", "2021", "1").Return(mockVersions.Items[0], nil)
			mockDatasetClient.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "1234").Return(dataset.DatasetDetails{IsBasedOn: &dataset.IsBasedOn{}}, nil)
			mockFilterClient := NewMockFilterClient(mockCtrl)
			mockFilterClient.EXPECT().CreateFlexibleBlueprint(ctx, userAuthToken, serviceAuthToken, "", collectionID, "1234", "2021", "1", gomock.Any(), "").Return("", "", errors.New("unable to create filter blueprint"))
			body := strings.NewReader("")

			testResponse(500, body, "/datasets/1234/editions/2021/versions/1/filter-flex", mockFilterClient, mockDatasetClient, true, mockCfg)
		})
	})
}

func testResponse(code int, body *strings.Reader, url string, fc FilterClient, dc DatasetClient, filterFlexRoute bool, cfg config.Config) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()

	router := mux.NewRouter()
	if filterFlexRoute {
		router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-flex", CreateFilterFlexID(fc, dc, cfg))
	} else {
		router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter", CreateFilterID(fc, dc))
	}

	router.ServeHTTP(w, req)

	So(w.Code, ShouldEqual, code)

	b, err := ioutil.ReadAll(w.Body)
	So(err, ShouldBeNil)
	// Writer body should be empty, we don't write a response
	So(b, ShouldBeEmpty)

	return w
}
