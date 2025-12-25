package html

import (
	"github.com/lukehoban/browser/dom"
	"testing"
)

func TestParseSimpleElement(t *testing.T) {
	input := "<div>Hello</div>"
	doc := Parse(input)

	if len(doc.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(doc.Children))
	}

	div := doc.Children[0]
	if div.Type != dom.ElementNode {
		t.Errorf("Expected ElementNode, got %v", div.Type)
	}
	if div.Data != "div" {
		t.Errorf("Expected tag 'div', got %v", div.Data)
	}
	if len(div.Children) != 1 {
		t.Fatalf("Expected 1 child in div, got %d", len(div.Children))
	}

	text := div.Children[0]
	if text.Type != dom.TextNode {
		t.Errorf("Expected TextNode, got %v", text.Type)
	}
	if text.Data != "Hello" {
		t.Errorf("Expected text 'Hello', got %v", text.Data)
	}
}

func TestParseNestedElements(t *testing.T) {
	input := "<html><body><div><p>Hello</p></div></body></html>"
	doc := Parse(input)

	if len(doc.Children) != 1 {
		t.Fatalf("Expected 1 child (html), got %d", len(doc.Children))
	}

	html := doc.Children[0]
	if html.Data != "html" {
		t.Errorf("Expected 'html', got %v", html.Data)
	}

	if len(html.Children) != 1 {
		t.Fatalf("Expected 1 child (body), got %d", len(html.Children))
	}

	body := html.Children[0]
	if body.Data != "body" {
		t.Errorf("Expected 'body', got %v", body.Data)
	}

	if len(body.Children) != 1 {
		t.Fatalf("Expected 1 child (div), got %d", len(body.Children))
	}

	div := body.Children[0]
	if div.Data != "div" {
		t.Errorf("Expected 'div', got %v", div.Data)
	}

	if len(div.Children) != 1 {
		t.Fatalf("Expected 1 child (p), got %d", len(div.Children))
	}

	p := div.Children[0]
	if p.Data != "p" {
		t.Errorf("Expected 'p', got %v", p.Data)
	}
}

func TestParseAttributes(t *testing.T) {
	input := `<div id="main" class="container active">`
	doc := Parse(input)

	if len(doc.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(doc.Children))
	}

	div := doc.Children[0]
	if div.GetAttribute("id") != "main" {
		t.Errorf("Expected id 'main', got %v", div.GetAttribute("id"))
	}
	if div.GetAttribute("class") != "container active" {
		t.Errorf("Expected class 'container active', got %v", div.GetAttribute("class"))
	}
}

func TestParseSelfClosingTag(t *testing.T) {
	input := "<div><br /></div>"
	doc := Parse(input)

	if len(doc.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(doc.Children))
	}

	div := doc.Children[0]
	if len(div.Children) != 1 {
		t.Fatalf("Expected 1 child (br), got %d", len(div.Children))
	}

	br := div.Children[0]
	if br.Data != "br" {
		t.Errorf("Expected 'br', got %v", br.Data)
	}
	if len(br.Children) != 0 {
		t.Errorf("Expected br to have no children, got %d", len(br.Children))
	}
}

func TestParseVoidElement(t *testing.T) {
	input := "<div><img src='test.jpg'><p>Text</p></div>"
	doc := Parse(input)

	if len(doc.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(doc.Children))
	}

	div := doc.Children[0]
	if len(div.Children) != 2 {
		t.Fatalf("Expected 2 children (img, p), got %d", len(div.Children))
	}

	img := div.Children[0]
	if img.Data != "img" {
		t.Errorf("Expected 'img', got %v", img.Data)
	}
	if img.GetAttribute("src") != "test.jpg" {
		t.Errorf("Expected src 'test.jpg', got %v", img.GetAttribute("src"))
	}

	p := div.Children[1]
	if p.Data != "p" {
		t.Errorf("Expected 'p', got %v", p.Data)
	}
}

func TestParseMixedContent(t *testing.T) {
	input := "<p>Hello <strong>World</strong>!</p>"
	doc := Parse(input)

	if len(doc.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(doc.Children))
	}

	p := doc.Children[0]
	if len(p.Children) != 3 {
		t.Fatalf("Expected 3 children, got %d", len(p.Children))
	}

	// First text node
	if p.Children[0].Type != dom.TextNode || p.Children[0].Data != "Hello " {
		t.Errorf("Expected 'Hello ', got %v", p.Children[0].Data)
	}

	// Strong element
	strong := p.Children[1]
	if strong.Data != "strong" {
		t.Errorf("Expected 'strong', got %v", strong.Data)
	}
	if len(strong.Children) != 1 {
		t.Fatalf("Expected 1 child in strong, got %d", len(strong.Children))
	}
	if strong.Children[0].Data != "World" {
		t.Errorf("Expected 'World', got %v", strong.Children[0].Data)
	}

	// Last text node
	if p.Children[2].Type != dom.TextNode || p.Children[2].Data != "!" {
		t.Errorf("Expected '!', got %v", p.Children[2].Data)
	}
}

// Tests for character reference decoding (HTML5 §12.2.4.2)

func TestParseCharacterReferences(t *testing.T) {
	// HTML5 §12.2.4.2 Character reference state
	// Character references like &amp;, &lt;, &gt;, &nbsp; should be decoded
	
	input := "<div>&lt;p&gt; &amp; &quot;</div>"
	doc := Parse(input)
	
	div := doc.Children[0]
	if len(div.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(div.Children))
	}
	
	text := div.Children[0]
	expected := "<p> & \""
	if text.Data != expected {
		t.Errorf("Expected text '%s', got '%s'", expected, text.Data)
	}
}

func TestParseNumericCharacterReferences(t *testing.T) {
	// HTML5 §12.2.4.3 Numeric character reference state
	// Both decimal (&#NNN;) and hexadecimal (&#xHHH;) forms should be supported
	
	input := "<div>&#60;&#x3E;&#169;</div>"
	doc := Parse(input)
	
	div := doc.Children[0]
	text := div.Children[0]
	expected := "<>©" // <, >, copyright symbol
	if text.Data != expected {
		t.Errorf("Expected text '%s', got '%s'", expected, text.Data)
	}
}

func TestParseNbsp(t *testing.T) {
	// HTML5 §12.2.4.4: Named character reference - non-breaking space
	// &nbsp; should be decoded to Unicode non-breaking space (U+00A0)
	
	input := "<div>Hello&nbsp;World</div>"
	doc := Parse(input)
	
	div := doc.Children[0]
	text := div.Children[0]
	expected := "Hello\u00A0World"
	if text.Data != expected {
		t.Errorf("Expected text with non-breaking space, got '%s'", text.Data)
	}
}

func TestParseScriptCDATA_Skipped(t *testing.T) {
	t.Skip("Script CDATA sections not implemented - HTML5 §12.2.5.14")
	// HTML5 §12.2.5.14 Script data state
	// Script tags should handle <![CDATA[ sections specially
	
	input := "<script><![CDATA[var x = 1 < 2;]]></script>"
	doc := Parse(input)
	
	if len(doc.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(doc.Children))
	}
	
	script := doc.Children[0]
	if script.Data != "script" {
		t.Errorf("Expected 'script', got %v", script.Data)
	}
	
	if len(script.Children) != 1 {
		t.Fatalf("Expected 1 text child, got %d", len(script.Children))
	}
	
	text := script.Children[0]
	expected := "var x = 1 < 2;"
	if text.Data != expected {
		t.Errorf("Expected text '%s', got '%s'", expected, text.Data)
	}
}

func TestParseStyleCDATA_Skipped(t *testing.T) {
	t.Skip("Style CDATA sections not implemented - HTML5 §12.2.5.16")
	// HTML5 §12.2.5.16 Style data state
	// Style tags should handle content without HTML parsing
	
	input := "<style>div > p { color: red; }</style>"
	doc := Parse(input)
	
	if len(doc.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(doc.Children))
	}
	
	style := doc.Children[0]
	if style.Data != "style" {
		t.Errorf("Expected 'style', got %v", style.Data)
	}
	
	if len(style.Children) != 1 {
		t.Fatalf("Expected 1 text child, got %d", len(style.Children))
	}
	
	text := style.Children[0]
	expected := "div > p { color: red; }"
	if text.Data != expected {
		t.Errorf("Expected text '%s', got '%s'", expected, text.Data)
	}
}

func TestParseSVGNamespace_Skipped(t *testing.T) {
	t.Skip("Namespace support not implemented - HTML5 §12.2.6.5")
	// HTML5 §12.2.6.5 Foreign elements
	// SVG and MathML elements should be parsed with proper namespace handling
	
	input := "<svg><circle cx='50' cy='50' r='40'/></svg>"
	doc := Parse(input)
	
	if len(doc.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(doc.Children))
	}
	
	svg := doc.Children[0]
	if svg.Data != "svg" {
		t.Errorf("Expected 'svg', got %v", svg.Data)
	}
	
	// Should have SVG namespace (when namespace support is added)
	// Expected: svg.Namespace == "http://www.w3.org/2000/svg"
	
	if len(svg.Children) != 1 {
		t.Fatalf("Expected 1 child (circle), got %d", len(svg.Children))
	}
	
	circle := svg.Children[0]
	if circle.Data != "circle" {
		t.Errorf("Expected 'circle', got %v", circle.Data)
	}
}

func TestParseMathMLNamespace_Skipped(t *testing.T) {
	t.Skip("Namespace support not implemented - HTML5 §12.2.6.5")
	// HTML5 §12.2.6.5 Foreign elements
	// MathML elements should be parsed with proper namespace
	
	input := "<math><mrow><mi>x</mi></mrow></math>"
	doc := Parse(input)
	
	if len(doc.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(doc.Children))
	}
	
	math := doc.Children[0]
	if math.Data != "math" {
		t.Errorf("Expected 'math', got %v", math.Data)
	}
	
	// Should have MathML namespace (when namespace support is added)
	// Expected: math.Namespace == "http://www.w3.org/1998/Math/MathML"
}
