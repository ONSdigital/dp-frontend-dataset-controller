package related

// Related stores the Title and URI for any related data (eg related publications on a dataset page)
type Related struct {
	Title   string `json:"title"`
	Summary string `json:"summary,omitempty"`
	URI     string `json:"uri"`
}
