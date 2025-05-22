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

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	datasetMdl "github.com/ONSdigital/dp-frontend-dataset-controller/model/dataset"
	filterable "github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetLandingPageFilterable"
	edition "github.com/ONSdigital/dp-frontend-dataset-controller/model/editions"
	dpTopicApiModels "github.com/ONSdigital/dp-topic-api/models"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-cookies/cookies"
	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"

	"github.com/ONSdigital/dp-frontend-dataset-controller/model/version"
	dpRendererModel "github.com/ONSdigital/dp-renderer/v2/model"

	"github.com/ONSdigital/log.go/v2/log"
)

// TimeSlice allows sorting of a list of time.Time
type TimeSlice []time.Time

// Constants names
const (
	DimensionTime      = "time"
	DimensionAge       = "age"
	DimensionGeography = "geography"
	SixteensVersion    = "2c5867a"
)

var (
	cfg, _ = config.Get()
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
//
//nolint:gocyclo //complexity 21
func CreateFilterableLandingPage(ctx context.Context, basePage dpRendererModel.Page, d dpDatasetApiModels.Dataset, ver dpDatasetApiModels.Version,
	datasetID string, opts []dpDatasetApiSdk.VersionDimensionOptionsList, dims dpDatasetApiSdk.VersionDimensionsList, displayOtherVersionsLink bool,
	breadcrumbs []zebedee.Breadcrumb, latestVersionNumber int, latestVersionURL, apiRouterVersion string, maxNumOpts int) filterable.Page {
	// Set default values to be used if fields are null pointers
	datasetSelfHRef := ""
	isLatest := false
	isNationalStatistic := false
	QMIURL := ""

	p := filterable.Page{
		Page: basePage,
	}
	p.Type = "dataset_landing_page"
	p.DatasetLandingPage.UnitOfMeasurement = d.UnitOfMeasure
	p.ReleaseDate = ver.ReleaseDate
	p.FeatureFlags.SixteensVersion = SixteensVersion

	if d.Type == "nomis" {
		p.DatasetLandingPage.NomisReferenceURL = ""
		homeBreadcrumb := dpRendererModel.TaxonomyNode{
			Title: "Home",
			URI:   "/",
		}
		p.Breadcrumb = append(p.Breadcrumb, homeBreadcrumb)
	}

	p.HasJSONLD = true

	// Trim API version path prefix from breadcrumb URIs, if present.
	for _, breadcrumb := range breadcrumbs {
		p.Page.Breadcrumb = append(p.Page.Breadcrumb, dpRendererModel.TaxonomyNode{
			Title: breadcrumb.Description.Title,
			URI:   getTrimmedBreadcrumbURI(ctx, breadcrumb, apiRouterVersion),
		})
	}

	// breadcrumbs won't contain this page or it's parent page in it's response
	// from Zebedee, so add it to the slice
	currentPageBreadcrumbTitle := ""
	// Update the breadcrumb title if edition link is available
	if ver.Links.Edition != nil {
		currentPageBreadcrumbTitle = ver.Links.Edition.ID
	}

	if currentPageBreadcrumbTitle == "time-series" {
		currentPageBreadcrumbTitle = "Current"
	}
	if d.Links.Self != nil {
		datasetSelfHRef = d.Links.Self.HRef
	}

	datasetURL, err := url.Parse(datasetSelfHRef)
	if err != nil {
		log.Warn(ctx, "failed to parse url, self link", log.FormatErrors([]error{err}))
	}
	datasetPath := strings.TrimPrefix(datasetURL.Path, apiRouterVersion)
	datasetBreadcrumbs := []dpRendererModel.TaxonomyNode{
		{
			Title: d.Title,
			URI:   datasetPath,
		},
		{
			Title: currentPageBreadcrumbTitle,
		},
	}
	p.Breadcrumb = append(p.Breadcrumb, datasetBreadcrumbs...)

	if len(d.Contacts) > 0 {
		contacts := d.Contacts
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

	// Update LatestVersion boolean if not null pointer
	if d.Links.LatestVersion != nil {
		if d.Links.LatestVersion.HRef == ver.Links.Self.HRef {
			isLatest = true
		}
	}
	p.DatasetLandingPage.IsLatest = isLatest
	p.DatasetLandingPage.LatestVersionURL = latestVersionURL
	p.DatasetLandingPage.IsLatestVersionOfEdition = latestVersionNumber == ver.Version

	// Update QMIURL if not null pointer
	if d.QMI != nil {
		QMIURL = d.QMI.HRef
	}
	p.DatasetLandingPage.QMIURL = QMIURL

	// Update IsNationalStatistic if not null pointer
	if d.NationalStatistic != nil {
		isNationalStatistic = *d.NationalStatistic
	}
	p.DatasetLandingPage.IsNationalStatistic = isNationalStatistic
	p.DatasetLandingPage.ReleaseFrequency = cases.Title(language.English).String(d.ReleaseFrequency)
	p.DatasetLandingPage.Citation = d.License

	if d.Type == "nomis" {
		if ver.UsageNotes != nil {
			for _, usageNote := range *ver.UsageNotes {
				p.DatasetLandingPage.UsageNotes = append(p.DatasetLandingPage.UsageNotes, filterable.UsageNote{
					Title: usageNote.Title,
					Note:  usageNote.Note,
				})
			}
		}
	}
	if d.Methodologies != nil {
		for _, meth := range d.Methodologies {
			p.DatasetLandingPage.Methodologies = append(p.DatasetLandingPage.Methodologies, filterable.Methodology{
				Title:       meth.Title,
				URL:         meth.HRef,
				Description: meth.Description,
			})
		}
	}

	if d.Publications != nil {
		for _, pub := range d.Publications {
			p.DatasetLandingPage.Publications = append(p.DatasetLandingPage.Publications, filterable.Publication{
				Title: pub.Title,
				URL:   pub.HRef,
			})
		}
	}

	if d.RelatedDatasets != nil {
		for _, link := range d.RelatedDatasets {
			p.DatasetLandingPage.RelatedLinks = append(p.DatasetLandingPage.RelatedLinks, filterable.Publication{
				Title: link.Title,
				URL:   link.HRef,
			})
		}
	}

	if ver.LatestChanges != nil {
		for _, change := range *ver.LatestChanges {
			p.DatasetLandingPage.LatestChanges = append(p.DatasetLandingPage.LatestChanges, filterable.Change{
				Name:        change.Name,
				Description: change.Description,
			})
		}
	}

	var v sharedModel.Version
	v.Title = d.Title
	v.Description = d.Description
	v.Edition = ver.Edition
	v.Version = strconv.Itoa(ver.Version)

	p.DatasetLandingPage.HasOlderVersions = displayOtherVersionsLink

	if ver.Downloads != nil {
		helpers.MapVersionDownloads(&v, ver.Downloads)
	}

	p.DatasetLandingPage.Version = v

	if len(opts) > 0 {
		p.DatasetLandingPage.Dimensions = mapOptionsToDimensions(ctx, d.Type, dims, opts, latestVersionURL, maxNumOpts)
	}

	return p
}

// CreateVersionsList creates a versions list page based on api model responses
func CreateVersionsList(basePage dpRendererModel.Page, req *http.Request, d dataset.DatasetDetails, ed dataset.Edition, versions []dataset.Version, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) version.Page {
	p := version.Page{
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

	p.Data.LatestVersionURL = helpers.DatasetVersionURL(d.ID, ed.Edition, ed.Links.LatestVersion.ID)
	p.DatasetId = d.ID
	p.URI = req.URL.Path
	p.FeatureFlags.SixteensVersion = SixteensVersion

	p.ServiceMessage = serviceMessage
	p.EmergencyBanner = mapEmergencyBanner(emergencyBannerContent)

	p.FeatureFlags.FeedbackAPIURL = cfg.FeedbackAPIURL

	latestVersionNumber := 1
	for i := range versions {
		var v sharedModel.Version
		v.IsLatest = false
		v.VersionNumber = versions[i].Version
		v.Title = d.Title
		v.Date = versions[i].ReleaseDate
		versionURL := helpers.DatasetVersionURL(versions[i].Links.Dataset.ID, versions[i].Edition, strconv.Itoa(versions[i].Version))
		v.VersionURL = versionURL
		v.FilterURL = versionURL + "/filter"

		// Not the 'created' first version and more than one stored version
		if versions[i].Version > 1 && len(p.Data.Versions) >= 1 {
			previousVersion := p.Data.Versions[len(p.Data.Versions)-1].VersionNumber
			v.Superseded = helpers.DatasetVersionURL(versions[i].Links.Dataset.ID, versions[i].Edition, strconv.Itoa(previousVersion))
		}

		if versions[i].Version > latestVersionNumber {
			latestVersionNumber = versions[i].Version
		}

		for ext, download := range versions[i].Downloads {
			v.Downloads = append(v.Downloads, sharedModel.Download{
				Extension: ext,
				Size:      download.Size,
				URI:       download.URL,
			})
		}

		// Map the correction alerts
		if versions[i].Alerts != nil {
			for _, alert := range *versions[i].Alerts {
				if alert.Type == CorrectionAlertType {
					v.Corrections = append(v.Corrections, sharedModel.Correction{
						Reason: alert.Description,
						Date:   alert.Date,
					})
				}
			}
		}

		p.Data.Versions = append(p.Data.Versions, v)
	}

	for i := range p.Data.Versions {
		if p.Data.Versions[i].VersionNumber == latestVersionNumber {
			p.Data.Versions[i].IsLatest = true
			break
		}
	}

	sort.Slice(p.Data.Versions, func(i, j int) bool { return p.Data.Versions[i].VersionNumber > p.Data.Versions[j].VersionNumber })
	return p
}

// CreateEditionsList creates an editions list page based on api model responses
func CreateEditionsList(ctx context.Context, basePage dpRendererModel.Page, req *http.Request, d dpDatasetApiModels.Dataset, editions dpDatasetApiSdk.EditionsList, datasetID string, breadcrumbs []zebedee.Breadcrumb, apiRouterVersion string) edition.Page {
	p := edition.Page{
		Page: basePage,
	}
	MapCookiePreferences(req, &p.Page.CookiesPreferencesSet, &p.Page.CookiesPolicy)
	p.FeatureFlags.SixteensVersion = SixteensVersion

	for _, bc := range breadcrumbs {
		p.Breadcrumb = append(p.Breadcrumb, dpRendererModel.TaxonomyNode{
			Title: bc.Description.Title,
			URI:   getTrimmedBreadcrumbURI(ctx, bc, apiRouterVersion),
		})
	}

	// breadcrumbs won't contain this page in it's response from Zebedee, so add it to the slice
	p.Breadcrumb = append(p.Breadcrumb, dpRendererModel.TaxonomyNode{
		Title: d.Title,
	})

	if len(d.Contacts) > 0 {
		contacts := d.Contacts
		p.ContactDetails.Name = contacts[0].Name
		p.ContactDetails.Telephone = contacts[0].Telephone
		p.ContactDetails.Email = contacts[0].Email
	}

	p.DatasetLandingPage.DatasetLandingPage.NextRelease = d.NextRelease
	p.DatasetLandingPage.DatasetID = datasetID

	// Get editions list
	editionItems := editions.Items
	if len(editionItems) > 0 {
		for i := range editionItems {
			var el edition.List
			el.Title = editionItems[i].Edition
			el.LatestVersionURL = helpers.DatasetVersionURL(datasetID, editionItems[i].Edition, editionItems[i].Links.LatestVersion.ID)
			p.Editions = append(p.Editions, el)
		}
	}

	// Prepares table of components object for use in dp-renderer
	p.TableOfContents = buildEditionsListTableOfContents(d)

	return p
}

// CreateEditionsListForStaticDatasetType creates an editions list page when dataset type is static, based on api model responses
func CreateEditionsListForStaticDatasetType(ctx context.Context, basePage dpRendererModel.Page, req *http.Request, d dpDatasetApiModels.Dataset, editions dpDatasetApiSdk.EditionsList, datasetID string, apiRouterVersion string, topicObjectList []dpTopicApiModels.Topic) edition.Page {
	p := edition.Page{
		Page: basePage,
	}
	MapCookiePreferences(req, &p.Page.CookiesPreferencesSet, &p.Page.CookiesPolicy)

	// Unset SixteensVersion when dp-design-system is needed
	p.FeatureFlags.SixteensVersion = ""

	if len(d.Contacts) > 0 {
		contacts := d.Contacts
		p.ContactDetails.Name = contacts[0].Name
		p.ContactDetails.Telephone = contacts[0].Telephone
		p.ContactDetails.Email = contacts[0].Email
	}

	p.DatasetLandingPage.DatasetLandingPage.NextRelease = d.NextRelease
	p.DatasetLandingPage.DatasetID = datasetID

	// BREADCRUMBS
	baseURL := "https://www.ons.gov.uk/"

	p.Breadcrumb = CreateBreadcrumbsFromTopicList(baseURL, topicObjectList)

	// Get editions list
	editionItems := editions.Items
	if len(editionItems) > 0 {
		for i := range editionItems {
			var el edition.List
			el.Title = editionItems[i].Edition
			el.LatestVersionURL = helpers.DatasetVersionURL(datasetID, editionItems[i].Edition, editionItems[i].Links.LatestVersion.ID)
			p.Editions = append(p.Editions, el)
		}
	}

	// Prepares table of components object for use in dp-renderer
	p.TableOfContents = buildEditionsListTableOfContents(d)

	return p
}


func mapCorrectionAlert(ver *dpDatasetApiModels.Version, model *sharedModel.Version) {
	if ver.Alerts != nil {
		for _, alert := range *ver.Alerts {
			if alert.Type == CorrectionAlertType {
				model.Corrections = append(model.Corrections, sharedModel.Correction{
					Reason: alert.Description,
					Date:   alert.Date,
				})
			}
		}
	}
}

//nolint:all // legacy code with poor test coverage
func mapOptionsToDimensions(ctx context.Context, datasetType string, dims dpDatasetApiSdk.VersionDimensionsList, opts []dpDatasetApiSdk.VersionDimensionOptionsList, latestVersionURL string, maxNumberOfOptions int) []sharedModel.Dimension {
	dimensions := []sharedModel.Dimension{}
	for _, opt := range opts {
		var pDim sharedModel.Dimension
		totalCount := len(opt.Items)

		var title string
		if totalCount > 0 {
			title = cases.Title(language.English).String(opt.Items[0].Name)
		}

		if datasetType != "nomis" {
			pDim.Title = title
			versionURL, err := url.Parse(latestVersionURL)
			if err != nil {
				log.Warn(ctx, "failed to parse url, last_version link", log.FormatErrors([]error{err}))
			}
			for _, dimension := range dims.Items {
				if dimension.Name == opt.Items[0].Name {
					pDim.Name = dimension.Name
					pDim.Description = dimension.Description
					if len(dimension.Label) > 0 {
						pDim.Title = dimension.Label
					}
				}
			}

			pDim.OptionsURL = fmt.Sprintf("%s/dimensions/%s/options", versionURL.Path, opt.Items[0].Name)
			pDim.TotalItems = totalCount

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
					if totalCount > maxNumberOfOptions {
						if i > 9 {
							break
						}
					}
					pDim.Values = append(pDim.Values, val.Label)
				}

				if opt.Items[0].Name == DimensionTime || opt.Items[0].Name == DimensionAge {
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
func MapCookiePreferences(req *http.Request, preferencesIsSet *bool, policy *dpRendererModel.CookiesPolicy) {
	preferencesCookie := cookies.GetCookiePreferences(req)
	*preferencesIsSet = preferencesCookie.IsPreferenceSet
	*policy = dpRendererModel.CookiesPolicy{
		Essential: preferencesCookie.Policy.Essential,
		Usage:     preferencesCookie.Policy.Usage,
	}
}

func mapEmergencyBanner(bannerData zebedee.EmergencyBanner) dpRendererModel.EmergencyBanner {
	var mappedEmergencyBanner dpRendererModel.EmergencyBanner
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
	for i := range versionList {
		if versionURI == versionList[i].URI {
			return versionList[i]
		}
	}
	return zebedee.Dataset{}
}

func MapDownloads(downloadsList []zebedee.Download, versionURI string) []datasetMdl.Download {
	dl := []datasetMdl.Download{}
	for _, d := range downloadsList {
		dl = append(dl, datasetMdl.Download{
			Extension:   filepath.Ext(d.File),
			Size:        d.Size,
			URI:         versionURI + "/" + d.File,
			DownloadURL: determineDownloadURL(d, versionURI),
		})
	}
	return dl
}

// Updates input `basePage` to include common dataset overview attributes, homepage content and
// dataset details across all dataset types
func UpdateBasePage(basePage *dpRendererModel.Page, dataset dpDatasetApiModels.Dataset,
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

	basePage.DatasetId = dataset.ID
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
	basePage.Metadata.Description = dataset.Description
	basePage.Metadata.Title = dataset.Title
	basePage.ServiceMessage = homepageContent.ServiceMessage
	basePage.Type = dataset.Type
	basePage.URI = request.URL.Path
}

func buildEditionsListTableOfContents(d dpDatasetApiModels.Dataset) dpRendererModel.TableOfContents {
	sections := make(map[string]dpRendererModel.ContentSection)
	displayOrder := make([]string, 0)

	tableOfContents := dpRendererModel.TableOfContents{
		AriaLabel: dpRendererModel.Localisation{
			LocaleKey: "ContentsAria",
			Plural:    1,
		},
		Title: dpRendererModel.Localisation{
			LocaleKey: "StaticTocHeading",
			Plural:    1,
		},
	}

	sections["editions-list"] = dpRendererModel.ContentSection{
		Title: dpRendererModel.Localisation{
			LocaleKey: "EditionListForDataset",
			Plural:    1,
		},
	}
	displayOrder = append(displayOrder, "editions-list")

	_, hasContactDetails := helpers.GetContactDetails(d)
	if hasContactDetails {
		sections["contact"] = dpRendererModel.ContentSection{
			Title: dpRendererModel.Localisation{
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

func CreateBreadcrumbsFromTopicList(baseURL string, topicObjectList []dpTopicApiModels.Topic) []dpRendererModel.TaxonomyNode {
	breadcrumbsObject := []dpRendererModel.TaxonomyNode{
		{
			Title: "Home",
			URI:   baseURL,
		},
	}
	path := baseURL

	for i, topicObject := range topicObjectList {
		if i == 0 {
			path += topicObject.Slug
		} else {
			// topic1 URI => /slug1 ... topic2 URI => /slug1/slug2... etc..
			path += "/" + topicObject.Slug
		}
		breadcrumbsObject = append(breadcrumbsObject, dpRendererModel.TaxonomyNode{
			Title: topicObject.Title,
			URI:   path,
		})
	}
	return breadcrumbsObject
}
