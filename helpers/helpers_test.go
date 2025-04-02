package helpers

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/ONSdigital/dp-dataset-api/models"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/osrlogo"
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
		So(GenerateSharingLink("x", url, title), ShouldResemble, fmt.Sprintf("https://x.com/intent/tweet?original_referer&text=%s&url=%s", title, url))
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

func TestGetOSRLogoDetails(t *testing.T) {
	Convey("Given language is 'en'", t, func() {
		result := GetOSRLogoDetails("en")
		So(result, ShouldResemble, osrlogo.OSRLogo{
			URL:     "https://cdn.ons.gov.uk/assets/images/ons-logo/kitemark/v2/uksa-kitemark-en.svg",
			AltText: "Official Statistics logo",
			Title:   "Accredited official statistics",
			About:   "Confirmed by the Office for Statistics Regulation as compliant with the Code of Practice for Statistics.",
		})
	})

	Convey("Given language is 'cy'", t, func() {
		result := GetOSRLogoDetails("cy")
		So(result, ShouldResemble, osrlogo.OSRLogo{
			URL:     "https://cdn.ons.gov.uk/assets/images/ons-logo/kitemark/v2/uksa-kitemark-cy.svg",
			AltText: "Official Statistics logo",
			Title:   "Accredited official statistics",
			About:   "Confirmed by the Office for Statistics Regulation as compliant with the Code of Practice for Statistics.",
		})
	})
}

// Tests for the `GetDistributionFileUrl` helper function
func TestGetDistributionFileUrl(t *testing.T) {
	distributionCsv := models.Distribution{
		Title:       "CSV file",
		Format:      models.DistributionFormatCSV,
		MediaType:   models.DistributionMediaTypeCSV,
		DownloadURL: "http://localhost:22000/datasets/123/editions/2017/versions/1.csv",
		ByteSize:    1234,
	}
	distributionXls := models.Distribution{
		Title:       "XLS file",
		Format:      models.DistributionFormatXLS,
		MediaType:   models.DistributionMediaTypeXLS,
		DownloadURL: "http://localhost:22000/datasets/123/editions/2017/versions/1.xls",
		ByteSize:    1234,
	}

	Convey("Test function returns empty string if input distributions is empty", t, func() {
		distributionList := []models.Distribution{}
		requestedFormat := "xls"

		result := GetDistributionFileUrl(&distributionList, requestedFormat)
		So(result, ShouldEqual, "")
	})
	Convey("Test function returns empty string if requested format doesn't match", t, func() {

		distributionList := []models.Distribution{distributionCsv}
		requestedFormat := "xls"

		// Just confirm requested format doesn't match the distribution format
		So(requestedFormat, ShouldNotEqual, distributionCsv.Format.String())
		result := GetDistributionFileUrl(&distributionList, requestedFormat)
		So(result, ShouldEqual, "")
	})
	Convey("Test function returns valid url if requested format matches distribution", t, func() {
		distributionList := []models.Distribution{
			distributionCsv,
			distributionXls,
		}
		requestedFormat := "csv"

		// Just confirm requested format matches the distribution format
		So(requestedFormat, ShouldEqual, distributionCsv.Format.String())
		result := GetDistributionFileUrl(&distributionList, requestedFormat)
		So(result, ShouldEqual, distributionCsv.DownloadURL)
	})
	Convey("Test function returns empty string if requested format is empty string", t, func() {
		distributionList := []models.Distribution{
			distributionCsv,
			distributionXls,
		}
		requestedFormat := ""

		result := GetDistributionFileUrl(&distributionList, requestedFormat)
		So(result, ShouldEqual, "")
	})
}

// Tests for the `GetDownloadFileUrl` helper function
func TestGetDownloadFileUrl(t *testing.T) {
	downloadObjectCsv := models.DownloadObject{
		HRef: "https://www.aws/123.csv",
		Size: "25",
	}
	downloadObjectXls := models.DownloadObject{
		HRef: "https://www.aws/1234.xls",
		Size: "25",
	}

	Convey("Test function returns empty string if input downloads is empty", t, func() {
		downloadList := models.DownloadList{}
		requestedFormat := "xls"

		result := GetDownloadFileUrl(&downloadList, requestedFormat)
		So(result, ShouldEqual, "")
	})
	Convey("Test function returns empty string if requested format doesn't match", t, func() {

		downloadList := models.DownloadList{
			CSV: &downloadObjectCsv,
		}
		requestedFormat := "xls"

		result := GetDownloadFileUrl(&downloadList, requestedFormat)
		So(result, ShouldEqual, "")
	})
	Convey("Test function returns valid url if requested format matches distribution", t, func() {
		downloadList := models.DownloadList{
			CSV: &downloadObjectCsv,
			XLS: &downloadObjectXls,
		}
		requestedFormat := "csv"

		result := GetDownloadFileUrl(&downloadList, requestedFormat)
		So(result, ShouldEqual, downloadList.CSV.HRef)
	})
	Convey("Test function returns empty string if requested format is empty string", t, func() {
		downloadList := models.DownloadList{
			CSV: &downloadObjectCsv,
			XLS: &downloadObjectXls,
		}
		requestedFormat := ""

		result := GetDownloadFileUrl(&downloadList, requestedFormat)
		So(result, ShouldEqual, "")
	})
}

// Tests for the `MapVersionDownloads` helper function
func TestMapVersionDownloads(t *testing.T) {

	sharedModelVersion := sharedModel.Version{}

	Convey("Test function doesn't update `sharedModelVersion.Downloads` if input `DownloadList` is empty", t, func() {
		downloadList := models.DownloadList{}

		MapVersionDownloads(&sharedModelVersion, &downloadList)
		So(sharedModelVersion.Downloads, ShouldBeEmpty)
	})
	Convey("Test function doesn't update `sharedModelVersion.Downloads` if input `DownloadList` is initialised with empty `DownloadObjects`", t, func() {
		downloadList := models.DownloadList{
			XLS:  &models.DownloadObject{},
			XLSX: &models.DownloadObject{},
			CSV:  &models.DownloadObject{},
			TXT:  &models.DownloadObject{},
			CSVW: &models.DownloadObject{},
		}

		MapVersionDownloads(&sharedModelVersion, &downloadList)
		So(sharedModelVersion.Downloads, ShouldBeEmpty)
	})

	Convey("Test function updates `sharedModelVersion.Downloads` if input `DownloadList` contains a valid `DownloadObject`", t, func() {
		xlsHref := "https://www.test.com/my-xls-download.xlx"
		xlsSize := "1234"

		downloadList := models.DownloadList{
			XLS: &models.DownloadObject{
				HRef: xlsHref,
				Size: xlsSize,
			},
			XLSX: &models.DownloadObject{},
			CSV:  &models.DownloadObject{},
			TXT:  &models.DownloadObject{},
			CSVW: &models.DownloadObject{},
		}

		MapVersionDownloads(&sharedModelVersion, &downloadList)
		So(sharedModelVersion.Downloads, ShouldNotBeEmpty)
		// Just check that only one download is returned, xls could have matched to xlsx too
		So(len(sharedModelVersion.Downloads), ShouldEqual, 1)
		So(sharedModelVersion.Downloads[0].Extension, ShouldEqual, "xls")
		So(sharedModelVersion.Downloads[0].Size, ShouldEqual, xlsSize)
		So(sharedModelVersion.Downloads[0].URI, ShouldEqual, xlsHref)
	})
}
