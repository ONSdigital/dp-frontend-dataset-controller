package mapper

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetPage"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-cookies/cookies"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetEditionsList"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetLandingPageCensus"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetLandingPageFilterable"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetVersionsList"
	coreModel "github.com/ONSdigital/dp-renderer/model"

	"github.com/ONSdigital/log.go/v2/log"
)

// TimeSlice allows sorting of a list of time.Time
type TimeSlice []time.Time

// Constants names
const (
	DimensionTime       = "time"
	DimensionAge        = "age"
	DimensionGeography  = "geography"
	SixteensVersion     = "77f1d9b"
	CorrectionAlertType = "correction"
)

func (p TimeSlice) Len() int {
	return len(p)
}

func (p TimeSlice) Less(i, j int) bool {
	return p[i].Before(p[j])
}

func (p TimeSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// Trim API version path prefix from breadcrumb URI, if present.
func getTrimmedBreadcrumbURI(ctx context.Context, breadcrumb zebedee.Breadcrumb, apiRouterVersion string) string {
	trimmedURI := breadcrumb.URI
	urlParsed, err := url.Parse(breadcrumb.URI)
	if err != nil {
		log.Warn(ctx, "wrong format for breadcrumb uri", log.Data{"breadcrumb": breadcrumb})
	} else {
		urlParsed.Path = strings.TrimPrefix(urlParsed.Path, apiRouterVersion)
		trimmedURI = urlParsed.String()
	}
	return trimmedURI
}

// CreateFilterableLandingPage creates a filterable dataset landing page based on api model responses
func CreateFilterableLandingPage(basePage coreModel.Page, ctx context.Context, req *http.Request, d dataset.DatasetDetails, ver dataset.Version, datasetID string, opts []dataset.Options, dims dataset.VersionDimensions, displayOtherVersionsLink bool, breadcrumbs []zebedee.Breadcrumb, latestVersionNumber int, latestVersionURL, lang, apiRouterVersion string, maxNumOpts int, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) datasetLandingPageFilterable.Page {
	p := datasetLandingPageFilterable.Page{
		Page: basePage,
	}
	MapCookiePreferences(req, &p.Page.CookiesPreferencesSet, &p.Page.CookiesPolicy)
	p.Type = "dataset_landing_page"
	p.Metadata.Title = d.Title
	p.Language = lang
	p.URI = req.URL.Path
	p.DatasetLandingPage.UnitOfMeasurement = d.UnitOfMeasure
	p.Metadata.Description = d.Description
	p.Metadata.Survey = d.Description.Survey
	p.DatasetId = datasetID
	p.ReleaseDate = ver.ReleaseDate
	p.BetaBannerEnabled = true
	p.FeatureFlags.SixteensVersion = SixteensVersion

	p.ServiceMessage = serviceMessage
	p.EmergencyBanner = mapEmergencyBanner(emergencyBannerContent)

	if d.Type == "nomis" {
		p.DatasetLandingPage.NomisReferenceURL = d.NomisReferenceURL
		homeBreadcrumb := coreModel.TaxonomyNode{
			Title: "Home",
			URI:   "/",
		}
		p.Breadcrumb = append(p.Breadcrumb, homeBreadcrumb)

	}

	p.HasJSONLD = true

	// Trim API version path prefix from breadcrumb URIs, if present.
	for _, breadcrumb := range breadcrumbs {
		p.Page.Breadcrumb = append(p.Page.Breadcrumb, coreModel.TaxonomyNode{
			Title: breadcrumb.Description.Title,
			URI:   getTrimmedBreadcrumbURI(ctx, breadcrumb, apiRouterVersion),
		})
	}

	// breadcrumbs won't contain this page or it's parent page in it's response
	// from Zebedee, so add it to the slice
	currentPageBreadcrumbTitle := ver.Links.Edition.ID
	if currentPageBreadcrumbTitle == "time-series" {
		currentPageBreadcrumbTitle = "Current"
	}
	datasetURL, err := url.Parse(d.Links.Self.URL)
	if err != nil {
		log.Warn(ctx, "failed to parse url, self link", log.FormatErrors([]error{err}))
	}
	datasetPath := strings.TrimPrefix(datasetURL.Path, apiRouterVersion)
	datasetBreadcrumbs := []coreModel.TaxonomyNode{
		{
			Title: d.Title,
			URI:   datasetPath,
		},
		{
			Title: currentPageBreadcrumbTitle,
		},
	}
	p.Breadcrumb = append(p.Breadcrumb, datasetBreadcrumbs...)

	if d.Contacts != nil && len(*d.Contacts) > 0 {
		contacts := *d.Contacts
		p.ContactDetails.Name = contacts[0].Name
		p.ContactDetails.Telephone = contacts[0].Telephone
		p.ContactDetails.Email = contacts[0].Email
	}

	p.DatasetLandingPage.DatasetLandingPage.NextRelease = d.NextRelease
	p.DatasetLandingPage.DatasetID = datasetID

	p.DatasetLandingPage.Edition = ver.Edition

	if ver.Edition != "time-series" {
		p.DatasetLandingPage.ShowEditionName = true
	}

	p.DatasetLandingPage.IsLatest = d.Links.LatestVersion.URL == ver.Links.Self.URL
	p.DatasetLandingPage.LatestVersionURL = latestVersionURL
	p.DatasetLandingPage.IsLatestVersionOfEdition = latestVersionNumber == ver.Version
	p.DatasetLandingPage.QMIURL = d.QMI.URL
	p.DatasetLandingPage.IsNationalStatistic = d.NationalStatistic
	p.DatasetLandingPage.ReleaseFrequency = strings.Title(d.ReleaseFrequency)
	p.DatasetLandingPage.Citation = d.License

	if d.Type == "nomis" {
		if ver.UsageNotes != nil {
			for _, usageNote := range *ver.UsageNotes {
				p.DatasetLandingPage.UsageNotes = append(p.DatasetLandingPage.UsageNotes, datasetLandingPageFilterable.UsageNote{
					Title: usageNote.Title,
					Note:  usageNote.Note,
				})
			}
		}
	}
	if d.Methodologies != nil {
		for _, meth := range *d.Methodologies {
			p.DatasetLandingPage.Methodologies = append(p.DatasetLandingPage.Methodologies, datasetLandingPageFilterable.Methodology{
				Title:       meth.Title,
				URL:         meth.URL,
				Description: meth.Description,
			})
		}
	}

	if d.Publications != nil {
		for _, pub := range *d.Publications {
			p.DatasetLandingPage.Publications = append(p.DatasetLandingPage.Publications, datasetLandingPageFilterable.Publication{
				Title: pub.Title,
				URL:   pub.URL,
			})
		}
	}

	if d.RelatedDatasets != nil {
		for _, link := range *d.RelatedDatasets {
			p.DatasetLandingPage.RelatedLinks = append(p.DatasetLandingPage.RelatedLinks, datasetLandingPageFilterable.Publication{
				Title: link.Title,
				URL:   link.URL,
			})
		}
	}

	for _, changes := range ver.LatestChanges {
		p.DatasetLandingPage.LatestChanges = append(p.DatasetLandingPage.LatestChanges, datasetLandingPageFilterable.Change{
			Name:        changes.Name,
			Description: changes.Description,
		})
	}

	var v sharedModel.Version
	v.Title = d.Title
	v.Description = d.Description
	v.Edition = ver.Edition
	v.Version = strconv.Itoa(ver.Version)

	p.DatasetLandingPage.HasOlderVersions = displayOtherVersionsLink

	for k, download := range ver.Downloads {
		if len(download.URL) > 0 {
			v.Downloads = append(v.Downloads, sharedModel.Download{
				Extension: k,
				Size:      download.Size,
				URI:       download.URL,
			})
		}
	}

	p.DatasetLandingPage.Version = v

	if len(opts) > 0 {
		p.DatasetLandingPage.Dimensions = mapOptionsToDimensions(ctx, d.Type, dims.Items, opts, d.Links.LatestVersion.URL, maxNumOpts)
	}

	return p
}

// CreateVersionsList creates a versions list page based on api model responses
func CreateVersionsList(basePage coreModel.Page, req *http.Request, d dataset.DatasetDetails, edition dataset.Edition, versions []dataset.Version, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) datasetVersionsList.Page {
	p := datasetVersionsList.Page{
		Page: basePage,
	}
	MapCookiePreferences(req, &p.Page.CookiesPreferencesSet, &p.Page.CookiesPolicy)
	// TODO refactor and make Welsh compatible.
	p.Metadata.Title = "All versions of " + d.Title
	if len(versions) > 0 {
		p.Metadata.Title += " " + versions[0].Edition
	}
	p.Metadata.Title += " dataset"
	p.BetaBannerEnabled = true

	p.Data.LatestVersionURL = helpers.DatasetVersionUrl(d.ID, edition.Edition, edition.Links.LatestVersion.ID)
	p.DatasetId = d.ID
	p.URI = req.URL.Path
	p.FeatureFlags.SixteensVersion = SixteensVersion

	p.ServiceMessage = serviceMessage
	p.EmergencyBanner = mapEmergencyBanner(emergencyBannerContent)

	latestVersionNumber := 1
	for _, ver := range versions {
		var version sharedModel.Version
		version.IsLatest = false
		version.VersionNumber = ver.Version
		version.Title = d.Title
		version.Date = ver.ReleaseDate
		versionUrl := helpers.DatasetVersionUrl(ver.Links.Dataset.ID, ver.Edition, strconv.Itoa(ver.Version))
		version.VersionURL = versionUrl
		version.FilterURL = versionUrl + "/filter"

		// Not the 'created' first version and more than one stored version
		if ver.Version > 1 && len(p.Data.Versions) >= 1 {
			previousVersion := p.Data.Versions[len(p.Data.Versions)-1].VersionNumber
			version.Superseded = helpers.DatasetVersionUrl(ver.Links.Dataset.ID, ver.Edition, strconv.Itoa(previousVersion))
		}

		if ver.Version > latestVersionNumber {
			latestVersionNumber = ver.Version
		}

		for ext, download := range ver.Downloads {
			version.Downloads = append(version.Downloads, sharedModel.Download{
				Extension: ext,
				Size:      download.Size,
				URI:       download.URL,
			})
		}

		mapCorrectionAlert(&ver, &version)

		p.Data.Versions = append(p.Data.Versions, version)
	}

	for i, ver := range p.Data.Versions {
		if ver.VersionNumber == latestVersionNumber {
			p.Data.Versions[i].IsLatest = true
			break
		}
	}

	sort.Slice(p.Data.Versions, func(i, j int) bool { return p.Data.Versions[i].VersionNumber > p.Data.Versions[j].VersionNumber })
	return p
}

// CreateEditionsList creates a editions list page based on api model responses
func CreateEditionsList(basePage coreModel.Page, ctx context.Context, req *http.Request, d dataset.DatasetDetails, editions []dataset.Edition, datasetID string, breadcrumbs []zebedee.Breadcrumb, lang, apiRouterVersion string, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) datasetEditionsList.Page {
	p := datasetEditionsList.Page{
		Page: basePage,
	}
	MapCookiePreferences(req, &p.Page.CookiesPreferencesSet, &p.Page.CookiesPolicy)
	p.Type = "dataset_edition_list"
	p.Language = lang
	p.Metadata.Title = d.Title
	p.URI = req.URL.Path
	p.Metadata.Description = d.Description
	p.DatasetId = datasetID
	p.BetaBannerEnabled = true
	p.FeatureFlags.SixteensVersion = SixteensVersion

	p.ServiceMessage = serviceMessage
	p.EmergencyBanner = mapEmergencyBanner(emergencyBannerContent)

	for _, bc := range breadcrumbs {
		p.Breadcrumb = append(p.Breadcrumb, coreModel.TaxonomyNode{
			Title: bc.Description.Title,
			URI:   getTrimmedBreadcrumbURI(ctx, bc, apiRouterVersion),
		})
	}

	// breadcrumbs won't contain this page in it's response from Zebedee, so add it to the slice
	p.Breadcrumb = append(p.Breadcrumb, coreModel.TaxonomyNode{
		Title: d.Title,
	})

	if d.Contacts != nil && len(*d.Contacts) > 0 {
		contacts := *d.Contacts
		p.ContactDetails.Name = contacts[0].Name
		p.ContactDetails.Telephone = contacts[0].Telephone
		p.ContactDetails.Email = contacts[0].Email
	}

	p.DatasetLandingPage.DatasetLandingPage.NextRelease = d.NextRelease
	p.DatasetLandingPage.DatasetID = datasetID

	if editions != nil && len(editions) > 0 {
		for _, edition := range editions {
			var e datasetEditionsList.Edition
			e.Title = edition.Edition
			e.LatestVersionURL = helpers.DatasetVersionUrl(datasetID, edition.Edition, edition.Links.LatestVersion.ID)
			p.Editions = append(p.Editions, e)
		}
	}

	return p
}

// CreateCensusDatasetLandingPage creates a census-landing page based on api model responses
func CreateCensusDatasetLandingPage(ctx context.Context, req *http.Request, basePage coreModel.Page, d dataset.DatasetDetails, version dataset.Version, opts []dataset.Options, initialVersionReleaseDate string, hasOtherVersions bool, allVersions []dataset.Version, latestVersionNumber int, latestVersionURL, lang string, maxNumberOfOptions int, isValidationError, hasFilterOutput bool, filter filter.Model) datasetLandingPageCensus.Page {
	p := datasetLandingPageCensus.Page{
		Page: basePage,
	}

	MapCookiePreferences(req, &p.Page.CookiesPreferencesSet, &p.Page.CookiesPolicy)

	p.Type = d.Type
	p.Language = lang
	p.URI = req.URL.Path
	p.ID = d.ID

	p.Version.ReleaseDate = version.ReleaseDate
	if initialVersionReleaseDate == "" {
		p.InitialReleaseDate = p.Version.ReleaseDate
	} else {
		p.InitialReleaseDate = initialVersionReleaseDate
	}

	p.DatasetLandingPage.HasOtherVersions = hasOtherVersions

	p.Metadata.Title = d.Title
	p.Metadata.Description = d.Description

	if hasFilterOutput {
		for ext, download := range filter.Downloads {
			p.Version.Downloads = append(p.Version.Downloads, sharedModel.Download{
				Extension: strings.ToLower(ext),
				Size:      download.Size,
				URI:       download.URL,
			})
		}
	} else {
		for ext, download := range version.Downloads {
			p.Version.Downloads = append(p.Version.Downloads, sharedModel.Download{
				Extension: strings.ToLower(ext),
				Size:      download.Size,
				URI:       download.URL,
			})
		}
	}

	if d.Contacts != nil && len(*d.Contacts) > 0 {
		contacts := *d.Contacts
		if contacts[0].Telephone != "" {
			p.ContactDetails.Telephone = contacts[0].Telephone
			p.HasContactDetails = true
		}
		if contacts[0].Email != "" {
			p.ContactDetails.Email = contacts[0].Email
			p.HasContactDetails = true
		}
	}

	p.DatasetLandingPage.Description = strings.Split(d.Description, "\n")

	var collapsibleContentItems []coreModel.CollapsibleItem

	for _, dims := range version.Dimensions {
		if dims.Description != "" {
			var collapsibleContent coreModel.CollapsibleItem
			collapsibleContent.Subheading = dims.Name
			collapsibleContent.Content = strings.Split(dims.Description, "\n")
			collapsibleContentItems = append(collapsibleContentItems, collapsibleContent)
		}
	}

	if len(collapsibleContentItems) > 0 {
		p.Collapsible = coreModel.Collapsible{
			Title: coreModel.Localisation{
				LocaleKey: "VariablesExplanation",
				Plural:    4,
			},
			CollapsibleItems: collapsibleContentItems,
		}
	}

	hasMethodologies := false
	if d.Methodologies != nil {
		for _, meth := range *d.Methodologies {
			p.DatasetLandingPage.Methodologies = append(p.DatasetLandingPage.Methodologies, datasetLandingPageCensus.Methodology{
				Title:       meth.Title,
				URL:         meth.URL,
				Description: meth.Description,
			})
		}
		hasMethodologies = true
	}

	p.Breadcrumb = []coreModel.TaxonomyNode{
		{
			Title: "Home",
			URI:   "/",
		},
		{
			Title: "Census",
			URI:   "/census",
		},
		{
			Title: "Datasets",
			URI:   "/census/datasets",
		},
	}

	sections := make(map[string]coreModel.ContentSection)
	displayOrder := make([]string, 0)

	p.TableOfContents = coreModel.TableOfContents{
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

	if len(version.Downloads) > 0 && !hasFilterOutput {
		p.DatasetLandingPage.HasDownloads = true
	}

	if hasFilterOutput && len(filter.Downloads) > 0 {
		p.DatasetLandingPage.HasDownloads = true
	}

	if p.HasContactDetails {
		sections["contact"] = coreModel.ContentSection{
			Title: coreModel.Localisation{
				LocaleKey: "ContactDetails",
				Plural:    1,
			},
		}
		displayOrder = append(displayOrder, "contact")
	}

	sections["stats-disclosure"] = coreModel.ContentSection{
		Title: coreModel.Localisation{
			LocaleKey: "StatisticalDisclosureControl",
			Plural:    1,
		},
	}
	displayOrder = append(displayOrder, "stats-disclosure")

	if hasMethodologies {
		sections["methodology"] = coreModel.ContentSection{
			Title: coreModel.Localisation{
				LocaleKey: "Methodology",
				Plural:    1,
			},
		}
		displayOrder = append(displayOrder, "methodology")
	}

	if hasOtherVersions {
		sections["version-history"] = coreModel.ContentSection{
			Title: coreModel.Localisation{
				LocaleKey: "VersionHistory",
				Plural:    1,
			},
		}
		displayOrder = append(displayOrder, "version-history")

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

		p.DatasetLandingPage.ShowOtherVersionsPanel = latestVersionNumber != version.Version && hasOtherVersions
		p.DatasetLandingPage.LatestVersionURL = latestVersionURL
	}

	p.TableOfContents.Sections = sections
	p.TableOfContents.DisplayOrder = displayOrder

	p.DatasetLandingPage.ShareDetails.Language = lang
	currentUrl := helpers.GetCurrentUrl(lang, p.SiteDomain, req.URL.Path)

	p.DatasetLandingPage.ShareDetails.ShareLocations = []datasetLandingPageCensus.Share{
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

	p.BetaBannerEnabled = true

	if len(opts) > 0 && !hasFilterOutput {
		p.DatasetLandingPage.Dimensions = mapOptionsToDimensions(ctx, d.Type, version.Dimensions, opts, d.Links.LatestVersion.URL, maxNumberOfOptions)
	}

	if hasFilterOutput {
		for _, dim := range filter.Dimensions {
			p.DatasetLandingPage.Dimensions = append(p.DatasetLandingPage.Dimensions, sharedModel.Dimension{
				Title:      dim.Label,
				Values:     dim.Options,
				IsAreaType: helpers.IsBoolPtr(dim.IsAreaType),
				TotalItems: len(dim.Options),
			})
		}
	}

	if isValidationError {
		p.Error.Title = fmt.Sprintf("Error: %s", d.Title)
	}

	if strings.Contains(d.Type, "flex") {
		p.DatasetLandingPage.IsFlexible = true
		p.DatasetLandingPage.FormAction = fmt.Sprintf("/datasets/%s/editions/%s/versions/%s/filter-flex", d.ID, version.Edition, strconv.Itoa(version.Version))
	}

	p.BackTo = coreModel.BackTo{
		Text: coreModel.Localisation{
			LocaleKey: "BackToContents",
			Plural:    4,
		},
		AnchorFragment: "toc",
	}

	return p
}

func mapCorrectionAlert(ver *dataset.Version, version *sharedModel.Version) {
	if ver.Alerts != nil {
		for _, alert := range *ver.Alerts {
			if alert.Type == CorrectionAlertType {
				version.Corrections = append(version.Corrections, sharedModel.Correction{
					Reason: alert.Description,
					Date:   alert.Date,
				})
			}
		}
	}
}

func mapOptionsToDimensions(ctx context.Context, datasetType string, dims []dataset.VersionDimension, opts []dataset.Options, latestVersionURL string, maxNumberOfOptions int) []sharedModel.Dimension {
	dimensions := []sharedModel.Dimension{}
	for _, opt := range opts {

		var pDim sharedModel.Dimension

		var title string
		if len(opt.Items) > 0 {
			title = strings.Title(opt.Items[0].DimensionID)
		}

		if datasetType != "nomis" {
			pDim.Title = title
			versionURL, err := url.Parse(latestVersionURL)
			if err != nil {
				log.Warn(ctx, "failed to parse url, last_version link", log.FormatErrors([]error{err}))
			}
			for _, dimension := range dims {
				if dimension.Name == opt.Items[0].DimensionID {
					pDim.Name = dimension.Name
					pDim.Description = dimension.Description
					pDim.IsAreaType = helpers.IsBoolPtr(dimension.IsAreaType)
					if len(dimension.Label) > 0 {
						pDim.Title = dimension.Label
					}
				}
			}

			pDim.OptionsURL = fmt.Sprintf("%s/dimensions/%s/options", versionURL.Path, opt.Items[0].DimensionID)
			pDim.TotalItems = opt.TotalCount

			if _, err = time.Parse("Jan-06", opt.Items[0].Label); err == nil {
				var ts TimeSlice
				for _, val := range opt.Items {
					t, err := convertMMMYYToTime(val.Label)
					if err != nil {
						log.Warn(ctx, "unable to convert date (MMYY) to time", log.FormatErrors([]error{err}), log.Data{"label": val.Label})
					}
					ts = append(ts, t)
				}
				sort.Sort(ts)

				if len(ts) > 0 {
					startDate := ts[0]

					for i, t := range ts {
						if i != len(ts)-1 {
							if ((ts[i+1].Month() - t.Month()) == 1) || (t.Month() == 12 && ts[i+1].Month() == 1) {
								continue
							}
							if startDate.Year() == t.Year() && startDate.Month().String() == t.Month().String() {
								pDim.Values = append(pDim.Values, fmt.Sprintf("This year %d contains data for the month %s", startDate.Year(), startDate.Month().String()))
							} else {
								pDim.Values = append(pDim.Values, fmt.Sprintf("All months between %s %d and %s %d", startDate.Month().String(), startDate.Year(), t.Month().String(), t.Year()))
							}
							startDate = ts[i+1]
						} else {
							if startDate.Year() == t.Year() && startDate.Month().String() == t.Month().String() {
								pDim.Values = append(pDim.Values, fmt.Sprintf("This year %d contains data for the month %s", startDate.Year(), startDate.Month().String()))
							} else {
								pDim.Values = append(pDim.Values, fmt.Sprintf("All months between %s %d and %s %d", startDate.Month().String(), startDate.Year(), t.Month().String(), t.Year()))
							}
						}
					}
				}
			} else if _, err = time.Parse("2006", opt.Items[0].Label); err == nil {
				var ts TimeSlice
				for _, val := range opt.Items {
					t, err := convertYYYYToTime(val.Label)
					if err != nil {
						log.Warn(ctx, "unable to convert date (YYYY) to time", log.FormatErrors([]error{err}), log.Data{"label": val.Label})
					}
					ts = append(ts, t)
				}
				sort.Sort(ts)

				if len(ts) > 0 {
					startDate := ts[0]

					for i, t := range ts {
						if i != len(ts)-1 {
							if (ts[i+1].Year() - t.Year()) == 1 {
								continue
							}
							if startDate.Year() == t.Year() {
								pDim.Values = append(pDim.Values, fmt.Sprintf("This year contains data for %d", startDate.Year()))
							} else {
								pDim.Values = append(pDim.Values, fmt.Sprintf("All years between %d and %d", startDate.Year(), t.Year()))
							}
							startDate = ts[i+1]
						} else {

							if startDate.Year() == t.Year() {
								pDim.Values = append(pDim.Values, fmt.Sprintf("This year contains data for %d", startDate.Year()))
							} else {
								pDim.Values = append(pDim.Values, fmt.Sprintf("All years between %d and %d", startDate.Year(), t.Year()))
							}
						}
					}
				}
			} else {
				for i, val := range opt.Items {
					if opt.TotalCount > maxNumberOfOptions {
						if i > 9 {
							break
						}
					}
					pDim.Values = append(pDim.Values, val.Label)
				}

				if opt.Items[0].DimensionID == DimensionTime || opt.Items[0].DimensionID == DimensionAge {
					isValid := true
					var intVals []int
					for _, val := range pDim.Values {
						intVal, err := strconv.Atoi(val)
						if err != nil {
							isValid = false
							break
						}
						intVals = append(intVals, intVal)
					}

					if isValid {
						sort.Ints(intVals)
						for i, val := range intVals {
							pDim.Values[i] = strconv.Itoa(val)
						}
					}
				}
			}
		}
		dimensions = append(dimensions, pDim)
	}
	return dimensions
}

func convertMMMYYToTime(input string) (t time.Time, err error) {
	return time.Parse("Jan-06", input)
}

func convertYYYYToTime(input string) (t time.Time, err error) {
	return time.Parse("2006", input)
}

// MapCookiePreferences reads cookie policy and preferences cookies and then maps the values to the page model
func MapCookiePreferences(req *http.Request, preferencesIsSet *bool, policy *coreModel.CookiesPolicy) {
	preferencesCookie := cookies.GetCookiePreferences(req)
	*preferencesIsSet = preferencesCookie.IsPreferenceSet
	*policy = coreModel.CookiesPolicy{
		Essential: preferencesCookie.Policy.Essential,
		Usage:     preferencesCookie.Policy.Usage,
	}
}

func mapEmergencyBanner(bannerData zebedee.EmergencyBanner) coreModel.EmergencyBanner {
	var mappedEmergencyBanner coreModel.EmergencyBanner
	emptyBannerObj := zebedee.EmergencyBanner{}
	if bannerData != emptyBannerObj {
		mappedEmergencyBanner.Title = bannerData.Title
		mappedEmergencyBanner.Type = strings.Replace(bannerData.Type, "_", "-", -1)
		mappedEmergencyBanner.Description = bannerData.Description
		mappedEmergencyBanner.URI = bannerData.URI
		mappedEmergencyBanner.LinkText = bannerData.LinkText
	}
	return mappedEmergencyBanner
}

func FindVersion(versionList []zebedee.Dataset, versionURI string) zebedee.Dataset {
	for _, ver := range versionList {
		if versionURI == ver.URI {
			return ver
		}
	}
	return zebedee.Dataset{}
}

func MapDownloads(downloadsList []zebedee.Download, versionURI string) []datasetPage.Download {
	var dl []datasetPage.Download
	for _, d := range downloadsList {
		dl = append(dl, datasetPage.Download{
			Extension:   filepath.Ext(d.File),
			Size:        d.Size,
			URI:         versionURI + "/" + d.File,
			DownloadUrl: determineDownloadUrl(d, versionURI),
		})
	}
	return dl
}
