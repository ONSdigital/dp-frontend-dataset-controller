package mapper

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/census"
	"github.com/ONSdigital/dp-renderer/v2/helper"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
)

// CreateCensusLandingPage creates a census-landing page based on api model responses
func CreateCensusLandingPage(basePage coreModel.Page, d dpDatasetApiModels.Dataset, version dpDatasetApiModels.Version,
	opts []dpDatasetApiSdk.VersionDimensionOptionsList, categorisationsMap map[string]int, allVersions []dpDatasetApiModels.Version, queryStrValues []string,
	isEnableMultivariate bool, pop population.GetPopulationTypeResponse,
) census.Page {
	p := CreateCensusBasePage(basePage, d, version, allVersions, isEnableMultivariate)

	// DOWNLOADS
	if version.Downloads != nil {
		helpers.MapVersionDownloads(&p.Version, version.Downloads)
		p.Version.Downloads = orderDownloads(p.Version.Downloads)

		if len(p.Version.Downloads) >= 3 {
			p.DatasetLandingPage.HasDownloads = true
		}
	}

	// DIMENSIONS
	if len(opts) > 0 {
		area, dims, qs := mapCensusOptionsToDimensions(version.Dimensions, opts, categorisationsMap, queryStrValues, basePage.URI, basePage.Language, p.DatasetLandingPage.IsMultivariate)
		p.DatasetLandingPage.QualityStatements = qs
		sort.Slice(dims, func(i, j int) bool {
			return dims[i].Name < dims[j].Name
		})

		pop := sharedModel.Dimension{
			Title:            pop.PopulationType.Label,
			IsPopulationType: true,
		}
		coverage := sharedModel.Dimension{
			IsCoverage:        true,
			IsDefaultCoverage: true,
			Title:             Coverage,
			Name:              strings.ToLower(Coverage),
			ShowChange:        true,
			ID:                strings.ToLower(Coverage),
		}
		p.DatasetLandingPage.Dimensions = append([]sharedModel.Dimension{pop, area, coverage}, dims...)
	}

	// COLLAPSIBLE
	p.Collapsible = coreModel.Collapsible{
		Title: coreModel.Localisation{
			LocaleKey: "VariablesExplanation",
			Plural:    4,
		},
		CollapsibleItems: mapLandingCollapsible(version.Dimensions),
	}

	// ANALYTICS
	p.PreGTMJavaScript = append(p.PreGTMJavaScript, getDataLayerJavaScript(getAnalytics(p.DatasetLandingPage.Dimensions)))

	// FINAL FORMATTING
	p.DatasetLandingPage.QualityStatements = formatPanels(p.DatasetLandingPage.QualityStatements)

	return p
}

// mapCensusOptionsToDimensions links dimension options to dimensions and prepares them for display
func mapCensusOptionsToDimensions(dims []dpDatasetApiModels.Dimension, opts []dpDatasetApiSdk.VersionDimensionOptionsList, categorisationsMap map[string]int, queryStrValues []string, path, lang string, isMultivariate bool) (area sharedModel.Dimension, dimensions []sharedModel.Dimension, qs []census.Panel) {
	for _, opt := range opts {
		var pDim sharedModel.Dimension
		totalItems := len(opt.Items)

		for i := range dims {
			if dims[i].Name != opt.Items[0].Name {
				continue
			}
			pDim.Name = dims[i].Name
			pDim.Description = dims[i].Description
			pDim.IsAreaType = helpers.IsBoolPtr(dims[i].IsAreaType)

			categorisationCount := categorisationsMap[dims[i].ID]
			pDim.ShowChange = pDim.IsAreaType || (isMultivariate && categorisationCount > 1)

			pDim.Title = cleanDimensionLabel(dims[i].Label)
			pDim.ID = dims[i].ID
			if dims[i].QualityStatementText != "" && dims[i].QualityStatementURL != "" {
				qs = append(qs, census.Panel{
					Body:       []string{fmt.Sprintf("<p>%s</p>%s", dims[i].QualityStatementText, helper.Localise("QualityNoticeReadMore", lang, 1, dims[i].QualityStatementURL))},
					CSSClasses: []string{"ons-u-mt-no"},
				})
			}
		}

		pDim.TotalItems = totalItems
		midFloor, midCeiling := getTruncationMidRange(totalItems)

		var displayedOptions []dpDatasetApiModels.PublicDimensionOption
		if pDim.TotalItems > 9 && !helpers.HasStringInSlice(pDim.ID, queryStrValues) {
			displayedOptions = opt.Items[:3]
			displayedOptions = append(displayedOptions, opt.Items[midFloor:midCeiling]...)
			displayedOptions = append(displayedOptions, opt.Items[len(opt.Items)-3:]...)
			pDim.IsTruncated = true
		} else {
			displayedOptions = opt.Items
		}

		for i := range displayedOptions {
			pDim.Values = append(pDim.Values, displayedOptions[i].Label)
		}

		q := url.Values{}
		if pDim.IsTruncated {
			q.Add(queryStrKey, pDim.ID)
		}
		pDim.TruncateLink = generateTruncatePath(path, pDim.ID, q)
		dimensions = append(dimensions, pDim)
	}

	sort.Slice(dimensions, func(i, j int) bool {
		return dimensions[i].IsAreaType
	})

	return dimensions[0], dimensions[1:], qs
}

// getAnalytics returns a map to add to the data layer which will be used on file download
func getAnalytics(dimensions []sharedModel.Dimension) map[string]string {
	analytics := make(map[string]string, 5)
	var dimensionIDs []string
	for i := range dimensions {
		if dimensions[i].IsAreaType {
			analytics["areaType"] = dimensions[i].ID
			analytics["coverageCount"] = "0"
		} else if !dimensions[i].IsCoverage {
			dimensionIDs = append(dimensionIDs, dimensions[i].ID)
		}
	}
	analytics["dimensions"] = strings.Join(dimensionIDs, ",")

	return analytics
}
