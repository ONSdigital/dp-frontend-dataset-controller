package model

import "github.com/ONSdigital/dp-frontend-dataset-controller/model/contactDetails"

// Version represents the data for a single version
type Version struct {
	Title       string                        `json:"title"`
	Description string                        `json:"description"`
	URL         string                        `json:"url"`
	ReleaseDate string                        `json:"release_date"`
	NextRelease string                        `json:"next_release"`
	Downloads   []Download                    `json:"downloads"`
	Edition     string                        `json:"edition"`
	Version     string                        `json:"version"`
	Contact     contactDetails.ContactDetails `json:"contact"`
}
