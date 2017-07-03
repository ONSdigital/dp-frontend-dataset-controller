package zebedeeModels

type TaxonomyNode struct {
	URI         string          `json:"uri"`
	Description NodeDescription `json:"description"`
	Type        string          `json:"type"`
}

type NodeDescription struct {
	Title string `json:"title"`
}
