package handlers

import (
	coreContext "context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

type testCliError struct{}

func (e *testCliError) Error() string { return "client error" }
func (e *testCliError) Code() int     { return http.StatusNotFound }

const serviceAuthToken = ""
const userAuthToken = ""
const collectionID = ""
const locale = "en"

// datasetOptions returns a mocked dataset.Options struct according to the provided offset and limit
func datasetOptions(offset, limit int) dataset.Options {
	allItems := []dataset.Option{
		{
			Label:  "1",
			Option: "abd",
		},
		{
			Label:  "2",
			Option: "fjd",
		},
	}
	o := dataset.Options{
		Offset:     offset,
		Limit:      limit,
		TotalCount: len(allItems),
	}
	o.Items = slice(allItems, offset, limit)
	o.Count = len(o.Items)
	return o
}

func slice(full []dataset.Option, offset, limit int) (sliced []dataset.Option) {
	end := offset + limit
	if end > len(full) {
		end = len(full)
	}

	if offset > len(full) || limit == 0 {
		return []dataset.Option{}
	}

	return full[offset:end]
}

func TestUnitHandlers(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := coreContext.Background()

	Convey("test setStatusCode", t, func() {

		Convey("test status code handles 404 response from client", func() {
			w := httptest.NewRecorder()
			err := &testCliError{}

			setStatusCode(ctx, w, err)

			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("test status code handles internal server error", func() {
			w := httptest.NewRecorder()
			err := errors.New("internal server error")

			setStatusCode(ctx, w, err)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func TestSortOptionsByCode(t *testing.T) {

	Convey("Dimension options are sorted", t, func() {
		getOptionList := func(items []dataset.Option) []string {
			results := []string{}
			for _, item := range items {
				results = append(results, item.Option)
			}
			return results
		}

		Convey("given non-numeric options", func() {
			nonNumeric := []dataset.Option{
				{
					DimensionID: "dim_2",
					Option:      "option 2",
				},
				{
					DimensionID: "dim_1",
					Option:      "option 1",
				},
			}
			Convey("when they are sorted", func() {
				sorted := sortOptionsByCode(nonNumeric)

				Convey("then options are sorted alphabetically", func() {
					actual := getOptionList(sorted)
					expected := []string{"option 1", "option 2"}
					So(actual, ShouldResemble, expected)
				})
			})
		})

		Convey("given simple numeric options", func() {
			numeric := []dataset.Option{
				{
					DimensionID: "dim_2",
					Option:      "2",
				},
				{
					DimensionID: "dim_10",
					Option:      "10",
				},
				{
					DimensionID: "dim_1",
					Option:      "1",
				},
			}
			Convey("when they are sorted", func() {
				sorted := sortOptionsByCode(numeric)

				Convey("then options are sorted numerically", func() {
					actual := getOptionList(sorted)
					expected := []string{"1", "2", "10"}
					So(actual, ShouldResemble, expected)
				})
			})
		})

		Convey("given numeric options with negatives", func() {
			numericWithNegatives := []dataset.Option{
				{
					DimensionID: "dim_2",
					Option:      "2",
				},
				{
					DimensionID: "dim_-1",
					Option:      "-1",
				},
				{
					DimensionID: "dim_10",
					Option:      "10",
				},
				{
					DimensionID: "dim_-10",
					Option:      "-10",
				},
				{
					DimensionID: "dim_1",
					Option:      "1",
				},
			}
			Convey("when they are sorted", func() {
				sorted := sortOptionsByCode(numericWithNegatives)

				Convey("then options are sorted numerically with negatives at the end", func() {
					actual := getOptionList(sorted)
					expected := []string{"1", "2", "10", "-1", "-10"}
					So(actual, ShouldResemble, expected)
				})
			})
		})

		Convey("given mixed numeric and non-numeric options", func() {
			alphanumeric := []dataset.Option{
				{
					DimensionID: "dim_2",
					Option:      "2nd Option",
				},
				{
					DimensionID: "dim_1",
					Option:      "1",
				},
				{
					DimensionID: "dim_10",
					Option:      "10",
				},
			}
			Convey("when they are sorted", func() {
				sorted := sortOptionsByCode(alphanumeric)

				Convey("then options are sorted alphanumerically", func() {
					actual := getOptionList(sorted)
					expected := []string{"1", "10", "2nd Option"}
					So(actual, ShouldResemble, expected)
				})
			})
		})
	})
}

func initialiseMockConfig() config.Config {
	return config.Config{
		PatternLibraryAssetsPath: "http://localhost:9000/dist",
		SiteDomain:               "ons",
		SupportedLanguages:       []string{"en", "cy"},
	}
}
