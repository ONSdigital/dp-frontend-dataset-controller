package helpers

import (
	"fmt"
	"net/url"
	"regexp"
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

// DatasetVersionUrl constructs a dataset version URL from the provided datasetID, edition and version values
func DatasetVersionUrl(datasetID string, edition string, version string) string {
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

// GetCurrentUrl returns a string of the current URL from language, site domain and url path parameters
func GetCurrentUrl(lang string, siteDomain string, urlPath string) string {
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
func GenerateSharingLink(socialType string, currentUrl string, title string) string {
	switch socialType {
	case "facebook":
		return fmt.Sprintf("https://www.facebook.com/sharer/sharer.php?u=%s", currentUrl)
	case "twitter":
		return fmt.Sprintf("https://twitter.com/intent/tweet?original_referer&text=%s&url=%s", title, currentUrl)
	case "linkedin":
		return fmt.Sprintf("https://www.linkedin.com/sharing/share-offsite/?url=%s", currentUrl)
	case "email":
		return fmt.Sprintf("mailto:?subject=%s&body=%s\n%s", title, title, currentUrl)
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
