package handlers

import (
	"errors"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	Filter           string = "filter"
	FilterFlex       string = "filter-flex"
	FilterFlexOutput string = "filter-flex-output"
)

func TestCreateFilterID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()

	Convey("test CreateFilterID", t, func() {
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
			w := testResponse(301, body, "/datasets/1234/editions/5678/versions/2017/filter", mockClient, mockDatasetClient, Filter)

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

			testResponse(500, body, "/datasets/1234/editions/5678/versions/2017/filter", mockClient, mockDatasetClient, Filter)
		})
	})

	Convey("test CreateFilterFlexID", t, func() {
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
			w := testResponse(301, body, "/datasets/1234/editions/2021/versions/1", mockClient, mockDatasetClient, FilterFlex)

			location := w.Header().Get("Location")
			So(location, ShouldNotBeEmpty)

			So(location, ShouldEqual, "/filters/12345/dimensions/geography")
		})

		Convey("test CreateFilterFlexID handler, creates a filter id and redirect for multivariate", func() {
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

			body := strings.NewReader("dimension=change")
			w := testResponse(301, body, "/datasets/1234/editions/2021/versions/1", mockClient, mockDatasetClient, FilterFlex)

			location := w.Header().Get("Location")
			So(location, ShouldNotBeEmpty)

			So(location, ShouldEqual, "/filters/12345/dimensions/change")
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
			w := testResponse(301, body, "/datasets/1234/editions/2021/versions/1", mockClient, mockDatasetClient, FilterFlex)

			location := w.Header().Get("Location")
			So(location, ShouldNotBeEmpty)

			So(location, ShouldEqual, "/filters/12345/dimensions/geography/coverage")
		})

		Convey("test CreateFilterFlexID returns 500 if unable to create a blueprint on filter api", func() {
			mockDatasetClient := NewMockDatasetClient(mockCtrl)
			mockDatasetClient.EXPECT().GetVersion(ctx, userAuthToken, serviceAuthToken, "", collectionID, "1234", "2021", "1").Return(mockVersions.Items[0], nil)
			mockDatasetClient.EXPECT().Get(ctx, userAuthToken, serviceAuthToken, collectionID, "1234").Return(dataset.DatasetDetails{IsBasedOn: &dataset.IsBasedOn{}}, nil)
			mockFilterClient := NewMockFilterClient(mockCtrl)
			mockFilterClient.EXPECT().CreateFlexibleBlueprint(ctx, userAuthToken, serviceAuthToken, "", collectionID, "1234", "2021", "1", gomock.Any(), "").Return("", "", errors.New("unable to create filter blueprint"))
			body := strings.NewReader("")

			testResponse(500, body, "/datasets/1234/editions/2021/versions/1", mockFilterClient, mockDatasetClient, FilterFlex)
		})
	})

	Convey("test CreateFilterFlexIDFromOutput", t, func() {
		mockFo := filter.Model{
			Dataset: filter.Dataset{
				DatasetID: "1234",
				Edition:   "2021",
				Version:   1,
			},
			PopulationType: "Example",
			Dimensions: []filter.ModelDimension{
				{
					Name:       "geography",
					IsAreaType: toBoolPtr(true),
					Options: []string{
						"option 1", "option 2",
					},
					FilterByParent: "country",
				},
				{
					Name:           "another dim",
					IsAreaType:     new(bool),
					Options:        []string{},
					FilterByParent: "",
				},
			},
		}

		Convey("test CreateFilterFlexIDFromOutput handler, creates a filter id and redirect for multivariate", func() {
			mockFc := NewMockFilterClient(mockCtrl)
			mockFc.
				EXPECT().
				GetOutput(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(mockFo, nil)
			mockFc.
				EXPECT().
				CreateFlexibleBlueprint(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), mockFo.Dataset.DatasetID, mockFo.Dataset.Edition, "1", mockFo.Dimensions, mockFo.PopulationType).
				Return("12345", "testETag", nil)

			body := strings.NewReader("dimension=change")
			w := testResponse(301, body, "/datasets/1234/editions/2021/versions/1/filter-outputs/5678", mockFc, NewMockDatasetClient(mockCtrl), FilterFlexOutput)

			location := w.Header().Get("Location")
			So(location, ShouldNotBeEmpty)

			So(location, ShouldEqual, "/filters/12345/dimensions/change")
		})

		Convey("test CreateFilterFlexIDFromOutput handler, creates a filter id and redirect includes dimension name", func() {
			mockFc := NewMockFilterClient(mockCtrl)
			mockFc.
				EXPECT().
				GetOutput(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(mockFo, nil)
			mockFc.
				EXPECT().
				CreateFlexibleBlueprint(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), mockFo.Dataset.DatasetID, mockFo.Dataset.Edition, "1", mockFo.Dimensions, mockFo.PopulationType).
				Return("12345", "testETag", nil)

			body := strings.NewReader("dimension=geography")
			w := testResponse(301, body, "/datasets/1234/editions/2021/versions/1/filter-outputs/5678", mockFc, NewMockDatasetClient(mockCtrl), FilterFlexOutput)

			location := w.Header().Get("Location")
			So(location, ShouldNotBeEmpty)

			So(location, ShouldEqual, "/filters/12345/dimensions/geography")
		})

		Convey("test CreateFilterFlexIDFromOutput handler, creates a filter id and redirect for coverage appends to geography", func() {
			mockFc := NewMockFilterClient(mockCtrl)
			mockFc.
				EXPECT().
				GetOutput(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(mockFo, nil)
			mockFc.
				EXPECT().
				CreateFlexibleBlueprint(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), mockFo.Dataset.DatasetID, mockFo.Dataset.Edition, "1", mockFo.Dimensions, mockFo.PopulationType).
				Return("12345", "testETag", nil)

			body := strings.NewReader("dimension=coverage")
			w := testResponse(301, body, "/datasets/1234/editions/2021/versions/1/filter-outputs/5678", mockFc, NewMockDatasetClient(mockCtrl), FilterFlexOutput)

			location := w.Header().Get("Location")
			So(location, ShouldNotBeEmpty)

			So(location, ShouldEqual, "/filters/12345/dimensions/geography/coverage")
		})

		Convey("test CreateFilterFlexIDFromOutput returns 500 if unable to get filter record on filter api", func() {
			mockFc := NewMockFilterClient(mockCtrl)
			mockFc.
				EXPECT().
				GetOutput(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(filter.Model{}, errors.New("unable to get filter job"))
			body := strings.NewReader("")

			testResponse(500, body, "/datasets/1234/editions/2021/versions/1/filter-outputs/5678", mockFc, NewMockDatasetClient(mockCtrl), FilterFlexOutput)
		})

		Convey("test CreateFilterFlexIDFromOutput returns 500 if unable to create a blueprint on filter api", func() {
			mockFc := NewMockFilterClient(mockCtrl)
			mockFc.
				EXPECT().
				GetOutput(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(mockFo, nil)
			mockFc.
				EXPECT().
				CreateFlexibleBlueprint(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), mockFo.Dataset.DatasetID, mockFo.Dataset.Edition, "1", mockFo.Dimensions, mockFo.PopulationType).
				Return("", "", errors.New("unable to create filter blueprint"))
			body := strings.NewReader("")

			testResponse(500, body, "/datasets/1234/editions/2021/versions/1/filter-outputs/5678", mockFc, NewMockDatasetClient(mockCtrl), FilterFlexOutput)
		})
	})
}

func testResponse(code int, body *strings.Reader, url string, fc FilterClient, dc DatasetClient, filterFlexRoute string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()

	router := mux.NewRouter()
	switch filterFlexRoute {
	case FilterFlex:
		router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}", CreateFilterFlexID(fc, dc))
	case FilterFlexOutput:
		router.HandleFunc("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter-outputs/{filterOutputID}", CreateFilterFlexIDFromOutput(fc))
	case Filter:
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
