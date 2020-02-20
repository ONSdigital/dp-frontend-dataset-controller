package mapper

import (
	"context"
	"strconv"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/go-ns/zebedee/data"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitMapper(t *testing.T) {
	ctx := context.Background()

	Convey("test CreateFilterableLandingPage", t, func() {
		contact := dataset.Contact{
			Name:      "Matt Rout",
			Telephone: "01622 734721",
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

		breadcrumbItem := data.Breadcrumb{
			URI:         "/economy/grossdomesticproduct/datasets/gdpjanuary2018",
			Description: data.NodeDescription{Title: "GDP: January 2018"},
		}

		p := CreateFilterableLandingPage(ctx, d, v[0], datasetID, []dataset.Options{
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
		}, dataset.Dimensions{}, false, []data.Breadcrumb{breadcrumbItem}, 1, "/datasets/83jd98fkflg/editions/124/versions/1", true, false)

		So(p.Type, ShouldEqual, "dataset_landing_page")
		So(p.Metadata.Title, ShouldEqual, d.Title)
		So(p.URI, ShouldEqual, d.Links.Self.URL)
		So(p.ShowFeedbackForm, ShouldEqual, true)
		So(p.ContactDetails.Name, ShouldEqual, contact.Name)
		So(p.ContactDetails.Telephone, ShouldEqual, contact.Telephone)
		So(p.ContactDetails.Email, ShouldEqual, contact.Email)
		So(p.DatasetLandingPage.NextRelease, ShouldEqual, d.NextRelease)
		So(p.DatasetLandingPage.DatasetID, ShouldEqual, datasetID)
		So(p.ReleaseDate, ShouldEqual, v[0].ReleaseDate)
		So(p.ShowFeedbackForm, ShouldEqual, true)
		So(p.Breadcrumb[0].Title, ShouldEqual, breadcrumbItem.Description.Title)
		So(p.Breadcrumb[0].URI, ShouldEqual, breadcrumbItem.URI)
		So(p.EnableLoop11, ShouldEqual, true)

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
		So(p.ReleaseDate, ShouldEqual, v[0].ReleaseDate)
		So(v0.Downloads[0].Size, ShouldEqual, "438290")
		So(v0.Downloads[0].Extension, ShouldEqual, "XLSX")
		So(v0.Downloads[0].URI, ShouldEqual, "my-url")
	})

}

// TestCreateVersionsList Tests the CreateVersionsList function in the mapper
func TestCreateVersionsList(t *testing.T) {
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

	ctx := context.Background()
	Convey("test latest version page", t, func() {
		dummySingleVersionList := []dataset.Version{dummyVersion3}

		page := CreateVersionsList(ctx, dummyModelData, dummyEditionData, dummySingleVersionList, false, false)
		Convey("title", func() {
			So(page.Metadata.Title, ShouldEqual, "All versions of Consumer Prices Index including owner occupiers? housing costs (CPIH) time-series dataset")
		})
		Convey("has correct number of versions when only one should be present", func() {
			So(len(page.Data.Versions), ShouldEqual, 1)
		})

		dummyMultipleVersionList := []dataset.Version{dummyVersion1, dummyVersion2, dummyVersion3}
		page = CreateVersionsList(ctx, dummyModelData, dummyEditionData, dummyMultipleVersionList, false, false)

		Convey("has correct number of versions when multiple should be present", func() {
			So(len(page.Data.Versions), ShouldEqual, 3)
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
