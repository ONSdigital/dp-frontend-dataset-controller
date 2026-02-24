package helpers

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

// PrefixPathWithTopic returns the path of rawURL prefixed with topicID.
// The first path segment (assumed to be the API Router version) is always removed
// because frontend routes do not include the API version prefix.
func PrefixPathWithTopic(topicID, rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Clean path and trim leading slash
	cleanPath := path.Clean(parsedURL.Path)
	cleanPath = strings.TrimPrefix(cleanPath, "/")

	// Remove first segment of the path
	segments := strings.Split(cleanPath, "/")
	if len(segments) > 0 {
		segments = segments[1:]
	}

	return path.Join("/", topicID, path.Join(segments...)), nil
}
