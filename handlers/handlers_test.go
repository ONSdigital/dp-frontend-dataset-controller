package handlers

import (
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

	Convey("test setStatusCode", t, func() {

		Convey("test status code handles 404 response from client", func() {
			req := httptest.NewRequest("GET", "http://localhost:20000", nil)
			w := httptest.NewRecorder()
			err := &testCliError{}

			setStatusCode(req, w, err)

			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("test status code handles internal server error", func() {
			req := httptest.NewRequest("GET", "http://localhost:20000", nil)
			w := httptest.NewRecorder()
			err := errors.New("internal server error")

			setStatusCode(req, w, err)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
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
