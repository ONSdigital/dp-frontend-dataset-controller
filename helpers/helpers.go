package helpers

import (
	"fmt"
	"regexp"
)

// ExtractDatasetInfoFromPath gets the datasetID, edition and version from a given path
func ExtractDatasetInfoFromPath(path string) (datasetID, edition, version string, err error) {
	pathReg := regexp.MustCompile(`\/datasets\/(.+)\/editions\/(.+)\/versions\/(.+)`)
	subs := pathReg.FindStringSubmatch(path)
	if len(subs) < 4 {
		err = fmt.Errorf("unabled to extract datasetID, edition and version from path: %s", path)
		return
	}
	return subs[1], subs[2], subs[3], nil
}
