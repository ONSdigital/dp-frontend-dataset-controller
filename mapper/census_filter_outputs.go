package mapper

import (
	"context"
	"net/http"
	"net/url"
	"sort"
	"strconv"
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

const (
	FilterOutput = "_filter_output"
)

// CreateCensusFilterOutputsPage creates a filter output page based on api model responses
func CreateCensusFilterOutputsPage(ctx context.Context, req *http.Request, basePage coreModel.Page, d dataset.DatasetDetails, version dataset.Version, initialVersionReleaseDate string, hasOtherVersions bool, allVersions []dataset.Version, latestVersionNumber int, latestVersionURL, lang string, queryStrValues []string, maxNumberOfOptions int, isValidationError, hasNoAreaOptions bool, filterOutput map[string]filter.Download, fDims []sharedModel.FilterDimension, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner, isEnableMultivariate bool, dimDesc population.GetDimensionsResponse) datasetLandingPageCensus.Page {
	p := CreateCensusBasePage(ctx, req, basePage, d, version, initialVersionReleaseDate, hasOtherVersions, allVersions, latestVersionNumber, latestVersionURL, lang, isValidationError, serviceMessage, emergencyBannerContent, isEnableMultivariate)

	p.Type += FilterOutput
	p.SearchNoIndexEnabled = true

	// DOWNLOADS
	for ext, download := range filterOutput {
		p.Version.Downloads = append(p.Version.Downloads, sharedModel.Download{
			Extension: strings.ToLower(ext),
			Size:      download.Size,
			URI:       download.URL,
		})
	}
	p.Version.Downloads = orderDownloads(p.Version.Downloads)

	if len(filterOutput) >= 3 {
		p.DatasetLandingPage.HasDownloads = true
		p.DatasetLandingPage.ShowXLSXInfo = true
	}

	// DIMENSIONS
	p.DatasetLandingPage.Dimensions = mapFilterOutputDims(fDims, queryStrValues, req.URL.Path, p.DatasetLandingPage.IsMultivariate)
	coverage := []sharedModel.Dimension{
		{
			IsCoverage:        true,
			IsDefaultCoverage: hasNoAreaOptions,
			Title:             Coverage,
			Name:              strings.ToLower(Coverage),
			ID:                strings.ToLower(Coverage),
			Values:            fDims[0].Options,
			ShowChange:        true,
		},
	}
	temp := append(coverage, p.DatasetLandingPage.Dimensions[1:]...)
	p.DatasetLandingPage.Dimensions = append(p.DatasetLandingPage.Dimensions[:1], temp...)

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

	return p
}

// mapFilterOutputDims links dimension options to FilterDimensions and prepares them for display
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
		pDim.ShowChange = isAreaType || (isMultivariate && dim.CategorisationCount > 1)
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

// getFilterAnalytics returns a map to add to the data layer which will be used on file download
func getFilterAnalytics(filterDimensions []sharedModel.FilterDimension, defaultCoverage bool) map[string]string {
	analytics := make(map[string]string, 5)
	var dimensionIDs []string
	for _, filterDimension := range filterDimensions {
		dimension := filterDimension.ModelDimension
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
