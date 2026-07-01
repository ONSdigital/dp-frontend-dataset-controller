package mapper

import (
	"context"
	"strconv"
	"strings"
	"time"

	core "github.com/ONSdigital/dis-design-system-go/model"
	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/static"
	dpTopicApiModels "github.com/ONSdigital/dp-topic-api/models"
	"github.com/ONSdigital/log.go/v2/log"
)

// CreateStaticLandingPage creates a static-overview page based on api model responses
func CreateStaticOverviewPage(ctx context.Context, basePage core.Page, datasetDetails dpDatasetApiModels.Dataset,
	version dpDatasetApiModels.Version, allVersions []dpDatasetApiModels.Version, isEnableMultivariate bool, topicObjectList []*dpTopicApiModels.Topic, isPublishing bool, enableApprovalButton bool,
) static.Page {
	p := CreateStaticBasePage(basePage, datasetDetails, version, allVersions, isEnableMultivariate, topicObjectList)

	// SET STATE FOR APPROVAL BUTTON USE
	p.DatasetLandingPage.State = version.State
	if isPublishing && enableApprovalButton {
		p.ShowApprove = true
	}

	// DOWNLOADS
	if version.Distributions != nil {
		distributions := *version.Distributions
		for _, distribution := range distributions {
			p.Version.Downloads = append(p.Version.Downloads, sharedModel.Download{
				Extension: strings.ToLower(distribution.Format.String()),
				Size:      strconv.FormatInt(distribution.ByteSize, 10),
				URI:       distribution.DownloadURL,
			})
		}
		p.Version.Downloads = orderDownloads(p.Version.Downloads)
		p.DatasetLandingPage.HasDownloads = true
	}

	// ANALYTICS
	p.PreGTMJavaScript = append(
		p.PreGTMJavaScript,
		getDataLayerJavaScript(setGTMDataLayerValuesForStaticDatasets(ctx, datasetDetails, version, topicObjectList)),
	)

	// FINAL FORMATTING
	p.DatasetLandingPage.QualityStatements = formatStaticPanels(p.DatasetLandingPage.QualityStatements)
	return p
}

// formatPanels is a helper function given an array of panels will format the final panel with the appropriate css class
func formatStaticPanels(panels []static.Panel) []static.Panel {
	if len(panels) > 0 {
		panelLen := len(panels)
		panels[panelLen-1].CSSClasses = append(panels[panelLen-1].CSSClasses, "ons-u-mb-l")
	}
	return panels
}

// setGTMDataLayerValuesForStaticDatasets returns a map to add to the data layer which will be used on static dataset version page
func setGTMDataLayerValuesForStaticDatasets(ctx context.Context, datasetDetails dpDatasetApiModels.Dataset, version dpDatasetApiModels.Version, topics []*dpTopicApiModels.Topic) map[string]string {
	dataLayer := make(map[string]string, 11)
	dataLayer["product"] = "dataset-catalogue"
	dataLayer["contentType"] = "datasets"
	dataLayer["contentSubtype"] = "versions"

	if len(topics) > 0 {
		dataLayer["contentGroup"] = topics[0].Title
	}

	dataLayer["contentTitle"] = datasetDetails.Title + ": " + version.EditionTitle
	dataLayer["outputSeries"] = datasetDetails.ID
	dataLayer["outputEdition"] = version.Edition
	dataLayer["outputVersion"] = strconv.Itoa(version.Version)
	if version.ReleaseDate != "" {
		relDate, err := time.Parse("2006-01-02T15:04:05Z07:00", version.ReleaseDate)
		if err == nil {
			dataLayer["releaseDate"] = relDate.Format("20060102")
		} else {
			log.Error(ctx, "failed to parse release date for GTM dataLayer", err)
		}
	}

	dataLayer["lastUpdateDate"] = version.LastUpdated.Format("20060102")

	// "yes" / "no" were requested by analytics team rather than true / false
	isLatestRelease := "no"
	if datasetDetails.Links != nil && datasetDetails.Links.LatestVersion != nil && datasetDetails.Links.LatestVersion.ID == strconv.Itoa(version.Version) {
		isLatestRelease = "yes"
	}
	dataLayer["latestRelease"] = isLatestRelease

	return dataLayer
}
