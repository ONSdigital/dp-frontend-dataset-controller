package mapper

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/census"
	"github.com/ONSdigital/dp-renderer/v2/helper"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
)

// CreateCensusLandingPage creates a census-landing page based on api model responses
func CreateCensusLandingPage(req *http.Request, basePage coreModel.Page, d dataset.DatasetDetails, version dataset.Version, opts []dataset.Options, categorisationsMap map[string]int, initialVersionReleaseDate string, hasOtherVersions bool, allVersions []dataset.Version, latestVersionNumber int, latestVersionURL, lang string, queryStrValues []string, isValidationError bool, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner, isEnableMultivariate bool, pop population.GetPopulationTypeResponse) census.Page {
	p := CreateCensusBasePage(req, basePage, d, version, initialVersionReleaseDate, hasOtherVersions, allVersions, latestVersionNumber, latestVersionURL, lang, isValidationError, serviceMessage, emergencyBannerContent, isEnableMultivariate)

	// DOWNLOADS
	for ext, download := range version.Downloads {
		p.Version.Downloads = append(p.Version.Downloads, sharedModel.Download{
			Extension: strings.ToLower(ext),
			Size:      download.Size,
			URI:       download.URL,
		})
	}
	p.Version.Downloads = orderDownloads(p.Version.Downloads)

	if len(version.Downloads) >= 3 {
		p.DatasetLandingPage.HasDownloads = true
	}

	// DIMENSIONS
	if len(opts) > 0 {
		area, dims, qs := mapCensusOptionsToDimensions(version.Dimensions, opts, categorisationsMap, queryStrValues, req.URL.Path, lang, p.DatasetLandingPage.IsMultivariate)
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
		CollapsibleItems: mapLandingCollapsible(&version.Dimensions),
	}

	// ANALYTICS
	p.PreGTMJavaScript = append(p.PreGTMJavaScript, getDataLayerJavaScript(getAnalytics(p.DatasetLandingPage.Dimensions)))

	// FINAL FORMATTING
	p.DatasetLandingPage.QualityStatements = formatPanels(p.DatasetLandingPage.QualityStatements)

	return p
}

// mapCensusOptionsToDimensions links dimension options to dimensions and prepares them for display
func mapCensusOptionsToDimensions(dims []dataset.VersionDimension, opts []dataset.Options, categorisationsMap map[string]int, queryStrValues []string, path, lang string, isMultivariate bool) (area sharedModel.Dimension, dimensions []sharedModel.Dimension, qs []census.Panel) {
	for _, opt := range opts {
		var pDim sharedModel.Dimension

		for _, dimension := range dims {
			if dimension.Name != opt.Items[0].DimensionID {
				continue
			}
			pDim.Name = dimension.Name
			pDim.Description = dimension.Description
			pDim.IsAreaType = helpers.IsBoolPtr(dimension.IsAreaType)

			categorisationCount := categorisationsMap[dimension.ID]
			pDim.ShowChange = pDim.IsAreaType || (isMultivariate && categorisationCount > 1)

			pDim.Title = cleanDimensionLabel(dimension.Label)
			pDim.ID = dimension.ID
			if dimension.QualityStatementText != "" && dimension.QualityStatementURL != "" {
				qs = append(qs, census.Panel{
					Body:       []string{fmt.Sprintf("<p>%s</p>%s", dimension.QualityStatementText, helper.Localise("QualityNoticeReadMore", lang, 1, dimension.QualityStatementURL))},
					CSSClasses: []string{"ons-u-mt-no"},
				})
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

	sort.Slice(dimensions, func(i, j int) bool {
		return dimensions[i].IsAreaType
	})

	return dimensions[0], dimensions[1:], qs
}

// getAnalytics returns a map to add to the data layer which will be used on file download
func getAnalytics(dimensions []sharedModel.Dimension) map[string]string {
	analytics := make(map[string]string, 5)
	var dimensionIDs []string
	for _, dimension := range dimensions {
		if dimension.IsAreaType {
			analytics["areaType"] = dimension.ID
			analytics["coverageCount"] = "0"
		} else if !dimension.IsCoverage {
			dimensionIDs = append(dimensionIDs, dimension.ID)
		}
	}
	analytics["dimensions"] = strings.Join(dimensionIDs, ",")

	return analytics
}
