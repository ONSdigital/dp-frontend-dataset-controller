package datasetEditionsList

import (
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/datasetLandingPageFilterable"
)

type Page struct {
	model.Page
	datasetLandingPageFilterable.DatasetLandingPage
	ContactDetails model.ContactDetails `json:"contact_details"`
	Editions       []Edition            `json:"editions"`
}

type Edition struct {
	Title            string `json:"title"`
	LatestVersionURL string `json:"url"`
}
