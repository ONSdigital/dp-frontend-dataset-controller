package mocks

import "strings"

var cyLocale = []string{
	"[HasCorrectionNotice]",
	"one = \"Correction notice\"",
	"[HasNewVersion]",
	"one = \"New version\"",
	"[QualityNoticeReadMore]",
	"one = \"Read more about this\"",
}

var enLocale = []string{
	"[HasCorrectionNotice]",
	"one = \"Correction notice\"",
	"[HasNewVersion]",
	"one = \"New version\"",
	"[QualityNoticeReadMore]",
	"one = \"Read more about this\"",
}

// MockAssetFunction returns mocked toml []bytes
func MockAssetFunction(name string) ([]byte, error) {
	if strings.Contains(name, ".cy.toml") {
		return []byte(strings.Join(cyLocale, "\n")), nil
	}
	return []byte(strings.Join(enLocale, "\n")), nil
}
