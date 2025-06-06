package staticlegacy

import (
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/contact"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/osrlogo"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/related"
	"github.com/ONSdigital/dp-renderer/v2/model"
)

// Page contains data re-used for each page type a Data struct for data specific to the page type
type Page struct {
	model.Page
	DatasetLandingPage DatasetLandingPage `json:"data"`
	FilterID           string             `json:"filter_id"`
	contact.Details
}

// DatasetLandingPage represents a frontend dataset landing page
type DatasetLandingPage struct {
	DatasetID           string          `json:"dataset_id"`
	FilterID            string          `json:"filter_id"`
	Related             Related         `json:"related"`
	Datasets            []Dataset       `json:"datasets"`
	Notes               string          `json:"markdown"`
	MetaDescription     string          `json:"meta_description"`
	IsNationalStatistic bool            `json:"national_statistic"`
	Survey              string          `json:"survey"`
	ReleaseDate         string          `json:"release_date"`
	NextRelease         string          `json:"next_release"`
	IsTimeseries        bool            `json:"is_timeseries"`
	Corrections         []Message       `json:"corrections"`
	Notices             []Message       `json:"notices"`
	ParentPath          string          `json:"parent_path"`
	OSRLogo             osrlogo.OSRLogo `json:"osr_logo"`
	FeedbackAPIURL      string          `json:"feedback_api_url"`
}

// Related content (split by type) to this page
type Related struct {
	Publications       []related.Related `json:"related_publications"`
	FilterableDatasets []related.Related `json:"related_filterable_datasets"`
	Datasets           []related.Related `json:"related_datasets"`
	Methodology        []related.Related `json:"related_methodology"`
	Links              []related.Related `json:"related_links"`
}

// Dataset has the file and title information for an individual dataset
type Dataset struct {
	Title              string              `json:"title"`
	Downloads          []Download          `json:"downloads"`
	URI                string              `json:"uri"`
	HasVersions        bool                `json:"has_versions"`
	SupplementaryFiles []SupplementaryFile `json:"supplementary_files"`
	VersionLabel       string              `json:"version_label"`
	IsLast             bool                `json:"is_last"`
}

// Download has the details for the an individual dataset's downloadable files
type Download struct {
	Extension   string `json:"extension"`
	Size        string `json:"size"`
	URI         string `json:"uri"`
	DownloadURL string `json:"download_url"`
}

// SupplementaryFile is a downloadable file that is associated to an individual dataset
type SupplementaryFile struct {
	Title       string `json:"title"`
	Extension   string `json:"extension"`
	Size        string `json:"size"`
	URI         string `json:"uri"`
	DownloadURL string `json:"download_url"`
}

// Message has a date and time, used for either correction or notices
type Message struct {
	Date     string `json:"date"`
	Markdown string `json:"markdown"`
}
