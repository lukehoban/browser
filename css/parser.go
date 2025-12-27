package css

import (
	"github.com/lukehoban/browser/log"
)

// Stylesheet represents a CSS stylesheet.
// CSS 2.1 §4 Syntax and basic data types
type Stylesheet struct {
	Rules []*Rule
}

// Rule represents a CSS rule.
// CSS 2.1 §4.1.7 Rule sets, declaration blocks, and selectors
type Rule struct {
	Selectors    []*Selector
	Declarations []*Declaration
}

// Selector represents a CSS selector.
// CSS 2.1 §5 Selectors
type Selector struct {
	Simple []*SimpleSelector // List of simple selectors (for descendant combinator)
}

// SimpleSelector represents a simple selector.
// CSS 2.1 §5.2 Selector syntax
type SimpleSelector struct {
	TagName string   // Element type selector (e.g., "div", "*" for universal)
	ID      string   // ID selector (e.g., "header")
	Classes []string // Class selectors (e.g., ["container", "main"])
}

// Declaration represents a CSS declaration.
// CSS 2.1 §4.1.8 Declarations and properties
type Declaration struct {
	Property string
	Value    string
}

// Parser parses CSS stylesheets.
type Parser struct {
	tokenizer *Tokenizer
}

// NewParser creates a new CSS parser.
func NewParser(input string) *Parser {
	return &Parser{
		tokenizer: NewTokenizer(input),
	}
}

// Parse parses the CSS input and returns a stylesheet.
func (p *Parser) Parse() *Stylesheet {
	stylesheet := &Stylesheet{
		Rules: make([]*Rule, 0),
	}

	for {
		p.tokenizer.SkipWhitespace()
		token := p.tokenizer.Peek()
		if token.Type == EOFToken {
			break
		}

		// Skip @-rules (media queries, imports, etc.)
		// CSS 2.1 §4.1.5 At-rules - not implementing for simplicity
		if token.Type == AtKeywordToken {
			log.Debugf("Skipping unsupported @-rule: %s", token.Value)
			p.skipAtRule()
			continue
		}

		rule := p.parseRule()
		if rule != nil {
			stylesheet.Rules = append(stylesheet.Rules, rule)
		}
	}

	return stylesheet
}

// skipAtRule skips an @-rule (like @media, @import, @keyframes).
// CSS 2.1 §4.1.5 At-rules
// We skip these because we don't implement them, but we need to properly
// parse past them to avoid infinite loops.
func (p *Parser) skipAtRule() {
	// Consume the @keyword token
	p.tokenizer.Next()

	// Skip tokens until we find either a semicolon (for simple @rules like @import)
	// or a block (for complex @rules like @media)
	braceDepth := 0
	for {
		token := p.tokenizer.Next()
		if token.Type == EOFToken {
			break
		}
		if token.Type == SemicolonToken && braceDepth == 0 {
			break
		}
		if token.Type == LeftBraceToken {
			braceDepth++
		}
		if token.Type == RightBraceToken {
			braceDepth--
			if braceDepth <= 0 {
				break
			}
		}
	}
}

// parseRule parses a CSS rule.
// CSS 2.1 §4.1.7 Rule sets
func (p *Parser) parseRule() *Rule {
	selectors := p.parseSelectors()
	if len(selectors) == 0 {
		return nil
	}

	p.tokenizer.SkipWhitespace()

	// Expect '{'
	token := p.tokenizer.Next()
	if token.Type != LeftBraceToken {
		return nil
	}

	declarations := p.parseDeclarations()

	p.tokenizer.SkipWhitespace()

	// Expect '}'
	token = p.tokenizer.Next()
	if token.Type != RightBraceToken {
		// Error recovery: skip to next '}'
		log.Debugf("CSS parse error: expected '}', got %v, recovering...", token.Type)
		for token.Type != RightBraceToken && token.Type != EOFToken {
			token = p.tokenizer.Next()
		}
	}

	return &Rule{
		Selectors:    selectors,
		Declarations: declarations,
	}
}

// parseSelectors parses a comma-separated list of selectors.
// CSS 2.1 §5.2 Selector syntax
func (p *Parser) parseSelectors() []*Selector {
	selectors := make([]*Selector, 0)

	for {
		p.tokenizer.SkipWhitespace()

		selector := p.parseSelector()
		if selector != nil {
			selectors = append(selectors, selector)
		}

		p.tokenizer.SkipWhitespace()
		token := p.tokenizer.Peek()

		if token.Type == CommaToken {
			p.tokenizer.Next() // consume comma
			continue
		}

		break
	}

	return selectors
}

// parseSelector parses a single selector.
// This handles descendant combinators (space-separated).
// CSS 2.1 §5.5 Descendant selectors
func (p *Parser) parseSelector() *Selector {
	selector := &Selector{
		Simple: make([]*SimpleSelector, 0),
	}

	for {
		p.tokenizer.SkipWhitespace()

		simple := p.parseSimpleSelector()
		if simple == nil {
			break
		}

		selector.Simple = append(selector.Simple, simple)

		// Check for descendant combinator (whitespace followed by another selector)
		savedPos := p.tokenizer.pos
		p.tokenizer.SkipWhitespace()
		next := p.tokenizer.Peek()

		// If next is not a selector start, restore position
		if next.Type != IdentToken && next.Type != HashToken && next.Type != DotToken {
			p.tokenizer.pos = savedPos
			break
		}
	}

	if len(selector.Simple) == 0 {
		return nil
	}

	return selector
}

// parseSimpleSelector parses a simple selector.
// CSS 2.1 §5.2 Selector syntax
func (p *Parser) parseSimpleSelector() *SimpleSelector {
	simple := &SimpleSelector{
		Classes: make([]string, 0),
	}

	token := p.tokenizer.Peek()

	// Type selector
	if token.Type == IdentToken {
		p.tokenizer.Next()
		simple.TagName = token.Value
	}

	// ID and class selectors
	for {
		token = p.tokenizer.Peek()

		if token.Type == HashToken {
			p.tokenizer.Next()
			simple.ID = token.Value
		} else if token.Type == DotToken {
			p.tokenizer.Next()
			// Next token should be class name
			token = p.tokenizer.Next()
			if token.Type == IdentToken {
				simple.Classes = append(simple.Classes, token.Value)
			}
		} else if token.Type == LeftBracketToken {
			// Skip attribute selectors [attr=value]
			// CSS 2.1 §5.8 Attribute selectors - not implementing for simplicity
			// Note: Attribute selectors are part of CSS 2.1 but not core to basic rendering
			log.Debug("Skipping attribute selector (not implemented)")
			p.tokenizer.Next() // consume '['
			// Skip everything until ']'
			for {
				token = p.tokenizer.Next()
				if token.Type == RightBracketToken || token.Type == EOFToken {
					break
				}
			}
		} else if token.Type == ColonToken {
			// Skip pseudo-classes and pseudo-elements (:hover, ::before, etc.)
			// CSS 2.1 §5.11 Pseudo-classes, §5.12 Pseudo-elements
			// Note: We treat selectors with pseudo-classes the same as without them
			// (e.g., "a:link" is treated as "a", "a:visited" is treated as "a")
			log.Debug("Skipping pseudo-class/pseudo-element (partial support)")
			p.tokenizer.Next() // consume ':'
			
			// Check for double colon (pseudo-element ::before)
			token = p.tokenizer.Peek()
			if token.Type == ColonToken {
				p.tokenizer.Next() // consume second ':'
			}
			
			// Consume the pseudo-class/pseudo-element name
			token = p.tokenizer.Peek()
			if token.Type == IdentToken {
				p.tokenizer.Next() // consume identifier (e.g., "link", "hover", "before")
			}
			
			// Handle functional pseudo-classes like :nth-child(2n+1)
			token = p.tokenizer.Peek()
			if token.Type == LeftParenToken {
				p.tokenizer.Next() // consume '('
				// Skip everything until ')'
				parenDepth := 1
				for parenDepth > 0 {
					token = p.tokenizer.Next()
					if token.Type == EOFToken {
						break
					}
					if token.Type == LeftParenToken {
						parenDepth++
					} else if token.Type == RightParenToken {
						parenDepth--
					}
				}
			}
		} else {
			break
		}
	}

	// Check if we actually parsed anything
	if simple.TagName == "" && simple.ID == "" && len(simple.Classes) == 0 {
		return nil
	}

	return simple
}

// parseDeclarations parses declarations within a rule.
// CSS 2.1 §4.1.8 Declarations and properties
func (p *Parser) parseDeclarations() []*Declaration {
	declarations := make([]*Declaration, 0)

	for {
		p.tokenizer.SkipWhitespace()

		token := p.tokenizer.Peek()
		if token.Type == RightBraceToken || token.Type == EOFToken {
			break
		}

		decl := p.parseDeclaration()
		if decl != nil {
			declarations = append(declarations, decl)
		}

		p.tokenizer.SkipWhitespace()

		// Expect ';' or '}'
		token = p.tokenizer.Peek()
		if token.Type == SemicolonToken {
			p.tokenizer.Next()
		} else if token.Type == RightBraceToken {
			break
		}
	}

	return declarations
}

// parseDeclaration parses a single declaration.
// CSS 2.1 §4.1.8 Declarations and properties
func (p *Parser) parseDeclaration() *Declaration {
	p.tokenizer.SkipWhitespace()

	// Property name
	token := p.tokenizer.Next()
	if token.Type != IdentToken {
		return nil
	}
	property := token.Value

	p.tokenizer.SkipWhitespace()

	// Expect ':'
	token = p.tokenizer.Next()
	if token.Type != ColonToken {
		return nil
	}

	p.tokenizer.SkipWhitespace()

	// Parse value (simplified - just concatenate tokens until ';' or '}')
	value := ""
	for {
		token = p.tokenizer.Peek()
		if token.Type == SemicolonToken || token.Type == RightBraceToken || token.Type == EOFToken {
			break
		}

		p.tokenizer.Next()

		if token.Type == WhitespaceToken {
			if value != "" {
				value += " "
			}
		} else if token.Type == HashToken {
			// CSS 2.1 §4.3.6: Preserve # prefix for color values
			value += "#" + token.Value
		} else {
			value += token.Value
		}
	}

	return &Declaration{
		Property: property,
		Value:    value,
	}
}

// Parse is a convenience function to parse CSS.
func Parse(input string) *Stylesheet {
	parser := NewParser(input)
	return parser.Parse()
}

// ParseInlineStyle parses inline style declarations from a style attribute.
// CSS 2.1 §6.4.3: Inline styles have specificity A=1, higher than any other selector.
// Unlike regular CSS rules, inline styles don't have selectors or braces - just declarations.
func ParseInlineStyle(styleAttr string) []*Declaration {
	if styleAttr == "" {
		return nil
	}
	
	parser := NewParser(styleAttr)
	declarations := make([]*Declaration, 0)
	
	for {
		parser.tokenizer.SkipWhitespace()
		
		token := parser.tokenizer.Peek()
		if token.Type == EOFToken {
			break
		}
		
		decl := parser.parseDeclaration()
		if decl != nil {
			declarations = append(declarations, decl)
		}
		
		parser.tokenizer.SkipWhitespace()
		
		// Expect ';' or EOF
		token = parser.tokenizer.Peek()
		if token.Type == SemicolonToken {
			parser.tokenizer.Next()
		}
	}
	
	return declarations
}
