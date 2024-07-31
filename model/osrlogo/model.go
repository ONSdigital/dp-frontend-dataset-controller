package osrlogo

// OSRLogo stores the url, alt text and text for the OSR logo
type OSRLogo struct {
	URL     string `json:"url"`
	AltText string `json:"alt_text"`
	Title   string `json:"title"`
	About   string `json:"about"`
}
