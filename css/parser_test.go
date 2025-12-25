package css

import "testing"

func TestParseSimpleRule(t *testing.T) {
	input := "div { color: red; }"
	stylesheet := Parse(input)

	if len(stylesheet.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(stylesheet.Rules))
	}

	rule := stylesheet.Rules[0]
	if len(rule.Selectors) != 1 {
		t.Fatalf("Expected 1 selector, got %d", len(rule.Selectors))
	}

	selector := rule.Selectors[0]
	if len(selector.Simple) != 1 {
		t.Fatalf("Expected 1 simple selector, got %d", len(selector.Simple))
	}

	simple := selector.Simple[0]
	if simple.TagName != "div" {
		t.Errorf("Expected tag 'div', got %v", simple.TagName)
	}

	if len(rule.Declarations) != 1 {
		t.Fatalf("Expected 1 declaration, got %d", len(rule.Declarations))
	}

	decl := rule.Declarations[0]
	if decl.Property != "color" {
		t.Errorf("Expected property 'color', got %v", decl.Property)
	}
	if decl.Value != "red" {
		t.Errorf("Expected value 'red', got %v", decl.Value)
	}
}

func TestParseIDSelector(t *testing.T) {
	input := "#header { font-size: 20px; }"
	stylesheet := Parse(input)

	if len(stylesheet.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(stylesheet.Rules))
	}

	simple := stylesheet.Rules[0].Selectors[0].Simple[0]
	if simple.ID != "header" {
		t.Errorf("Expected ID 'header', got %v", simple.ID)
	}
}

func TestParseClassSelector(t *testing.T) {
	input := ".container { width: 100px; }"
	stylesheet := Parse(input)

	if len(stylesheet.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(stylesheet.Rules))
	}

	simple := stylesheet.Rules[0].Selectors[0].Simple[0]
	if len(simple.Classes) != 1 {
		t.Fatalf("Expected 1 class, got %d", len(simple.Classes))
	}
	if simple.Classes[0] != "container" {
		t.Errorf("Expected class 'container', got %v", simple.Classes[0])
	}
}

func TestParseCombinedSelector(t *testing.T) {
	input := "div#main.container { margin: 10px; }"
	stylesheet := Parse(input)

	if len(stylesheet.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(stylesheet.Rules))
	}

	simple := stylesheet.Rules[0].Selectors[0].Simple[0]
	if simple.TagName != "div" {
		t.Errorf("Expected tag 'div', got %v", simple.TagName)
	}
	if simple.ID != "main" {
		t.Errorf("Expected ID 'main', got %v", simple.ID)
	}
	if len(simple.Classes) != 1 || simple.Classes[0] != "container" {
		t.Errorf("Expected class 'container', got %v", simple.Classes)
	}
}

func TestParseMultipleClasses(t *testing.T) {
	input := ".container.active { display: block; }"
	stylesheet := Parse(input)

	if len(stylesheet.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(stylesheet.Rules))
	}

	simple := stylesheet.Rules[0].Selectors[0].Simple[0]
	if len(simple.Classes) != 2 {
		t.Fatalf("Expected 2 classes, got %d", len(simple.Classes))
	}
	if simple.Classes[0] != "container" {
		t.Errorf("Expected first class 'container', got %v", simple.Classes[0])
	}
	if simple.Classes[1] != "active" {
		t.Errorf("Expected second class 'active', got %v", simple.Classes[1])
	}
}

func TestParseDescendantSelector(t *testing.T) {
	input := "div p { color: blue; }"
	stylesheet := Parse(input)

	if len(stylesheet.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(stylesheet.Rules))
	}

	selector := stylesheet.Rules[0].Selectors[0]
	if len(selector.Simple) != 2 {
		t.Fatalf("Expected 2 simple selectors (descendant), got %d", len(selector.Simple))
	}

	if selector.Simple[0].TagName != "div" {
		t.Errorf("Expected first selector 'div', got %v", selector.Simple[0].TagName)
	}
	if selector.Simple[1].TagName != "p" {
		t.Errorf("Expected second selector 'p', got %v", selector.Simple[1].TagName)
	}
}

func TestParseMultipleSelectors(t *testing.T) {
	input := "h1, h2, h3 { font-weight: bold; }"
	stylesheet := Parse(input)

	if len(stylesheet.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(stylesheet.Rules))
	}

	rule := stylesheet.Rules[0]
	if len(rule.Selectors) != 3 {
		t.Fatalf("Expected 3 selectors, got %d", len(rule.Selectors))
	}

	tags := []string{"h1", "h2", "h3"}
	for i, tag := range tags {
		if rule.Selectors[i].Simple[0].TagName != tag {
			t.Errorf("Expected selector %d to be '%s', got %v", i, tag, rule.Selectors[i].Simple[0].TagName)
		}
	}
}

func TestParseMultipleDeclarations(t *testing.T) {
	input := "div { color: red; background: blue; margin: 10px; }"
	stylesheet := Parse(input)

	if len(stylesheet.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(stylesheet.Rules))
	}

	rule := stylesheet.Rules[0]
	if len(rule.Declarations) != 3 {
		t.Fatalf("Expected 3 declarations, got %d", len(rule.Declarations))
	}

	expected := map[string]string{
		"color":      "red",
		"background": "blue",
		"margin":     "10px",
	}

	for _, decl := range rule.Declarations {
		expectedValue, ok := expected[decl.Property]
		if !ok {
			t.Errorf("Unexpected property: %v", decl.Property)
			continue
		}
		if decl.Value != expectedValue {
			t.Errorf("Property %v: expected value %v, got %v", decl.Property, expectedValue, decl.Value)
		}
	}
}

func TestParseMultipleRules(t *testing.T) {
	input := `
		div { color: red; }
		p { font-size: 14px; }
		.container { width: 100%; }
	`
	stylesheet := Parse(input)

	if len(stylesheet.Rules) != 3 {
		t.Fatalf("Expected 3 rules, got %d", len(stylesheet.Rules))
	}
}

func TestParseComplexValue(t *testing.T) {
	input := "div { border: 1px solid black; }"
	stylesheet := Parse(input)

	if len(stylesheet.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(stylesheet.Rules))
	}

	decl := stylesheet.Rules[0].Declarations[0]
	if decl.Property != "border" {
		t.Errorf("Expected property 'border', got %v", decl.Property)
	}
	if decl.Value != "1px solid black" {
		t.Errorf("Expected value '1px solid black', got %v", decl.Value)
	}
}

// TestParseAttributeSelector tests that attribute selectors are skipped gracefully.
// CSS 2.1 §5.8 Attribute selectors
func TestParseAttributeSelector(t *testing.T) {
	input := `
input[type='submit'] { font-family: Verdana; }
.class { color: red; }
`
	stylesheet := Parse(input)

	// Should parse successfully and have at least the .class rule
	if len(stylesheet.Rules) < 1 {
		t.Errorf("Expected at least 1 rule, got %d", len(stylesheet.Rules))
	}

	// The .class rule should be parsed correctly
	foundClassRule := false
	for _, rule := range stylesheet.Rules {
		if len(rule.Selectors) > 0 && len(rule.Selectors[0].Simple) > 0 {
			simple := rule.Selectors[0].Simple[0]
			if len(simple.Classes) > 0 && simple.Classes[0] == "class" {
				foundClassRule = true
				break
			}
		}
	}

	if !foundClassRule {
		t.Error("Expected .class rule to be parsed")
	}
}

// TestParseAtRule tests that @-rules are skipped gracefully.
// CSS 2.1 §4.1.5 At-rules
func TestParseAtRule(t *testing.T) {
	input := `
body { color: black; }
@media screen and (max-width: 600px) {
body { color: blue; }
}
.test { color: red; }
`
	stylesheet := Parse(input)

	// Should parse successfully and have the body and .test rules
	if len(stylesheet.Rules) < 2 {
		t.Errorf("Expected at least 2 rules, got %d", len(stylesheet.Rules))
	}

	// Check that we have body and .test rules
	foundBody := false
	foundTest := false
	for _, rule := range stylesheet.Rules {
		if len(rule.Selectors) > 0 && len(rule.Selectors[0].Simple) > 0 {
			simple := rule.Selectors[0].Simple[0]
			if simple.TagName == "body" {
				foundBody = true
			}
			if len(simple.Classes) > 0 && simple.Classes[0] == "test" {
				foundTest = true
			}
		}
	}

	if !foundBody {
		t.Error("Expected body rule to be parsed")
	}
	if !foundTest {
		t.Error("Expected .test rule to be parsed")
	}
}

// SKIPPED TESTS FOR KNOWN BROKEN/UNIMPLEMENTED FEATURES
// These tests document known limitations that need to be implemented.
// See MILESTONES.md for more details.

func TestParsePseudoClasses_Skipped(t *testing.T) {
	t.Skip("Pseudo-classes not implemented - CSS 2.1 §5.11")
	// CSS 2.1 §5.11 Pseudo-classes
	// Common pseudo-classes include :hover, :active, :focus, :first-child, :last-child, :nth-child
	
	input := "a:hover { color: red; } p:first-child { margin-top: 0; }"
	_ = Parse(input)
	
	// When implemented, parser should recognize pseudo-classes
	// Expected structure would include PseudoClass field in SimpleSelector
	// For now, this gracefully fails/skips the selector
}

func TestParsePseudoElements_Skipped(t *testing.T) {
	t.Skip("Pseudo-elements not implemented - CSS 2.1 §5.12")
	// CSS 2.1 §5.12 Pseudo-elements
	// Common pseudo-elements include ::before, ::after, ::first-line, ::first-letter
	
	input := "p::before { content: '→ '; } p::after { content: ' ←'; }"
	_ = Parse(input)
	
	// When implemented, parser should recognize pseudo-elements
	// Expected structure would include PseudoElement field in SimpleSelector
	// For now, this gracefully fails/skips the selector
}

func TestParseChildCombinator_Skipped(t *testing.T) {
	t.Skip("Child combinator not implemented - CSS 2.1 §5.5")
	// CSS 2.1 §5.5 Child selectors
	// The child combinator (>) selects direct children only
	
	input := "div > p { color: blue; }"
	_ = Parse(input)
	
	// When implemented, the Selector structure would include a Combinator field
	// to distinguish between descendant (space) and child (>) combinators
	// For now, this might parse but treat '>' as a descendant combinator
}

func TestParseAdjacentSiblingCombinator_Skipped(t *testing.T) {
	t.Skip("Adjacent sibling combinator not implemented - CSS 2.1 §5.7")
	// CSS 2.1 §5.7 Adjacent sibling selectors
	// The adjacent sibling combinator (+) selects the immediately following sibling
	
	input := "h1 + p { margin-top: 0; }"
	_ = Parse(input)
	
	// When implemented, would need Combinator field to distinguish sibling selectors
	// For now, this might not parse correctly
}

func TestParseGeneralSiblingCombinator_Skipped(t *testing.T) {
	t.Skip("General sibling combinator not implemented - CSS 3 Selectors")
	// CSS 3 Selectors §8.3.2 General sibling combinator
	// The general sibling combinator (~) selects all following siblings
	
	input := "h1 ~ p { color: gray; }"
	_ = Parse(input)
	
	// When implemented, would need Combinator field for general sibling matching
	// For now, this might not parse correctly
}

func TestParseAttributeSelectorMatching_Skipped(t *testing.T) {
	t.Skip("Attribute selector matching not implemented - CSS 2.1 §5.8")
	// CSS 2.1 §5.8 Attribute selectors
	// While parsing is gracefully handled, actual matching is not implemented
	// This test would verify that matching works correctly
	
	// The parser skips attribute selectors, but they should be parsed
	// into the data structure and matched during style computation
	input := `input[type="text"] { border: 1px solid black; }`
	_ = Parse(input)
	
	// When implemented, SimpleSelector would have Attributes field
	// For now, attribute selectors are skipped during parsing
}

// TestParseInlineStyle tests parsing of inline style attributes.
// CSS 2.1 §6.4.3: Inline styles have the highest specificity.
func TestParseInlineStyle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*Declaration
	}{
		{
			name:  "single declaration",
			input: "color: red",
			expected: []*Declaration{
				{Property: "color", Value: "red"},
			},
		},
		{
			name:  "single declaration with semicolon",
			input: "color: red;",
			expected: []*Declaration{
				{Property: "color", Value: "red"},
			},
		},
		{
			name:  "multiple declarations",
			input: "color: red; font-size: 16px",
			expected: []*Declaration{
				{Property: "color", Value: "red"},
				{Property: "font-size", Value: "16px"},
			},
		},
		{
			name:  "multiple declarations with trailing semicolon",
			input: "color: red; font-size: 16px;",
			expected: []*Declaration{
				{Property: "color", Value: "red"},
				{Property: "font-size", Value: "16px"},
			},
		},
		{
			name:  "multiple declarations with spaces",
			input: "  color: red;  font-size: 16px;  background: blue;  ",
			expected: []*Declaration{
				{Property: "color", Value: "red"},
				{Property: "font-size", Value: "16px"},
				{Property: "background", Value: "blue"},
			},
		},
		{
			name:  "complex values",
			input: "margin: 10px 20px 30px 40px; padding: 5px;",
			expected: []*Declaration{
				{Property: "margin", Value: "10px 20px 30px 40px"},
				{Property: "padding", Value: "5px"},
			},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: []*Declaration{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			declarations := ParseInlineStyle(tt.input)

			if tt.expected == nil {
				if declarations != nil {
					t.Errorf("Expected nil, got %v", declarations)
				}
				return
			}

			if len(declarations) != len(tt.expected) {
				t.Errorf("Expected %d declarations, got %d", len(tt.expected), len(declarations))
				return
			}

			for i, expected := range tt.expected {
				if declarations[i].Property != expected.Property {
					t.Errorf("Declaration %d: expected property '%s', got '%s'",
						i, expected.Property, declarations[i].Property)
				}
				if declarations[i].Value != expected.Value {
					t.Errorf("Declaration %d: expected value '%s', got '%s'",
						i, expected.Value, declarations[i].Value)
				}
			}
		})
	}
}
