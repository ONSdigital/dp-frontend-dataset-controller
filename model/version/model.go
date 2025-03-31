package version

import (
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-renderer/v2/model"
)

// Page contains the data re-used on each page as well as the data for the current page
type Page struct {
	model.Page
	Data VersionsList `json:"data"`
}

// VersionsList represents the data on the versions list page
type VersionsList struct {
	LatestVersionURL string                `json:"latest_version_url"`
	Versions         []sharedModel.Version `json:"versions"`
	FeedbackAPIURL   string                `json:"feedback_api_url"`
}

// Download has the details for the an individual dataset's downloadable files
type Download struct {
	Extension string `json:"extension"`
	Size      string `json:"size"`
	URI       string `json:"uri"`
}
