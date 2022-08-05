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

func TestDatasetVerionURL(t *testing.T) {
	Convey("The dataset version URL is correctly constructed from the provided parameters", t, func() {
		So(DatasetVersionUrl("myDataset", "myEdition", "myVersion"), ShouldResemble,
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
		So(GetCurrentUrl("en", "mydomain.com", "/page1/page2"), ShouldResemble, "mydomain.com/page1/page2")
		So(GetCurrentUrl("en", "mydomain.com", ""), ShouldResemble, "mydomain.com")
		So(GetCurrentUrl("cy", "mydomain.com", ""), ShouldResemble, "cy.mydomain.com")
		So(GetCurrentUrl("cy", "mydomain.com", "/page1"), ShouldResemble, "cy.mydomain.com/page1")
		So(GetCurrentUrl("en", "localhost", "/page1"), ShouldResemble, "ons.gov.uk/page1")
		So(GetCurrentUrl("cy", "localhost", "/page1"), ShouldResemble, "cy.ons.gov.uk/page1")
		So(GetCurrentUrl("en", "", "/page1"), ShouldResemble, "ons.gov.uk/page1")
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
	Convey("no string given and empty array returns false", t, func() {
		So(HasStringInSlice("", []string{}), ShouldBeFalse)
	})
	Convey("no string given valid string array returns false", t, func() {
		So(HasStringInSlice("", []string{"hello", "world"}), ShouldBeFalse)
	})
	Convey("valid string given empty array returns false", t, func() {
		So(HasStringInSlice("hello", []string{}), ShouldBeFalse)
	})
	Convey("valid string given with valid string array returns true", t, func() {
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
