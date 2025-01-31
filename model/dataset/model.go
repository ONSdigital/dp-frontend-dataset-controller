package dataset

import (
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/contact"
	"github.com/ONSdigital/dp-renderer/v2/model"
)

// Page contains data re-used for each page type a Data struct for data specific to the page type
type Page struct {
	model.Page
	DatasetPage DatasetPage `json:"data"`
	contact.Details
}

// DatasetPage has the file and title information for an individual dataset
type DatasetPage struct { //nolint:revive //renaming to page would mix types and have unintended consequences
	Versions            []Version           `json:"versions"`
	SupplementaryFiles  []SupplementaryFile `json:"supplementary_files"`
	Downloads           []Download          `json:"downloads"`
	IsNationalStatistic bool                `json:"national_statistic"`
	ReleaseDate         string              `json:"release_date"`
	NextRelease         string              `json:"next_release"`
	DatasetID           string              `json:"dataset_id"`
	URI                 string              `json:"uri"`
	Edition             string              `json:"edition"`
	Markdown            string              `json:"markdown"`
	ParentPath          string              `json:"parent_path"`
	EnableFeedbackAPI   bool                `json:"enable_feedback_api"`
	FeedbackAPIURL      string              `json:"feedback_api_url"`
}

// Download has the details for an individual dataset's downloadable files
type Download struct {
	Extension   string `json:"extension"`
	Size        string `json:"size"`
	URI         string `json:"uri"`
	File        string `json:"file"`
	DownloadURL string `json:"download_url,omitempty"`
}

// SupplementaryFile is a downloadable file that is associated to an individual dataset
type SupplementaryFile struct {
	Title       string `json:"title"`
	Extension   string `json:"extension"`
	Size        string `json:"size"`
	URI         string `json:"uri"`
	DownloadURL string `json:"download_url,omitempty"`
}

// Version has the details for a previous version of the dataset
type Version struct {
	URI              string     `json:"url"`
	UpdateDate       string     `json:"update_date"`
	CorrectionNotice string     `json:"correction_notice"`
	Label            string     `json:"label"`
	Downloads        []Download `json:"downloads"`
}
