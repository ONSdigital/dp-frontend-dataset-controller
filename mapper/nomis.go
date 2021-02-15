package mapper

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/zebedee"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/datasetLandingPageFilterable"
	"github.com/ONSdigital/log.go/log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// CreateNomisLandingPage creates a nomis dataset landing page based on api model responses
func CreateNomisLandingPage(ctx context.Context, req *http.Request, d dataset.DatasetDetails, ver dataset.Version, datasetID string, displayOtherVersionsLink bool, breadcrumbs []zebedee.Breadcrumb, latestVersionNumber int, latestVersionURL, lang string) datasetLandingPageFilterable.Page {
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
		log.Event(ctx, "failed to parse url, self link", log.WARN, log.Error(err))
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
	p.DatasetLandingPage.Version = v

	return p
}
