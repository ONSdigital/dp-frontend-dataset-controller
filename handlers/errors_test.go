package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	datasetAPIErrors "github.com/ONSdigital/dp-dataset-api/apierrors"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSetStatusCode(t *testing.T) {
	ctx := context.Background()

	Convey("test setStatusCode", t, func() {
		testCases := []struct {
			name     string
			err      error
			expected int
		}{
			{
				name:     "test status code handles 404 response from client",
				err:      &testCliError{},
				expected: http.StatusNotFound,
			},
			{
				name:     "test status code handles internal server error",
				err:      errors.New("internal server error"),
				expected: http.StatusInternalServerError,
			},
			{
				name:     "test status code handles known errors with mapped status codes",
				err:      errDatasetTypeNotSupported,
				expected: http.StatusNotFound,
			},
			{
				name:     "test status code handles dataset API client error",
				err:      datasetAPIModels.Error{Code: "404"},
				expected: http.StatusNotFound,
			},
			{
				name:     "test status code handles known dataset API not found error",
				err:      datasetAPIErrors.ErrDatasetNotFound,
				expected: http.StatusNotFound,
			},
		}

		for _, tc := range testCases {
			Convey(tc.name, func() {
				w := httptest.NewRecorder()
				setStatusCode(ctx, w, tc.err)
				So(w.Code, ShouldEqual, tc.expected)
			})
		}
	})
}
