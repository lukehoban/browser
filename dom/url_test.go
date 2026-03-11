package dom

import (
	"os"
	"path/filepath"
	"strings"
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

func TestResolveURLStringDataURL(t *testing.T) {
dataURL := "data:image/png;base64,abc123"
result := ResolveURLString("/base/dir", dataURL)
if result != dataURL {
t.Errorf("Data URLs should not be resolved, got %q", result)
}
}

func TestResolveURLStringAbsoluteHTTP(t *testing.T) {
absURL := "http://example.com/image.png"
result := ResolveURLString("/base/dir", absURL)
if result != absURL {
t.Errorf("Absolute HTTP URLs should not be changed, got %q", result)
}
}

func TestResolveURLStringAbsoluteHTTPS(t *testing.T) {
absURL := "https://example.com/style.css"
result := ResolveURLString("/base/dir", absURL)
if result != absURL {
t.Errorf("Absolute HTTPS URLs should not be changed, got %q", result)
}
}

func TestResolveURLStringRelativeWithHTTPBase(t *testing.T) {
base := "http://example.com/pages/"
rel := "images/logo.png"
result := ResolveURLString(base, rel)
expected := "http://example.com/pages/images/logo.png"
if result != expected {
t.Errorf("Expected %q, got %q", expected, result)
}
}

func TestResolveURLStringRelativeWithHTTPBaseAbsolutePath(t *testing.T) {
base := "http://example.com/pages/index.html"
rel := "/images/logo.png"
result := ResolveURLString(base, rel)
expected := "http://example.com/images/logo.png"
if result != expected {
t.Errorf("Expected %q, got %q", expected, result)
}
}

func TestResolveURLStringFilePath(t *testing.T) {
base := "/home/user/project"
rel := "images/logo.png"
result := ResolveURLString(base, rel)
expected := "/home/user/project/images/logo.png"
if result != expected {
t.Errorf("Expected %q, got %q", expected, result)
}
}

func TestResolveURLsLinkElement(t *testing.T) {
doc := NewDocument()
head := NewElement("head")
link := NewElement("link")
link.SetAttribute("href", "style.css")
head.AppendChild(link)
doc.AppendChild(head)

ResolveURLs(doc, "/home/test")

got := link.GetAttribute("href")
expected := "/home/test/style.css"
if got != expected {
t.Errorf("Expected href=%q, got %q", expected, got)
}
}

func TestResolveURLsNilNode(t *testing.T) {
// Should not panic on nil node
ResolveURLs(nil, "/base")
}

func TestFetchExternalStylesheets(t *testing.T) {
// Create a temporary CSS file
tmpDir := "/tmp/test_fetch_css"
os.MkdirAll(tmpDir, 0755)
defer os.RemoveAll(tmpDir)

cssContent := "body { color: red; }"
cssFile := tmpDir + "/style.css"
os.WriteFile(cssFile, []byte(cssContent), 0644)

// Create DOM tree with a link element
doc := NewDocument()
head := NewElement("head")
link := NewElement("link")
link.SetAttribute("rel", "stylesheet")
link.SetAttribute("href", cssFile)
head.AppendChild(link)
doc.AppendChild(head)

result := FetchExternalStylesheets(doc)
if result == "" {
t.Error("Expected CSS content from external stylesheet")
}
if !strings.Contains(result, "color: red") {
t.Errorf("Expected 'color: red' in result, got %q", result)
}
}

func TestFetchExternalStylesheetsNoLink(t *testing.T) {
doc := NewDocument()
body := NewElement("body")
doc.AppendChild(body)

result := FetchExternalStylesheets(doc)
if result != "" {
t.Errorf("Expected empty result for no link elements, got %q", result)
}
}

func TestFetchExternalStylesheetsNonStylesheet(t *testing.T) {
doc := NewDocument()
head := NewElement("head")
link := NewElement("link")
link.SetAttribute("rel", "icon")
link.SetAttribute("href", "/favicon.ico")
head.AppendChild(link)
doc.AppendChild(head)

result := FetchExternalStylesheets(doc)
if result != "" {
t.Errorf("Expected empty result for non-stylesheet link, got %q", result)
}
}

func TestFetchExternalStylesheetsNilNode(t *testing.T) {
result := FetchExternalStylesheets(nil)
if result != "" {
t.Errorf("Expected empty result for nil node, got %q", result)
}
}

func TestFetchExternalStylesheetsMissingFile(t *testing.T) {
doc := NewDocument()
head := NewElement("head")
link := NewElement("link")
link.SetAttribute("rel", "stylesheet")
link.SetAttribute("href", "/nonexistent/style.css")
head.AppendChild(link)
doc.AppendChild(head)

// Should not panic, just skip the missing file
result := FetchExternalStylesheets(doc)
if result != "" {
t.Errorf("Expected empty result for missing file, got %q", result)
}
}
