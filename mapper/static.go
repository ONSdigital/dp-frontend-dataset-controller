package mapper

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
)

// MapStaticDatasetToZebedee maps a dataset of type static from the dataset API to the equivalent zebedee format.
func MapStaticDatasetToZebedee(ctx context.Context, dataset datasetAPIModels.Dataset, topicSlugs []string) (*zebedee.DatasetLandingPage, error) {
	zebedeeDataset := &zebedee.DatasetLandingPage{
		Description: zebedee.Description{
			DatasetID:       dataset.ID,
			Title:           dataset.Title,
			Summary:         dataset.Description,
			MetaDescription: dataset.Description,
			Keywords:        dataset.Keywords,
			NextRelease:     dataset.NextRelease,
			CanonicalTopic:  topicSlugs[0],
			Topics:          topicSlugs[1:],
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
