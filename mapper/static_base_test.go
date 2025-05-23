package mapper

import (
	"testing"

	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
	dpTopicApiModels "github.com/ONSdigital/dp-topic-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateStaticBasePage(t *testing.T) {
	basePage := coreModel.Page{}
	dataset := dpDatasetApiModels.Dataset{}
	allVersions := []dpDatasetApiModels.Version{}
	isEnableMultivariate := true
	topicObjectList := []dpTopicApiModels.Topic{}

	Convey("If `version.EditionTitle` field value is valid", t, func() {
		editionTitleStr := "My edition title"
		editionSlug := "my-edition-title"
		version := dpDatasetApiModels.Version{
			EditionTitle: editionTitleStr,
			Edition:      editionSlug,
		}
		Convey("When CreateStaticBasePage is called", func() {
			staticPage := CreateStaticBasePage(basePage, dataset, version, allVersions, isEnableMultivariate, topicObjectList)

			Convey("Then the resulting static.Page should have expected values", func() {
				So(staticPage.Version.Edition, ShouldEqual, editionTitleStr)
			})
		})
	})
	Convey("If `version.EditionTitle` field value is not valid", t, func() {
		editionTitleStr := ""
		editionSlug := "my-edition-title"
		version := dpDatasetApiModels.Version{
			EditionTitle: editionTitleStr,
			Edition:      editionSlug,
		}
		Convey("When CreateStaticBasePage is called", func() {
			staticPage := CreateStaticBasePage(basePage, dataset, version, allVersions, isEnableMultivariate, topicObjectList)

			Convey("Then the resulting static.Page should have expected values", func() {
				So(staticPage.Version.Edition, ShouldEqual, editionSlug)
			})
		})
	})
}
