package datasetEditionsList

import (
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/contactDetails"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetLandingPageFilterable"
	"github.com/ONSdigital/dp-renderer/model"
)

type Page struct {
	model.Page
	datasetLandingPageFilterable.DatasetLandingPage
	ContactDetails contactDetails.ContactDetails `json:"contact_details"`
	Editions       []Edition                     `json:"editions"`
	DatasetTitle   string                        `json:"dataset_title"`
}

type Edition struct {
	Title            string `json:"title"`
	LatestVersionURL string `json:"url"`
}
