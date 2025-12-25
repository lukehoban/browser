// Package layout implements the CSS 2.1 visual formatting model.
// It converts styled nodes into a tree of layout boxes with computed dimensions.
//
// Spec references:
// - CSS 2.1 §8 Box model: https://www.w3.org/TR/CSS21/box.html
// - CSS 2.1 §9 Visual formatting model: https://www.w3.org/TR/CSS21/visuren.html
// - CSS 2.1 §10 Visual formatting model details: https://www.w3.org/TR/CSS21/visudet.html
package layout

import (
	"regexp"
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
			"width":               "200px",
			"height":              "auto",
			"padding-top":         "0",
			"padding-bottom":      "0",
			"border-top-width":    "0",
			"border-bottom-width": "0",
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
			"width":              "auto",
			"margin-left":        "10px",
			"margin-right":       "10px",
			"padding-left":       "0",
			"padding-right":      "0",
			"border-left-width":  "0",
			"border-right-width": "0",
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

func TestIntegrationLayoutHackerNews(t *testing.T) {
	// Test that we can parse and layout the Hacker News test file
	// This is a simplified static version of the Hacker News homepage
	htmlContent := `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Hacker News</title>
    <style>
        body {
            font-family: Verdana;
            margin: 0;
            padding: 0;
            background: #f6f6ef;
        }
        #hnmain {
            width: 85%;
            background: #f6f6ef;
        }
        #header {
            background: #ff6600;
            padding: 2px;
        }
        .rank {
            color: #828282;
        }
    </style>
</head>
<body>
    <center>
        <table id="hnmain">
            <tr>
                <td id="header">
                    <span class="pagetop">
                        <b class="hnname"><a href="news">Hacker News</a></b>
                    </span>
                </td>
            </tr>
            <tr>
                <td>
                    <table>
                        <tr class="athing" id="item1">
                            <td class="title">
                                <span class="rank">1.</span>
                            </td>
                            <td class="title">
                                <span class="titleline">
                                    <a href="https://example.com">Example Article</a>
                                </span>
                            </td>
                        </tr>
                    </table>
                </td>
            </tr>
        </table>
    </center>
</body>
</html>`

	// Parse the HTML
	doc := parseHTML(htmlContent)
	if doc == nil {
		t.Fatal("Failed to parse HTML")
	}

	// Extract CSS and create stylesheet
	cssContent := extractCSS(htmlContent)
	stylesheet := css.Parse(cssContent)
	if stylesheet == nil {
		t.Fatal("Failed to parse CSS")
	}

	// Compute styles
	styledTree := style.StyleTree(doc, stylesheet)
	if styledTree == nil {
		t.Fatal("Failed to compute styled tree")
	}

	// Create layout
	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 0},
	}
	layoutTree := LayoutTree(styledTree, containingBlock)

	// Verify layout was created
	if layoutTree == nil {
		t.Fatal("Expected layout tree to be created")
	}

	// Verify some basic structure
	if layoutTree.StyledNode == nil {
		t.Fatal("Layout tree should have a styled node")
	}

	// Verify width was calculated correctly for body
	if layoutTree.Dimensions.Content.Width != 800.0 {
		t.Errorf("Expected body width 800, got %v", layoutTree.Dimensions.Content.Width)
	}
}

// parseHTML is a helper that parses HTML using the html package
func parseHTML(input string) *dom.Node {
	tokenizer := newHTMLTokenizer(input)
	doc := dom.NewDocument()
	stack := []*dom.Node{doc}

	for {
		token, ok := tokenizer.Next()
		if !ok {
			break
		}

		switch token.Type {
		case startTagToken, selfClosingTagToken:
			elem := dom.NewElement(token.Data)
			for name, value := range token.Attributes {
				elem.SetAttribute(name, value)
			}
			current := stack[len(stack)-1]
			current.AppendChild(elem)
			if token.Type != selfClosingTagToken && !isVoidElement(token.Data) {
				stack = append(stack, elem)
			}
		case endTagToken:
			for i := len(stack) - 1; i >= 0; i-- {
				if stack[i].Type == dom.ElementNode && stack[i].Data == token.Data {
					stack = stack[:i]
					break
				}
			}
		case textToken:
			if len(stack) > 1 {
				text := dom.NewText(token.Data)
				current := stack[len(stack)-1]
				current.AppendChild(text)
			}
		}
	}

	return doc
}

// Token types for test helper
const (
	textToken = iota
	startTagToken
	endTagToken
	selfClosingTagToken
)

// tokenForTest is a simplified token for testing
type tokenForTest struct {
	Type       int
	Data       string
	Attributes map[string]string
}

// newHTMLTokenizer creates a simple tokenizer for testing
func newHTMLTokenizer(input string) *testTokenizer {
	return &testTokenizer{input: input, pos: 0}
}

type testTokenizer struct {
	input string
	pos   int
}

func (t *testTokenizer) Next() (tokenForTest, bool) {
	for t.pos < len(t.input) && (t.input[t.pos] == ' ' || t.input[t.pos] == '\n' || t.input[t.pos] == '\t' || t.input[t.pos] == '\r') {
		t.pos++
	}

	if t.pos >= len(t.input) {
		return tokenForTest{}, false
	}

	if t.input[t.pos] == '<' {
		return t.scanTag()
	}
	return t.scanText()
}

func (t *testTokenizer) scanText() (tokenForTest, bool) {
	start := t.pos
	for t.pos < len(t.input) && t.input[t.pos] != '<' {
		t.pos++
	}
	return tokenForTest{Type: textToken, Data: t.input[start:t.pos]}, true
}

func (t *testTokenizer) scanTag() (tokenForTest, bool) {
	t.pos++ // skip '<'

	// DOCTYPE
	if t.pos+8 <= len(t.input) && t.input[t.pos:t.pos+8] == "!DOCTYPE" {
		for t.pos < len(t.input) && t.input[t.pos] != '>' {
			t.pos++
		}
		t.pos++ // skip '>'
		return t.Next()
	}

	// Comment
	if t.pos+3 <= len(t.input) && t.input[t.pos:t.pos+3] == "!--" {
		for t.pos < len(t.input) {
			if t.pos+3 <= len(t.input) && t.input[t.pos:t.pos+3] == "-->" {
				t.pos += 3
				break
			}
			t.pos++
		}
		return t.Next()
	}

	// End tag
	if t.input[t.pos] == '/' {
		t.pos++
		start := t.pos
		for t.pos < len(t.input) && t.input[t.pos] != '>' {
			t.pos++
		}
		tagName := t.input[start:t.pos]
		t.pos++ // skip '>'
		return tokenForTest{Type: endTagToken, Data: tagName}, true
	}

	// Start tag
	start := t.pos
	for t.pos < len(t.input) && t.input[t.pos] != ' ' && t.input[t.pos] != '>' && t.input[t.pos] != '/' {
		t.pos++
	}
	tagName := t.input[start:t.pos]

	attrs := make(map[string]string)
	for t.pos < len(t.input) && t.input[t.pos] != '>' && t.input[t.pos] != '/' {
		for t.pos < len(t.input) && (t.input[t.pos] == ' ' || t.input[t.pos] == '\n' || t.input[t.pos] == '\t') {
			t.pos++
		}
		if t.pos >= len(t.input) || t.input[t.pos] == '>' || t.input[t.pos] == '/' {
			break
		}

		// Read attribute name
		attrStart := t.pos
		for t.pos < len(t.input) && t.input[t.pos] != '=' && t.input[t.pos] != ' ' && t.input[t.pos] != '>' && t.input[t.pos] != '/' {
			t.pos++
		}
		attrName := t.input[attrStart:t.pos]

		if t.pos < len(t.input) && t.input[t.pos] == '=' {
			t.pos++
			var attrValue string
			if t.pos < len(t.input) && (t.input[t.pos] == '"' || t.input[t.pos] == '\'') {
				quote := t.input[t.pos]
				t.pos++
				valueStart := t.pos
				for t.pos < len(t.input) && t.input[t.pos] != quote {
					t.pos++
				}
				attrValue = t.input[valueStart:t.pos]
				t.pos++ // skip closing quote
			}
			attrs[attrName] = attrValue
		}
	}

	selfClosing := false
	if t.pos < len(t.input) && t.input[t.pos] == '/' {
		selfClosing = true
		t.pos++
	}
	if t.pos < len(t.input) && t.input[t.pos] == '>' {
		t.pos++
	}

	tokenType := startTagToken
	if selfClosing {
		tokenType = selfClosingTagToken
	}
	return tokenForTest{Type: tokenType, Data: tagName, Attributes: attrs}, true
}

func isVoidElement(tagName string) bool {
	voidElements := map[string]bool{
		"area": true, "base": true, "br": true, "col": true, "embed": true,
		"hr": true, "img": true, "input": true, "link": true, "meta": true,
		"param": true, "source": true, "track": true, "wbr": true,
	}
	return voidElements[tagName]
}

// cssStyleRegexp is a pre-compiled regex for extracting CSS from style tags
var cssStyleRegexp = regexp.MustCompile(`(?is)<style[^>]*>(.*?)</style>`)

func extractCSS(htmlContent string) string {
	// Use regex for efficient CSS extraction from <style> tags
	matches := cssStyleRegexp.FindAllStringSubmatch(htmlContent, -1)
	result := ""
	for _, match := range matches {
		if len(match) > 1 {
			result += match[1]
		}
	}
	return result
}

func TestLayoutText(t *testing.T) {
	// Create a text node
	textNode := dom.NewText("Hello World")
	styledNode := &style.StyledNode{
		Node:     textNode,
		Styles:   map[string]string{},
		Children: []*style.StyledNode{},
	}

	// Build layout tree
	box := buildLayoutTree(styledNode)

	// Layout the text
	containingBlock := Dimensions{
		Content: Rect{
			X:      10,
			Y:      20,
			Width:  800,
			Height: 0,
		},
	}
	box.Layout(containingBlock)

	// Check that text dimensions were calculated
	if box.Dimensions.Content.Width == 0 {
		t.Errorf("expected non-zero width for text, got 0")
	}
	if box.Dimensions.Content.Height == 0 {
		t.Errorf("expected non-zero height for text, got 0")
	}

	// Check position
	if box.Dimensions.Content.X != 10 {
		t.Errorf("expected X position 10, got %v", box.Dimensions.Content.X)
	}
	if box.Dimensions.Content.Y != 20 {
		t.Errorf("expected Y position 20, got %v", box.Dimensions.Content.Y)
	}
}

func TestTableLayout(t *testing.T) {
	// Create a simple table: table > tr > td, td
	table := dom.NewElement("table")
tr := dom.NewElement("tr")
td1 := dom.NewElement("td")
td2 := dom.NewElement("td")
text1 := dom.NewText("Cell 1")
text2 := dom.NewText("Cell 2")

td1.AppendChild(text1)
td2.AppendChild(text2)
tr.AppendChild(td1)
tr.AppendChild(td2)
table.AppendChild(tr)

// Create styled nodes
styledTable := &style.StyledNode{
Node: table,
Styles: map[string]string{
"width": "400px",
},
Children: []*style.StyledNode{
{
Node:   tr,
Styles: map[string]string{},
Children: []*style.StyledNode{
{
Node: td1,
Styles: map[string]string{
"padding": "10px",
},
Children: []*style.StyledNode{
{
Node:     text1,
Styles:   map[string]string{},
Children: []*style.StyledNode{},
},
},
},
{
Node: td2,
Styles: map[string]string{
"padding": "10px",
},
Children: []*style.StyledNode{
{
Node:     text2,
Styles:   map[string]string{},
Children: []*style.StyledNode{},
},
},
},
},
},
},
}

// Build layout tree
containingBlock := Dimensions{
Content: Rect{X: 0, Y: 0, Width: 800, Height: 0},
}
layoutBox := buildLayoutTree(styledTable)

// Verify box types
if layoutBox.BoxType != TableBox {
t.Errorf("Expected TableBox, got %v", layoutBox.BoxType)
}
if len(layoutBox.Children) != 1 {
t.Fatalf("Expected 1 table row, got %d", len(layoutBox.Children))
}
if layoutBox.Children[0].BoxType != TableRowBox {
t.Errorf("Expected TableRowBox, got %v", layoutBox.Children[0].BoxType)
}
if len(layoutBox.Children[0].Children) != 2 {
t.Fatalf("Expected 2 table cells, got %d", len(layoutBox.Children[0].Children))
}
if layoutBox.Children[0].Children[0].BoxType != TableCellBox {
t.Errorf("Expected TableCellBox, got %v", layoutBox.Children[0].Children[0].BoxType)
}

// Layout the table
layoutBox.Layout(containingBlock)

// Verify table width
if layoutBox.Dimensions.Content.Width != 400.0 {
t.Errorf("Expected table width 400, got %v", layoutBox.Dimensions.Content.Width)
}

// Verify cells are laid out horizontally
row := layoutBox.Children[0]
cell1 := row.Children[0]
cell2 := row.Children[1]

// Each cell should be approximately 200px wide (400 / 2)
expectedCellWidth := 200.0
if cell1.Dimensions.Content.Width < expectedCellWidth-20 || cell1.Dimensions.Content.Width > expectedCellWidth+20 {
t.Errorf("Expected cell1 width around %v, got %v", expectedCellWidth, cell1.Dimensions.Content.Width)
}

// Cell 2 should be positioned to the right of cell 1
if cell2.Dimensions.Content.X <= cell1.Dimensions.Content.X {
t.Errorf("Cell 2 should be positioned to the right of cell 1. Cell1 X=%v, Cell2 X=%v",
cell1.Dimensions.Content.X, cell2.Dimensions.Content.X)
}

// Cells should be on the same row (same Y position approximately)
if cell1.Dimensions.Content.Y != cell2.Dimensions.Content.Y {
t.Errorf("Cells should be on same row. Cell1 Y=%v, Cell2 Y=%v",
cell1.Dimensions.Content.Y, cell2.Dimensions.Content.Y)
}
}

// SKIPPED TESTS FOR KNOWN BROKEN/UNIMPLEMENTED FEATURES
// These tests document known limitations that need to be implemented.
// See MILESTONES.md for more details.

func TestInlineLayout(t *testing.T) {
	// CSS 2.1 §9.4.2 Inline formatting contexts
	// Inline elements should flow horizontally within line boxes
	
	// Create DOM: p with inline spans
	p := dom.NewElement("p")
	span1 := dom.NewElement("span")
	text1 := dom.NewText("Hello ")
	span1.AppendChild(text1)
	span2 := dom.NewElement("span")
	text2 := dom.NewText("World")
	span2.AppendChild(text2)
	p.AppendChild(span1)
	p.AppendChild(span2)
	
	styledP := &style.StyledNode{
		Node: p,
		Styles: map[string]string{
			"display": "block",
			"width":   "200px",
		},
		Children: []*style.StyledNode{
			{
				Node: span1,
				Styles: map[string]string{
					"display": "inline",
				},
				Children: []*style.StyledNode{
					{Node: text1, Styles: map[string]string{}},
				},
			},
			{
				Node: span2,
				Styles: map[string]string{
					"display": "inline",
				},
				Children: []*style.StyledNode{
					{Node: text2, Styles: map[string]string{}},
				},
			},
		},
	}
	
	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 0},
	}
	
	box := buildLayoutTree(styledP)
	box.Layout(containingBlock)
	
	// Spans should be laid out inline (horizontally)
	if len(box.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(box.Children))
	}
	
	span1Box := box.Children[0]
	span2Box := box.Children[1]
	
	// Span2 should be to the right of span1, not below it
	if span2Box.Dimensions.Content.X <= span1Box.Dimensions.Content.X {
		t.Errorf("Span2 should be to the right of span1")
	}
	
	// Both spans should be on the same line (same Y)
	if span1Box.Dimensions.Content.Y != span2Box.Dimensions.Content.Y {
		t.Errorf("Spans should be on the same line")
	}
}

func TestAbsolutePositioning_Skipped(t *testing.T) {
	t.Skip("Absolute positioning not implemented - CSS 2.1 §9.6")
	// CSS 2.1 §9.6 Absolute positioning
	// Elements with position: absolute should be positioned relative to containing block
	
	parent := dom.NewElement("div")
	child := dom.NewElement("div")
	parent.AppendChild(child)
	
	styledParent := &style.StyledNode{
		Node: parent,
		Styles: map[string]string{
			"position": "relative",
			"width":    "400px",
			"height":   "300px",
		},
		Children: []*style.StyledNode{
			{
				Node: child,
				Styles: map[string]string{
					"position": "absolute",
					"top":      "20px",
					"left":     "30px",
					"width":    "100px",
					"height":   "50px",
				},
			},
		},
	}
	
	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 600},
	}
	
	box := buildLayoutTree(styledParent)
	box.Layout(containingBlock)
	
	if len(box.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(box.Children))
	}
	
	childBox := box.Children[0]
	
	// Child should be positioned at (30, 20) relative to parent
	if childBox.Dimensions.Content.X != 30.0 {
		t.Errorf("Expected X position 30, got %v", childBox.Dimensions.Content.X)
	}
	if childBox.Dimensions.Content.Y != 20.0 {
		t.Errorf("Expected Y position 20, got %v", childBox.Dimensions.Content.Y)
	}
	
	// Child should not affect parent's height (out of flow)
	if box.Dimensions.Content.Height > 300.0 {
		t.Errorf("Absolute positioned child should not affect parent height")
	}
}

func TestRelativePositioning_Skipped(t *testing.T) {
	t.Skip("Relative positioning not implemented - CSS 2.1 §9.4.3")
	// CSS 2.1 §9.4.3 Relative positioning
	// Elements with position: relative should be offset from their normal position
	
	parent := dom.NewElement("div")
	child := dom.NewElement("div")
	parent.AppendChild(child)
	
	styledParent := &style.StyledNode{
		Node: parent,
		Styles: map[string]string{
			"width":  "400px",
			"height": "300px",
		},
		Children: []*style.StyledNode{
			{
				Node: child,
				Styles: map[string]string{
					"position": "relative",
					"top":      "10px",
					"left":     "20px",
					"width":    "100px",
					"height":   "50px",
				},
			},
		},
	}
	
	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 600},
	}
	
	box := buildLayoutTree(styledParent)
	box.Layout(containingBlock)
	
	childBox := box.Children[0]
	
	// Child should be offset by (20, 10) from its normal position
	// Normal position would be (0, 0) within parent
	if childBox.Dimensions.Content.X != 20.0 {
		t.Errorf("Expected X position 20 (0 + 20 offset), got %v", childBox.Dimensions.Content.X)
	}
	if childBox.Dimensions.Content.Y != 10.0 {
		t.Errorf("Expected Y position 10 (0 + 10 offset), got %v", childBox.Dimensions.Content.Y)
	}
}

func TestFixedPositioning_Skipped(t *testing.T) {
	t.Skip("Fixed positioning not implemented - CSS 2.1 §9.6")
	// CSS 2.1 §9.6.1 Fixed positioning
	// Elements with position: fixed should be positioned relative to viewport
	
	div := dom.NewElement("div")
	styledDiv := &style.StyledNode{
		Node: div,
		Styles: map[string]string{
			"position": "fixed",
			"top":      "10px",
			"right":    "20px",
			"width":    "100px",
			"height":   "50px",
		},
	}
	
	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 600},
	}
	
	box := buildLayoutTree(styledDiv)
	box.Layout(containingBlock)
	
	// Should be positioned at (680, 10) - 20px from right edge
	expectedX := 680.0 // 800 - 20 - 100
	if box.Dimensions.Content.X != expectedX {
		t.Errorf("Expected X position %v, got %v", expectedX, box.Dimensions.Content.X)
	}
	if box.Dimensions.Content.Y != 10.0 {
		t.Errorf("Expected Y position 10, got %v", box.Dimensions.Content.Y)
	}
}

func TestFloatLayout_Skipped(t *testing.T) {
	t.Skip("Float layout not implemented - CSS 2.1 §9.5")
	// CSS 2.1 §9.5 Floats
	// Floated elements should be moved to the left or right edge
	
	container := dom.NewElement("div")
	float1 := dom.NewElement("div")
	float2 := dom.NewElement("div")
	content := dom.NewElement("p")
	container.AppendChild(float1)
	container.AppendChild(float2)
	container.AppendChild(content)
	
	styledContainer := &style.StyledNode{
		Node: container,
		Styles: map[string]string{
			"width": "400px",
		},
		Children: []*style.StyledNode{
			{
				Node: float1,
				Styles: map[string]string{
					"float":  "left",
					"width":  "100px",
					"height": "50px",
				},
			},
			{
				Node: float2,
				Styles: map[string]string{
					"float":  "left",
					"width":  "100px",
					"height": "50px",
				},
			},
			{
				Node: content,
				Styles: map[string]string{
					"width": "auto",
				},
			},
		},
	}
	
	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 0},
	}
	
	box := buildLayoutTree(styledContainer)
	box.Layout(containingBlock)
	
	float1Box := box.Children[0]
	float2Box := box.Children[1]
	contentBox := box.Children[2]
	
	// Float1 should be at left edge
	if float1Box.Dimensions.Content.X != 0.0 {
		t.Errorf("Expected float1 X at 0, got %v", float1Box.Dimensions.Content.X)
	}
	
	// Float2 should be to the right of float1
	expectedFloat2X := 100.0
	if float2Box.Dimensions.Content.X != expectedFloat2X {
		t.Errorf("Expected float2 X at %v, got %v", expectedFloat2X, float2Box.Dimensions.Content.X)
	}
	
	// Content should flow around the floats
	expectedContentX := 200.0 // After both floats
	if contentBox.Dimensions.Content.X != expectedContentX {
		t.Errorf("Expected content X at %v, got %v", expectedContentX, contentBox.Dimensions.Content.X)
	}
	
	// Content width should be reduced by float widths
	expectedContentWidth := 200.0 // 400 - 100 - 100
	if contentBox.Dimensions.Content.Width != expectedContentWidth {
		t.Errorf("Expected content width %v, got %v", expectedContentWidth, contentBox.Dimensions.Content.Width)
	}
}

func TestFlexboxLayout_Skipped(t *testing.T) {
	t.Skip("Flexbox not implemented - CSS Flexible Box Layout Module Level 1")
	// CSS Flexible Box Layout Module Level 1
	// Flexbox provides efficient layout for complex alignments
	
	container := dom.NewElement("div")
	item1 := dom.NewElement("div")
	item2 := dom.NewElement("div")
	item3 := dom.NewElement("div")
	container.AppendChild(item1)
	container.AppendChild(item2)
	container.AppendChild(item3)
	
	styledContainer := &style.StyledNode{
		Node: container,
		Styles: map[string]string{
			"display":         "flex",
			"flex-direction":  "row",
			"justify-content": "space-between",
			"width":           "400px",
		},
		Children: []*style.StyledNode{
			{Node: item1, Styles: map[string]string{"width": "100px", "height": "50px"}},
			{Node: item2, Styles: map[string]string{"width": "100px", "height": "50px"}},
			{Node: item3, Styles: map[string]string{"width": "100px", "height": "50px"}},
		},
	}
	
	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 0},
	}
	
	box := buildLayoutTree(styledContainer)
	box.Layout(containingBlock)
	
	// Items should be distributed with space between them
	item1Box := box.Children[0]
	item2Box := box.Children[1]
	item3Box := box.Children[2]
	
	// Item1 at start
	if item1Box.Dimensions.Content.X != 0.0 {
		t.Errorf("Expected item1 X at 0, got %v", item1Box.Dimensions.Content.X)
	}
	
	// Item2 in middle
	expectedItem2X := 150.0 // (400 - 300) / 2 + 100
	if item2Box.Dimensions.Content.X != expectedItem2X {
		t.Errorf("Expected item2 X at %v, got %v", expectedItem2X, item2Box.Dimensions.Content.X)
	}
	
	// Item3 at end
	expectedItem3X := 300.0
	if item3Box.Dimensions.Content.X != expectedItem3X {
		t.Errorf("Expected item3 X at %v, got %v", expectedItem3X, item3Box.Dimensions.Content.X)
	}
}

func TestGridLayout_Skipped(t *testing.T) {
	t.Skip("Grid layout not implemented - CSS Grid Layout Module Level 1")
	// CSS Grid Layout Module Level 1
	// Grid provides two-dimensional layout system
	
	container := dom.NewElement("div")
	for i := 0; i < 6; i++ {
		item := dom.NewElement("div")
		container.AppendChild(item)
	}
	
	children := make([]*style.StyledNode, 6)
	for i := range children {
		children[i] = &style.StyledNode{
			Node:   container.Children[i],
			Styles: map[string]string{},
		}
	}
	
	styledContainer := &style.StyledNode{
		Node: container,
		Styles: map[string]string{
			"display":               "grid",
			"grid-template-columns": "1fr 1fr 1fr",
			"grid-gap":              "10px",
			"width":                 "400px",
		},
		Children: children,
	}
	
	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 0},
	}
	
	box := buildLayoutTree(styledContainer)
	box.Layout(containingBlock)
	
	// Items should be arranged in 3 columns, 2 rows
	// Each column should be approximately 130px wide (400 - 20 gap) / 3
	
	// First row items
	item0 := box.Children[0]
	item1 := box.Children[1]
	item2 := box.Children[2]
	
	// All first row items should be on same Y
	if item0.Dimensions.Content.Y != item1.Dimensions.Content.Y ||
		item1.Dimensions.Content.Y != item2.Dimensions.Content.Y {
		t.Error("First row items should be on same Y position")
	}
	
	// Items should be spaced horizontally with gaps
	expectedGap := 10.0
	gap1 := item1.Dimensions.Content.X - (item0.Dimensions.Content.X + item0.Dimensions.Content.Width)
	if gap1 < expectedGap-1 || gap1 > expectedGap+1 {
		t.Errorf("Expected gap of %v between items 0 and 1, got %v", expectedGap, gap1)
	}
}
