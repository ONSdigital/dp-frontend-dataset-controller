package handlers

import (
	coreContext "context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/population"
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
	Convey("Population categories are sorted", t, func() {
		getCategoryList := func(items []population.DimensionCategoryItem) []string {
			results := make([]string, 0, len(items))
			for _, item := range items {
				results = append(results, item.ID)
			}
			return results
		}

		Convey("given non-numeric options", func() {
			nonNumeric := []population.DimensionCategoryItem{
				{
					ID: "option 2",
				},
				{
					ID: "option 1",
				},
			}
			Convey("when they are sorted", func() {
				sorted := sortCategoriesByID(nonNumeric)

				Convey("then categories are sorted alphabetically", func() {
					actual := getCategoryList(sorted)
					expected := []string{"option 1", "option 2"}
					So(actual, ShouldResemble, expected)
				})
			})
		})

		Convey("given simple numeric options", func() {
			simpleNumeric := []population.DimensionCategoryItem{
				{
					ID: "2",
				},
				{
					ID: "10",
				},
				{
					ID: "1",
				},
			}
			Convey("when they are sorted", func() {
				sorted := sortCategoriesByID(simpleNumeric)

				Convey("then options are sorted numerically", func() {
					actual := getCategoryList(sorted)
					expected := []string{"1", "2", "10"}
					So(actual, ShouldResemble, expected)
				})
			})
		})

		Convey("given numeric options with negatives", func() {
			numeric := []population.DimensionCategoryItem{
				{
					ID: "2",
				},
				{
					ID: "-1",
				},
				{
					ID: "10",
				},
				{
					ID: "-10",
				},
				{
					ID: "1",
				},
			}

			Convey("when they are sorted", func() {
				sorted := sortCategoriesByID(numeric)

				Convey("then options are sorted numerically with negatives at the end", func() {
					actual := getCategoryList(sorted)
					expected := []string{"1", "2", "10", "-1", "-10"}
					So(actual, ShouldResemble, expected)
				})
			})
		})

		Convey("given mixed numeric and non-numeric options", func() {
			mixed := []population.DimensionCategoryItem{
				{
					ID: "2nd Option",
				},
				{
					ID: "1",
				},
				{
					ID: "10",
				},
			}
			Convey("when they are sorted", func() {
				sorted := sortCategoriesByID(mixed)

				Convey("then options are sorted alphanumerically", func() {
					actual := getCategoryList(sorted)
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
