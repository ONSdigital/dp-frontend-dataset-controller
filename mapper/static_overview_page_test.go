package mapper

import (
	"testing"
	"time"

	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSetGTMDataLayerValuesForStaticDatasets(t *testing.T) {
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

		Convey("When setGTMDataLayerValuesForStaticDatasets is called", func() {
			dataLayer := setGTMDataLayerValuesForStaticDatasets(dataset, version)

			Convey("Then the data layer should contain the expected analytics values", func() {
				So(dataLayer, ShouldResemble, map[string]string{
					"product":        "dataset-catalogue",
					"contentType":    "datasets",
					"contentSubtype": "versions",
					"contentGroup":   "topic-economy-id",
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

		Convey("When setGTMDataLayerValuesForStaticDatasets is called", func() {
			dataLayer := setGTMDataLayerValuesForStaticDatasets(dataset, version)

			Convey("Then latestRelease should be no", func() {
				So(dataLayer["latestRelease"], ShouldEqual, "no")
			})
		})
	})

	Convey("Given a dataset with no topics", t, func() {
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

		Convey("When setGTMDataLayerValuesForStaticDatasets is called", func() {
			dataLayer := setGTMDataLayerValuesForStaticDatasets(dataset, version)

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

		Convey("When setGTMDataLayerValuesForStaticDatasets is called", func() {
			dataLayer := setGTMDataLayerValuesForStaticDatasets(dataset, version)

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

		Convey("When setGTMDataLayerValuesForStaticDatasets is called", func() {
			dataLayer := setGTMDataLayerValuesForStaticDatasets(dataset, version)

			Convey("Then releaseDate should not be set", func() {
				_, ok := dataLayer["releaseDate"]
				So(ok, ShouldBeFalse)
			})
		})
	})
}
