package mapper

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitMapperLegacy(t *testing.T) {
	Convey("test MapZebedeeDatasetLandingPageToFrontendModel", t, func() {
		cfg := config.Config{}
		ctx := context.Background()
		dlp := getTestDatasetLandingPage()
		bcs := getTestBreadcrumbs()
		ds := getTestDatsets()
		lang := "cy"
		req := httptest.NewRequest("GET", "/", nil)

		sdlp := CreateLegacyDatasetLanding(cfg, ctx, req, dlp, bcs, ds, lang)
		So(sdlp, ShouldNotBeEmpty)

		So(sdlp.Type, ShouldEqual, "legacy_dataset")
		So(sdlp.URI, ShouldEqual, dlp.URI)
		So(sdlp.Metadata.Title, ShouldEqual, dlp.Description.Title)
		So(sdlp.Metadata.Description, ShouldEqual, dlp.Description.Summary)

		So(sdlp.DatasetLandingPage.Related.Datasets[0].Title, ShouldEqual, dlp.RelatedDatasets[0].Title)
		So(sdlp.DatasetLandingPage.Related.Datasets[0].URI, ShouldEqual, dlp.RelatedDatasets[0].URI)

		So(sdlp.DatasetLandingPage.Related.Publications[0].Title, ShouldEqual, dlp.RelatedDocuments[0].Title)
		So(sdlp.DatasetLandingPage.Related.Publications[0].URI, ShouldEqual, dlp.RelatedDocuments[0].URI)

		So(sdlp.DatasetLandingPage.Related.Methodology[0].Title, ShouldEqual, dlp.RelatedMethodology[0].Title)
		So(sdlp.DatasetLandingPage.Related.Methodology[0].URI, ShouldEqual, dlp.RelatedMethodology[0].URI)

		So(sdlp.ContactDetails.Email, ShouldEqual, dlp.Description.Contact.Email)
		So(sdlp.ContactDetails.Name, ShouldEqual, dlp.Description.Contact.Name)
		So(sdlp.ContactDetails.Telephone, ShouldEqual, dlp.Description.Contact.Telephone)

		So(sdlp.DatasetLandingPage.IsNationalStatistic, ShouldEqual, dlp.Description.NationalStatistic)
		So(sdlp.DatasetLandingPage.IsTimeseries, ShouldEqual, dlp.Timeseries)

		So(sdlp.DatasetLandingPage.ReleaseDate, ShouldNotBeEmpty)
		So(sdlp.DatasetLandingPage.NextRelease, ShouldEqual, dlp.Description.NextRelease)

		So(sdlp.Page.Breadcrumb[0].Title, ShouldEqual, bcs[0].Description.Title)

		So(sdlp.DatasetLandingPage.Datasets, ShouldHaveLength, 1)
		So(sdlp.DatasetLandingPage.Datasets[0].URI, ShouldEqual, "google.com")
		So(sdlp.DatasetLandingPage.Datasets[0].Downloads, ShouldHaveLength, 1)
		So(sdlp.DatasetLandingPage.Datasets[0].Downloads[0].URI, ShouldEqual, "helloworld.csv")
		So(sdlp.DatasetLandingPage.Datasets[0].Downloads[0].Extension, ShouldEqual, "csv")
		So(sdlp.DatasetLandingPage.Datasets[0].Downloads[0].Size, ShouldEqual, "452456")
	})
}

func getTestDatsets() []zebedee.Dataset {
	return []zebedee.Dataset{
		{
			Type: "dataset",
			URI:  "google.com",
			Description: zebedee.Description{
				Title:             "hello world",
				Edition:           "2016",
				Summary:           "a nice big old dataset",
				Keywords:          []string{"hello"},
				MetaDescription:   "this is so meta",
				NationalStatistic: false,
				Contact: zebedee.Contact{
					Email:     "testemail@123.com",
					Name:      "matt",
					Telephone: "01234567892",
				},
				ReleaseDate: "11/07/2016",
				NextRelease: "11/07/2017",
				DatasetID:   "12345",
				Unit:        "Joules",
				PreUnit:     "kg",
				Source:      "word of mouth",
			},
			Downloads: []zebedee.Download{
				{
					File: "helloworld.csv",
					Size: "452456",
				},
			},
			SupplementaryFiles: []zebedee.SupplementaryFile{
				{
					Title: "moredata.xls",
					File:  "helloworld.csv",
					Size:  "372920",
				},
			},
			Versions: []zebedee.Version{
				{
					URI:         "google.com",
					ReleaseDate: "01/01/2017",
					Notice:      "missing data",
					Label:       "missing data",
				},
			},
		},
	}
}

func getTestBreadcrumbs() []zebedee.Breadcrumb {
	return []zebedee.Breadcrumb{
		{
			URI: "google.com",
			Description: zebedee.NodeDescription{
				Title: "google",
			},
			Type: "web",
		},
	}
}

func getTestDatasetLandingPage() zebedee.DatasetLandingPage {
	return zebedee.DatasetLandingPage{
		Type: "dataset",
		URI:  "www.google.com",
		Description: zebedee.Description{
			Title:             "hello world",
			Edition:           "2016",
			Summary:           "a nice big old dataset",
			Keywords:          []string{"hello"},
			MetaDescription:   "this is so meta",
			NationalStatistic: false,
			Contact: zebedee.Contact{
				Email:     "testemail@123.com",
				Name:      "matt",
				Telephone: "01234567892",
			},
			ReleaseDate: "11/07/2016",
			NextRelease: "11/07/2017",
			DatasetID:   "12345",
			Unit:        "Joules",
			PreUnit:     "kg",
			Source:      "word of mouth",
		},
		Section: zebedee.Section{
			Markdown: "markdown",
		},
		Datasets: []zebedee.Related{
			{
				Title: "google",
				URI:   "google.com",
			},
		},
		RelatedLinks: []zebedee.Related{
			{
				Title: "google",
				URI:   "google.com",
			},
		},
		RelatedDatasets: []zebedee.Related{
			{
				Title: "google",
				URI:   "google.com",
			},
		},
		RelatedDocuments: []zebedee.Related{
			{
				Title: "google",
				URI:   "google.com",
			},
		},
		RelatedMethodology: []zebedee.Related{
			{
				Title: "google",
				URI:   "google.com",
			},
		},
		RelatedMethodologyArticle: []zebedee.Related{
			{
				Title: "google",
				URI:   "google.com",
			},
		},
		Alerts: []zebedee.Alert{
			{
				Date:     "05/05/2017",
				Type:     "alert",
				Markdown: "12345",
			},
			{
				Date:     "05/05/2017",
				Type:     "correction",
				Markdown: "12345",
			},
			{
				Date:     "05/05/2017",
				Type:     "unrecognised",
				Markdown: "12345",
			},
		},
		Timeseries: true,
	}
}
