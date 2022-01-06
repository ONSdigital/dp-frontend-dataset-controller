package mapper

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/contactDetails"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetLandingPageStatic"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/related"
	coreModel "github.com/ONSdigital/dp-renderer/model"
	"github.com/ONSdigital/log.go/v2/log"
)

// StaticDatasetLandingPage is a StaticDatasetLandingPage representation
type StaticDatasetLandingPage datasetLandingPageStatic.Page

// CreateLegacyDatasetLanding maps a zebedee response struct into a frontend model to be used for rendering
func CreateLegacyDatasetLanding(basePage coreModel.Page, ctx context.Context, req *http.Request, dlp zebedee.DatasetLandingPage, bcs []zebedee.Breadcrumb, ds []zebedee.Dataset, localeCode string, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) StaticDatasetLandingPage {

	sdlp := StaticDatasetLandingPage{
		Page: basePage,
	}

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
	sdlp.FeatureFlags.SixteensVersion = SixteensVersion

	sdlp.ServiceMessage = serviceMessage
	sdlp.EmergencyBanner = mapEmergencyBanner(emergencyBannerContent)

	for _, d := range dlp.RelatedDatasets {
		sdlp.DatasetLandingPage.Related.Datasets = append(sdlp.DatasetLandingPage.Related.Datasets, related.Related(d))
	}

	for _, d := range dlp.RelatedFilterableDatasets {
		sdlp.DatasetLandingPage.Related.FilterableDatasets = append(sdlp.DatasetLandingPage.Related.FilterableDatasets, related.Related(d))
	}

	for _, d := range dlp.RelatedDocuments {
		sdlp.DatasetLandingPage.Related.Publications = append(sdlp.DatasetLandingPage.Related.Publications, related.Related(d))
	}

	for _, d := range dlp.RelatedMethodology {
		sdlp.DatasetLandingPage.Related.Methodology = append(sdlp.DatasetLandingPage.Related.Methodology, related.Related(d))
	}
	for _, d := range dlp.RelatedMethodologyArticle {
		sdlp.DatasetLandingPage.Related.Methodology = append(sdlp.DatasetLandingPage.Related.Methodology, related.Related(d))
	}

	for _, d := range dlp.RelatedLinks {
		sdlp.DatasetLandingPage.Related.Links = append(sdlp.DatasetLandingPage.Related.Links, related.Related(d))
	}

	sdlp.DatasetLandingPage.IsNationalStatistic = dlp.Description.NationalStatistic
	sdlp.DatasetLandingPage.IsTimeseries = dlp.Timeseries
	sdlp.ContactDetails = contactDetails.ContactDetails(dlp.Description.Contact)

	// HACK FIX TODO REMOVE WHEN TIME IS SAVED CORRECTLY (GMT/UTC Issue)
	if strings.Contains(dlp.Description.ReleaseDate, "T23:00:00") {
		releaseDateInTimeFormat, err := time.Parse(time.RFC3339, dlp.Description.ReleaseDate)
		if err != nil {
			log.Error(ctx, "failed to parse release date", err, log.Data{"release_date": dlp.Description.ReleaseDate})
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
		sdlp.Page.Breadcrumb = append(sdlp.Page.Breadcrumb, coreModel.TaxonomyNode{
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
			log.Error(ctx, "Unrecognised alert type", errors.New("Unrecognised alert type"), log.Data{"alert": value})
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
