package mapper

import (
	"testing"

	dpTopicApiModels "github.com/ONSdigital/dp-topic-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateStaticBasePage(t *testing.T) {
	Convey("Given a topicObjectList with two topics", t, func() {
		topicObjectList := []dpTopicApiModels.Topic{
			{Title: "Topic One", Slug: "slug1"},
			{Title: "Topic Two", Slug: "slug2"},
		}

		baseURL := "https://www.ons.gov.uk/"

		Convey("When CreateBreadcrumbsFromTopicList is called to build the breadcrumbs from the topics list", func() {
			breadcrumbObject := CreateBreadcrumbsFromTopicList(baseURL, topicObjectList)

			Convey("Then the breadcrumbs object should contain a TaxonomyNode for each topic ", func() {
				So(breadcrumbObject, ShouldHaveLength, 3)
				So(breadcrumbObject[0].Title, ShouldEqual, "Home")
				So(breadcrumbObject[0].URI, ShouldEqual, "https://www.ons.gov.uk/")
				So(breadcrumbObject[1].Title, ShouldEqual, "Topic One")
				So(breadcrumbObject[1].URI, ShouldEqual, "https://www.ons.gov.uk/slug1")
				So(breadcrumbObject[2].Title, ShouldEqual, "Topic Two")
				So(breadcrumbObject[2].URI, ShouldEqual, "https://www.ons.gov.uk/slug1/slug2")
			})
		})
	})
}
