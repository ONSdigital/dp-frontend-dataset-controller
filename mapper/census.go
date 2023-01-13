package mapper

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetLandingPageCensus"
	"github.com/ONSdigital/dp-renderer/helper"
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
		return CreateCensusFilterOutputsPage(isEnableMultivariate, ctx, req, basePage, d, version, opts, initialVersionReleaseDate, hasOtherVersions, allVersions, latestVersionNumber, latestVersionURL, lang, queryStrValues, maxNumberOfOptions, isValidationError, isFilterOutput, hasNoAreaOptions, filterOutput, fDims, serviceMessage, emergencyBannerContent)
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

func mapCensusOptionsToDimensions(dims []dataset.VersionDimension, opts []dataset.Options, queryStrValues []string, path, lang string, isFlex, isMultivariate bool) ([]sharedModel.Dimension, []datasetLandingPageCensus.Panel) {
	dimensions := []sharedModel.Dimension{}
	qs := []datasetLandingPageCensus.Panel{}
	for _, opt := range opts {
		var pDim sharedModel.Dimension

		for _, dimension := range dims {
			if dimension.Name == opt.Items[0].DimensionID {
				pDim.Name = dimension.Name
				pDim.Description = dimension.Description
				pDim.IsAreaType = helpers.IsBoolPtr(dimension.IsAreaType)
				pDim.ShowChange = pDim.IsAreaType && isFlex || isMultivariate
				pDim.Title = cleanDimensionLabel(dimension.Label)
				pDim.ID = dimension.ID
				if dimension.QualityStatementText != "" && dimension.QualityStatementURL != "" {
					qs = append(qs, datasetLandingPageCensus.Panel{
						Body:       fmt.Sprintf("<p>%s</p>%s", dimension.QualityStatementText, helper.Localise("QualityNoticeReadMore", lang, 1, dimension.QualityStatementURL)),
						CssClasses: []string{"ons-u-mt-no"},
					})
				}
			}
		}

		pDim.TotalItems = opt.TotalCount
		midFloor, midCeiling := getTruncationMidRange(opt.TotalCount)

		var displayedOptions []dataset.Option
		if pDim.TotalItems > 9 && !helpers.HasStringInSlice(pDim.ID, queryStrValues) {
			displayedOptions = opt.Items[:3]
			displayedOptions = append(displayedOptions, opt.Items[midFloor:midCeiling]...)
			displayedOptions = append(displayedOptions, opt.Items[len(opt.Items)-3:]...)
			pDim.IsTruncated = true
		} else {
			displayedOptions = opt.Items
		}

		for _, opt := range displayedOptions {
			pDim.Values = append(pDim.Values, opt.Label)
		}

		q := url.Values{}
		if pDim.IsTruncated {
			q.Add(queryStrKey, pDim.ID)
		}
		pDim.TruncateLink = generateTruncatePath(path, pDim.ID, q)
		dimensions = append(dimensions, pDim)
	}
	return dimensions, qs
}

func mapFilterOutputDims(dims []sharedModel.FilterDimension, queryStrValues []string, path string, isMultivariate bool) []sharedModel.Dimension {
	sort.Slice(dims, func(i, j int) bool {
		return *dims[i].IsAreaType
	})
	dimensions := []sharedModel.Dimension{}
	for _, dim := range dims {
		var isAreaType bool
		if helpers.IsBoolPtr(dim.IsAreaType) {
			isAreaType = true
		}
		pDim := sharedModel.Dimension{}
		pDim.Title = cleanDimensionLabel(dim.Label)
		pDim.ID = dim.ID
		pDim.Name = dim.Name
		pDim.IsAreaType = isAreaType
		pDim.ShowChange = isAreaType || isMultivariate
		pDim.TotalItems = dim.OptionsCount
		midFloor, midCeiling := getTruncationMidRange(pDim.TotalItems)

		var displayedOptions []string
		if pDim.TotalItems > 9 && !helpers.HasStringInSlice(pDim.ID, queryStrValues) && !pDim.IsAreaType {
			displayedOptions = dim.Options[:3]
			displayedOptions = append(displayedOptions, dim.Options[midFloor:midCeiling]...)
			displayedOptions = append(displayedOptions, dim.Options[len(dim.Options)-3:]...)
			pDim.IsTruncated = true
		} else {
			displayedOptions = dim.Options
		}

		pDim.Values = append(pDim.Values, displayedOptions...)

		q := url.Values{}
		if pDim.IsTruncated {
			q.Add(queryStrKey, pDim.ID)
		}
		pDim.TruncateLink = generateTruncatePath(path, pDim.ID, q)
		dimensions = append(dimensions, pDim)
	}
	return dimensions
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
