package census

import (
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/contact"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/osrlogo"

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
}

// DatasetLandingPage contains properties related to the census dataset landing page
type DatasetLandingPage struct {
	HasOtherVersions    bool                             `json:"has_other_versions"`
	HasDownloads        bool                             `json:"has_downloads"`
	LatestVersionURL    string                           `json:"latest_version_url"`
	Dimensions          []sharedModel.Dimension          `json:"dimensions"`
	ShareDetails        sharedModel.ShareDetails         `json:"share_details"`
	Description         []string                         `json:"description"`
	IsCustom            bool                             `json:"is_custom"`
	IsFlexibleForm      bool                             `json:"is_flexible_form"`
	DatasetURL          string                           `json:"dataset_url"`
	Panels              []Panel                          `json:"panels"`
	QualityStatements   []Panel                          `json:"quality_statements"`
	SDC                 []Panel                          `json:"sdc"`
	HasSDC              bool                             `json:"has_sdc"`
	RelatedContentItems []sharedModel.RelatedContentItem `json:"related_content_items"`
	IsMultivariate      bool                             `json:"is_multivariate"`
	ShowXLSXInfo        bool                             `json:"show_xlsx_info"`
	OSRLogo             osrlogo.OSRLogo                  `json:"osr_logo"`
	FeedbackAPIURL      string                           `json:"feedback_api_url"`
	ImproveResults      model.Collapsible                `json:"improve_results"`
}
