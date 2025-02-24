package mapper

import (
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/static"
)

// formatPanels is a helper function given an array of panels will format the final panel with the appropriate css class
func formatStaticPanels(panels []static.Panel) []static.Panel {
	if len(panels) > 0 {
		panelLen := len(panels)
		panels[panelLen-1].CSSClasses = append(panels[panelLen-1].CSSClasses, "ons-u-mb-l")
	}
	return panels
}
