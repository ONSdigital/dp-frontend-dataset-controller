package mapper

import (
	"testing"

	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
	dpTopicApiModels "github.com/ONSdigital/dp-topic-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateStaticBasePage(t *testing.T) {
	Convey("Given a topicObjectList with two topics", t, func() {
		basePage := coreModel.Page{}
		dataset := dpDatasetApiModels.Dataset{}
		version := dpDatasetApiModels.Version{}
		allVersions := []dpDatasetApiModels.Version{}
		isEnableMultivariate := false

		topicObjectList := []dpTopicApiModels.Topic{
			{Title: "Topic One"},
			{Title: "Topic Two"},
		}

		Convey("When CreateStaticBasePage is called", func() {
			breadcrumbObject := CreateStaticBasePage(basePage, dataset, version, allVersions, isEnableMultivariate, topicObjectList)

			Convey("Then the breadcrumbs object should contain a TaxonomyNode for each topic ", func() {
				So(breadcrumbObject.Breadcrumb, ShouldHaveLength, 2)
				So(breadcrumbObject.Breadcrumb[0].Title, ShouldEqual, "Topic One")
				So(breadcrumbObject.Breadcrumb[0].URI, ShouldEqual, "#")
				So(breadcrumbObject.Breadcrumb[1].Title, ShouldEqual, "Topic Two")
				So(breadcrumbObject.Breadcrumb[1].URI, ShouldEqual, "#")
			})
		})
	})
}
