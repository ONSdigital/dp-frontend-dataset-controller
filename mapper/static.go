package mapper

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
)

// MapStaticDatasetToZebedee maps a dataset of type static from the dataset API to the equivalent zebedee format.
func MapStaticDatasetToZebedee(ctx context.Context, dataset datasetAPIModels.Dataset, topicSlugs []string) (*zebedee.DatasetLandingPage, error) {
	if len(topicSlugs) == 0 {
		return nil, fmt.Errorf("at least one topic slug is required to map a static dataset to zebedee format")
	}

	zebedeeDataset := &zebedee.DatasetLandingPage{
		Description: zebedee.Description{
			DatasetID:       dataset.ID,
			Title:           dataset.Title,
			Summary:         dataset.Description,
			MetaDescription: dataset.Description,
			Contact:         mapContactsToZebedeeContact(dataset.Contacts),
			Keywords:        dataset.Keywords,
			NextRelease:     dataset.NextRelease,
			CanonicalTopic:  topicSlugs[0],
			Topics:          topicSlugs[1:],
		},
		Type: zebedee.PageTypeDatasetLandingPage, // "dataset_landing_page" is the zebedee equivalent for a "dataset" in the dataset API
		URI:  fmt.Sprintf("/%s/datasets/%s", topicSlugs[0], dataset.ID),
	}

	if dataset.QMI != nil {
		parsedQMIURL, err := url.Parse(dataset.QMI.HRef)
		if err != nil {
			return nil, fmt.Errorf("failed to parse QMI URL: %w", err)
		}

		zebedeeDataset.RelatedMethodology = []zebedee.Link{
			{
				Title:   dataset.QMI.Title,
				Summary: dataset.QMI.Description,
				URI:     parsedQMIURL.Path,
			},
		}
	}

	return zebedeeDataset, nil
}

// MapStaticVersionToZebedee maps a version of type static from the dataset API to the equivalent zebedee format.
func MapStaticVersionToZebedee(dataset datasetAPIModels.Dataset, version datasetAPIModels.Version, previousVersions []datasetAPIModels.Version, topicSlugs []string) (*zebedee.Dataset, error) {
	if len(topicSlugs) == 0 {
		return nil, fmt.Errorf("at least one topic slug is required to map a static version to zebedee format")
	}

	zebedeeVersion := &zebedee.Dataset{
		Description: zebedee.Description{
			DatasetID:       dataset.ID,
			Title:           dataset.Title,
			Summary:         dataset.Description,
			MetaDescription: dataset.Description,
			Contact:         mapContactsToZebedeeContact(dataset.Contacts),
			Keywords:        dataset.Keywords,
			ReleaseDate:     version.ReleaseDate,
			NextRelease:     dataset.NextRelease,
			CanonicalTopic:  topicSlugs[0],
			Topics:          topicSlugs[1:],
		},
		Type:      zebedee.PageTypeDataset, // "dataset" is the zebedee equivalent for an "edition" or "version" in the dataset API
		Downloads: mapDistributionsToDownloads(version.Distributions),
		URI:       fmt.Sprintf("/%s/datasets/%s/editions/%s/versions/%d", topicSlugs[0], dataset.ID, version.Edition, version.Version),
	}

	editionTitle := version.EditionTitle
	if editionTitle == "Historical" {
		editionTitle = "Current"
	}
	zebedeeVersion.Description.Edition = editionTitle

	if version.QualityDesignation == datasetAPIModels.QualityDesignationAccreditedOfficial {
		zebedeeVersion.Description.NationalStatistic = true
	}

	if len(previousVersions) > 0 {
		zebedeeVersions, err := mapPreviousVersionsToZebedeeVersions(previousVersions, topicSlugs[0], dataset.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to map previous versions to zebedee versions: %w", err)
		}
		zebedeeVersion.Versions = zebedeeVersions
	}

	return zebedeeVersion, nil
}

// mapContactsToZebedeeContact maps dataset API contacts to zebedee contact format.
// The dataset API supports multiple contacts but zebedee only supports a single contact, so the first contact is used.
func mapContactsToZebedeeContact(contacts []datasetAPIModels.ContactDetails) zebedee.Contact {
	if len(contacts) == 0 {
		return zebedee.Contact{}
	}

	return zebedee.Contact{
		Name:      contacts[0].Name,
		Email:     contacts[0].Email,
		Telephone: contacts[0].Telephone,
	}
}

// mapDistributionsToDownloads maps dataset API distributions to zebedee downloads.
func mapDistributionsToDownloads(distributions *[]datasetAPIModels.Distribution) []zebedee.Download {
	if distributions == nil || len(*distributions) == 0 {
		return nil
	}

	downloads := make([]zebedee.Download, len(*distributions))

	for i, distribution := range *distributions {
		downloads[i] = zebedee.Download{
			File: distribution.Title,
			URI:  distribution.DownloadURL,
		}
	}

	return downloads
}

// mapPreviousVersionsToZebedeeVersions maps dataset API previous versions to zebedee versions.
func mapPreviousVersionsToZebedeeVersions(previousVersions []datasetAPIModels.Version, topicSlug, datasetID string) ([]zebedee.Version, error) {
	if len(previousVersions) == 0 {
		return nil, nil
	}

	if topicSlug == "" || datasetID == "" {
		return nil, fmt.Errorf("topic slug and dataset ID are required to map previous versions to zebedee versions")
	}

	zebedeeVersions := make([]zebedee.Version, len(previousVersions))

	for i := range previousVersions {
		version := previousVersions[i]
		zebedeeVersions[i] = zebedee.Version{
			URI:         fmt.Sprintf("/%s/datasets/%s/editions/%s/versions/%d", topicSlug, datasetID, version.Edition, version.Version),
			ReleaseDate: version.ReleaseDate,
			Notice:      mapAlertsToZebedeeCorrectionNotice(version.Alerts),
		}
	}

	return zebedeeVersions, nil
}

// mapAlertsToZebedeeCorrectionNotice maps dataset API alerts to a zebedee correction notice.
// This only includes alerts of type "correction".
func mapAlertsToZebedeeCorrectionNotice(alerts *[]datasetAPIModels.Alert) string {
	if alerts == nil || len(*alerts) == 0 {
		return ""
	}

	var b strings.Builder

	for _, alert := range *alerts {
		if alert.Type != datasetAPIModels.AlertTypeCorrection {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString(alert.Description)
	}

	return b.String()
}
