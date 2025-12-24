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
