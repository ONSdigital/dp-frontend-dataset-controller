package mapper

import (
	"context"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testTopicSlugs = []string{"economy", "inflation"}

	testStaticDataset = datasetAPIModels.Dataset{
		ID:          "dataset-123",
		Title:       "Producer price inflation (MM22)",
		Description: "UK price index data at manufacturing, aggregated industry and product group levels. Data supplied from individual manufacturers, importers and exporters. Monthly and annual data.",
		Keywords:    []string{"manufacturing", "input prices", "output prices", "producer prices"},
		NextRelease: "To be announced",
		Topics:      []string{"topic-economy-id", "topic-inflation-id"},
		Contacts: []datasetAPIModels.ContactDetails{
			{
				Name:      "Business Prices team",
				Email:     "business.prices@ons.gov.uk",
				Telephone: "+44 1633 456907",
			},
		},
		QMI: &datasetAPIModels.GeneralDetails{
			Title:       "Producer Prices QMI",
			Description: "Quality and Methodology Information for Producer Price Indices",
			HRef:        "https://www.ons.gov.uk/economy/inflationandpriceindices/qmis/producerpriceindicesqmi",
		},
		Type: datasetAPIModels.Static.String(),
	}

	testStaticVersion = datasetAPIModels.Version{
		Version:            3,
		Edition:            "2025",
		EditionTitle:       "2025 edition",
		ReleaseDate:        "2025-01-15T00:00:00.000Z",
		QualityDesignation: datasetAPIModels.QualityDesignationAccreditedOfficial,
		Distributions: &[]datasetAPIModels.Distribution{
			{
				Title:       "file.csv",
				DownloadURL: "http://localhost:23600/downloads/file.csv",
			},
			{
				Title:       "file.xls",
				DownloadURL: "http://localhost:23600/downloads/file.xls",
			},
		},
	}

	testStaticPreviousVersions = []datasetAPIModels.Version{
		{
			Version:     2,
			Edition:     "2024",
			ReleaseDate: "2024-01-15T00:00:00.000Z",
			Alerts: &[]datasetAPIModels.Alert{
				{Type: datasetAPIModels.AlertTypeCorrection, Description: "Correction for version 2"},
			},
		},
		{
			Version:     1,
			Edition:     "2023",
			ReleaseDate: "2023-01-15T00:00:00.000Z",
			Alerts: &[]datasetAPIModels.Alert{
				{Type: datasetAPIModels.AlertTypeAlert, Description: "Alert for version 1"},
			},
		},
	}
)

func TestMapStaticDatasetToZebedee(t *testing.T) {
	ctx := context.Background()

	Convey("Given a static dataset and a list of topic slugs", t, func() {
		dataset := testStaticDataset
		topicSlugs := testTopicSlugs

		Convey("When MapStaticDatasetToZebedee is called", func() {
			result, err := MapStaticDatasetToZebedee(ctx, dataset, topicSlugs)

			Convey("Then the result should be the expected zebedee dataset and no error should be returned", func() {
				expected := &zebedee.DatasetLandingPage{
					Type: zebedee.PageTypeDatasetLandingPage,
					URI:  "/economy/datasets/dataset-123",
					Description: zebedee.Description{
						DatasetID:       "dataset-123",
						Title:           "Producer price inflation (MM22)",
						Summary:         "UK price index data at manufacturing, aggregated industry and product group levels. Data supplied from individual manufacturers, importers and exporters. Monthly and annual data.",
						MetaDescription: "UK price index data at manufacturing, aggregated industry and product group levels. Data supplied from individual manufacturers, importers and exporters. Monthly and annual data.",
						Keywords:        []string{"manufacturing", "input prices", "output prices", "producer prices"},
						NextRelease:     "To be announced",
						CanonicalTopic:  "economy",
						Topics:          []string{"inflation"},
						Contact: zebedee.Contact{
							Name:      "Business Prices team",
							Email:     "business.prices@ons.gov.uk",
							Telephone: "+44 1633 456907",
						},
					},
					RelatedMethodology: []zebedee.Link{
						{
							Title:   "Producer Prices QMI",
							Summary: "Quality and Methodology Information for Producer Price Indices",
							URI:     "/economy/inflationandpriceindices/qmis/producerpriceindicesqmi",
						},
					},
				}
				So(err, ShouldBeNil)
				So(result, ShouldResemble, expected)
			})
		})

		Convey("When no topic slugs are provided", func() {
			result, err := MapStaticDatasetToZebedee(ctx, dataset, []string{})

			Convey("Then an error should be returned indicating at least one topic slug is required", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "at least one topic slug is required to map a static dataset to zebedee format")
				So(result, ShouldBeNil)
			})
		})

		Convey("When the dataset QMI URL is invalid", func() {
			datasetWithInvalidQMI := dataset
			datasetWithInvalidQMI.QMI.HRef = "://invalid-url"

			result, err := MapStaticDatasetToZebedee(ctx, datasetWithInvalidQMI, topicSlugs)

			Convey("Then an error should be returned indicating the QMI URL is invalid", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to parse QMI URL")
				So(result, ShouldBeNil)
			})
		})
	})
}

func TestMapStaticVersionToZebedee(t *testing.T) {
	Convey("Given a static dataset, version, previous versions and topic slugs", t, func() {
		dataset := testStaticDataset
		version := testStaticVersion
		previousVersions := testStaticPreviousVersions
		topicSlugs := testTopicSlugs

		Convey("When MapStaticVersionToZebedee is called", func() {
			result, err := MapStaticVersionToZebedee(dataset, version, previousVersions, topicSlugs)

			Convey("Then the result should be the expected zebedee dataset version and no error should be returned", func() {
				expected := &zebedee.Dataset{
					Description: zebedee.Description{
						DatasetID:         "dataset-123",
						Title:             "Producer price inflation (MM22)",
						Summary:           "UK price index data at manufacturing, aggregated industry and product group levels. Data supplied from individual manufacturers, importers and exporters. Monthly and annual data.",
						MetaDescription:   "UK price index data at manufacturing, aggregated industry and product group levels. Data supplied from individual manufacturers, importers and exporters. Monthly and annual data.",
						Contact:           zebedee.Contact{Name: "Business Prices team", Email: "business.prices@ons.gov.uk", Telephone: "+44 1633 456907"},
						Keywords:          []string{"manufacturing", "input prices", "output prices", "producer prices"},
						ReleaseDate:       "2025-01-15T00:00:00.000Z",
						NextRelease:       "To be announced",
						CanonicalTopic:    "economy",
						Topics:            []string{"inflation"},
						Edition:           "2025 edition",
						NationalStatistic: true,
					},
					Type: zebedee.PageTypeDataset,
					Downloads: []zebedee.Download{
						{
							File: "file.csv",
							URI:  "http://localhost:23600/downloads/file.csv",
						},
						{
							File: "file.xls",
							URI:  "http://localhost:23600/downloads/file.xls",
						},
					},
					URI: "/economy/datasets/dataset-123/editions/2025/versions/3",
					Versions: []zebedee.Version{
						{
							URI:         "/economy/datasets/dataset-123/editions/2024/versions/2",
							ReleaseDate: "2024-01-15T00:00:00.000Z",
							Notice:      "Correction for version 2",
						},
						{
							URI:         "/economy/datasets/dataset-123/editions/2023/versions/1",
							ReleaseDate: "2023-01-15T00:00:00.000Z",
							Notice:      "",
						},
					},
				}

				So(err, ShouldBeNil)
				So(result, ShouldResemble, expected)
			})
		})

		Convey("When no topic slugs are provided", func() {
			result, err := MapStaticVersionToZebedee(dataset, version, previousVersions, []string{})

			Convey("Then an error should be returned indicating at least one topic slug is required", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "at least one topic slug is required to map a static version to zebedee format")
				So(result, ShouldBeNil)
			})
		})

		Convey("When edition title is Historical", func() {
			historicalVersion := version
			historicalVersion.EditionTitle = "Historical"

			result, err := MapStaticVersionToZebedee(dataset, historicalVersion, previousVersions, topicSlugs)

			Convey("Then edition should be mapped to Current", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.Description.Edition, ShouldEqual, "Current")
			})
		})
	})
}

func TestMapContactsToZebedeeContact(t *testing.T) {
	Convey("Given dataset API contacts", t, func() {
		Convey("When contacts are empty", func() {
			result := mapContactsToZebedeeContact([]datasetAPIModels.ContactDetails{})

			Convey("Then an empty zebedee contact should be returned", func() {
				So(result, ShouldResemble, zebedee.Contact{})
			})
		})

		Convey("When contacts contain one item", func() {
			contacts := []datasetAPIModels.ContactDetails{
				{
					Name:      "Business Prices team",
					Email:     "business.prices@ons.gov.uk",
					Telephone: "+44 1633 456907",
				},
			}

			result := mapContactsToZebedeeContact(contacts)

			Convey("Then the first contact should be mapped", func() {
				expected := zebedee.Contact{
					Name:      "Business Prices team",
					Email:     "business.prices@ons.gov.uk",
					Telephone: "+44 1633 456907",
				}

				So(result, ShouldResemble, expected)
			})
		})

		Convey("When contacts contain multiple items", func() {
			contacts := []datasetAPIModels.ContactDetails{
				{
					Name:      "First Contact",
					Email:     "first@example.com",
					Telephone: "111",
				},
				{
					Name:      "Second Contact",
					Email:     "second@example.com",
					Telephone: "222",
				},
			}

			result := mapContactsToZebedeeContact(contacts)

			Convey("Then only the first contact should be mapped", func() {
				expected := zebedee.Contact{
					Name:      "First Contact",
					Email:     "first@example.com",
					Telephone: "111",
				}

				So(result, ShouldResemble, expected)
			})
		})
	})
}

func TestMapDistributionsToDownloads(t *testing.T) {
	Convey("Given dataset API distributions", t, func() {
		Convey("When distributions is nil", func() {
			result := mapDistributionsToDownloads(nil)

			Convey("Then nil should be returned", func() {
				So(result, ShouldBeNil)
			})
		})

		Convey("When distributions is empty", func() {
			distributions := &[]datasetAPIModels.Distribution{}

			result := mapDistributionsToDownloads(distributions)

			Convey("Then nil should be returned", func() {
				So(result, ShouldBeNil)
			})
		})

		Convey("When distributions contains items", func() {
			distributions := &[]datasetAPIModels.Distribution{
				{
					Title:       "file.csv",
					DownloadURL: "http://localhost:23600/downloads/file.csv",
				},
				{
					Title:       "file.xls",
					DownloadURL: "http://localhost:23600/downloads/file.xls",
				},
			}

			result := mapDistributionsToDownloads(distributions)

			Convey("Then distributions should be mapped to zebedee downloads", func() {
				expected := []zebedee.Download{
					{
						File: "file.csv",
						URI:  "http://localhost:23600/downloads/file.csv",
					},
					{
						File: "file.xls",
						URI:  "http://localhost:23600/downloads/file.xls",
					},
				}

				So(result, ShouldResemble, expected)
			})
		})
	})
}

func TestMapPreviousVersionsToZebedeeVersions(t *testing.T) {
	Convey("Given dataset API previous versions", t, func() {
		Convey("When previous versions is empty", func() {
			result, err := mapPreviousVersionsToZebedeeVersions([]datasetAPIModels.Version{}, "economy", "dataset-123")

			Convey("Then nil and no error should be returned", func() {
				So(err, ShouldBeNil)
				So(result, ShouldBeNil)
			})
		})

		Convey("When topic slug is empty", func() {
			previousVersions := []datasetAPIModels.Version{{Version: 1, Edition: "2024"}}

			result, err := mapPreviousVersionsToZebedeeVersions(previousVersions, "", "dataset-123")

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "topic slug and dataset ID are required")
				So(result, ShouldBeNil)
			})
		})

		Convey("When dataset ID is empty", func() {
			previousVersions := []datasetAPIModels.Version{{Version: 1, Edition: "2024"}}

			result, err := mapPreviousVersionsToZebedeeVersions(previousVersions, "economy", "")

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "topic slug and dataset ID are required")
				So(result, ShouldBeNil)
			})
		})

		Convey("When previous versions contain items", func() {
			previousVersions := []datasetAPIModels.Version{
				{
					Version:     1,
					Edition:     "2023",
					ReleaseDate: "2023-01-15T00:00:00.000Z",
				},
				{
					Version:     2,
					Edition:     "2024",
					ReleaseDate: "2024-01-15T00:00:00.000Z",
				},
			}

			result, err := mapPreviousVersionsToZebedeeVersions(previousVersions, "economy", "dataset-123")

			Convey("Then versions should be mapped to zebedee versions", func() {
				expected := []zebedee.Version{
					{
						URI:         "/economy/datasets/dataset-123/editions/2023/versions/1",
						ReleaseDate: "2023-01-15T00:00:00.000Z",
					},
					{
						URI:         "/economy/datasets/dataset-123/editions/2024/versions/2",
						ReleaseDate: "2024-01-15T00:00:00.000Z",
					},
				}

				So(err, ShouldBeNil)
				So(result, ShouldResemble, expected)
			})
		})
	})
}

func TestMapAlertsToZebedeeCorrectionNotice(t *testing.T) {
	Convey("Given dataset API alerts", t, func() {
		Convey("When alerts is nil", func() {
			result := mapAlertsToZebedeeCorrectionNotice(nil)

			Convey("Then an empty string should be returned", func() {
				So(result, ShouldEqual, "")
			})
		})

		Convey("When alerts is empty", func() {
			alerts := &[]datasetAPIModels.Alert{}

			result := mapAlertsToZebedeeCorrectionNotice(alerts)

			Convey("Then an empty string should be returned", func() {
				So(result, ShouldEqual, "")
			})
		})

		Convey("When alerts contain no correction types", func() {
			alerts := &[]datasetAPIModels.Alert{
				{Type: datasetAPIModels.AlertTypeAlert, Description: "general notice"},
				{Type: datasetAPIModels.AlertTypeAlert, Description: "another general notice"},
			}

			result := mapAlertsToZebedeeCorrectionNotice(alerts)

			Convey("Then an empty string should be returned", func() {
				So(result, ShouldEqual, "")
			})
		})

		Convey("When alerts contain correction types", func() {
			alerts := &[]datasetAPIModels.Alert{
				{Type: datasetAPIModels.AlertTypeAlert, Description: "general notice"},
				{Type: datasetAPIModels.AlertTypeCorrection, Description: "Correction one"},
				{Type: datasetAPIModels.AlertTypeAlert, Description: "another general notice"},
				{Type: datasetAPIModels.AlertTypeCorrection, Description: "Correction two"},
			}

			result := mapAlertsToZebedeeCorrectionNotice(alerts)

			Convey("Then only corrections should be included in order separated by new lines", func() {
				So(result, ShouldEqual, "Correction one\nCorrection two")
			})
		})
	})
}
