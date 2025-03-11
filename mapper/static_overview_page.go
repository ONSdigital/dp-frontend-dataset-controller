package mapper

import (
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/static"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
)

// CreateStaticLandingPage creates a static-overview page based on api model responses
func CreateStaticOverviewPage(basePage coreModel.Page, datasetDetails dataset.DatasetDetails,
	version dataset.Version, allVersions []dataset.Version, isEnableMultivariate bool,
) static.Page {
	p := CreateStaticBasePage(basePage, datasetDetails, version, allVersions, isEnableMultivariate)

	// DOWNLOADS
	for ext, download := range version.Downloads {
		p.Version.Downloads = append(p.Version.Downloads, sharedModel.Download{
			Extension: strings.ToLower(ext),
			Size:      download.Size,
			URI:       download.URL,
		})
	}
	p.Version.Downloads = orderDownloads(p.Version.Downloads)

	// HasDownloads is the flag used to render the template
	if len(version.Downloads) > 0 {
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
