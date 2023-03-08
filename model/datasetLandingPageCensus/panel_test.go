package datasetLandingPageCensus

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFuncGetPanelType(t *testing.T) {
	Convey("Given a panel struct", t, func() {
		Convey("When the FuncGetPanelType function is called", func() {
			Convey("Then it returns the expected panel type", func() {
				tc := []struct {
					given    int
					expected string
				}{
					{
						// info returned to ensure backwards compatibility
						given:    0,
						expected: "info",
					},
					{
						given:    int(Info),
						expected: "info",
					},
					{
						given:    int(Pending),
						expected: "pending",
					},
					{
						given:    int(Success),
						expected: "success",
					},
					{
						given:    int(Error),
						expected: "error",
					},
					{
						given:    25,
						expected: "",
					},
				}
				for _, t := range tc {
					mockPanel := Panel{
						Type: PanelType(t.given),
					}
					So(mockPanel.FuncGetPanelType(), ShouldEqual, t.expected)
				}
			})
		})
	})
}
