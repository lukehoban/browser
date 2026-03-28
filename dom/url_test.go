package dom

import (
	"path/filepath"
	"testing"
)

func TestResolveURLs(t *testing.T) {
	// Create a simple DOM tree with img elements
	doc := NewDocument()
	body := NewElement("body")
	doc.AppendChild(body)

	// Add an img with relative path
	img1 := NewElement("img")
	img1.SetAttribute("src", "logo.png")
	body.AppendChild(img1)

	// Add an img with path in subdirectory
	img2 := NewElement("img")
	img2.SetAttribute("src", "images/icon.png")
	body.AppendChild(img2)

	// Resolve URLs against /home/test
	baseDir := "/home/test"
	ResolveURLs(doc, baseDir)

	// Check that img1 src was resolved
	expectedPath1 := filepath.Join(baseDir, "logo.png")
	if img1.GetAttribute("src") != expectedPath1 {
		t.Errorf("expected src=%s, got %s", expectedPath1, img1.GetAttribute("src"))
	}

	// Check that img2 src was resolved
	expectedPath2 := filepath.Join(baseDir, "images/icon.png")
	if img2.GetAttribute("src") != expectedPath2 {
		t.Errorf("expected src=%s, got %s", expectedPath2, img2.GetAttribute("src"))
	}
}

func TestResolveURLsNestedElements(t *testing.T) {
	// Create a nested DOM tree
	doc := NewDocument()
	html := NewElement("html")
	doc.AppendChild(html)

	body := NewElement("body")
	html.AppendChild(body)

	div := NewElement("div")
	body.AppendChild(div)

	img := NewElement("img")
	img.SetAttribute("src", "test.png")
	div.AppendChild(img)

	// Resolve URLs
	baseDir := "/var/www"
	ResolveURLs(doc, baseDir)

	// Check that nested img src was resolved
	expectedPath := filepath.Join(baseDir, "test.png")
	if img.GetAttribute("src") != expectedPath {
		t.Errorf("expected src=%s, got %s", expectedPath, img.GetAttribute("src"))
	}
}

func TestResolveURLsNoSrc(t *testing.T) {
	// Create an img without src attribute
	doc := NewDocument()
	img := NewElement("img")
	img.SetAttribute("alt", "test")
	doc.AppendChild(img)

	// Resolve URLs - should not panic
	ResolveURLs(doc, "/home/test")

	// Check that alt attribute is unchanged
	if img.GetAttribute("alt") != "test" {
		t.Errorf("expected alt=test, got %s", img.GetAttribute("alt"))
	}
}

func TestResolveURLsNonImgElements(t *testing.T) {
	// Create elements that are not img
	doc := NewDocument()
	div := NewElement("div")
	div.SetAttribute("data-src", "test.png") // Not a src attribute
	doc.AppendChild(div)

	// Resolve URLs
	ResolveURLs(doc, "/home/test")

	// Check that data-src is unchanged (we only resolve src on img elements)
	if div.GetAttribute("data-src") != "test.png" {
		t.Errorf("expected data-src=test.png, got %s", div.GetAttribute("data-src"))
	}
}

func TestResolveURLsLinkElement(t *testing.T) {
	doc := NewDocument()
	link := NewElement("link")
	link.SetAttribute("rel", "stylesheet")
	link.SetAttribute("href", "style.css")
	doc.AppendChild(link)

	ResolveURLs(doc, "/var/www")

	expected := filepath.Join("/var/www", "style.css")
	if link.GetAttribute("href") != expected {
		t.Errorf("expected href=%s, got %s", expected, link.GetAttribute("href"))
	}
}

func TestResolveURLString(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		relativeURL string
		expected    string
	}{
		{
			name:        "data URL passthrough",
			baseURL:     "/home/test",
			relativeURL: "data:text/plain;base64,SGVsbG8=",
			expected:    "data:text/plain;base64,SGVsbG8=",
		},
		{
			name:        "absolute http URL passthrough",
			baseURL:     "/home/test",
			relativeURL: "http://example.com/image.png",
			expected:    "http://example.com/image.png",
		},
		{
			name:        "absolute https URL passthrough",
			baseURL:     "/home/test",
			relativeURL: "https://example.com/style.css",
			expected:    "https://example.com/style.css",
		},
		{
			name:        "relative path with file base",
			baseURL:     "/home/test",
			relativeURL: "image.png",
			expected:    filepath.Join("/home/test", "image.png"),
		},
		{
			name:        "relative path with subdirectory",
			baseURL:     "/var/www",
			relativeURL: "images/logo.png",
			expected:    filepath.Join("/var/www", "images/logo.png"),
		},
		{
			name:        "HTTP base with relative URL",
			baseURL:     "http://example.com/page/",
			relativeURL: "image.png",
			expected:    "http://example.com/page/image.png",
		},
		{
			name:        "HTTPS base with relative URL",
			baseURL:     "https://example.com/",
			relativeURL: "style.css",
			expected:    "https://example.com/style.css",
		},
		{
			name:        "HTTP base with path traversal",
			baseURL:     "http://example.com/dir/page.html",
			relativeURL: "../images/logo.png",
			expected:    "http://example.com/images/logo.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveURLString(tt.baseURL, tt.relativeURL)
			if result != tt.expected {
				t.Errorf("ResolveURLString(%q, %q) = %q, want %q",
					tt.baseURL, tt.relativeURL, result, tt.expected)
			}
		})
	}
}

func TestFetchExternalStylesheets(t *testing.T) {
	// Test with no link elements
	doc := NewDocument()
	body := NewElement("body")
	doc.AppendChild(body)

	result := FetchExternalStylesheets(doc)
	if result != "" {
		t.Errorf("FetchExternalStylesheets() with no links = %q, want %q", result, "")
	}
}

func TestFetchExternalStylesheets_NonStylesheet(t *testing.T) {
	// Test with a link element that is not a stylesheet
	doc := NewDocument()
	link := NewElement("link")
	link.SetAttribute("rel", "icon")
	link.SetAttribute("href", "favicon.ico")
	doc.AppendChild(link)

	result := FetchExternalStylesheets(doc)
	if result != "" {
		t.Errorf("FetchExternalStylesheets() with non-stylesheet link = %q, want empty", result)
	}
}

func TestResolveURLs_NilNode(t *testing.T) {
	// Should not panic on nil node
	ResolveURLs(nil, "/home/test")
}
