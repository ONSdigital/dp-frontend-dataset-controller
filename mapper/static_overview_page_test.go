package mapper

import (
	"context"
	"testing"
	"time"

	core "github.com/ONSdigital/dis-design-system-go/model"
	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	dpTopicApiModels "github.com/ONSdigital/dp-topic-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateStaticOverviewPageRelatedContent(t *testing.T) {
	Convey("Given a dataset with related content", t, func() {
		dataset := dpDatasetApiModels.Dataset{
			RelatedContent: []dpDatasetApiModels.GeneralDetails{
				{Title: "Dataset content", HRef: "/dataset-content", Description: "Dataset description"},
			},
		}
		version := dpDatasetApiModels.Version{}
		topics := []*dpTopicApiModels.Topic{{Slug: "economy"}}
		expected := []sharedModel.RelatedContentItem{
			{Title: "Dataset content", Link: "/dataset-content", Text: "Dataset description"},
		}

		Convey("When the version has related content", func() {
			relatedContent := []dpDatasetApiModels.GeneralDetails{
				{Title: "First version content", HRef: "/version-content-1", Description: "First version description"},
				{Title: "Second version content", HRef: "/version-content-2", Description: "Second version description"},
			}
			version.RelatedContent = &relatedContent

			page := CreateStaticOverviewPage(context.Background(), core.Page{}, dataset, version, nil, false, topics, false, false)

			Convey("Then version items are mapped and appended after dataset content in order", func() {
				expected = append(expected,
					sharedModel.RelatedContentItem{Title: "First version content", Link: "/version-content-1", Text: "First version description"},
					sharedModel.RelatedContentItem{Title: "Second version content", Link: "/version-content-2", Text: "Second version description"},
				)
				So(page.DatasetLandingPage.RelatedContentItems, ShouldResemble, expected)
			})
		})

		Convey("When version related content is nil", func() {
			page := CreateStaticOverviewPage(context.Background(), core.Page{}, dataset, version, nil, false, topics, false, false)

			Convey("Then dataset related content is preserved", func() {
				So(page.DatasetLandingPage.RelatedContentItems, ShouldResemble, expected)
			})
		})

		Convey("When version related content points to an empty slice", func() {
			relatedContent := []dpDatasetApiModels.GeneralDetails{}
			version.RelatedContent = &relatedContent

			page := CreateStaticOverviewPage(context.Background(), core.Page{}, dataset, version, nil, false, topics, false, false)

			Convey("Then dataset related content is preserved", func() {
				So(page.DatasetLandingPage.RelatedContentItems, ShouldResemble, expected)
			})
		})
	})
}

func TestSetGTMDataLayerValuesForStaticDatasets(t *testing.T) {
	ctx := context.Background()

	Convey("Given a static dataset and version with full details", t, func() {
		lastUpdated := time.Date(2025, 6, 10, 12, 0, 0, 0, time.UTC)
		dataset := dpDatasetApiModels.Dataset{
			ID:     "dataset-123",
			Title:  "Producer price inflation (MM22)",
			Topics: []string{"topic-economy-id"},
			Links: &dpDatasetApiModels.DatasetLinks{
				LatestVersion: &dpDatasetApiModels.LinkObject{ID: "3"},
			},
		}
		version := dpDatasetApiModels.Version{
			Version:      3,
			Edition:      "2025",
			EditionTitle: "2025 edition",
			ReleaseDate:  "2025-01-15T00:00:00.000Z",
			LastUpdated:  lastUpdated,
		}
		topics := []*dpTopicApiModels.Topic{
			{ID: "topic-economy-id", Slug: "economy", Title: "Economy"},
		}

		Convey("When setGTMDataLayerValuesForStaticDatasets is called", func() {
			dataLayer := setGTMDataLayerValuesForStaticDatasets(ctx, dataset, version, topics)

			Convey("Then the data layer should contain the expected analytics values", func() {
				So(dataLayer, ShouldResemble, map[string]string{
					"product":        "dataset-catalogue",
					"contentType":    "datasets",
					"contentSubtype": "versions",
					"contentGroup":   "economy",
					"contentTitle":   "Producer price inflation (MM22): 2025 edition",
					"outputSeries":   "dataset-123",
					"outputEdition":  "2025",
					"outputVersion":  "3",
					"releaseDate":    "20250115",
					"lastUpdateDate": "20250610",
					"latestRelease":  "yes",
				})
			})
		})
	})

	Convey("Given a version that is not the latest release", t, func() {
		dataset := dpDatasetApiModels.Dataset{
			ID:    "dataset-123",
			Title: "Producer price inflation (MM22)",
			Links: &dpDatasetApiModels.DatasetLinks{
				LatestVersion: &dpDatasetApiModels.LinkObject{ID: "3"},
			},
		}
		version := dpDatasetApiModels.Version{
			Version:      2,
			Edition:      "2024",
			EditionTitle: "2024 edition",
		}
		topics := []*dpTopicApiModels.Topic{
			{ID: "topic-economy-id", Slug: "economy", Title: "Economy"},
		}

		Convey("When setGTMDataLayerValuesForStaticDatasets is called", func() {
			dataLayer := setGTMDataLayerValuesForStaticDatasets(ctx, dataset, version, topics)

			Convey("Then latestRelease should be no", func() {
				So(dataLayer["latestRelease"], ShouldEqual, "no")
			})
		})
	})

	Convey("Given no topics", t, func() {
		dataset := dpDatasetApiModels.Dataset{
			ID:    "dataset-123",
			Title: "Producer price inflation (MM22)",
			Links: &dpDatasetApiModels.DatasetLinks{
				LatestVersion: &dpDatasetApiModels.LinkObject{ID: "1"},
			},
		}
		version := dpDatasetApiModels.Version{
			Version:      1,
			Edition:      "2025",
			EditionTitle: "2025 edition",
		}
		topics := []*dpTopicApiModels.Topic{}

		Convey("When setGTMDataLayerValuesForStaticDatasets is called", func() {
			dataLayer := setGTMDataLayerValuesForStaticDatasets(ctx, dataset, version, topics)

			Convey("Then contentGroup should not be set", func() {
				_, ok := dataLayer["contentGroup"]
				So(ok, ShouldBeFalse)
			})
		})
	})

	Convey("Given a version with no release date", t, func() {
		dataset := dpDatasetApiModels.Dataset{
			ID:    "dataset-123",
			Title: "Producer price inflation (MM22)",
			Links: &dpDatasetApiModels.DatasetLinks{
				LatestVersion: &dpDatasetApiModels.LinkObject{ID: "1"},
			},
		}
		version := dpDatasetApiModels.Version{
			Version:      1,
			Edition:      "2025",
			EditionTitle: "2025 edition",
			ReleaseDate:  "",
		}
		topics := []*dpTopicApiModels.Topic{
			{ID: "topic-economy-id", Slug: "economy", Title: "Economy"},
		}

		Convey("When setGTMDataLayerValuesForStaticDatasets is called", func() {
			dataLayer := setGTMDataLayerValuesForStaticDatasets(ctx, dataset, version, topics)

			Convey("Then releaseDate should not be set", func() {
				_, ok := dataLayer["releaseDate"]
				So(ok, ShouldBeFalse)
			})
		})
	})

	Convey("Given a version with an invalid release date", t, func() {
		dataset := dpDatasetApiModels.Dataset{
			ID:    "dataset-123",
			Title: "Producer price inflation (MM22)",
			Links: &dpDatasetApiModels.DatasetLinks{
				LatestVersion: &dpDatasetApiModels.LinkObject{ID: "1"},
			},
		}
		version := dpDatasetApiModels.Version{
			Version:      1,
			Edition:      "2025",
			EditionTitle: "2025 edition",
			ReleaseDate:  "not-a-date",
		}
		topics := []*dpTopicApiModels.Topic{
			{ID: "topic-economy-id", Slug: "economy", Title: "Economy"},
		}

		Convey("When setGTMDataLayerValuesForStaticDatasets is called", func() {
			dataLayer := setGTMDataLayerValuesForStaticDatasets(ctx, dataset, version, topics)

			Convey("Then releaseDate should not be set", func() {
				_, ok := dataLayer["releaseDate"]
				So(ok, ShouldBeFalse)
			})
		})
	})
}
