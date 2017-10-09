package mapper

import (
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
			NextRelease: "11-11-2018",
			Periodicity: "yearly",
			Publisher: dataset.Publisher{
				URL:  "ons.gov.uk",
				Name: "ONS",
				Type: "Government Agency",
			},
			State: "created",
			Theme: "purple",
			Title: "Penguins of the Antarctic Ocean",
		}

		v := []dataset.Version{
			dataset.Version{
				CollectionID: "abcdefg",
				Edition:      "2017",
				ID:           "tehnskofjios-ashbc7",
				InstanceID:   "31241592",
				License:      "ons",
				Version:      1,
				Links: dataset.Links{
					Self: dataset.Link{
						URL: "/datasets/83jd98fkflg/editions/124/versions/1",
					},
				},
				ReleaseDate: "11-11-2017",
				State:       "published",
			},
		}
		datasetID := "038847784-2874757-23784854905"

		p := CreateFilterableLandingPage(d, v, datasetID, []dataset.Options{})

		So(p.Type, ShouldEqual, "dataset_landing_page")
		So(p.Metadata.Title, ShouldEqual, d.Title)
		So(p.URI, ShouldEqual, d.Links.Self.URL)
		So(p.Metadata.Footer.Contact, ShouldEqual, d.Contacts[0].Name)
		So(p.Metadata.Footer.DatasetID, ShouldEqual, datasetID)
		So(p.ContactDetails.Name, ShouldEqual, d.Contacts[0].Name)
		So(p.ContactDetails.Telephone, ShouldEqual, d.Contacts[0].Telephone)
		So(p.ContactDetails.Email, ShouldEqual, d.Contacts[0].Email)
		So(p.DatasetLandingPage.NextRelease, ShouldEqual, d.NextRelease)
		So(p.DatasetLandingPage.DatasetID, ShouldEqual, datasetID)
		So(p.DatasetLandingPage.ReleaseDate, ShouldEqual, v[0].ReleaseDate)

		v0 := p.DatasetLandingPage.Versions[0]
		So(v0.Title, ShouldEqual, d.Title)
		So(v0.Description, ShouldEqual, d.Description)
		So(v0.Edition, ShouldEqual, v[0].Edition)
		So(v0.Version, ShouldEqual, strconv.Itoa(v[0].Version))
		So(v0.ReleaseDate, ShouldEqual, v[0].ReleaseDate)
		So(v0.Downloads[0].Size, ShouldEqual, "438290")
		So(v0.Downloads[0].Extension, ShouldEqual, "XLSX")
	})
}
