// Package dom provides resource loading functionality for the Document Object Model.
// This handles fetching resources (HTML, CSS, images) from network or filesystem.
//
// Spec references:
// - HTML5 §2.5 URLs: URL resolution and resource fetching
// - HTML5 §12.2.8.3: The end (loading resources during parsing)
package dom

import (
	"fmt"
	"io"
	"net/http"
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
func (rl *ResourceLoader) LoadResource(path string) ([]byte, error) {
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
