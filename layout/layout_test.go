package layout

import (
	"testing"

	"github.com/lukehoban/browser/css"
	"github.com/lukehoban/browser/dom"
	"github.com/lukehoban/browser/style"
)

func TestParseLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		reference float64
		expected  float64
	}{
		{"pixels", "10px", 0, 10.0},
		{"percentage", "50%", 100, 50.0},
		{"auto", "auto", 0, -1.0},
		{"empty", "", 0, -1.0},
		{"number", "10", 0, 10.0},
		{"percentage calculation", "25%", 200, 50.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLength(tt.value, tt.reference)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLayoutSimpleBlock(t *testing.T) {
	// Create a simple styled node
	node := dom.NewElement("div")
	styledNode := &style.StyledNode{
		Node: node,
		Styles: map[string]string{
			"width":          "100px",
			"height":         "50px",
			"margin-top":     "10px",
			"margin-bottom":  "10px",
			"margin-left":    "10px",
			"margin-right":   "10px",
			"padding-top":    "5px",
			"padding-bottom": "5px",
			"padding-left":   "5px",
			"padding-right":  "5px",
		},
		Children: []*style.StyledNode{},
	}

	// Create layout
	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 0},
	}
	
	box := buildLayoutTree(styledNode)
	box.Layout(containingBlock)

	// Check dimensions
	if box.Dimensions.Content.Width != 100.0 {
		t.Errorf("Expected width 100, got %v", box.Dimensions.Content.Width)
	}
	if box.Dimensions.Content.Height != 50.0 {
		t.Errorf("Expected height 50, got %v", box.Dimensions.Content.Height)
	}
}

func TestLayoutNestedBlocks(t *testing.T) {
	// Create nested structure: div > div
	parent := dom.NewElement("div")
	child := dom.NewElement("div")
	parent.AppendChild(child)

	styledParent := &style.StyledNode{
		Node: parent,
		Styles: map[string]string{
			"width":                "200px",
			"height":               "auto",
			"padding-top":          "0",
			"padding-bottom":       "0",
			"border-top-width":     "0",
			"border-bottom-width":  "0",
		},
		Children: []*style.StyledNode{
			{
				Node: child,
				Styles: map[string]string{
					"width":          "100px",
					"height":         "50px",
					"margin-top":     "0",
					"margin-bottom":  "0",
					"margin-left":    "0",
					"margin-right":   "0",
					"padding-top":    "0",
					"padding-bottom": "0",
					"padding-left":   "0",
					"padding-right":  "0",
				},
				Children: []*style.StyledNode{},
			},
		},
	}

	// Create layout
	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 0},
	}
	
	box := buildLayoutTree(styledParent)
	box.Layout(containingBlock)

	// Check parent height includes child
	if box.Dimensions.Content.Height != 50.0 {
		t.Errorf("Expected parent height 50 (from child), got %v", box.Dimensions.Content.Height)
	}

	// Check child dimensions
	if len(box.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(box.Children))
	}
	
	childBox := box.Children[0]
	if childBox.Dimensions.Content.Width != 100.0 {
		t.Errorf("Expected child width 100, got %v", childBox.Dimensions.Content.Width)
	}
}

func TestLayoutWithMarginPadding(t *testing.T) {
	node := dom.NewElement("div")
	styledNode := &style.StyledNode{
		Node: node,
		Styles: map[string]string{
			"width":         "100px",
			"margin-left":   "10px",
			"margin-right":  "10px",
			"padding-left":  "5px",
			"padding-right": "5px",
		},
		Children: []*style.StyledNode{},
	}

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 0},
	}
	
	box := buildLayoutTree(styledNode)
	box.Layout(containingBlock)

	if box.Dimensions.Content.Width != 100.0 {
		t.Errorf("Expected content width 100, got %v", box.Dimensions.Content.Width)
	}
	if box.Dimensions.Margin.Left != 10.0 {
		t.Errorf("Expected margin-left 10, got %v", box.Dimensions.Margin.Left)
	}
	if box.Dimensions.Margin.Right != 10.0 {
		t.Errorf("Expected margin-right 10, got %v", box.Dimensions.Margin.Right)
	}
	if box.Dimensions.Padding.Left != 5.0 {
		t.Errorf("Expected padding-left 5, got %v", box.Dimensions.Padding.Left)
	}
	if box.Dimensions.Padding.Right != 5.0 {
		t.Errorf("Expected padding-right 5, got %v", box.Dimensions.Padding.Right)
	}
}

func TestLayoutAutoWidth(t *testing.T) {
	node := dom.NewElement("div")
	styledNode := &style.StyledNode{
		Node: node,
		Styles: map[string]string{
			"width":                "auto",
			"margin-left":          "10px",
			"margin-right":         "10px",
			"padding-left":         "0",
			"padding-right":        "0",
			"border-left-width":    "0",
			"border-right-width":   "0",
		},
		Children: []*style.StyledNode{},
	}

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 0},
	}
	
	box := buildLayoutTree(styledNode)
	box.Layout(containingBlock)

	// Width should fill container minus margins
	expected := 800.0 - 10.0 - 10.0
	if box.Dimensions.Content.Width != expected {
		t.Errorf("Expected auto width %v, got %v", expected, box.Dimensions.Content.Width)
	}
}

func TestBoxModel(t *testing.T) {
	box := &LayoutBox{
		Dimensions: Dimensions{
			Content: Rect{X: 10, Y: 10, Width: 100, Height: 50},
			Padding: EdgeSizes{Top: 5, Right: 5, Bottom: 5, Left: 5},
			Border:  EdgeSizes{Top: 2, Right: 2, Bottom: 2, Left: 2},
			Margin:  EdgeSizes{Top: 10, Right: 10, Bottom: 10, Left: 10},
		},
	}

	// Check padding box
	paddingBox := box.paddingBox()
	if paddingBox.Width != 110.0 { // 100 + 5 + 5
		t.Errorf("Expected padding box width 110, got %v", paddingBox.Width)
	}

	// Check border box
	borderBox := box.borderBox()
	if borderBox.Width != 114.0 { // 100 + 5 + 5 + 2 + 2
		t.Errorf("Expected border box width 114, got %v", borderBox.Width)
	}

	// Check margin box
	marginBox := box.marginBox()
	if marginBox.Width != 134.0 { // 100 + 5 + 5 + 2 + 2 + 10 + 10
		t.Errorf("Expected margin box width 134, got %v", marginBox.Width)
	}
}

func TestIntegrationLayoutFromHTML(t *testing.T) {
	// Create a simple HTML structure with styles
	doc := dom.NewDocument()
	body := dom.NewElement("body")
	div := dom.NewElement("div")
	div.SetAttribute("id", "main")
	body.AppendChild(div)
	doc.AppendChild(body)

	// Create stylesheet
	stylesheet := &css.Stylesheet{
		Rules: []*css.Rule{
			{
				Selectors: []*css.Selector{
					{Simple: []*css.SimpleSelector{{TagName: "body"}}},
				},
				Declarations: []*css.Declaration{
					{Property: "margin-top", Value: "0"},
					{Property: "margin-bottom", Value: "0"},
					{Property: "margin-left", Value: "0"},
					{Property: "margin-right", Value: "0"},
				},
			},
			{
				Selectors: []*css.Selector{
					{Simple: []*css.SimpleSelector{{ID: "main"}}},
				},
				Declarations: []*css.Declaration{
					{Property: "width", Value: "400px"},
					{Property: "height", Value: "200px"},
				},
			},
		},
	}

	// Compute styles
	styledTree := style.StyleTree(doc, stylesheet)

	// Create layout
	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 0},
	}
	layoutTree := LayoutTree(styledTree, containingBlock)

	// Verify layout was created
	if layoutTree == nil {
		t.Fatal("Expected layout tree to be created")
	}
}
