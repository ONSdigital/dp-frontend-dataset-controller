package mapper

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/publisher"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/static"
	"github.com/ONSdigital/dp-renderer/v2/helper"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
)

// CreateCensusBasePage builds a base datasetLandingPageCensus.Page with shared functionality between Dataset Landing Pages and Filter Output pages
func CreateStaticBasePage(
	req *http.Request,
	basePage coreModel.Page,
	d dataset.DatasetDetails,
	version dataset.Version,
	initialVersionReleaseDate string,
	hasOtherVersions bool,
	allVersions []dataset.Version,
	latestVersionNumber int,
	latestVersionURL,
	lang string,
	isValidationError bool,
	serviceMessage string,
	emergencyBannerContent zebedee.EmergencyBanner,
	isEnableMultivariate bool,
	topicSlugList []string,
) static.Page {
	p := static.Page{
		Page: basePage,
	}

	MapCookiePreferences(req, &p.Page.CookiesPreferencesSet, &p.Page.CookiesPolicy)
	// PAGE META-DATA
	p.Type = d.Type
	p.Metadata.Title = d.Title
	p.Language = lang
	p.URI = req.URL.Path
	p.Metadata.Description = d.Description
	p.IsNationalStatistic = d.NationalStatistic
	p.DatasetId = d.ID
	p.ContactDetails, p.HasContactDetails = getContactDetails(d)

	p.Version.ReleaseDate = version.ReleaseDate
	p.ReleaseDate = getReleaseDate(initialVersionReleaseDate, p.Version.ReleaseDate)

	p.DatasetLandingPage.Description = strings.Split(d.Description, "\n")
	p.DatasetLandingPage.HasOtherVersions = hasOtherVersions

	p.DatasetLandingPage.IsMultivariate = strings.Contains(d.Type, "multivariate") && isEnableMultivariate
	p.DatasetLandingPage.IsFlexibleForm = p.DatasetLandingPage.IsMultivariate || strings.Contains(d.Type, "flexible")

	p.Publisher = getPublisherDetails(d)

	p.UsageNotes = getUsageDetails(version)

	p.DatasetLandingPage.OSRLogo = helpers.GetOSRLogoDetails(lang)
	p.DatasetLandingPage.NextRelease = d.NextRelease

	

	// SITE-WIDE BANNERS
	p.BetaBannerEnabled = true
	p.ServiceMessage = serviceMessage
	p.EmergencyBanner = mapEmergencyBanner(emergencyBannerContent)

	// CENSUS BRANDING
	p.ShowCensusBranding = false

	// BREADCRUMBS
	var breadcrumbsObject []coreModel.TaxonomyNode

	for _, item := range topicSlugList {
		entry := coreModel.TaxonomyNode{
			Title: item,
			URI: "#",
		}
		breadcrumbsObject = append(breadcrumbsObject, entry)
	}
	p.Breadcrumb = breadcrumbsObject
	
	// p.Breadcrumb = []coreModel.TaxonomyNode{
	// 	{
	// 		Title: "Home",
	// 		URI:   "https://www.ons.gov.uk/",
	// 	},
	// 	{
	// 		Title: "Overview page",
	// 		URI:   "#",
	// 	},
	// }



	// FEEDBACK API
	p.FeatureFlags.FeedbackAPIURL = cfg.FeedbackAPIURL

	// BACK LINK
	p.BackTo = coreModel.BackTo{
		Text: coreModel.Localisation{
			LocaleKey: "BackToContents",
			Plural:    4,
		},
		AnchorFragment: "toc",
	}

	// ALERTS
	if version.Alerts != nil {
		for _, alert := range *version.Alerts {
			switch alert.Type {
			case CorrectionAlertType:
				p.DatasetLandingPage.Panels = append(p.DatasetLandingPage.Panels, static.Panel{
					DisplayIcon: true,
					Body:        []string{helper.Localise("HasCorrectionNotice", lang, 1)},
					CSSClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
				})
			case AlertType:
				p.DatasetLandingPage.Panels = append(p.DatasetLandingPage.Panels, static.Panel{
					DisplayIcon: true,
					Body:        []string{helper.Localise("HasAlert", lang, 1, alert.Description)},
					CSSClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
				})
			}
		}
	}

	// TABLE OF CONTENTS
	p.TableOfContents = buildStaticTableOfContents(p, d, hasOtherVersions)

	// VERSIONS TABLE
	if hasOtherVersions {
		for i := range allVersions {
			var version sharedModel.Version
			version.VersionNumber = allVersions[i].Version
			version.ReleaseDate = allVersions[i].ReleaseDate
			versionURL := helpers.DatasetVersionURL(
				allVersions[i].Links.Dataset.ID,
				allVersions[i].Edition,
				strconv.Itoa(allVersions[i].Version))
			version.VersionURL = versionURL
			version.IsCurrentPage = versionURL == req.URL.Path
			mapCorrectionAlert(&allVersions[i], &version)

			p.Versions = append(p.Versions, version)
		}

		sort.Slice(p.Versions, func(i, j int) bool { return p.Versions[i].VersionNumber > p.Versions[j].VersionNumber })

		p.DatasetLandingPage.LatestVersionURL = latestVersionURL
	}

	// LATEST VERSIONS PANEL
	if latestVersionNumber != version.Version && hasOtherVersions {
		p.DatasetLandingPage.Panels = append(p.DatasetLandingPage.Panels, static.Panel{
			DisplayIcon: true,
			Body:        []string{helper.Localise("HasNewVersion", lang, 1, latestVersionURL)},
			CSSClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
		})
	}

	// SHARING LINKS
	currentURL := helpers.GetCurrentURL(lang, p.SiteDomain, req.URL.Path)
	p.DatasetLandingPage.DatasetURL = currentURL
	p.DatasetLandingPage.ShareDetails = buildStaticSharingDetails(d, lang, currentURL)

	// RELATED CONTENT
	p.DatasetLandingPage.RelatedContentItems = []static.RelatedContentItem{}
	if d.RelatedContent != nil {
		for _, content := range *d.RelatedContent {
			p.DatasetLandingPage.RelatedContentItems = append(p.DatasetLandingPage.RelatedContentItems, static.RelatedContentItem{
				Title: content.Title,
				Link:  content.HRef,
				Text:  content.Description,
			})
		}
	}

	// ERRORS
	if isValidationError {
		p.Error = coreModel.Error{
			Title: p.Metadata.Title,
			ErrorItems: []coreModel.ErrorItem{
				{
					Description: coreModel.Localisation{
						LocaleKey: "GetDataValidationError",
						Plural:    1,
					},
					URL: "#select-format-error",
				},
			},
			Language: lang,
		}
	}

	return p
}

func buildStaticSharingDetails(d dataset.DatasetDetails, lang, currentURL string) static.ShareDetails {
	shareDetails := static.ShareDetails{}
	shareDetails.Language = lang
	shareDetails.ShareLocations = []static.Share{
		{
			Title: "Email",
			Link:  helpers.GenerateSharingLink("email", currentURL, d.Title),
			Icon:  "email",
		},
		{
			Title: "Facebook",
			Link:  helpers.GenerateSharingLink("facebook", currentURL, d.Title),
			Icon:  "facebook",
		},
		{
			Title: "LinkedIn",
			Link:  helpers.GenerateSharingLink("linkedin", currentURL, d.Title),
			Icon:  "linkedin",
		},
		{
			Title: "X",
			Link:  helpers.GenerateSharingLink("x", currentURL, d.Title),
			Icon:  "x",
		},
	}
	return shareDetails
}

func buildStaticTableOfContents(p static.Page, d dataset.DatasetDetails, hasOtherVersions bool) coreModel.TableOfContents {
	sections := make(map[string]coreModel.ContentSection)
	displayOrder := make([]string, 0)

	tableOfContents := coreModel.TableOfContents{
		AriaLabel: coreModel.Localisation{
			LocaleKey: "ContentsAria",
			Plural:    1,
		},
		Title: coreModel.Localisation{
			LocaleKey: "StaticTocHeading",
			Plural:    1,
		},
	}

	sections["get-data"] = coreModel.ContentSection{
		Title: coreModel.Localisation{
			LocaleKey: "GetData",
			Plural:    1,
		},
	}
	displayOrder = append(displayOrder, "get-data")

	if len(p.UsageNotes) > 0 {
		sections["usage-notes"] = coreModel.ContentSection{
			Title: coreModel.Localisation{
				LocaleKey: "UsageNotes",
				Plural:    1,
			},
		}
		displayOrder = append(displayOrder, "usage-notes")
	}

	if d.RelatedContent != nil {
		sections["related-content"] = coreModel.ContentSection{
			Title: coreModel.Localisation{
				LocaleKey: "RelatedContentTitle",
				Plural:    1,
			},
		}
		displayOrder = append(displayOrder, "related-content")
	}

	if hasOtherVersions {
		sections["version-history"] = coreModel.ContentSection{
			Title: coreModel.Localisation{
				LocaleKey: "VersionHistory",
				Plural:    1,
			},
		}
		displayOrder = append(displayOrder, "version-history")
	}

	if p.HasContactDetails {
		sections["contact"] = coreModel.ContentSection{
			Title: coreModel.Localisation{
				LocaleKey: "DatasetContactDetailsStatic",
				Plural:    1,
			},
		}
		displayOrder = append(displayOrder, "contact")
	}

	tableOfContents.Sections = sections
	tableOfContents.DisplayOrder = displayOrder

	return tableOfContents
}

func getPublisherDetails(d dataset.DatasetDetails) publisher.Publisher {
	publisherObject := publisher.Publisher{}

	// TODO: this code should be refactored to be uncoupled from predefined variables
	// Currennt available variables:
	// 		URL  string `json:"href"`
	// 		Name string `json:"name"`
	// 		Type string `json:"type"`

	if d.Publisher != nil {
		incomingPublisherDataset := *d.Publisher

		if incomingPublisherDataset.URL != "" {
			publisherObject.URL = incomingPublisherDataset.URL
		}

		if incomingPublisherDataset.Name != "" {
			publisherObject.Name = incomingPublisherDataset.Name
		}
		if incomingPublisherDataset.Type != "" {
			publisherObject.Type = incomingPublisherDataset.Type
		}
	}

	return publisherObject
}

// grab the usage notes
func getUsageDetails(v dataset.Version) []static.UsageNote {
	usageNotesList := []static.UsageNote{}

	if v.UsageNotes != nil {
		for _, usageNote := range *v.UsageNotes {
			usageNotesList = append(usageNotesList, static.UsageNote{
				Title: usageNote.Title,
				Note:  usageNote.Note,
			})
		}
	}
	return usageNotesList
}
