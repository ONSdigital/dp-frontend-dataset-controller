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
	Methodologies []Methodology `json:"methodology"`
}

// Section corresponds to a section of the landing page with title, description and collapsible section (optional)
type Section struct {
	Title       string      `json:"title"`
	Description []string    `json:"description"`
	Collapsible Collapsible `json:"collapsible"`
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
