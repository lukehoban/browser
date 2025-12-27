// Package dom provides URL resolution for the Document Object Model.
// This handles resolving relative URLs against a base URL as per HTML5 §2.5 URLs.
package dom

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/lukehoban/browser/log"
)

// ResolveURLs resolves all relative URLs in a DOM tree against a base URL.
// For now, this handles file system paths. In the future, this should handle
// full URL resolution as per HTML5 §2.5 URLs.
//
// HTML5 §2.5: A URL is a string used to identify a resource.
// HTML5 §2.5.1: The document's base URL is used to resolve relative URLs.
func ResolveURLs(root *Node, baseDir string) {
	resolveNode(root, baseDir)
}

// resolveNode recursively resolves URLs in a node and its children.
func resolveNode(node *Node, baseDir string) {
	if node == nil {
		return
	}

	// Only process element nodes
	if node.Type == ElementNode {
		// Handle img elements - resolve the src attribute
		if node.Data == "img" {
			if src := node.GetAttribute("src"); src != "" {
				resolvedPath := resolveURL(baseDir, src)
				node.SetAttribute("src", resolvedPath)
			}
		}

		// Handle link elements - resolve the href attribute
		// HTML5 §4.2.4: The link element allows authors to link their document to other resources
		if node.Data == "link" {
			if href := node.GetAttribute("href"); href != "" {
				resolvedPath := resolveURL(baseDir, href)
				node.SetAttribute("href", resolvedPath)
			}
		}

		// Future: Handle other elements with URL attributes
		// - <script src="...">
		// - <a href="...">
		// - CSS background-image
	}

	// Recursively process children
	for _, child := range node.Children {
		resolveNode(child, baseDir)
	}
}

// resolveURL resolves a potentially relative URL against a base URL.
// HTML5 §2.5: URLs in documents are resolved against a base URL.
func resolveURL(baseURL, relativeURL string) string {
	return ResolveURLString(baseURL, relativeURL)
}

// ResolveURLString resolves a potentially relative URL against a base URL.
// This is exported for use by other packages that need to resolve URLs.
// HTML5 §2.5: URLs in documents are resolved against a base URL.
func ResolveURLString(baseURL, relativeURL string) string {
	// If the URL is already absolute (http:// or https://), return as-is
	if strings.HasPrefix(relativeURL, "http://") || strings.HasPrefix(relativeURL, "https://") {
		return relativeURL
	}

	// If base is a URL, do proper URL resolution
	if strings.HasPrefix(baseURL, "http://") || strings.HasPrefix(baseURL, "https://") {
		base, err := url.Parse(baseURL)
		if err != nil {
			log.Warnf("Failed to parse base URL '%s': %v", baseURL, err)
			return relativeURL
		}
		rel, err := url.Parse(relativeURL)
		if err != nil {
			log.Warnf("Failed to parse relative URL '%s': %v", relativeURL, err)
			return relativeURL
		}
		return base.ResolveReference(rel).String()
	}

	// Otherwise, treat as file paths
	return filepath.Join(baseURL, relativeURL)
}

// FetchExternalStylesheets finds all <link rel="stylesheet"> elements in the DOM tree
// and fetches their CSS content.
// HTML5 §4.2.4: The link element allows authors to link their document to other resources.
// This should ideally be done during HTML parsing, but for simplicity we do it post-parse.
func FetchExternalStylesheets(root *Node) string {
	loader := NewResourceLoader("")
	var cssBuilder strings.Builder
	fetchStylesheetsFromNode(root, loader, &cssBuilder)
	return cssBuilder.String()
}

// fetchStylesheetsFromNode recursively finds and fetches external stylesheets.
func fetchStylesheetsFromNode(node *Node, loader *ResourceLoader, builder *strings.Builder) {
	if node == nil {
		return
	}

	if node.Type == ElementNode && node.Data == "link" {
		rel := node.GetAttribute("rel")
		href := node.GetAttribute("href")

		// HTML5 §4.2.4: rel="stylesheet" indicates the linked resource is a stylesheet
		if rel == "stylesheet" && href != "" {
			// The href should already be resolved by ResolveURLs
			cssContent, err := loader.LoadResourceAsString(href)
			if err != nil {
				// Skip failed stylesheets (non-blocking per HTML5 spec)
				log.Warnf("Failed to load external stylesheet '%s': %v", href, err)
				return
			}
			builder.WriteString(cssContent)
			builder.WriteString("\n")
		}
	}

	for _, child := range node.Children {
		fetchStylesheetsFromNode(child, loader, builder)
	}
}
