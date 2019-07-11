package datasetVersionsList

import "github.com/ONSdigital/dp-frontend-models/model"

// Page contains the data re-used on each page as well as the data for the current page
type Page struct {
	model.Page
	Data VersionsList `json:"data"`
}

// VersionsList represents the data on the versions list page
type VersionsList struct {
	LatestVersionURL string    `json:"latest_version_url"`
	Versions         []Version `json:"versions"`
}

// Version represents an edition version on the version list page
type Version struct {
	Date          string       `json:"date"`
	Corrections   []Correction `json:"correction"`
	Downloads     []Download   `json:"downloads"`
	FilterURL     string       `json:"filter_url"`
	VersionURL    string       `json:"version_url"`
	VersionNumber int          `json:"version_number"`
	Superseded    string       `json:"superseded"`
	IsLatest      bool         `json:"is_latest"`
	Title         string       `json:"title"`
	Edition       string       `json:"edition"`
}

// Correction represents a single correction on the versions list page
type Correction struct {
	Reason string `json:"reason"`
	Date   string `json:"date"`
}

//Download has the details for the an individual dataset's downloadable files
type Download struct {
	Extension string `json:"extension"`
	Size      string `json:"size"`
	URI       string `json:"uri"`
}
