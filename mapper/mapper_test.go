package mapper

import (
	"os"
	"strconv"
	"testing"

	"github.com/ONSdigital/go-ns/clients/dataset"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitMapper(t *testing.T) {
	Convey("test CreateFilterableLandingPage", t, func() {
		d := dataset.Model{
			CollectionID: "abcdefg",
			Contacts: []dataset.Contact{
				dataset.Contact{
					Name:      "Matt Rout",
					Telephone: "01622 734721",
					Email:     "mattrout@test.com",
				},
			},
			Description: "A really awesome dataset for you to look at",
			Links: dataset.Links{
				Self: dataset.Link{
					URL: "/datasets/83jd98fkflg",
				},
			},
			NextRelease:      "11-11-2018",
			ReleaseFrequency: "Yearly",
			Publisher: dataset.Publisher{
				URL:  "ons.gov.uk",
				Name: "ONS",
				Type: "Government Agency",
			},
			State:   "created",
			Theme:   "purple",
			Title:   "Penguins of the Antarctic Ocean",
			License: "ons",
		}

		v := []dataset.Version{
			dataset.Version{
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
					"XLSX": dataset.Download{
						Size: "438290",
						URL:  "my-url",
					},
				},
			},
		}
		datasetID := "038847784-2874757-23784854905"

		p := CreateFilterableLandingPage(d, v[0], datasetID, []dataset.Options{
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
		}, dataset.Dimensions{}, false)

		So(p.Type, ShouldEqual, "dataset_landing_page")
		So(p.Metadata.Title, ShouldEqual, d.Title)
		So(p.URI, ShouldEqual, d.Links.Self.URL)
		So(p.ShowFeedbackForm, ShouldEqual, true)
		So(p.ContactDetails.Name, ShouldEqual, d.Contacts[0].Name)
		So(p.ContactDetails.Telephone, ShouldEqual, d.Contacts[0].Telephone)
		So(p.ContactDetails.Email, ShouldEqual, d.Contacts[0].Email)
		So(p.DatasetLandingPage.NextRelease, ShouldEqual, d.NextRelease)
		So(p.DatasetLandingPage.DatasetID, ShouldEqual, datasetID)
		So(p.DatasetLandingPage.ReleaseDate, ShouldEqual, v[0].ReleaseDate)
		So(p.ShowFeedbackForm, ShouldEqual, true)

		So(len(p.DatasetLandingPage.Dimensions), ShouldEqual, 2)
		So(p.DatasetLandingPage.Dimensions[0].Title, ShouldEqual, "Age")
		So(len(p.DatasetLandingPage.Dimensions[0].Values), ShouldEqual, 5)
		So(p.DatasetLandingPage.Dimensions[0].Values[0], ShouldEqual, "3")
		So(p.DatasetLandingPage.Dimensions[0].Values[1], ShouldEqual, "6")
		So(p.DatasetLandingPage.Dimensions[0].Values[2], ShouldEqual, "19")
		So(p.DatasetLandingPage.Dimensions[0].Values[3], ShouldEqual, "23")
		So(p.DatasetLandingPage.Dimensions[0].Values[4], ShouldEqual, "24")
		So(len(p.DatasetLandingPage.Dimensions[1].Values), ShouldEqual, 1)
		So(p.DatasetLandingPage.Dimensions[1].Title, ShouldEqual, "Time")
		So(p.DatasetLandingPage.Dimensions[1].Values[0], ShouldEqual, "All months between January 2005 and February 2005")

		v0 := p.DatasetLandingPage.Version
		So(v0.Title, ShouldEqual, d.Title)
		So(v0.Description, ShouldEqual, d.Description)
		So(v0.Edition, ShouldEqual, v[0].Edition)
		So(v0.Version, ShouldEqual, strconv.Itoa(v[0].Version))
		So(v0.ReleaseDate, ShouldEqual, v[0].ReleaseDate)
		So(v0.Downloads[0].Size, ShouldEqual, "438290")
		So(v0.Downloads[0].Extension, ShouldEqual, "XLSX")
		So(v0.Downloads[0].URI, ShouldEqual, "my-url")
	})

	Convey("test taxonomy domain is set on page when environment variable is set", t, func() {
		os.Setenv("TAXONOMY_DOMAIN", "my-domain")

		p := CreateFilterableLandingPage(dataset.Model{}, dataset.Version{}, "", []dataset.Options{}, dataset.Dimensions{}, false)

		So(p.TaxonomyDomain, ShouldEqual, "my-domain")

	})
}
