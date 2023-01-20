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
func CreateCensusDatasetLandingPage(ctx context.Context, req *http.Request, basePage coreModel.Page, d dataset.DatasetDetails, version dataset.Version, opts []dataset.Options, initialVersionReleaseDate string, hasOtherVersions bool, allVersions []dataset.Version, latestVersionNumber int, latestVersionURL, lang string, queryStrValues []string, maxNumberOfOptions int, isValidationError, isFilterOutput, hasNoAreaOptions bool, filterOutput map[string]filter.Download, fDims []sharedModel.FilterDimension, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner, isEnableMultivariate bool) datasetLandingPageCensus.Page {
	if isFilterOutput {
		return CreateCensusFilterOutputsPage(ctx, req, basePage, d, version, initialVersionReleaseDate, hasOtherVersions, allVersions, latestVersionNumber, latestVersionURL, lang, queryStrValues, maxNumberOfOptions, isValidationError, hasNoAreaOptions, filterOutput, fDims, serviceMessage, emergencyBannerContent, isEnableMultivariate)
	} else {
		return CreateCensusLandingPage(ctx, req, basePage, d, version, opts, initialVersionReleaseDate, hasOtherVersions, allVersions, latestVersionNumber, latestVersionURL, lang, queryStrValues, maxNumberOfOptions, isValidationError, serviceMessage, emergencyBannerContent, isEnableMultivariate)
	}
}

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

// populateCollapsible maps dimension data for the collapsible section
func populateCollapsible(Dimensions []dataset.VersionDimension, isFilterOutput bool) []coreModel.CollapsibleItem {
	// TODO: This helper func will be re-written when filter output mapping work is done
	var collapsibleContentItems []coreModel.CollapsibleItem
	collapsibleContentItems = append(collapsibleContentItems, []coreModel.CollapsibleItem{
		{
			Subheading: AreaType,
			SafeHTML: coreModel.Localisation{
				LocaleKey: "VariableInfoAreaType",
				Plural:    1,
			},
		},
		{
			Subheading: Coverage,
			SafeHTML: coreModel.Localisation{
				LocaleKey: "VariableInfoCoverage",
				Plural:    1,
			},
		},
	}...)

	// TODO: Temporarily removing mapping on filter output pages until API is updated
	if !isFilterOutput {
		for _, dims := range Dimensions {
			if helpers.IsBoolPtr(dims.IsAreaType) && dims.Description != "" {
				collapsibleContentItems = append(collapsibleContentItems[:1], []coreModel.CollapsibleItem{
					{
						Subheading: cleanDimensionLabel(dims.Label),
						Content:    strings.Split(dims.Description, "\n"),
					},
					{
						Subheading: Coverage,
						SafeHTML: coreModel.Localisation{
							LocaleKey: "VariableInfoCoverage",
							Plural:    1,
						},
					},
				}...)
			}
			if !helpers.IsBoolPtr(dims.IsAreaType) && dims.Description != "" {
				var collapsibleContent coreModel.CollapsibleItem
				collapsibleContent.Subheading = cleanDimensionLabel(dims.Label)
				collapsibleContent.Content = strings.Split(dims.Description, "\n")
				collapsibleContentItems = append(collapsibleContentItems, collapsibleContent)
			}
		}
	}

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

// getDataLayerJavaScript returns a template.JS for page.PreGTMJavaScript that maps a map to the data layer
func getDataLayerJavaScript(analytics map[string]string) template.JS {
	jsonStr, _ := json.Marshal(analytics)
	return template.JS(`dataLayer.push(` + string(jsonStr) + `);`)
}
