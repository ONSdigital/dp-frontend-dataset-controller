package data

// Dataset ...
type Dataset struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	ReleaseDate string  `json:"release_date"`
	NextRelease string  `json:"next_release"`
	Edition     string  `json:"edition"`
	Version     string  `json:"version"`
	Contact     Contact `json:"contact"`
}

// Contact ...
type Contact struct {
	Name      string `json:"name"`
	Telephone string `json:"telephone"`
	Email     string `json:"email"`
}

// Dimension ...
type Dimension struct {
	CodeListID string   `json:"code_list_id"`
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Values     []string `json:"values"`
}
