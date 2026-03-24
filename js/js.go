// Package js provides a simple JavaScript interpreter for executing scripts
// embedded in HTML documents. It supports a practical subset of ECMAScript 5/6
// including variables, closures, control flow, and DOM manipulation.
//
// The interpreter executes <script> tags found in the DOM, allowing JavaScript
// to modify the document before it is laid out and rendered.
//
// Supported language features:
//   - Variable declarations: var, let, const
//   - Functions: declarations, expressions, arrow functions, closures
//   - Control flow: if/else, while, do/while, for, for...in, for...of, switch
//   - Operators: arithmetic, comparison, logical, bitwise, assignment
//   - Exception handling: try/catch/finally, throw
//   - Classes: basic ES6 class syntax
//   - Template literals: `hello ${name}`
//   - Built-in objects: Math, JSON, Array, Object, String, Number, Boolean
//   - Collections: Map, Set
//   - Promises: basic synchronous execution
//
// Supported DOM APIs:
//   - document.getElementById, querySelector, querySelectorAll
//   - document.createElement, createTextNode
//   - element.innerHTML, textContent, innerText (get/set)
//   - element.style.propertyName (camelCase)
//   - element.classList.add/remove/toggle/contains
//   - element.setAttribute, getAttribute, removeAttribute
//   - element.appendChild, removeChild, insertBefore
//   - element.dataset for data-* attributes
//   - console.log, warn, error, info
package js

import (
	"strings"

	"github.com/lukehoban/browser/dom"
	browserlog "github.com/lukehoban/browser/log"
)

// Execute finds all <script> tags in doc and executes their JavaScript content.
// The script may modify the DOM tree in place (e.g., setting innerHTML, appending nodes).
// Scripts are executed in order of appearance.
//
// Any JavaScript runtime errors are logged as warnings but do not abort execution of
// subsequent scripts.
func Execute(doc *dom.Node) {
	interp := NewInterpreter()
	interp.SetupDOM(doc)

	// Also set `window` to point to a global-like object with document
	if win, ok := interp.global.Get("window"); ok && win.typ == TypeObject {
		if docVal, ok := interp.global.Get("document"); ok {
			win.objVal.Set("document", docVal)
		}
	}

	// Collect all script node content before executing
	// (executing one script may modify the DOM)
	scripts := collectScripts(doc)

	for _, script := range scripts {
		if script == "" {
			continue
		}
		browserlog.Debugf("Executing script (%d chars)", len(script))
		if err := executeScript(interp, script); err != nil {
			browserlog.Warnf("JavaScript error: %v", err)
		}
	}
}

// executeScript parses and evaluates a single JavaScript string.
func executeScript(interp *Interpreter, src string) error {
	parser := NewParser(src)
	prog, err := parser.ParseProgram()
	if err != nil {
		return err
	}
	return interp.Eval(prog)
}

// collectScripts extracts the text content of all <script> tags in the document.
// Inline scripts are returned; external scripts (with src=) are skipped.
func collectScripts(node *dom.Node) []string {
	var scripts []string
	collectScriptsFrom(node, &scripts)
	return scripts
}

func collectScriptsFrom(node *dom.Node, scripts *[]string) {
	if node.Type == dom.ElementNode && node.Data == "script" {
		// Skip external scripts (src attribute)
		if node.GetAttribute("src") != "" {
			return
		}
		// Skip non-JS scripts (type != text/javascript and type != module)
		scriptType := strings.ToLower(node.GetAttribute("type"))
		if scriptType != "" && scriptType != "text/javascript" && scriptType != "application/javascript" && scriptType != "module" {
			return
		}
		// Extract text content
		var content strings.Builder
		for _, child := range node.Children {
			if child.Type == dom.TextNode {
				content.WriteString(child.Data)
			}
		}
		text := strings.TrimSpace(content.String())
		if text != "" {
			*scripts = append(*scripts, text)
		}
		return
	}
	for _, child := range node.Children {
		collectScriptsFrom(child, scripts)
	}
}

// SetProperty intercepts property writes on DOM element wrappers.
// This is called when the interpreter assigns to elem.propertyName.
// It handles special properties like textContent and innerHTML.
func (interp *Interpreter) setPropertyOnDOMWrapper(obj *Object, key string, val *Value) bool {
	// Check if this object is a DOM wrapper
	nodeRef := obj.Get("__domNode__")
	if nodeRef.IsUndefined() {
		return false
	}
	nodeIDVal := nodeRef.ToObject()
	if nodeIDVal == nil {
		return false
	}
	nodeID := int(nodeIDVal.Get("__nodeID__").ToNumber())
	node := nodeRegistry[nodeID]
	if node == nil {
		return false
	}

	switch key {
	case "textContent", "innerText":
		setTextContent(node, val.ToString())
		obj.props["textContent"] = StringVal(val.ToString())
		obj.props["innerText"] = StringVal(val.ToString())
		obj.props["innerHTML"] = StringVal(getInnerHTML(node))
		return true
	case "innerHTML":
		setInnerHTML(node, val.ToString())
		obj.props["innerHTML"] = StringVal(val.ToString())
		obj.props["textContent"] = StringVal(getTextContent(node))
		obj.props["children"] = interp.makeChildrenArray(node)
		return true
	case "id":
		node.SetAttribute("id", val.ToString())
		obj.props["id"] = StringVal(val.ToString())
		return true
	case "className":
		node.SetAttribute("class", val.ToString())
		obj.props["className"] = StringVal(val.ToString())
		obj.props["classList"] = interp.makeClassList(node)
		return true
	case "hidden":
		if val.ToBoolean() {
			node.SetAttribute("hidden", "")
		} else {
			delete(node.Attributes, "hidden")
		}
		obj.props["hidden"] = BoolVal(val.ToBoolean())
		return true
	case "value":
		node.SetAttribute("value", val.ToString())
		obj.props["value"] = StringVal(val.ToString())
		return true
	case "href":
		node.SetAttribute("href", val.ToString())
		obj.props["href"] = StringVal(val.ToString())
		return true
	case "src":
		node.SetAttribute("src", val.ToString())
		obj.props["src"] = StringVal(val.ToString())
		return true
	case "disabled":
		if val.ToBoolean() {
			node.SetAttribute("disabled", "")
		} else {
			delete(node.Attributes, "disabled")
		}
		obj.props["disabled"] = BoolVal(val.ToBoolean())
		return true
	case "checked":
		if val.ToBoolean() {
			node.SetAttribute("checked", "")
		} else {
			delete(node.Attributes, "checked")
		}
		obj.props["checked"] = BoolVal(val.ToBoolean())
		return true
	}
	return false
}
