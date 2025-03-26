package mapper

import (
	"net/http"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/static"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
)

// CreateStaticLandingPage creates a static-overview page based on api model responses
func CreateStaticOverviewPage(req *http.Request, basePage coreModel.Page, d dataset.DatasetDetails, version dataset.Version,
	categorisationsMap map[string]int, initialVersionReleaseDate string, hasOtherVersions bool, allVersions []dataset.Version,
	latestVersionNumber int, latestVersionURL, lang string, queryStrValues []string, isValidationError bool, serviceMessage string,
	emergencyBannerContent zebedee.EmergencyBanner, isEnableMultivariate bool,
) static.Page {
	p := CreateStaticBasePage(req, basePage, d, version, initialVersionReleaseDate, hasOtherVersions, allVersions,
		latestVersionNumber, latestVersionURL, lang, isValidationError, serviceMessage, emergencyBannerContent,
		isEnableMultivariate)

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

	// FEEDBACK API
	p.FeatureFlags.FeedbackAPIURL = cfg.FeedbackAPIURL

	return p
}
