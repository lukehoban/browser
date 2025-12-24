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
	if divStyled.Styles["margin"] != "10px" {
		t.Errorf("Expected margin '10px', got %v", divStyled.Styles["margin"])
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

func TestStyleInheritance(t *testing.T) {
	// Create DOM: div > p > text
	// CSS 2.1 §6.2: Certain properties (font properties, color, etc.) are inherited
	div := dom.NewElement("div")
	p := dom.NewElement("p")
	text := dom.NewText("Hello")
	div.AppendChild(p)
	p.AppendChild(text)

	// Create stylesheet with font properties on div
	stylesheet := &css.Stylesheet{
		Rules: []*css.Rule{
			{
				Selectors: []*css.Selector{
					{Simple: []*css.SimpleSelector{{TagName: "div"}}},
				},
				Declarations: []*css.Declaration{
					{Property: "font-size", Value: "20px"},
					{Property: "color", Value: "red"},
					{Property: "font-weight", Value: "bold"},
				},
			},
		},
	}

	// Style the tree
	doc := dom.NewDocument()
	doc.AppendChild(div)
	styledTree := StyleTree(doc, stylesheet)

	// Check that div has the styles
	divStyled := styledTree.Children[0]
	if divStyled.Styles["font-size"] != "20px" {
		t.Errorf("Expected div font-size '20px', got %v", divStyled.Styles["font-size"])
	}
	if divStyled.Styles["color"] != "red" {
		t.Errorf("Expected div color 'red', got %v", divStyled.Styles["color"])
	}

	// Check that p inherits the styles
	pStyled := divStyled.Children[0]
	if pStyled.Styles["font-size"] != "20px" {
		t.Errorf("Expected p to inherit font-size '20px', got %v", pStyled.Styles["font-size"])
	}
	if pStyled.Styles["color"] != "red" {
		t.Errorf("Expected p to inherit color 'red', got %v", pStyled.Styles["color"])
	}

	// Check that text node inherits the styles
	textStyled := pStyled.Children[0]
	if textStyled.Styles["font-size"] != "20px" {
		t.Errorf("Expected text to inherit font-size '20px', got %v", textStyled.Styles["font-size"])
	}
	if textStyled.Styles["color"] != "red" {
		t.Errorf("Expected text to inherit color 'red', got %v", textStyled.Styles["color"])
	}
	if textStyled.Styles["font-weight"] != "bold" {
		t.Errorf("Expected text to inherit font-weight 'bold', got %v", textStyled.Styles["font-weight"])
	}
}
