// Package main provides a WebAssembly entry point for the browser.
// It exposes browser rendering functions to JavaScript via syscall/js.
//
// This allows the browser to run entirely in a web client by compiling
// to WebAssembly and loading in an HTML page.
package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"strings"
	"syscall/js"

	"github.com/lukehoban/browser/css"
	"github.com/lukehoban/browser/dom"
	"github.com/lukehoban/browser/html"
	"github.com/lukehoban/browser/layout"
	"github.com/lukehoban/browser/render"
	"github.com/lukehoban/browser/style"
)

// renderHTML is the main function exposed to JavaScript.
// It takes HTML content, width, and height, and returns a base64-encoded PNG.
func renderHTML(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return map[string]interface{}{
			"error": "renderHTML requires 3 arguments: htmlContent (string), width (int), height (int)",
		}
	}

	htmlContent := args[0].String()
	width := args[1].Int()
	height := args[2].Int()

	// Parse HTML
	doc := html.Parse(htmlContent)

	// Extract CSS from <style> tags
	cssContent := extractCSS(doc)

	// Parse CSS
	stylesheet := css.Parse(cssContent)

	// Compute styles
	styledTree := style.StyleTree(doc, stylesheet)

	// Build layout tree
	containingBlock := layout.Dimensions{
		Content: layout.Rect{
			Width:  float64(width),
			Height: 0,
		},
	}
	layoutTree := layout.LayoutTree(styledTree, containingBlock)

	// Render to canvas
	canvas := render.Render(layoutTree, width, height)

	// Encode PNG to base64
	var buf bytes.Buffer
	img := canvas.ToImage()
	if err := png.Encode(&buf, img); err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("Failed to encode PNG: %v", err),
		}
	}

	base64Img := base64.StdEncoding.EncodeToString(buf.Bytes())

	return map[string]interface{}{
		"success": true,
		"image":   base64Img,
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

func main() {
	// Register the renderHTML function with JavaScript
	js.Global().Set("renderHTML", js.FuncOf(renderHTML))

	// Signal that Go is ready
	js.Global().Set("goReady", true)

	// Keep the Go program running
	<-make(chan bool)
}
