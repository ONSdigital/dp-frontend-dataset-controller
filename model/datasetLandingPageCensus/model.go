package datasetLandingPageCensus

import (
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/contactDetails"
	"github.com/ONSdigital/dp-renderer/model"
)

// Page contains data for the census landing page
type Page struct {
	model.Page
	DatasetLandingPage DatasetLandingPage            `json:"data"`
	ContactDetails     contactDetails.ContactDetails `json:"contact_details"`
}

// DatasetLandingPage contains properties related to the census dataset landing page
type DatasetLandingPage struct {
	Sections      map[string]Section
	ShareDetails  ShareDetails
	Methodologies []Methodology `json:"methodology"`
}

// ShareDetails contains the locations the page can be shared to, as well as the language attribute for localisation
type ShareDetails struct {
	ShareLocations []Share `json:"share_locations"`
	Language       string  `json:"language"`
}

// Section corresponds to a section of the landing page with title, description and collapsible section (optional)
type Section struct {
	Title       string      `json:"title"`
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
