package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPostCreateCustomDataset(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()

	Convey("test PostCreateCustomDataset", t, func() {
		Convey("happy path creates a filter id and redirects to filter page", func() {

			// GIVEN - a valid request
			formData := url.Values{}
			formData.Add("populationType", "UR")
			encodedFormData := formData.Encode()

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/datasets/create"), strings.NewReader(encodedFormData))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Content-Length", strconv.Itoa(len(encodedFormData)))

			// AND - a filter client mocked to return a happy response
			mockFc := NewMockFilterClient(mockCtrl)
			mockFc.
				EXPECT().
				CreateCustomFilter(ctx, gomock.Any(), gomock.Any(), "UR").
				Return("12345", nil)

			// WHEN - we post the request
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/datasets/create", PostCreateCustomDataset(mockFc))
			router.ServeHTTP(w, req)

			// THEN - we are rerouted to the filter review page (without the filter client being called)
			So(w.Code, ShouldEqual, http.StatusMovedPermanently)

			location := w.Header().Get("Location")
			So(location, ShouldNotBeEmpty)
			So(location, ShouldEqual, "/filters/12345/dimensions")
		})

		Convey("redirects to page with error=true if form data invalid", func() {
			// GIVEN - a form where the populationType field is missing
			formData := url.Values{}
			encodedFormData := formData.Encode()

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/datasets/create"), strings.NewReader(encodedFormData))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Content-Length", strconv.Itoa(len(encodedFormData)))

			// AND - a filter client mocked to expect no usage
			mockFc := NewMockFilterClient(mockCtrl)
			mockFc.
				EXPECT().
				CreateCustomFilter(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
				Times(0)

			// WHEN - we post the request
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/datasets/create", PostCreateCustomDataset(mockFc))
			router.ServeHTTP(w, req)

			// THEN - we are rerouted to the filter review page
			location := w.Header().Get("Location")
			So(location, ShouldNotBeEmpty)
			So(location, ShouldEqual, "/datasets/create?error=true")
		})

		Convey("returns 500 when the filter API client responds with an error", func() {
			// GIVEN - a valid request
			formData := url.Values{}
			formData.Add("populationType", "UR")
			encodedFormData := formData.Encode()

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/datasets/create"), strings.NewReader(encodedFormData))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Content-Length", strconv.Itoa(len(encodedFormData)))

			// BUT - a filter client mocked to return an error
			mockFc := NewMockFilterClient(mockCtrl)
			mockFc.
				EXPECT().
				CreateCustomFilter(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
				Return("", errors.New("internal error"))

			// WHEN - we post the request
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/datasets/create", PostCreateCustomDataset(mockFc))
			router.ServeHTTP(w, req)

			// THEN - the client should not be redirected
			So(w.Header().Get("Location"), ShouldBeEmpty)

			// AND - the status code should be 500
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
