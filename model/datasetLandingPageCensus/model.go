package datasetLandingPageCensus

import (
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/contactDetails"
	"github.com/ONSdigital/dp-renderer/model"
)

// Page contains data for the census landing page
type Page struct {
	model.Page
	DatasetLandingPage DatasetLandingPage            `json:"data"`
	Version            Version                       `json:"version"`
	InitialReleaseDate string                        `json:"initial_release_date"`
	ID                 string                        `json:"id"`
	ContactDetails     contactDetails.ContactDetails `json:"contact_details"`
	HasContactDetails  bool                          `json:"has_contact_details"`
}

// Version contains dataset version specific metadata
type Version struct {
	ReleaseDate string `json:"release_date"`
}

// DatasetLandingPage contains properties related to the census dataset landing page
type DatasetLandingPage struct {
	HasOtherVersions bool        `json:"has_other_versions"`
	Dimensions       []Dimension `json:"dimensions"`
	GuideContents    GuideContents
	Sections         map[string]Section
	ShareDetails     ShareDetails
	Methodologies    []Methodology `json:"methodology"`
}

// Dimension represents the data for a single dimension
type Dimension struct {
	Title       string   `json:"title"`
	Values      []string `json:"values"`
	OptionsURL  string   `json:"options_url"`
	TotalItems  int      `json:"total_items"`
	Description string   `json:"description"`
}

// GuideContents contains the contents of the page and the language attribute
type GuideContents struct {
	GuideContent []Content `json:"guide_content"`
	Language     string    `json:"language"`
}

/* Content maps the content details.
The visible text can be either a 'Title' or a 'LocaliseKey'.
The 'LocaliseKey' has to correspond to the localisation key found in the toml files within assets/locales, otherwise the page will error.
ID refers to the html element's ID that is needed to form the href.
*/
type Content struct {
	Title       string `json:"title"`
	ID          string `json:"id"`
	LocaliseKey string `json:"localise_key"`
}

// ShareDetails contains the locations the page can be shared to, as well as the language attribute for localisation
type ShareDetails struct {
	ShareLocations []Share `json:"share_locations"`
	Language       string  `json:"language"`
}

// Section corresponds to a section of the landing page with title, description and collapsible section (optional)
type Section struct {
	Title       string      `json:"title"`
	ID          string      `json:"id"`
	Description []string    `json:"description"`
	Collapsible Collapsible `json:"collapsible"`
}

/* Share includes details for a specific place the dataset can be shared
   Included icons: 'facebook', 'twitter', 'email', 'linkedin'
*/
type Share struct {
	Title string `json:"title"`
	Link  string `json:"link"`
	Icon  string `json:"icon"`
}

// Collapsible is a representation of the data required in a collapsible UI component
type Collapsible struct {
	Language string   `json:"language"`
	Title    string   `json:"title"`
	Content  []string `json:"content"`
}

// Methodology links
type Methodology struct {
	Description string `json:"description"`
	URL         string `json:"href"`
	Title       string `json:"title"`
}
