// Package css provides CSS parsing and stylesheet management.
// It follows the CSS 2.1 specification.
//
// Spec references:
// - CSS 2.1 §4 Syntax and basic data types: https://www.w3.org/TR/CSS21/syndata.html
// - CSS 2.1 §5 Selectors: https://www.w3.org/TR/CSS21/selector.html
// - CSS 2.1 §4.1.7 Rule sets, declaration blocks, and selectors: https://www.w3.org/TR/CSS21/syndata.html#rule-sets
//
// Implemented features:
// - CSS tokenization (identifiers, strings, numbers, hash, operators)
// - Rule parsing (selectors and declarations)
// - Simple selectors: element, class (.class), ID (#id)
// - Descendant combinators (space-separated selectors)
// - Multiple selectors (comma-separated)
// - Graceful handling of @-rules (skipped, not parsed)
// - Graceful handling of attribute selectors (skipped)
// - Partial pseudo-class support (stripped from selector)
//
// Not yet implemented (logged as warnings when encountered):
// - Child combinator > (CSS 2.1 §5.6)
// - Sibling combinators +, ~ (CSS 2.1 §5.7, CSS3)
// - Attribute selector matching (CSS 2.1 §5.8)
// - Pseudo-classes :hover, :focus (CSS 2.1 §5.11)
// - Pseudo-elements ::before, ::after (CSS 2.1 §5.12)
// - @media queries, @import, @font-face (CSS 2.1 §4.1.5)
// - !important declarations (CSS 2.1 §6.4.2)
// - Full shorthand property parsing
package css

import (
	"unicode"
)

// TokenType represents the type of a CSS token.
type TokenType int

const (
	// ErrorToken indicates an error
	ErrorToken TokenType = iota
	// IdentToken represents an identifier
	IdentToken
	// StringToken represents a string literal
	StringToken
	// NumberToken represents a number
	NumberToken
	// HashToken represents a hash (#id or #color)
	HashToken
	// ColonToken represents ':'
	ColonToken
	// SemicolonToken represents ';'
	SemicolonToken
	// CommaToken represents ','
	CommaToken
	// LeftBraceToken represents '{'
	LeftBraceToken
	// RightBraceToken represents '}'
	RightBraceToken
	// LeftParenToken represents '('
	LeftParenToken
	// RightParenToken represents ')'
	RightParenToken
	// LeftBracketToken represents '['
	LeftBracketToken
	// RightBracketToken represents ']'
	RightBracketToken
	// WhitespaceToken represents whitespace
	WhitespaceToken
	// DotToken represents '.'
	DotToken
	// AtKeywordToken represents an @-rule (e.g., @media, @import)
	AtKeywordToken
	// EOFToken represents end of input
	EOFToken
)

// Token represents a CSS token.
type Token struct {
	Type  TokenType
	Value string
}

// Tokenizer tokenizes CSS input.
// CSS 2.1 §4.1.1 Tokenization
type Tokenizer struct {
	input string
	pos   int
}

// NewTokenizer creates a new CSS tokenizer.
func NewTokenizer(input string) *Tokenizer {
	return &Tokenizer{
		input: input,
		pos:   0,
	}
}

// Next returns the next token.
func (t *Tokenizer) Next() Token {
	if t.pos >= len(t.input) {
		return Token{Type: EOFToken}
	}

	c := t.input[t.pos]

	// Whitespace
	if unicode.IsSpace(rune(c)) {
		return t.readWhitespace()
	}

	// String literals
	if c == '"' || c == '\'' {
		return t.readString(c)
	}

	// Numbers
	if unicode.IsDigit(rune(c)) || (c == '.' && t.pos+1 < len(t.input) && unicode.IsDigit(rune(t.input[t.pos+1]))) {
		return t.readNumber()
	}

	// Hash
	if c == '#' {
		t.pos++
		name := t.readName()
		return Token{Type: HashToken, Value: name}
	}

	// Dot (class selector)
	if c == '.' && t.pos+1 < len(t.input) && isNameStart(rune(t.input[t.pos+1])) {
		t.pos++
		return Token{Type: DotToken, Value: "."}
	}

	// At-keyword (@media, @import, etc.) - CSS 2.1 §4.1.5 At-rules
	if c == '@' {
		t.pos++
		name := t.readName()
		return Token{Type: AtKeywordToken, Value: name}
	}

	// Single-character tokens
	switch c {
	case ':':
		t.pos++
		return Token{Type: ColonToken, Value: ":"}
	case ';':
		t.pos++
		return Token{Type: SemicolonToken, Value: ";"}
	case ',':
		t.pos++
		return Token{Type: CommaToken, Value: ","}
	case '{':
		t.pos++
		return Token{Type: LeftBraceToken, Value: "{"}
	case '}':
		t.pos++
		return Token{Type: RightBraceToken, Value: "}"}
	case '(':
		t.pos++
		return Token{Type: LeftParenToken, Value: "("}
	case ')':
		t.pos++
		return Token{Type: RightParenToken, Value: ")"}
	case '[':
		t.pos++
		return Token{Type: LeftBracketToken, Value: "["}
	case ']':
		t.pos++
		return Token{Type: RightBracketToken, Value: "]"}
	}

	// Comments
	if c == '/' && t.pos+1 < len(t.input) && t.input[t.pos+1] == '*' {
		return t.readComment()
	}

	// Identifiers and keywords
	if isNameStart(rune(c)) {
		return t.readIdent()
	}

	// Unknown character
	t.pos++
	return Token{Type: ErrorToken, Value: string(c)}
}

// readWhitespace reads whitespace characters.
func (t *Tokenizer) readWhitespace() Token {
	start := t.pos
	for t.pos < len(t.input) && unicode.IsSpace(rune(t.input[t.pos])) {
		t.pos++
	}
	return Token{Type: WhitespaceToken, Value: t.input[start:t.pos]}
}

// readString reads a string literal.
// CSS 2.1 §4.1.3 Characters and case
func (t *Tokenizer) readString(quote byte) Token {
	t.pos++ // consume opening quote
	start := t.pos

	for t.pos < len(t.input) {
		c := t.input[t.pos]
		if c == quote {
			value := t.input[start:t.pos]
			t.pos++ // consume closing quote
			return Token{Type: StringToken, Value: value}
		}
		if c == '\\' && t.pos+1 < len(t.input) {
			t.pos += 2 // skip escape sequence
			continue
		}
		t.pos++
	}

	// Unclosed string
	return Token{Type: StringToken, Value: t.input[start:]}
}

// readNumber reads a number.
// CSS 2.1 §4.3.1 Integers and real numbers
func (t *Tokenizer) readNumber() Token {
	start := t.pos

	// Read integer part
	for t.pos < len(t.input) && unicode.IsDigit(rune(t.input[t.pos])) {
		t.pos++
	}

	// Read decimal part
	if t.pos < len(t.input) && t.input[t.pos] == '.' {
		t.pos++
		for t.pos < len(t.input) && unicode.IsDigit(rune(t.input[t.pos])) {
			t.pos++
		}
	}

	value := t.input[start:t.pos]

	// Read unit (if present)
	if t.pos < len(t.input) && isNameStart(rune(t.input[t.pos])) {
		for t.pos < len(t.input) && isNameChar(rune(t.input[t.pos])) {
			t.pos++
		}
		value = t.input[start:t.pos] // include unit
	}

	return Token{Type: NumberToken, Value: value}
}

// readIdent reads an identifier.
// CSS 2.1 §4.1.3 Characters and case
func (t *Tokenizer) readIdent() Token {
	start := t.pos
	t.pos++ // consume first character

	for t.pos < len(t.input) && isNameChar(rune(t.input[t.pos])) {
		t.pos++
	}

	return Token{Type: IdentToken, Value: t.input[start:t.pos]}
}

// readName reads a name (after # for hash).
func (t *Tokenizer) readName() string {
	start := t.pos
	for t.pos < len(t.input) && isNameChar(rune(t.input[t.pos])) {
		t.pos++
	}
	return t.input[start:t.pos]
}

// readComment reads and skips a comment.
// CSS 2.1 §4.1.9 Comments
func (t *Tokenizer) readComment() Token {
	t.pos += 2 // consume '/*'

	for t.pos < len(t.input)-1 {
		if t.input[t.pos] == '*' && t.input[t.pos+1] == '/' {
			t.pos += 2 // consume '*/'
			// Skip to next meaningful token
			for t.pos < len(t.input) && unicode.IsSpace(rune(t.input[t.pos])) {
				t.pos++
			}
			return t.Next() // return next token after comment
		}
		t.pos++
	}

	// Unclosed comment, skip to end
	t.pos = len(t.input)
	return Token{Type: EOFToken}
}

// isNameStart returns true if the character can start a name.
// CSS 2.1 §4.1.3
func isNameStart(c rune) bool {
	return unicode.IsLetter(c) || c == '_' || c == '-' || c > 127
}

// isNameChar returns true if the character can be part of a name.
// CSS 2.1 §4.1.3
func isNameChar(c rune) bool {
	return isNameStart(c) || unicode.IsDigit(c)
}

// Peek returns the next token without consuming it.
func (t *Tokenizer) Peek() Token {
	savedPos := t.pos
	token := t.Next()
	t.pos = savedPos
	return token
}

// SkipWhitespace skips whitespace tokens.
func (t *Tokenizer) SkipWhitespace() {
	for {
		token := t.Peek()
		if token.Type != WhitespaceToken {
			break
		}
		t.Next()
	}
}
