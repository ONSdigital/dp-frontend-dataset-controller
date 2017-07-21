package datasetLandingPageFilterable

import (
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/datasetLandingPageStatic"
)

type Page struct {
	model.Page
	DatasetLandingPage DatasetLandingPage   `json:"data"`
	ContactDetails     model.ContactDetails `json:"contact_details"`
}

type DatasetLandingPage struct {
	datasetLandingPageStatic.DatasetLandingPage
	Dimensions []Dimension `json:"dimensions"`
	Versions   []Version   `json:"versions"`
}

type Dimension struct {
	Title       string   `json:"title"`
	Values      []string `json:"values"`
	Description string   `json:"description"`
}

type Version struct {
	Title       string               `json:"title"`
	Description string               `json:"description"`
	URL         string               `json:"url"`
	ReleaseDate string               `json:"release_date"`
	NextRelease string               `json:"next_release"`
	Downloads   []Download           `json:"downloads"`
	Edition     string               `json:"edition"`
	Version     string               `json:"version"`
	Contact     model.ContactDetails `json:"contact"`
}

//Download has the details for the an individual dataset's downloadable files
type Download struct {
	Extension string `json:"extension"`
	Size      string `json:"size"`
	URI       string `json:"uri"`
}
