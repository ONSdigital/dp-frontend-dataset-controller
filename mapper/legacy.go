package mapper

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/contact"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/related"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/staticlegacy"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
	topicModel "github.com/ONSdigital/dp-topic-api/models"
	"github.com/ONSdigital/log.go/v2/log"
)

const staticFilesDownloadEndpoint = "downloads-new"

// StaticDatasetLandingPage is a StaticDatasetLandingPage representation
type StaticDatasetLandingPage staticlegacy.Page

// CreateLegacyDatasetLanding maps a zebedee response struct into a frontend model to be used for rendering
//
//nolint:gocyclo //complexity 21
func CreateLegacyDatasetLanding(ctx context.Context, basePage coreModel.Page, req *http.Request, dlp zebedee.DatasetLandingPage, bcs []zebedee.Breadcrumb, ds []zebedee.Dataset, localeCode, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner, navigationContent *topicModel.Navigation) StaticDatasetLandingPage {
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

	sdlp.FeatureFlags.FeedbackAPIURL = cfg.FeedbackAPIURL

	if navigationContent != nil {
		sdlp.NavigationContent = MapNavigationContent(*navigationContent)
	}

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
	sdlp.DatasetLandingPage.Survey = dlp.Description.Survey
	sdlp.Details = contact.Details(dlp.Description.Contact)

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
		sdlp.Breadcrumb = append(sdlp.Breadcrumb, coreModel.TaxonomyNode{
			Title: bc.Description.Title,
			URI:   bc.URI,
		})
	}

	if len(sdlp.Breadcrumb) > 0 {
		sdlp.DatasetLandingPage.ParentPath = sdlp.Breadcrumb[len(sdlp.Breadcrumb)-1].Title
	}

	for i := range ds {
		var dataset staticlegacy.Dataset
		dataset.URI = ds[i].URI
		dataset.VersionLabel = ds[i].Description.VersionLabel

		for _, download := range ds[i].Downloads {
			if download.URI != "" { // i.e. new static files sourced files
				filePath := strings.TrimPrefix(download.URI, "/")
				dataset.Downloads = append(dataset.Downloads, staticlegacy.Download{
					URI:         download.URI,
					DownloadURL: fmt.Sprintf("/%s/%s", staticFilesDownloadEndpoint, filePath),
					Extension:   strings.TrimPrefix(filepath.Ext(download.URI), "."),
					Size:        download.Size,
				})
			} else { // old legacy Zebedee-source files
				dataset.Downloads = append(dataset.Downloads, staticlegacy.Download{
					URI:         download.File,
					DownloadURL: fmt.Sprintf("/file?uri=%s/%s", dataset.URI, download.File),
					Extension:   strings.TrimPrefix(filepath.Ext(download.File), "."),
					Size:        download.Size,
				})
			}
		}
		for _, supplementaryFile := range ds[i].SupplementaryFiles {
			if supplementaryFile.URI != "" { // i.e. new static files sourced files
				filePath := strings.TrimPrefix(supplementaryFile.URI, "/")
				dataset.SupplementaryFiles = append(dataset.SupplementaryFiles, staticlegacy.SupplementaryFile{
					Title:       supplementaryFile.Title,
					URI:         supplementaryFile.URI,
					DownloadURL: fmt.Sprintf("/%s/%s", staticFilesDownloadEndpoint, filePath),
					Extension:   strings.TrimPrefix(filepath.Ext(supplementaryFile.URI), "."),
					Size:        supplementaryFile.Size,
				})
			} else { // old legacy Zebedee-source files
				dataset.SupplementaryFiles = append(dataset.SupplementaryFiles, staticlegacy.SupplementaryFile{
					Title:       supplementaryFile.Title,
					URI:         supplementaryFile.File,
					DownloadURL: fmt.Sprintf("/file?uri=%s/%s", dataset.URI, supplementaryFile.File),
					Extension:   strings.TrimPrefix(filepath.Ext(supplementaryFile.File), "."),
					Size:        supplementaryFile.Size,
				})
			}
		}

		dataset.Title = ds[i].Description.Edition

		if len(ds[i].Versions) > 0 {
			dataset.HasVersions = true
		}
		dataset.IsLast = i+1 == len(ds)
		sdlp.DatasetLandingPage.Datasets = append(sdlp.DatasetLandingPage.Datasets, dataset)
	}

	for _, value := range dlp.Alerts {
		switch value.Type {
		default:
			log.Error(ctx, "unrecognised alert type", errors.New("unrecognised alert type"), log.Data{"alert": value})
			fallthrough
		case "alert":
			sdlp.DatasetLandingPage.Notices = append(sdlp.DatasetLandingPage.Notices, staticlegacy.Message{
				Date:     value.Date,
				Markdown: value.Markdown,
			})
		case "correction":
			sdlp.DatasetLandingPage.Corrections = append(sdlp.DatasetLandingPage.Corrections, staticlegacy.Message{
				Date:     value.Date,
				Markdown: value.Markdown,
			})
		}
	}

	return sdlp
}
