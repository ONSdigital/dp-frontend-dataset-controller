package datasetLandingPageCensus

import (
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/contactDetails"

	"github.com/ONSdigital/dp-renderer/model"
)

// Page contains data for the census landing page
type Page struct {
	model.Page
	DatasetLandingPage DatasetLandingPage            `json:"data"`
	Version            sharedModel.Version           `json:"version"`
	Versions           []sharedModel.Version         `json:"versions"`
	ID                 string                        `json:"id"`
	ContactDetails     contactDetails.ContactDetails `json:"contact_details"`
	HasContactDetails  bool                          `json:"has_contact_details"`
}

// DatasetLandingPage contains properties related to the census dataset landing page
type DatasetLandingPage struct {
	HasOtherVersions bool                    `json:"has_other_versions"`
	HasDownloads     bool                    `json:"has_downloads"`
	LatestVersionURL string                  `json:"latest_version_url"`
	Dimensions       []sharedModel.Dimension `json:"dimensions"`
	ShareDetails     ShareDetails
	Description      []string `json:"description"`
	IsFlexibleForm   bool     `json:"is_flexible_form"`
	FormAction       string   `json:"form_action"`
	DatasetURL       string   `json:"dataset_url"`
	Panels           []Panel  `json:"panels"`
}

// Panel contains the data required to populate a panel UI component
type Panel struct {
	IsCorrection bool `json:"is_correction"`
}

// ShareDetails contains the locations the page can be shared to, as well as the language attribute for localisation
type ShareDetails struct {
	ShareLocations []Share `json:"share_locations"`
	Language       string  `json:"language"`
}

/* Share includes details for a specific place the dataset can be shared
   Included icons: 'facebook', 'twitter', 'email', 'linkedin'
*/
type Share struct {
	Title string `json:"title"`
	Link  string `json:"link"`
	Icon  string `json:"icon"`
}
