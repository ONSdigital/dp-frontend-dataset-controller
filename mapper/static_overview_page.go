package mapper

import (
	"strconv"
	"strings"

	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/static"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
	dpTopicApiModels "github.com/ONSdigital/dp-topic-api/models"
)

// CreateStaticLandingPage creates a static-overview page based on api model responses
func CreateStaticOverviewPage(basePage coreModel.Page, datasetDetails dpDatasetApiModels.Dataset,
	version dpDatasetApiModels.Version, allVersions []dpDatasetApiModels.Version, isEnableMultivariate bool, topicObjectList []dpTopicApiModels.Topic, isPublishing bool,
) static.Page {
	p := CreateStaticBasePage(basePage, datasetDetails, version, allVersions, isEnableMultivariate, topicObjectList)

	// SET STATE FOR APPROVAL BUTTON USE
	p.DatasetLandingPage.State = version.State
	if isPublishing {
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
		getDataLayerJavaScript(getAnalytics(p.DatasetLandingPage.Dimensions)),
	)

	// FINAL FORMATTING
	p.DatasetLandingPage.QualityStatements = formatStaticPanels(p.DatasetLandingPage.QualityStatements)
	return p
}
