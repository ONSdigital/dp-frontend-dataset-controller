package permissions

import (
	"errors"
	"testing"

	authorisation "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"
	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	"github.com/ONSdigital/dp-permissions-api/sdk"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPermissionsCheck(t *testing.T) {
	Convey("When a valid JWT token is submitted", t, func() {
		Convey("test user is a publishing manager (admin)", func() {
			adminJWTToken := authorisationtest.AdminJWTToken

			middlewareMock := &authorisation.MiddlewareMock{
				ParseFunc: func(token string) (*sdk.EntityData, error) {
					return &sdk.EntityData{
						Groups: []string{"role-admin"},
					}, nil
				},
			}
			adminUser, err := CheckIsAdmin(t.Context(), adminJWTToken, middlewareMock)
			So(err, ShouldBeNil)
			So(adminUser, ShouldBeTrue)
		})

		Convey("test user is a publisher", func() {
			publisherJWTToken := authorisationtest.PublisherJWTToken

			middlewareMock := &authorisation.MiddlewareMock{
				ParseFunc: func(token string) (*sdk.EntityData, error) {
					return &sdk.EntityData{
						Groups: []string{"role-publisher"},
					}, nil
				},
			}
			adminUser, err := CheckIsAdmin(t.Context(), publisherJWTToken, middlewareMock)
			So(err, ShouldBeNil)
			So(adminUser, ShouldBeFalse)
		})
	})

	Convey("When an invalid JWT token is submitted", t, func() {
		Convey("an error will be returned", func() {
			invalidToken := "test-123-token"

			middlewareMock := &authorisation.MiddlewareMock{
				ParseFunc: func(token string) (*sdk.EntityData, error) {
					return nil, errors.New("jwt token is not valid")
				},
			}
			_, err := CheckIsAdmin(t.Context(), invalidToken, middlewareMock)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "jwt token is not valid")
		})
	})
}
