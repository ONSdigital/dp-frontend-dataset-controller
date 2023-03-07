package mocks

import "strings"

var cyLocale = []string{
	"[HasCorrectionNotice]",
	"one = \"Correction notice\"",
	"[HasNewVersion]",
	"one = \"New version\"",
	"[QualityNoticeReadMore]",
	"one = \"Read more about this\"",
	"[HasAlert]",
	"one = \"Important notice\"",
	"[ImproveResultsSubHeading]",
	"one = \"Try the following\"",
	"[ImproveResultsList]",
	"one = \"A list of suggestions\"",
	"[SDCAreasAvailable]",
	"one = \"10 out of 25 areas available\"",
	"[SDCRestrictedAreas]",
	"one = \"Protecting personal data will prevent 1 area from being published\"",
	"other = \"Protecting personal data will prevent 15 areas from being published\"",
	"[SDCAllAreasAvailable]",
	"one = \"1 area available\"",
	"other = \"All 10 areas available\"",
	"[CreateCustomDatasetTitle]",
	"one = \"Create a custom dataset\"",
}

var enLocale = []string{
	"[HasCorrectionNotice]",
	"one = \"Correction notice\"",
	"[HasNewVersion]",
	"one = \"New version\"",
	"[QualityNoticeReadMore]",
	"one = \"Read more about this\"",
	"[HasAlert]",
	"one = \"Important notice\"",
	"[ImproveResultsSubHeading]",
	"one = \"Try the following\"",
	"[ImproveResultsList]",
	"one = \"A list of suggestions\"",
	"[SDCAreasAvailable]",
	"one = \"10 out of 25 areas available\"",
	"[SDCRestrictedAreas]",
	"one = \"Protecting personal data will prevent 1 area from being published\"",
	"other = \"Protecting personal data will prevent 15 areas from being published\"",
	"[SDCAllAreasAvailable]",
	"one = \"1 area available\"",
	"other = \"All 10 areas available\"",
	"[CreateCustomDatasetTitle]",
	"one = \"Create a custom dataset\"",
}

// MockAssetFunction returns mocked toml []bytes
func MockAssetFunction(name string) ([]byte, error) {
	if strings.Contains(name, ".cy.toml") {
		return []byte(strings.Join(cyLocale, "\n")), nil
	}
	return []byte(strings.Join(enLocale, "\n")), nil
}
