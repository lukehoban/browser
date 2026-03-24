// Package js provides a JavaScript execution engine for the browser.
// It uses the goja ECMAScript engine to execute JavaScript found in
// <script> tags and provides DOM bindings for document manipulation.
//
// The engine implements a subset of the Web API:
//   - document.getElementById, getElementsByTagName, getElementsByClassName
//   - document.createElement, createTextNode
//   - document.body, document.documentElement
//   - Element: appendChild, removeChild, insertBefore, setAttribute, getAttribute
//   - Element: textContent, innerHTML, id, className, tagName, style
//   - Element: children, parentNode, parentElement, firstChild, lastChild
//   - Element: nextSibling, previousSibling
//   - console.log, console.warn, console.error
package js

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
	"github.com/lukehoban/browser/dom"
	"github.com/lukehoban/browser/log"
)

// Engine wraps the goja JavaScript runtime with DOM bindings.
type Engine struct {
	runtime *goja.Runtime
	doc     *dom.Node
	// nodeMap tracks DOM nodes to their JS object wrappers for identity preservation
	nodeMap map[*dom.Node]*goja.Object
}

// NewEngine creates a new JavaScript engine with DOM bindings for the given document.
func NewEngine(doc *dom.Node) *Engine {
	rt := goja.New()

	e := &Engine{
		runtime: rt,
		doc:     doc,
		nodeMap: make(map[*dom.Node]*goja.Object),
	}

	e.setupConsole()
	e.setupDocument()

	return e
}

// Execute runs a JavaScript program and returns any error.
func (e *Engine) Execute(script string) error {
	_, err := e.runtime.RunString(script)
	if err != nil {
		return fmt.Errorf("javascript error: %w", err)
	}
	return nil
}

// ExtractScripts extracts JavaScript source code from <script> tags in the DOM tree.
// It returns each script's content in document order. External scripts (with src
// attribute) are skipped with a warning.
func ExtractScripts(doc *dom.Node) []string {
	var scripts []string
	extractScriptsFromNode(doc, &scripts)
	return scripts
}

func extractScriptsFromNode(node *dom.Node, scripts *[]string) {
	if node.Type == dom.ElementNode && node.Data == "script" {
		// Skip external scripts (src attribute)
		if src := node.GetAttribute("src"); src != "" {
			log.Warnf("External script not supported: %s", src)
			return
		}
		// Skip non-JavaScript script types
		if scriptType := node.GetAttribute("type"); scriptType != "" &&
			scriptType != "text/javascript" &&
			scriptType != "application/javascript" {
			return
		}
		// Collect text content from the script element
		var content strings.Builder
		for _, child := range node.Children {
			if child.Type == dom.TextNode {
				content.WriteString(child.Data)
			}
		}
		if s := content.String(); strings.TrimSpace(s) != "" {
			*scripts = append(*scripts, s)
		}
		return
	}

	for _, child := range node.Children {
		extractScriptsFromNode(child, scripts)
	}
}

// setupConsole sets up the console object (console.log, console.warn, console.error).
func (e *Engine) setupConsole() {
	console := e.runtime.NewObject()
	_ = console.Set("log", func(call goja.FunctionCall) goja.Value {
		args := formatArgs(call)
		log.Infof("[JS console.log] %s", args)
		return goja.Undefined()
	})
	_ = console.Set("warn", func(call goja.FunctionCall) goja.Value {
		args := formatArgs(call)
		log.Warnf("[JS console.warn] %s", args)
		return goja.Undefined()
	})
	_ = console.Set("error", func(call goja.FunctionCall) goja.Value {
		args := formatArgs(call)
		log.Errorf("[JS console.error] %s", args)
		return goja.Undefined()
	})
	_ = e.runtime.Set("console", console)

	// Also provide alert() as a global
	_ = e.runtime.Set("alert", func(call goja.FunctionCall) goja.Value {
		args := formatArgs(call)
		log.Infof("[JS alert] %s", args)
		return goja.Undefined()
	})
}

// formatArgs formats goja function call arguments as a space-separated string.
func formatArgs(call goja.FunctionCall) string {
	parts := make([]string, len(call.Arguments))
	for i, arg := range call.Arguments {
		parts[i] = arg.String()
	}
	return strings.Join(parts, " ")
}

// setupDocument creates the global document object with DOM bindings.
func (e *Engine) setupDocument() {
	doc := e.runtime.NewDynamicObject(&documentAccessor{engine: e})
	_ = e.runtime.Set("document", doc)
}

// documentAccessor implements goja.DynamicObject for the document object.
type documentAccessor struct {
	engine *Engine
}

func (d *documentAccessor) Get(key string) goja.Value {
	e := d.engine
	switch key {
	case "getElementById":
		return e.runtime.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 1 {
				return goja.Null()
			}
			id := call.Arguments[0].String()
			node := e.doc.GetElementByID(id)
			if node == nil {
				return goja.Null()
			}
			return e.wrapNode(node)
		})
	case "getElementsByTagName":
		return e.runtime.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 1 {
				return e.runtime.NewArray()
			}
			tagName := call.Arguments[0].String()
			nodes := e.doc.GetElementsByTagName(tagName)
			return e.wrapNodeList(nodes)
		})
	case "getElementsByClassName":
		return e.runtime.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 1 {
				return e.runtime.NewArray()
			}
			className := call.Arguments[0].String()
			nodes := e.doc.GetElementsByClassName(className)
			return e.wrapNodeList(nodes)
		})
	case "createElement":
		return e.runtime.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 1 {
				return goja.Null()
			}
			tagName := strings.ToLower(call.Arguments[0].String())
			node := dom.NewElement(tagName)
			return e.wrapNode(node)
		})
	case "createTextNode":
		return e.runtime.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 1 {
				return goja.Null()
			}
			text := call.Arguments[0].String()
			node := dom.NewText(text)
			return e.wrapNode(node)
		})
	case "body":
		bodies := e.doc.GetElementsByTagName("body")
		if len(bodies) > 0 {
			return e.wrapNode(bodies[0])
		}
		return goja.Null()
	case "documentElement":
		for _, child := range e.doc.Children {
			if child.Type == dom.ElementNode && child.Data == "html" {
				return e.wrapNode(child)
			}
		}
		return goja.Null()
	case "head":
		heads := e.doc.GetElementsByTagName("head")
		if len(heads) > 0 {
			return e.wrapNode(heads[0])
		}
		return goja.Null()
	case "title":
		titles := e.doc.GetElementsByTagName("title")
		if len(titles) > 0 {
			return e.runtime.ToValue(titles[0].TextContent())
		}
		return e.runtime.ToValue("")
	}
	return goja.Undefined()
}

func (d *documentAccessor) Set(key string, val goja.Value) bool {
	return false
}

func (d *documentAccessor) Has(key string) bool {
	switch key {
	case "getElementById", "getElementsByTagName", "getElementsByClassName",
		"createElement", "createTextNode", "body", "documentElement", "head", "title":
		return true
	}
	return false
}

func (d *documentAccessor) Delete(key string) bool {
	return false
}

func (d *documentAccessor) Keys() []string {
	return []string{
		"getElementById", "getElementsByTagName", "getElementsByClassName",
		"createElement", "createTextNode", "body", "documentElement", "head", "title",
	}
}

// nodeAccessor implements goja.DynamicObject for DOM node wrappers.
type nodeAccessor struct {
	engine *Engine
	node   *dom.Node
}

func (n *nodeAccessor) Get(key string) goja.Value {
	e := n.engine
	node := n.node

	switch key {
	// Node type info
	case "nodeType":
		switch node.Type {
		case dom.ElementNode:
			return e.runtime.ToValue(1)
		case dom.TextNode:
			return e.runtime.ToValue(3)
		case dom.DocumentNode:
			return e.runtime.ToValue(9)
		}
		return e.runtime.ToValue(0)
	case "tagName", "nodeName":
		if node.Type == dom.ElementNode {
			return e.runtime.ToValue(strings.ToUpper(node.Data))
		}
		if node.Type == dom.TextNode {
			return e.runtime.ToValue("#text")
		}
		return goja.Undefined()

	// Properties
	case "id":
		return e.runtime.ToValue(node.GetAttribute("id"))
	case "className":
		return e.runtime.ToValue(node.GetAttribute("class"))
	case "textContent":
		return e.runtime.ToValue(node.TextContent())
	case "innerHTML":
		return e.runtime.ToValue(getInnerHTML(node))
	case "nodeValue":
		if node.Type == dom.TextNode {
			return e.runtime.ToValue(node.Data)
		}
		return goja.Null()

	// Tree navigation
	case "parentNode", "parentElement":
		if node.Parent == nil {
			return goja.Null()
		}
		if key == "parentElement" && node.Parent.Type != dom.ElementNode {
			return goja.Null()
		}
		return e.wrapNode(node.Parent)
	case "children":
		var elements []*dom.Node
		for _, child := range node.Children {
			if child.Type == dom.ElementNode {
				elements = append(elements, child)
			}
		}
		return e.wrapNodeList(elements)
	case "childNodes":
		return e.wrapNodeList(node.Children)
	case "firstChild":
		if len(node.Children) == 0 {
			return goja.Null()
		}
		return e.wrapNode(node.Children[0])
	case "lastChild":
		if len(node.Children) == 0 {
			return goja.Null()
		}
		return e.wrapNode(node.Children[len(node.Children)-1])
	case "firstElementChild":
		for _, child := range node.Children {
			if child.Type == dom.ElementNode {
				return e.wrapNode(child)
			}
		}
		return goja.Null()
	case "lastElementChild":
		for i := len(node.Children) - 1; i >= 0; i-- {
			if node.Children[i].Type == dom.ElementNode {
				return e.wrapNode(node.Children[i])
			}
		}
		return goja.Null()
	case "nextSibling":
		if s := nextSibling(node); s != nil {
			return e.wrapNode(s)
		}
		return goja.Null()
	case "previousSibling":
		if s := previousSibling(node); s != nil {
			return e.wrapNode(s)
		}
		return goja.Null()

	// Style property
	case "style":
		return e.runtime.NewDynamicObject(&styleAccessor{engine: e, node: node})

	// Methods
	case "appendChild":
		return e.runtime.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 1 {
				return goja.Null()
			}
			childNode := e.unwrapNode(call.Arguments[0])
			if childNode == nil {
				return goja.Null()
			}
			if childNode.Parent != nil {
				childNode.Parent.RemoveChild(childNode)
			}
			node.AppendChild(childNode)
			return e.wrapNode(childNode)
		})
	case "removeChild":
		return e.runtime.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 1 {
				return goja.Null()
			}
			childNode := e.unwrapNode(call.Arguments[0])
			if childNode == nil {
				return goja.Null()
			}
			removed := node.RemoveChild(childNode)
			if removed == nil {
				return goja.Null()
			}
			return e.wrapNode(removed)
		})
	case "insertBefore":
		return e.runtime.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 1 {
				return goja.Null()
			}
			newChild := e.unwrapNode(call.Arguments[0])
			if newChild == nil {
				return goja.Null()
			}
			var refChild *dom.Node
			if len(call.Arguments) >= 2 && !goja.IsNull(call.Arguments[1]) && !goja.IsUndefined(call.Arguments[1]) {
				refChild = e.unwrapNode(call.Arguments[1])
			}
			if newChild.Parent != nil {
				newChild.Parent.RemoveChild(newChild)
			}
			node.InsertBefore(newChild, refChild)
			return e.wrapNode(newChild)
		})
	case "setAttribute":
		return e.runtime.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 2 {
				return goja.Undefined()
			}
			name := call.Arguments[0].String()
			value := call.Arguments[1].String()
			node.SetAttribute(name, value)
			return goja.Undefined()
		})
	case "getAttribute":
		return e.runtime.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 1 {
				return goja.Null()
			}
			name := call.Arguments[0].String()
			if node.Attributes == nil {
				return goja.Null()
			}
			val, exists := node.Attributes[name]
			if !exists {
				return goja.Null()
			}
			return e.runtime.ToValue(val)
		})
	case "hasAttribute":
		return e.runtime.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 1 {
				return e.runtime.ToValue(false)
			}
			name := call.Arguments[0].String()
			if node.Attributes == nil {
				return e.runtime.ToValue(false)
			}
			_, exists := node.Attributes[name]
			return e.runtime.ToValue(exists)
		})
	case "removeAttribute":
		return e.runtime.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 1 {
				return goja.Undefined()
			}
			name := call.Arguments[0].String()
			if node.Attributes != nil {
				delete(node.Attributes, name)
			}
			return goja.Undefined()
		})
	case "getElementsByTagName":
		return e.runtime.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 1 {
				return e.runtime.NewArray()
			}
			tagName := call.Arguments[0].String()
			nodes := node.GetElementsByTagName(tagName)
			return e.wrapNodeList(nodes)
		})
	case "getElementsByClassName":
		return e.runtime.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 1 {
				return e.runtime.NewArray()
			}
			className := call.Arguments[0].String()
			nodes := node.GetElementsByClassName(className)
			return e.wrapNodeList(nodes)
		})
	case "hasChildNodes":
		return e.runtime.ToValue(func(call goja.FunctionCall) goja.Value {
			return e.runtime.ToValue(len(node.Children) > 0)
		})
	case "cloneNode":
		return e.runtime.ToValue(func(call goja.FunctionCall) goja.Value {
			deep := false
			if len(call.Arguments) >= 1 {
				deep = call.Arguments[0].ToBoolean()
			}
			clone := cloneNode(node, deep)
			return e.wrapNode(clone)
		})
	}

	return goja.Undefined()
}

func (n *nodeAccessor) Set(key string, val goja.Value) bool {
	node := n.node
	switch key {
	case "id":
		node.SetAttribute("id", val.String())
		return true
	case "className":
		node.SetAttribute("class", val.String())
		return true
	case "textContent":
		node.SetTextContent(val.String())
		return true
	case "innerHTML":
		setInnerHTML(node, val.String())
		return true
	case "nodeValue":
		if node.Type == dom.TextNode {
			node.Data = val.String()
			return true
		}
		return false
	}
	return false
}

func (n *nodeAccessor) Has(key string) bool {
	switch key {
	case "nodeType", "tagName", "nodeName", "id", "className", "textContent", "innerHTML",
		"nodeValue", "parentNode", "parentElement", "children", "childNodes",
		"firstChild", "lastChild", "firstElementChild", "lastElementChild",
		"nextSibling", "previousSibling", "style",
		"appendChild", "removeChild", "insertBefore",
		"setAttribute", "getAttribute", "hasAttribute", "removeAttribute",
		"getElementsByTagName", "getElementsByClassName",
		"hasChildNodes", "cloneNode":
		return true
	}
	return false
}

func (n *nodeAccessor) Delete(key string) bool {
	return false
}

func (n *nodeAccessor) Keys() []string {
	return []string{
		"nodeType", "tagName", "nodeName", "id", "className", "textContent", "innerHTML",
		"nodeValue", "parentNode", "parentElement", "children", "childNodes",
		"firstChild", "lastChild", "firstElementChild", "lastElementChild",
		"nextSibling", "previousSibling", "style",
		"appendChild", "removeChild", "insertBefore",
		"setAttribute", "getAttribute", "hasAttribute", "removeAttribute",
		"getElementsByTagName", "getElementsByClassName",
		"hasChildNodes", "cloneNode",
	}
}

// styleAccessor implements goja.DynamicObject for element.style.
type styleAccessor struct {
	engine *Engine
	node   *dom.Node
}

func (s *styleAccessor) Get(key string) goja.Value {
	cssKey := camelToCSSProperty(key)
	styles := parseInlineStyle(s.node.GetAttribute("style"))
	if val, ok := styles[cssKey]; ok {
		return s.engine.runtime.ToValue(val)
	}
	return s.engine.runtime.ToValue("")
}

func (s *styleAccessor) Set(key string, val goja.Value) bool {
	cssKey := camelToCSSProperty(key)
	value := val.String()
	styles := parseInlineStyle(s.node.GetAttribute("style"))
	if value == "" {
		delete(styles, cssKey)
	} else {
		styles[cssKey] = value
	}
	s.node.SetAttribute("style", serializeInlineStyle(styles))
	return true
}

func (s *styleAccessor) Has(key string) bool {
	cssKey := camelToCSSProperty(key)
	styles := parseInlineStyle(s.node.GetAttribute("style"))
	_, ok := styles[cssKey]
	return ok
}

func (s *styleAccessor) Delete(key string) bool {
	cssKey := camelToCSSProperty(key)
	styles := parseInlineStyle(s.node.GetAttribute("style"))
	delete(styles, cssKey)
	s.node.SetAttribute("style", serializeInlineStyle(styles))
	return true
}

func (s *styleAccessor) Keys() []string {
	styles := parseInlineStyle(s.node.GetAttribute("style"))
	keys := make([]string, 0, len(styles))
	for k := range styles {
		keys = append(keys, k)
	}
	return keys
}

// wrapNode wraps a dom.Node as a JavaScript DynamicObject with DOM-like properties and methods.
// It maintains identity: the same *dom.Node always returns the same JS object.
func (e *Engine) wrapNode(node *dom.Node) *goja.Object {
	if node == nil {
		return nil
	}

	// Check cache for identity preservation
	if obj, ok := e.nodeMap[node]; ok {
		return obj
	}

	obj := e.runtime.NewDynamicObject(&nodeAccessor{engine: e, node: node})
	e.nodeMap[node] = obj
	return obj
}

// wrapNodeList wraps a slice of dom.Nodes into a JavaScript array-like object.
func (e *Engine) wrapNodeList(nodes []*dom.Node) *goja.Object {
	arr := e.runtime.NewArray()
	for i, node := range nodes {
		_ = arr.Set(fmt.Sprintf("%d", i), e.wrapNode(node))
	}
	_ = arr.Set("length", len(nodes))
	return arr
}

// unwrapNode extracts the dom.Node from a wrapped JavaScript value.
func (e *Engine) unwrapNode(val goja.Value) *dom.Node {
	if val == nil || goja.IsNull(val) || goja.IsUndefined(val) {
		return nil
	}

	// Export the dynamic object and check if it's a nodeAccessor
	obj := val.ToObject(e.runtime)
	exported := obj.Export()
	if na, ok := exported.(*nodeAccessor); ok {
		return na.node
	}

	// Fallback: search the node map
	for node, wrapped := range e.nodeMap {
		if wrapped == obj {
			return node
		}
	}
	return nil
}

// Helper functions

// nextSibling returns the next sibling of the node, or nil.
func nextSibling(node *dom.Node) *dom.Node {
	if node.Parent == nil {
		return nil
	}
	for i, child := range node.Parent.Children {
		if child == node && i+1 < len(node.Parent.Children) {
			return node.Parent.Children[i+1]
		}
	}
	return nil
}

// previousSibling returns the previous sibling of the node, or nil.
func previousSibling(node *dom.Node) *dom.Node {
	if node.Parent == nil {
		return nil
	}
	for i, child := range node.Parent.Children {
		if child == node && i > 0 {
			return node.Parent.Children[i-1]
		}
	}
	return nil
}

// cloneNode creates a copy of the node. If deep is true, it also clones all descendants.
func cloneNode(node *dom.Node, deep bool) *dom.Node {
	clone := &dom.Node{
		Type: node.Type,
		Data: node.Data,
	}
	if node.Attributes != nil {
		clone.Attributes = make(map[string]string)
		for k, v := range node.Attributes {
			clone.Attributes[k] = v
		}
	}
	clone.Children = make([]*dom.Node, 0)
	if deep {
		for _, child := range node.Children {
			childClone := cloneNode(child, true)
			clone.AppendChild(childClone)
		}
	}
	return clone
}

// getInnerHTML returns a simple HTML serialization of the node's children.
func getInnerHTML(node *dom.Node) string {
	var sb strings.Builder
	for _, child := range node.Children {
		serializeNode(child, &sb)
	}
	return sb.String()
}

// serializeNode serializes a DOM node to HTML.
func serializeNode(node *dom.Node, sb *strings.Builder) {
	switch node.Type {
	case dom.TextNode:
		sb.WriteString(node.Data)
	case dom.ElementNode:
		sb.WriteString("<")
		sb.WriteString(node.Data)
		for name, value := range node.Attributes {
			sb.WriteString(fmt.Sprintf(` %s="%s"`, name, value))
		}
		sb.WriteString(">")
		for _, child := range node.Children {
			serializeNode(child, sb)
		}
		sb.WriteString(fmt.Sprintf("</%s>", node.Data))
	}
}

// setInnerHTML parses simple HTML and replaces the node's children.
func setInnerHTML(node *dom.Node, htmlStr string) {
	// Clear existing children
	for _, child := range node.Children {
		child.Parent = nil
	}
	node.Children = make([]*dom.Node, 0)

	if htmlStr == "" {
		return
	}

	// For simple text content (no tags), create a text node
	if !strings.Contains(htmlStr, "<") {
		node.AppendChild(dom.NewText(htmlStr))
		return
	}

	// Use a minimal inline parser for basic HTML fragments
	parseHTMLFragment(node, htmlStr)
}

// parseHTMLFragment does minimal HTML parsing for innerHTML.
func parseHTMLFragment(parent *dom.Node, htmlStr string) {
	pos := 0
	for pos < len(htmlStr) {
		if htmlStr[pos] == '<' {
			// Check for end tag
			if pos+1 < len(htmlStr) && htmlStr[pos+1] == '/' {
				// Skip end tag
				end := strings.IndexByte(htmlStr[pos:], '>')
				if end >= 0 {
					pos += end + 1
				} else {
					pos = len(htmlStr)
				}
				continue
			}
			// Start tag
			end := strings.IndexByte(htmlStr[pos:], '>')
			if end < 0 {
				break
			}
			tagContent := htmlStr[pos+1 : pos+end]
			parts := strings.Fields(tagContent)
			if len(parts) == 0 {
				pos += end + 1
				continue
			}
			tagName := strings.ToLower(strings.TrimRight(parts[0], "/"))
			elem := dom.NewElement(tagName)

			// Parse simple attributes
			for _, part := range parts[1:] {
				if eqIdx := strings.IndexByte(part, '='); eqIdx >= 0 {
					attrName := part[:eqIdx]
					attrValue := strings.Trim(part[eqIdx+1:], `"'`)
					elem.SetAttribute(attrName, attrValue)
				}
			}

			parent.AppendChild(elem)
			pos += end + 1

			// Find inner content and closing tag
			if !isHTMLVoidElement(tagName) {
				closeTag := fmt.Sprintf("</%s>", tagName)
				closeIdx := strings.Index(strings.ToLower(htmlStr[pos:]), closeTag)
				if closeIdx >= 0 {
					inner := htmlStr[pos : pos+closeIdx]
					if strings.Contains(inner, "<") {
						parseHTMLFragment(elem, inner)
					} else if inner != "" {
						elem.AppendChild(dom.NewText(inner))
					}
					pos += closeIdx + len(closeTag)
				}
			}
		} else {
			// Text content
			end := strings.IndexByte(htmlStr[pos:], '<')
			if end < 0 {
				end = len(htmlStr) - pos
			}
			text := htmlStr[pos : pos+end]
			if strings.TrimSpace(text) != "" {
				parent.AppendChild(dom.NewText(text))
			}
			pos += end
		}
	}
}

// isHTMLVoidElement checks if a tag is a void/self-closing element.
func isHTMLVoidElement(tagName string) bool {
	switch tagName {
	case "area", "base", "br", "col", "embed", "hr", "img", "input",
		"link", "meta", "param", "source", "track", "wbr":
		return true
	}
	return false
}

// parseInlineStyle parses a CSS inline style string into a map.
func parseInlineStyle(style string) map[string]string {
	styles := make(map[string]string)
	if style == "" {
		return styles
	}
	for _, decl := range strings.Split(style, ";") {
		decl = strings.TrimSpace(decl)
		if decl == "" {
			continue
		}
		colonIdx := strings.IndexByte(decl, ':')
		if colonIdx < 0 {
			continue
		}
		prop := strings.TrimSpace(decl[:colonIdx])
		val := strings.TrimSpace(decl[colonIdx+1:])
		if prop != "" {
			styles[prop] = val
		}
	}
	return styles
}

// serializeInlineStyle converts a style map back to a CSS inline style string.
func serializeInlineStyle(styles map[string]string) string {
	if len(styles) == 0 {
		return ""
	}
	var parts []string
	for prop, val := range styles {
		parts = append(parts, fmt.Sprintf("%s: %s", prop, val))
	}
	return strings.Join(parts, "; ")
}

// camelToCSSProperty converts a camelCase JavaScript property name to a CSS property name.
// For example: "backgroundColor" -> "background-color", "fontSize" -> "font-size".
func camelToCSSProperty(name string) string {
	var sb strings.Builder
	for i, ch := range name {
		if ch >= 'A' && ch <= 'Z' {
			if i > 0 {
				sb.WriteByte('-')
			}
			sb.WriteRune(ch + ('a' - 'A'))
		} else {
			sb.WriteRune(ch)
		}
	}
	return sb.String()
}
