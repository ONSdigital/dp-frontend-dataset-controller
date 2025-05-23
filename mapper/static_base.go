package mapper

import (
	"sort"
	"strconv"
	"strings"

	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/publisher"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/static"
	"github.com/ONSdigital/dp-renderer/v2/helper"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
	dpTopicApiModels "github.com/ONSdigital/dp-topic-api/models"
)

// CreateCensusBasePage builds a base datasetLandingPageCensus.Page with shared functionality between Dataset Landing Pages and Filter Output pages
func CreateStaticBasePage(basePage coreModel.Page, d dpDatasetApiModels.Dataset, version dpDatasetApiModels.Version,
	allVersions []dpDatasetApiModels.Version, isEnableMultivariate bool, topicObjectList []dpTopicApiModels.Topic,
) static.Page {
	var editionStr string

	p := static.Page{
		Page: basePage,
	}

	// Use edition title string if available
	if version.EditionTitle != "" {
		editionStr = version.EditionTitle
	} else {
		editionStr = version.Edition
	}
	p.Version.Edition = editionStr

	hasOtherVersions := false
	initialVersionReleaseDate := ""
	isNationalStatistic := false
	latestVersionNumber := 1

	// Loop through versions to find info
	for i := range allVersions {
		singleVersion := &allVersions[i]
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

	latestVersionURL := helpers.DatasetVersionURL(d.ID, version.Edition, strconv.Itoa(latestVersionNumber))

	if d.NationalStatistic != nil {
		isNationalStatistic = *d.NationalStatistic
	}
	p.IsNationalStatistic = isNationalStatistic
	p.ContactDetails, p.HasContactDetails = getContactDetails(d)

	p.Version.ReleaseDate = version.ReleaseDate
	p.ReleaseDate = getReleaseDate(initialVersionReleaseDate, p.Version.ReleaseDate)

	p.DatasetLandingPage.Description = strings.Split(d.Description, "\n")
	p.DatasetLandingPage.HasOtherVersions = hasOtherVersions

	p.DatasetLandingPage.IsMultivariate = strings.Contains(d.Type, "multivariate") && isEnableMultivariate
	p.DatasetLandingPage.IsFlexibleForm = p.DatasetLandingPage.IsMultivariate || strings.Contains(d.Type, "flexible")

	p.Publisher = getPublisherDetails(d)

	p.UsageNotes = getUsageDetails(version)

	p.DatasetLandingPage.OSRLogo = helpers.GetOSRLogoDetails(basePage.Language)
	p.DatasetLandingPage.NextRelease = d.NextRelease

	// CENSUS BRANDING
	p.ShowCensusBranding = false

	// BREADCRUMBS
	baseURL := "https://www.ons.gov.uk/"

	p.Breadcrumb = CreateBreadcrumbsFromTopicList(baseURL, topicObjectList)

	// ALERTS
	if version.Alerts != nil {
		for _, alert := range *version.Alerts {
			switch alert.Type {
			case CorrectionAlertType:
				p.DatasetLandingPage.Panels = append(p.DatasetLandingPage.Panels, static.Panel{
					DisplayIcon: true,
					Body:        []string{helper.Localise("HasCorrectionNotice", basePage.Language, 1)},
					CSSClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
				})
			case AlertType:
				p.DatasetLandingPage.Panels = append(p.DatasetLandingPage.Panels, static.Panel{
					DisplayIcon: true,
					Body:        []string{helper.Localise("HasAlert", basePage.Language, 1, alert.Description)},
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
			version.IsCurrentPage = versionURL == basePage.URI
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
			Body:        []string{helper.Localise("HasNewVersion", basePage.Language, 1, latestVersionURL)},
			CSSClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
		})
	}

	// SHARING LINKS
	currentURL := helpers.GetCurrentURL(basePage.Language, p.SiteDomain, basePage.URI)
	p.DatasetLandingPage.DatasetURL = currentURL
	p.DatasetLandingPage.ShareDetails = buildStaticSharingDetails(d, basePage.Language, currentURL)

	// RELATED CONTENT
	p.DatasetLandingPage.RelatedContentItems = []sharedModel.RelatedContentItem{}
	if d.RelatedContent != nil {
		for _, content := range d.RelatedContent {
			p.DatasetLandingPage.RelatedContentItems = append(p.DatasetLandingPage.RelatedContentItems, sharedModel.RelatedContentItem{
				Title: content.Title,
				Link:  content.HRef,
				Text:  content.Description,
			})
		}
	}

	return p
}

func buildStaticSharingDetails(d dpDatasetApiModels.Dataset, lang, currentURL string) sharedModel.ShareDetails {
	shareDetails := sharedModel.ShareDetails{}
	shareDetails.Language = lang
	shareDetails.ShareLocations = []sharedModel.Share{
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

func buildStaticTableOfContents(p static.Page, d dpDatasetApiModels.Dataset, hasOtherVersions bool) coreModel.TableOfContents {
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

func getPublisherDetails(d dpDatasetApiModels.Dataset) publisher.Publisher {
	publisherObject := publisher.Publisher{}

	// TODO: this code should be refactored to be uncoupled from predefined variables
	// Currennt available variables:
	// 		URL  string `json:"href"`
	// 		Name string `json:"name"`
	// 		Type string `json:"type"`

	if d.Publisher != nil {
		incomingPublisherDataset := *d.Publisher

		if incomingPublisherDataset.HRef != "" {
			publisherObject.URL = incomingPublisherDataset.HRef
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
func getUsageDetails(v dpDatasetApiModels.Version) []static.UsageNote {
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
