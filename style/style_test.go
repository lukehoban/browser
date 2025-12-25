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
