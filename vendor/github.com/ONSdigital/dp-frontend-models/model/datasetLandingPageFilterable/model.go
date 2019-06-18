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
	Dimensions               []Dimension   `json:"dimensions"`
	Version                  Version       `json:"version"`
	HasOlderVersions         bool          `json:"has_older_versions"`
	ShowEditionName          bool          `json:"show_edition_name"`
	Edition                  string        `json:"edition"`
	ReleaseFrequency         string        `json:"release_frequency"`
	IsLatest                 bool          `json:"is_latest"`
	LatestVersionURL         string        `json:"latest_version_url"`
	IsLatestVersionOfEdition bool          `json:"is_latest_version_of_edition_url"`
	QMIURL                   string        `json:"qmi_url"`
	IsNationalStatistic      bool          `json:"is_national_statistic"`
	Publications             []Publication `json:"publications"`
	RelatedLinks             []Publication `json:"related_links"`
	LatestChanges            []Change      `json:"latest_changes"`
	Citation                 string        `json:"citation"`
	DatasetTitle             string        `json:"dataset_title"`
	UnitOfMeasurement        string        `json:"unit_of_measurement"`
	Methodologies            []Methodology `json:"methodology"`
}

type Publication struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type Change struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Publisher struct {
	Name string `json:"name"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

type Dimension struct {
	Title       string   `json:"title"`
	Values      []string `json:"values"`
	OptionsURL  string   `json:"options_url"`
	TotalItems  int      `json:"total_items"`
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

// Methodology links
type Methodology struct {
	Description string `json:"description"`
	URL         string `json:"href"`
	Title       string `json:"title"`
}
