package zebedeeModels

type Dataset struct {
	Type               string      `json:"type"`
	URI                string      `json:"uri"`
	Description        Description `json:"description"`
	Downloads          []Download  `json:"downloads"`
	SupplementaryFiles []Download  `json:"supplementaryFiles"`
	Versions           []Version   `json:"versions"`
}

type Download struct {
	File string `json:"file"`
}

type Version struct {
	URI         string `json:"uri"`
	ReleaseDate string `json:"updateDate"`
	Notice      string `json:"correctionNotice"`
	Label       string `json:"label"`
}
