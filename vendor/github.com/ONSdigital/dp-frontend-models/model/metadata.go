package model

//Metadata ...
type Metadata struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	ServiceName string   `json:"serviceName"`
	Keywords    []string `json:"keywords"`
}
