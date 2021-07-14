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
