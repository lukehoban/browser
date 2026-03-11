package js

import (
	"fmt"
	"strings"
	"unicode"
)

// TokenType is the type of a JavaScript token.
type TokenType int

const (
	// Literals
	tokNumber TokenType = iota
	tokString
	tokTemplate
	tokIdent
	// Punctuation
	tokLParen    // (
	tokRParen    // )
	tokLBrace    // {
	tokRBrace    // }
	tokLBracket  // [
	tokRBracket  // ]
	tokSemicolon // ;
	tokColon     // :
	tokComma     // ,
	tokDot       // .
	tokDotDotDot // ...
	tokFatArrow  // =>
	tokQuestion  // ?
	tokOptChain  // ?.
	tokNullCoal  // ??
	// Arithmetic operators
	tokPlus     // +
	tokMinus    // -
	tokStar     // *
	tokSlash    // /
	tokPercent  // %
	tokStarStar // **
	// Bitwise
	tokAmp    // &
	tokPipe   // |
	tokCaret  // ^
	tokTilde  // ~
	tokLShift // <<
	tokRShift // >>
	// Assignment operators
	tokAssign       // =
	tokPlusAssign   // +=
	tokMinusAssign  // -=
	tokStarAssign   // *=
	tokSlashAssign  // /=
	tokPercentAssign // %=
	tokAmpAssign    // &=
	tokPipeAssign   // |=
	tokCaretAssign  // ^=
	// Comparison
	tokEqEq    // ==
	tokEqEqEq  // ===
	tokNeq     // !=
	tokNeqEq   // !==
	tokLt      // <
	tokGt      // >
	tokLte     // <=
	tokGte     // >=
	// Logical
	tokAnd // &&
	tokOr  // ||
	tokNot // !
	// Update
	tokPlusPlus   // ++
	tokMinusMinus // --
	// Keywords (stored as tokIdent and checked by value)
	tokEOF
)

var keywords = map[string]bool{
	"var": true, "let": true, "const": true,
	"function": true, "return": true,
	"if": true, "else": true,
	"while": true, "do": true,
	"for": true, "break": true, "continue": true,
	"new": true, "this": true,
	"typeof": true, "instanceof": true, "in": true, "of": true,
	"void": true, "delete": true,
	"true": true, "false": true, "null": true, "undefined": true,
	"throw": true, "try": true, "catch": true, "finally": true,
	"switch": true, "case": true, "default": true,
	"class": true, "extends": true, "super": true,
	"import": true, "export": true,
	"async": true, "await": true,
}

// Token is a single lexical token.
type Token struct {
	Type    TokenType
	Value   string // raw text
	Line    int
	AfterNL bool // true if a newline occurred before this token
}

// Lexer tokenizes JavaScript source code.
type Lexer struct {
	src    []rune
	pos    int
	line   int
	nlSeen bool // newline seen since last token
	peeked *Token
}

// NewLexer creates a new JavaScript lexer.
func NewLexer(src string) *Lexer {
	return &Lexer{src: []rune(src), pos: 0, line: 1}
}

// Peek returns the next token without consuming it.
func (l *Lexer) Peek() Token {
	if l.peeked == nil {
		t := l.next()
		l.peeked = &t
	}
	return *l.peeked
}

// Next consumes and returns the next token.
func (l *Lexer) Next() Token {
	if l.peeked != nil {
		t := *l.peeked
		l.peeked = nil
		return t
	}
	return l.next()
}

func (l *Lexer) next() Token {
	afterNL := l.nlSeen
	l.nlSeen = false

	for l.pos < len(l.src) {
		ch := l.src[l.pos]

		// Whitespace
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.pos++
			continue
		}
		if ch == '\n' {
			l.pos++
			l.line++
			afterNL = true
			continue
		}

		// Line comment
		if ch == '/' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '/' {
			for l.pos < len(l.src) && l.src[l.pos] != '\n' {
				l.pos++
			}
			continue
		}

		// Block comment
		if ch == '/' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '*' {
			l.pos += 2
			for l.pos+1 < len(l.src) {
				if l.src[l.pos] == '\n' {
					l.line++
					afterNL = true
				}
				if l.src[l.pos] == '*' && l.src[l.pos+1] == '/' {
					l.pos += 2
					break
				}
				l.pos++
			}
			continue
		}

		// Number literal
		if unicode.IsDigit(ch) || (ch == '.' && l.pos+1 < len(l.src) && unicode.IsDigit(l.src[l.pos+1])) {
			return l.readNumber(afterNL)
		}

		// String literal
		if ch == '"' || ch == '\'' {
			return l.readString(ch, afterNL)
		}

		// Template literal
		if ch == '`' {
			return l.readTemplate(afterNL)
		}

		// Identifier or keyword
		if unicode.IsLetter(ch) || ch == '_' || ch == '$' {
			return l.readIdent(afterNL)
		}

		// Punctuation and operators
		return l.readOp(afterNL)
	}

	return Token{Type: tokEOF, Line: l.line, AfterNL: afterNL}
}

func (l *Lexer) readNumber(afterNL bool) Token {
	start := l.pos
	// Hex
	if l.src[l.pos] == '0' && l.pos+1 < len(l.src) && (l.src[l.pos+1] == 'x' || l.src[l.pos+1] == 'X') {
		l.pos += 2
		for l.pos < len(l.src) && isHexDigit(l.src[l.pos]) {
			l.pos++
		}
	} else {
		for l.pos < len(l.src) && unicode.IsDigit(l.src[l.pos]) {
			l.pos++
		}
		if l.pos < len(l.src) && l.src[l.pos] == '.' {
			l.pos++
			for l.pos < len(l.src) && unicode.IsDigit(l.src[l.pos]) {
				l.pos++
			}
		}
		if l.pos < len(l.src) && (l.src[l.pos] == 'e' || l.src[l.pos] == 'E') {
			l.pos++
			if l.pos < len(l.src) && (l.src[l.pos] == '+' || l.src[l.pos] == '-') {
				l.pos++
			}
			for l.pos < len(l.src) && unicode.IsDigit(l.src[l.pos]) {
				l.pos++
			}
		}
	}
	return Token{Type: tokNumber, Value: string(l.src[start:l.pos]), Line: l.line, AfterNL: afterNL}
}

func (l *Lexer) readString(quote rune, afterNL bool) Token {
	l.pos++ // skip opening quote
	var sb strings.Builder
	for l.pos < len(l.src) {
		ch := l.src[l.pos]
		if rune(ch) == quote {
			l.pos++
			break
		}
		if ch == '\\' && l.pos+1 < len(l.src) {
			l.pos++
			esc := l.src[l.pos]
			l.pos++
			switch esc {
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case 'r':
				sb.WriteRune('\r')
			case '\\':
				sb.WriteRune('\\')
			case '\'':
				sb.WriteRune('\'')
			case '"':
				sb.WriteRune('"')
			case '`':
				sb.WriteRune('`')
			case '0':
				sb.WriteRune(0)
			case 'u':
				// \uXXXX
				if l.pos+4 <= len(l.src) {
					hex := string(l.src[l.pos : l.pos+4])
					var r rune
					if _, err := fmt.Sscanf(hex, "%x", &r); err == nil {
						sb.WriteRune(r)
						l.pos += 4
					}
				}
			default:
				sb.WriteRune(esc)
			}
			continue
		}
		if ch == '\n' {
			l.line++
		}
		sb.WriteRune(rune(ch))
		l.pos++
	}
	return Token{Type: tokString, Value: sb.String(), Line: l.line, AfterNL: afterNL}
}

// readTemplate reads a template literal (backtick string).
// Returns a token whose Value is the raw text between backticks,
// including ${ and } markers. The parser will further process this.
func (l *Lexer) readTemplate(afterNL bool) Token {
	l.pos++ // skip opening `
	var sb strings.Builder
	depth := 0
	for l.pos < len(l.src) {
		ch := l.src[l.pos]
		if ch == '`' && depth == 0 {
			l.pos++
			break
		}
		if ch == '$' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '{' {
			depth++
			sb.WriteRune('$')
			sb.WriteRune('{')
			l.pos += 2
			continue
		}
		if ch == '}' && depth > 0 {
			depth--
			sb.WriteRune('}')
			l.pos++
			continue
		}
		if ch == '\\' && l.pos+1 < len(l.src) {
			sb.WriteRune(ch)
			l.pos++
			sb.WriteRune(l.src[l.pos])
			l.pos++
			continue
		}
		if ch == '\n' {
			l.line++
		}
		sb.WriteRune(rune(ch))
		l.pos++
	}
	return Token{Type: tokTemplate, Value: sb.String(), Line: l.line, AfterNL: afterNL}
}

func (l *Lexer) readIdent(afterNL bool) Token {
	start := l.pos
	for l.pos < len(l.src) && (unicode.IsLetter(l.src[l.pos]) || unicode.IsDigit(l.src[l.pos]) || l.src[l.pos] == '_' || l.src[l.pos] == '$') {
		l.pos++
	}
	return Token{Type: tokIdent, Value: string(l.src[start:l.pos]), Line: l.line, AfterNL: afterNL}
}

func (l *Lexer) readOp(afterNL bool) Token {
	ch := l.src[l.pos]
	l.pos++

	peek := func() rune {
		if l.pos < len(l.src) {
			return l.src[l.pos]
		}
		return 0
	}
	consume := func() {
		l.pos++
	}

	tok := func(t TokenType, v string) Token {
		return Token{Type: t, Value: v, Line: l.line, AfterNL: afterNL}
	}

	switch ch {
	case '(':
		return tok(tokLParen, "(")
	case ')':
		return tok(tokRParen, ")")
	case '{':
		return tok(tokLBrace, "{")
	case '}':
		return tok(tokRBrace, "}")
	case '[':
		return tok(tokLBracket, "[")
	case ']':
		return tok(tokRBracket, "]")
	case ';':
		return tok(tokSemicolon, ";")
	case ':':
		return tok(tokColon, ":")
	case ',':
		return tok(tokComma, ",")
	case '~':
		return tok(tokTilde, "~")
	case '.':
		if peek() == '.' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '.' {
			consume()
			consume()
			return tok(tokDotDotDot, "...")
		}
		return tok(tokDot, ".")
	case '?':
		if peek() == '?' {
			consume()
			return tok(tokNullCoal, "??")
		}
		if peek() == '.' {
			consume()
			return tok(tokOptChain, "?.")
		}
		return tok(tokQuestion, "?")
	case '+':
		if peek() == '+' {
			consume()
			return tok(tokPlusPlus, "++")
		}
		if peek() == '=' {
			consume()
			return tok(tokPlusAssign, "+=")
		}
		return tok(tokPlus, "+")
	case '-':
		if peek() == '-' {
			consume()
			return tok(tokMinusMinus, "--")
		}
		if peek() == '=' {
			consume()
			return tok(tokMinusAssign, "-=")
		}
		return tok(tokMinus, "-")
	case '*':
		if peek() == '*' {
			consume()
			return tok(tokStarStar, "**")
		}
		if peek() == '=' {
			consume()
			return tok(tokStarAssign, "*=")
		}
		return tok(tokStar, "*")
	case '/':
		if peek() == '=' {
			consume()
			return tok(tokSlashAssign, "/=")
		}
		return tok(tokSlash, "/")
	case '%':
		if peek() == '=' {
			consume()
			return tok(tokPercentAssign, "%=")
		}
		return tok(tokPercent, "%")
	case '&':
		if peek() == '&' {
			consume()
			return tok(tokAnd, "&&")
		}
		if peek() == '=' {
			consume()
			return tok(tokAmpAssign, "&=")
		}
		return tok(tokAmp, "&")
	case '|':
		if peek() == '|' {
			consume()
			return tok(tokOr, "||")
		}
		if peek() == '=' {
			consume()
			return tok(tokPipeAssign, "|=")
		}
		return tok(tokPipe, "|")
	case '^':
		if peek() == '=' {
			consume()
			return tok(tokCaretAssign, "^=")
		}
		return tok(tokCaret, "^")
	case '<':
		if peek() == '<' {
			consume()
			return tok(tokLShift, "<<")
		}
		if peek() == '=' {
			consume()
			return tok(tokLte, "<=")
		}
		return tok(tokLt, "<")
	case '>':
		if peek() == '>' {
			consume()
			return tok(tokRShift, ">>")
		}
		if peek() == '=' {
			consume()
			return tok(tokGte, ">=")
		}
		return tok(tokGt, ">")
	case '=':
		if peek() == '>' {
			consume()
			return tok(tokFatArrow, "=>")
		}
		if peek() == '=' {
			consume()
			if peek() == '=' {
				consume()
				return tok(tokEqEqEq, "===")
			}
			return tok(tokEqEq, "==")
		}
		return tok(tokAssign, "=")
	case '!':
		if peek() == '=' {
			consume()
			if peek() == '=' {
				consume()
				return tok(tokNeqEq, "!==")
			}
			return tok(tokNeq, "!=")
		}
		return tok(tokNot, "!")
	}

	// Unknown character - skip it
	return Token{Type: tokIdent, Value: string(ch), Line: l.line, AfterNL: afterNL}
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}
