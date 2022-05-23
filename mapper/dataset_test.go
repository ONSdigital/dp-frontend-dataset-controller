package mapper

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-renderer/model"
	. "github.com/smartystreets/goconvey/convey"
	"net/http/httptest"
	"testing"
)

// TestCreateDatasetPage tests the CreateDatasetPage method in the mapper
func TestCreateDatasetPage(t *testing.T) {
	req := httptest.NewRequest("", "/", nil)
	expectedType := "dataset"
	dummyModelData := zebedee.Dataset{
		Type: "dataset",
		URI:  "/economy/inflationandpriceindices/datasets/consumerpriceinflation/current",
		Downloads: []zebedee.Download{
			{
				File: "consumerpriceinflationdetailedreferencetables18052021121126.xls",
				Size: "400",
			},
		},
		SupplementaryFiles: []zebedee.SupplementaryFile{
			{
				Title: "Re-referencing of the CPI and CPIH indices to 2015\u003d100",
				File:  "cpirereferenceddata2015100tcm77432809.xls",
				Size:  "900",
			},
		},
		Versions: []zebedee.Version{
			{
				URI:         "/economy/inflationandpriceindices/datasets/consumerpriceinflation/current/previous/v1",
				ReleaseDate: "2015-11-17T09:30:00.000Z",
				Notice:      "A small error occurred in the division-level CPI contributions to the monthly change (between December 2018 and January 2019) published in column C of Table 26 of the Consumer price inflation tables. With the exception of the contributions presented in Table 26, the CPI headline rate or all other series are unaffected. The error occurred in updating the CPI divisional weights within the calculation of the contributions to the monthly change in the CPI (within the January 2019 publication). We have corrected this error. You can see all previous versions of this data on the previous versions page. We apologise for any inconvenience.\n",
				Label:       "",
			},
			{
				URI:         "/economy/inflationandpriceindices/datasets/consumerpriceinflation/current/previous/v2",
				ReleaseDate: "2015-12-15T09:30:00.000Z",
				Notice:      "",
				Label:       "",
			},
		},
		Description: zebedee.Description{
			Title:   "Consumer Price Inflation",
			Summary: "Measures of inflation data including CPI, CPIH, RPI and RPIJ. These tables complement the Consumer Price Inflation time series data sets available on our website.",
			Keywords: []string{
				"Economy",
				"Weights",
				"Index",
				"Indices",
				"Retail",
			},
			MetaDescription:   "Measures of inflation data including CPI, CPIH, RPI and RPIJ. These tables complement the Consumer Price Inflation time series data sets available on our website.",
			NationalStatistic: true,
			Contact: zebedee.Contact{
				Email:     "tester001@ons.gov.uk ",
				Name:      "Test Tester",
				Telephone: "+44 (0)1633 456900 ",
			},
			ReleaseDate:  "2015-12-16T09:38:39.038Z",
			NextRelease:  "17 November 2015",
			Edition:      "Current",
			DatasetID:    "",
			Unit:         "",
			PreUnit:      "",
			Source:       "",
			VersionLabel: "",
		},
	}

	dummyLandingPage := zebedee.DatasetLandingPage{
		Section: zebedee.Section{
			Markdown: "The interactive Personal Inflation Calculator, which could be used by people to calculate their personal inflation based on their spending patterns, has currently been removed from the website. The facility was used by a very small number of people.",
		},
		RelatedFilterableDatasets: []zebedee.Related{
			{
				URI: "/datasets/cpih01",
			},
		},

		RelatedDatasets: []zebedee.Related{},
		RelatedDocuments: []zebedee.Related{
			{
				URI: "/economy/inflationandpriceindices/bulletins/consumerpriceinflation/latest",
			},
		},
		Datasets: []zebedee.Related{
			{
				URI: "/economy/inflationandpriceindices/datasets/consumerpriceinflation/current",
			},
		},
		RelatedLinks:              []zebedee.Related{},
		Alerts:                    []zebedee.Alert{},
		RelatedMethodology:        []zebedee.Related{},
		RelatedMethodologyArticle: []zebedee.Related{},
		Type:                      "dataset_landing_page",
		URI:                       "/economy/inflationandpriceindices/datasets/consumerpriceinflation",
		Description: zebedee.Description{
			Title:   "Consumer Price Inflation",
			Summary: "Measures of inflation data including CPI, CPIH, RPI and RPIJ. These tables complement the Consumer Price Inflation time series data sets available on our website.",
			Keywords: []string{
				"Economy",
				"Weights",
				"Index",
				"Indices",
				"Retail",
			},
			MetaDescription:   "Measures of inflation data including CPI, CPIH, RPI and RPIJ. These tables complement the Consumer Price Inflation time series data sets available on our website.",
			NationalStatistic: true,
			Contact: zebedee.Contact{
				Email:     "tester001@ons.gov.uk ",
				Name:      "Test Tester",
				Telephone: "+44 (0)1633 123456 ",
			},
			ReleaseDate: "2015-12-16T09:38:39.038Z",
			NextRelease: "17 November 2015",
			Edition:     "Current",
			DatasetID:   "",
			Unit:        "",
			PreUnit:     "",
			Source:      "",
		},
	}

	dummyVersions := []zebedee.Dataset{
		{
			Downloads: []zebedee.Download{
				{
					Size: "500",
					File: "consumerpriceinflationdetailedreferencetables_tcm77-419243.xls",
				},
			},
			Type: "dataset",
			URI:  "/economy/inflationandpriceindices/datasets/consumerpriceinflation/current/previous/v1",
			Description: zebedee.Description{
				Title:   "Consumer Price Inflation",
				Summary: "Measures of inflation data including CPI, CPIH, RPI and RPIJ. These tables complement the Consumer Price Inflation time series data sets available on our website.",
				Keywords: []string{
					"Economy",
					"Weights",
					"Index",
					"Indices",
					"Retail",
				},
				MetaDescription:   "Measures of inflation data including CPI, CPIH, RPI and RPIJ. These tables complement the Consumer Price Inflation time series data sets available on our website.",
				NationalStatistic: true,
				Contact: zebedee.Contact{
					Email:     "tester001@ons.gov.uk ",
					Name:      "Test Tester",
					Telephone: "+44 (0)1633 123456 ",
				},
				ReleaseDate: "2015-12-16T09:38:39.038Z",
				NextRelease: "17 November 2015",
				Edition:     "Current",
				DatasetID:   "",
				Unit:        "",
				PreUnit:     "",
				Source:      "",
			},
		},
		{
			Downloads: []zebedee.Download{
				{
					Size: "600",
					File: "consumerpriceinflationdetailedreferencetables_tcm77-423330.xls",
				},
			},
			Type:               "dataset",
			URI:                "/economy/inflationandpriceindices/datasets/consumerpriceinflation/current/previous/v2",
			SupplementaryFiles: []zebedee.SupplementaryFile{},
			Versions: []zebedee.Version{
				{
					URI:         "/economy/inflationandpriceindices/datasets/consumerpriceinflation/current/previous/v1",
					ReleaseDate: "2015-11-17T00:00:00.000Z",
					Notice:      "",
					Label:       "November 2015",
				},
			},
			Description: zebedee.Description{
				Title:   "Consumer Price Inflation",
				Summary: "Measures of inflation data including CPI, CPIH, RPI and RPIJ. These tables complement the Consumer Price Inflation time series data sets available on our website.",
				Keywords: []string{
					"Economy",
					"Weights",
					"Index",
					"Indices",
					"Retail",
				},
				MetaDescription:   "Measures of inflation data including CPI, CPIH, RPI and RPIJ. These tables complement the Consumer Price Inflation time series data sets available on our website.",
				NationalStatistic: true,
				Contact: zebedee.Contact{
					Email:     "tester001@ons.gov.uk ",
					Name:      "Test Tester",
					Telephone: "+44 (0)1633 123456 ",
				},
				ReleaseDate:  "2015-12-16T09:38:39.038Z",
				NextRelease:  "17 November 2015",
				Edition:      "Current",
				DatasetID:    "",
				Unit:         "",
				PreUnit:      "",
				Source:       "",
				VersionLabel: "November 2015",
			},
		},
	}

	// breadcrumbItem returned by zebedee after being proxied through API router
	breadcrumbItem0 := zebedee.Breadcrumb{
		URI:         "http://myHost:1234/v1/economy/grossdomesticproduct/datasets/gdpjanuary2018",
		Description: zebedee.NodeDescription{Title: "GDP: January 2018"},
	}

	// breadcrumbItem as expected as a result of CreateFilterableLandingPage
	expectedBreadcrumbItem0 := zebedee.Breadcrumb{
		URI:         "http://myHost:1234/economy/grossdomesticproduct/datasets/gdpjanuary2018",
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

	ctx := context.Background()
	mdl := model.Page{}

	Convey("test dataset page correctly returns", t, func() {
		dp := CreateDatasetPage(mdl, ctx, req, dummyModelData, dummyLandingPage,
			[]zebedee.Breadcrumb{breadcrumbItem0, breadcrumbItem1, breadcrumbItemWrongURI},
			dummyVersions, "en", "/v1")

		So(dp.Breadcrumb[0].Title, ShouldEqual, expectedBreadcrumbItem0.Description.Title)
		So(dp.Breadcrumb[0].URI, ShouldEqual, expectedBreadcrumbItem0.URI)
		So(dp.Breadcrumb[1].Title, ShouldEqual, expectedBreadcrumbItem1.Description.Title)
		So(dp.Breadcrumb[1].URI, ShouldEqual, expectedBreadcrumbItem1.URI)
		So(dp.Breadcrumb[2].Title, ShouldEqual, expectedBreadcrumbItemWrongURI.Description.Title)
		So(dp.Breadcrumb[2].URI, ShouldEqual, expectedBreadcrumbItemWrongURI.URI)
		So(dp.Breadcrumb[3].Title, ShouldEqual, dummyModelData.Description.Edition)

		So(dp.Type, ShouldEqual, expectedType)
		So(dp.Metadata.Title, ShouldEqual, dummyLandingPage.Description.Title)
		So(dp.URI, ShouldEqual, dummyModelData.URI)
		So(dp.DatasetPage.URI, ShouldEqual, dummyLandingPage.URI)
		So(dp.Metadata.Description, ShouldEqual, dummyLandingPage.Description.Summary)
		So(dp.DatasetPage.ReleaseDate, ShouldEqual, dummyLandingPage.Description.ReleaseDate)
		So(dp.DatasetPage.Edition, ShouldEqual, dummyModelData.Description.Edition)
		So(dp.DatasetPage.Markdown, ShouldEqual, dummyLandingPage.Section.Markdown)
		So(dp.DatasetPage.IsNationalStatistic, ShouldEqual, dummyLandingPage.Description.NationalStatistic)
		So(dp.DatasetPage.NextRelease, ShouldEqual, dummyLandingPage.Description.NextRelease)
		So(dp.DatasetPage.DatasetID, ShouldEqual, dummyLandingPage.Description.DatasetID)
		So(dp.ContactDetails.Email, ShouldEqual, "tester001@ons.gov.uk")
		So(dp.ContactDetails.Telephone, ShouldEqual, dummyLandingPage.Description.Contact.Telephone)
		So(dp.ContactDetails.Name, ShouldEqual, dummyLandingPage.Description.Contact.Name)

		So(dp.DatasetPage.Downloads[0].Size, ShouldEqual, "400")
		So(dp.DatasetPage.Downloads[0].Extension, ShouldEqual, ".xls")
		So(dp.DatasetPage.Downloads[0].URI, ShouldEqual, "/economy/inflationandpriceindices/datasets/consumerpriceinflation/current/consumerpriceinflationdetailedreferencetables18052021121126.xls")

		So(dp.DatasetPage.SupplementaryFiles[0].Title, ShouldEqual, "Re-referencing of the CPI and CPIH indices to 2015\u003d100")
		So(dp.DatasetPage.SupplementaryFiles[0].Size, ShouldEqual, "900")
		So(dp.DatasetPage.SupplementaryFiles[0].Extension, ShouldEqual, ".xls")
		So(dp.DatasetPage.SupplementaryFiles[0].URI, ShouldEqual, "/economy/inflationandpriceindices/datasets/consumerpriceinflation/current/cpirereferenceddata2015100tcm77432809.xls")

		v0 := dp.DatasetPage.Versions[0]
		So(v0.URI, ShouldEqual, dummyModelData.Versions[1].URI)
		So(v0.UpdateDate, ShouldEqual, dummyModelData.Versions[1].ReleaseDate)
		So(v0.CorrectionNotice, ShouldEqual, dummyModelData.Versions[1].Notice)
		So(v0.Label, ShouldEqual, dummyModelData.Versions[1].Label)
		So(v0.Downloads[0].Size, ShouldEqual, "600")
		So(v0.Downloads[0].Extension, ShouldEqual, ".xls")
		So(v0.Downloads[0].URI, ShouldEqual, "/economy/inflationandpriceindices/datasets/consumerpriceinflation/current/previous/v2/consumerpriceinflationdetailedreferencetables_tcm77-423330.xls")

		v1 := dp.DatasetPage.Versions[1]
		So(v1.URI, ShouldEqual, dummyModelData.Versions[0].URI)
		So(v1.UpdateDate, ShouldEqual, dummyModelData.Versions[0].ReleaseDate)
		So(v1.CorrectionNotice, ShouldEqual, dummyModelData.Versions[0].Notice)
		So(v1.Label, ShouldEqual, dummyModelData.Versions[0].Label)
		So(v1.Downloads[0].Size, ShouldEqual, "500")
		So(v1.Downloads[0].Extension, ShouldEqual, ".xls")
		So(v1.Downloads[0].URI, ShouldEqual, "/economy/inflationandpriceindices/datasets/consumerpriceinflation/current/previous/v1/consumerpriceinflationdetailedreferencetables_tcm77-419243.xls")
	})
}
