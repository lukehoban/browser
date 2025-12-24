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
