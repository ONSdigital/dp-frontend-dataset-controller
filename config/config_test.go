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
				So(cfg.BindAddr, ShouldEqual, "localhost:20200")
				So(cfg.Debug, ShouldBeFalse)
				So(cfg.EnableCensusPages, ShouldBeFalse)
				So(cfg.APIRouterURL, ShouldEqual, "http://localhost:23200/v1")
				So(cfg.DownloadServiceURL, ShouldEqual, "http://localhost:23600")
				So(cfg.SiteDomain, ShouldEqual, "localhost")
				So(cfg.SupportedLanguages, ShouldResemble, []string{"en", "cy"})
				So(cfg.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				So(cfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(cfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(cfg.EnableProfiler, ShouldBeFalse)
				So(cfg.PatternLibraryAssetsPath, ShouldEqual, "//cdn.ons.gov.uk/dp-design-system/ba32e79")
			})
		})
	})
}
