// Package html provides HTML tokenization and parsing.
// It follows the HTML5 tokenization algorithm.
//
// Spec references:
// - HTML5 §12.2.5 Tokenization: https://html.spec.whatwg.org/multipage/parsing.html#tokenization
package html

import (
	"strings"
	"unicode"
)

// TokenType represents the type of an HTML token.
type TokenType int

const (
	// ErrorToken indicates an error occurred during tokenization
	ErrorToken TokenType = iota
	// TextToken represents text content
	TextToken
	// StartTagToken represents an opening tag (e.g., <div>)
	StartTagToken
	// EndTagToken represents a closing tag (e.g., </div>)
	EndTagToken
	// SelfClosingTagToken represents a self-closing tag (e.g., <br />)
	SelfClosingTagToken
	// CommentToken represents an HTML comment
	CommentToken
	// DoctypeToken represents a DOCTYPE declaration
	DoctypeToken
)

// Token represents an HTML token.
type Token struct {
	Type       TokenType
	Data       string            // Tag name or text content
	Attributes map[string]string // Attributes for tags
}

// Tokenizer tokenizes HTML input.
// This is a simplified implementation based on HTML5 §12.2.5.
type Tokenizer struct {
	input string
	pos   int
}

// NewTokenizer creates a new HTML tokenizer.
func NewTokenizer(input string) *Tokenizer {
	return &Tokenizer{
		input: input,
		pos:   0,
	}
}

// Next returns the next token from the input.
func (t *Tokenizer) Next() (Token, bool) {
	if t.pos >= len(t.input) {
		return Token{}, false
	}

	// HTML5 §12.2.5.1 Data state
	if t.input[t.pos] != '<' {
		return t.readText(), true
	}

	// Start of tag
	t.pos++ // consume '<'
	
	if t.pos >= len(t.input) {
		return Token{Type: TextToken, Data: "<"}, true
	}

	// HTML5 §12.2.5.6 Tag open state
	switch t.input[t.pos] {
	case '!':
		// Comment or DOCTYPE
		t.pos++
		if strings.HasPrefix(t.input[t.pos:], "--") {
			return t.readComment(), true
		}
		if strings.HasPrefix(strings.ToUpper(t.input[t.pos:]), "DOCTYPE") {
			return t.readDoctype(), true
		}
		// Invalid, treat as text
		return Token{Type: TextToken, Data: "<!"}, true
	
	case '/':
		// End tag
		t.pos++
		return t.readEndTag(), true
	
	default:
		// Start tag
		return t.readStartTag(), true
	}
}

// readText reads text content until the next '<'.
// HTML5 §12.2.5.1 Data state
func (t *Tokenizer) readText() Token {
	start := t.pos
	for t.pos < len(t.input) && t.input[t.pos] != '<' {
		t.pos++
	}
	return Token{
		Type: TextToken,
		Data: t.input[start:t.pos],
	}
}

// readStartTag reads a start tag.
// HTML5 §12.2.5.8 Tag name state
func (t *Tokenizer) readStartTag() Token {
	tagName := t.readTagName()
	attrs := t.readAttributes()
	
	selfClosing := false
	if t.pos < len(t.input) && t.input[t.pos] == '/' {
		selfClosing = true
		t.pos++
	}
	
	// Consume '>'
	if t.pos < len(t.input) && t.input[t.pos] == '>' {
		t.pos++
	}
	
	tokenType := StartTagToken
	if selfClosing {
		tokenType = SelfClosingTagToken
	}
	
	return Token{
		Type:       tokenType,
		Data:       strings.ToLower(tagName),
		Attributes: attrs,
	}
}

// readEndTag reads an end tag.
// HTML5 §12.2.5.9 End tag open state
func (t *Tokenizer) readEndTag() Token {
	tagName := t.readTagName()
	
	// Skip to '>'
	for t.pos < len(t.input) && t.input[t.pos] != '>' {
		t.pos++
	}
	
	// Consume '>'
	if t.pos < len(t.input) {
		t.pos++
	}
	
	return Token{
		Type: EndTagToken,
		Data: strings.ToLower(tagName),
	}
}

// readTagName reads a tag name.
func (t *Tokenizer) readTagName() string {
	start := t.pos
	for t.pos < len(t.input) {
		c := t.input[t.pos]
		if c == '>' || c == '/' || unicode.IsSpace(rune(c)) {
			break
		}
		t.pos++
	}
	return t.input[start:t.pos]
}

// readAttributes reads tag attributes.
// HTML5 §12.2.5.32 Before attribute name state
func (t *Tokenizer) readAttributes() map[string]string {
	attrs := make(map[string]string)
	
	for t.pos < len(t.input) {
		t.skipWhitespace()
		
		if t.pos >= len(t.input) {
			break
		}
		
		c := t.input[t.pos]
		if c == '>' || c == '/' {
			break
		}
		
		// Read attribute name
		name := t.readAttrName()
		if name == "" {
			break
		}
		
		t.skipWhitespace()
		
		// Check for '='
		value := ""
		if t.pos < len(t.input) && t.input[t.pos] == '=' {
			t.pos++ // consume '='
			t.skipWhitespace()
			value = t.readAttrValue()
		}
		
		attrs[strings.ToLower(name)] = value
	}
	
	return attrs
}

// readAttrName reads an attribute name.
func (t *Tokenizer) readAttrName() string {
	start := t.pos
	for t.pos < len(t.input) {
		c := t.input[t.pos]
		if c == '=' || c == '>' || c == '/' || unicode.IsSpace(rune(c)) {
			break
		}
		t.pos++
	}
	return t.input[start:t.pos]
}

// readAttrValue reads an attribute value.
// HTML5 §12.2.5.37 Attribute value states
func (t *Tokenizer) readAttrValue() string {
	if t.pos >= len(t.input) {
		return ""
	}
	
	quote := t.input[t.pos]
	if quote == '"' || quote == '\'' {
		// Quoted value
		t.pos++ // consume quote
		start := t.pos
		for t.pos < len(t.input) && t.input[t.pos] != quote {
			t.pos++
		}
		value := t.input[start:t.pos]
		if t.pos < len(t.input) {
			t.pos++ // consume closing quote
		}
		return value
	}
	
	// Unquoted value
	start := t.pos
	for t.pos < len(t.input) {
		c := t.input[t.pos]
		if unicode.IsSpace(rune(c)) || c == '>' {
			break
		}
		t.pos++
	}
	return t.input[start:t.pos]
}

// readComment reads an HTML comment.
// HTML5 §12.2.5.42 Comment start state
func (t *Tokenizer) readComment() Token {
	t.pos += 2 // consume '--'
	start := t.pos
	
	// Find end of comment
	for t.pos < len(t.input)-2 {
		if t.input[t.pos] == '-' && t.input[t.pos+1] == '-' && t.input[t.pos+2] == '>' {
			data := t.input[start:t.pos]
			t.pos += 3 // consume '-->'
			return Token{Type: CommentToken, Data: data}
		}
		t.pos++
	}
	
	// Unclosed comment
	return Token{Type: CommentToken, Data: t.input[start:]}
}

// readDoctype reads a DOCTYPE declaration.
func (t *Tokenizer) readDoctype() Token {
	start := t.pos
	
	// Skip to '>'
	for t.pos < len(t.input) && t.input[t.pos] != '>' {
		t.pos++
	}
	
	data := t.input[start:t.pos]
	
	// Consume '>'
	if t.pos < len(t.input) {
		t.pos++
	}
	
	return Token{Type: DoctypeToken, Data: data}
}

// skipWhitespace skips whitespace characters.
func (t *Tokenizer) skipWhitespace() {
	for t.pos < len(t.input) && unicode.IsSpace(rune(t.input[t.pos])) {
		t.pos++
	}
}
