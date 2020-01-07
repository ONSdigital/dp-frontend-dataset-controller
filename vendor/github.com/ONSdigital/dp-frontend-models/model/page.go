package model

//Page contains data re-used for each page type a Data struct for data specific to the page type
type Page struct {
	Type                             string         `json:"type"`
	DatasetId                        string         `json:"dataset_id"`
	DatasetTitle                     string         `json:"dataset_title"`
	URI                              string         `json:"uri"`
	Taxonomy                         []TaxonomyNode `json:"taxonomy"`
	Breadcrumb                       []TaxonomyNode `json:"breadcrumb"`
	IsInFilterBreadcrumb             bool           `json:"is_in_filter_breadcrumb"`
	ServiceMessage                   string         `json:"service_message"`
	Metadata                         Metadata       `json:"metadata"`
	SearchDisabled                   bool           `json:"search_disabled"`
	SiteDomain                       string         `json:"-"`
	PatternLibraryAssetsPath         string         `json:"-"`
	Language                         string         `json:"language"`
	IncludeAssetsIntegrityAttributes bool           `json:"-"`
	ShowFeedbackForm                 bool           `json:"show_feedback_form"`
	ReleaseDate                      string         `json:"release_date"`
	BetaBannerEnabled                bool           `json:"beta_banner_enabled"`
	EnableLoop11                     bool           `json:"enable_loop11"`
}
