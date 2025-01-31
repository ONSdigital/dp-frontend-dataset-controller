package mapper

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/census"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/contact"
	"github.com/ONSdigital/dp-renderer/v2/helper"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
)

const (
	FilterOutput = "_filter_output"
)

// CreateCensusFilterOutputsPage creates a filter output page based on api model responses
func CreateCensusFilterOutputsPage(req *http.Request, basePage coreModel.Page, d dataset.DatasetDetails, version dataset.Version, initialVersionReleaseDate string, hasOtherVersions bool, allVersions []dataset.Version, latestVersionNumber int, latestVersionURL, lang string, queryStrValues []string, isValidationError, hasNoAreaOptions bool, filterOutput filter.Model, fDims []sharedModel.FilterDimension, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner, isEnableMultivariate bool, dimDesc population.GetDimensionsResponse, sdc cantabular.GetBlockedAreaCountResult, pop population.GetPopulationTypeResponse) census.Page {
	p := CreateCensusBasePage(req, basePage, d, version, initialVersionReleaseDate, hasOtherVersions, allVersions, latestVersionNumber, latestVersionURL, lang, isValidationError, serviceMessage, emergencyBannerContent, isEnableMultivariate)

	p.Type += FilterOutput
	p.SearchNoIndexEnabled = true

	// CUSTOM PAGE METADATA OVERRIDES
	if helpers.IsBoolPtr(filterOutput.Custom) {
		p.ReleaseDate = ""
		p.DatasetId = ""
		p.IsNationalStatistic = true
		p.ShowCensusBranding = true
		p.DatasetLandingPage.IsCustom = true
		p.HasContactDetails = true
		p.ContactDetails = contact.Details{
			Email:     "census.customerservices@ons.gov.uk",
			Telephone: "+44 1329 444972",
		}
		p.TableOfContents = buildTableOfContents(p, dataset.DatasetDetails{}, false)
	}

	// DOWNLOADS
	for ext, download := range filterOutput.Downloads {
		p.Version.Downloads = append(p.Version.Downloads, sharedModel.Download{
			Extension: strings.ToLower(ext),
			Size:      download.Size,
			URI:       download.URL,
		})
	}
	p.Version.Downloads = orderDownloads(p.Version.Downloads)

	if len(filterOutput.Downloads) >= 3 {
		p.DatasetLandingPage.HasDownloads = true
		p.DatasetLandingPage.ShowXLSXInfo = true
	}

	popDim := sharedModel.Dimension{
		IsPopulationType: true,
		Title:            pop.PopulationType.Label,
	}

	// DIMENSIONS
	p.DatasetLandingPage.Dimensions, p.DatasetLandingPage.QualityStatements = mapFilterOutputDims(fDims, queryStrValues, req.URL.Path, lang, p.DatasetLandingPage.IsMultivariate)
	coverage := sharedModel.Dimension{
		IsCoverage:        true,
		IsDefaultCoverage: hasNoAreaOptions,
		Title:             Coverage,
		Name:              strings.ToLower(Coverage),
		ID:                strings.ToLower(Coverage),
		Values:            fDims[0].Options,
		ShowChange:        true,
	}
	area := p.DatasetLandingPage.Dimensions[0]
	displayedDims := p.DatasetLandingPage.Dimensions[1:]
	sort.Slice(displayedDims, func(i, j int) bool {
		return displayedDims[i].Title < displayedDims[j].Title
	})
	p.DatasetLandingPage.Dimensions = append([]sharedModel.Dimension{popDim, area, coverage}, displayedDims...)

	// CUSTOM TITLE
	if helpers.IsBoolPtr(filterOutput.Custom) || p.DatasetLandingPage.IsMultivariate {
		nonGeoDims := getNonGeographyDims(p.DatasetLandingPage.Dimensions)
		dimensionStr := buildConjoinedList(nonGeoDims, true)
		vDims := getNonGeographyVersionDims(version.Dimensions)
		vTitle := buildConjoinedList(vDims, true)
		if vTitle != dimensionStr || helpers.IsBoolPtr(filterOutput.Custom) {
			p.Metadata.Title = strings.ToUpper(dimensionStr[:1]) + strings.ToLower(dimensionStr[1:])
			p.DatasetLandingPage.Description = []string{
				helper.Localise("CustomDatasetSummary", lang, 1, strings.ToLower(popDim.Title), strings.ToLower(dimensionStr)),
			}
		}
	}

	// COLLAPSIBLE CONTENT
	p.Collapsible = coreModel.Collapsible{
		Title: coreModel.Localisation{
			LocaleKey: "VariablesExplanation",
			Plural:    4,
		},
		CollapsibleItems: mapOutputCollapsible(dimDesc, p.DatasetLandingPage.Dimensions),
	}

	// ANALYTICS
	p.PreGTMJavaScript = append(p.PreGTMJavaScript, getDataLayerJavaScript(getFilterAnalytics(fDims, hasNoAreaOptions)))

	// SDC
	if p.DatasetLandingPage.IsMultivariate {
		switch {
		case sdc.Blocked > 0: // areas blocked
			p.DatasetLandingPage.HasSDC = true
			p.DatasetLandingPage.SDC = mapBlockedAreasPanel(&sdc, census.Pending, lang)
			p.DatasetLandingPage.ImproveResults = mapImproveResultsCollapsible(p.DatasetLandingPage.Dimensions, lang)
		case sdc.Passed == sdc.Total && sdc.Total > 0: // all areas passing
			p.DatasetLandingPage.HasSDC = true
			p.DatasetLandingPage.SDC = mapBlockedAreasPanel(&sdc, census.Success, lang)
		}
	}

	// FINAL FORMATTING
	p.DatasetLandingPage.QualityStatements = formatPanels(p.DatasetLandingPage.QualityStatements)

	// FEEDBACK API
	p.FeatureFlags.EnableFeedbackAPI = cfg.EnableFeedbackAPI
	p.FeatureFlags.FeedbackAPIURL = cfg.FeedbackAPIURL

	return p
}

// mapFilterOutputDims links dimension options to FilterDimensions and prepares them for display
func mapFilterOutputDims(dims []sharedModel.FilterDimension, queryStrValues []string, path, lang string, isMultivariate bool) (dimensions []sharedModel.Dimension, qs []census.Panel) {
	sort.Slice(dims, func(i, j int) bool {
		return *dims[i].IsAreaType
	})
	for i := range dims {
		var isAreaType bool
		if helpers.IsBoolPtr(dims[i].IsAreaType) {
			isAreaType = true
		}
		pDim := sharedModel.Dimension{}
		pDim.Title = cleanDimensionLabel(dims[i].Label)
		pDim.ID = dims[i].ID
		pDim.Name = dims[i].Name
		pDim.IsAreaType = isAreaType
		pDim.ShowChange = isAreaType || (isMultivariate && dims[i].CategorisationCount > 1)
		pDim.TotalItems = dims[i].OptionsCount
		midFloor, midCeiling := getTruncationMidRange(pDim.TotalItems)

		var displayedOptions []string
		if pDim.TotalItems > 9 && !helpers.HasStringInSlice(pDim.ID, queryStrValues) && !pDim.IsAreaType {
			displayedOptions = dims[i].Options[:3]
			displayedOptions = append(displayedOptions, dims[i].Options[midFloor:midCeiling]...)
			displayedOptions = append(displayedOptions, dims[i].Options[len(dims[i].Options)-3:]...)
			pDim.IsTruncated = true
		} else {
			displayedOptions = dims[i].Options
		}

		pDim.Values = append(pDim.Values, displayedOptions...)

		q := url.Values{}
		if pDim.IsTruncated {
			q.Add(queryStrKey, pDim.ID)
		}
		pDim.TruncateLink = generateTruncatePath(path, pDim.ID, q)
		if dims[i].QualityStatementText != "" && dims[i].QualitySummaryURL != "" {
			qs = append(qs, census.Panel{
				Body:       []string{fmt.Sprintf("<p>%s</p>%s", dims[i].QualityStatementText, helper.Localise("QualityNoticeReadMore", lang, 1, dims[i].QualitySummaryURL))},
				CSSClasses: []string{"ons-u-mt-no"},
			})
		}
		dimensions = append(dimensions, pDim)
	}
	return dimensions, qs
}

// getFilterAnalytics returns a map to add to the data layer which will be used on file download
func getFilterAnalytics(filterDimensions []sharedModel.FilterDimension, defaultCoverage bool) map[string]string {
	analytics := make(map[string]string, 5)
	var dimensionIDs []string
	for i := range filterDimensions {
		dimension := filterDimensions[i].ModelDimension
		if dimension.IsAreaType != nil && *dimension.IsAreaType {
			analytics["areaType"] = dimension.ID

			if defaultCoverage {
				analytics["coverageCount"] = "0"
			} else {
				analytics["coverageCount"] = strconv.Itoa(len(dimension.Options))

				if len(dimension.Options) > 0 {
					if len(dimension.Options) <= AnalyticsMaxItems {
						analytics["coverage"] = strings.Join(dimension.Options, ",")
					}
					if dimension.FilterByParent == "" {
						analytics["coverageAreaType"] = dimension.ID
					} else {
						analytics["coverageAreaType"] = dimension.FilterByParent
					}
				}
			}
		} else {
			dimensionIDs = append(dimensionIDs, dimension.ID)
		}
	}
	analytics["dimensions"] = strings.Join(dimensionIDs, ",")

	return analytics
}

// mapBlockedAreasPanel is a helper function that maps the blocked areas panel by panel type
func mapBlockedAreasPanel(sdc *cantabular.GetBlockedAreaCountResult, panelType census.PanelType, lang string) (p []census.Panel) {
	switch panelType {
	case census.Pending:
		p = []census.Panel{
			{
				Type:        census.Pending,
				DisplayIcon: false,
				CSSClasses:  []string{"ons-u-mt-xl", "ons-u-mb-s"},
				Body: []string{
					helper.Localise("SDCAreasAvailable", lang, 1, helper.ThousandsSeparator(sdc.Passed), helper.ThousandsSeparator(sdc.Total)),
					helper.Localise("SDCRestrictedAreas", lang, sdc.Blocked, helper.ThousandsSeparator(sdc.Blocked)),
				},
				Language: lang,
			},
		}
	case census.Success:
		p = []census.Panel{
			{
				Type:        census.Success,
				DisplayIcon: false,
				CSSClasses:  []string{"ons-u-mt-xl", "ons-u-mb-s"},
				Body: []string{
					helper.Localise("SDCAllAreasAvailable", lang, sdc.Total, helper.ThousandsSeparator(sdc.Total)),
				},
				Language: lang,
			},
		}
	}
	return p
}

func mapImproveResultsCollapsible(dims []sharedModel.Dimension, lang string) coreModel.Collapsible {
	dimsList := getNonGeographyDims(dims)
	stringList := buildConjoinedList(dimsList, false)

	return coreModel.Collapsible{
		Title: coreModel.Localisation{
			LocaleKey: "ImproveResultsTitle",
			Plural:    4,
		},
		CollapsibleItems: []coreModel.CollapsibleItem{
			{
				Subheading: helper.Localise("ImproveResultsSubHeading", lang, 1),
				SafeHTML: coreModel.Localisation{
					Text: helper.Localise("ImproveResultsList", lang, 1, stringList),
				},
			},
		},
	}
}

// getNonGeographyDims returns all dimensions that are non-geography (not area type, coverage or population type)
func getNonGeographyDims(dims []sharedModel.Dimension) (dimsList []string) {
	for i := range dims {
		if !dims[i].IsAreaType && !dims[i].IsCoverage && !dims[i].IsPopulationType {
			dimsList = append(dimsList, dims[i].Title)
		}
	}
	return dimsList
}

// getNonGeographyVersionDims returns all version dimensions that is not an area type
func getNonGeographyVersionDims(dims []dataset.VersionDimension) (dimsList []string) {
	for i := range dims {
		if !helpers.IsBoolPtr(dims[i].IsAreaType) {
			dimsList = append(dimsList, cleanDimensionLabel(dims[i].Label))
		}
	}
	sort.Strings(dimsList)
	return dimsList
}

// buildConjoinedList returns a single string from an array that is conjoined with a comma, 'or', 'and'
func buildConjoinedList(dimsList []string, useAnd bool) (str string) {
	var penultimateItem = len(dimsList) - 2
	for i, item := range dimsList {
		switch {
		case i < penultimateItem:
			str += fmt.Sprintf("%s, ", item)
		case i == penultimateItem:
			if useAnd {
				str += fmt.Sprintf("%s and ", item)
			} else {
				str += fmt.Sprintf("%s or ", item)
			}
		default:
			str += item
		}
	}
	return str
}
