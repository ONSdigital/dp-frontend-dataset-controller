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
	case "facebook":
		return fmt.Sprintf("https://www.facebook.com/sharer/sharer.php?u=%s", currentURL)
	case "twitter":
		return fmt.Sprintf("https://twitter.com/intent/tweet?original_referer&text=%s&url=%s", title, currentURL)
	case "linkedin":
		return fmt.Sprintf("https://www.linkedin.com/sharing/share-offsite/?url=%s", currentURL)
	case "email":
		return fmt.Sprintf("mailto:?subject=%s&body=%s\n%s", title, title, currentURL)
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

// GetOfficialStatisticsLogo returns the official statistics logo based on the enableOfficialStatisticsLogo and language
func GetOfficialStatisticsLogo(enableOfficialStatisticsLogo, useSvg bool, language string) string {
	extension := ".png"
	if useSvg {
		extension = ".svg"
	}

	if enableOfficialStatisticsLogo {
		return fmt.Sprintf("https://cdn.ons.gov.uk/assets/images/ons-logo/kitemark/v2/uksa-kitemark-%s%s", language, extension)
	}

	if useSvg {
		return "https://cdn.ons.gov.uk/assets/images/ons-logo/kitemark/uksa-kitemark.svg"
	}

	return "/img/national-statistics.png"
}
