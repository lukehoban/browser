// Package main provides the browser command-line application.
// It parses HTML and CSS, computes styles, calculates layout, and renders to PNG.
//
// Network support:
// - HTTP/HTTPS URL fetching follows standard Go net/http practices
// - HTML5 §2.5 URLs: Relative URL resolution against base URL
// - External stylesheet loading via <link rel="stylesheet">
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/lukehoban/browser/css"
	"github.com/lukehoban/browser/dom"
	"github.com/lukehoban/browser/html"
	"github.com/lukehoban/browser/layout"
	"github.com/lukehoban/browser/render"
	"github.com/lukehoban/browser/style"
)

func main() {
	// Parse command-line flags
	outputFile := flag.String("output", "", "Output PNG file path (optional)")
	width := flag.Int("width", 800, "Viewport width in pixels")
	height := flag.Int("height", 600, "Viewport height in pixels")
	flag.Parse()

	// Check for input file or URL
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: browser [options] <html-file-or-url>\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	input := args[0]

	// Determine if input is a URL or file path
	var content string
	var baseURL string
	var err error

	if isURL(input) {
		// Fetch from network
		fmt.Fprintf(os.Stderr, "Fetching from URL: %s\n", input)
		content, err = fetchURL(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching URL: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Fetched %d bytes\n", len(content))
		baseURL = input
	} else {
		// Read from local file
		fileContent, err := os.ReadFile(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		content = string(fileContent)
		baseURL = filepath.Dir(input)
	}

	// Parse HTML
	fmt.Fprintf(os.Stderr, "Parsing HTML...\n")
	doc := html.Parse(content)
	fmt.Fprintf(os.Stderr, "HTML parsed\n")

	// Resolve relative URLs (e.g., image paths) against the document's base URL
	// HTML5 §2.5: URLs in documents are resolved against a base URL
	fmt.Fprintf(os.Stderr, "Resolving URLs...\n")
	dom.ResolveURLs(doc, baseURL)
	fmt.Fprintf(os.Stderr, "URLs resolved\n")

	// Extract CSS from <style> tags and <link> tags
	fmt.Fprintf(os.Stderr, "Extracting CSS...\n")
	cssContent := extractCSS(doc)

	// Fetch external stylesheets
	// HTML5 §4.2.4: External stylesheets are fetched via <link rel="stylesheet">
	fmt.Fprintf(os.Stderr, "Fetching external stylesheets...\n")
	externalCSS := dom.FetchExternalStylesheets(doc)
	cssContent = externalCSS + "\n" + cssContent
	fmt.Fprintf(os.Stderr, "External stylesheets fetched\n")

	// Parse CSS
	fmt.Fprintf(os.Stderr, "Parsing CSS...\n")
	stylesheet := css.Parse(cssContent)
	fmt.Fprintf(os.Stderr, "CSS parsed\n")

	// Compute styles
	styledTree := style.StyleTree(doc, stylesheet)

	// Build layout tree
	// Note: Height starts at 0 - it accumulates as children are laid out
	containingBlock := layout.Dimensions{
		Content: layout.Rect{
			Width:  float64(*width),
			Height: 0,
		},
	}
	layoutTree := layout.LayoutTree(styledTree, containingBlock)

	// Print summary
	fmt.Println("=== Browser Rendering ===")
	fmt.Printf("Input: %s\n", input)
	fmt.Printf("Viewport: %dx%d\n", *width, *height)

	// Render to PNG if output specified
	if *outputFile != "" {
		canvas := render.Render(layoutTree, *width, *height)
		if err := canvas.SavePNG(*outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving PNG: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Output: %s\n", *outputFile)
		fmt.Println("Rendering complete!")
	} else {
		// Print layout tree for debugging
		fmt.Println("\n=== Layout Tree ===")
		printLayoutTree(layoutTree, 0)
	}
}

// extractCSS extracts CSS from <style> tags in the document.
func extractCSS(doc *dom.Node) string {
	var cssBuilder strings.Builder
	extractCSSFromNode(doc, &cssBuilder)
	return cssBuilder.String()
}

// extractCSSFromNode recursively extracts CSS from style elements.
func extractCSSFromNode(node *dom.Node, builder *strings.Builder) {
	if node.Type == dom.ElementNode && node.Data == "style" {
		for _, child := range node.Children {
			if child.Type == dom.TextNode {
				builder.WriteString(child.Data)
				builder.WriteString("\n")
			}
		}
	}

	for _, child := range node.Children {
		extractCSSFromNode(child, builder)
	}
}

// printLayoutTree prints the layout tree for debugging.
func printLayoutTree(box *layout.LayoutBox, indent int) {
	prefix := strings.Repeat("  ", indent)

	boxType := "Block"
	if box.BoxType == layout.InlineBox {
		boxType = "Inline"
	} else if box.BoxType == layout.AnonymousBox {
		boxType = "Anonymous"
	}

	nodeName := "?"
	if box.StyledNode != nil && box.StyledNode.Node != nil {
		nodeName = box.StyledNode.Node.Data
	}

	fmt.Printf("%s%s <%s> [x:%.0f y:%.0f w:%.0f h:%.0f]\n",
		prefix, boxType, nodeName,
		box.Dimensions.Content.X,
		box.Dimensions.Content.Y,
		box.Dimensions.Content.Width,
		box.Dimensions.Content.Height,
	)

	for _, child := range box.Children {
		printLayoutTree(child, indent+1)
	}
}

// isURL checks if the input string is a URL (http:// or https://)
func isURL(input string) bool {
	return strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://")
}

// fetchURL fetches content from a URL and returns it as a string
func fetchURL(urlStr string) (string, error) {
	resp, err := http.Get(urlStr)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}
