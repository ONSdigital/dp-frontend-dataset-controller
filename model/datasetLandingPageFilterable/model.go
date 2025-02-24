package filterable

import (
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"

	"github.com/ONSdigital/dp-frontend-dataset-controller/model/contact"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/osrlogo"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/staticlegacy"
	"github.com/ONSdigital/dp-renderer/v2/model"
)

// Page contains data re-used for each page type a Data struct for data specific to the page type
type Page struct {
	model.Page
	DatasetLandingPage DatasetLandingPage `json:"data"`
	ContactDetails     contact.Details    `json:"contact_details"`
}

// DatasetLandingPage represents the data on the dataset landing page
type DatasetLandingPage struct {
	staticlegacy.DatasetLandingPage
	Dimensions               []sharedModel.Dimension `json:"dimensions"`
	Version                  sharedModel.Version     `json:"version"`
	HasOlderVersions         bool                    `json:"has_older_versions"`
	ShowEditionName          bool                    `json:"show_edition_name"`
	Edition                  string                  `json:"edition"`
	ReleaseFrequency         string                  `json:"release_frequency"`
	IsLatest                 bool                    `json:"is_latest"`
	LatestVersionURL         string                  `json:"latest_version_url"`
	IsLatestVersionOfEdition bool                    `json:"is_latest_version_of_edition_url"`
	QMIURL                   string                  `json:"qmi_url"`
	IsNationalStatistic      bool                    `json:"is_national_statistic"`
	Publications             []Publication           `json:"publications"`
	RelatedLinks             []Publication           `json:"related_links"`
	LatestChanges            []Change                `json:"latest_changes"`
	Citation                 string                  `json:"citation"`
	UnitOfMeasurement        string                  `json:"unit_of_measurement"`
	Methodologies            []Methodology           `json:"methodology"`
	NomisReferenceURL        string                  `json:"nomis_reference_url,omitempty"`
	UsageNotes               []UsageNote             `json:"UsageNotes"`
	OSRLogo                  osrlogo.OSRLogo         `json:"osr_logo"`
}

// UsageNote represents data for a single usage note
type UsageNote struct {
	Note  string `json:"note,omitempty"`
	Title string `json:"title,omitempty"`
}

// Publication represents the data for a single publication
type Publication struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// Change represents the data for a single change
type Change struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Publisher represents the data for a single publisher
type Publisher struct {
	Name string `json:"name"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

// Methodology links
type Methodology struct {
	Description string `json:"description"`
	URL         string `json:"href"`
	Title       string `json:"title"`
}
