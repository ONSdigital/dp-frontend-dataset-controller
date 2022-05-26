package model

// Dimension represents the data for a single dimension
type Dimension struct {
	Title       string   `json:"title"`
	Name        string   `json:"name"`
	Values      []string `json:"values"`
	OptionsURL  string   `json:"options_url"`
	TotalItems  int      `json:"total_items"`
	Description string   `json:"description"`
	IsAreaType  bool     `json:"is_area_type"`
}
