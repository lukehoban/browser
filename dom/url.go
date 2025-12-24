// Package dom provides URL resolution for the Document Object Model.
// This handles resolving relative URLs against a base URL as per HTML5 §2.5 URLs.
package dom

import (
	"path/filepath"
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
				// Resolve relative path to absolute path
				resolvedPath := filepath.Join(baseDir, src)
				node.SetAttribute("src", resolvedPath)
			}
		}

		// Future: Handle other elements with URL attributes
		// - <link href="...">
		// - <script src="...">
		// - <a href="...">
		// - CSS background-image
	}

	// Recursively process children
	for _, child := range node.Children {
		resolveNode(child, baseDir)
	}
}
