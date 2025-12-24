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
