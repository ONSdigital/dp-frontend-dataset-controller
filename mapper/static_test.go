package mapper

import (
	"context"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMapStaticDatasetToZebedee(t *testing.T) {
	ctx := context.Background()

	Convey("Given a static dataset and a list of topic slugs", t, func() {
		dataset := datasetAPIModels.Dataset{
			ID:          "dataset-123",
			Title:       "Producer price inflation (MM22)",
			Description: "UK price index data at manufacturing, aggregated industry and product group levels. Data supplied from individual manufacturers, importers and exporters. Monthly and annual data.",
			Keywords:    []string{"manufacturing", "input prices", "output prices", "producer prices"},
			NextRelease: "To be announced",
			Topics:      []string{"economy", "inflation"},
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
		}
		topicSlugs := []string{"economy", "inflation"}

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
					RelatedMethodology: []zebedee.Related{
						zebedee.Link{
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
