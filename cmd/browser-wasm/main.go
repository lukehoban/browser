//go:build js && wasm
// +build js,wasm

// Package main provides a WebAssembly entry point for the browser.
// It exposes browser rendering functions to JavaScript via syscall/js.
//
// This allows the browser to run entirely in a web client by compiling
// to WebAssembly and loading in an HTML page.
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"strings"
	"syscall/js"

	"github.com/lukehoban/browser/css"
	"github.com/lukehoban/browser/dom"
	"github.com/lukehoban/browser/html"
	"github.com/lukehoban/browser/layout"
	"github.com/lukehoban/browser/log"
	"github.com/lukehoban/browser/render"
	"github.com/lukehoban/browser/style"
)

// pageLogWriter sends log messages to the web page via JavaScript callback
type pageLogWriter struct{}

func (w *pageLogWriter) Write(p []byte) (n int, err error) {
	// Check if the JavaScript callback exists
	if logCallback := js.Global().Get("logToPage"); !logCallback.IsUndefined() {
		logCallback.Invoke(string(p))
	}
	return len(p), nil
}

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

	// Validate dimensions
	if width <= 0 || height <= 0 {
		return map[string]interface{}{
			"error": "Width and height must be positive",
		}
	}
	if width > 10000 || height > 10000 {
		return map[string]interface{}{
			"error": "Width and height must not exceed 10000 pixels",
		}
	}

	// Parse HTML (html.Parse doesn't return errors, it's resilient)
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

// getLayoutTree generates a JSON representation of the layout tree
func getLayoutTree(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return map[string]interface{}{
			"error": "getLayoutTree requires 3 arguments: htmlContent (string), width (int), height (int)",
		}
	}

	htmlContent := args[0].String()
	width := args[1].Int()
	_ = args[2].Int() // height - not used for layout tree building

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

	// Convert layout tree to JSON
	treeData := layoutBoxToMap(layoutTree)
	jsonData, err := json.Marshal(treeData)
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("Failed to serialize layout tree: %v", err),
		}
	}

	return map[string]interface{}{
		"success": true,
		"tree":    string(jsonData),
	}
}

// getRenderTree generates a JSON representation of the render tree (styled nodes)
func getRenderTree(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return map[string]interface{}{
			"error": "getRenderTree requires 1 argument: htmlContent (string)",
		}
	}

	htmlContent := args[0].String()

	// Parse HTML
	doc := html.Parse(htmlContent)

	// Extract CSS from <style> tags
	cssContent := extractCSS(doc)

	// Parse CSS
	stylesheet := css.Parse(cssContent)

	// Compute styles
	styledTree := style.StyleTree(doc, stylesheet)

	// Convert styled tree to JSON
	treeData := styledNodeToMap(styledTree)
	jsonData, err := json.Marshal(treeData)
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("Failed to serialize render tree: %v", err),
		}
	}

	return map[string]interface{}{
		"success": true,
		"tree":    string(jsonData),
	}
}

// setLogLevel configures the logging level from JavaScript
func setLogLevel(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return map[string]interface{}{
			"error": "setLogLevel requires 1 argument: level (string)",
		}
	}

	level := strings.ToLower(args[0].String())
	switch level {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		return map[string]interface{}{
			"error": fmt.Sprintf("Invalid log level: %s. Use debug, info, warn, or error", level),
		}
	}

	return map[string]interface{}{
		"success": true,
		"level":   level,
	}
}

// layoutBoxToMap converts a layout box to a map for JSON serialization
func layoutBoxToMap(box *layout.LayoutBox) map[string]interface{} {
	if box == nil {
		return nil
	}

	boxType := "Block"
	switch box.BoxType {
	case layout.InlineBox:
		boxType = "Inline"
	case layout.AnonymousBox:
		boxType = "Anonymous"
	case layout.TableBox:
		boxType = "Table"
	case layout.TableRowBox:
		boxType = "TableRow"
	case layout.TableCellBox:
		boxType = "TableCell"
	}

	nodeName := "?"
	if box.StyledNode != nil && box.StyledNode.Node != nil {
		nodeName = box.StyledNode.Node.Data
	}

	result := map[string]interface{}{
		"type":   boxType,
		"node":   nodeName,
		"x":      box.Dimensions.Content.X,
		"y":      box.Dimensions.Content.Y,
		"width":  box.Dimensions.Content.Width,
		"height": box.Dimensions.Content.Height,
	}

	if len(box.Children) > 0 {
		children := make([]map[string]interface{}, 0, len(box.Children))
		for _, child := range box.Children {
			children = append(children, layoutBoxToMap(child))
		}
		result["children"] = children
	}

	return result
}

// styledNodeToMap converts a styled node to a map for JSON serialization
func styledNodeToMap(node *style.StyledNode) map[string]interface{} {
	if node == nil {
		return nil
	}

	nodeType := "Unknown"
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
		case dom.DocumentNode:
			nodeType = "Document"
			nodeData = "document"
		}
	}

	result := map[string]interface{}{
		"type": nodeType,
		"data": nodeData,
	}

	if len(node.Styles) > 0 {
		result["styles"] = node.Styles
	}

	if len(node.Children) > 0 {
		children := make([]map[string]interface{}, 0, len(node.Children))
		for _, child := range node.Children {
			children = append(children, styledNodeToMap(child))
		}
		result["children"] = children
	}

	return result
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
	// Set up page logging for WASM
	log.SetOutput(&pageLogWriter{})
	log.SetLevel(log.WarnLevel) // Default level

	// Register the renderHTML function with JavaScript
	js.Global().Set("renderHTML", js.FuncOf(renderHTML))
	
	// Register layout and render tree functions
	js.Global().Set("getLayoutTree", js.FuncOf(getLayoutTree))
	js.Global().Set("getRenderTree", js.FuncOf(getRenderTree))
	
	// Register log level control function
	js.Global().Set("setLogLevel", js.FuncOf(setLogLevel))

	// Signal that Go is ready
	js.Global().Set("goReady", true)

	// Keep the Go program running indefinitely
	select {}
}
