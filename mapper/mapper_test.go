package mapper

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/ONSdigital/dp-frontend-dataset-controller/data"
	"github.com/ONSdigital/dp-frontend-models/model/datasetLandingPageFilterable"
	"github.com/ONSdigital/go-ns/zebedee/zebedeeMapper"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitMapper(t *testing.T) {
	Convey("test CreateFilterableLandingPage", t, func() {
		ds := getTestDatasets()
		dims := getTestDimensions()
		sp := getTestStaticLandingPage()

		p := CreateFilterableLandingPage(ds, dims, sp)
		So(p.Type, ShouldEqual, sp.Type)
		So(p.URI, ShouldEqual, sp.URI)
		So(p.Metadata, ShouldResemble, sp.Metadata)
		So(p.DatasetLandingPage.DatasetLandingPage, ShouldResemble, sp.DatasetLandingPage)
		So(p.Breadcrumb, ShouldResemble, sp.Breadcrumb)
		So(p.ContactDetails, ShouldResemble, sp.ContactDetails)

		v0 := p.DatasetLandingPage.Versions[0]

		So(v0.Title, ShouldEqual, ds[0].Title)
		So(v0.Description, ShouldEqual, sp.DatasetLandingPage.MetaDescription)
		So(v0.Edition, ShouldEqual, ds[0].Edition)
		So(v0.Version, ShouldEqual, ds[0].Version)
		So(v0.ReleaseDate, ShouldEqual, ds[0].ReleaseDate)
		So(v0.Downloads[0], ShouldResemble, datasetLandingPageFilterable.Download(sp.DatasetLandingPage.Datasets[0].Downloads[0]))

		d0 := p.DatasetLandingPage.Dimensions[0]

		So(d0.Title, ShouldResemble, dims[0].Name)
		So(d0.Values, ShouldResemble, dims[0].Values)
	})
}

func getTestDatasets() []data.Dataset {
	return []data.Dataset{
		{
			ID:          "12345",
			Title:       "Dataset title",
			URL:         "google.com",
			ReleaseDate: "11 November 2017",
			NextRelease: "11 November 2018",
			Edition:     "2017",
			Version:     "1",
			Contact: data.Contact{
				Name:      "Matt Rout",
				Telephone: "012346 382012",
				Email:     "matt@gmail.com",
			},
		},
	}
}

func getTestDimensions() []data.Dimension {
	return []data.Dimension{
		{
			CodeListID: "ABDCSKA",
			ID:         "siojxuidhc",
			Name:       "Geography",
			Type:       "Hierarchy",
			Values:     []string{"Region", "County"},
		},
		{
			CodeListID: "AHDHSID",
			ID:         "eorihfieorf",
			Name:       "Age List",
			Type:       "List",
			Values:     []string{"0", "1", "2"},
		},
	}
}

func getTestStaticLandingPage() zebedeeMapper.StaticDatasetLandingPage {
	var sdlp zebedeeMapper.StaticDatasetLandingPage

	f, _ := ioutil.ReadFile("testdata/staticlandingpage.json")

	json.Unmarshal(f, &sdlp)
	return sdlp
}
