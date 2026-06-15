package helpers

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPrefixPathWithTopic(t *testing.T) {
	testCases := []struct {
		name     string
		topicID  string
		rawURL   string
		expected string
	}{
		{
			name:     "full URL",
			topicID:  "topic",
			rawURL:   "http://example.com/v1/some/path",
			expected: "/topic/some/path",
		},
		{
			name:     "relative URL",
			topicID:  "topic",
			rawURL:   "/v1/some/path",
			expected: "/topic/some/path",
		},
		{
			name:     "relative URL without leading slash",
			topicID:  "topic",
			rawURL:   "v1/some/path",
			expected: "/topic/some/path",
		},
		{
			name:     "relative URL with multiple slashes",
			topicID:  "topic",
			rawURL:   "///v1/some///path///",
			expected: "/topic/some/path",
		},
	}

	for _, tc := range testCases {
		Convey(tc.name, t, func() {
			result, err := PrefixPathWithTopic(tc.topicID, tc.rawURL)
			So(err, ShouldBeNil)
			So(result, ShouldEqual, tc.expected)
		})
	}

	Convey("Given an invalid url", t, func() {
		url := "://invalid-url"

		Convey("When PrefixPathWithTopic is called", func() {
			_, err := PrefixPathWithTopic("topic", url)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to parse URL")
			})
		})
	})
}

func TestReplaceFirstPathSegment(t *testing.T) {
	testCases := []struct {
		name        string
		rawPath     string
		replaceWith string
		expected    string
	}{
		{
			name:        "path with multiple segments",
			rawPath:     "/path/with/multiple/segments",
			replaceWith: "new",
			expected:    "/new/with/multiple/segments",
		},
		{
			name:        "path without leading slash",
			rawPath:     "path/without/leading/slash",
			replaceWith: "new",
			expected:    "/new/without/leading/slash",
		},
		{
			name:        "root path",
			rawPath:     "/",
			replaceWith: "new",
			expected:    "/new",
		},
		{
			name:        "empty path",
			rawPath:     "",
			replaceWith: "new",
			expected:    "/new",
		},
	}

	for _, tc := range testCases {
		Convey(tc.name, t, func() {
			result := ReplaceFirstPathSegment(tc.rawPath, tc.replaceWith)
			So(result, ShouldEqual, tc.expected)
		})
	}
}
