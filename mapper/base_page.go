package mapper

import (
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	dpRendererModel "github.com/ONSdigital/dp-renderer/v2/model"
)

// Updates input `basePage` to include common dataset overview attributes, homepage content and
// dataset details across all dataset types
func UpdateBasePage(basePage *dpRendererModel.Page, datasetDetails dataset.DatasetDetails,
	homepageContent zebedee.HomepageContent, isValidationError bool, lang string, request *http.Request) {

	basePage.BackTo = dpRendererModel.BackTo{
		Text: dpRendererModel.Localisation{
			LocaleKey: "BackToContents",
			Plural:    4,
		},
		AnchorFragment: "toc",
	}
	basePage.BetaBannerEnabled = true

	// Cookies
	MapCookiePreferences(request, &basePage.CookiesPreferencesSet, &basePage.CookiesPolicy)

	basePage.DatasetId = datasetDetails.ID
	basePage.EmergencyBanner = mapEmergencyBanner(homepageContent.EmergencyBanner)

	if isValidationError {
		basePage.Error = dpRendererModel.Error{
			Title: basePage.Metadata.Title,
			ErrorItems: []dpRendererModel.ErrorItem{
				{
					Description: dpRendererModel.Localisation{
						LocaleKey: "GetDataValidationError",
						Plural:    1,
					},
					URL: "#select-format-error",
				},
			},
			Language: lang,
		}
	}

	basePage.FeatureFlags.FeedbackAPIURL = cfg.FeedbackAPIURL
	basePage.Language = lang
	basePage.Metadata.Description = datasetDetails.Description
	basePage.Metadata.Title = datasetDetails.Title
	basePage.ServiceMessage = homepageContent.ServiceMessage
	basePage.Type = datasetDetails.Type
	basePage.URI = request.URL.Path

	// basePage.DatasetTitle = ""
	// basePage.URI = ""
	// basePage.Taxonomy = ""
	// basePage.Breadcrumb = ""
	// basePage.IsInFilterBreadcrumb = ""
	// basePage.ServiceMessage = ""
	// basePage.SearchDisabled = ""
	// basePage.SiteDomain = ""
	// basePage.PatternLibraryAssetsPath = ""
	// basePage.Language = ""
	// basePage.IncludeAssetsIntegrityAttributes = ""
	// basePage.ReleaseDate = ""
	// basePage.BetaBannerEnabled = ""
	// basePage.CookiesPreferencesSet = ""
	// basePage.CookiesPolicy = ""
	// basePage.HasJSONLD = ""
	// basePage.FeatureFlags = ""
	// basePage.Error = ""
	// basePage.EmergencyBanner = ""
	// basePage.Collapsible = ""
	// basePage.Pagination = ""
	// basePage.TableOfContents = ""
	// basePage.BackTo = ""
	// basePage.SearchNoIndexEnabled = ""
	// basePage.NavigationContent = ""
	// basePage.PreGTMJavaScript = ""
	// basePage.RemoveGalleryBackground = ""
	// basePage.Feedback = ""
	// basePage.Enable500ErrorPageStyling = ""
}
