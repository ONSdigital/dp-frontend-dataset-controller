package helpers

import (
	"fmt"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitHelpers(t *testing.T) {
	Convey("test ExtractDatasetInfoFromPath", t, func() {
		Convey("extracts datasetID, edition and version from path", func() {
			datasetID, edition, version, err := ExtractDatasetInfoFromPath("/datasets/12345/editions/2016/versions/1")
			So(err, ShouldBeNil)
			So(datasetID, ShouldEqual, "12345")
			So(edition, ShouldEqual, "2016")
			So(version, ShouldEqual, "1")
		})

		Convey("returns an error if it is unable to extract the information", func() {
			datasetID, edition, version, err := ExtractDatasetInfoFromPath("invalid")
			So(err, ShouldBeError, "unable to extract datasetID, edition and version from path: invalid")
			So(datasetID, ShouldEqual, "")
			So(edition, ShouldEqual, "")
			So(version, ShouldEqual, "")
		})
	})
}

func TestDatasetVersionURL(t *testing.T) {
	Convey("The dataset version URL is correctly constructed from the provided parameters", t, func() {
		So(DatasetVersionURL("myDataset", "myEdition", "myVersion"), ShouldResemble,
			"/datasets/myDataset/editions/myEdition/versions/myVersion")
	})
}

func TestGetAPIRouterVersion(t *testing.T) {
	Convey("The api router version is correctly extracted from a valid API Router URL", t, func() {
		version, err := GetAPIRouterVersion("http://localhost:23200/v1")
		So(err, ShouldBeNil)
		So(version, ShouldEqual, "/v1")
	})

	Convey("An empty string version is extracted from a valid unversioned API Router URL", t, func() {
		version, err := GetAPIRouterVersion("http://localhost:23200")
		So(err, ShouldBeNil)
		So(version, ShouldEqual, "")
	})

	Convey("Extracting a version from an invalid API Router URL results in the parsing error being returned", t, func() {
		version, err := GetAPIRouterVersion("hello%goodbye")
		So(err, ShouldResemble, &url.Error{
			Op:  "parse",
			URL: "hello%goodbye",
			Err: url.EscapeError("%go"),
		})
		So(version, ShouldEqual, "")
	})
}

func TestGetCurrentUrl(t *testing.T) {
	Convey("The current URL is correctly constructed from the parameters", t, func() {
		So(GetCurrentURL("en", "mydomain.com", "/page1/page2"), ShouldResemble, "mydomain.com/page1/page2")
		So(GetCurrentURL("en", "mydomain.com", ""), ShouldResemble, "mydomain.com")
		So(GetCurrentURL("cy", "mydomain.com", ""), ShouldResemble, "cy.mydomain.com")
		So(GetCurrentURL("cy", "mydomain.com", "/page1"), ShouldResemble, "cy.mydomain.com/page1")
		So(GetCurrentURL("en", "localhost", "/page1"), ShouldResemble, "ons.gov.uk/page1")
		So(GetCurrentURL("cy", "localhost", "/page1"), ShouldResemble, "cy.ons.gov.uk/page1")
		So(GetCurrentURL("en", "", "/page1"), ShouldResemble, "ons.gov.uk/page1")
	})
}

func TestGenerateSharingLink(t *testing.T) {
	Convey("The sharing link is correctly constructed from the parameters", t, func() {
		const title = "a title"
		const url = "mydomain.com/page"
		So(GenerateSharingLink("", url, title), ShouldBeBlank)
		So(GenerateSharingLink("facebook", url, title), ShouldResemble, fmt.Sprintf("https://www.facebook.com/sharer/sharer.php?u=%s", url))
		So(GenerateSharingLink("twitter", url, title), ShouldResemble, fmt.Sprintf("https://twitter.com/intent/tweet?original_referer&text=%s&url=%s", title, url))
		So(GenerateSharingLink("linkedin", url, title), ShouldResemble, fmt.Sprintf("https://www.linkedin.com/sharing/share-offsite/?url=%s", url))
		So(GenerateSharingLink("email", url, title), ShouldResemble, fmt.Sprintf("mailto:?subject=%s&body=%s\n%s", title, title, url))
	})
}

func TestIsBoolPtr(t *testing.T) {
	Convey("When the value is nil", t, func() {
		Convey("Then the returned value is false", func() {
			So(IsBoolPtr(nil), ShouldBeFalse)
		})
	})
	Convey("When the value is a false pointer", t, func() {
		ptr := *new(bool)
		Convey("Then the returned value is false", func() {
			So(IsBoolPtr(&ptr), ShouldBeFalse)
		})
	})
	Convey("When the value is a true pointer", t, func() {
		ptr := bool(true)
		Convey("Then the returned value is true", func() {
			So(IsBoolPtr(&ptr), ShouldBeTrue)
		})
	})
}

func TestHasStringInSlice(t *testing.T) {
	Convey("Should return false when searching an empty array for the empty string", t, func() {
		So(HasStringInSlice("", []string{}), ShouldBeFalse)
	})
	Convey("Should return false when searching a populated array for the empty string", t, func() {
		So(HasStringInSlice("", []string{"hello", "world"}), ShouldBeFalse)
	})
	Convey("Should return false when searching an empty array for a given string", t, func() {
		So(HasStringInSlice("hello", []string{}), ShouldBeFalse)
	})
	Convey("Should return true when searching a populated array known to contain the given string", t, func() {
		So(HasStringInSlice("hello", []string{"hello", "world"}), ShouldBeTrue)
	})
}

func TestCheckForExistingParams(t *testing.T) {
	Convey("persist existing query string values and ignore given value", t, func() {
		queryStrValues := []string{"Value 1", "Value 2"}
		ignoreValue := "Value 1"
		key := "testKey"
		q := url.Values{}

		PersistExistingParams(queryStrValues, key, ignoreValue, q)
		So(q[key], ShouldResemble, []string{"Value 2"})
	})

	Convey("persist existing query string values no value to ignore", t, func() {
		queryStrValues := []string{"Value 1", "Value 2"}
		existingValue := ""
		key := "testKey"
		q := url.Values{}

		PersistExistingParams(queryStrValues, key, existingValue, q)
		So(q[key], ShouldResemble, queryStrValues)
	})

	Convey("invalid key given no values persisted", t, func() {
		queryStrValues := []string{"Value 1", "Value 2"}
		existingValue := ""
		key := "testKey"
		q := url.Values{}

		PersistExistingParams(queryStrValues, key, existingValue, q)
		So(q["another key"], ShouldBeNil)
		So(q[key], ShouldResemble, queryStrValues)
	})
}

func TestToBoolPtr(t *testing.T) {
	Convey("Given a bool, a pointer is returned", t, func() {
		So(ToBoolPtr(false), ShouldResemble, new(bool))
		So(ToBoolPtr(false), ShouldNotBeNil)
		truePtr := func(b bool) *bool { return &b }(true)
		So(ToBoolPtr(true), ShouldResemble, truePtr)
	})
}

func TestGetOfficialStatisticsLogo(t *testing.T) {
    Convey("Given enableOfficialStatisticsLogo is true and useSvg is true", t, func() {
        result := GetOfficialStatisticsLogo(true, true, "en")
        So(result, ShouldEqual, "https://cdn.ons.gov.uk/assets/images/ons-logo/kitemark/v2/uksa-kitemark-en.svg")
    })

    Convey("Given enableOfficialStatisticsLogo is true and useSvg is false", t, func() {
        result := GetOfficialStatisticsLogo(true, false, "en")
        So(result, ShouldEqual, "https://cdn.ons.gov.uk/assets/images/ons-logo/kitemark/v2/uksa-kitemark-en.png")
    })

    Convey("Given enableOfficialStatisticsLogo is false and useSvg is true", t, func() {
        result := GetOfficialStatisticsLogo(false, true, "en")
        So(result, ShouldEqual, "https://cdn.ons.gov.uk/assets/images/ons-logo/kitemark/uksa-kitemark.svg")
    })

    Convey("Given enableOfficialStatisticsLogo is false and useSvg is false", t, func() {
        result := GetOfficialStatisticsLogo(false, false, "en")
        So(result, ShouldEqual, "/img/national-statistics.png")
    })

	Convey("Given enableOfficialStatisticsLogo is true, useSvg is true, and language is 'cy'", t, func() {
		result := GetOfficialStatisticsLogo(true, true, "cy")
		So(result, ShouldEqual, "https://cdn.ons.gov.uk/assets/images/ons-logo/kitemark/v2/uksa-kitemark-cy.svg")
	})
	
	Convey("Given enableOfficialStatisticsLogo is true, useSvg is false, and language is 'cy'", t, func() {
		result := GetOfficialStatisticsLogo(true, false, "cy")
		So(result, ShouldEqual, "https://cdn.ons.gov.uk/assets/images/ons-logo/kitemark/v2/uksa-kitemark-cy.png")
	})
}
