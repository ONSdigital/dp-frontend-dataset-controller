package mapper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/zebedee"
	"github.com/ONSdigital/dp-renderer/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitMapper(t *testing.T) {
	ctx := context.Background()
	req := httptest.NewRequest("", "/", nil)
	mdl := model.Page{}

	nomisRefURL := "https://www.nomisweb.co.uk/census/2011/ks101ew"
	contact := dataset.Contact{
		Name:      "Matt Rout",
		Telephone: "01111 222222",
		Email:     "mattrout@test.com",
	}
	d := dataset.DatasetDetails{
		CollectionID: "abcdefg",
		Contacts: &[]dataset.Contact{
			contact,
		},
		Description: "A really awesome dataset for you to look at",
		Links: dataset.Links{
			Self: dataset.Link{
				URL: "/datasets/83jd98fkflg",
			},
		},
		NextRelease:      "11-11-2018",
		ReleaseFrequency: "Yearly",
		Publisher: &dataset.Publisher{
			URL:  "ons.gov.uk",
			Name: "ONS",
			Type: "Government Agency",
		},
		State:   "created",
		Theme:   "purple",
		Title:   "Penguins of the Antarctic Ocean",
		License: "ons",
	}
	nomisD := dataset.DatasetDetails{
		CollectionID: "abcdefg",
		Contacts: &[]dataset.Contact{
			contact,
		},
		Description: "A really awesome dataset for you to look at",
		Links: dataset.Links{
			Self: dataset.Link{
				URL: "/datasets/83jd98fkflg",
			},
		},
		NextRelease:      "11-11-2018",
		ReleaseFrequency: "Yearly",
		Publisher: &dataset.Publisher{
			URL:  "ons.gov.uk",
			Name: "ONS",
			Type: "Government Agency",
		},
		State:             "created",
		Theme:             "purple",
		Title:             "Penguins of the Antarctic Ocean",
		License:           "ons",
		Type:              "nomis",
		NomisReferenceURL: nomisRefURL,
	}

	v := []dataset.Version{
		{
			CollectionID: "abcdefg",
			Edition:      "2017",
			ID:           "tehnskofjios-ashbc7",
			InstanceID:   "31241592",
			Version:      1,
			Links: dataset.Links{
				Self: dataset.Link{
					URL: "/datasets/83jd98fkflg/editions/124/versions/1",
				},
			},
			ReleaseDate: "11-11-2017",
			State:       "published",
			Downloads: map[string]dataset.Download{
				"XLSX": {
					Size: "438290",
					URL:  "my-url",
				},
			},
		},
	}

	datasetID := "038847784-2874757-23784854905"

	Convey("test CreateFilterableLandingPage for CMD pages", t, func() {
		// breadcrumbItem returned by zebedee after being proxied through API router
		breadcrumbItem0 := zebedee.Breadcrumb{
			URI:         "https://myHost:1234/v1/economy/grossdomesticproduct/datasets/gdpjanuary2018",
			Description: zebedee.NodeDescription{Title: "GDP: January 2018"},
		}

		// breadcrumbItem as expected as a result of CreateFilterableLandingPage
		expectedBreadcrumbItem0 := zebedee.Breadcrumb{
			URI:         "https://myHost:1234/economy/grossdomesticproduct/datasets/gdpjanuary2018",
			Description: zebedee.NodeDescription{Title: "GDP: January 2018"},
		}

		// breadcrumbItem returned by zebedee directly (without proxying through API router)
		breadcrumbItem1 := zebedee.Breadcrumb{
			URI:         "/economy/grossdomesticproduct/datasets/gdpjanuary2019",
			Description: zebedee.NodeDescription{Title: "GDP: January 2019"},
		}
		expectedBreadcrumbItem1 := breadcrumbItem1

		// breadcrumbItemWrongURI with wrong URI value
		breadcrumbItemWrongURI := zebedee.Breadcrumb{
			URI:         "/v1/%&*$^@$(@!@±£8",
			Description: zebedee.NodeDescription{Title: "Something wrong"},
		}
		expectedBreadcrumbItemWrongURI := breadcrumbItemWrongURI

		p := CreateFilterableLandingPage(mdl, ctx, req, d, v[0], datasetID, []dataset.Options{
			{
				Items: []dataset.Option{
					{
						DimensionID: "age",
						Label:       "6",
						Option:      "6",
					},
					{
						DimensionID: "age",
						Label:       "3",
						Option:      "3",
					},
					{
						DimensionID: "age",
						Label:       "24",
						Option:      "24",
					},
					{
						DimensionID: "age",
						Label:       "23",
						Option:      "23",
					},
					{
						DimensionID: "age",
						Label:       "19",
						Option:      "19",
					},
				},
			},
			{
				Items: []dataset.Option{
					{
						DimensionID: "time",
						Label:       "Jan-05",
						Option:      "Jan-05",
					},
					{
						DimensionID: "time",
						Label:       "Feb-05",
						Option:      "Feb-05",
					},
				},
			},
		}, dataset.VersionDimensions{}, false, []zebedee.Breadcrumb{breadcrumbItem0, breadcrumbItem1, breadcrumbItemWrongURI},
			1, "/datasets/83jd98fkflg/editions/124/versions/1", "en", "/v1", 50)

		So(p.Type, ShouldEqual, "dataset_landing_page")
		So(p.Metadata.Title, ShouldEqual, d.Title)
		So(p.URI, ShouldEqual, req.URL.Path)
		So(p.ContactDetails.Name, ShouldEqual, contact.Name)
		So(p.ContactDetails.Telephone, ShouldEqual, contact.Telephone)
		So(p.ContactDetails.Email, ShouldEqual, contact.Email)
		So(p.DatasetLandingPage.NextRelease, ShouldEqual, d.NextRelease)
		So(p.DatasetLandingPage.DatasetID, ShouldEqual, datasetID)
		So(p.ReleaseDate, ShouldEqual, v[0].ReleaseDate)
		So(p.Breadcrumb[0].Title, ShouldEqual, expectedBreadcrumbItem0.Description.Title)
		So(p.Breadcrumb[0].URI, ShouldEqual, expectedBreadcrumbItem0.URI)
		So(p.Breadcrumb[1].Title, ShouldEqual, expectedBreadcrumbItem1.Description.Title)
		So(p.Breadcrumb[1].URI, ShouldEqual, expectedBreadcrumbItem1.URI)
		So(p.Breadcrumb[2].Title, ShouldEqual, expectedBreadcrumbItemWrongURI.Description.Title)
		So(p.Breadcrumb[2].URI, ShouldEqual, expectedBreadcrumbItemWrongURI.URI)

		So(p.DatasetLandingPage.Dimensions, ShouldHaveLength, 2)
		So(p.DatasetLandingPage.Dimensions[0].Title, ShouldEqual, "Age")
		So(p.DatasetLandingPage.Dimensions[0].Values, ShouldHaveLength, 5)
		So(p.DatasetLandingPage.Dimensions[0].Values[0], ShouldEqual, "3")
		So(p.DatasetLandingPage.Dimensions[0].Values[1], ShouldEqual, "6")
		So(p.DatasetLandingPage.Dimensions[0].Values[2], ShouldEqual, "19")
		So(p.DatasetLandingPage.Dimensions[0].Values[3], ShouldEqual, "23")
		So(p.DatasetLandingPage.Dimensions[0].Values[4], ShouldEqual, "24")
		So(p.DatasetLandingPage.Dimensions[1].Values, ShouldHaveLength, 1)
		So(p.DatasetLandingPage.Dimensions[1].Title, ShouldEqual, "Time")
		So(p.DatasetLandingPage.Dimensions[1].Values[0], ShouldEqual, "All months between January 2005 and February 2005")

		v0 := p.DatasetLandingPage.Version
		So(v0.Title, ShouldEqual, d.Title)
		So(v0.Description, ShouldEqual, d.Description)
		So(v0.Edition, ShouldEqual, v[0].Edition)
		So(v0.Version, ShouldEqual, strconv.Itoa(v[0].Version))
		So(p.ReleaseDate, ShouldEqual, v[0].ReleaseDate)
		So(v0.Downloads[0].Size, ShouldEqual, "438290")
		So(v0.Downloads[0].Extension, ShouldEqual, "XLSX")
		So(v0.Downloads[0].URI, ShouldEqual, "my-url")
	})

	Convey("test CreateFilterableLandingPage for Nomis pages", t, func() {
		// breadcrumbItem returned by zebedee after being proxied through API router
		breadcrumbItem0 := zebedee.Breadcrumb{
			URI:         "https://myHost:1234/v1/economy/grossdomesticproduct/datasets/gdpjanuary2018",
			Description: zebedee.NodeDescription{Title: "GDP: January 2018"},
		}

		// breadcrumbItem as expected as a result of CreateFilterableLandingPage
		expectedBreadcrumbItem0 := zebedee.Breadcrumb{
			URI:         "/",
			Description: zebedee.NodeDescription{Title: "Home"},
		}

		// breadcrumbItem returned by zebedee directly (without proxying through API router)
		breadcrumbItem1 := zebedee.Breadcrumb{
			URI:         "https://myHost:1234/economy/grossdomesticproduct/datasets/gdpjanuary2018",
			Description: zebedee.NodeDescription{Title: "GDP: January 2018"},
		}
		expectedBreadcrumbItem1 := breadcrumbItem1

		// breadcrumbItemWrongURI with wrong URI value
		breadcrumbItemWrongURI := zebedee.Breadcrumb{
			URI:         "/v1/%&*$^@$(@!@±£8",
			Description: zebedee.NodeDescription{Title: "Something wrong"},
		}

		p := CreateFilterableLandingPage(mdl, ctx, req, nomisD, v[0], datasetID, []dataset.Options{
			{
				Items: []dataset.Option{
					{
						DimensionID: "age",
						Label:       "6",
						Option:      "6",
					},
					{
						DimensionID: "age",
						Label:       "3",
						Option:      "3",
					},
					{
						DimensionID: "age",
						Label:       "24",
						Option:      "24",
					},
					{
						DimensionID: "age",
						Label:       "23",
						Option:      "23",
					},
					{
						DimensionID: "age",
						Label:       "19",
						Option:      "19",
					},
				},
			},
			{
				Items: []dataset.Option{
					{
						DimensionID: "time",
						Label:       "Jan-05",
						Option:      "Jan-05",
					},
					{
						DimensionID: "time",
						Label:       "Feb-05",
						Option:      "Feb-05",
					},
				},
			},
		}, dataset.VersionDimensions{}, false, []zebedee.Breadcrumb{breadcrumbItem0, breadcrumbItem1, breadcrumbItemWrongURI},
			1, "/datasets/83jd98fkflg/editions/124/versions/1", "en", "/v1", 50)

		So(p.Type, ShouldEqual, "dataset_landing_page")
		So(p.Metadata.Title, ShouldEqual, d.Title)
		So(p.URI, ShouldEqual, req.URL.Path)
		So(p.ContactDetails.Name, ShouldEqual, contact.Name)
		So(p.ContactDetails.Telephone, ShouldEqual, contact.Telephone)
		So(p.ContactDetails.Email, ShouldEqual, contact.Email)
		So(p.DatasetLandingPage.NextRelease, ShouldEqual, d.NextRelease)
		So(p.DatasetLandingPage.DatasetID, ShouldEqual, datasetID)
		So(p.ReleaseDate, ShouldEqual, v[0].ReleaseDate)
		So(p.Breadcrumb[0].Title, ShouldEqual, expectedBreadcrumbItem0.Description.Title)
		So(p.Breadcrumb[0].URI, ShouldEqual, expectedBreadcrumbItem0.URI)
		So(p.Breadcrumb[1].Title, ShouldEqual, expectedBreadcrumbItem1.Description.Title)
		So(p.Breadcrumb[1].URI, ShouldEqual, expectedBreadcrumbItem1.URI)

		So(p.DatasetLandingPage.NomisReferenceURL, ShouldEqual, nomisRefURL)

		v0 := p.DatasetLandingPage.Version
		So(v0.Title, ShouldEqual, d.Title)
		So(v0.Description, ShouldEqual, d.Description)
		So(v0.Edition, ShouldEqual, v[0].Edition)
		So(v0.Version, ShouldEqual, strconv.Itoa(v[0].Version))
		So(p.ReleaseDate, ShouldEqual, v[0].ReleaseDate)
	})

	Convey("test time dimensions when parsing Jan-06 format for CreateFilterableLandingPage ", t, func() {

		p := CreateFilterableLandingPage(mdl, ctx, req, d, v[0], datasetID, []dataset.Options{
			{
				Items: []dataset.Option{
					{
						DimensionID: "time",
						Label:       "Jan-05",
						Option:      "Jan-05",
					},
					{
						DimensionID: "time",
						Label:       "May-07",
						Option:      "May-07",
					},
					{
						DimensionID: "time",
						Label:       "Jun-07",
						Option:      "Jun-07",
					},
				},
			},
		}, dataset.VersionDimensions{}, false, []zebedee.Breadcrumb{},
			1, "/datasets/83jd98fkflg/editions/124/versions/1", "en", "/v1", 50)

		So(p.Type, ShouldEqual, "dataset_landing_page")
		So(p.DatasetLandingPage.Dimensions[0].Values, ShouldHaveLength, 2)
		So(p.DatasetLandingPage.Dimensions[0].Title, ShouldEqual, "Time")
		So(p.DatasetLandingPage.Dimensions[0].Values[0], ShouldEqual, "This year 2005 contains data for the month January")
		So(p.DatasetLandingPage.Dimensions[0].Values[1], ShouldEqual, "All months between May 2007 and June 2007")

	})

	Convey("test time dimensions for CreateFilterableLandingPage ", t, func() {

		p := CreateFilterableLandingPage(mdl, ctx, req, d, v[0], datasetID, []dataset.Options{
			{
				Items: []dataset.Option{
					{
						DimensionID: "time",
						Label:       "2016",
						Option:      "2016",
					},
					{
						DimensionID: "time",
						Label:       "2018",
						Option:      "2018",
					},
					{
						DimensionID: "time",
						Label:       "2019",
						Option:      "2019",
					},
					{
						DimensionID: "time",
						Label:       "2020",
						Option:      "2020",
					},
				},
			},
		}, dataset.VersionDimensions{}, false, []zebedee.Breadcrumb{},
			1, "/datasets/83jd98fkflg/editions/124/versions/1", "en", "/v1", 50)

		So(p.Type, ShouldEqual, "dataset_landing_page")
		So(p.DatasetLandingPage.Dimensions[0].Values, ShouldHaveLength, 2)
		So(p.DatasetLandingPage.Dimensions[0].Title, ShouldEqual, "Time")
		So(p.DatasetLandingPage.Dimensions[0].Values[0], ShouldEqual, "This year contains data for 2016")
		So(p.DatasetLandingPage.Dimensions[0].Values[1], ShouldEqual, "All years between 2018 and 2020")

	})

}

// TestCreateVersionsList Tests the CreateVersionsList function in the mapper
func TestCreateVersionsList(t *testing.T) {
	mdl := model.Page{}
	req := httptest.NewRequest("", "/", nil)
	dummyModelData := dataset.DatasetDetails{
		ID:    "cpih01",
		Title: "Consumer Prices Index including owner occupiers? housing costs (CPIH)",
		Links: dataset.Links{
			Editions: dataset.Link{
				URL: "http://localhost:22000/datasets/cpih01/editions",
				ID:  ""},
			LatestVersion: dataset.Link{
				URL: "http://localhost:22000/datasets/cpih01/editions/time-series/versions/3",
				ID:  "3"},
			Self: dataset.Link{
				URL: "http://localhost:22000/datasets/cpih01",
				ID:  ""},
			Taxonomy: dataset.Link{
				URL: "/economy/environmentalaccounts/datasets/consumerpricesindexincludingowneroccupiershousingcostscpih",
				ID:  ""},
		},
	}
	dummyEditionData := dataset.Edition{}
	dummyVersion1 := dataset.Version{
		Alerts:        nil,
		CollectionID:  "",
		Downloads:     nil,
		Edition:       "time-series",
		Dimensions:    nil,
		ID:            "",
		InstanceID:    "",
		LatestChanges: nil,
		Links: dataset.Links{
			Dataset: dataset.Link{
				URL: "http://localhost:22000/datasets/cpih01",
				ID:  "cpih01",
			},
		},
		ReleaseDate: "2019-08-15T00:00:00.000Z",
		State:       "published",
		Temporal:    nil,
		Version:     1,
	}
	dummyVersion2 := dummyVersion1
	dummyVersion2.Version = 2
	dummyVersion3 := dummyVersion1
	dummyVersion3.Version = 3

	Convey("test latest version page", t, func() {
		dummySingleVersionList := []dataset.Version{dummyVersion3}

		page := CreateVersionsList(mdl, req, dummyModelData, dummyEditionData, dummySingleVersionList)
		Convey("title", func() {
			So(page.Metadata.Title, ShouldEqual, "All versions of Consumer Prices Index including owner occupiers? housing costs (CPIH) time-series dataset")
		})
		Convey("has correct number of versions when only one should be present", func() {
			So(page.Data.Versions, ShouldHaveLength, 1)
		})

		dummyMultipleVersionList := []dataset.Version{dummyVersion1, dummyVersion2, dummyVersion3}
		page = CreateVersionsList(mdl, req, dummyModelData, dummyEditionData, dummyMultipleVersionList)

		Convey("has correct number of versions when multiple should be present", func() {
			So(page.Data.Versions, ShouldHaveLength, 3)
		})
		Convey("is latest version correctly tagged", func() {
			So(page.Data.Versions[0].IsLatest, ShouldEqual, true)
			So(page.Data.Versions[1].IsLatest, ShouldEqual, false)
			So(page.Data.Versions[2].IsLatest, ShouldEqual, false)
		})
		Convey("are version numbers accurate", func() {
			So(page.Data.Versions[0].VersionNumber, ShouldEqual, 3)
			So(page.Data.Versions[1].VersionNumber, ShouldEqual, 2)
			So(page.Data.Versions[2].VersionNumber, ShouldEqual, 1)
		})
		Convey("superseded links accurate", func() {
			So(page.Data.Versions[0].Superseded, ShouldEqual, "/datasets/cpih01/editions/time-series/versions/2")
			So(page.Data.Versions[1].Superseded, ShouldEqual, "/datasets/cpih01/editions/time-series/versions/1")
			So(page.Data.Versions[2].Superseded, ShouldEqual, "")
		})
	})
}

func TestUnitMapCookiesPreferences(t *testing.T) {
	req := httptest.NewRequest("", "/", nil)
	pageModel := model.Page{
		CookiesPreferencesSet: false,
		CookiesPolicy: model.CookiesPolicy{
			Essential: false,
			Usage:     false,
		},
	}

	Convey("maps cookies preferences cookie data to page model correctly", t, func() {
		So(pageModel.CookiesPreferencesSet, ShouldEqual, false)
		So(pageModel.CookiesPolicy.Essential, ShouldEqual, false)
		So(pageModel.CookiesPolicy.Usage, ShouldEqual, false)
		req.AddCookie(&http.Cookie{Name: "cookies_preferences_set", Value: "true"})
		req.AddCookie(&http.Cookie{Name: "cookies_policy", Value: "%7B%22essential%22%3Atrue%2C%22usage%22%3Atrue%7D"})
		MapCookiePreferences(req, &pageModel.CookiesPreferencesSet, &pageModel.CookiesPolicy)
		So(pageModel.CookiesPreferencesSet, ShouldEqual, true)
		So(pageModel.CookiesPolicy.Essential, ShouldEqual, true)
		So(pageModel.CookiesPolicy.Usage, ShouldEqual, true)
	})
}

func TestCreateCensusDatasetLandingPage(t *testing.T) {
	req := httptest.NewRequest("", "/", nil)
	pageModel := model.Page{}
	contact := dataset.Contact{
		Telephone: "01232 123 123",
		Email:     "hello@testing.com",
	}
	methodology := dataset.Methodology{
		Description: "An interesting methodology description",
		URL:         "http://www.google.com",
		Title:       "The methodology title",
	}
	datasetModel := dataset.DatasetDetails{
		Contacts: &[]dataset.Contact{
			contact,
		},
		ID:          "12345",
		Description: "An interesting test description",
		Methodologies: &[]dataset.Methodology{
			methodology,
		},
		Title: "Test title",
		Type:  "cantabular",
	}

	versionOneDetails := dataset.Version{
		ReleaseDate: "01-01-2021",
		Downloads: map[string]dataset.Download{
			"XLSX": {
				Size:   "438290",
				Public: "my-url",
			},
		},
	}

	versionTwoDetails := dataset.Version{
		ReleaseDate: "15-02-2021",
	}

	datasetOptions := []dataset.Options{
		{
			Items: []dataset.Option{
				{
					DimensionID: "time",
					Label:       "2016",
					Option:      "2016",
				},
				{
					DimensionID: "time",
					Label:       "2018",
					Option:      "2018",
				},
				{
					DimensionID: "time",
					Label:       "2019",
					Option:      "2019",
				},
				{
					DimensionID: "time",
					Label:       "2020",
					Option:      "2020",
				},
			},
		},
	}

	Convey("Census dataset landing page maps correctly as version 1", t, func() {
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionOneDetails, datasetOptions, dataset.VersionDimensions{}, "", false, "", 50)
		So(page.Type, ShouldEqual, datasetModel.Type)
		So(page.ID, ShouldEqual, datasetModel.ID)
		So(page.Version.ReleaseDate, ShouldEqual, versionOneDetails.ReleaseDate)
		So(page.InitialReleaseDate, ShouldEqual, page.Version.ReleaseDate)
		So(page.DatasetLandingPage.HasOtherVersions, ShouldEqual, false)
		So(page.Version.Downloads[0].Size, ShouldEqual, "438290")
		So(page.Version.Downloads[0].Extension, ShouldEqual, "XLSX")
		So(page.Version.Downloads[0].URI, ShouldEqual, "my-url")
		So(page.Version.Downloads[0].Name, ShouldEqual, "test-title.xlsx")
		So(page.Metadata.Title, ShouldEqual, datasetModel.Title)
		So(page.Metadata.Description, ShouldEqual, datasetModel.Description)
		So(page.ContactDetails.Name, ShouldEqual, contact.Name)
		So(page.ContactDetails.Email, ShouldEqual, contact.Email)
		So(page.ContactDetails.Telephone, ShouldEqual, contact.Telephone)
		So(page.HasContactDetails, ShouldBeTrue)
		So(page.DatasetLandingPage.Methodologies[0].Description, ShouldEqual, methodology.Description)
		So(page.DatasetLandingPage.Methodologies[0].Title, ShouldEqual, methodology.Title)
		So(page.DatasetLandingPage.Methodologies[0].URL, ShouldEqual, methodology.URL)
	})

	Convey("Release date and hasOtherVersions is mapped correctly when v2 of Census DLP dataset is loaded", t, func() {
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionTwoDetails, datasetOptions, dataset.VersionDimensions{}, versionOneDetails.ReleaseDate, true, "", 50)
		So(page.InitialReleaseDate, ShouldEqual, versionOneDetails.ReleaseDate)
		So(page.Version.ReleaseDate, ShouldEqual, versionTwoDetails.ReleaseDate)
		So(page.DatasetLandingPage.HasOtherVersions, ShouldEqual, true)
	})

	Convey("Filename builder should not add hyphens when no spaces are provided", t, func() {
		datasetModel := dataset.DatasetDetails{
			Title: "NoSPACES",
		}
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionOneDetails, datasetOptions, dataset.VersionDimensions{}, versionOneDetails.ReleaseDate, false, "", 50)
		So(page.Version.Downloads[0].Name, ShouldEqual, "nospaces.xlsx")
	})

	Convey("Filename builder should trim whitespace before and after string", t, func() {
		datasetModel := dataset.DatasetDetails{
			Title: " spaces before and after ",
		}
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionOneDetails, datasetOptions, dataset.VersionDimensions{}, versionOneDetails.ReleaseDate, false, "", 50)
		So(page.Version.Downloads[0].Name, ShouldEqual, "spaces-before-and-after.xlsx")
	})

	Convey("Filename builder should suffix file extension in lower case", t, func() {
		versionDetails := dataset.Version{
			Downloads: map[string]dataset.Download{
				"XLSX": {
					Size:   "438290",
					Public: "my-url",
				},
			},
		}
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionDetails, datasetOptions, dataset.VersionDimensions{}, versionOneDetails.ReleaseDate, false, "", 50)
		So(page.Version.Downloads[0].Name, ShouldEqual, "test-title.xlsx")
	})

	noContact := dataset.Contact{
		Telephone: "",
		Email:     "",
	}
	noContactDM := dataset.DatasetDetails{
		Contacts: &[]dataset.Contact{
			noContact,
		},
	}

	Convey("No contacts provided, contact section is not displayed", t, func() {
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, noContactDM, versionOneDetails, datasetOptions, dataset.VersionDimensions{}, "", false, "", 50)
		So(page.ContactDetails.Email, ShouldEqual, noContact.Email)
		So(page.ContactDetails.Telephone, ShouldEqual, noContact.Telephone)
		So(page.HasContactDetails, ShouldBeFalse)
	})

	oneContactDetail := dataset.Contact{
		Telephone: "123",
		Email:     "",
	}
	oneContactDetailDM := dataset.DatasetDetails{
		Contacts: &[]dataset.Contact{
			oneContactDetail,
		},
	}

	Convey("One contact detail provided, contact section is displayed", t, func() {
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, oneContactDetailDM, versionOneDetails, datasetOptions, dataset.VersionDimensions{}, "", false, "", 50)
		So(page.ContactDetails.Email, ShouldEqual, oneContactDetail.Email)
		So(page.ContactDetails.Telephone, ShouldEqual, oneContactDetail.Telephone)
		So(page.HasContactDetails, ShouldBeTrue)
	})
}
