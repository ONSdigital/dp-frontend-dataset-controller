package mapper

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
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
			Dimensions: []dataset.VersionDimension{
				{
					ID:    "city",
					Name:  "geography",
					Label: "City",
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

	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()

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
			1, "/datasets/83jd98fkflg/editions/124/versions/1", "en", "/v1", 50, serviceMessage, emergencyBanner)

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

		So(p.ServiceMessage, ShouldEqual, serviceMessage)

		So(p.EmergencyBanner.Type, ShouldEqual, strings.Replace(emergencyBanner.Type, "_", "-", -1))
		So(p.EmergencyBanner.Title, ShouldEqual, emergencyBanner.Title)
		So(p.EmergencyBanner.Description, ShouldEqual, emergencyBanner.Description)
		So(p.EmergencyBanner.URI, ShouldEqual, emergencyBanner.URI)
		So(p.EmergencyBanner.LinkText, ShouldEqual, emergencyBanner.LinkText)

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
			1, "/datasets/83jd98fkflg/editions/124/versions/1", "en", "/v1", 50, serviceMessage, emergencyBanner)

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

		So(p.ServiceMessage, ShouldEqual, serviceMessage)

		So(p.EmergencyBanner.Type, ShouldEqual, strings.Replace(emergencyBanner.Type, "_", "-", -1))
		So(p.EmergencyBanner.Title, ShouldEqual, emergencyBanner.Title)
		So(p.EmergencyBanner.Description, ShouldEqual, emergencyBanner.Description)
		So(p.EmergencyBanner.URI, ShouldEqual, emergencyBanner.URI)
		So(p.EmergencyBanner.LinkText, ShouldEqual, emergencyBanner.LinkText)

		So(p.DatasetLandingPage.NomisReferenceURL, ShouldEqual, nomisRefURL)

		v0 := p.DatasetLandingPage.Version
		So(v0.Title, ShouldEqual, d.Title)
		So(v0.Description, ShouldEqual, d.Description)
		So(v0.Edition, ShouldEqual, v[0].Edition)
		So(v0.Version, ShouldEqual, strconv.Itoa(v[0].Version))
		So(p.ReleaseDate, ShouldEqual, v[0].ReleaseDate)
	})

	Convey("test CreateFilterableLandingPage dimension options are mapped into landing page dimensions", t, func() {
		const (
			dimensionName        = "geography"
			dimensionID          = "city"
			dimensionLabel       = "City"
			dimensionOptionLabel = "London"
		)

		dims := dataset.VersionDimensions{
			Items: []dataset.VersionDimension{
				{
					ID:    dimensionID,
					Name:  dimensionName,
					Label: dimensionLabel,
				},
			},
		}
		opts := []dataset.Options{
			{
				Items: []dataset.Option{
					{
						DimensionID: dimensionName,
						Label:       dimensionOptionLabel,
						Option:      "0",
					},
				},
				Count:      1,
				TotalCount: 1,
			},
		}

		p := CreateFilterableLandingPage(mdl, ctx, req, d, v[0], datasetID, opts, dims, false, []zebedee.Breadcrumb{},
			1, "", "en", "/v1", 50, serviceMessage, emergencyBanner)

		So(p.DatasetLandingPage.Dimensions, ShouldResemble, []sharedModel.Dimension{
			{
				Title:      dimensionLabel,
				Name:       dimensionName,
				Values:     []string{dimensionOptionLabel},
				OptionsURL: "/dimensions/geography/options",
				TotalItems: 1,
			},
		})
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
			1, "/datasets/83jd98fkflg/editions/124/versions/1", "en", "/v1", 50, serviceMessage, emergencyBanner)

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
			1, "/datasets/83jd98fkflg/editions/124/versions/1", "en", "/v1", 50, serviceMessage, emergencyBanner)

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
	dummyVersion3.Alerts = &[]dataset.Alert{
		{
			Date:        "",
			Description: "This is a correction",
			Type:        "correction",
		},
	}
	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()

	Convey("test latest version page", t, func() {
		dummySingleVersionList := []dataset.Version{dummyVersion3}

		page := CreateVersionsList(mdl, req, dummyModelData, dummyEditionData, dummySingleVersionList, serviceMessage, emergencyBanner)
		Convey("title", func() {
			So(page.Metadata.Title, ShouldEqual, "All versions of Consumer Prices Index including owner occupiers? housing costs (CPIH) time-series dataset")
		})
		Convey("has correct number of versions when only one should be present", func() {
			So(page.Data.Versions, ShouldHaveLength, 1)
		})

		dummyMultipleVersionList := []dataset.Version{dummyVersion1, dummyVersion2, dummyVersion3}
		page = CreateVersionsList(mdl, req, dummyModelData, dummyEditionData, dummyMultipleVersionList, serviceMessage, emergencyBanner)

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
		Convey("correction notice maps when present", func() {
			So(page.Data.Versions[2].Corrections, ShouldBeEmpty)
			So(page.Data.Versions[1].Corrections, ShouldBeEmpty)
			So(page.Data.Versions[0].Corrections[0].Reason, ShouldEqual, "This is a correction")
		})
		Convey("service message maps correctly", func() {
			So(page.ServiceMessage, ShouldEqual, serviceMessage)
		})
		Convey("emergency banner maps correctly", func() {
			So(page.EmergencyBanner.Type, ShouldEqual, strings.Replace(emergencyBanner.Type, "_", "-", -1))
			So(page.EmergencyBanner.Title, ShouldEqual, emergencyBanner.Title)
			So(page.EmergencyBanner.Description, ShouldEqual, emergencyBanner.Description)
			So(page.EmergencyBanner.URI, ShouldEqual, emergencyBanner.URI)
			So(page.EmergencyBanner.LinkText, ShouldEqual, emergencyBanner.LinkText)
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
		Description: "An interesting test description \n with a line break",
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
				Size: "438290",
				URL:  "https://mydomain.com/my-request",
			},
		},
		Edition: "2021",
		Version: 1,
		Links: dataset.Links{
			Dataset: dataset.Link{
				URL: "http://localhost:22000/datasets/cantabular-1",
				ID:  "cantabular-1",
			},
		},
		Dimensions: []dataset.VersionDimension{
			{
				Description: "A description on one line",
				Name:        "Dimension 1",
				ID:          "dim_1",
				IsAreaType:  helpers.ToBoolPtr(true),
			},
			{
				Description: "A description on one line \n Then a line break",
				Name:        "Dimension 2",
				ID:          "dim_2",
			},
			{
				Description: "",
				Name:        "Only a name - I shouldn't map",
				ID:          "dim_3",
			},
		},
	}

	versionTwoDetails := dataset.Version{
		ReleaseDate: "15-02-2021",
		Version:     2,
		Edition:     "2021",
		Links: dataset.Links{
			Dataset: dataset.Link{
				URL: "http://localhost:22000/datasets/cantabular-1",
				ID:  "cantabular-1",
			},
		},
		Alerts: &[]dataset.Alert{
			{
				Date:        "",
				Description: "This is a correction",
				Type:        "correction",
			},
		},
	}

	versionThreeDetails := versionTwoDetails
	versionThreeDetails.Version = 3

	datasetOptions := []dataset.Options{
		{
			Items: []dataset.Option{
				{
					DimensionID: "dim_1",
					Option:      "option 1",
				},
				{
					DimensionID: "dim_1",
					Option:      "option 2",
				},
			},
		},
	}

	filterOutput := filter.Model{
		Dimensions: []filter.ModelDimension{
			{
				Label:   "A label",
				Options: []string{"An option", "and another"},
			},
		},
		Downloads: map[string]filter.Download{
			"CSV": {
				Size: "12345",
				URL:  "https://mydomain.com/my-request",
			},
		},
	}

	Convey("Census dataset landing page maps correctly as version 1", t, func() {
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionOneDetails, datasetOptions, "", false, []dataset.Version{versionOneDetails}, 1, "/a/version/1", "", []string{}, 50, false, false, false, filter.Model{})
		So(page.Type, ShouldEqual, datasetModel.Type)
		So(page.ID, ShouldEqual, datasetModel.ID)
		So(page.Version.ReleaseDate, ShouldEqual, versionOneDetails.ReleaseDate)
		So(page.InitialReleaseDate, ShouldEqual, page.Version.ReleaseDate)
		So(page.DatasetLandingPage.HasOtherVersions, ShouldBeFalse)
		So(page.Version.Downloads[0].Size, ShouldEqual, "438290")
		So(page.Version.Downloads[0].Extension, ShouldEqual, "xlsx")
		So(page.Version.Downloads[0].URI, ShouldEqual, "https://mydomain.com/my-request")
		So(page.Version.Downloads, ShouldHaveLength, 1)
		So(page.Metadata.Title, ShouldEqual, datasetModel.Title)
		So(page.Metadata.Description, ShouldEqual, datasetModel.Description)
		So(page.DatasetLandingPage.Description, ShouldResemble, strings.Split(datasetModel.Description, "\n"))
		So(page.ContactDetails.Name, ShouldEqual, contact.Name)
		So(page.ContactDetails.Email, ShouldEqual, contact.Email)
		So(page.ContactDetails.Telephone, ShouldEqual, contact.Telephone)
		So(page.HasContactDetails, ShouldBeTrue)
		So(page.DatasetLandingPage.Methodologies[0].Description, ShouldEqual, methodology.Description)
		So(page.DatasetLandingPage.Methodologies[0].Title, ShouldEqual, methodology.Title)
		So(page.DatasetLandingPage.Methodologies[0].URL, ShouldEqual, methodology.URL)
		So(page.DatasetLandingPage.LatestVersionURL, ShouldBeBlank)
		So(page.Collapsible.CollapsibleItems[0].Subheading, ShouldEqual, versionOneDetails.Dimensions[0].Name)
		So(page.Collapsible.CollapsibleItems[0].Content[0], ShouldEqual, versionOneDetails.Dimensions[0].Description)
		So(page.Collapsible.CollapsibleItems[1].Subheading, ShouldEqual, versionOneDetails.Dimensions[1].Name)
		So(page.Collapsible.CollapsibleItems[1].Content, ShouldResemble, strings.Split(versionOneDetails.Dimensions[1].Description, "\n"))
		So(page.Collapsible.CollapsibleItems, ShouldHaveLength, 2)
		So(page.DatasetLandingPage.IsFlexible, ShouldBeFalse)
		So(page.DatasetLandingPage.Dimensions, ShouldHaveLength, 2) // coverage is inserted
		So(page.DatasetLandingPage.Dimensions[1].IsCoverage, ShouldBeTrue)
		So(page.DatasetLandingPage.Dimensions[1].Title, ShouldEqual, "Coverage")
		So(page.DatasetLandingPage.Dimensions[1].Name, ShouldEqual, "coverage")
		So(page.DatasetLandingPage.Dimensions[1].ShowChange, ShouldBeFalse)
		So(page.DatasetLandingPage.Dimensions[0].ShowChange, ShouldBeFalse)
	})

	Convey("Census dataset landing page maps correctly with filter output", t, func() {
		datasetModel.Type = "flexible"
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionOneDetails, datasetOptions, "", false, []dataset.Version{versionOneDetails}, 1, "/a/version/1", "", []string{}, 50, false, true, true, filterOutput)
		So(page.Type, ShouldEqual, datasetModel.Type)
		So(page.ID, ShouldEqual, datasetModel.ID)
		So(page.Version.ReleaseDate, ShouldEqual, versionOneDetails.ReleaseDate)
		So(page.InitialReleaseDate, ShouldEqual, page.Version.ReleaseDate)
		So(page.DatasetLandingPage.HasOtherVersions, ShouldBeFalse)
		// Downloads are mapped from filterOutput
		So(page.Version.Downloads[0].Size, ShouldEqual, "12345")
		So(page.Version.Downloads[0].Extension, ShouldEqual, "csv")
		So(page.Version.Downloads[0].URI, ShouldEqual, "https://mydomain.com/my-request")
		So(page.Version.Downloads, ShouldHaveLength, 1)
		So(page.Metadata.Title, ShouldEqual, datasetModel.Title)
		So(page.Metadata.Description, ShouldEqual, datasetModel.Description)
		So(page.DatasetLandingPage.Description, ShouldResemble, strings.Split(datasetModel.Description, "\n"))
		So(page.ContactDetails.Name, ShouldEqual, contact.Name)
		So(page.ContactDetails.Email, ShouldEqual, contact.Email)
		So(page.ContactDetails.Telephone, ShouldEqual, contact.Telephone)
		So(page.HasContactDetails, ShouldBeTrue)
		So(page.DatasetLandingPage.Methodologies[0].Description, ShouldEqual, methodology.Description)
		So(page.DatasetLandingPage.Methodologies[0].Title, ShouldEqual, methodology.Title)
		So(page.DatasetLandingPage.Methodologies[0].URL, ShouldEqual, methodology.URL)
		So(page.DatasetLandingPage.LatestVersionURL, ShouldBeBlank)
		So(page.Collapsible.CollapsibleItems[0].Subheading, ShouldEqual, versionOneDetails.Dimensions[0].Name)
		So(page.Collapsible.CollapsibleItems[0].Content[0], ShouldEqual, versionOneDetails.Dimensions[0].Description)
		So(page.Collapsible.CollapsibleItems[1].Subheading, ShouldEqual, versionOneDetails.Dimensions[1].Name)
		So(page.Collapsible.CollapsibleItems[1].Content, ShouldResemble, strings.Split(versionOneDetails.Dimensions[1].Description, "\n"))
		So(page.Collapsible.CollapsibleItems, ShouldHaveLength, 2)
		So(page.DatasetLandingPage.IsFlexible, ShouldBeTrue)
		So(page.DatasetLandingPage.Dimensions[0].Title, ShouldEqual, filterOutput.Dimensions[0].Label)
		So(page.DatasetLandingPage.Dimensions[0].Values, ShouldResemble, filterOutput.Dimensions[0].Options)
		So(page.DatasetLandingPage.Dimensions[1].IsCoverage, ShouldBeTrue)
		So(page.DatasetLandingPage.Dimensions[1].Values, ShouldResemble, filterOutput.Dimensions[0].Options)
	})

	Convey("Release date and hasOtherVersions is mapped correctly when v2 of Census DLP dataset is loaded", t, func() {
		req := httptest.NewRequest("", "/datasets/cantabular-1/editions/2021/versions/2", nil)
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionTwoDetails, datasetOptions, versionOneDetails.ReleaseDate, true, []dataset.Version{versionOneDetails, versionTwoDetails}, 2, "/a/version/123", "", []string{}, 50, false, false, false, filter.Model{})
		So(page.InitialReleaseDate, ShouldEqual, versionOneDetails.ReleaseDate)
		So(page.Version.ReleaseDate, ShouldEqual, versionTwoDetails.ReleaseDate)
		So(page.DatasetLandingPage.HasOtherVersions, ShouldBeTrue)
		So(page.Versions[0].VersionURL, ShouldEqual, "/datasets/cantabular-1/editions/2021/versions/2")
		So(page.Versions[0].VersionNumber, ShouldEqual, 2)
		So(page.Versions[0].ReleaseDate, ShouldEqual, versionTwoDetails.ReleaseDate)
		So(page.Versions[0].IsCurrentPage, ShouldBeTrue)
		So(page.Versions[0].Corrections[0].Reason, ShouldEqual, "This is a correction")
		So(page.DatasetLandingPage.LatestVersionURL, ShouldEqual, "/a/version/123")
	})

	Convey("IsCurrent returns false when request is for a different page", t, func() {
		req := httptest.NewRequest("", "/datasets/cantabular-1/editions/2021/versions/1", nil)
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionTwoDetails, datasetOptions, versionOneDetails.ReleaseDate, true, []dataset.Version{versionOneDetails, versionTwoDetails}, 2, "", "", []string{}, 50, false, false, false, filter.Model{})
		So(page.Versions[0].VersionURL, ShouldEqual, "/datasets/cantabular-1/editions/2021/versions/2")
		So(page.Versions[0].IsCurrentPage, ShouldBeFalse)
	})

	Convey("Versions history is in descending order", t, func() {
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionTwoDetails, datasetOptions, versionOneDetails.ReleaseDate, true, []dataset.Version{versionOneDetails, versionTwoDetails, versionThreeDetails}, 3, "", "", []string{}, 50, false, false, false, filter.Model{})
		So(page.Versions[0].VersionNumber, ShouldEqual, 3)
		So(page.Versions[1].VersionNumber, ShouldEqual, 2)
		So(page.Versions[2].VersionNumber, ShouldEqual, 1)
	})

	Convey("ShowOtherVersionsPanel is set correctly", t, func() {
		// Landed on version 1, more than one version available = panel displays
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionOneDetails, datasetOptions, versionOneDetails.ReleaseDate, true, []dataset.Version{versionOneDetails, versionTwoDetails, versionThreeDetails}, 3, "", "", []string{}, 50, false, false, false, filter.Model{})
		So(page.DatasetLandingPage.ShowOtherVersionsPanel, ShouldBeTrue)

		// Only one version = panel hidden
		page = CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionOneDetails, datasetOptions, versionOneDetails.ReleaseDate, false, []dataset.Version{versionOneDetails}, 1, "", "", []string{}, 50, false, false, false, filter.Model{})
		So(page.DatasetLandingPage.ShowOtherVersionsPanel, ShouldBeFalse)

		// More than one version, landed on latest version (3) = panel hidden
		page = CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionThreeDetails, datasetOptions, versionThreeDetails.ReleaseDate, true, []dataset.Version{versionOneDetails, versionTwoDetails, versionThreeDetails}, 3, "", "", []string{}, 50, false, false, false, filter.Model{})
		So(page.DatasetLandingPage.ShowOtherVersionsPanel, ShouldBeFalse)
	})

	Convey("Validation error passed as true, error title should be populated", t, func() {
		req := httptest.NewRequest("", "/?f=get-data", nil)
		versionDetails := dataset.Version{
			Downloads: map[string]dataset.Download{
				"XLSX": {
					Size: "1234",
					URL:  "https://mydomain.com/my-request.xlsx",
				},
			},
		}
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionDetails, datasetOptions, versionOneDetails.ReleaseDate, false, []dataset.Version{}, 1, "", "", []string{}, 50, true, false, false, filter.Model{})
		So(page.Error.Title, ShouldEqual, fmt.Sprintf("Error: %s", datasetModel.Title))
	})

	Convey("Validation error passed as false, error title should be empty", t, func() {
		req := httptest.NewRequest("", "/?f=get-data&format=xlsx", nil)
		versionDetails := dataset.Version{
			Downloads: map[string]dataset.Download{
				"XLSX": {
					Size: "1234",
					URL:  "https://mydomain.com/my-request.xlsx",
				},
			},
		}
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionDetails, datasetOptions, versionOneDetails.ReleaseDate, false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, filter.Model{})
		So(page.Error.Title, ShouldBeBlank)
	})

	Convey("Unknown get query request made, format selection error title should be empty", t, func() {
		req := httptest.NewRequest("", "/?f=blah-blah", nil)
		versionDetails := dataset.Version{
			Downloads: map[string]dataset.Download{
				"XLSX": {
					Size: "1234",
					URL:  "https://mydomain.com/my-request.xlsx",
				},
			},
		}
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionDetails, datasetOptions, versionOneDetails.ReleaseDate, false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, filter.Model{})
		So(page.Error.Title, ShouldBeBlank)
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
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, noContactDM, versionOneDetails, datasetOptions, "", false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, filter.Model{})
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
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, oneContactDetailDM, versionOneDetails, datasetOptions, "", false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, filter.Model{})
		So(page.ContactDetails.Email, ShouldEqual, oneContactDetail.Email)
		So(page.ContactDetails.Telephone, ShouldEqual, oneContactDetail.Telephone)
		So(page.HasContactDetails, ShouldBeTrue)
	})

	Convey("Dataset type is flexible, additional mapping is correct", t, func() {
		flexDm := dataset.DatasetDetails{
			Type: "cantabular_flexible_table",
			ID:   "test-flex",
		}
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, flexDm, versionOneDetails, datasetOptions, "", false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, filter.Model{})
		So(page.DatasetLandingPage.IsFlexible, ShouldBeTrue)
		So(page.DatasetLandingPage.FormAction, ShouldEqual, fmt.Sprintf("/datasets/%s/editions/%s/versions/%s/filter-flex", flexDm.ID, versionOneDetails.Edition, strconv.Itoa(versionOneDetails.Version)))
	})

	Convey("Test HasDownloads", t, func() {
		Convey("On version", func() {
			Convey("HasDownloads set to true when downloads are greater than zero", func() {
				page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, oneContactDetailDM, dataset.Version{Downloads: map[string]dataset.Download{
					"XLSX": {
						Size: "1234",
						URL:  "https://mydomain.com/my-request.xlsx",
					},
				}}, datasetOptions, "", false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, filter.Model{})
				So(page.DatasetLandingPage.HasDownloads, ShouldBeTrue)
			})
			Convey("HasDownloads set to false when downloads are zero", func() {
				page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, oneContactDetailDM, dataset.Version{Downloads: nil}, datasetOptions, "", false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, filter.Model{})
				So(page.DatasetLandingPage.HasDownloads, ShouldBeFalse)
			})
		})
		Convey("On filterOutput", func() {
			Convey("HasDownloads set to true when downloads are greater than zero", func() {
				page := CreateCensusDatasetLandingPage(context.Background(),
					req,
					pageModel,
					oneContactDetailDM,
					dataset.Version{},
					datasetOptions,
					"",
					false,
					[]dataset.Version{},
					1,
					"",
					"",
					[]string{},
					50,
					false,
					true,
					false,
					filter.Model{
						Dimensions: []filter.ModelDimension{
							{
								Name:       "Area 1",
								IsAreaType: helpers.ToBoolPtr(true),
								Options:    []string{"one", "two", "three"},
							},
						},
						Downloads: map[string]filter.Download{
							"XLSX": {
								Size: "1234",
								URL:  "https://mydomain.com/my-request.xlsx",
							},
						}})
				So(page.DatasetLandingPage.HasDownloads, ShouldBeTrue)
			})
			Convey("HasDownloads set to false when downloads are zero", func() {
				page := CreateCensusDatasetLandingPage(context.Background(),
					req,
					pageModel,
					oneContactDetailDM,
					dataset.Version{},
					datasetOptions,
					"",
					false,
					[]dataset.Version{},
					1,
					"",
					"",
					[]string{},
					50,
					false,
					true,
					false,
					filter.Model{
						Dimensions: []filter.ModelDimension{
							{
								Name:       "Area 1",
								IsAreaType: helpers.ToBoolPtr(true),
								Options:    []string{"one", "two", "three"},
							},
						},
						Downloads: nil,
					})
				So(page.DatasetLandingPage.HasDownloads, ShouldBeFalse)
			})
		})
	})

	Convey("given a dimension to truncate on census dataset landing page", t, func() {
		datasetOptions := []dataset.Options{
			{
				Items: []dataset.Option{
					{
						DimensionID: "Dimension 1",
						Label:       "Label 1",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 2",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 3",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 4",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 5",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 6",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 7",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 8",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 9",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 10",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 11",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 12",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 13",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 14",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 15",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 16",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 17",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 18",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 19",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 20",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 21",
					},
				},
				TotalCount: 20,
			},
			{
				Items: []dataset.Option{
					{
						DimensionID: "Dimension 2",
						Label:       "Label 1",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 2",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 3",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 4",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 5",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 6",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 7",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 8",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 9",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 10",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 11",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 12",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 13",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 14",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 15",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 16",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 17",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 18",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 19",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 20",
					},
				},
				TotalCount: 20,
			},
		}

		Convey("when valid parameters are provided", func() {
			p := CreateCensusDatasetLandingPage(
				context.Background(),
				req,
				pageModel,
				oneContactDetailDM,
				versionOneDetails,
				datasetOptions,
				"",
				false,
				[]dataset.Version{},
				1,
				"",
				"",
				[]string{},
				50,
				false,
				false,
				false,
				filter.Model{Downloads: nil})
			Convey("then the list should be truncated to show the first, middle, and last three values", func() {
				So(p.DatasetLandingPage.Dimensions[0].TotalItems, ShouldEqual, datasetOptions[0].TotalCount)
				So(p.DatasetLandingPage.Dimensions[0].Values, ShouldHaveLength, 9)
				So(p.DatasetLandingPage.Dimensions[0].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3",
					"Label 9", "Label 10", "Label 11",
					"Label 19", "Label 20", "Label 21",
				})
				So(p.DatasetLandingPage.Dimensions[0].IsTruncated, ShouldBeTrue)
				So(p.DatasetLandingPage.Dimensions[0].TruncateLink, ShouldEqual, "/?showAll=dim_1#dim_1")
			})
		})

		Convey("when 'showAll' parameter provided", func() {
			p := CreateCensusDatasetLandingPage(
				context.Background(),
				req,
				pageModel,
				oneContactDetailDM,
				versionOneDetails,
				datasetOptions,
				"",
				false,
				[]dataset.Version{},
				1,
				"",
				"",
				[]string{"dim_1"},
				50,
				false,
				false,
				false,
				filter.Model{Downloads: nil})

			Convey("then the dimension is no longer truncated", func() {
				So(p.DatasetLandingPage.Dimensions[0].TotalItems, ShouldEqual, datasetOptions[0].TotalCount)
				So(p.DatasetLandingPage.Dimensions[0].Values, ShouldHaveLength, 21)
				So(p.DatasetLandingPage.Dimensions[0].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3", "Label 4", "Label 5",
					"Label 6", "Label 7", "Label 8", "Label 9", "Label 10",
					"Label 11", "Label 12", "Label 13", "Label 14", "Label 15",
					"Label 16", "Label 17", "Label 18", "Label 19", "Label 20",
					"Label 21",
				})
				So(p.DatasetLandingPage.Dimensions[0].IsTruncated, ShouldBeFalse)
				So(p.DatasetLandingPage.Dimensions[0].TruncateLink, ShouldEqual, "/#dim_1")
			})

			Convey("then other truncated dimensions are persisted", func() {
				So(p.DatasetLandingPage.Dimensions[0].Values, ShouldHaveLength, 21)
				So(p.DatasetLandingPage.Dimensions[0].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3", "Label 4", "Label 5",
					"Label 6", "Label 7", "Label 8", "Label 9", "Label 10",
					"Label 11", "Label 12", "Label 13", "Label 14", "Label 15",
					"Label 16", "Label 17", "Label 18", "Label 19", "Label 20",
					"Label 21",
				})
				So(p.DatasetLandingPage.Dimensions[0].TruncateLink, ShouldEqual, "/#dim_1")
				So(p.DatasetLandingPage.Dimensions[2].TotalItems, ShouldEqual, datasetOptions[1].TotalCount)
				So(p.DatasetLandingPage.Dimensions[2].Values, ShouldHaveLength, 9)
				So(p.DatasetLandingPage.Dimensions[2].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3",
					"Label 9", "Label 10", "Label 11",
					"Label 18", "Label 19", "Label 20",
				})
				So(p.DatasetLandingPage.Dimensions[2].IsTruncated, ShouldBeTrue)
				So(p.DatasetLandingPage.Dimensions[2].TruncateLink, ShouldEqual, "/?showAll=dim_2#dim_2")
			})
		})
	})

	Convey("given a dimension to truncate on filter output landing page", t, func() {
		filterDims := filter.Model{
			Dimensions: []filter.ModelDimension{
				{
					Label: "Dimension 1",
					ID:    "dim_1",
					Options: []string{
						"Label 1", "Label 2", "Label 3", "Label 4", "Label 5",
						"Label 6", "Label 7", "Label 8", "Label 9", "Label 10",
						"Label 11", "Label 12", "Label 13", "Label 14", "Label 15",
						"Label 16", "Label 17", "Label 18", "Label 19", "Label 20",
						"Label 21",
					},
				},
				{
					Label: "Dimension 2",
					ID:    "dim_2",
					Options: []string{
						"Label 1", "Label 2", "Label 3", "Label 4",
						"Label 5", "Label 6", "Label 7", "Label 8",
						"Label 9", "Label 10", "Label 11", "Label 12",
					},
				},
			},
		}

		Convey("when valid parameters are provided", func() {
			p := CreateCensusDatasetLandingPage(
				context.Background(),
				req,
				pageModel,
				oneContactDetailDM,
				versionOneDetails,
				datasetOptions,
				"",
				false,
				[]dataset.Version{},
				1,
				"",
				"",
				[]string{},
				50,
				false,
				true,
				false,
				filterDims)
			Convey("then the list should be truncated to show the first, middle, and last three values", func() {
				So(p.DatasetLandingPage.Dimensions[0].TotalItems, ShouldEqual, 21)
				So(p.DatasetLandingPage.Dimensions[0].Values, ShouldHaveLength, 9)
				So(p.DatasetLandingPage.Dimensions[0].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3",
					"Label 9", "Label 10", "Label 11",
					"Label 19", "Label 20", "Label 21",
				})
				So(p.DatasetLandingPage.Dimensions[0].IsTruncated, ShouldBeTrue)
				So(p.DatasetLandingPage.Dimensions[0].TruncateLink, ShouldEqual, "/?showAll=dim_1#dim_1")
			})
		})

		Convey("when 'showAll' parameter provided", func() {
			p := CreateCensusDatasetLandingPage(
				context.Background(),
				req,
				pageModel,
				oneContactDetailDM,
				versionOneDetails,
				datasetOptions,
				"",
				false,
				[]dataset.Version{},
				1,
				"",
				"",
				[]string{"dim_1"},
				50,
				false,
				true,
				false,
				filterDims)

			Convey("then the dimension is no longer truncated", func() {
				So(p.DatasetLandingPage.Dimensions[0].TotalItems, ShouldEqual, 21)
				So(p.DatasetLandingPage.Dimensions[0].Values, ShouldHaveLength, 21)
				So(p.DatasetLandingPage.Dimensions[0].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3", "Label 4", "Label 5",
					"Label 6", "Label 7", "Label 8", "Label 9", "Label 10",
					"Label 11", "Label 12", "Label 13", "Label 14", "Label 15",
					"Label 16", "Label 17", "Label 18", "Label 19", "Label 20",
					"Label 21",
				})
				So(p.DatasetLandingPage.Dimensions[0].IsTruncated, ShouldBeFalse)
				So(p.DatasetLandingPage.Dimensions[0].TruncateLink, ShouldEqual, "/#dim_1")
			})

			Convey("then other truncated dimensions are persisted", func() {
				So(p.DatasetLandingPage.Dimensions[0].Values, ShouldHaveLength, 21)
				So(p.DatasetLandingPage.Dimensions[0].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3", "Label 4", "Label 5",
					"Label 6", "Label 7", "Label 8", "Label 9", "Label 10",
					"Label 11", "Label 12", "Label 13", "Label 14", "Label 15",
					"Label 16", "Label 17", "Label 18", "Label 19", "Label 20",
					"Label 21",
				})
				So(p.DatasetLandingPage.Dimensions[0].TruncateLink, ShouldEqual, "/#dim_1")
				So(p.DatasetLandingPage.Dimensions[2].TotalItems, ShouldEqual, 12)
				So(p.DatasetLandingPage.Dimensions[2].Values, ShouldHaveLength, 9)
				So(p.DatasetLandingPage.Dimensions[2].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3",
					"Label 5", "Label 6", "Label 7",
					"Label 10", "Label 11", "Label 12",
				})
				So(p.DatasetLandingPage.Dimensions[2].IsTruncated, ShouldBeTrue)
				So(p.DatasetLandingPage.Dimensions[2].TruncateLink, ShouldEqual, "/?showAll=dim_2#dim_2")
			})
		})
	})
}

func getTestEmergencyBanner() zebedee.EmergencyBanner {
	return zebedee.EmergencyBanner{
		Type:        "notable_death",
		Title:       "This is not not an emergency",
		Description: "Something has gone wrong",
		URI:         "google.com",
		LinkText:    "More info",
	}
}

func getTestServiceMessage() string {
	return "Test service message"
}
