package mapper

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-frontend-dataset-controller/clients"
	topicAPISDK "github.com/ONSdigital/dp-topic-api/sdk"
)

// MapStaticDatasetToZebedee maps a dataset of type static from the dataset API to the equivalent zebedee format.
func MapStaticDatasetToZebedee(ctx context.Context, dataset datasetAPIModels.Dataset, topicAPIClient clients.TopicAPIClient) (*zebedee.DatasetLandingPage, error) {
	zebedeeDataset := &zebedee.DatasetLandingPage{
		Description: zebedee.Description{
			DatasetID:       dataset.ID,
			Title:           dataset.Title,
			Summary:         dataset.Description,
			MetaDescription: dataset.Description,
			Keywords:        dataset.Keywords,
			NextRelease:     dataset.NextRelease,
		},
		Type: zebedee.PageTypeDatasetLandingPage, // "dataset_landing_page" is the zebedee equivalent for a "dataset" in the dataset API
	}

	if dataset.QMI != nil {
		parsedQMIURL, err := url.Parse(dataset.QMI.HRef)
		if err != nil {
			return nil, fmt.Errorf("failed to parse QMI URL: %w", err)
		}

		zebedeeDataset.RelatedMethodology = []zebedee.Related{
			zebedee.Link{
				Title:   dataset.QMI.Title,
				Summary: dataset.QMI.Description,
				URI:     parsedQMIURL.Path,
			},
		}
	}

	// The dataset API returns a list of topic IDs, but zebedee requires a list of topic slugs
	topicSlugs := make([]string, len(dataset.Topics))
	for i, topicID := range dataset.Topics {
		topic, err := topicAPIClient.GetTopicPublic(ctx, topicAPISDK.Headers{}, topicID)
		if err != nil {
			return nil, fmt.Errorf("failed to get topic with ID %s: %w", topicID, err)
		}
		topicSlugs[i] = topic.Slug
	}

	zebedeeDataset.Description.Topics = topicSlugs
	zebedeeDataset.URI = fmt.Sprintf("/%s/datasets/%s", topicSlugs[0], dataset.ID)

	// Contacts is a mandatory field so dataset.Contacts[0] should always exist.
	// The dataset API supports multiple contacts but zebedee only supports a single contact, so the first contact is used here.
	zebedeeDataset.Description.Contact = zebedee.Contact{
		Name:      dataset.Contacts[0].Name,
		Email:     dataset.Contacts[0].Email,
		Telephone: dataset.Contacts[0].Telephone,
	}

	return zebedeeDataset, nil
}
