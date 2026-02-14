package util

import (
	"fmt"
	"net/url"
	"strings"
)

// GetDomain takes a raw URL string and returns the hostname.
func GetDomain(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// u.Hostname() strips the port (e.g., :8080) if it exists
	host := u.Hostname()

	// If the URL didn't have a scheme (like "google.com" instead of "https://google.com"),
	// url.Parse might put the domain in the Path instead of the Host.
	if host == "" && strings.Contains(rawURL, ".") {
		return "", fmt.Errorf("invalid URL: missing scheme (e.g., https://)")
	}

	return host, nil
}

// GetPath takes a raw URL string and returns the path.
func GetPath(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	path := parsed.Path
	if path == "" {
		path = "/"
	}

	// If we need to add query params to path, use this
	//if parsed.RawQuery != "" {
	//	path = path + "?" + parsed.RawQuery
	//}

	return path, nil
}
