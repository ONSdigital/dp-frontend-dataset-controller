package mapper

import (
	"sort"
	"strconv"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model"
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
func CreateCensusBasePage(basePage coreModel.Page, datasetDetails dataset.DatasetDetails, version dataset.Version,
	allVersions []dataset.Version, isEnableMultivariate bool,
) census.Page {
	censusPage := census.Page{
		Page: basePage,
	}

	hasOtherVersions := false
	initialVersionReleaseDate := ""
	latestVersionNumber := 1

	// Loop through versions to find info
	for _, singleVersion := range allVersions {
		// Find the initial version release data
		if singleVersion.Version == 1 {
			initialVersionReleaseDate = singleVersion.ReleaseDate
		}
		// Find the latest version number
		if singleVersion.Version > latestVersionNumber {
			latestVersionNumber = singleVersion.Version
		}
	}

	// Set `hasOtherVersions` based on length of input `allVersions`
	if len(allVersions) > 1 {
		hasOtherVersions = true
	}

	latestVersionURL := helpers.DatasetVersionURL(datasetDetails.ID, version.Edition, strconv.Itoa(latestVersionNumber))

	censusPage.IsNationalStatistic = datasetDetails.NationalStatistic
	censusPage.ContactDetails, censusPage.HasContactDetails = getContactDetails(datasetDetails)

	censusPage.Version.ReleaseDate = version.ReleaseDate
	censusPage.ReleaseDate = getReleaseDate(initialVersionReleaseDate, censusPage.Version.ReleaseDate)

	censusPage.DatasetLandingPage.Description = strings.Split(datasetDetails.Description, "\n")
	censusPage.DatasetLandingPage.HasOtherVersions = hasOtherVersions

	censusPage.DatasetLandingPage.IsMultivariate = strings.Contains(datasetDetails.Type, "multivariate") && isEnableMultivariate
	censusPage.DatasetLandingPage.IsFlexibleForm = censusPage.DatasetLandingPage.IsMultivariate || strings.Contains(datasetDetails.Type, "flexible")

	// CENSUS BRANDING
	censusPage.ShowCensusBranding = datasetDetails.Survey == "census"

	// BREADCRUMBS
	censusPage.Breadcrumb = []coreModel.TaxonomyNode{
		{
			Title: "Home",
			URI:   "/",
		},
		{
			Title: "Census",
			URI:   "/census",
		},
	}

	// ALERTS
	if version.Alerts != nil {
		for _, alert := range *version.Alerts {
			switch alert.Type {
			case CorrectionAlertType:
				censusPage.DatasetLandingPage.Panels = append(censusPage.DatasetLandingPage.Panels, census.Panel{
					DisplayIcon: true,
					Body:        []string{helper.Localise("HasCorrectionNotice", censusPage.Language, 1)},
					CSSClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
				})
			case AlertType:
				censusPage.DatasetLandingPage.Panels = append(censusPage.DatasetLandingPage.Panels, census.Panel{
					DisplayIcon: true,
					Body:        []string{helper.Localise("HasAlert", censusPage.Language, 1, alert.Description)},
					CSSClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
				})
			}
		}
	}

	// TABLE OF CONTENTS
	censusPage.TableOfContents = buildTableOfContents(censusPage, datasetDetails, hasOtherVersions)

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
			version.IsCurrentPage = versionURL == censusPage.URI
			mapCorrectionAlert(&allVersions[i], &version)

			censusPage.Versions = append(censusPage.Versions, version)
		}

		sort.Slice(censusPage.Versions, func(i, j int) bool {
			return censusPage.Versions[i].VersionNumber > censusPage.Versions[j].VersionNumber
		})

		censusPage.DatasetLandingPage.LatestVersionURL = latestVersionURL
	}

	// LATEST VERSIONS PANEL
	if latestVersionNumber != version.Version && hasOtherVersions {
		censusPage.DatasetLandingPage.Panels = append(censusPage.DatasetLandingPage.Panels, census.Panel{
			DisplayIcon: true,
			Body:        []string{helper.Localise("HasNewVersion", censusPage.Language, 1, latestVersionURL)},
			CSSClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
		})
	}

	// SHARING LINKS
	currentURL := helpers.GetCurrentURL(censusPage.Language, censusPage.SiteDomain, censusPage.URI)
	censusPage.DatasetLandingPage.DatasetURL = currentURL
	censusPage.DatasetLandingPage.ShareDetails = buildSharingDetails(datasetDetails, censusPage.Language, currentURL)

	// RELATED CONTENT
	censusPage.DatasetLandingPage.RelatedContentItems = []model.RelatedContentItem{}
	if datasetDetails.RelatedContent != nil {
		for _, content := range *datasetDetails.RelatedContent {
			censusPage.DatasetLandingPage.RelatedContentItems = append(censusPage.DatasetLandingPage.RelatedContentItems, model.RelatedContentItem{
				Title: content.Title,
				Link:  content.HRef,
				Text:  content.Description,
			})
		}
	}

	return censusPage
}

func getContactDetails(d dataset.DatasetDetails) (contact.Details, bool) {
	details := contact.Details{}
	hasContactDetails := false

	if d.Contacts != nil && len(*d.Contacts) > 0 {
		contacts := *d.Contacts
		if d.Type == "static" {
			if contacts[0].Name != "" {
				details.Name = contacts[0].Name
				hasContactDetails = true
			}
		}
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

func buildSharingDetails(d dataset.DatasetDetails, lang, currentURL string) model.ShareDetails {
	shareDetails := model.ShareDetails{}
	shareDetails.Language = lang
	shareDetails.ShareLocations = []model.Share{
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
