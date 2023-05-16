package model

import "github.com/ONSdigital/dp-frontend-dataset-controller/model/contact"

// Version represents the data for a single version
type Version struct {
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	URL           string          `json:"url"`
	ReleaseDate   string          `json:"release_date"`
	NextRelease   string          `json:"next_release"`
	Downloads     []Download      `json:"downloads"`
	Edition       string          `json:"edition"`
	Version       string          `json:"version"`
	Contact       contact.Details `json:"contact"`
	IsCurrentPage bool            `json:"is_current"`
	VersionURL    string          `json:"version_url"`
	Superseded    string          `json:"superseded"`
	VersionNumber int             `json:"version_number"`
	Date          string          `json:"date"`
	Corrections   []Correction    `json:"correction"`
	FilterURL     string          `json:"filter_url"`
	IsLatest      bool            `json:"is_latest"`
}
