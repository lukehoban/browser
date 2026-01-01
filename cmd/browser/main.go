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
	"sort"
	"strings"

	"github.com/lukehoban/browser/css"
	"github.com/lukehoban/browser/dom"
	"github.com/lukehoban/browser/html"
	"github.com/lukehoban/browser/layout"
	"github.com/lukehoban/browser/log"
	"github.com/lukehoban/browser/render"
	"github.com/lukehoban/browser/style"
)

const (
	// maxDisplayedStyles is the maximum number of styles to display per layout box
	// to keep the output readable while still providing useful debugging information.
	maxDisplayedStyles = 8
)

// importantStyles lists CSS properties that are most relevant for layout debugging,
// shown in priority order when displaying computed styles in the layout tree.
var importantStyles = []string{
	"display",
	"width",
	"height",
	"color",
	"background-color",
	"font-size",
	"font-weight",
	"font-style",
}

// importantStylesMap provides O(1) lookup for important styles.
// Initialized once to avoid repeated allocations.
var importantStylesMap = func() map[string]bool {
	m := make(map[string]bool)
	for _, key := range importantStyles {
		m[key] = true
	}
	return m
}()

func main() {
	// Parse command-line flags
	outputFile := flag.String("output", "", "Output PNG file path (optional)")
	width := flag.Int("width", 800, "Viewport width in pixels")
	height := flag.Int("height", 600, "Viewport height in pixels")
	logLevel := flag.String("log-level", "warn", "Log level: debug, info, warn, error")
	verbose := flag.Bool("verbose", false, "Enable verbose logging (equivalent to -log-level=info)")
	showLayout := flag.Bool("show-layout", false, "Display layout tree instead of rendering")
	showRender := flag.Bool("show-render", false, "Display render tree (styled nodes) instead of rendering")
	flag.Parse()

	// Configure logging
	if *verbose {
		log.SetLevel(log.InfoLevel)
	} else {
		switch strings.ToLower(*logLevel) {
		case "debug":
			log.SetLevel(log.DebugLevel)
		case "info":
			log.SetLevel(log.InfoLevel)
		case "warn":
			log.SetLevel(log.WarnLevel)
		case "error":
			log.SetLevel(log.ErrorLevel)
		default:
			log.SetLevel(log.WarnLevel)
		}
	}

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

	// Resolve CSS URLs (like background-image) against base URL
	// HTML5 §2.5.1: URLs should be resolved against the document's base URL
	style.ResolveCSSURLs(styledTree, baseURL)

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

	// Display render tree if requested
	if *showRender {
		fmt.Println("\n=== Render Tree (Styled Nodes) ===")
		printRenderTree(styledTree, 0)
		return
	}

	// Display layout tree if requested
	if *showLayout {
		fmt.Println("\n=== Layout Tree ===")
		printLayoutTree(layoutTree, 0)
		return
	}

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
// Each node displays its type, name, dimensions, and computed styles.
func printLayoutTree(box *layout.LayoutBox, indent int) {
	prefix := strings.Repeat("  ", indent)

	boxType := "Block"
	if box.BoxType == layout.InlineBox {
		boxType = "Inline"
	} else if box.BoxType == layout.AnonymousBox {
		boxType = "Anonymous"
	} else if box.BoxType == layout.TableBox {
		boxType = "Table"
	} else if box.BoxType == layout.TableRowBox {
		boxType = "TableRow"
	} else if box.BoxType == layout.TableCellBox {
		boxType = "TableCell"
	}

	nodeName := "?"
	if box.StyledNode != nil && box.StyledNode.Node != nil {
		nodeName = box.StyledNode.Node.Data
	}

	// Format the basic layout info
	layoutInfo := fmt.Sprintf("%s%s <%s> [x:%.0f y:%.0f w:%.0f h:%.0f]",
		prefix, boxType, nodeName,
		box.Dimensions.Content.X,
		box.Dimensions.Content.Y,
		box.Dimensions.Content.Width,
		box.Dimensions.Content.Height,
	)

	// Add computed styles if available
	if box.StyledNode != nil && len(box.StyledNode.Styles) > 0 {
		layoutInfo += " {"
		styleCount := 0

		// First, show important styles in order
		for _, key := range importantStyles {
			if value, ok := box.StyledNode.Styles[key]; ok {
				if styleCount > 0 {
					layoutInfo += ", "
				}
				layoutInfo += fmt.Sprintf("%s:%s", key, value)
				styleCount++
				if styleCount >= maxDisplayedStyles {
					break
				}
			}
		}

		// If we haven't hit the limit, show additional styles in sorted order
		if styleCount < maxDisplayedStyles {
			// Collect remaining style keys with pre-allocated capacity
			remainingKeys := make([]string, 0, len(box.StyledNode.Styles))
			for key := range box.StyledNode.Styles {
				// Skip if already shown (O(1) lookup)
				if !importantStylesMap[key] {
					remainingKeys = append(remainingKeys, key)
				}
			}
			// Sort for deterministic output
			sort.Strings(remainingKeys)

			// Display remaining styles in sorted order
			for _, key := range remainingKeys {
				if styleCount > 0 {
					layoutInfo += ", "
				}
				layoutInfo += fmt.Sprintf("%s:%s", key, box.StyledNode.Styles[key])
				styleCount++
				if styleCount >= maxDisplayedStyles {
					break
				}
			}
		}

		// Indicate if there are more styles than what we displayed
		if len(box.StyledNode.Styles) > styleCount {
			layoutInfo += ", ..."
		}
		layoutInfo += "}"
	}

	fmt.Println(layoutInfo)

	for _, child := range box.Children {
		printLayoutTree(child, indent+1)
	}
}

// printRenderTree prints the render tree (styled nodes) for debugging.
func printRenderTree(node *style.StyledNode, indent int) {
	if node == nil {
		return
	}

	prefix := strings.Repeat("  ", indent)

	// Get node info
	nodeType := "?"
	nodeData := ""
	if node.Node != nil {
		switch node.Node.Type {
		case dom.ElementNode:
			nodeType = "Element"
			nodeData = node.Node.Data
		case dom.TextNode:
			nodeType = "Text"
			nodeData = node.Node.Data
			if len(nodeData) > 40 {
				nodeData = nodeData[:40] + "..."
			}
			// Escape newlines for display
			nodeData = strings.ReplaceAll(nodeData, "\n", "\\n")
			nodeData = strings.ReplaceAll(nodeData, "\t", "\\t")
		case dom.DocumentNode:
			nodeType = "Document"
			nodeData = "document"
		}
	}

	// Print node with key styles
	fmt.Printf("%s%s: %s", prefix, nodeType, nodeData)
	if len(node.Styles) > 0 {
		fmt.Printf(" {")
		count := 0
		maxStyles := 5
		for k, v := range node.Styles {
			if count < maxStyles {
				if count > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("%s:%s", k, v)
				count++
			}
		}
		if len(node.Styles) > maxStyles {
			fmt.Printf(", ...")
		}
		fmt.Printf("}")
	}
	fmt.Println()

	// Recursively print children
	for _, child := range node.Children {
		printRenderTree(child, indent+1)
	}
}

// isURL checks if the input string is a URL (http:// or https://)
func isURL(input string) bool {
	return strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://")
}

// fetchURL fetches content from a URL and returns it as a string
func fetchURL(urlStr string) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow redirects but maintain user-agent
			req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GoBrowser/1.0; +https://github.com/lukehoban/browser)")
			return nil
		},
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set a proper user-agent to avoid being blocked by websites like Wikipedia
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GoBrowser/1.0; +https://github.com/lukehoban/browser)")

	resp, err := client.Do(req)
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
