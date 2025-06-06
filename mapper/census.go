package mapper

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"regexp"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/census"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
)

// Constants...
const (
	queryStrKey       = "showAll"
	Coverage          = "Coverage"
	AreaType          = "Area type"
	AnalyticsMaxItems = 4
)

// orderDownloads orders a set of sharedModel.Downloads using a hardcoded download order
func orderDownloads(downloads []sharedModel.Download) []sharedModel.Download {
	downloadOrder := []string{"xls", "xlsx", "csv", "txt", "csvw"}
	mapped := make(map[string]sharedModel.Download, 5)
	for _, download := range downloads {
		mapped[download.Extension] = download
	}
	var ordered []sharedModel.Download
	for _, ext := range downloadOrder {
		if download, ok := mapped[ext]; ok {
			ordered = append(ordered, download)
		}
	}
	return ordered
}

func areaTypeItem() coreModel.CollapsibleItem {
	return coreModel.CollapsibleItem{
		Subheading: AreaType,
		SafeHTML: coreModel.Localisation{
			LocaleKey: "VariableInfoAreaType",
			Plural:    1,
		},
	}
}

func coverageItem() coreModel.CollapsibleItem {
	return coreModel.CollapsibleItem{
		Subheading: Coverage,
		SafeHTML: coreModel.Localisation{
			LocaleKey: "VariableInfoCoverage",
			Plural:    1,
		},
	}
}

// mapOutputCollapsible maps the collapsible on the output page
func mapOutputCollapsible(dimDescriptions population.GetDimensionsResponse, dims []sharedModel.Dimension) []coreModel.CollapsibleItem {
	var collapsibleContentItems []coreModel.CollapsibleItem
	var areaItem coreModel.CollapsibleItem

	for i := range dims {
		for _, dimDescription := range dimDescriptions.Dimensions {
			if dims[i].ID == dimDescription.ID && dims[i].IsAreaType {
				areaItem.Subheading = cleanDimensionLabel(dimDescription.Label)
				areaItem.Content = strings.Split(dimDescription.Description, "\n")
			} else if dims[i].ID == dimDescription.ID && !dims[i].IsAreaType {
				collapsibleContentItems = append(collapsibleContentItems, coreModel.CollapsibleItem{
					Subheading: cleanDimensionLabel(dimDescription.Label),
					Content:    strings.Split(dimDescription.Description, "\n"),
				})
			}
		}
	}

	return concatenateCollapsibleItems(collapsibleContentItems, areaItem)
}

// mapLandingCollapsible maps the collapsible on the landing page
func mapLandingCollapsible(dimensions []dpDatasetApiModels.Dimension) []coreModel.CollapsibleItem {
	var collapsibleContentItems []coreModel.CollapsibleItem
	var areaItem coreModel.CollapsibleItem
	for i := range dimensions {
		if helpers.IsBoolPtr(dimensions[i].IsAreaType) && dimensions[i].Description != "" {
			areaItem.Subheading = cleanDimensionLabel(dimensions[i].Label)
			areaItem.Content = strings.Split(dimensions[i].Description, "\n")
		} else if dimensions[i].Description != "" {
			collapsibleContentItems = append(collapsibleContentItems, coreModel.CollapsibleItem{
				Subheading: cleanDimensionLabel(dimensions[i].Label),
				Content:    strings.Split(dimensions[i].Description, "\n"),
			})
		}
	}

	return concatenateCollapsibleItems(collapsibleContentItems, areaItem)
}

// concatenateCollapsibleItems returns the collapsible in the order: area type, area type description, coverage then other dimensions
func concatenateCollapsibleItems(collapsibleContentItems []coreModel.CollapsibleItem, areaItem coreModel.CollapsibleItem) []coreModel.CollapsibleItem {
	collapsibleContentItems = append([]coreModel.CollapsibleItem{
		areaTypeItem(),
		areaItem,
		coverageItem(),
	}, collapsibleContentItems...)

	return collapsibleContentItems
}

// getTruncationMidRange returns ints that can be used as the truncation mid range
func getTruncationMidRange(total int) (midFloor, midCeiling int) {
	mid := total / 2
	midFloor = mid - 2
	midCeiling = midFloor + 3
	if midFloor < 0 {
		midFloor = 0
	}
	return midFloor, midCeiling
}

// generateTruncatePath returns the path to truncate or show all
func generateTruncatePath(path, dimID string, q url.Values) string {
	truncatePath := path
	if q.Encode() != "" {
		truncatePath += fmt.Sprintf("?%s", q.Encode())
	}
	if dimID != "" {
		truncatePath += fmt.Sprintf("#%s", dimID)
	}
	return truncatePath
}

// cleanDimensionLabel is a helper function that parses dimension labels from cantabular into display text
func cleanDimensionLabel(label string) string {
	matcher := regexp.MustCompile(`(\(\d+ (([Cc])ategories|([Cc])ategory)\))`)
	result := matcher.ReplaceAllString(label, "")
	return strings.TrimSpace(result)
}

// getDataLayerJavaScript returns a template.JS for page.PreGTMJavaScript that maps a map to the data layer
func getDataLayerJavaScript(analytics map[string]string) template.JS {
	jsonStr, _ := json.Marshal(analytics)
	//nolint:gosec //cannot html escape string as JS is required in output
	return template.JS(`dataLayer.push(` + string(jsonStr) + `);`)
}

// formatPanels is a helper function given an array of panels will format the final panel with the appropriate css class
func formatPanels(panels []census.Panel) []census.Panel {
	if len(panels) > 0 {
		panelLen := len(panels)
		panels[panelLen-1].CSSClasses = append(panels[panelLen-1].CSSClasses, "ons-u-mb-l")
	}
	return panels
}
