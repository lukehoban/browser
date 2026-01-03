package style

import (
	"testing"

	"github.com/lukehoban/browser/css"
	"github.com/lukehoban/browser/dom"
)

func TestMatchesSimpleSelector(t *testing.T) {
	tests := []struct {
		name     string
		node     *dom.Node
		selector *css.SimpleSelector
		expected bool
	}{
		{
			name:     "match tag name",
			node:     dom.NewElement("div"),
			selector: &css.SimpleSelector{TagName: "div"},
			expected: true,
		},
		{
			name:     "no match tag name",
			node:     dom.NewElement("div"),
			selector: &css.SimpleSelector{TagName: "p"},
			expected: false,
		},
		{
			name: "match ID",
			node: func() *dom.Node {
				n := dom.NewElement("div")
				n.SetAttribute("id", "header")
				return n
			}(),
			selector: &css.SimpleSelector{ID: "header"},
			expected: true,
		},
		{
			name: "no match ID",
			node: func() *dom.Node {
				n := dom.NewElement("div")
				n.SetAttribute("id", "header")
				return n
			}(),
			selector: &css.SimpleSelector{ID: "footer"},
			expected: false,
		},
		{
			name: "match class",
			node: func() *dom.Node {
				n := dom.NewElement("div")
				n.SetAttribute("class", "container")
				return n
			}(),
			selector: &css.SimpleSelector{Classes: []string{"container"}},
			expected: true,
		},
		{
			name: "match multiple classes",
			node: func() *dom.Node {
				n := dom.NewElement("div")
				n.SetAttribute("class", "container active main")
				return n
			}(),
			selector: &css.SimpleSelector{Classes: []string{"container", "active"}},
			expected: true,
		},
		{
			name: "no match class",
			node: func() *dom.Node {
				n := dom.NewElement("div")
				n.SetAttribute("class", "container")
				return n
			}(),
			selector: &css.SimpleSelector{Classes: []string{"footer"}},
			expected: false,
		},
		{
			name: "match tag and ID",
			node: func() *dom.Node {
				n := dom.NewElement("div")
				n.SetAttribute("id", "main")
				return n
			}(),
			selector: &css.SimpleSelector{TagName: "div", ID: "main"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesSimpleSelector(tt.node, tt.selector)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCalculateSpecificity(t *testing.T) {
	tests := []struct {
		name     string
		selector *css.Selector
		expected Specificity
	}{
		{
			name: "element selector",
			selector: &css.Selector{
				Simple: []*css.SimpleSelector{
					{TagName: "div"},
				},
			},
			expected: Specificity{A: 0, B: 0, C: 0, D: 1},
		},
		{
			name: "ID selector",
			selector: &css.Selector{
				Simple: []*css.SimpleSelector{
					{ID: "header"},
				},
			},
			expected: Specificity{A: 0, B: 1, C: 0, D: 0},
		},
		{
			name: "class selector",
			selector: &css.Selector{
				Simple: []*css.SimpleSelector{
					{Classes: []string{"container"}},
				},
			},
			expected: Specificity{A: 0, B: 0, C: 1, D: 0},
		},
		{
			name: "combined selector",
			selector: &css.Selector{
				Simple: []*css.SimpleSelector{
					{TagName: "div", ID: "main", Classes: []string{"container", "active"}},
				},
			},
			expected: Specificity{A: 0, B: 1, C: 2, D: 1},
		},
		{
			name: "descendant selector",
			selector: &css.Selector{
				Simple: []*css.SimpleSelector{
					{TagName: "div"},
					{TagName: "p"},
				},
			},
			expected: Specificity{A: 0, B: 0, C: 0, D: 2},
		},
		{
			name: "pseudo-class selector (a:link)",
			selector: &css.Selector{
				Simple: []*css.SimpleSelector{
					{TagName: "a", PseudoClasses: []string{"link"}},
				},
			},
			expected: Specificity{A: 0, B: 0, C: 1, D: 1},
		},
		{
			name: "class with pseudo-class (.comhead a:link)",
			selector: &css.Selector{
				Simple: []*css.SimpleSelector{
					{Classes: []string{"comhead"}},
					{TagName: "a", PseudoClasses: []string{"link"}},
				},
			},
			expected: Specificity{A: 0, B: 0, C: 2, D: 1},
		},
		{
			name: "multiple pseudo-classes",
			selector: &css.Selector{
				Simple: []*css.SimpleSelector{
					{TagName: "a", PseudoClasses: []string{"link", "hover"}},
				},
			},
			expected: Specificity{A: 0, B: 0, C: 2, D: 1},
		},
		{
			name: "pseudo-element (::before)",
			selector: &css.Selector{
				Simple: []*css.SimpleSelector{
					{TagName: "p", PseudoElements: []string{"before"}},
				},
			},
			expected: Specificity{A: 0, B: 0, C: 0, D: 2},
		},
		{
			name: "pseudo-class and pseudo-element",
			selector: &css.Selector{
				Simple: []*css.SimpleSelector{
					{TagName: "a", PseudoClasses: []string{"hover"}, PseudoElements: []string{"after"}},
				},
			},
			expected: Specificity{A: 0, B: 0, C: 1, D: 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateSpecificity(tt.selector)
			if result != tt.expected {
				t.Errorf("Expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}

func TestSpecificityCompare(t *testing.T) {
	tests := []struct {
		name     string
		s1       Specificity
		s2       Specificity
		expected int // positive if s1 > s2, 0 if equal, negative if s1 < s2
	}{
		{
			name:     "equal",
			s1:       Specificity{A: 0, B: 0, C: 1, D: 1},
			s2:       Specificity{A: 0, B: 0, C: 1, D: 1},
			expected: 0,
		},
		{
			name:     "ID beats class",
			s1:       Specificity{A: 0, B: 1, C: 0, D: 0},
			s2:       Specificity{A: 0, B: 0, C: 10, D: 10},
			expected: 1,
		},
		{
			name:     "class beats element",
			s1:       Specificity{A: 0, B: 0, C: 1, D: 0},
			s2:       Specificity{A: 0, B: 0, C: 0, D: 10},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.s1.Compare(tt.s2)
			if (result > 0) != (tt.expected > 0) ||
				(result == 0) != (tt.expected == 0) ||
				(result < 0) != (tt.expected < 0) {
				t.Errorf("Expected comparison result %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestStyleTree(t *testing.T) {
	// Create a simple DOM tree
	doc := dom.NewDocument()
	div := dom.NewElement("div")
	div.SetAttribute("id", "main")
	div.SetAttribute("class", "container")
	p := dom.NewElement("p")
	text := dom.NewText("Hello")
	p.AppendChild(text)
	div.AppendChild(p)
	doc.AppendChild(div)

	// Create a stylesheet
	stylesheet := &css.Stylesheet{
		Rules: []*css.Rule{
			{
				Selectors: []*css.Selector{
					{Simple: []*css.SimpleSelector{{TagName: "div"}}},
				},
				Declarations: []*css.Declaration{
					{Property: "color", Value: "red"},
				},
			},
			{
				Selectors: []*css.Selector{
					{Simple: []*css.SimpleSelector{{ID: "main"}}},
				},
				Declarations: []*css.Declaration{
					{Property: "background", Value: "blue"},
				},
			},
			{
				Selectors: []*css.Selector{
					{Simple: []*css.SimpleSelector{{Classes: []string{"container"}}}},
				},
				Declarations: []*css.Declaration{
					{Property: "margin", Value: "10px"},
				},
			},
		},
	}

	// Style the tree
	styledTree := StyleTree(doc, stylesheet)

	// Check that div has all three properties
	divStyled := styledTree.Children[0]
	if divStyled.Styles["color"] != "red" {
		t.Errorf("Expected color 'red', got %v", divStyled.Styles["color"])
	}
	if divStyled.Styles["background"] != "blue" {
		t.Errorf("Expected background 'blue', got %v", divStyled.Styles["background"])
	}
	// Check that margin shorthand was expanded to longhand properties
	if divStyled.Styles["margin-top"] != "10px" {
		t.Errorf("Expected margin-top '10px', got %v", divStyled.Styles["margin-top"])
	}
	if divStyled.Styles["margin-right"] != "10px" {
		t.Errorf("Expected margin-right '10px', got %v", divStyled.Styles["margin-right"])
	}
	if divStyled.Styles["margin-bottom"] != "10px" {
		t.Errorf("Expected margin-bottom '10px', got %v", divStyled.Styles["margin-bottom"])
	}
	if divStyled.Styles["margin-left"] != "10px" {
		t.Errorf("Expected margin-left '10px', got %v", divStyled.Styles["margin-left"])
	}
}

func TestDescendantSelector(t *testing.T) {
	// Create DOM: div > p > span
	div := dom.NewElement("div")
	p := dom.NewElement("p")
	span := dom.NewElement("span")
	div.AppendChild(p)
	p.AppendChild(span)

	// Selector: div span
	selector := &css.Selector{
		Simple: []*css.SimpleSelector{
			{TagName: "div"},
			{TagName: "span"},
		},
	}

	if !matchesSelector(span, selector) {
		t.Error("Expected span to match 'div span' selector")
	}

	// Selector: div p
	selector2 := &css.Selector{
		Simple: []*css.SimpleSelector{
			{TagName: "div"},
			{TagName: "p"},
		},
	}

	if !matchesSelector(p, selector2) {
		t.Error("Expected p to match 'div p' selector")
	}

	// div should not match "div span"
	if matchesSelector(div, selector) {
		t.Error("Expected div not to match 'div span' selector")
	}
}

func TestExpandShorthand(t *testing.T) {
	tests := []struct {
		name     string
		property string
		value    string
		expected map[string]string
	}{
		{
			name:     "margin with 1 value",
			property: "margin",
			value:    "10px",
			expected: map[string]string{
				"margin-top":    "10px",
				"margin-right":  "10px",
				"margin-bottom": "10px",
				"margin-left":   "10px",
			},
		},
		{
			name:     "margin with 2 values",
			property: "margin",
			value:    "10px 20px",
			expected: map[string]string{
				"margin-top":    "10px",
				"margin-right":  "20px",
				"margin-bottom": "10px",
				"margin-left":   "20px",
			},
		},
		{
			name:     "margin with 3 values",
			property: "margin",
			value:    "10px 20px 30px",
			expected: map[string]string{
				"margin-top":    "10px",
				"margin-right":  "20px",
				"margin-bottom": "30px",
				"margin-left":   "20px",
			},
		},
		{
			name:     "margin with 4 values",
			property: "margin",
			value:    "10px 20px 30px 40px",
			expected: map[string]string{
				"margin-top":    "10px",
				"margin-right":  "20px",
				"margin-bottom": "30px",
				"margin-left":   "40px",
			},
		},
		{
			name:     "padding with 1 value",
			property: "padding",
			value:    "5px",
			expected: map[string]string{
				"padding-top":    "5px",
				"padding-right":  "5px",
				"padding-bottom": "5px",
				"padding-left":   "5px",
			},
		},
		{
			name:     "padding with 2 values",
			property: "padding",
			value:    "5px 10px",
			expected: map[string]string{
				"padding-top":    "5px",
				"padding-right":  "10px",
				"padding-bottom": "5px",
				"padding-left":   "10px",
			},
		},
		{
			name:     "non-shorthand property",
			property: "color",
			value:    "red",
			expected: map[string]string{
				"color": "red",
			},
		},
		{
			name:     "margin-top longhand",
			property: "margin-top",
			value:    "15px",
			expected: map[string]string{
				"margin-top": "15px",
			},
		},
		{
			name:     "border shorthand with width, style, and color",
			property: "border",
			value:    "2px solid #2196F3",
			expected: map[string]string{
				"border-top-width":    "2px",
				"border-right-width":  "2px",
				"border-bottom-width": "2px",
				"border-left-width":   "2px",
				"border-style":        "solid",
				"border-color":        "#2196F3",
			},
		},
		{
			name:     "border-bottom shorthand",
			property: "border-bottom",
			value:    "3px solid #4CAF50",
			expected: map[string]string{
				"border-bottom-width": "3px",
				"border-style":        "solid",
				"border-color":        "#4CAF50",
			},
		},
		{
			name:     "border shorthand with named color",
			property: "border",
			value:    "1px dashed red",
			expected: map[string]string{
				"border-top-width":    "1px",
				"border-right-width":  "1px",
				"border-bottom-width": "1px",
				"border-left-width":   "1px",
				"border-style":        "dashed",
				"border-color":        "red",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandShorthand(tt.property, tt.value)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d properties, got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				if result[key] != expectedValue {
					t.Errorf("For property %s, expected %s, got %s", key, expectedValue, result[key])
				}
			}
		})
	}
}

func TestSplitWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single value",
			input:    "10px",
			expected: []string{"10px"},
		},
		{
			name:     "multiple values with spaces",
			input:    "10px 20px 30px",
			expected: []string{"10px", "20px", "30px"},
		},
		{
			name:     "multiple spaces",
			input:    "10px  20px   30px",
			expected: []string{"10px", "20px", "30px"},
		},
		{
			name:     "tabs and spaces",
			input:    "10px\t20px 30px",
			expected: []string{"10px", "20px", "30px"},
		},
		{
			name:     "newlines",
			input:    "10px\n20px",
			expected: []string{"10px", "20px"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitWhitespace(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d values, got %d", len(tt.expected), len(result))
			}

			for i, val := range tt.expected {
				if i >= len(result) || result[i] != val {
					t.Errorf("Expected value %d to be %s, got %s", i, val, result[i])
				}
			}
		})
	}
}

func TestShorthandIntegration(t *testing.T) {
	// Create a simple DOM tree
	doc := dom.NewDocument()
	div := dom.NewElement("div")
	doc.AppendChild(div)

	// Create a stylesheet with shorthand properties
	stylesheet := &css.Stylesheet{
		Rules: []*css.Rule{
			{
				Selectors: []*css.Selector{
					{Simple: []*css.SimpleSelector{{TagName: "div"}}},
				},
				Declarations: []*css.Declaration{
					{Property: "margin", Value: "20px"},
					{Property: "padding", Value: "10px"},
				},
			},
		},
	}

	// Style the tree
	styledTree := StyleTree(doc, stylesheet)

	// Check that div has expanded longhand properties
	divStyled := styledTree.Children[0]

	// Check margin properties
	if divStyled.Styles["margin-top"] != "20px" {
		t.Errorf("Expected margin-top '20px', got %v", divStyled.Styles["margin-top"])
	}
	if divStyled.Styles["margin-right"] != "20px" {
		t.Errorf("Expected margin-right '20px', got %v", divStyled.Styles["margin-right"])
	}
	if divStyled.Styles["margin-bottom"] != "20px" {
		t.Errorf("Expected margin-bottom '20px', got %v", divStyled.Styles["margin-bottom"])
	}
	if divStyled.Styles["margin-left"] != "20px" {
		t.Errorf("Expected margin-left '20px', got %v", divStyled.Styles["margin-left"])
	}

	// Check padding properties
	if divStyled.Styles["padding-top"] != "10px" {
		t.Errorf("Expected padding-top '10px', got %v", divStyled.Styles["padding-top"])
	}
	if divStyled.Styles["padding-right"] != "10px" {
		t.Errorf("Expected padding-right '10px', got %v", divStyled.Styles["padding-right"])
	}
	if divStyled.Styles["padding-bottom"] != "10px" {
		t.Errorf("Expected padding-bottom '10px', got %v", divStyled.Styles["padding-bottom"])
	}
	if divStyled.Styles["padding-left"] != "10px" {
		t.Errorf("Expected padding-left '10px', got %v", divStyled.Styles["padding-left"])
	}
}

// SKIPPED TESTS FOR KNOWN BROKEN/UNIMPLEMENTED FEATURES
// These tests document known limitations that need to be implemented.
// See MILESTONES.md for more details.

func TestCSSInheritance_Skipped(t *testing.T) {
	t.Skip("CSS inheritance not implemented - CSS 2.1 §6.2")
	// CSS 2.1 §6.2 Inheritance
	// Inheritable properties (color, font-*, text-*, etc.) should propagate from parent to child
	
	// Create DOM: div > p > span
	doc := dom.NewDocument()
	div := dom.NewElement("div")
	p := dom.NewElement("p")
	span := dom.NewElement("span")
	text := dom.NewText("Hello")
	span.AppendChild(text)
	p.AppendChild(span)
	div.AppendChild(p)
	doc.AppendChild(div)
	
	// Style only the div with color
	stylesheet := &css.Stylesheet{
		Rules: []*css.Rule{
			{
				Selectors: []*css.Selector{
					{Simple: []*css.SimpleSelector{{TagName: "div"}}},
				},
				Declarations: []*css.Declaration{
					{Property: "color", Value: "red"},
					{Property: "font-size", Value: "16px"},
				},
			},
		},
	}
	
	styledTree := StyleTree(doc, stylesheet)
	divStyled := styledTree.Children[0]
	pStyled := divStyled.Children[0]
	spanStyled := pStyled.Children[0]
	
	// Div should have the explicit styles
	if divStyled.Styles["color"] != "red" {
		t.Errorf("Expected div color 'red', got %v", divStyled.Styles["color"])
	}
	if divStyled.Styles["font-size"] != "16px" {
		t.Errorf("Expected div font-size '16px', got %v", divStyled.Styles["font-size"])
	}
	
	// P should inherit color and font-size from div
	if pStyled.Styles["color"] != "red" {
		t.Errorf("Expected p to inherit color 'red', got %v", pStyled.Styles["color"])
	}
	if pStyled.Styles["font-size"] != "16px" {
		t.Errorf("Expected p to inherit font-size '16px', got %v", pStyled.Styles["font-size"])
	}
	
	// Span should inherit color and font-size from p
	if spanStyled.Styles["color"] != "red" {
		t.Errorf("Expected span to inherit color 'red', got %v", spanStyled.Styles["color"])
	}
	if spanStyled.Styles["font-size"] != "16px" {
		t.Errorf("Expected span to inherit font-size '16px', got %v", spanStyled.Styles["font-size"])
	}
}

func TestImportantDeclarations_Skipped(t *testing.T) {
	t.Skip("!important declarations not implemented - CSS 2.1 §6.4.2")
	// CSS 2.1 §6.4.2 !important rules
	// !important declarations should override normal declarations regardless of specificity
	
	doc := dom.NewDocument()
	div := dom.NewElement("div")
	div.SetAttribute("id", "main")
	doc.AppendChild(div)
	
	// When implemented, Declaration would have an Important field
	stylesheet := &css.Stylesheet{
		Rules: []*css.Rule{
			{
				Selectors: []*css.Selector{
					{Simple: []*css.SimpleSelector{{TagName: "div"}}},
				},
				Declarations: []*css.Declaration{
					{Property: "color", Value: "red"}, // Would have Important: true
				},
			},
			{
				Selectors: []*css.Selector{
					{Simple: []*css.SimpleSelector{{ID: "main"}}},
				},
				Declarations: []*css.Declaration{
					{Property: "color", Value: "blue"},
				},
			},
		},
	}
	
	styledTree := StyleTree(doc, stylesheet)
	divStyled := styledTree.Children[0]
	
	// The !important declaration should win even though ID selector has higher specificity
	// Currently, ID selector wins due to higher specificity
	if divStyled.Styles["color"] != "blue" {
		t.Errorf("Expected color 'blue' (ID has higher specificity), got %v", divStyled.Styles["color"])
	}
}

func TestComputedValues_Skipped(t *testing.T) {
	t.Skip("Computed value calculation not implemented - CSS 2.1 §6.1.2")
	// CSS 2.1 §6.1.2 Computed values
	// Relative values (em, %, etc.) should be converted to absolute values
	
	doc := dom.NewDocument()
	div := dom.NewElement("div")
	p := dom.NewElement("p")
	div.AppendChild(p)
	doc.AppendChild(div)
	
	stylesheet := &css.Stylesheet{
		Rules: []*css.Rule{
			{
				Selectors: []*css.Selector{
					{Simple: []*css.SimpleSelector{{TagName: "div"}}},
				},
				Declarations: []*css.Declaration{
					{Property: "font-size", Value: "16px"},
					{Property: "width", Value: "100%"},
				},
			},
			{
				Selectors: []*css.Selector{
					{Simple: []*css.SimpleSelector{{TagName: "p"}}},
				},
				Declarations: []*css.Declaration{
					{Property: "font-size", Value: "1.5em"}, // Should be 24px (16px * 1.5)
					{Property: "width", Value: "50%"},        // Should be 50% of parent
				},
			},
		},
	}
	
	styledTree := StyleTree(doc, stylesheet)
	divStyled := styledTree.Children[0]
	pStyled := divStyled.Children[0]
	
	// When implemented, would have ComputedStyles field
	// P's font-size should be computed to 24px (1.5em of parent's 16px)
	// For now, values are used as-is
	if pStyled.Styles["font-size"] != "1.5em" {
		t.Errorf("Expected font-size '1.5em' (not yet computed), got %v", pStyled.Styles["font-size"])
	}
}

func TestPseudoClassMatching_Skipped(t *testing.T) {
	t.Skip("Pseudo-class matching not implemented - CSS 2.1 §5.11")
	// CSS 2.1 §5.11 Pseudo-classes
	// Pseudo-classes should be matched based on element state/position
	
	doc := dom.NewDocument()
	div := dom.NewElement("div")
	p1 := dom.NewElement("p")
	p2 := dom.NewElement("p")
	div.AppendChild(p1)
	div.AppendChild(p2)
	doc.AppendChild(div)
	
	// Note: Parser would need to support pseudo-classes first
	// This test assumes SimpleSelector would have a PseudoClass field
	stylesheet := &css.Stylesheet{
		Rules: []*css.Rule{
			{
				Selectors: []*css.Selector{
					{Simple: []*css.SimpleSelector{{TagName: "p"}}}, // Would have PseudoClass: "first-child"
				},
				Declarations: []*css.Declaration{
					{Property: "margin-top", Value: "0"},
				},
			},
		},
	}
	
	styledTree := StyleTree(doc, stylesheet)
	divStyled := styledTree.Children[0]
	p1Styled := divStyled.Children[0]
	p2Styled := divStyled.Children[1]
	
	// When implemented, first p should match :first-child and have margin-top: 0
	// For now, both match the plain 'p' selector
	if p1Styled.Styles["margin-top"] != "0" {
		t.Errorf("Expected first p to have margin-top '0', got %v", p1Styled.Styles["margin-top"])
	}
	
	// Second p should also match (no pseudo-class filtering yet)
	if p2Styled.Styles["margin-top"] != "0" {
		t.Error("Without pseudo-class support, second p also matches 'p' selector")
	}
}

func TestChildCombinatorMatching_Skipped(t *testing.T) {
	t.Skip("Child combinator matching not implemented - CSS 2.1 §5.5")
	// CSS 2.1 §5.5 Child selectors
	// Child combinator (>) should only match direct children
	
	// Create DOM: div > p > span
	div := dom.NewElement("div")
	p := dom.NewElement("p")
	span := dom.NewElement("span")
	div.AppendChild(p)
	p.AppendChild(span)
	
	// When implemented, Selector would have a Combinator field
	// Selector: div > span (should NOT match - span is grandchild)
	selector := &css.Selector{
		Simple: []*css.SimpleSelector{
			{TagName: "div"},
			{TagName: "span"},
		},
		// Would have: Combinator: ">"
	}
	
	// Currently treats all multi-part selectors as descendant
	if !matchesSelector(span, selector) {
		t.Error("Currently matches as descendant (space), should NOT match with child combinator (>)")
	}
	
	// Selector: div > p (should match - p is direct child)
	selector2 := &css.Selector{
		Simple: []*css.SimpleSelector{
			{TagName: "div"},
			{TagName: "p"},
		},
		// Would have: Combinator: ">"
	}
	
	if !matchesSelector(p, selector2) {
		t.Error("Expected p to match 'div > p' (direct child)")
	}
}

// TestPresentationalHints tests HTML presentational attributes
func TestPresentationalHints(t *testing.T) {
tests := []struct {
name     string
node     *dom.Node
expected map[string]string
}{
{
name: "font color attribute",
node: func() *dom.Node {
n := dom.NewElement("font")
n.SetAttribute("color", "red")
return n
}(),
expected: map[string]string{"color": "red"},
},
{
name: "font color hex",
node: func() *dom.Node {
n := dom.NewElement("font")
n.SetAttribute("color", "#0000FF")
return n
}(),
expected: map[string]string{"color": "#0000FF"},
},
{
name: "bgcolor on table cell",
node: func() *dom.Node {
n := dom.NewElement("td")
n.SetAttribute("bgcolor", "yellow")
return n
}(),
expected: map[string]string{"background-color": "yellow"},
},
{
name: "bgcolor on table row",
node: func() *dom.Node {
n := dom.NewElement("tr")
n.SetAttribute("bgcolor", "#FF0000")
return n
}(),
expected: map[string]string{"background-color": "#FF0000"},
},
{
name: "no presentational attributes",
node: dom.NewElement("div"),
expected: map[string]string{},
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
styles := make(map[string]string)
applyPresentationalHints(tt.node, styles)

for key, expectedVal := range tt.expected {
if actualVal, ok := styles[key]; !ok {
t.Errorf("Expected style %s to be set", key)
} else if actualVal != expectedVal {
t.Errorf("Style %s = %v, expected %v", key, actualVal, expectedVal)
}
}

if len(styles) != len(tt.expected) {
t.Errorf("Got %d styles, expected %d", len(styles), len(tt.expected))
}
})
}
}

// TestInlineStyles tests that inline style attributes are applied correctly.
// CSS 2.1 §6.4.3: Inline styles have the highest specificity.
func TestInlineStyles(t *testing.T) {
	tests := []struct {
		name         string
		html         *dom.Node
		css          *css.Stylesheet
		expectedDiv  map[string]string
	}{
		{
			name: "inline style with no CSS rules",
			html: func() *dom.Node {
				doc := dom.NewDocument()
				div := dom.NewElement("div")
				div.SetAttribute("style", "color: red; font-size: 20px")
				doc.AppendChild(div)
				return doc
			}(),
			css: &css.Stylesheet{Rules: []*css.Rule{}},
			expectedDiv: map[string]string{
				"color":     "red",
				"font-size": "20px",
			},
		},
		{
			name: "inline style overrides CSS rule",
			html: func() *dom.Node {
				doc := dom.NewDocument()
				div := dom.NewElement("div")
				div.SetAttribute("style", "color: red")
				doc.AppendChild(div)
				return doc
			}(),
			css: &css.Stylesheet{
				Rules: []*css.Rule{
					{
						Selectors: []*css.Selector{
							{Simple: []*css.SimpleSelector{{TagName: "div"}}},
						},
						Declarations: []*css.Declaration{
							{Property: "color", Value: "blue"},
						},
					},
				},
			},
			expectedDiv: map[string]string{
				"color": "red", // Inline style wins
			},
		},
		{
			name: "inline style overrides ID selector",
			html: func() *dom.Node {
				doc := dom.NewDocument()
				div := dom.NewElement("div")
				div.SetAttribute("id", "main")
				div.SetAttribute("style", "color: red")
				doc.AppendChild(div)
				return doc
			}(),
			css: &css.Stylesheet{
				Rules: []*css.Rule{
					{
						Selectors: []*css.Selector{
							{Simple: []*css.SimpleSelector{{ID: "main"}}},
						},
						Declarations: []*css.Declaration{
							{Property: "color", Value: "blue"},
						},
					},
				},
			},
			expectedDiv: map[string]string{
				"color": "red", // Inline style wins over ID selector
			},
		},
		{
			name: "inline style combines with CSS rules",
			html: func() *dom.Node {
				doc := dom.NewDocument()
				div := dom.NewElement("div")
				div.SetAttribute("style", "color: red")
				doc.AppendChild(div)
				return doc
			}(),
			css: &css.Stylesheet{
				Rules: []*css.Rule{
					{
						Selectors: []*css.Selector{
							{Simple: []*css.SimpleSelector{{TagName: "div"}}},
						},
						Declarations: []*css.Declaration{
							{Property: "font-size", Value: "16px"},
						},
					},
				},
			},
			expectedDiv: map[string]string{
				"color":     "red",  // From inline style
				"font-size": "16px", // From CSS rule
			},
		},
		{
			name: "inline style with shorthand property",
			html: func() *dom.Node {
				doc := dom.NewDocument()
				div := dom.NewElement("div")
				div.SetAttribute("style", "margin: 10px 20px")
				doc.AppendChild(div)
				return doc
			}(),
			css: &css.Stylesheet{Rules: []*css.Rule{}},
			expectedDiv: map[string]string{
				"margin-top":    "10px",
				"margin-right":  "20px",
				"margin-bottom": "10px",
				"margin-left":   "20px",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			styledTree := StyleTree(tt.html, tt.css)
			
			// Get the div (first child of document)
			if len(styledTree.Children) == 0 {
				t.Fatal("Expected at least one child in styled tree")
			}
			divStyled := styledTree.Children[0]

			// Check expected styles
			for prop, expectedVal := range tt.expectedDiv {
				if actualVal, ok := divStyled.Styles[prop]; !ok {
					t.Errorf("Expected style %s to be set", prop)
				} else if actualVal != expectedVal {
					t.Errorf("Style %s = %v, expected %v", prop, actualVal, expectedVal)
				}
			}
		})
	}
}

// TestInlineStyleSpecificity tests that inline styles have higher specificity than any selector.
// CSS 2.1 §6.4.3: Inline style declarations have the highest specificity (A=1).
func TestInlineStyleSpecificity(t *testing.T) {
	doc := dom.NewDocument()
	div := dom.NewElement("div")
	div.SetAttribute("id", "unique")
	div.SetAttribute("class", "special highlight")
	div.SetAttribute("style", "color: red")
	doc.AppendChild(div)

	// Create stylesheet with rules of varying specificity
	stylesheet := &css.Stylesheet{
		Rules: []*css.Rule{
			{
				// Element selector (specificity: 0,0,0,1)
				Selectors: []*css.Selector{
					{Simple: []*css.SimpleSelector{{TagName: "div"}}},
				},
				Declarations: []*css.Declaration{
					{Property: "color", Value: "blue"},
				},
			},
			{
				// Class selector (specificity: 0,0,1,0)
				Selectors: []*css.Selector{
					{Simple: []*css.SimpleSelector{{Classes: []string{"special"}}}},
				},
				Declarations: []*css.Declaration{
					{Property: "color", Value: "green"},
				},
			},
			{
				// ID selector (specificity: 0,1,0,0)
				Selectors: []*css.Selector{
					{Simple: []*css.SimpleSelector{{ID: "unique"}}},
				},
				Declarations: []*css.Declaration{
					{Property: "color", Value: "yellow"},
				},
			},
			{
				// Combined selector (specificity: 0,1,2,1)
				Selectors: []*css.Selector{
					{Simple: []*css.SimpleSelector{{
						TagName: "div",
						ID:      "unique",
						Classes: []string{"special", "highlight"},
					}}},
				},
				Declarations: []*css.Declaration{
					{Property: "color", Value: "purple"},
				},
			},
		},
	}

	styledTree := StyleTree(doc, stylesheet)
	divStyled := styledTree.Children[0]

	// Inline style should win over all CSS rules
	if divStyled.Styles["color"] != "red" {
		t.Errorf("Expected inline style 'red' to win, got %v", divStyled.Styles["color"])
	}
}
