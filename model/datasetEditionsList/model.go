package datasetEditionsList

import (
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/contactDetails"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetLandingPageFilterable"
	"github.com/ONSdigital/dp-renderer/v2/model"
)

// Page contains data re-used for each page type a Data struct for data specific to the page type
type Page struct {
	model.Page
	datasetLandingPageFilterable.DatasetLandingPage
	ContactDetails contactDetails.ContactDetails `json:"contact_details"`
	Editions       []Edition                     `json:"editions"`
	DatasetTitle   string                        `json:"dataset_title"`
}

// Edition contains data for a single edition
type Edition struct {
	Title            string `json:"title"`
	LatestVersionURL string `json:"url"`
}
