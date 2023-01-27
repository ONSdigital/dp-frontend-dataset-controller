package mapper

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetLandingPageCensus"
	coreModel "github.com/ONSdigital/dp-renderer/model"
)

// Constants...
const (
	queryStrKey       = "showAll"
	Coverage          = "Coverage"
	AreaType          = "Area type"
	AnalyticsMaxItems = 4
)

// CreateCensusDatasetLandingPage creates a census-landing page based on api model responses
func CreateCensusDatasetLandingPage(isEnableMultivariate bool, ctx context.Context, req *http.Request, basePage coreModel.Page, d dataset.DatasetDetails, version dataset.Version, opts []dataset.Options, initialVersionReleaseDate string, hasOtherVersions bool, allVersions []dataset.Version, latestVersionNumber int, latestVersionURL, lang string, queryStrValues []string, maxNumberOfOptions int, isValidationError, isFilterOutput, hasNoAreaOptions bool, filterOutput map[string]filter.Download, fDims []sharedModel.FilterDimension, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) datasetLandingPageCensus.Page {
	if isFilterOutput {
		return CreateCensusFilterOutputsPage(isEnableMultivariate, ctx, req, basePage, d, version, opts, initialVersionReleaseDate, hasOtherVersions, allVersions, latestVersionNumber, latestVersionURL, lang, queryStrValues, maxNumberOfOptions, isValidationError, isFilterOutput, hasNoAreaOptions, filterOutput, fDims, serviceMessage, emergencyBannerContent, population.GetDimensionsResponse{})
	} else {
		return CreateCensusLandingPage(isEnableMultivariate, ctx, req, basePage, d, version, opts, initialVersionReleaseDate, hasOtherVersions, allVersions, latestVersionNumber, latestVersionURL, lang, queryStrValues, maxNumberOfOptions, isValidationError, isFilterOutput, hasNoAreaOptions, filterOutput, fDims, serviceMessage, emergencyBannerContent)
	}
}

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

	for _, dim := range dims {
		for _, dimDescription := range dimDescriptions.Dimensions {
			if dim.ID == dimDescription.ID && dim.IsAreaType {
				areaItem.Subheading = cleanDimensionLabel(dimDescription.Label)
				areaItem.Content = strings.Split(dimDescription.Description, "\n")
			} else if dim.ID == dimDescription.ID && !dim.IsAreaType {
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
func mapLandingCollapsible(Dimensions []dataset.VersionDimension) []coreModel.CollapsibleItem {
	var collapsibleContentItems []coreModel.CollapsibleItem
	var areaItem coreModel.CollapsibleItem
	for _, dim := range Dimensions {
		if helpers.IsBoolPtr(dim.IsAreaType) && dim.Description != "" {
			areaItem.Subheading = cleanDimensionLabel(dim.Label)
			areaItem.Content = strings.Split(dim.Description, "\n")
		} else if dim.Description != "" {
			collapsibleContentItems = append(collapsibleContentItems, coreModel.CollapsibleItem{
				Subheading: cleanDimensionLabel(dim.Label),
				Content:    strings.Split(dim.Description, "\n"),
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
func getTruncationMidRange(total int) (int, int) {
	mid := total / 2
	midFloor := mid - 2
	midCeiling := midFloor + 3
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
	matcher := regexp.MustCompile(`(\(\d+ ((C|c)ategories|(C|c)ategory)\))`)
	result := matcher.ReplaceAllString(label, "")
	return strings.TrimSpace(result)
}

func getDataLayerJavaScript(analytics map[string]string) template.JS {
	jsonStr, _ := json.Marshal(analytics)
	return template.JS(`dataLayer.push(` + string(jsonStr) + `);`)
}
