package mapper

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/zebedee"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/datasetLandingPageStatic"
	"github.com/ONSdigital/log.go/log"
)

// StaticDatasetLandingPage is a StaticDatasetLandingPage representation
type StaticDatasetLandingPage datasetLandingPageStatic.Page

// CreateLegacyDatasetLanding maps a zebedee response struct into a frontend model to be used for rendering
func CreateLegacyDatasetLanding(ctx context.Context, req *http.Request, dlp zebedee.DatasetLandingPage, bcs []zebedee.Breadcrumb, ds []zebedee.Dataset, localeCode string) StaticDatasetLandingPage {

	var sdlp StaticDatasetLandingPage

	MapCookiePreferences(req, &sdlp.CookiesPreferencesSet, &sdlp.CookiesPolicy)

	// Prepend 'legacy_' to type value to make it easier to differentiate between
	// filterable and legacy dataset pages - their models are different
	var pageType strings.Builder
	pageType.WriteString("legacy_")
	pageType.WriteString(dlp.Type)

	sdlp.Type = pageType.String()
	sdlp.URI = dlp.URI
	sdlp.Metadata.Title = dlp.Description.Title
	sdlp.Metadata.Description = dlp.Description.Summary
	sdlp.Language = localeCode
	sdlp.HasJSONLD = true

	for _, d := range dlp.RelatedDatasets {
		sdlp.DatasetLandingPage.Related.Datasets = append(sdlp.DatasetLandingPage.Related.Datasets, model.Related(d))
	}

	for _, d := range dlp.RelatedFilterableDatasets {
		sdlp.DatasetLandingPage.Related.FilterableDatasets = append(sdlp.DatasetLandingPage.Related.FilterableDatasets, model.Related(d))
	}

	for _, d := range dlp.RelatedDocuments {
		sdlp.DatasetLandingPage.Related.Publications = append(sdlp.DatasetLandingPage.Related.Publications, model.Related(d))
	}

	for _, d := range dlp.RelatedMethodology {
		sdlp.DatasetLandingPage.Related.Methodology = append(sdlp.DatasetLandingPage.Related.Methodology, model.Related(d))
	}
	for _, d := range dlp.RelatedMethodologyArticle {
		sdlp.DatasetLandingPage.Related.Methodology = append(sdlp.DatasetLandingPage.Related.Methodology, model.Related(d))
	}

	for _, d := range dlp.RelatedLinks {
		sdlp.DatasetLandingPage.Related.Links = append(sdlp.DatasetLandingPage.Related.Links, model.Related(d))
	}

	sdlp.DatasetLandingPage.IsNationalStatistic = dlp.Description.NationalStatistic
	sdlp.DatasetLandingPage.IsTimeseries = dlp.Timeseries
	sdlp.ContactDetails = model.ContactDetails(dlp.Description.Contact)

	// HACK FIX TODO REMOVE WHEN TIME IS SAVED CORRECTLY (GMT/UTC Issue)
	if strings.Contains(dlp.Description.ReleaseDate, "T23:00:00") {
		releaseDateInTimeFormat, err := time.Parse(time.RFC3339, dlp.Description.ReleaseDate)
		if err != nil {
			log.Event(ctx, "failed to parse release date", log.Error(err), log.Data{"release_date": dlp.Description.ReleaseDate})
			sdlp.DatasetLandingPage.ReleaseDate = dlp.Description.ReleaseDate
		}
		sdlp.DatasetLandingPage.ReleaseDate = releaseDateInTimeFormat.Add(1 * time.Hour).Format(time.RFC3339)
	} else {
		sdlp.DatasetLandingPage.ReleaseDate = dlp.Description.ReleaseDate
	}
	// END of hack fix
	sdlp.DatasetLandingPage.NextRelease = dlp.Description.NextRelease
	sdlp.DatasetLandingPage.DatasetID = dlp.Description.DatasetID
	sdlp.DatasetLandingPage.Notes = dlp.Section.Markdown

	for _, bc := range bcs {
		sdlp.Page.Breadcrumb = append(sdlp.Page.Breadcrumb, model.TaxonomyNode{
			Title: bc.Description.Title,
			URI:   bc.URI,
		})
	}

	if len(sdlp.Page.Breadcrumb) > 0 {
		sdlp.DatasetLandingPage.ParentPath = sdlp.Page.Breadcrumb[len(sdlp.Page.Breadcrumb)-1].Title
	}

	for i, d := range ds {
		var dataset datasetLandingPageStatic.Dataset
		for _, value := range d.Downloads {
			dataset.URI = d.URI
			dataset.VersionLabel = d.Description.VersionLabel
			dataset.Downloads = append(dataset.Downloads, datasetLandingPageStatic.Download{
				URI:       value.File,
				Extension: strings.TrimPrefix(filepath.Ext(value.File), "."),
				Size:      value.Size,
			})
		}
		for _, value := range d.SupplementaryFiles {
			dataset.SupplementaryFiles = append(dataset.SupplementaryFiles, datasetLandingPageStatic.SupplementaryFile{
				Title:     value.Title,
				URI:       value.File,
				Extension: strings.TrimPrefix(filepath.Ext(value.File), "."),
				Size:      value.Size,
			})
		}
		dataset.Title = d.Description.Edition
		if len(d.Versions) > 0 {
			dataset.HasVersions = true
		}
		dataset.IsLast = i+1 == len(ds)
		sdlp.DatasetLandingPage.Datasets = append(sdlp.DatasetLandingPage.Datasets, dataset)
	}

	for _, value := range dlp.Alerts {
		switch value.Type {
		default:
			log.Event(ctx, "Unrecognised alert type", log.Data{"alert": value})
			fallthrough
		case "alert":
			sdlp.DatasetLandingPage.Notices = append(sdlp.DatasetLandingPage.Notices, datasetLandingPageStatic.Message{
				Date:     value.Date,
				Markdown: value.Markdown,
			})
		case "correction":
			sdlp.DatasetLandingPage.Corrections = append(sdlp.DatasetLandingPage.Corrections, datasetLandingPageStatic.Message{
				Date:     value.Date,
				Markdown: value.Markdown,
			})
		}
	}

	return sdlp
}
