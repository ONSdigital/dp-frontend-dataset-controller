package mapper

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-frontend-dataset-controller/clients"
	topicAPIModels "github.com/ONSdigital/dp-topic-api/models"
	topicAPISDK "github.com/ONSdigital/dp-topic-api/sdk"
	topicAPISDKErrors "github.com/ONSdigital/dp-topic-api/sdk/errors"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMapStaticDatasetToZebedee(t *testing.T) {
	ctx := context.Background()

	Convey("Given a static dataset and a topic API client", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTopicClient := clients.NewMockTopicAPIClient(ctrl)

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

		Convey("When MapStaticDatasetToZebedee is called", func() {
			mockTopicClient.EXPECT().
				GetTopicPublic(ctx, topicAPISDK.Headers{}, "economy").
				Return(&topicAPIModels.Topic{ID: "economy", Slug: "economy"}, nil).
				Times(1)

			mockTopicClient.EXPECT().
				GetTopicPublic(ctx, topicAPISDK.Headers{}, "inflation").
				Return(&topicAPIModels.Topic{ID: "inflation", Slug: "inflation"}, nil).
				Times(1)

			result, err := MapStaticDatasetToZebedee(ctx, dataset, mockTopicClient)

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
						Topics:          []string{"economy", "inflation"},
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

		Convey("When the dataset QMI URL is invalid", func() {
			datasetWithInvalidQMI := dataset
			datasetWithInvalidQMI.QMI.HRef = "://invalid-url"

			result, err := MapStaticDatasetToZebedee(ctx, datasetWithInvalidQMI, mockTopicClient)

			Convey("Then an error should be returned indicating the QMI URL is invalid", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to parse QMI URL")
				So(result, ShouldBeNil)
			})
		})

		Convey("When the topic API client returns an error", func() {
			mockTopicClient.EXPECT().
				GetTopicPublic(ctx, topicAPISDK.Headers{}, "economy").
				Return(nil, topicAPISDKErrors.StatusError{Code: 500, Err: errors.New("topic API error")}).
				Times(1)

			result, err := MapStaticDatasetToZebedee(ctx, dataset, mockTopicClient)

			Convey("Then an error should be returned indicating the topic could not be fetched", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to get topic with ID economy")
				So(result, ShouldBeNil)
			})
		})
	})
}
