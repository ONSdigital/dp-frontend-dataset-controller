package mapper

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/dataset"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
	topicModel "github.com/ONSdigital/dp-topic-api/models"
)

// DatasetPage is a DatasetPage representation
type DatasetPage dataset.Page

func CreateDatasetPage(basePage coreModel.Page, req *http.Request, d zebedee.Dataset, dlp zebedee.DatasetLandingPage, bc []zebedee.Breadcrumb, versions []zebedee.Dataset, lang, serviceMessage string, emergencyBannerContent zebedee.EmergencyBanner, navigationContent *topicModel.Navigation) DatasetPage {
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

	dp.FeatureFlags.FeedbackAPIURL = cfg.FeedbackAPIURL

	for _, breadcrumb := range bc {
		dp.Page.Breadcrumb = append(dp.Page.Breadcrumb, coreModel.TaxonomyNode{
			Title: breadcrumb.Description.Title,
			URI:   breadcrumb.URI,
		})
	}

	dp.Page.Breadcrumb = append(dp.Page.Breadcrumb, coreModel.TaxonomyNode{
		Title: dp.DatasetPage.Edition,
	})

	if len(dp.Page.Breadcrumb) > 0 {
		dp.DatasetPage.ParentPath = dp.Page.Breadcrumb[len(dp.Page.Breadcrumb)-1].Title
	}

	dp.DatasetPage.IsNationalStatistic = dlp.Description.NationalStatistic
	dp.DatasetPage.NextRelease = dlp.Description.NextRelease
	dp.DatasetPage.DatasetID = dlp.Description.DatasetID

	dp.Details.Email = strings.TrimSpace(dlp.Description.Contact.Email)
	dp.Details.Name = dlp.Description.Contact.Name
	dp.Details.Telephone = dlp.Description.Contact.Telephone

	if navigationContent != nil {
		dp.NavigationContent = MapNavigationContent(*navigationContent)
	}

	for _, download := range d.Downloads {
		dp.DatasetPage.Downloads = append(
			dp.DatasetPage.Downloads,
			dataset.Download{
				Extension:   filepath.Ext(download.File),
				Size:        download.Size,
				URI:         dp.URI + "/" + download.File,
				File:        download.File,
				DownloadURL: determineDownloadURL(download, dp.URI),
			})
	}

	for _, supplementaryFile := range d.SupplementaryFiles {
		dp.DatasetPage.SupplementaryFiles = append(
			dp.DatasetPage.SupplementaryFiles,
			dataset.SupplementaryFile{
				Title:       supplementaryFile.Title,
				Extension:   filepath.Ext(supplementaryFile.File),
				Size:        supplementaryFile.Size,
				URI:         dp.URI + "/" + supplementaryFile.File,
				DownloadURL: determineSupplementaryFileURL(supplementaryFile, dp.URI),
			})
	}

	var reversed = dp.DatasetPage.Versions
	for _, ver := range d.Versions {
		dp.DatasetPage.Versions = append(
			dp.DatasetPage.Versions,
			dataset.Version{
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

func determineDownloadURL(download zebedee.Download, datasetPageURI string) string {
	var downloadURL string
	if download.URI != "" {
		downloadURL = "/" + staticFilesDownloadEndpoint + download.URI
	} else {
		downloadURL = "/file?uri=" + datasetPageURI + "/" + download.File
	}
	return downloadURL
}

func determineSupplementaryFileURL(supplementaryFile zebedee.SupplementaryFile, datasetPageURI string) string {
	var downloadURL string
	if supplementaryFile.URI != "" {
		downloadURL = "/" + staticFilesDownloadEndpoint + supplementaryFile.URI
	} else {
		downloadURL = "/file?uri=" + datasetPageURI + "/" + supplementaryFile.File
	}
	return downloadURL
}

// mapNavigationContent takes navigationContent as returned from the client and returns information needed for the navigation bar
func MapNavigationContent(navigationContent topicModel.Navigation) []coreModel.NavigationItem {
	var mappedNavigationContent []coreModel.NavigationItem
	if navigationContent.Items != nil {
		for _, rootContent := range *navigationContent.Items {
			var subItems []coreModel.NavigationItem
			if rootContent.SubtopicItems != nil {
				for _, subtopicContent := range *rootContent.SubtopicItems {
					subItems = append(subItems, coreModel.NavigationItem{
						Uri:   subtopicContent.URI,
						Label: subtopicContent.Label,
					})
				}
			}
			mappedNavigationContent = append(mappedNavigationContent, coreModel.NavigationItem{
				Uri:      rootContent.URI,
				Label:    rootContent.Label,
				SubItems: subItems,
			})
		}
	}
	return mappedNavigationContent
}
