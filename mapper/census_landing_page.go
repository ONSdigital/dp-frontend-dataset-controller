package mapper

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetLandingPageCensus"
	"github.com/ONSdigital/dp-renderer/helper"
	coreModel "github.com/ONSdigital/dp-renderer/model"
)

// CreateCensusLandingPage creates a census-landing page based on api model responses
func CreateCensusLandingPage(isEnableMultivariate bool, ctx context.Context, req *http.Request, basePage coreModel.Page, d dataset.DatasetDetails, version dataset.Version, opts []dataset.Options, initialVersionReleaseDate string, hasOtherVersions bool, allVersions []dataset.Version, latestVersionNumber int, latestVersionURL, lang string, queryStrValues []string, maxNumberOfOptions int, isValidationError, isFilterOutput, hasNoAreaOptions bool, filterOutput map[string]filter.Download, fDims []sharedModel.FilterDimension, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) datasetLandingPageCensus.Page {
	p := CreateCensusBasePage(isEnableMultivariate, ctx, req, basePage, d, version, initialVersionReleaseDate, hasOtherVersions, allVersions, latestVersionNumber, latestVersionURL, lang, isValidationError, serviceMessage, emergencyBannerContent)

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
		p.DatasetLandingPage.Dimensions, p.DatasetLandingPage.QualityStatements = mapCensusOptionsToDimensions(version.Dimensions, opts, queryStrValues, req.URL.Path, lang, true, p.DatasetLandingPage.IsMultivariate)
		coverage := []sharedModel.Dimension{
			{
				IsCoverage:        true,
				IsDefaultCoverage: true,
				Title:             Coverage,
				Name:              strings.ToLower(Coverage),
				ShowChange:        true,
				ID:                strings.ToLower(Coverage),
			},
		}
		temp := append(coverage, p.DatasetLandingPage.Dimensions[1:]...)
		p.DatasetLandingPage.Dimensions = append(p.DatasetLandingPage.Dimensions[:1], temp...)
	}

	// COLLAPSIBLE
	p.Collapsible = coreModel.Collapsible{
		Title: coreModel.Localisation{
			LocaleKey: "VariablesExplanation",
			Plural:    4,
		},
		CollapsibleItems: populateCollapsible(version.Dimensions, false),
	}

	// ANALYTICS
	p.PreGTMJavaScript = append(p.PreGTMJavaScript, getDataLayerJavaScript(getAnalytics(p.DatasetLandingPage.Dimensions)))

	// FINAL FORMATTING
	if len(p.DatasetLandingPage.QualityStatements) > 0 {
		qsLen := len(p.DatasetLandingPage.QualityStatements)
		p.DatasetLandingPage.QualityStatements[qsLen-1].CssClasses = append(p.DatasetLandingPage.QualityStatements[qsLen-1].CssClasses, "ons-u-mb-l")
	}

	return p
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
				pDim.ShowChange = pDim.IsAreaType || isMultivariate
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

func getAnalytics(dimensions []model.Dimension) map[string]string {
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
