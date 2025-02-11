package publisher

// Publisher represents the data for a single publisher
type Publisher struct {
	URL  string `json:"href"`
	Name string `json:"name"`
	Type string `json:"type"`
}
