package createCustomDatasetPage

import (
	"github.com/ONSdigital/dp-renderer/v2/model"
)

// Page contains data for the census landing page
type Page struct {
	model.Page
	CreateCustomDatasetPage CreateCustomDatasetPage `json:"data"`
	IsNationalStatistic     bool                    `json:"is_national_statistic"`
	ShowCensusBranding      bool                    `json:"show_census_branding"`
}

// CreateDatasetPage contains properties related to the create dataset  page
type CreateCustomDatasetPage struct {
	PopulationTypes []PopulationType
}

// CreateDatasetPage contains properties related to the create dataset  page
type PopulationType struct {
	Name        string
	Label       string
	Description string
}
