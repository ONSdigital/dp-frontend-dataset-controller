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
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/census"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/contact"
	"github.com/ONSdigital/dp-renderer/v2/helper"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
)

// Constants...
const (
	AlertType           = "alert"
	CorrectionAlertType = "correction"
)

// CreateCensusBasePage builds a base datasetLandingPageCensus.Page with shared functionality between Dataset Landing Pages and Filter Output pages
func CreateCensusBasePage(req *http.Request, basePage coreModel.Page, d dataset.DatasetDetails, version dataset.Version, initialVersionReleaseDate string, hasOtherVersions bool, allVersions []dataset.Version, latestVersionNumber int, latestVersionURL, lang string, isValidationError bool, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner, isEnableMultivariate bool) census.Page {
	p := census.Page{
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

	// SITE-WIDE BANNERS
	p.BetaBannerEnabled = true
	p.ServiceMessage = serviceMessage
	p.EmergencyBanner = mapEmergencyBanner(emergencyBannerContent)

	// CENSUS BRANDING
	p.ShowCensusBranding = d.Survey == "census"

	// BREADCRUMBS
	p.Breadcrumb = []coreModel.TaxonomyNode{
		{
			Title: "Home",
			URI:   "/",
		},
		{
			Title: "Census",
			URI:   "/census",
		},
	}

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
				p.DatasetLandingPage.Panels = append(p.DatasetLandingPage.Panels, census.Panel{
					DisplayIcon: true,
					Body:        []string{helper.Localise("HasCorrectionNotice", lang, 1)},
					CSSClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
				})
			case AlertType:
				p.DatasetLandingPage.Panels = append(p.DatasetLandingPage.Panels, census.Panel{
					DisplayIcon: true,
					Body:        []string{helper.Localise("HasAlert", lang, 1, alert.Description)},
					CSSClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
				})
			}
		}
	}

	// TABLE OF CONTENTS
	p.TableOfContents = buildTableOfContents(p, d, hasOtherVersions)

	// VERSIONS TABLE
	if hasOtherVersions {
		for i, ver := range allVersions {
			var version sharedModel.Version
			version.VersionNumber = ver.Version
			version.ReleaseDate = ver.ReleaseDate
			versionURL := helpers.DatasetVersionURL(ver.Links.Dataset.ID, ver.Edition, strconv.Itoa(ver.Version))
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
		p.DatasetLandingPage.Panels = append(p.DatasetLandingPage.Panels, census.Panel{
			DisplayIcon: true,
			Body:        []string{helper.Localise("HasNewVersion", lang, 1, latestVersionURL)},
			CSSClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
		})
	}

	// SHARING LINKS
	currentURL := helpers.GetCurrentURL(lang, p.SiteDomain, req.URL.Path)
	p.DatasetLandingPage.DatasetURL = currentURL
	p.DatasetLandingPage.ShareDetails = buildSharingDetails(d, lang, currentURL)

	// RELATED CONTENT
	p.DatasetLandingPage.RelatedContentItems = []census.RelatedContentItem{}
	if d.RelatedContent != nil {
		for _, content := range *d.RelatedContent {
			p.DatasetLandingPage.RelatedContentItems = append(p.DatasetLandingPage.RelatedContentItems, census.RelatedContentItem{
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

func getContactDetails(d dataset.DatasetDetails) (contact.Details, bool) {
	details := contact.Details{}
	hasContactDetails := false

	if d.Contacts != nil && len(*d.Contacts) > 0 {
		contacts := *d.Contacts
		if contacts[0].Telephone != "" {
			details.Telephone = contacts[0].Telephone
			hasContactDetails = true
		}
		if contacts[0].Email != "" {
			details.Email = contacts[0].Email
			hasContactDetails = true
		}
	}

	return details, hasContactDetails
}

func getReleaseDate(initialDate, alternateDate string) string {
	if initialDate == "" {
		return alternateDate
	}
	return initialDate
}

func buildSharingDetails(d dataset.DatasetDetails, lang, currentURL string) census.ShareDetails {
	shareDetails := census.ShareDetails{}
	shareDetails.Language = lang
	shareDetails.ShareLocations = []census.Share{
		{
			Title: "Facebook",
			Link:  helpers.GenerateSharingLink("facebook", currentURL, d.Title),
			Icon:  "facebook",
		},
		{
			Title: "Twitter",
			Link:  helpers.GenerateSharingLink("twitter", currentURL, d.Title),
			Icon:  "twitter",
		},
		{
			Title: "LinkedIn",
			Link:  helpers.GenerateSharingLink("linkedin", currentURL, d.Title),
			Icon:  "linkedin",
		},
		{
			Title: "Email",
			Link:  helpers.GenerateSharingLink("email", currentURL, d.Title),
			Icon:  "email",
		},
	}
	return shareDetails
}

func buildTableOfContents(p census.Page, d dataset.DatasetDetails, hasOtherVersions bool) coreModel.TableOfContents {
	sections := make(map[string]coreModel.ContentSection)
	displayOrder := make([]string, 0)

	tableOfContents := coreModel.TableOfContents{
		AriaLabel: coreModel.Localisation{
			LocaleKey: "ContentsAria",
			Plural:    1,
		},
		Title: coreModel.Localisation{
			LocaleKey: "Contents",
			Plural:    1,
		},
	}

	sections["summary"] = coreModel.ContentSection{
		Title: coreModel.Localisation{
			LocaleKey: "Summary",
			Plural:    1,
		},
	}
	displayOrder = append(displayOrder, "summary")

	sections["variables"] = coreModel.ContentSection{
		Title: coreModel.Localisation{
			LocaleKey: "Variables",
			Plural:    4,
		},
	}
	displayOrder = append(displayOrder, "variables")

	sections["get-data"] = coreModel.ContentSection{
		Title: coreModel.Localisation{
			LocaleKey: "GetData",
			Plural:    1,
		},
	}
	displayOrder = append(displayOrder, "get-data")

	if p.HasContactDetails {
		sections["contact"] = coreModel.ContentSection{
			Title: coreModel.Localisation{
				LocaleKey: "ContactUs",
				Plural:    1,
			},
		}
		displayOrder = append(displayOrder, "contact")
	}

	sections["protecting-personal-data"] = coreModel.ContentSection{
		Title: coreModel.Localisation{
			LocaleKey: "ProtectingPersonalDataTitle",
			Plural:    1,
		},
	}
	displayOrder = append(displayOrder, "protecting-personal-data")

	if hasOtherVersions {
		sections["version-history"] = coreModel.ContentSection{
			Title: coreModel.Localisation{
				LocaleKey: "VersionHistory",
				Plural:    1,
			},
		}
		displayOrder = append(displayOrder, "version-history")
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

	tableOfContents.Sections = sections
	tableOfContents.DisplayOrder = displayOrder

	return tableOfContents
}
