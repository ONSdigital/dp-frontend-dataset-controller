package model

// Correction represents a single correction on a version
type Correction struct {
	Reason string `json:"reason"`
	Date   string `json:"date"`
}
