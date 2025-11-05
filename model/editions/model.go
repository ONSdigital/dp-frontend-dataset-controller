package editions

import (
	"github.com/ONSdigital/dis-design-system-go/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/contact"
	filterable "github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetLandingPageFilterable"
)

// Page contains data re-used for each page type a Data struct for data specific to the page type
type Page struct {
	model.Page
	filterable.DatasetLandingPage
	ContactDetails contact.Details `json:"contact_details"`
	Editions       []List          `json:"editions"`
	ShowApprove    bool            `json:"show_approve"`
}

// List contains data for a single edition
type List struct {
	Title            string `json:"title"`
	LatestVersionURL string `json:"url"`
}
