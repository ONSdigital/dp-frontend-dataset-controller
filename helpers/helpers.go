package helpers

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/contact"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/osrlogo"
)

// ExtractDatasetInfoFromPath gets the datasetID, edition and version from a given path
func ExtractDatasetInfoFromPath(path string) (datasetID, edition, version string, err error) {
	pathReg := regexp.MustCompile(`/datasets/(.+)/editions/(.+)/versions/(.+)`)
	subs := pathReg.FindStringSubmatch(path)
	if len(subs) < 4 {
		err = fmt.Errorf("unable to extract datasetID, edition and version from path: %s", path)
		return
	}
	return subs[1], subs[2], subs[3], nil
}

// DatasetVersionURL constructs a dataset version URL from the provided datasetID, edition and version values
func DatasetVersionURL(datasetID, edition, version string) string {
	return fmt.Sprintf("/datasets/%s/editions/%s/versions/%s", datasetID, edition, version)
}

// GetAPIRouterVersion returns the path of the provided url, which corresponds to the api router version
func GetAPIRouterVersion(rawurl string) (string, error) {
	apiRouterURL, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	return apiRouterURL.Path, nil
}

// GetCurrentURL returns a string of the current URL from language, site domain and url path parameters
func GetCurrentURL(lang, siteDomain, urlPath string) string {
	var welshPrepend string
	if lang == "cy" {
		welshPrepend = "cy."
	}

	if siteDomain == "localhost" || siteDomain == "" {
		siteDomain = "ons.gov.uk"
	}

	return welshPrepend + siteDomain + urlPath
}

// GenerateSharingLink returns a sharing link for different types of social media
func GenerateSharingLink(socialType, currentURL, title string) string {
	switch socialType {
	case "email":
		return fmt.Sprintf("mailto:?subject=%s&body=%s\n%s", title, title, currentURL)
	case "facebook":
		return fmt.Sprintf("https://www.facebook.com/sharer/sharer.php?u=%s", currentURL)
	case "linkedin":
		return fmt.Sprintf("https://www.linkedin.com/sharing/share-offsite/?url=%s", currentURL)
	case "twitter", "x":
		return fmt.Sprintf("https://%s.com/intent/tweet?original_referer&text=%s&url=%s", socialType, title, currentURL)
	default:
		return ""
	}
}

// IsBoolPtr determines if the given value is a pointer
func IsBoolPtr(val *bool) bool {
	if val == nil {
		return false
	}

	return *val
}

// HasStringInSlice checks for a string within in a string array
func HasStringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// PersistExistingParams persists existing query string values and ignores a given value
func PersistExistingParams(values []string, key, ignoreValue string, q url.Values) {
	for _, value := range values {
		if value != ignoreValue {
			q.Add(key, value)
		}
	}
}

// ToBoolPtr converts a boolean to a pointer
func ToBoolPtr(val bool) *bool {
	return &val
}

// GetOSRLogoDetails returns the official statistics logo details based on the language
func GetOSRLogoDetails(language string) osrlogo.OSRLogo {
	altText := "Official Statistics logo"
	title := "Accredited official statistics"
	about := "Confirmed by the Office for Statistics Regulation as compliant with the Code of Practice for Statistics."

	return osrlogo.OSRLogo{
		URL:     fmt.Sprintf("https://cdn.ons.gov.uk/assets/images/ons-logo/kitemark/v2/uksa-kitemark-%s%s", language, ".svg"),
		AltText: altText,
		Title:   title,
		About:   about,
	}
}

// Loops through the input list of distributions to find a distribution with a format that matches the
// requested format.
// If a matching distribution is found, the download URL of that distribution is returned.
// If no matching distribution is found, it returns an empty string.
func GetDistributionFileURL(distributionList *[]dpDatasetApiModels.Distribution, requestedFormat string) string {
	distributions := *distributionList
	for _, distribution := range distributions {
		if strings.EqualFold(distribution.Format.String(), requestedFormat) {
			return distribution.DownloadURL
		}
	}
	return ""
}

// Loops through a list of download objects to find a download with a format that matches the
// requested format.
// If a matching download object is found, it returns the HRef (URL) of that download object.
// If no matching download object is found, it returns an empty string.
func GetDownloadFileURL(downloadList *dpDatasetApiModels.DownloadList, requestedFormat string) string {
	if downloadList != nil {
		// We need a way to map `DownloadObject` identifiers to extension strings
		downloadObjects := MapDownloadObjectExtensions(downloadList)
		// Loop through the possible downloadobjects and redirect to the requested one
		for downloadObject, extension := range downloadObjects {
			if strings.EqualFold(extension, requestedFormat) {
				if downloadObject != nil {
					return downloadObject.HRef
				}
			}
		}
	}
	return ""
}

// Creates a mapping from download objects to their corresponding file extension strings
func MapDownloadObjectExtensions(downloadList *dpDatasetApiModels.DownloadList) map[*dpDatasetApiModels.DownloadObject]string {
	downloads := *downloadList
	return map[*dpDatasetApiModels.DownloadObject]string{
		downloads.XLS:  "xls",
		downloads.XLSX: "xlsx",
		downloads.CSV:  "csv",
		downloads.TXT:  "txt",
		downloads.CSVW: "csvw",
	}
}

// Maps download objects from a dp-dataset-api Version to download details, including file extensions, sizes, and URIs
func MapVersionDownloads(sharedModelVersion *sharedModel.Version, downloadList *dpDatasetApiModels.DownloadList) {
	if downloadList != nil {
		downloadObjects := MapDownloadObjectExtensions(downloadList)
		// Loop through the possible downloadobjects and add to downloads if valid
		for downloadObject, extension := range downloadObjects {
			if downloadObject != nil {
				// We need a valid `HRef` at the very least to create a valid download
				if downloadObject.HRef != "" {
					sharedModelVersion.Downloads = append(sharedModelVersion.Downloads, sharedModel.Download{
						Extension: strings.ToLower(extension),
						Size:      downloadObject.Size,
						URI:       downloadObject.HRef,
					})
				}
			}
		}
	}
}

// Collects the contact details of a datset into a contacts object and provides a HasContactDetails flag.
func GetContactDetails(d dpDatasetApiModels.Dataset) (contact.Details, bool) {
	details := contact.Details{}
	hasContactDetails := false

	if len(d.Contacts) > 0 {
		contacts := d.Contacts
		if d.Type == "static" {
			if contacts[0].Name != "" {
				details.Name = contacts[0].Name
				hasContactDetails = true
			}
		}
		if contacts[0].Telephone != "" {
			details.Telephone = contacts[0].Telephone
			hasContactDetails = true
		}
		if contacts[0].Email != "" {
			details.Email = contacts[0].Email
			hasContactDetails = true
		}
	}

	return details, hasContactDetails
}
