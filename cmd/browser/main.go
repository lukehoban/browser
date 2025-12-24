// Command browser parses HTML files and displays the parsed DOM tree,
// style computations, and layout information.
package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/lukehoban/browser/css"
	"github.com/lukehoban/browser/dom"
	"github.com/lukehoban/browser/html"
	"github.com/lukehoban/browser/layout"
	"github.com/lukehoban/browser/style"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: browser <html-file>")
		os.Exit(1)
	}

	filename := os.Args[1]
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	htmlContent := string(content)

	// Parse HTML
	fmt.Println("=== Parsing HTML ===")
	doc := html.Parse(htmlContent)
	fmt.Println("DOM tree parsed successfully.")

	// Print DOM tree summary
	fmt.Println("\n=== DOM Tree ===")
	printDOMTree(doc, 0)

	// Extract and parse CSS from <style> tags
	fmt.Println("\n=== Parsing CSS ===")
	cssContent := extractCSS(htmlContent)
	var stylesheet *css.Stylesheet
	if cssContent != "" {
		stylesheet = css.Parse(cssContent)
		fmt.Printf("Found %d CSS rules.\n", len(stylesheet.Rules))
	} else {
		stylesheet = &css.Stylesheet{Rules: []*css.Rule{}}
		fmt.Println("No embedded CSS found.")
	}

	// Compute styles
	fmt.Println("\n=== Computing Styles ===")
	styledTree := style.StyleTree(doc, stylesheet)
	fmt.Println("Styles computed successfully.")

	// Print styled tree
	fmt.Println("\n=== Styled Tree ===")
	printStyledTree(styledTree, 0)

	// Create layout
	fmt.Println("\n=== Creating Layout ===")
	containingBlock := layout.Dimensions{
		Content: layout.Rect{X: 0, Y: 0, Width: 800, Height: 0},
	}
	layoutTree := layout.LayoutTree(styledTree, containingBlock)
	fmt.Println("Layout computed successfully.")

	// Print layout tree
	fmt.Println("\n=== Layout Tree ===")
	printLayoutTree(layoutTree, 0)

	fmt.Println("\n=== Done ===")
}

// printDOMTree prints a DOM tree with indentation.
func printDOMTree(node *dom.Node, indent int) {
	prefix := strings.Repeat("  ", indent)

	switch node.Type {
	case dom.DocumentNode:
		fmt.Printf("%s[Document]\n", prefix)
	case dom.ElementNode:
		attrs := ""
		if id := node.GetAttribute("id"); id != "" {
			attrs += fmt.Sprintf(" id=%q", id)
		}
		if class := node.GetAttribute("class"); class != "" {
			attrs += fmt.Sprintf(" class=%q", class)
		}
		fmt.Printf("%s<%s%s>\n", prefix, node.Data, attrs)
	case dom.TextNode:
		text := strings.TrimSpace(node.Data)
		if text != "" {
			if len(text) > 50 {
				text = text[:47] + "..."
			}
			fmt.Printf("%s\"%s\"\n", prefix, text)
		}
	}

	for _, child := range node.Children {
		printDOMTree(child, indent+1)
	}
}

// printStyledTree prints a styled tree with computed styles.
func printStyledTree(node *style.StyledNode, indent int) {
	prefix := strings.Repeat("  ", indent)

	if node.Node.Type == dom.ElementNode {
		fmt.Printf("%s<%s>", prefix, node.Node.Data)
		if len(node.Styles) > 0 {
			fmt.Printf(" [%d styles]", len(node.Styles))
		}
		fmt.Println()
	}

	for _, child := range node.Children {
		printStyledTree(child, indent+1)
	}
}

// printLayoutTree prints a layout tree with dimensions.
func printLayoutTree(box *layout.LayoutBox, indent int) {
	prefix := strings.Repeat("  ", indent)

	boxTypeName := "block"
	switch box.BoxType {
	case layout.InlineBox:
		boxTypeName = "inline"
	case layout.AnonymousBox:
		boxTypeName = "anonymous"
	}

	tagName := ""
	if box.StyledNode != nil && box.StyledNode.Node != nil {
		tagName = box.StyledNode.Node.Data
	}

	fmt.Printf("%s[%s] <%s> x=%.0f y=%.0f w=%.0f h=%.0f\n",
		prefix, boxTypeName, tagName,
		box.Dimensions.Content.X,
		box.Dimensions.Content.Y,
		box.Dimensions.Content.Width,
		box.Dimensions.Content.Height)

	for _, child := range box.Children {
		printLayoutTree(child, indent+1)
	}
}

// extractCSS extracts CSS content from <style> tags in HTML.
func extractCSS(htmlContent string) string {
	re := regexp.MustCompile(`(?is)<style[^>]*>(.*?)</style>`)
	matches := re.FindAllStringSubmatch(htmlContent, -1)

	var cssContent strings.Builder
	for _, match := range matches {
		if len(match) > 1 {
			cssContent.WriteString(match[1])
			cssContent.WriteString("\n")
		}
	}
	return cssContent.String()
}
