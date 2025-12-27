// Package dom provides resource loading functionality for the Document Object Model.
// This handles fetching resources (HTML, CSS, images) from network or filesystem.
//
// Spec references:
// - HTML5 §2.5 URLs: URL resolution and resource fetching
// - HTML5 §12.2.8.3: The end (loading resources during parsing)
// - RFC 2397: The "data" URL scheme
package dom

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// ResourceLoader handles loading resources from URLs or file paths.
type ResourceLoader struct {
	BaseURL string
}

// NewResourceLoader creates a new resource loader with the given base URL.
func NewResourceLoader(baseURL string) *ResourceLoader {
	return &ResourceLoader{
		BaseURL: baseURL,
	}
}

// LoadResource loads content from a URL or file path.
// HTML5 §2.5: Resources are identified by URLs and can be fetched over network or from filesystem.
// RFC 2397: Supports data URLs (data:[<mediatype>][;base64],<data>)
func (rl *ResourceLoader) LoadResource(path string) ([]byte, error) {
	// RFC 2397: Handle data URLs
	if isDataURL(path) {
		return loadFromDataURL(path)
	}
	if isURL(path) {
		return loadFromURL(path)
	}
	return os.ReadFile(path)
}

// LoadResourceAsString loads content as a string.
func (rl *ResourceLoader) LoadResourceAsString(path string) (string, error) {
	data, err := rl.LoadResource(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// isURL checks if the input string is a URL (http:// or https://).
func isURL(input string) bool {
	return strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://")
}

// isDataURL checks if the input string is a data URL.
// RFC 2397: data URLs have the format data:[<mediatype>][;base64],<data>
func isDataURL(input string) bool {
	return strings.HasPrefix(input, "data:")
}

// loadFromURL fetches content from a URL.
func loadFromURL(urlStr string) ([]byte, error) {
	resp, err := http.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// loadFromDataURL decodes a data URL and returns its content.
// RFC 2397: data:[<mediatype>][;base64],<data>
// Examples:
//   - data:text/plain;base64,SGVsbG8sIFdvcmxkIQ==
//   - data:image/svg+xml,%3Csvg...%3E%3C/svg%3E
//   - data:image/png;base64,iVBORw0KGgo...
func loadFromDataURL(dataURL string) ([]byte, error) {
	// Parse the data URL
	parsedURL, err := url.Parse(dataURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse data URL: %w", err)
	}

	if parsedURL.Scheme != "data" {
		return nil, fmt.Errorf("not a data URL")
	}

	// The opaque part contains [<mediatype>][;base64],<data>
	dataStr := parsedURL.Opaque

	// Find the comma that separates metadata from data
	commaIdx := strings.Index(dataStr, ",")
	if commaIdx == -1 {
		return nil, fmt.Errorf("invalid data URL: missing comma")
	}

	metadata := dataStr[:commaIdx]
	data := dataStr[commaIdx+1:]

	// Check if base64 encoded
	isBase64 := strings.HasSuffix(metadata, ";base64")

	if isBase64 {
		// Decode base64 data
		decoded, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 data: %w", err)
		}
		return decoded, nil
	}

	// Otherwise, URL decode the data
	decoded, err := url.QueryUnescape(data)
	if err != nil {
		return nil, fmt.Errorf("failed to URL decode data: %w", err)
	}
	return []byte(decoded), nil
}
