package mapper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	dpRendererModel "github.com/ONSdigital/dp-renderer/v2/model"
	dpTopicApiModels "github.com/ONSdigital/dp-topic-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitMapper(t *testing.T) {
	ctx := context.Background()
	mdl := dpRendererModel.Page{}

	contact := dpDatasetApiModels.ContactDetails{
		Name:      "Matt Rout",
		Telephone: "01111 222222",
		Email:     "mattrout@test.com",
	}
	d := dpDatasetApiModels.Dataset{
		CollectionID: "abcdefg",
		Contacts: []dpDatasetApiModels.ContactDetails{
			contact,
		},
		Description: "A really awesome dataset for you to look at",
		Links: &dpDatasetApiModels.DatasetLinks{
			Self: &dpDatasetApiModels.LinkObject{
				HRef: "/datasets/83jd98fkflg",
			},
		},
		NextRelease:      "11-11-2018",
		ReleaseFrequency: "Yearly",
		Publisher: &dpDatasetApiModels.Publisher{
			HRef: "ons.gov.uk",
			Name: "ONS",
			Type: "Government Agency",
		},
		State:   "created",
		Theme:   "purple",
		Title:   "Penguins of the Antarctic Ocean",
		License: "ons",
	}
	nomisD := dpDatasetApiModels.Dataset{
		CollectionID: "abcdefg",
		Contacts: []dpDatasetApiModels.ContactDetails{
			contact,
		},
		Description: "A really awesome dataset for you to look at",
		Links: &dpDatasetApiModels.DatasetLinks{
			Self: &dpDatasetApiModels.LinkObject{
				HRef: "/datasets/83jd98fkflg",
			},
		},
		NextRelease:      "11-11-2018",
		ReleaseFrequency: "Yearly",
		Publisher: &dpDatasetApiModels.Publisher{
			HRef: "ons.gov.uk",
			Name: "ONS",
			Type: "Government Agency",
		},
		State:   "created",
		Theme:   "purple",
		Title:   "Penguins of the Antarctic Ocean",
		License: "ons",
		Type:    "nomis",
	}

	v := []dpDatasetApiModels.Version{
		{
			CollectionID: "abcdefg",
			Edition:      "2017",
			ID:           "tehnskofjios-ashbc7",
			Version:      1,
			Links: &dpDatasetApiModels.VersionLinks{
				Self: &dpDatasetApiModels.LinkObject{
					HRef: "/datasets/83jd98fkflg/editions/124/versions/1",
				},
			},
			Dimensions: []dpDatasetApiModels.Dimension{
				{
					ID:    "city",
					Name:  "geography",
					Label: "City",
				},
			},
			ReleaseDate: "11-11-2017",
			State:       "published",
			Downloads: &dpDatasetApiModels.DownloadList{
				XLSX: &dpDatasetApiModels.DownloadObject{
					Size: "438290",
					HRef: "my-url",
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
		options := []dpDatasetApiSdk.VersionDimensionOptionsList{
			{
				Items: []dpDatasetApiModels.PublicDimensionOption{
					{
						Name:   "age",
						Label:  "6",
						Option: "6",
					},
					{
						Name:   "age",
						Label:  "3",
						Option: "3",
					},
					{
						Name:   "age",
						Label:  "24",
						Option: "24",
					},
					{
						Name:   "age",
						Label:  "23",
						Option: "23",
					},
					{
						Name:   "age",
						Label:  "19",
						Option: "19",
					},
				},
			},
			{
				Items: []dpDatasetApiModels.PublicDimensionOption{
					{
						Name:   "time",
						Label:  "Jan-05",
						Option: "Jan-05",
					},
					{
						Name:   "time",
						Label:  "Feb-05",
						Option: "Feb-05",
					},
				},
			},
		}
		p := CreateFilterableLandingPage(ctx, mdl, d, v[0], datasetID, options, dpDatasetApiSdk.VersionDimensionsList{}, false,
			[]zebedee.Breadcrumb{breadcrumbItem0, breadcrumbItem1, breadcrumbItemWrongURI}, 1, "/datasets/83jd98fkflg/editions/124/versions/1", "/v1", 50)

		So(p.Type, ShouldEqual, "dataset_landing_page")
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
		So(v0.Downloads[0].Extension, ShouldEqual, "xlsx")
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
		options := []dpDatasetApiSdk.VersionDimensionOptionsList{
			{
				Items: []dpDatasetApiModels.PublicDimensionOption{
					{
						Name:   "age",
						Label:  "6",
						Option: "6",
					},
					{
						Name:   "age",
						Label:  "3",
						Option: "3",
					},
					{
						Name:   "age",
						Label:  "24",
						Option: "24",
					},
					{
						Name:   "age",
						Label:  "23",
						Option: "23",
					},
					{
						Name:   "age",
						Label:  "19",
						Option: "19",
					},
				},
			},
			{
				Items: []dpDatasetApiModels.PublicDimensionOption{
					{
						Name:   "time",
						Label:  "Jan-05",
						Option: "Jan-05",
					},
					{
						Name:   "time",
						Label:  "Feb-05",
						Option: "Feb-05",
					},
				},
			},
		}
		p := CreateFilterableLandingPage(ctx, mdl, nomisD, v[0], datasetID, options, dpDatasetApiSdk.VersionDimensionsList{}, false,
			[]zebedee.Breadcrumb{breadcrumbItem0, breadcrumbItem1, breadcrumbItemWrongURI}, 1, "/datasets/83jd98fkflg/editions/124/versions/1", "/v1", 50)

		So(p.Type, ShouldEqual, "dataset_landing_page")
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
		So(p.DatasetLandingPage.NomisReferenceURL, ShouldBeEmpty)

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

		dims := dpDatasetApiSdk.VersionDimensionsList{
			Items: []dpDatasetApiModels.Dimension{
				{
					ID:    dimensionID,
					Name:  dimensionName,
					Label: dimensionLabel,
				},
			},
		}
		opts := []dpDatasetApiSdk.VersionDimensionOptionsList{
			{
				Items: []dpDatasetApiModels.PublicDimensionOption{
					{
						Name:   dimensionName,
						Label:  dimensionOptionLabel,
						Option: "0",
					},
				},
			},
		}

		p := CreateFilterableLandingPage(ctx, mdl, d, v[0], datasetID, opts, dims, false, []zebedee.Breadcrumb{},
			1, "", "/v1", 50)

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
		options := []dpDatasetApiSdk.VersionDimensionOptionsList{
			{
				Items: []dpDatasetApiModels.PublicDimensionOption{
					{
						Name:   "time",
						Label:  "Jan-05",
						Option: "Jan-05",
					},
					{
						Name:   "time",
						Label:  "May-07",
						Option: "May-07",
					},
					{
						Name:   "time",
						Label:  "Jun-07",
						Option: "Jun-07",
					},
				},
			},
		}
		p := CreateFilterableLandingPage(ctx, mdl, d, v[0], datasetID, options, dpDatasetApiSdk.VersionDimensionsList{}, false, []zebedee.Breadcrumb{},
			1, "/datasets/83jd98fkflg/editions/124/versions/1", "/v1", 50)

		So(p.Type, ShouldEqual, "dataset_landing_page")
		So(p.DatasetLandingPage.Dimensions[0].Values, ShouldHaveLength, 2)
		So(p.DatasetLandingPage.Dimensions[0].Title, ShouldEqual, "Time")
		So(p.DatasetLandingPage.Dimensions[0].Values[0], ShouldEqual, "This year 2005 contains data for the month January")
		So(p.DatasetLandingPage.Dimensions[0].Values[1], ShouldEqual, "All months between May 2007 and June 2007")
	})

	Convey("test time dimensions for CreateFilterableLandingPage ", t, func() {
		options := []dpDatasetApiSdk.VersionDimensionOptionsList{
			{
				Items: []dpDatasetApiModels.PublicDimensionOption{
					{
						Name:   "time",
						Label:  "2016",
						Option: "2016",
					},
					{
						Name:   "time",
						Label:  "2018",
						Option: "2018",
					},
					{
						Name:   "time",
						Label:  "2019",
						Option: "2019",
					},
					{
						Name:   "time",
						Label:  "2020",
						Option: "2020",
					},
				},
			},
		}
		p := CreateFilterableLandingPage(ctx, mdl, d, v[0], datasetID, options, dpDatasetApiSdk.VersionDimensionsList{}, false, []zebedee.Breadcrumb{},
			1, "/datasets/83jd98fkflg/editions/124/versions/1", "/v1", 50)

		So(p.Type, ShouldEqual, "dataset_landing_page")
		So(p.DatasetLandingPage.Dimensions[0].Values, ShouldHaveLength, 2)
		So(p.DatasetLandingPage.Dimensions[0].Title, ShouldEqual, "Time")
		So(p.DatasetLandingPage.Dimensions[0].Values[0], ShouldEqual, "This year contains data for 2016")
		So(p.DatasetLandingPage.Dimensions[0].Values[1], ShouldEqual, "All years between 2018 and 2020")
	})
}

// TestCreateVersionsList Tests the CreateVersionsList function in the mapper
func TestCreateVersionsList(t *testing.T) {
	mdl := dpRendererModel.Page{}
	req := httptest.NewRequest("", "/", http.NoBody)
	dummyModelData := dpDatasetApiModels.Dataset{
		ID:    "cpih01",
		Title: "Consumer Prices Index including owner occupiers? housing costs (CPIH)",
		Links: &dpDatasetApiModels.DatasetLinks{
			Editions: &dpDatasetApiModels.LinkObject{
				HRef: "http://localhost:22000/datasets/cpih01/editions",
				ID:   ""},
			LatestVersion: &dpDatasetApiModels.LinkObject{
				HRef: "http://localhost:22000/datasets/cpih01/editions/time-series/versions/3",
				ID:   "3"},
			Self: &dpDatasetApiModels.LinkObject{
				HRef: "http://localhost:22000/datasets/cpih01",
				ID:   ""},
			Taxonomy: &dpDatasetApiModels.LinkObject{
				HRef: "/economy/environmentalaccounts/datasets/consumerpricesindexincludingowneroccupiershousingcostscpih",
				ID:   ""},
		},
	}
	dummyEditionData := dpDatasetApiModels.Edition{}
	dummyVersion1 := dpDatasetApiModels.Version{
		Alerts:        nil,
		CollectionID:  "",
		Downloads:     nil,
		Edition:       "time-series",
		Dimensions:    nil,
		ID:            "",
		LatestChanges: nil,
		Links: &dpDatasetApiModels.VersionLinks{
			Dataset: &dpDatasetApiModels.LinkObject{
				HRef: "http://localhost:22000/datasets/cpih01",
				ID:   "cpih01",
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
	dummyVersion3.Alerts = &[]dpDatasetApiModels.Alert{
		{
			Date:        "",
			Description: "This is a correction",
			Type:        "correction",
		},
	}
	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()

	Convey("test latest version page", t, func() {
		dummySingleVersionList := []dpDatasetApiModels.Version{dummyVersion3}

		page := CreateVersionsList(mdl, req, dummyModelData, dummyEditionData, dummySingleVersionList, serviceMessage, emergencyBanner)
		Convey("title", func() {
			So(page.Metadata.Title, ShouldEqual, "All versions of Consumer Prices Index including owner occupiers? housing costs (CPIH) time-series dataset")
		})
		Convey("has correct number of versions when only one should be present", func() {
			So(page.Data.Versions, ShouldHaveLength, 1)
		})

		dummyMultipleVersionList := []dpDatasetApiModels.Version{dummyVersion1, dummyVersion2, dummyVersion3}
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
	req := httptest.NewRequest("", "/", http.NoBody)
	pageModel := dpRendererModel.Page{
		CookiesPreferencesSet: false,
		CookiesPolicy: dpRendererModel.CookiesPolicy{
			Communications: false,
			Essential:      false,
			Settings:       false,
			Usage:          false,
		},
	}

	Convey("maps cookies preferences cookie data to page model correctly", t, func() {
		So(pageModel.CookiesPreferencesSet, ShouldBeFalse)
		So(pageModel.CookiesPolicy.Communications, ShouldBeFalse)
		So(pageModel.CookiesPolicy.Essential, ShouldBeFalse)
		So(pageModel.CookiesPolicy.Settings, ShouldBeFalse)
		So(pageModel.CookiesPolicy.Usage, ShouldBeFalse)
		req.AddCookie(&http.Cookie{Name: "ons_cookie_message_displayed", Value: "true"})
		req.AddCookie(&http.Cookie{Name: "ons_cookie_policy", Value: "{'essential':true,'settings':true,'usage':true,'campaigns':true}"})
		MapCookiePreferences(req, &pageModel.CookiesPreferencesSet, &pageModel.CookiesPolicy)
		So(pageModel.CookiesPreferencesSet, ShouldBeTrue)
		So(pageModel.CookiesPolicy.Communications, ShouldBeTrue)
		So(pageModel.CookiesPolicy.Essential, ShouldBeTrue)
		So(pageModel.CookiesPolicy.Settings, ShouldBeTrue)
		So(pageModel.CookiesPolicy.Usage, ShouldBeTrue)
	})
}

func TestUpdateBasePage(t *testing.T) {
	basePageModel := dpRendererModel.Page{}
	contacts := getTestContacts()
	isValidationError := false
	lang := "en"
	mockRequest := httptest.NewRequest("", "/", http.NoBody)
	relatedContent := getTestRelatedContent()
	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()

	datasetDetails := getTestDatasetDetails(contacts, relatedContent)
	homepageContent := zebedee.HomepageContent{
		EmergencyBanner: emergencyBanner,
		ServiceMessage:  serviceMessage,
	}

	Convey("Test `UpdateBasePage` updates page attributes correctly default parameters", t, func() {
		UpdateBasePage(&basePageModel, datasetDetails, homepageContent, isValidationError, lang, mockRequest)

		// These parameters are set by default and are not dependent on conditional inputs
		So(basePageModel.BetaBannerEnabled, ShouldEqual, true)
		So(basePageModel.DatasetId, ShouldEqual, datasetDetails.ID)
		So(basePageModel.EmergencyBanner.Type, ShouldEqual, strings.Replace(emergencyBanner.Type, "_", "-", -1))
		So(basePageModel.EmergencyBanner.Title, ShouldEqual, emergencyBanner.Title)
		So(basePageModel.EmergencyBanner.Description, ShouldEqual, emergencyBanner.Description)
		So(basePageModel.EmergencyBanner.URI, ShouldEqual, emergencyBanner.URI)
		So(basePageModel.EmergencyBanner.LinkText, ShouldEqual, emergencyBanner.LinkText)
		So(basePageModel.Language, ShouldEqual, lang)
		So(basePageModel.Metadata.Description, ShouldEqual, datasetDetails.Description)
		So(basePageModel.Metadata.Title, ShouldEqual, datasetDetails.Title)
		So(basePageModel.ReleaseDate, ShouldEqual, "")
		So(basePageModel.ServiceMessage, ShouldEqual, serviceMessage)
		So(basePageModel.Type, ShouldEqual, datasetDetails.Type)
		So(basePageModel.URI, ShouldEqual, mockRequest.URL.Path)
	})

	Convey("Test `UpdateBasePage` does not update `Error` if `isValidationError` is `false", t, func() {
		isValidationError = false

		// Instantiation of `dpRendererModel.Page{}` sets `Error` to an empty struct
		expectedError := dpRendererModel.Error{
			Description: "",
			ErrorCode:   0,
			ErrorItems:  []dpRendererModel.ErrorItem(nil),
			Language:    "",
			Title:       "",
		}

		UpdateBasePage(&basePageModel, datasetDetails, homepageContent, isValidationError, lang, mockRequest)

		So(basePageModel.Error, ShouldEqual, expectedError)
	})

	Convey("Test `UpdateBasePage` updates `Error` if `isValidationError` is `true", t, func() {
		isValidationError = true

		// Error should be updated to show error details
		expectedError := dpRendererModel.Error{
			Description: "",
			ErrorCode:   0,
			ErrorItems: []dpRendererModel.ErrorItem{
				{
					Description: dpRendererModel.Localisation{
						LocaleKey: "GetDataValidationError",
						Plural:    1,
					},
					URL: "#select-format-error",
				},
			},
			Language: lang,
			Title:    datasetDetails.Title,
		}

		UpdateBasePage(&basePageModel, datasetDetails, homepageContent, isValidationError, lang, mockRequest)

		So(basePageModel.Error, ShouldEqual, expectedError)
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

func TestCreateBreadcrumbsFromTopicList(t *testing.T) {
	Convey("Given a topicObjectList with two topics", t, func() {
		topicObjectList := []dpTopicApiModels.Topic{
			{Title: "Topic One", Slug: "slug1"},
			{Title: "Topic Two", Slug: "slug2"},
		}

		Convey("When CreateBreadcrumbsFromTopicList is called to build the breadcrumbs from the topics list", func() {
			breadcrumbObject := CreateBreadcrumbsFromTopicList(topicObjectList)

			Convey("Then the breadcrumbs object should contain a TaxonomyNode for each topic ", func() {
				So(breadcrumbObject, ShouldHaveLength, 3)
				So(breadcrumbObject[0].Title, ShouldEqual, "Home")
				So(breadcrumbObject[0].URI, ShouldEqual, "/")
				So(breadcrumbObject[1].Title, ShouldEqual, "Topic One")
				So(breadcrumbObject[1].URI, ShouldEqual, "/slug1")
				So(breadcrumbObject[2].Title, ShouldEqual, "Topic Two")
				So(breadcrumbObject[2].URI, ShouldEqual, "/slug1/slug2")
			})
		})
	})
}
