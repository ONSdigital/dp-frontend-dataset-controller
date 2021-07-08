package config

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// TestConfig tests config options correctly default if not set
func TestConfig(t *testing.T) {
	t.Parallel()
	Convey("Given an environment with no environment variables set", t, func() {
		cfg, err := Get()

		Convey("When the config values are retrieved", func() {

			Convey("Then there should be no error returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("That the values should be set to the expected defaults", func() {
				So(cfg.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				So(cfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(cfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(cfg.BindAddr, ShouldEqual, ":20200")
				So(cfg.PatternLibraryAssetsPath, ShouldEqual, "//cdn.ons.gov.uk/sixteens/c690dcd")
				So(cfg.SiteDomain, ShouldEqual, "localhost")
			})
		})
	})
}
