package mapper

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/contactDetails"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetLandingPageCensus"
	"github.com/ONSdigital/dp-renderer/helper"
	coreModel "github.com/ONSdigital/dp-renderer/model"
)

// Constants...
const (
	AlertType           = "alert"
	CorrectionAlertType = "correction"
)

func CreateCensusBasePage(isEnableMultivariate bool, ctx context.Context, req *http.Request, basePage coreModel.Page, d dataset.DatasetDetails, version dataset.Version, opts []dataset.Options, initialVersionReleaseDate string, hasOtherVersions bool, allVersions []dataset.Version, latestVersionNumber int, latestVersionURL, lang string, queryStrValues []string, maxNumberOfOptions int, isValidationError, isFilterOutput, hasNoAreaOptions bool, filterOutput map[string]filter.Download, fDims []sharedModel.FilterDimension, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) datasetLandingPageCensus.Page {
	p := datasetLandingPageCensus.Page{
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
				p.DatasetLandingPage.Panels = append(p.DatasetLandingPage.Panels, datasetLandingPageCensus.Panel{
					DisplayIcon: true,
					Body:        helper.Localise("HasCorrectionNotice", lang, 1),
					CssClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
				})
			case AlertType:
				p.DatasetLandingPage.Panels = append(p.DatasetLandingPage.Panels, datasetLandingPageCensus.Panel{
					DisplayIcon: true,
					Body:        helper.Localise("HasAlert", lang, 1, alert.Description),
					CssClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
				})
			}
		}
	}

	// TABLE OF CONTENTS
	p.TableOfContents = buildTableOfContents(p, d, hasOtherVersions)

	// VERSIONS TABLE
	if hasOtherVersions {
		for _, ver := range allVersions {
			var version sharedModel.Version
			version.VersionNumber = ver.Version
			version.ReleaseDate = ver.ReleaseDate
			versionUrl := helpers.DatasetVersionUrl(ver.Links.Dataset.ID, ver.Edition, strconv.Itoa(ver.Version))
			version.VersionURL = versionUrl
			version.IsCurrentPage = versionUrl == req.URL.Path
			mapCorrectionAlert(&ver, &version)

			p.Versions = append(p.Versions, version)
		}

		sort.Slice(p.Versions, func(i, j int) bool { return p.Versions[i].VersionNumber > p.Versions[j].VersionNumber })

		p.DatasetLandingPage.LatestVersionURL = latestVersionURL
	}

	// LATEST VERSIONS PANEL
	if latestVersionNumber != version.Version && hasOtherVersions {
		p.DatasetLandingPage.Panels = append(p.DatasetLandingPage.Panels, datasetLandingPageCensus.Panel{
			DisplayIcon: true,
			Body:        helper.Localise("HasNewVersion", lang, 1, latestVersionURL),
			CssClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
		})
	}

	// SHARING LINKS
	currentUrl := helpers.GetCurrentUrl(lang, p.SiteDomain, req.URL.Path)
	p.DatasetLandingPage.DatasetURL = currentUrl
	p.DatasetLandingPage.ShareDetails = buildSharingDetails(d, lang, currentUrl)

	// RELATED CONTENT
	p.DatasetLandingPage.RelatedContentItems = []datasetLandingPageCensus.RelatedContentItem{}
	if d.RelatedContent != nil {
		for _, content := range *d.RelatedContent {
			p.DatasetLandingPage.RelatedContentItems = append(p.DatasetLandingPage.RelatedContentItems, datasetLandingPageCensus.RelatedContentItem{
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

func getContactDetails(d dataset.DatasetDetails) (contactDetails.ContactDetails, bool) {
	details := contactDetails.ContactDetails{}
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

func getReleaseDate(initialDate string, alternateDate string) string {
	if initialDate == "" {
		return alternateDate
	} else {
		return initialDate
	}
}

func buildSharingDetails(d dataset.DatasetDetails, lang, currentUrl string) datasetLandingPageCensus.ShareDetails {
	shareDetails := datasetLandingPageCensus.ShareDetails{}
	shareDetails.Language = lang
	shareDetails.ShareLocations = []datasetLandingPageCensus.Share{
		{
			Title: "Facebook",
			Link:  helpers.GenerateSharingLink("facebook", currentUrl, d.Title),
			Icon:  "facebook",
		},
		{
			Title: "Twitter",
			Link:  helpers.GenerateSharingLink("twitter", currentUrl, d.Title),
			Icon:  "twitter",
		},
		{
			Title: "LinkedIn",
			Link:  helpers.GenerateSharingLink("linkedin", currentUrl, d.Title),
			Icon:  "linkedin",
		},
		{
			Title: "Email",
			Link:  helpers.GenerateSharingLink("email", currentUrl, d.Title),
			Icon:  "email",
		},
	}
	return shareDetails
}

func buildTableOfContents(p datasetLandingPageCensus.Page, d dataset.DatasetDetails, hasOtherVersions bool) coreModel.TableOfContents {
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
