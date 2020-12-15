package mapper

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/zebedee"
	"github.com/ONSdigital/dp-cookies/cookies"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/datasetEditionsList"
	"github.com/ONSdigital/dp-frontend-models/model/datasetLandingPageFilterable"
	"github.com/ONSdigital/dp-frontend-models/model/datasetVersionsList"

	"github.com/ONSdigital/log.go/log"
)

// TimeSlice allows sorting of a list of time.Time
type TimeSlice []time.Time

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
		log.Event(ctx, "wrong format for breadcrumb uri", log.WARN, log.Data{"breadcrumb": breadcrumb})
	} else {
		urlParsed.Path = strings.TrimPrefix(urlParsed.Path, apiRouterVersion)
		trimmedURI = urlParsed.String()
	}
	return trimmedURI
}

// CreateFilterableLandingPage creates a filterable dataset landing page based on api model responses
func CreateFilterableLandingPage(ctx context.Context, req *http.Request, d dataset.DatasetDetails, ver dataset.Version, datasetID string, opts []dataset.Options, dims dataset.VersionDimensions, displayOtherVersionsLink bool, breadcrumbs []zebedee.Breadcrumb, latestVersionNumber int, latestVersionURL, lang, apiRouterVersion string) datasetLandingPageFilterable.Page {
	p := datasetLandingPageFilterable.Page{}
	MapCookiePreferences(req, &p.Page.CookiesPreferencesSet, &p.Page.CookiesPolicy)
	p.Type = "dataset_landing_page"
	p.Metadata.Title = d.Title
	p.Language = lang
	p.URI = d.Links.Self.URL
	p.DatasetLandingPage.UnitOfMeasurement = d.UnitOfMeasure
	p.Metadata.Description = d.Description
	p.DatasetId = datasetID
	p.ReleaseDate = ver.ReleaseDate
	p.BetaBannerEnabled = true
	p.HasJSONLD = true

	// Trim API version path prefix from breadcrumb URIs, if present.
	for _, breadcrumb := range breadcrumbs {
		p.Page.Breadcrumb = append(p.Page.Breadcrumb, model.TaxonomyNode{
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
		datasetURL.Path = ""
		log.Event(ctx, "failed to parse url, self link", log.WARN, log.Error(err))
	}
	datasetPath := strings.TrimPrefix(datasetURL.Path, apiRouterVersion)
	datasetBreadcrumbs := []model.TaxonomyNode{
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

	var v datasetLandingPageFilterable.Version
	v.Title = d.Title
	v.Description = d.Description
	v.Edition = ver.Edition
	v.Version = strconv.Itoa(ver.Version)

	p.DatasetLandingPage.HasOlderVersions = displayOtherVersionsLink

	for k, download := range ver.Downloads {
		if len(download.URL) > 0 {
			v.Downloads = append(v.Downloads, datasetLandingPageFilterable.Download{
				Extension: k,
				Size:      download.Size,
				URI:       download.URL,
			})
		}
	}

	p.DatasetLandingPage.Version = v

	if len(opts) > 0 {
		for _, opt := range opts {

			var pDim datasetLandingPageFilterable.Dimension

			var title string
			if len(opt.Items) > 0 {
				title = strings.Title(opt.Items[0].DimensionID)
			}

			pDim.Title = title
			versionURL, err := url.Parse(d.Links.LatestVersion.URL)
			if err != nil {
				log.Event(ctx, "failed to parse url, last_version link", log.WARN, log.Error(err))
			}
			for _, dimension := range dims.Items {
				if dimension.Name == opt.Items[0].DimensionID {
					pDim.Description = dimension.Description
					if len(dimension.Label) > 0 {
						pDim.Title = dimension.Label
					}
				}
			}
			pDim.OptionsURL = fmt.Sprintf("%s/dimensions/%s/options", versionURL.Path, opt.Items[0].DimensionID)
			pDim.TotalItems = len(opt.Items)

			if _, err = time.Parse("Jan-06", opt.Items[0].Label); err == nil {
				var ts TimeSlice
				for _, val := range opt.Items {
					t, err := convertMMMYYToTime(val.Label)
					if err != nil {
						log.Event(ctx, "unable to convert date (MMYY) to time", log.WARN, log.Error(err), log.Data{"label": val.Label})
					}
					ts = append(ts, t)
				}
				sort.Sort(ts)

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

			} else if _, err = time.Parse("2006", opt.Items[0].Label); err == nil {
				var ts TimeSlice
				for _, val := range opt.Items {
					t, err := convertYYYYToTime(val.Label)
					if err != nil {
						log.Event(ctx, "unable to convert date (YYYY) to time", log.WARN, log.Error(err), log.Data{"label": val.Label})
					}
					ts = append(ts, t)
				}
				sort.Sort(ts)

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
			} else {

				for i, val := range opt.Items {
					if len(opt.Items) > 50 {
						if i > 9 {
							break
						}
					}
					pDim.Values = append(pDim.Values, val.Label)
				}

				if opt.Items[0].DimensionID == "time" || opt.Items[0].DimensionID == "age" {
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

			p.DatasetLandingPage.Dimensions = append(p.DatasetLandingPage.Dimensions, pDim)
		}
	}

	return p
}

// CreateVersionsList creates a versions list page based on api model responses
func CreateVersionsList(ctx context.Context, req *http.Request, d dataset.DatasetDetails, edition dataset.Edition, versions []dataset.Version) datasetVersionsList.Page {
	var p datasetVersionsList.Page
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

	latestVersionNumber := 1
	for _, ver := range versions {
		var version datasetVersionsList.Version
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
			version.Downloads = append(version.Downloads, datasetVersionsList.Download{
				Extension: ext,
				Size:      download.Size,
				URI:       download.URL,
			})
		}

		const correctionAlertType = "correction"
		if ver.Alerts != nil {
			for _, alert := range *ver.Alerts {
				if alert.Type == correctionAlertType {
					version.Corrections = append(version.Corrections, datasetVersionsList.Correction{
						Reason: alert.Description,
						Date:   alert.Date,
					})
				}
			}
		}

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
func CreateEditionsList(ctx context.Context, req *http.Request, d dataset.DatasetDetails, editions []dataset.Edition, datasetID string, breadcrumbs []zebedee.Breadcrumb, lang, apiRouterVersion string) datasetEditionsList.Page {
	p := datasetEditionsList.Page{}
	MapCookiePreferences(req, &p.Page.CookiesPreferencesSet, &p.Page.CookiesPolicy)
	p.Type = "dataset_edition_list"
	p.Language = lang
	p.Metadata.Title = d.Title
	p.URI = d.Links.Self.URL
	p.Metadata.Description = d.Description
	p.DatasetId = datasetID
	p.BetaBannerEnabled = true

	for _, bc := range breadcrumbs {
		p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
			Title: bc.Description.Title,
			URI:   getTrimmedBreadcrumbURI(ctx, bc, apiRouterVersion),
		})
	}

	// breadcrumbs won't contain this page in it's response from Zebedee, so add it to the slice
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
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

func convertMMMYYToTime(input string) (t time.Time, err error) {
	return time.Parse("Jan-06", input)
}

func convertYYYYToTime(input string) (t time.Time, err error) {
	return time.Parse("2006", input)
}

// MapCookiePreferences reads cookie policy and preferences cookies and then maps the values to the page model
func MapCookiePreferences(req *http.Request, preferencesIsSet *bool, policy *model.CookiesPolicy) {
	preferencesCookie := cookies.GetCookiePreferences(req)
	*preferencesIsSet = preferencesCookie.IsPreferenceSet
	*policy = model.CookiesPolicy{
		Essential: preferencesCookie.Policy.Essential,
		Usage:     preferencesCookie.Policy.Usage,
	}
}
