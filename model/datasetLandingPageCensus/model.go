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
	InitialReleaseDate string                        `json:"initial_release_date"`
	ID                 string                        `json:"id"`
	ContactDetails     contactDetails.ContactDetails `json:"contact_details"`
	HasContactDetails  bool                          `json:"has_contact_details"`
}

// DatasetLandingPage contains properties related to the census dataset landing page
type DatasetLandingPage struct {
	HasOtherVersions       bool                    `json:"has_other_versions"`
	ShowOtherVersionsPanel bool                    `json:"show_other_versions_panel"`
	HasDownloads           bool                    `json:"has_downloads"`
	LatestVersionURL       string                  `json:"latest_version_url"`
	Dimensions             []sharedModel.Dimension `json:"dimensions"`
	ShareDetails           ShareDetails
	Methodologies          []Methodology `json:"methodology"`
	Description            []string      `json:"description"`
	IsFlexible             bool          `json:"is_flexible"`
	FormAction             string        `json:"form_action"`
	DatasetURL             string        `json:"dataset_url"`
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

// Methodology links
type Methodology struct {
	Description string `json:"description"`
	URL         string `json:"href"`
	Title       string `json:"title"`
}
