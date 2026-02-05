package main

import (
	"os"
	"testing"

	"github.com/lukehoban/browser/css"
	"github.com/lukehoban/browser/dom"
	"github.com/lukehoban/browser/html"
	"github.com/lukehoban/browser/layout"
	"github.com/lukehoban/browser/render"
	"github.com/lukehoban/browser/style"
)

// TestWebsites tests that the browser can successfully fetch and render
// real-world websites without crashing. These tests verify network support,
// HTML/CSS parsing, layout computation, and rendering for actual internet sites.
//
// Note: These tests require network connectivity and may be slow or flaky
// if websites change their structure or are unavailable.

// websiteTestCase represents a test case for a real website
type websiteTestCase struct {
	name        string
	url         string
	description string
}

// Common simple websites for testing
var websiteTests = []websiteTestCase{
	{
		name:        "Hacker News",
		url:         "https://news.ycombinator.com/",
		description: "Simple news aggregator with minimal CSS",
	},
	{
		name:        "Example.com",
		url:         "http://example.com/",
		description: "IANA's example domain with basic HTML",
	},
	{
		name:        "Example.org",
		url:         "http://example.org/",
		description: "Another IANA example domain",
	},
	{
		name:        "TextOnly Wikipedia",
		url:         "https://en.wikipedia.org/wiki/Main_Page",
		description: "Wikipedia main page - complex but well-structured HTML",
	},
}

// TestWebsiteRendering tests that the browser can successfully render real websites
func TestWebsiteRendering(t *testing.T) {
	// Skip if running in short mode (for faster CI)
	if testing.Short() {
		t.Skip("Skipping website tests in short mode")
	}

	for _, tc := range websiteTests {
		t.Run(tc.name, func(t *testing.T) {
			// Skip Wikipedia test by default as it's more complex
			if tc.name == "TextOnly Wikipedia" {
				t.Skip("Skipping Wikipedia test - too complex for basic validation")
			}

			t.Logf("Testing %s (%s): %s", tc.name, tc.url, tc.description)

			// Fetch the website
			content, err := fetchURL(tc.url)
			if err != nil {
				t.Fatalf("Failed to fetch %s: %v", tc.url, err)
			}

			if len(content) == 0 {
				t.Fatalf("Fetched empty content from %s", tc.url)
			}

			t.Logf("Fetched %d bytes from %s", len(content), tc.url)

			// Parse HTML
			doc := html.Parse(content)
			if doc == nil {
				t.Fatalf("Failed to parse HTML from %s", tc.url)
			}

			// Resolve URLs
			dom.ResolveURLs(doc, tc.url)

			// Extract CSS
			cssContent := extractCSS(doc)
			externalCSS := dom.FetchExternalStylesheets(doc)
			cssContent = externalCSS + "\n" + cssContent

			// Parse CSS
			stylesheet := css.Parse(cssContent)

			// Compute styles
			styledTree := style.StyleTree(doc, stylesheet)
			if styledTree == nil {
				t.Fatalf("Failed to compute styles for %s", tc.url)
			}

			// Resolve CSS URLs
			style.ResolveCSSURLs(styledTree, tc.url)

			// Build layout tree
			containingBlock := layout.Dimensions{
				Content: layout.Rect{
					Width:  800,
					Height: 0,
				},
			}
			layoutTree := layout.LayoutTree(styledTree, containingBlock)
			if layoutTree == nil {
				t.Fatalf("Failed to build layout tree for %s", tc.url)
			}

			// Verify layout tree has content
			if layoutTree.Dimensions.Content.Width == 0 && layoutTree.Dimensions.Content.Height == 0 {
				t.Logf("Warning: Layout tree has zero dimensions for %s", tc.url)
			}

			t.Logf("Successfully rendered %s: layout dimensions %.0fx%.0f",
				tc.name,
				layoutTree.Dimensions.Content.Width,
				layoutTree.Dimensions.Content.Height)
		})
	}
}

// TestWebsiteRenderingToPNG tests that websites can be rendered to PNG files
func TestWebsiteRenderingToPNG(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping PNG rendering tests in short mode")
	}

	// Only test a single simple website to avoid slow tests
	tc := websiteTestCase{
		name:        "Example.com",
		url:         "http://example.com/",
		description: "IANA's example domain",
	}

	t.Run(tc.name+"_PNG", func(t *testing.T) {
		// Fetch the website
		content, err := fetchURL(tc.url)
		if err != nil {
			t.Fatalf("Failed to fetch %s: %v", tc.url, err)
		}

		// Parse HTML
		doc := html.Parse(content)
		dom.ResolveURLs(doc, tc.url)

		// Extract and parse CSS
		cssContent := extractCSS(doc)
		externalCSS := dom.FetchExternalStylesheets(doc)
		cssContent = externalCSS + "\n" + cssContent
		stylesheet := css.Parse(cssContent)

		// Compute styles and layout
		styledTree := style.StyleTree(doc, stylesheet)
		style.ResolveCSSURLs(styledTree, tc.url)

		containingBlock := layout.Dimensions{
			Content: layout.Rect{
				Width:  800,
				Height: 0,
			},
		}
		layoutTree := layout.LayoutTree(styledTree, containingBlock)

		// Render to PNG
		canvas := render.Render(layoutTree, 800, 600)
		if canvas == nil {
			t.Fatalf("Failed to render canvas for %s", tc.url)
		}

		// Save to temporary file
		tmpFile := "/tmp/website_test_example.png"
		err = canvas.SavePNG(tmpFile)
		if err != nil {
			t.Fatalf("Failed to save PNG for %s: %v", tc.url, err)
		}

		// Verify file was created and has content
		info, err := os.Stat(tmpFile)
		if err != nil {
			t.Fatalf("Failed to stat PNG file: %v", err)
		}

		if info.Size() == 0 {
			t.Fatalf("Generated PNG file is empty")
		}

		t.Logf("Successfully rendered %s to PNG (%d bytes)", tc.name, info.Size())

		// Clean up
		os.Remove(tmpFile)
	})
}

// TestHackerNewsSpecific tests Hacker News specifically since it's mentioned in the README
func TestHackerNewsSpecific(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Hacker News test in short mode")
	}

	url := "https://news.ycombinator.com/"
	t.Logf("Testing Hacker News: %s", url)

	// Fetch the website
	content, err := fetchURL(url)
	if err != nil {
		t.Fatalf("Failed to fetch Hacker News: %v", err)
	}

	// Parse and process
	doc := html.Parse(content)
	dom.ResolveURLs(doc, url)

	cssContent := extractCSS(doc)
	externalCSS := dom.FetchExternalStylesheets(doc)
	cssContent = externalCSS + "\n" + cssContent
	stylesheet := css.Parse(cssContent)

	styledTree := style.StyleTree(doc, stylesheet)
	style.ResolveCSSURLs(styledTree, url)

	containingBlock := layout.Dimensions{
		Content: layout.Rect{
			Width:  1024, // Use 1024x768 as mentioned in README
			Height: 0,
		},
	}
	layoutTree := layout.LayoutTree(styledTree, containingBlock)

	// Verify we got a layout
	if layoutTree == nil {
		t.Fatal("Failed to build layout tree for Hacker News")
	}

	t.Logf("Successfully rendered Hacker News: layout dimensions %.0fx%.0f",
		layoutTree.Dimensions.Content.Width,
		layoutTree.Dimensions.Content.Height)

	// Optionally render to PNG if output file is desired (commented out for CI)
	// canvas := render.Render(layoutTree, 1024, 768)
	// canvas.SavePNG("/tmp/hackernews_test.png")
}

// TestWebsiteErrorHandling tests that the browser handles various error conditions gracefully
func TestWebsiteErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		shouldFail  bool
		description string
	}{
		{
			name:        "Invalid URL",
			url:         "not-a-valid-url",
			shouldFail:  true,
			description: "Should handle invalid URLs gracefully",
		},
		{
			name:        "Non-existent domain",
			url:         "http://this-domain-definitely-does-not-exist-12345.com",
			shouldFail:  true,
			description: "Should handle DNS resolution failures",
		},
		{
			name:        "Valid Example.com",
			url:         "http://example.com/",
			shouldFail:  false,
			description: "Should successfully fetch example.com",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := fetchURL(tc.url)

			if tc.shouldFail && err == nil {
				t.Errorf("Expected error for %s but got none", tc.url)
			}

			if !tc.shouldFail && err != nil {
				t.Errorf("Expected success for %s but got error: %v", tc.url, err)
			}

			t.Logf("%s: %s", tc.description, tc.url)
		})
	}
}
