package model

// Download has the details for the an individual dataset's downloadable files
type Download struct {
	Extension string `json:"extension"`
	Size      string `json:"size"`
	URI       string `json:"uri"`
	Name      string `json:"name"`
}
