package js

import (
	"strings"
	"testing"

	"github.com/lukehoban/browser/dom"
	"github.com/lukehoban/browser/html"
	"github.com/lukehoban/browser/log"
)

func init() {
	// Enable info-level logging so console.log messages are captured
	log.SetLevel(log.InfoLevel)
}

func TestExtractScripts(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected int
	}{
		{
			name:     "no scripts",
			html:     "<html><body><p>Hello</p></body></html>",
			expected: 0,
		},
		{
			name:     "single script",
			html:     `<html><body><script>var x = 1;</script></body></html>`,
			expected: 1,
		},
		{
			name:     "multiple scripts",
			html:     `<html><head><script>var a = 1;</script></head><body><script>var b = 2;</script></body></html>`,
			expected: 2,
		},
		{
			name:     "skip external script",
			html:     `<html><body><script src="app.js"></script></body></html>`,
			expected: 0,
		},
		{
			name:     "skip non-js type",
			html:     `<html><body><script type="application/json">{"key": "value"}</script></body></html>`,
			expected: 0,
		},
		{
			name:     "explicit js type",
			html:     `<html><body><script type="text/javascript">var x = 1;</script></body></html>`,
			expected: 1,
		},
		{
			name:     "empty script",
			html:     `<html><body><script>   </script></body></html>`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := html.Parse(tt.html)
			scripts := ExtractScripts(doc)
			if len(scripts) != tt.expected {
				t.Errorf("expected %d scripts, got %d", tt.expected, len(scripts))
			}
		})
	}
}

func TestEngineExecute(t *testing.T) {
	doc := html.Parse("<html><body></body></html>")
	engine := NewEngine(doc)

	// Simple expression
	if err := engine.Execute("var x = 1 + 2;"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Syntax error
	if err := engine.Execute("var = ;"); err == nil {
		t.Fatal("expected error for invalid syntax")
	}
}

func TestDocumentGetElementByID(t *testing.T) {
	doc := html.Parse(`<html><body><div id="main">Hello</div></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var el = document.getElementById("main");
		if (!el) throw new Error("element not found");
		if (el.tagName !== "DIV") throw new Error("expected DIV, got " + el.tagName);
		if (el.id !== "main") throw new Error("expected id=main, got " + el.id);
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDocumentGetElementByIDNotFound(t *testing.T) {
	doc := html.Parse(`<html><body></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var el = document.getElementById("nonexistent");
		if (el !== null) throw new Error("expected null");
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDocumentCreateElement(t *testing.T) {
	doc := html.Parse(`<html><body></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var div = document.createElement("div");
		if (div.tagName !== "DIV") throw new Error("expected DIV, got " + div.tagName);
		if (div.nodeType !== 1) throw new Error("expected nodeType 1, got " + div.nodeType);
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDocumentCreateTextNode(t *testing.T) {
	doc := html.Parse(`<html><body></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var text = document.createTextNode("Hello World");
		if (text.nodeType !== 3) throw new Error("expected nodeType 3, got " + text.nodeType);
		if (text.nodeValue !== "Hello World") throw new Error("expected nodeValue 'Hello World'");
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppendChild(t *testing.T) {
	doc := html.Parse(`<html><body><div id="container"></div></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var container = document.getElementById("container");
		var p = document.createElement("p");
		p.textContent = "Added by JavaScript";
		container.appendChild(p);
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the DOM was modified
	container := doc.GetElementByID("container")
	if container == nil {
		t.Fatal("container not found")
	}
	if len(container.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(container.Children))
	}
	p := container.Children[0]
	if p.Data != "p" {
		t.Errorf("expected <p>, got <%s>", p.Data)
	}
	if p.TextContent() != "Added by JavaScript" {
		t.Errorf("expected text 'Added by JavaScript', got %q", p.TextContent())
	}
}

func TestRemoveChild(t *testing.T) {
	doc := html.Parse(`<html><body><div id="parent"><p id="child">Remove me</p></div></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var parent = document.getElementById("parent");
		var child = document.getElementById("child");
		parent.removeChild(child);
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parent := doc.GetElementByID("parent")
	if parent == nil {
		t.Fatal("parent not found")
	}
	// Should have no element children after removal
	elementChildren := 0
	for _, c := range parent.Children {
		if c.Type == dom.ElementNode {
			elementChildren++
		}
	}
	if elementChildren != 0 {
		t.Errorf("expected 0 element children, got %d", elementChildren)
	}
}

func TestSetAttribute(t *testing.T) {
	doc := html.Parse(`<html><body><div id="target"></div></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var el = document.getElementById("target");
		el.setAttribute("data-value", "42");
		el.setAttribute("class", "highlight");
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	target := doc.GetElementByID("target")
	if target == nil {
		t.Fatal("target not found")
	}
	if target.GetAttribute("data-value") != "42" {
		t.Errorf("expected data-value=42, got %q", target.GetAttribute("data-value"))
	}
	if target.GetAttribute("class") != "highlight" {
		t.Errorf("expected class=highlight, got %q", target.GetAttribute("class"))
	}
}

func TestGetAttribute(t *testing.T) {
	doc := html.Parse(`<html><body><a id="link" href="https://example.com">Click</a></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var link = document.getElementById("link");
		var href = link.getAttribute("href");
		if (href !== "https://example.com") throw new Error("expected href, got " + href);
		var missing = link.getAttribute("nonexistent");
		if (missing !== null) throw new Error("expected null for missing attr, got " + missing);
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTextContent(t *testing.T) {
	doc := html.Parse(`<html><body><div id="test"><span>Hello</span> <span>World</span></div></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var el = document.getElementById("test");
		var text = el.textContent;
		if (text !== "Hello World") throw new Error("expected 'Hello World', got '" + text + "'");
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetTextContent(t *testing.T) {
	doc := html.Parse(`<html><body><div id="test"><span>Old</span></div></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var el = document.getElementById("test");
		el.textContent = "New content";
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	el := doc.GetElementByID("test")
	if el == nil {
		t.Fatal("element not found")
	}
	if el.TextContent() != "New content" {
		t.Errorf("expected 'New content', got %q", el.TextContent())
	}
	// Should have replaced children with single text node
	if len(el.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(el.Children))
	}
	if el.Children[0].Type != dom.TextNode {
		t.Error("expected text node child")
	}
}

func TestInnerHTML(t *testing.T) {
	doc := html.Parse(`<html><body><div id="test"><span>Hello</span></div></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var el = document.getElementById("test");
		el.innerHTML = "<p>New paragraph</p>";
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	el := doc.GetElementByID("test")
	if el == nil {
		t.Fatal("element not found")
	}
	if len(el.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(el.Children))
	}
	p := el.Children[0]
	if p.Data != "p" {
		t.Errorf("expected <p>, got <%s>", p.Data)
	}
	if p.TextContent() != "New paragraph" {
		t.Errorf("expected 'New paragraph', got %q", p.TextContent())
	}
}

func TestStyleProperty(t *testing.T) {
	doc := html.Parse(`<html><body><div id="box"></div></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var box = document.getElementById("box");
		box.style.backgroundColor = "red";
		box.style.fontSize = "16px";
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	box := doc.GetElementByID("box")
	if box == nil {
		t.Fatal("box not found")
	}
	style := box.GetAttribute("style")
	if !strings.Contains(style, "background-color: red") {
		t.Errorf("expected background-color in style, got %q", style)
	}
	if !strings.Contains(style, "font-size: 16px") {
		t.Errorf("expected font-size in style, got %q", style)
	}
}

func TestGetElementsByTagName(t *testing.T) {
	doc := html.Parse(`<html><body><p>One</p><p>Two</p><p>Three</p></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var ps = document.getElementsByTagName("p");
		if (ps.length !== 3) throw new Error("expected 3, got " + ps.length);
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetElementsByClassName(t *testing.T) {
	doc := html.Parse(`<html><body><div class="item">A</div><div class="item">B</div><div>C</div></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var items = document.getElementsByClassName("item");
		if (items.length !== 2) throw new Error("expected 2, got " + items.length);
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParentNode(t *testing.T) {
	doc := html.Parse(`<html><body><div id="parent"><p id="child">Text</p></div></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var child = document.getElementById("child");
		var parent = child.parentNode;
		if (parent.tagName !== "DIV") throw new Error("expected DIV, got " + parent.tagName);
		if (parent.id !== "parent") throw new Error("expected parent id");
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChildren(t *testing.T) {
	doc := html.Parse(`<html><body><div id="parent"><p>One</p><p>Two</p></div></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var parent = document.getElementById("parent");
		var children = parent.children;
		if (children.length !== 2) throw new Error("expected 2 children, got " + children.length);
		if (children[0].tagName !== "P") throw new Error("expected P");
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFirstLastChild(t *testing.T) {
	doc := html.Parse(`<html><body><ul id="list"><li>First</li><li>Last</li></ul></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var list = document.getElementById("list");
		if (list.firstChild.tagName !== "LI") throw new Error("expected LI for firstChild");
		if (list.lastChild.tagName !== "LI") throw new Error("expected LI for lastChild");
		if (list.firstChild.textContent !== "First") throw new Error("first child text wrong");
		if (list.lastChild.textContent !== "Last") throw new Error("last child text wrong");
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDocumentBody(t *testing.T) {
	doc := html.Parse(`<html><body><p>Hello</p></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var body = document.body;
		if (!body) throw new Error("document.body is null");
		if (body.tagName !== "BODY") throw new Error("expected BODY, got " + body.tagName);
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInsertBefore(t *testing.T) {
	doc := html.Parse(`<html><body><div id="container"><p id="existing">Existing</p></div></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var container = document.getElementById("container");
		var existing = document.getElementById("existing");
		var newEl = document.createElement("h1");
		newEl.textContent = "Inserted";
		container.insertBefore(newEl, existing);
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	container := doc.GetElementByID("container")
	if container == nil {
		t.Fatal("container not found")
	}
	// First element child should be h1
	var firstElement *dom.Node
	for _, c := range container.Children {
		if c.Type == dom.ElementNode {
			firstElement = c
			break
		}
	}
	if firstElement == nil || firstElement.Data != "h1" {
		t.Error("expected h1 as first element child")
	}
}

func TestNodeIdentityPreservation(t *testing.T) {
	doc := html.Parse(`<html><body><div id="test"></div></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var a = document.getElementById("test");
		var b = document.getElementById("test");
		if (a !== b) throw new Error("node identity not preserved");
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetID(t *testing.T) {
	doc := html.Parse(`<html><body><div id="old"></div></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var el = document.getElementById("old");
		el.id = "new";
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.GetElementByID("new") == nil {
		t.Error("expected element with id 'new'")
	}
	if doc.GetElementByID("old") != nil {
		t.Error("element with id 'old' should not exist")
	}
}

func TestSetClassName(t *testing.T) {
	doc := html.Parse(`<html><body><div id="test"></div></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var el = document.getElementById("test");
		el.className = "active highlight";
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	el := doc.GetElementByID("test")
	if el.GetAttribute("class") != "active highlight" {
		t.Errorf("expected class 'active highlight', got %q", el.GetAttribute("class"))
	}
}

func TestConsoleLog(t *testing.T) {
	doc := html.Parse(`<html><body></body></html>`)
	engine := NewEngine(doc)

	// console.log should not cause errors
	err := engine.Execute(`
		console.log("hello", "world");
		console.warn("warning");
		console.error("error");
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHasAttribute(t *testing.T) {
	doc := html.Parse(`<html><body><div id="test" class="box"></div></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var el = document.getElementById("test");
		if (!el.hasAttribute("class")) throw new Error("expected hasAttribute('class') to be true");
		if (el.hasAttribute("style")) throw new Error("expected hasAttribute('style') to be false");
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveAttribute(t *testing.T) {
	doc := html.Parse(`<html><body><div id="test" class="box"></div></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var el = document.getElementById("test");
		el.removeAttribute("class");
		if (el.hasAttribute("class")) throw new Error("class should be removed");
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCloneNode(t *testing.T) {
	doc := html.Parse(`<html><body><div id="original"><p>Child</p></div></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var orig = document.getElementById("original");
		var shallow = orig.cloneNode(false);
		if (shallow.id !== "original") throw new Error("clone should have same id");
		if (shallow.hasChildNodes()) throw new Error("shallow clone should have no children");

		var deep = orig.cloneNode(true);
		if (!deep.hasChildNodes()) throw new Error("deep clone should have children");
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestComplexDOMManipulation(t *testing.T) {
	doc := html.Parse(`<html><body><ul id="list"></ul></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var list = document.getElementById("list");
		for (var i = 0; i < 5; i++) {
			var li = document.createElement("li");
			li.textContent = "Item " + (i + 1);
			li.className = "list-item";
			list.appendChild(li);
		}
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	list := doc.GetElementByID("list")
	if list == nil {
		t.Fatal("list not found")
	}
	if len(list.Children) != 5 {
		t.Fatalf("expected 5 children, got %d", len(list.Children))
	}
	for i, child := range list.Children {
		if child.Data != "li" {
			t.Errorf("child %d: expected <li>, got <%s>", i, child.Data)
		}
		expected := "Item " + string(rune('1'+i))
		if child.TextContent() != expected {
			t.Errorf("child %d: expected %q, got %q", i, expected, child.TextContent())
		}
	}
}

func TestCamelToCSSProperty(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"backgroundColor", "background-color"},
		{"fontSize", "font-size"},
		{"marginTop", "margin-top"},
		{"color", "color"},
		{"borderBottomWidth", "border-bottom-width"},
	}

	for _, tt := range tests {
		result := camelToCSSProperty(tt.input)
		if result != tt.expected {
			t.Errorf("camelToCSSProperty(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestParseInlineStyle(t *testing.T) {
	tests := []struct {
		input    string
		expected map[string]string
	}{
		{"", map[string]string{}},
		{"color: red", map[string]string{"color": "red"}},
		{"color: red; font-size: 16px", map[string]string{"color": "red", "font-size": "16px"}},
		{"  margin : 10px ;  ", map[string]string{"margin": "10px"}},
	}

	for _, tt := range tests {
		result := parseInlineStyle(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("parseInlineStyle(%q): expected %d entries, got %d", tt.input, len(tt.expected), len(result))
			continue
		}
		for k, v := range tt.expected {
			if result[k] != v {
				t.Errorf("parseInlineStyle(%q): key %q = %q, want %q", tt.input, k, result[k], v)
			}
		}
	}
}

func TestDocumentGetElementsByTagNameOnElement(t *testing.T) {
	doc := html.Parse(`<html><body><div id="container"><span>A</span><span>B</span></div><span>C</span></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var container = document.getElementById("container");
		var spans = container.getElementsByTagName("span");
		if (spans.length !== 2) throw new Error("expected 2, got " + spans.length);
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNextSibling(t *testing.T) {
	doc := html.Parse(`<html><body><p id="first">First</p><p id="second">Second</p></body></html>`)
	engine := NewEngine(doc)

	err := engine.Execute(`
		var first = document.getElementById("first");
		var next = first.nextSibling;
		if (!next) throw new Error("nextSibling is null");
		if (next.id !== "second") throw new Error("expected id=second, got " + next.id);
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
