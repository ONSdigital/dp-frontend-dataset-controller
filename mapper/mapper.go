package mapper

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/datasetEditionsList"
	"github.com/ONSdigital/dp-frontend-models/model/datasetLandingPageFilterable"
	"github.com/ONSdigital/dp-frontend-models/model/datasetVersionsList"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/zebedee/data"
)

// SetTaxonomyDomain will set the taxonomy domain for a given pages
func SetTaxonomyDomain(p *model.Page) {
	p.TaxonomyDomain = os.Getenv("TAXONOMY_DOMAIN")
}

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

// CreateFilterableLandingPage creates a filterable dataset landing page based on api model responses
func CreateFilterableLandingPage(ctx context.Context, d dataset.Model, ver dataset.Version, datasetID string, opts []dataset.Options, dims dataset.Dimensions, displayOtherVersionsLink bool, breadcrumbs []data.Breadcrumb, latestVersionNumber int, latestVersionURL string) datasetLandingPageFilterable.Page {
	p := datasetLandingPageFilterable.Page{}
	SetTaxonomyDomain(&p.Page)
	p.Type = "dataset_landing_page"
	p.Metadata.Title = d.Title
	p.URI = d.Links.Self.URL
	p.DatasetLandingPage.UnitOfMeasurement = d.UnitOfMeasure
	p.Metadata.Description = d.Description
	p.ShowFeedbackForm = true
	p.DatasetId = datasetID
	p.ReleaseDate = ver.ReleaseDate
	p.BetaBannerEnabled = true

	for _, breadcrumb := range breadcrumbs {
		p.Page.Breadcrumb = append(p.Page.Breadcrumb, model.TaxonomyNode{
			Title: breadcrumb.Description.Title,
			URI:   breadcrumb.URI,
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
		log.ErrorCtx(ctx, err, nil)
	}
	datasetBreadcrumbs := []model.TaxonomyNode{
		{
			Title: d.Title,
			URI:   datasetURL.Path,
		},
		{
			Title: currentPageBreadcrumbTitle,
		},
	}
	p.Breadcrumb = append(p.Breadcrumb, datasetBreadcrumbs...)

	if len(d.Contacts) > 0 {
		p.ContactDetails.Name = d.Contacts[0].Name
		p.ContactDetails.Telephone = d.Contacts[0].Telephone
		p.ContactDetails.Email = d.Contacts[0].Email
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

	for _, meth := range d.Methodologies {
		p.DatasetLandingPage.Methodologies = append(p.DatasetLandingPage.Methodologies, datasetLandingPageFilterable.Methodology{
			Title:       meth.Title,
			URL:         meth.URL,
			Description: meth.Description,
		})
	}

	for _, pub := range d.Publications {
		p.DatasetLandingPage.Publications = append(p.DatasetLandingPage.Publications, datasetLandingPageFilterable.Publication{
			Title: pub.Title,
			URL:   pub.URL,
		})
	}

	for _, link := range d.RelatedDatasets {
		p.DatasetLandingPage.RelatedLinks = append(p.DatasetLandingPage.RelatedLinks, datasetLandingPageFilterable.Publication{
			Title: link.Title,
			URL:   link.URL,
		})
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
				log.ErrorCtx(ctx, err, nil)
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
						log.ErrorCtx(ctx, err, nil)
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
						pDim.Values = append(pDim.Values, fmt.Sprintf("All months between %s %d and %s %d", startDate.Month().String(), startDate.Year(), t.Month().String(), t.Year()))
						startDate = ts[i+1]
					} else {
						pDim.Values = append(pDim.Values, fmt.Sprintf("All months between %s %d and %s %d", startDate.Month().String(), startDate.Year(), t.Month().String(), t.Year()))
					}
				}

			} else if _, err = time.Parse("2006", opt.Items[0].Label); err == nil {
				var ts TimeSlice
				for _, val := range opt.Items {
					t, err := convertYYYYToTime(val.Label)
					if err != nil {
						log.Error(err, nil)
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
						pDim.Values = append(pDim.Values, fmt.Sprintf("All years between %d and %d", startDate.Year(), t.Year()))
						startDate = ts[i+1]
					} else {
						pDim.Values = append(pDim.Values, fmt.Sprintf("All years between %d and %d", startDate.Year(), t.Year()))
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
func CreateVersionsList(ctx context.Context, d dataset.Model, edition dataset.Edition, versions []dataset.Version) datasetVersionsList.Page {
	var p datasetVersionsList.Page
	SetTaxonomyDomain(&p.Page)
	// TODO refactor and make Welsh compatible.
	p.Metadata.Title = "All versions of " + d.Title
	if len(versions) > 0 {
		p.Metadata.Title += " " + versions[0].Edition
	}
	p.Metadata.Title += " dataset"
	p.BetaBannerEnabled = true
	uri, err := url.Parse(edition.Links.LatestVersion.URL)
	if err != nil {
		log.ErrorCtx(ctx, err, nil)
	}
	p.Data.LatestVersionURL = uri.Path
	p.DatasetId = d.ID

	latestVersionNumber := 1
	for i, ver := range versions {
		var version datasetVersionsList.Version
		version.IsLatest = false
		version.VersionNumber = ver.Version
		version.Title = d.Title
		version.Date = ver.ReleaseDate
		version.VersionURL = fmt.Sprintf("/datasets/%s/editions/%s/versions/%d", ver.Links.Dataset.ID, ver.Edition, ver.Version)
		version.FilterURL = fmt.Sprintf("/datasets/%s/editions/%s/versions/%d/filter", ver.Links.Dataset.ID, ver.Edition, ver.Version)

		if ver.Version > 1 {
			version.Superseded = fmt.Sprintf("/datasets/%s/editions/%s/versions/%d", ver.Links.Dataset.ID, ver.Edition, i)
		}

		if ver.Version > latestVersionNumber{
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
				if &alert != nil && alert.Type == correctionAlertType {
					version.Corrections = append(version.Corrections, datasetVersionsList.Correction{
						Reason: alert.Description,
						Date:   alert.Date,
					})
				}
			}
		}

		p.Data.Versions = append(p.Data.Versions, version)
	}

	for i, ver := range p.Data.Versions{
		if ver.VersionNumber == latestVersionNumber {
			p.Data.Versions[i].IsLatest = true
		}
	}

	sort.Slice(p.Data.Versions, func(i, j int) bool { return p.Data.Versions[i].VersionNumber > p.Data.Versions[j].VersionNumber })
	return p
}

// CreateEditionsList creates a editions list page based on api model responses
func CreateEditionsList(ctx context.Context, d dataset.Model, editions []dataset.Edition, datasetID string, breadcrumbs []data.Breadcrumb) datasetEditionsList.Page {
	p := datasetEditionsList.Page{}
	SetTaxonomyDomain(&p.Page)
	p.Type = "dataset_edition_list"
	p.Metadata.Title = d.Title
	p.URI = d.Links.Self.URL
	p.Metadata.Description = d.Description
	p.ShowFeedbackForm = true
	p.DatasetId = datasetID
	p.BetaBannerEnabled = true

	for _, bc := range breadcrumbs {
		p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
			Title: bc.Description.Title,
			URI:   bc.URI,
		})
	}

	// breadcrumbs won't contain this page in it's response from Zebedee, so add it to the slice
	p.Breadcrumb = append(p.Breadcrumb, model.TaxonomyNode{
		Title: d.Title,
	})

	if len(d.Contacts) > 0 {
		p.ContactDetails.Name = d.Contacts[0].Name
		p.ContactDetails.Telephone = d.Contacts[0].Telephone
		p.ContactDetails.Email = d.Contacts[0].Email
	}

	p.DatasetLandingPage.DatasetLandingPage.NextRelease = d.NextRelease
	p.DatasetLandingPage.DatasetID = datasetID

	if editions != nil && len(editions) > 0 {
		for _, edition := range editions {

			var latestVersionURL, err = url.Parse(edition.Links.LatestVersion.URL)
			if err != nil {
				log.ErrorCtx(ctx, err, nil)
			}
			var latestVersionPath = latestVersionURL.Path
			fmt.Println(latestVersionPath)

			var e datasetEditionsList.Edition
			e.Title = edition.Edition
			e.LatestVersionURL = latestVersionPath

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
