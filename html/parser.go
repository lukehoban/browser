package html

import (
	"github.com/lukehoban/browser/dom"
)

// Parser parses HTML and builds a DOM tree.
// This is a simplified parser based on HTML5 tree construction.
//
// Spec references:
// - HTML5 §12.2.6 Tree construction: https://html.spec.whatwg.org/multipage/parsing.html#tree-construction
type Parser struct {
	tokenizer *Tokenizer
	doc       *dom.Node
	stack     []*dom.Node // Stack of open elements
}

// NewParser creates a new HTML parser.
func NewParser(input string) *Parser {
	return &Parser{
		tokenizer: NewTokenizer(input),
		doc:       dom.NewDocument(),
		stack:     make([]*dom.Node, 0),
	}
}

// Parse parses the HTML input and returns a DOM tree.
func (p *Parser) Parse() *dom.Node {
	// Start with document on stack
	p.stack = append(p.stack, p.doc)

	for {
		token, ok := p.tokenizer.Next()
		if !ok {
			break
		}

		p.processToken(token)
	}

	return p.doc
}

// processToken processes a single token and updates the DOM tree.
// This is a simplified version of HTML5 tree construction algorithm.
func (p *Parser) processToken(token Token) {
	switch token.Type {
	case StartTagToken, SelfClosingTagToken:
		p.handleStartTag(token)
	case EndTagToken:
		p.handleEndTag(token)
	case TextToken:
		p.handleText(token)
	case CommentToken:
		// Ignore comments for now
	case DoctypeToken:
		// Ignore DOCTYPE for now
	}
}

// handleStartTag handles a start tag token.
// HTML5 §12.2.6.4.7 "in body" insertion mode (simplified)
func (p *Parser) handleStartTag(token Token) {
	elem := dom.NewElement(token.Data)

	// Set attributes
	for name, value := range token.Attributes {
		elem.SetAttribute(name, value)
	}

	// Append to current node
	current := p.currentNode()
	current.AppendChild(elem)

	// Push to stack unless self-closing or void element
	if token.Type != SelfClosingTagToken && !isVoidElement(token.Data) {
		p.stack = append(p.stack, elem)
	}
}

// handleEndTag handles an end tag token.
func (p *Parser) handleEndTag(token Token) {
	// Find matching element in stack and pop
	for i := len(p.stack) - 1; i >= 0; i-- {
		node := p.stack[i]
		if node.Type == dom.ElementNode && node.Data == token.Data {
			// Pop all elements up to and including this one
			p.stack = p.stack[:i]
			return
		}
	}
	// If not found, ignore (error recovery)
}

// handleText handles a text token.
func (p *Parser) handleText(token Token) {
	// Skip whitespace-only text nodes at document level
	if len(p.stack) == 1 {
		allWhitespace := true
		for _, c := range token.Data {
			if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
				allWhitespace = false
				break
			}
		}
		if allWhitespace {
			return
		}
	}

	text := dom.NewText(token.Data)
	current := p.currentNode()
	current.AppendChild(text)
}

// currentNode returns the current node (top of stack).
func (p *Parser) currentNode() *dom.Node {
	if len(p.stack) == 0 {
		return p.doc
	}
	return p.stack[len(p.stack)-1]
}

// isVoidElement returns true if the element is a void element.
// Void elements cannot have children.
// HTML5 §12.1.2 Elements: https://html.spec.whatwg.org/multipage/syntax.html#void-elements
func isVoidElement(tagName string) bool {
	voidElements := map[string]bool{
		"area":   true,
		"base":   true,
		"br":     true,
		"col":    true,
		"embed":  true,
		"hr":     true,
		"img":    true,
		"input":  true,
		"link":   true,
		"meta":   true,
		"param":  true,
		"source": true,
		"track":  true,
		"wbr":    true,
	}
	return voidElements[tagName]
}

// Parse is a convenience function to parse HTML.
func Parse(input string) *dom.Node {
	parser := NewParser(input)
	return parser.Parse()
}
