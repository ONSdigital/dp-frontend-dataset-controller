package model

// Dimension represents the data for a single dimension
type Dimension struct {
	Title             string   `json:"title"`
	Name              string   `json:"name"`
	Values            []string `json:"values"`
	OptionsURL        string   `json:"options_url"`
	TotalItems        int      `json:"total_items"`
	Description       string   `json:"description"`
	IsAreaType        bool     `json:"is_area_type"`
	IsCoverage        bool     `json:"is_coverage"`
	IsDefaultCoverage bool     `json:"is_default_coverage"`
	ShowChange        bool     `json:"show_change"`
	ChangeURL         string   `json:"change_url"`
	IsTruncated       bool     `json:"is_truncated"`
	TruncateLink      string   `json:"truncate_link"`
	ID                string   `json:"id"`
}
