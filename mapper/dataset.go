package mapper

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetPage"
	coreModel "github.com/ONSdigital/dp-renderer/model"
	"net/http"
	"path/filepath"
	"strings"
)

// DatasetPage is a DatasetPage representation
type DatasetPage datasetPage.Page

func CreateDatasetPage(basePage coreModel.Page, ctx context.Context, req *http.Request, d zebedee.Dataset, dlp zebedee.DatasetLandingPage, bc []zebedee.Breadcrumb, versions []zebedee.Dataset, lang string, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner) DatasetPage {

	dp := DatasetPage{
		Page: basePage,
	}

	MapCookiePreferences(req, &dp.CookiesPreferencesSet, &dp.CookiesPolicy)

	dp.Type = "dataset"
	dp.Metadata.Title = dlp.Description.Title
	dp.Language = lang
	dp.URI = d.URI
	dp.DatasetPage.URI = dlp.URI
	dp.Metadata.Description = dlp.Description.Summary
	dp.DatasetPage.ReleaseDate = dlp.Description.ReleaseDate
	dp.DatasetPage.Edition = d.Description.Edition
	dp.DatasetPage.Markdown = dlp.Section.Markdown
	dp.FeatureFlags.SixteensVersion = SixteensVersion

	dp.ServiceMessage = serviceMessage
	dp.EmergencyBanner = mapEmergencyBanner(emergencyBannerContent)

	for _, breadcrumb := range bc {
		dp.Page.Breadcrumb = append(dp.Page.Breadcrumb, coreModel.TaxonomyNode{
			Title: breadcrumb.Description.Title,
			URI:   breadcrumb.URI,
		})
	}

	dp.Page.Breadcrumb = append(dp.Page.Breadcrumb, coreModel.TaxonomyNode{
		Title: dp.DatasetPage.Edition,
	})

	dp.DatasetPage.IsNationalStatistic = dlp.Description.NationalStatistic
	dp.DatasetPage.NextRelease = dlp.Description.NextRelease
	dp.DatasetPage.DatasetID = dlp.Description.DatasetID

	dp.ContactDetails.Email = strings.TrimSpace(dlp.Description.Contact.Email)
	dp.ContactDetails.Name = dlp.Description.Contact.Name
	dp.ContactDetails.Telephone = dlp.Description.Contact.Telephone

	for _, download := range d.Downloads {
		dp.DatasetPage.Downloads = append(
			dp.DatasetPage.Downloads,
			datasetPage.Download{
				Extension: filepath.Ext(download.File),
				Size:      download.Size,
				URI:       dp.URI + "/" + download.File,
				File:      download.File})
	}

	for _, supplementaryFile := range d.SupplementaryFiles {
		dp.DatasetPage.SupplementaryFiles = append(
			dp.DatasetPage.SupplementaryFiles,
			datasetPage.SupplementaryFile{
				Title:     supplementaryFile.Title,
				Extension: filepath.Ext(supplementaryFile.File),
				Size:      supplementaryFile.Size,
				URI:       dp.URI + "/" + supplementaryFile.File})
	}

	var reversed = dp.DatasetPage.Versions
	for _, ver := range d.Versions {
		dp.DatasetPage.Versions = append(
			dp.DatasetPage.Versions,
			datasetPage.Version{
				URI:              ver.URI,
				UpdateDate:       ver.ReleaseDate,
				CorrectionNotice: ver.Notice,
				Label:            ver.Label,
				Downloads:        MapDownloads(FindVersion(versions, ver.URI).Downloads, ver.URI)})
	}

	for i := range dp.DatasetPage.Versions {
		n := dp.DatasetPage.Versions[len(dp.DatasetPage.Versions)-1-i]
		reversed = append(reversed, n)
	}

	dp.DatasetPage.Versions = reversed

	return dp
}
