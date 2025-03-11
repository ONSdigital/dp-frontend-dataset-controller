package model

/* RelatedContentItem contains details for a section of related content
 */
type RelatedContentItem struct {
	Title string `json:"title"`
	Link  string `json:"link"`
	Text  string `json:"text"`
}
