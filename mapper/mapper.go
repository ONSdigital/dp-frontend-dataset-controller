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

	datasetMdl "github.com/ONSdigital/dp-frontend-dataset-controller/model/dataset"
	filterable "github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetLandingPageFilterable"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-cookies/cookies"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/Edition"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/version"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"

	"github.com/ONSdigital/log.go/v2/log"
)

// TimeSlice allows sorting of a list of time.Time
type TimeSlice []time.Time

// Constants names
const (
	DimensionTime      = "time"
	DimensionAge       = "age"
	DimensionGeography = "geography"
	SixteensVersion    = "30948d6"
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
func CreateFilterableLandingPage(basePage coreModel.Page, ctx context.Context, req *http.Request, d dataset.DatasetDetails, ver dataset.Version, datasetID string, opts []dataset.Options, dims dataset.VersionDimensions, displayOtherVersionsLink bool, breadcrumbs []zebedee.Breadcrumb, latestVersionNumber int, latestVersionURL, lang, apiRouterVersion string, maxNumOpts int, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) filterable.Page {
	p := filterable.Page{
		Page: basePage,
	}
	MapCookiePreferences(req, &p.Page.CookiesPreferencesSet, &p.Page.CookiesPolicy)
	p.Type = "dataset_landing_page"
	p.Metadata.Title = d.Title
	p.Language = lang
	p.URI = req.URL.Path
	p.DatasetLandingPage.UnitOfMeasurement = d.UnitOfMeasure
	p.Metadata.Description = d.Description
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
				p.DatasetLandingPage.UsageNotes = append(p.DatasetLandingPage.UsageNotes, filterable.UsageNote{
					Title: usageNote.Title,
					Note:  usageNote.Note,
				})
			}
		}
	}
	if d.Methodologies != nil {
		for _, meth := range *d.Methodologies {
			p.DatasetLandingPage.Methodologies = append(p.DatasetLandingPage.Methodologies, filterable.Methodology{
				Title:       meth.Title,
				URL:         meth.URL,
				Description: meth.Description,
			})
		}
	}

	if d.Publications != nil {
		for _, pub := range *d.Publications {
			p.DatasetLandingPage.Publications = append(p.DatasetLandingPage.Publications, filterable.Publication{
				Title: pub.Title,
				URL:   pub.URL,
			})
		}
	}

	if d.RelatedDatasets != nil {
		for _, link := range *d.RelatedDatasets {
			p.DatasetLandingPage.RelatedLinks = append(p.DatasetLandingPage.RelatedLinks, filterable.Publication{
				Title: link.Title,
				URL:   link.URL,
			})
		}
	}

	for _, changes := range ver.LatestChanges {
		p.DatasetLandingPage.LatestChanges = append(p.DatasetLandingPage.LatestChanges, filterable.Change{
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
func CreateVersionsList(basePage coreModel.Page, req *http.Request, d dataset.DatasetDetails, edition dataset.Edition, versions []dataset.Version, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) version.Page {
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

	p.Data.LatestVersionURL = helpers.DatasetVersionURL(d.ID, edition.Edition, edition.Links.LatestVersion.ID)
	p.DatasetId = d.ID
	p.URI = req.URL.Path
	p.FeatureFlags.SixteensVersion = SixteensVersion

	p.ServiceMessage = serviceMessage
	p.EmergencyBanner = mapEmergencyBanner(emergencyBannerContent)

	latestVersionNumber := 1
	for i, ver := range versions {
		var v sharedModel.Version
		v.IsLatest = false
		v.VersionNumber = ver.Version
		v.Title = d.Title
		v.Date = ver.ReleaseDate
		versionURL := helpers.DatasetVersionURL(ver.Links.Dataset.ID, ver.Edition, strconv.Itoa(ver.Version))
		v.VersionURL = versionURL
		v.FilterURL = versionURL + "/filter"

		// Not the 'created' first version and more than one stored version
		if ver.Version > 1 && len(p.Data.Versions) >= 1 {
			previousVersion := p.Data.Versions[len(p.Data.Versions)-1].VersionNumber
			v.Superseded = helpers.DatasetVersionURL(ver.Links.Dataset.ID, ver.Edition, strconv.Itoa(previousVersion))
		}

		if ver.Version > latestVersionNumber {
			latestVersionNumber = ver.Version
		}

		for ext, download := range ver.Downloads {
			v.Downloads = append(v.Downloads, sharedModel.Download{
				Extension: ext,
				Size:      download.Size,
				URI:       download.URL,
			})
		}

		mapCorrectionAlert(&versions[i], &v)

		p.Data.Versions = append(p.Data.Versions, v)
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
func CreateEditionsList(ctx context.Context, basePage coreModel.Page, req *http.Request, d dataset.DatasetDetails, editions []dataset.Edition, datasetID string, breadcrumbs []zebedee.Breadcrumb, lang, apiRouterVersion, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) Edition.Page {
	p := Edition.Page{
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
			var e Edition.List
			e.Title = edition.Edition
			e.LatestVersionURL = helpers.DatasetVersionURL(datasetID, edition.Edition, edition.Links.LatestVersion.ID)
			p.Editions = append(p.Editions, e)
		}
	}

	return p
}

func mapCorrectionAlert(ver *dataset.Version, model *sharedModel.Version) {
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
