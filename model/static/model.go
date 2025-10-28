package static

import (
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/contact"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/osrlogo"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/publisher"

	"github.com/ONSdigital/dp-renderer/v2/model"
)

// Page contains data for the census landing page
type Page struct {
	model.Page
	DatasetLandingPage  DatasetLandingPage    `json:"data"`
	Version             sharedModel.Version   `json:"version"`
	Versions            []sharedModel.Version `json:"versions"`
	ID                  string                `json:"id"`
	ContactDetails      contact.Details       `json:"contact_details"`
	HasContactDetails   bool                  `json:"has_contact_details"`
	IsNationalStatistic bool                  `json:"is_national_statistic"`
	ShowCensusBranding  bool                  `json:"show_census_branding"`
	Publisher           publisher.Publisher   `json:"publisher,omitempty"`
	UsageNotes          []UsageNote           `json:"usage_notes"`
	ShowApprove         bool                  `json:"show_approve"`
}

// StaticOverviewPage contains properties related to the static dataset
type DatasetLandingPage struct {
	DatasetURL          string                           `json:"dataset_url"`
	Description         []string                         `json:"description"`
	Dimensions          []sharedModel.Dimension          `json:"dimensions"`
	FeedbackAPIURL      string                           `json:"feedback_api_url"`
	HasDownloads        bool                             `json:"has_downloads"`
	HasOtherVersions    bool                             `json:"has_other_versions"`
	HasSDC              bool                             `json:"has_sdc"`
	ImproveResults      model.Collapsible                `json:"improve_results"`
	IsPublishing        bool                             `json:"is_publishing"`
	IsCustom            bool                             `json:"is_custom"`
	IsFlexibleForm      bool                             `json:"is_flexible_form"`
	IsMultivariate      bool                             `json:"is_multivariate"`
	LatestVersionURL    string                           `json:"latest_version_url"`
	NextRelease         string                           `json:"next_release"`
	OSRLogo             osrlogo.OSRLogo                  `json:"osr_logo"`
	Panels              []Panel                          `json:"panels"`
	QualityStatements   []Panel                          `json:"quality_statements"`
	RelatedContentItems []sharedModel.RelatedContentItem `json:"related_content_items"`
	SDC                 []Panel                          `json:"sdc"`
	ShareDetails        sharedModel.ShareDetails         `json:"share_details"`
	ShowXLSXInfo        bool                             `json:"show_xlsx_info"`
	State               string                           `json:"state"`
	Version             sharedModel.Version              `json:"version"`
}

// UsageNote represents data for a single usage note
type UsageNote struct {
	Note  string `json:"note,omitempty"`
	Title string `json:"title,omitempty"`
}
