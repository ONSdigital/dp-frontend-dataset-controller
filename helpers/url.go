package helpers

import (
	"fmt"
	"net/url"
	"path"
)

// PrefixPathWithTopic returns the path of the provided URL prefixed with the provided topicID
func PrefixPathWithTopic(topicID, rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	parsedURL.Path = path.Join("/", topicID, parsedURL.Path)
	return parsedURL.Path, nil
}
