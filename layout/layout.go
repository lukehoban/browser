// Package layout implements the CSS 2.1 visual formatting model.
// It converts styled nodes into a tree of layout boxes with computed dimensions.
//
// Spec references:
// - CSS 2.1 §8 Box model: https://www.w3.org/TR/CSS21/box.html
// - CSS 2.1 §9 Visual formatting model: https://www.w3.org/TR/CSS21/visuren.html
// - CSS 2.1 §10 Visual formatting model details: https://www.w3.org/TR/CSS21/visudet.html
// - CSS 2.1 §17 Tables: https://www.w3.org/TR/CSS21/tables.html
// - CSS Flexible Box Layout Module Level 1: https://www.w3.org/TR/css-flexbox-1/
//
// Implemented features:
// - Box model (content, padding, border, margin) per CSS 2.1 §8
// - Block-level layout in normal flow (CSS 2.1 §9.4.1)
// - Inline formatting context with baseline alignment (CSS 2.1 §9.4.2, §10.8)
// - Table layout with auto width algorithm (CSS 2.1 §17.5)
// - Width calculation per CSS 2.1 §10.3.3
// - Height calculation per CSS 2.1 §10.6.3
// - Text alignment (left, center, right) via CSS text-align and HTML align attribute
// - Vertical alignment in table cells via HTML valign attribute
// - Flexbox layout (CSS3): display:flex, flex-direction:row, justify-content
//
// Not yet implemented (would log warnings if encountered):
// - Floats (CSS 2.1 §9.5)
// - Positioning schemes: absolute, relative, fixed (CSS 2.1 §9.3)
// - Inline layout with line wrapping (CSS 2.1 §9.4.2 - partial)
// - Z-index and stacking contexts (CSS 2.1 §9.9)
// - Table rowspan (CSS 2.1 §17.2)
// - Border-collapse model (CSS 2.1 §17.6.2)
// - Advanced flexbox features (column direction, wrap, align-items, flex-grow/shrink)
// - Grid layout (CSS3)
package layout

import (
	"math"
	"strconv"
	"strings"

	"github.com/lukehoban/browser/css"
	"github.com/lukehoban/browser/dom"
	"github.com/lukehoban/browser/font"
	"github.com/lukehoban/browser/log"
	"github.com/lukehoban/browser/style"
	"golang.org/x/image/font/basicfont"
)

// Layout constants
const (
	// CSS 2.1 §17.5.2.2: Maximum table column width to prevent unusable layouts
	maxColumnWidth = 400.0

	// HTML5 §4.9.9: Maximum colspan to prevent DoS attacks
	maxColspan = 1000

	// CSS 2.1 §16.4: Default word spacing as a fraction of font size (em units).
	// The spec states word-spacing 'normal' uses the font's default inter-word space,
	// which is typically around 0.25em (approximately the width of a space character).
	// Reference: https://www.w3.org/TR/CSS2/text.html#propdef-word-spacing
	defaultWordSpacingEm = 0.25

	// CSS 2.1 §10.8.1: Baseline position as a fraction of font size.
	// For most Latin fonts, the baseline is approximately 80% from the top of the em-box,
	// accounting for ascenders (the baseline sits below the cap height).
	baselinePositionEm = 0.8
)

// LayoutBox represents a box in the layout tree.
// CSS 2.1 §8.1 Box dimensions
type LayoutBox struct {
	BoxType    BoxType
	StyledNode *style.StyledNode
	Dimensions Dimensions
	Children   []*LayoutBox
}

// BoxType represents the type of a layout box.
// CSS 2.1 §9.2.1 Block-level elements and block boxes
// CSS 2.1 §17 Tables
type BoxType int

const (
	// BlockBox represents a block-level box
	BlockBox BoxType = iota
	// InlineBox represents an inline-level box
	InlineBox
	// AnonymousBox represents an anonymous block box
	AnonymousBox
	// TableBox represents a table box
	TableBox
	// TableRowBox represents a table row box
	TableRowBox
	// TableCellBox represents a table cell box
	TableCellBox
	// FlexBox represents a flex container box (CSS3 Flexbox)
	FlexBox
)

// Dimensions represents the dimensions of a box.
// CSS 2.1 §8.1 Box dimensions
type Dimensions struct {
	// Content area
	Content Rect

	// Padding edge
	Padding EdgeSizes

	// Border edge
	Border EdgeSizes

	// Margin edge
	Margin EdgeSizes
}

// Rect represents a rectangle.
type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// EdgeSizes represents the sizes of the four edges of a box.
type EdgeSizes struct {
	Top    float64
	Right  float64
	Bottom float64
	Left   float64
}

// LayoutTree builds a layout tree from a styled tree.
// CSS 2.1 §9.2 Controlling box generation
func LayoutTree(styledNode *style.StyledNode, containingBlock Dimensions) *LayoutBox {
	// Set initial containing block dimensions
	containingBlock.Content.Width = 800.0 // Default viewport width

	root := buildLayoutTree(styledNode)
	root.Layout(containingBlock)
	return root
}

// buildLayoutTree constructs the layout tree.
func buildLayoutTree(styledNode *style.StyledNode) *LayoutBox {
	// Skip whitespace-only text nodes
	// CSS 2.1 §16.6.1: Whitespace-only text nodes are collapsed.
	// However, they may contribute to word spacing in inline contexts.
	if styledNode.Node != nil && styledNode.Node.Type == dom.TextNode {
		text := styledNode.Node.Data
		isWhitespaceOnly := true
		for _, ch := range text {
			if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
				isWhitespaceOnly = false
				break
			}
		}
		if isWhitespaceOnly {
			log.Debug("Skipping whitespace-only text node in layout")
			return nil // Don't create a box for whitespace-only text
		}
	}

	// Determine box type based on display property or HTML element
	// CSS 2.1 §17.2.1: The table element generates a principal table box
	// CSS 2.1 §9.2.2: Inline-level elements
	boxType := BlockBox
	display := styledNode.Styles["display"]
	
	// CSS 2.1 §9.3: Positioning schemes (absolute, relative, fixed) - not yet implemented
	if position := styledNode.Styles["position"]; position != "" && position != "static" {
		log.Warnf("CSS 2.1 §9.3: position:%s not yet implemented (only 'static' positioning supported)", position)
	}
	
	// CSS 2.1 §9.5: Floats - not yet implemented
	if float := styledNode.Styles["float"]; float != "" && float != "none" {
		log.Warnf("CSS 2.1 §9.5: float:%s not yet implemented", float)
	}
	
	// CSS3: Flexbox - basic support for display:flex with flex-direction:row
	if display == "flex" {
		// Supported - will create FlexBox
	} else if display == "inline-flex" {
		log.Warnf("CSS3 Flexbox: display:%s not yet implemented, treating as flex", display)
		display = "flex"
	}
	
	// CSS3: Grid - not yet implemented
	if display == "grid" || display == "inline-grid" {
		log.Warnf("CSS3 Grid: display:%s not yet implemented, treating as block", display)
		display = "block"
	}

	// If no explicit display property, infer from HTML element
	if display == "" && styledNode.Node != nil && styledNode.Node.Type == dom.ElementNode {
		switch styledNode.Node.Data {
		case "table":
			display = "table"
		case "tr":
			display = "table-row"
		case "td", "th":
			display = "table-cell"
		// CSS 2.1 §9.2.2: Inline-level elements and inline boxes
		// These elements generate inline boxes by default
		case "a", "span", "b", "strong", "i", "em", "font", "code", "small", "big",
			"abbr", "cite", "kbd", "samp", "var", "sub", "sup", "mark", "u", "s", "del", "ins":
			display = "inline"
		// HTML5 §10.3.1: Elements that should not be rendered
		// These elements have display:none in the default UA stylesheet
		case "head", "title", "meta", "link", "style", "script", "noscript", "base":
			display = "none"
		}
	}

	switch display {
	case "inline":
		boxType = InlineBox
	case "none":
		if styledNode.Node != nil && styledNode.Node.Type == dom.ElementNode {
			log.Debugf("Element <%s> has display:none, skipping layout", styledNode.Node.Data)
		}
		return nil // Don't create a box
	case "table":
		boxType = TableBox
	case "table-row":
		boxType = TableRowBox
	case "table-cell":
		boxType = TableCellBox
	case "flex":
		boxType = FlexBox
	}

	box := &LayoutBox{
		BoxType:    boxType,
		StyledNode: styledNode,
		Dimensions: Dimensions{},
		Children:   make([]*LayoutBox, 0),
	}

	// Build children
	for _, child := range styledNode.Children {
		if childBox := buildLayoutTree(child); childBox != nil {
			box.Children = append(box.Children, childBox)
		}
	}

	return box
}

// Layout calculates the layout for this box and its children.
// CSS 2.1 §10 Visual formatting model details
// CSS 2.1 §17 Tables
// CSS Flexible Box Layout Module Level 1 (Flexbox)
func (box *LayoutBox) Layout(containingBlock Dimensions) {
	switch box.BoxType {
	case BlockBox:
		box.layoutBlock(containingBlock)
	case InlineBox:
		box.layoutInlineBox(containingBlock)
	case TableBox:
		box.layoutTable(containingBlock)
	case TableRowBox:
		box.layoutTableRow(containingBlock)
	case TableCellBox:
		box.layoutTableCell(containingBlock)
	case FlexBox:
		box.layoutFlex(containingBlock)
	}
}

// layoutBlock lays out a block-level element.
// CSS 2.1 §10.3 Calculating widths and margins
// CSS 2.1 §10.6 Calculating heights and margins
func (box *LayoutBox) layoutBlock(containingBlock Dimensions) {
	// Check if this is a text node
	if box.StyledNode != nil && box.StyledNode.Node != nil && box.StyledNode.Node.Type == dom.TextNode {
		box.layoutText(containingBlock)
		return
	}

	// Calculate width
	box.calculateBlockWidth(containingBlock)

	// Calculate position
	box.calculateBlockPosition(containingBlock)

	// Layout children
	box.layoutBlockChildren()

	// Calculate height
	box.calculateBlockHeight()

	// Handle <center> element - center children horizontally
	// HTML 4.01 §15.1.2: The CENTER element centers content
	if box.StyledNode != nil && box.StyledNode.Node != nil &&
		box.StyledNode.Node.Type == dom.ElementNode &&
		box.StyledNode.Node.Data == "center" {
		box.applyHorizontalAlignment("center")
	}
}

// calculateBlockWidth calculates the width of a block box.
// CSS 2.1 §10.3.3 Block-level, non-replaced elements in normal flow
func (box *LayoutBox) calculateBlockWidth(containingBlock Dimensions) {
	styles := box.StyledNode.Styles

	// Default to auto
	width := parseLength(styles["width"], containingBlock.Content.Width)

	// Margins (default to 0 if not specified)
	marginLeft := parseLengthOr0(styles["margin-left"], containingBlock.Content.Width)
	marginRight := parseLengthOr0(styles["margin-right"], containingBlock.Content.Width)

	// Padding (default to 0)
	paddingLeft := parseLengthOr0(styles["padding-left"], containingBlock.Content.Width)
	paddingRight := parseLengthOr0(styles["padding-right"], containingBlock.Content.Width)

	// Border (default to 0)
	borderLeft := parseLengthOr0(styles["border-left-width"], containingBlock.Content.Width)
	borderRight := parseLengthOr0(styles["border-right-width"], containingBlock.Content.Width)

	// Calculate total width
	total := marginLeft + marginRight + borderLeft + borderRight +
		paddingLeft + paddingRight + width

	// If width is not auto and total is greater than container, treat auto margins as 0
	// CSS 2.1 §10.3.3: over-constrained, solve for margin-right
	if width >= 0 && total > containingBlock.Content.Width {
		marginRight = containingBlock.Content.Width - width - marginLeft -
			borderLeft - borderRight - paddingLeft - paddingRight
	}

	// If width is auto, calculate it
	if width < 0 {
		width = containingBlock.Content.Width - marginLeft - marginRight -
			borderLeft - borderRight - paddingLeft - paddingRight
		if width < 0 {
			width = 0
		}
	}

	box.Dimensions.Content.Width = width
	box.Dimensions.Padding.Left = paddingLeft
	box.Dimensions.Padding.Right = paddingRight
	box.Dimensions.Border.Left = borderLeft
	box.Dimensions.Border.Right = borderRight
	box.Dimensions.Margin.Left = marginLeft
	box.Dimensions.Margin.Right = marginRight
}

// calculateBlockPosition calculates the position of a block box.
// CSS 2.1 §10.6.3 Block-level non-replaced elements in normal flow
func (box *LayoutBox) calculateBlockPosition(containingBlock Dimensions) {
	styles := box.StyledNode.Styles

	// Margin (default to 0)
	box.Dimensions.Margin.Top = parseLengthOr0(styles["margin-top"], containingBlock.Content.Width)
	box.Dimensions.Margin.Bottom = parseLengthOr0(styles["margin-bottom"], containingBlock.Content.Width)

	// Padding (default to 0)
	box.Dimensions.Padding.Top = parseLengthOr0(styles["padding-top"], containingBlock.Content.Width)
	box.Dimensions.Padding.Bottom = parseLengthOr0(styles["padding-bottom"], containingBlock.Content.Width)

	// Border (default to 0)
	box.Dimensions.Border.Top = parseLengthOr0(styles["border-top-width"], containingBlock.Content.Width)
	box.Dimensions.Border.Bottom = parseLengthOr0(styles["border-bottom-width"], containingBlock.Content.Width)

	// Position box below previous sibling or at top of container
	box.Dimensions.Content.X = containingBlock.Content.X +
		box.Dimensions.Margin.Left +
		box.Dimensions.Border.Left +
		box.Dimensions.Padding.Left

	box.Dimensions.Content.Y = containingBlock.Content.Y +
		containingBlock.Content.Height +
		box.Dimensions.Margin.Top +
		box.Dimensions.Border.Top +
		box.Dimensions.Padding.Top
}

// layoutBlockChildren lays out the children of a block box.
func (box *LayoutBox) layoutBlockChildren() {
	for i := 0; i < len(box.Children); {
		child := box.Children[i]

		// CSS 2.1 §9.4.2: Inline formatting context
		if child.isInlineLevel() {
			inlineRun := make([]*LayoutBox, 0)
			for i < len(box.Children) && box.Children[i].isInlineLevel() {
				inlineRun = append(inlineRun, box.Children[i])
				i++
			}
			box.layoutInlineChildren(inlineRun)
			continue
		}

		// Block-level layout (existing behavior)
		child.Layout(box.Dimensions)
		box.Dimensions.Content.Height += child.marginBox().Height
		i++
	}
}

// layoutInlineChildren lays out a run of inline-level children within this block box.
// CSS 2.1 §9.4.2 Inline formatting contexts: inline-level boxes are laid out in horizontal line boxes.
// CSS 2.1 §10.8: Line height and baseline alignment
func (box *LayoutBox) layoutInlineChildren(children []*LayoutBox) {
	if len(children) == 0 {
		return
	}

	currentX := box.Dimensions.Content.X
	currentY := box.Dimensions.Content.Y + box.Dimensions.Content.Height

	// First pass: layout all children to calculate their dimensions
	// and find the maximum baseline (for baseline alignment)
	maxBaseline := 0.0
	maxHeight := 0.0

	for i, child := range children {
		inlineCB := Dimensions{
			Content: Rect{
				X:      currentX,
				Y:      currentY,
				Width:  math.Max(0, box.Dimensions.Content.Width-(currentX-box.Dimensions.Content.X)),
				Height: 0,
			},
		}

		child.Layout(inlineCB)

		// Position child horizontally
		child.Dimensions.Content.X = currentX + child.Dimensions.Margin.Left + child.Dimensions.Border.Left + child.Dimensions.Padding.Left
		child.Dimensions.Content.Y = currentY + child.Dimensions.Margin.Top + child.Dimensions.Border.Top + child.Dimensions.Padding.Top

		currentX += child.marginBox().Width

		// CSS 2.1 §16.4: Add word-spacing between adjacent inline elements
		if i < len(children)-1 {
			currentX += calculateWordSpacing(child)
		}

		// Track maximum baseline offset (font size approximates baseline position)
		baseline := getBaseline(child)
		if baseline > maxBaseline {
			maxBaseline = baseline
		}

		if child.marginBox().Height > maxHeight {
			maxHeight = child.marginBox().Height
		}
	}

	// Second pass: align all children to the common baseline
	// CSS 2.1 §10.8.1: Inline elements are aligned by their baselines
	for _, child := range children {
		baseline := getBaseline(child)
		// Shift elements with smaller baselines down to align with the maximum baseline
		baselineOffset := maxBaseline - baseline
		if baselineOffset > 0 {
			child.shiftY(baselineOffset)
		}
	}

	// Increase parent height by the tallest inline box on this line
	box.Dimensions.Content.Height += maxHeight
}

// getBaseline returns the baseline offset for an inline element.
// CSS 2.1 §10.8.1: The baseline of an inline element is determined by its content.
// For text, the baseline is where the bottom of most letters (excluding descenders) sit.
// For replaced elements and other non-text content, we use the bottom edge as a fallback.
func getBaseline(box *LayoutBox) float64 {
	// For text nodes and elements with font styling, use font-based baseline
	if box.StyledNode != nil {
		fontSize := extractFontSize(box.StyledNode.Styles)
		// CSS 2.1 §10.8.1: Baseline is at baselinePositionEm of font size from top
		return fontSize * baselinePositionEm
	}
	// For elements without styling (rare), use content height as baseline
	// This is a fallback for replaced elements (CSS 2.1 §10.8.1)
	return box.Dimensions.Content.Height
}

// layoutInlineBox lays out an inline box and its inline children.
// CSS 2.1 §9.4.2 Inline formatting contexts
// CSS 2.1 §10.8: Line height and baseline alignment
func (box *LayoutBox) layoutInlineBox(containingBlock Dimensions) {
	styles := box.StyledNode.Styles

	// Apply inline box model properties
	box.Dimensions.Margin.Left = parseLengthOr0(styles["margin-left"], containingBlock.Content.Width)
	box.Dimensions.Margin.Right = parseLengthOr0(styles["margin-right"], containingBlock.Content.Width)
	box.Dimensions.Margin.Top = parseLengthOr0(styles["margin-top"], containingBlock.Content.Width)
	box.Dimensions.Margin.Bottom = parseLengthOr0(styles["margin-bottom"], containingBlock.Content.Width)

	box.Dimensions.Padding.Left = parseLengthOr0(styles["padding-left"], containingBlock.Content.Width)
	box.Dimensions.Padding.Right = parseLengthOr0(styles["padding-right"], containingBlock.Content.Width)
	box.Dimensions.Padding.Top = parseLengthOr0(styles["padding-top"], containingBlock.Content.Width)
	box.Dimensions.Padding.Bottom = parseLengthOr0(styles["padding-bottom"], containingBlock.Content.Width)

	box.Dimensions.Border.Left = parseLengthOr0(styles["border-left-width"], containingBlock.Content.Width)
	box.Dimensions.Border.Right = parseLengthOr0(styles["border-right-width"], containingBlock.Content.Width)
	box.Dimensions.Border.Top = parseLengthOr0(styles["border-top-width"], containingBlock.Content.Width)
	box.Dimensions.Border.Bottom = parseLengthOr0(styles["border-bottom-width"], containingBlock.Content.Width)

	// Position content relative to containing block and the box model edges
	box.Dimensions.Content.X = containingBlock.Content.X +
		box.Dimensions.Margin.Left +
		box.Dimensions.Border.Left +
		box.Dimensions.Padding.Left
	box.Dimensions.Content.Y = containingBlock.Content.Y +
		box.Dimensions.Margin.Top +
		box.Dimensions.Border.Top +
		box.Dimensions.Padding.Top

	currentX := box.Dimensions.Content.X
	currentY := box.Dimensions.Content.Y
	maxBaseline := 0.0
	maxHeight := 0.0

	// First pass: layout all children to calculate their dimensions
	for i, child := range box.Children {
		inlineCB := Dimensions{
			Content: Rect{
				X:      currentX,
				Y:      currentY,
				Width:  math.Max(0, containingBlock.Content.Width-(currentX-containingBlock.Content.X)),
				Height: 0,
			},
		}

		child.Layout(inlineCB)

		// Position child horizontally
		child.Dimensions.Content.X = currentX + child.Dimensions.Margin.Left + child.Dimensions.Border.Left + child.Dimensions.Padding.Left
		child.Dimensions.Content.Y = currentY + child.Dimensions.Margin.Top + child.Dimensions.Border.Top + child.Dimensions.Padding.Top

		currentX += child.marginBox().Width

		// CSS 2.1 §16.4: Add word-spacing between adjacent inline elements
		if i < len(box.Children)-1 {
			currentX += calculateWordSpacing(child)
		}

		// Track maximum baseline offset
		baseline := getBaseline(child)
		if baseline > maxBaseline {
			maxBaseline = baseline
		}

		if child.marginBox().Height > maxHeight {
			maxHeight = child.marginBox().Height
		}
	}

	// Second pass: align all children to the common baseline
	// CSS 2.1 §10.8.1: Inline elements are aligned by their baselines
	for _, child := range box.Children {
		baseline := getBaseline(child)
		baselineOffset := maxBaseline - baseline
		if baselineOffset > 0 {
			child.shiftY(baselineOffset)
		}
	}

	box.Dimensions.Content.Width = currentX - box.Dimensions.Content.X
	box.Dimensions.Content.Height = maxHeight
}

// calculateWordSpacing calculates the word spacing to add between inline elements.
// CSS 2.1 §16.4: Word spacing accounts for whitespace that was collapsed between elements.
// Returns defaultWordSpacingEm (0.25em) based on the element's font size.
func calculateWordSpacing(child *LayoutBox) float64 {
	fontSize := css.BaseFontHeight
	if child.StyledNode != nil {
		fontSize = extractFontSize(child.StyledNode.Styles)
	}
	return fontSize * defaultWordSpacingEm
}

// isInlineLevel returns true for inline boxes and text nodes.
// Note: Anonymous inline boxes (CSS 2.1 §9.2.2.1) are not generated yet; only explicit inline elements and text nodes are treated as inline.
func (box *LayoutBox) isInlineLevel() bool {
	if box.BoxType == InlineBox {
		return true
	}
	if box.StyledNode != nil && box.StyledNode.Node != nil && box.StyledNode.Node.Type == dom.TextNode {
		return true
	}
	return false
}

// calculateBlockHeight calculates the height of a block box.
// CSS 2.1 §10.6.3
func (box *LayoutBox) calculateBlockHeight() {
	// If height is explicitly set, use that
	if height := box.StyledNode.Styles["height"]; height != "" {
		if h := parseLength(height, 0); h >= 0 {
			box.Dimensions.Content.Height = h
		}
	}
	// Otherwise, height is already calculated from children
}

// parseLength parses a CSS length value.
// Returns -1 if the value is "auto" or invalid.
// CSS 2.1 §4.3.2 Lengths
func parseLength(value string, referenceLength float64) float64 {
	value = strings.TrimSpace(value)

	if value == "" || value == "auto" {
		return -1
	}

	// Parse percentage
	if strings.HasSuffix(value, "%") {
		if pct, err := strconv.ParseFloat(value[:len(value)-1], 64); err == nil {
			return referenceLength * pct / 100.0
		}
		return -1
	}

	// Parse pixels
	if strings.HasSuffix(value, "px") {
		if px, err := strconv.ParseFloat(value[:len(value)-2], 64); err == nil {
			return px
		}
		return -1
	}

	// Try parsing as a number (assume px)
	if num, err := strconv.ParseFloat(value, 64); err == nil {
		return num
	}

	return -1
}

// parseLengthOr0 parses a CSS length value, returning 0 if invalid or auto.
func parseLengthOr0(value string, referenceLength float64) float64 {
	result := parseLength(value, referenceLength)
	if result < 0 {
		return 0
	}
	return result
}

// marginBox returns the box including margin.
func (box *LayoutBox) marginBox() Rect {
	return expandRect(box.borderBox(), box.Dimensions.Margin)
}

// borderBox returns the box including border.
func (box *LayoutBox) borderBox() Rect {
	return expandRect(box.paddingBox(), box.Dimensions.Border)
}

// paddingBox returns the box including padding.
func (box *LayoutBox) paddingBox() Rect {
	return expandRect(box.Dimensions.Content, box.Dimensions.Padding)
}

// expandRect expands a rectangle by edge sizes.
func expandRect(rect Rect, edges EdgeSizes) Rect {
	return Rect{
		X:      rect.X - edges.Left,
		Y:      rect.Y - edges.Top,
		Width:  rect.Width + edges.Left + edges.Right,
		Height: rect.Height + edges.Top + edges.Bottom,
	}
}

// applyExplicitRowHeight applies an explicit height style to an empty table row.
// CSS 2.1 §17.5.3: Table row height can be explicitly set.
// Returns true if height was applied (indicating the caller should return early).
func applyExplicitRowHeight(box *LayoutBox, styles map[string]string) bool {
	if height := styles["height"]; height != "" {
		if h := parseLength(height, 0); h >= 0 {
			box.Dimensions.Content.Height = h
		}
	}
	return true
}

// layoutText lays out a text node.
// CSS 2.1 §16 Text
func (box *LayoutBox) layoutText(containingBlock Dimensions) {
	// Get the text content
	text := box.StyledNode.Node.Data
	if text == "" {
		box.Dimensions.Content.Width = 0
		box.Dimensions.Content.Height = 0
		return
	}

	// CSS 2.1 §16.6.1: Collapse whitespace for layout calculations
	// This ensures dimensions match what will actually be rendered
	text = collapseWhitespace(text)

	if text == "" {
		box.Dimensions.Content.Width = 0
		box.Dimensions.Content.Height = 0
		return
	}

	// Get font size and style from CSS styles (CSS 2.1 §15 Fonts)
	fontSize := extractFontSize(box.StyledNode.Styles)
	fontWeight := extractFontWeight(box.StyledNode.Styles)
	fontStyleStr := extractFontStyle(box.StyledNode.Styles)
	
	// Measure text using shared font.MeasureText
	// This ensures layout and rendering use the same measurements
	fontStyle := font.Style{
		Size:   fontSize,
		Weight: fontWeight,
		Style:  fontStyleStr,
	}
	width, height := font.MeasureText(text, fontStyle)

	// Position the text node
	box.Dimensions.Content.X = containingBlock.Content.X
	box.Dimensions.Content.Y = containingBlock.Content.Y + containingBlock.Content.Height
	box.Dimensions.Content.Width = width
	box.Dimensions.Content.Height = height
}

// extractFontSize extracts the font-size from CSS styles and returns it in pixels.
// CSS 2.1 §15.7 Font size: the 'font-size' property
func extractFontSize(styles map[string]string) float64 {
	fontSize := styles["font-size"]
	if fontSize == "" {
		return css.BaseFontHeight // Default font size
	}

	if size := css.ParseFontSize(fontSize); size > 0 {
		return size
	}

	return css.BaseFontHeight // Default font size
}

// collapseWhitespace collapses consecutive whitespace characters into a single space.
// CSS 2.1 §16.6.1: The white-space property
// For normal text (white-space: normal, which is the default):
// - Sequences of whitespace (space, tab, newline, carriage return) are collapsed into a single space
// - Leading and trailing whitespace is removed
func collapseWhitespace(text string) string {
	if text == "" {
		return text
	}

	var result strings.Builder
	lastWasSpace := true // Start as true to trim leading whitespace

	for _, ch := range text {
		// Check if character is whitespace (space, tab, newline, carriage return)
		isSpace := ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'

		if isSpace {
			// Only add a space if we haven't just added one
			if !lastWasSpace {
				result.WriteRune(' ')
				lastWasSpace = true
			}
		} else {
			// Regular character - add it
			result.WriteRune(ch)
			lastWasSpace = false
		}
	}

	// Trim trailing whitespace
	output := result.String()
	if lastWasSpace && len(output) > 0 {
		output = output[:len(output)-1]
	}

	return output
}

// layoutTable lays out a table element.
// CSS 2.1 §17.5 Visual layout of table contents
// CSS 2.1 §17.6.1: Border-spacing property adds space between cells
func (box *LayoutBox) layoutTable(containingBlock Dimensions) {
	// CSS 2.1 §17.6.2: Border-collapse model - not yet implemented
	if borderCollapse := box.StyledNode.Styles["border-collapse"]; borderCollapse == "collapse" {
		log.Warnf("CSS 2.1 §17.6.2: border-collapse:collapse not yet implemented, using separate borders")
	}
	
	// Calculate table width (similar to block)
	box.calculateBlockWidth(containingBlock)
	box.calculateBlockPosition(containingBlock)

	// Get border-spacing value (CSS 2.1 §17.6.1)
	// For simplicity, we use the same spacing for horizontal and vertical
	// (full spec supports "horizontal vertical" syntax)
	borderSpacing := 2.0 // Default from user-agent stylesheet
	if spacing := box.StyledNode.Styles["border-spacing"]; spacing != "" {
		if parsed := parseLength(spacing, 0); parsed >= 0 {
			borderSpacing = parsed
		}
	}

	// Calculate the number of columns in the table
	// CSS 2.1 §17.2.1: The number of columns is determined by examining all rows
	numColumns := box.calculateTableColumns()

	// Calculate column widths based on content
	// CSS 2.1 §17.5.2.2: Auto table layout
	columnWidths := box.calculateColumnWidths(numColumns, box.Dimensions.Content.Width)

	// Layout table rows with column widths and border spacing
	for _, row := range box.Children {
		if row.BoxType == TableRowBox {
			row.layoutWithColumnWidths(box.Dimensions, columnWidths, borderSpacing)
			box.Dimensions.Content.Height += row.marginBox().Height
		}
	}

	// If height is explicitly set, use that
	box.calculateBlockHeight()
}

// calculateTableColumns calculates the number of columns in a table.
// CSS 2.1 §17.2.1: Counts columns based on table cells and colspan attributes
func (box *LayoutBox) calculateTableColumns() int {
	maxColumns := 0

	for _, row := range box.Children {
		if row.BoxType == TableRowBox {
			columnCount := 0
			for _, cell := range row.Children {
				if cell.BoxType == TableCellBox {
					columnCount += getColspan(cell)
				}
			}
			if columnCount > maxColumns {
				maxColumns = columnCount
			}
		}
	}

	if maxColumns == 0 {
		maxColumns = 1
	}

	return maxColumns
}

// getColspan extracts the colspan attribute from a table cell.
// CSS 2.1 §17.2.1: Returns 1 if no colspan attribute is present
// HTML5 recommends a maximum colspan of 1000 to prevent performance issues
func getColspan(cell *LayoutBox) int {
	if cell.StyledNode == nil || cell.StyledNode.Node == nil {
		return 1
	}

	// CSS 2.1 §17.2: rowspan attribute - not yet implemented
	if rowspanStr := cell.StyledNode.Node.GetAttribute("rowspan"); rowspanStr != "" && rowspanStr != "1" {
		log.Warnf("CSS 2.1 §17.2: rowspan attribute not yet implemented (found rowspan=%s)", rowspanStr)
	}

	colspanStr := cell.StyledNode.Node.GetAttribute("colspan")
	if colspanStr == "" {
		return 1
	}

	if val, err := strconv.Atoi(colspanStr); err == nil && val > 0 {
		// Cap at maxColspan as recommended by HTML specification
		if val > maxColspan {
			return maxColspan
		}
		return val
	}

	return 1
}

// calculateColumnWidths calculates preferred widths for table columns.
// CSS 2.1 §17.5.2.2: Auto table layout algorithm (non-normative)
// The spec does not fully define this algorithm. This implementation:
// 1. Calculates minimum content width for each column
// 2. Distributes remaining space proportionally to content width
func (box *LayoutBox) calculateColumnWidths(numColumns int, tableWidth float64) []float64 {
	// Collect column content sizes from all rows
	columnMinWidths := make([]float64, numColumns)
	
	for _, row := range box.Children {
		if row.BoxType == TableRowBox {
			colIndex := 0
			for _, cell := range row.Children {
				if cell.BoxType == TableCellBox {
					colspan := getColspan(cell)
					minWidth := box.estimateCellMinWidth(cell)
					
					if colspan == 1 && colIndex < numColumns {
						if minWidth > columnMinWidths[colIndex] {
							columnMinWidths[colIndex] = minWidth
						}
					}

					colIndex += colspan
				}
			}
		}
	}

	totalMinWidth := 0.0
	for _, w := range columnMinWidths {
		totalMinWidth += w
	}

	columnWidths := make([]float64, numColumns)
	if totalMinWidth > 0 {
		if totalMinWidth <= tableWidth {
			// Distribute extra space proportionally to content width
			scale := tableWidth / totalMinWidth
			for i, w := range columnMinWidths {
				columnWidths[i] = w * scale
			}
		} else {
			// Content exceeds table width - use minimum widths
			copy(columnWidths, columnMinWidths)
		}
	} else {
		// Equal distribution if no content found
		equalWidth := tableWidth / float64(numColumns)
		for i := range columnWidths {
			columnWidths[i] = equalWidth
		}
	}

	return columnWidths
}

// estimateCellMinWidth estimates the minimum width needed for a cell's content
func (box *LayoutBox) estimateCellMinWidth(cell *LayoutBox) float64 {
	minWidth := 30.0 // Default minimum

	// Check for explicit width style
	if cell.StyledNode != nil {
		if widthStr := cell.StyledNode.Styles["width"]; widthStr != "" {
			if w := parseLength(widthStr, 0); w > 0 {
				return w + 20 // Add some padding
			}
		}
	}

	// Estimate based on content, but cap at reasonable maximum
	contentWidth := box.estimateContentWidth(cell)
	if contentWidth > minWidth {
		minWidth = contentWidth
	}

	// Cap at reasonable maximum to prevent extremely wide content from creating unusable layouts
	if minWidth > maxColumnWidth {
		minWidth = maxColumnWidth
	}

	return minWidth
}

// estimateContentWidth estimates the width of content in a box
func (box *LayoutBox) estimateContentWidth(layoutBox *LayoutBox) float64 {
	width := 0.0

	// Use the same character width as in layoutText
	face := basicfont.Face7x13
	charWidth := float64(face.Advance)

	// Check for explicit width on the element itself
	if layoutBox.StyledNode != nil {
		if widthStr := layoutBox.StyledNode.Styles["width"]; widthStr != "" {
			if w := parseLength(widthStr, 0); w > 0 {
				// Add padding if specified
				paddingLeft := parseLengthOr0(layoutBox.StyledNode.Styles["padding-left"], 0)
				paddingRight := parseLengthOr0(layoutBox.StyledNode.Styles["padding-right"], 0)
				return w + paddingLeft + paddingRight
			}
		}
	}

	// Recursively estimate width from children
	for _, child := range layoutBox.Children {
		if child.StyledNode != nil && child.StyledNode.Node != nil {
			if child.StyledNode.Node.Type == dom.TextNode {
				// Estimate text width accounting for font size and whitespace collapsing
				// CSS 2.1 §16.6.1: Collapse whitespace for width calculations
				text := collapseWhitespace(child.StyledNode.Node.Data)
				fontSize := extractFontSize(child.StyledNode.Styles)
				scale := fontSize / css.BaseFontHeight
				textWidth := float64(len(text)) * charWidth * scale
				width += textWidth
			} else {
				// Check for explicit width on child element
				if widthStr := child.StyledNode.Styles["width"]; widthStr != "" {
					if w := parseLength(widthStr, 0); w > 0 {
						paddingLeft := parseLengthOr0(child.StyledNode.Styles["padding-left"], 0)
						paddingRight := parseLengthOr0(child.StyledNode.Styles["padding-right"], 0)
						width += w + paddingLeft + paddingRight
						continue
					}
				}
				// Recursively estimate child width
				childWidth := box.estimateContentWidth(child)
				width += childWidth
			}
		}
	}

	// Add padding if specified
	if layoutBox.StyledNode != nil {
		paddingLeft := parseLengthOr0(layoutBox.StyledNode.Styles["padding-left"], 0)
		paddingRight := parseLengthOr0(layoutBox.StyledNode.Styles["padding-right"], 0)
		width += paddingLeft + paddingRight
	}

	return width
}

// layoutTableRow lays out a table row.
// CSS 2.1 §17.5.3 Table height algorithms
func (box *LayoutBox) layoutTableRow(containingBlock Dimensions) {
	// Calculate number of columns to maintain consistency with auto layout
	// This ensures equal distribution is based on the correct column count
	numColumns := 0
	for _, cell := range box.Children {
		if cell.BoxType == TableCellBox {
			numColumns += getColspan(cell)
		}
	}
	if numColumns == 0 {
		numColumns = len(box.Children)
	}

	box.layoutWithColumns(containingBlock, numColumns)
}

// layoutWithColumnWidths lays out a table row with pre-calculated column widths.
// CSS 2.1 §17.5.2.2: Auto table layout
func (box *LayoutBox) layoutWithColumnWidths(containingBlock Dimensions, columnWidths []float64, borderSpacing float64) {
	styles := box.StyledNode.Styles

	// Calculate position
	box.Dimensions.Margin.Top = parseLengthOr0(styles["margin-top"], containingBlock.Content.Width)
	box.Dimensions.Margin.Bottom = parseLengthOr0(styles["margin-bottom"], containingBlock.Content.Width)
	box.Dimensions.Padding.Top = parseLengthOr0(styles["padding-top"], containingBlock.Content.Width)
	box.Dimensions.Padding.Bottom = parseLengthOr0(styles["padding-bottom"], containingBlock.Content.Width)
	box.Dimensions.Border.Top = parseLengthOr0(styles["border-top-width"], containingBlock.Content.Width)
	box.Dimensions.Border.Bottom = parseLengthOr0(styles["border-bottom-width"], containingBlock.Content.Width)

	// Position row
	box.Dimensions.Content.X = containingBlock.Content.X
	box.Dimensions.Content.Y = containingBlock.Content.Y + containingBlock.Content.Height +
		box.Dimensions.Margin.Top + box.Dimensions.Border.Top + box.Dimensions.Padding.Top
	box.Dimensions.Content.Width = containingBlock.Content.Width

	// Handle empty rows (e.g., spacer rows) - apply explicit height and return
	if len(box.Children) == 0 {
		applyExplicitRowHeight(box, styles)
		return
	}

	// Layout each cell horizontally using calculated widths
	// CSS 2.1 §17.6.1: border-spacing adds space between cells
	currentX := box.Dimensions.Content.X + borderSpacing // Start with border-spacing on the left
	currentCol := 0
	maxHeight := 0.0

	for _, cell := range box.Children {
		if cell.BoxType == TableCellBox {
			colspan := getColspan(cell)

			// Calculate cell width by summing column widths
			cellWidth := 0.0
			for i := 0; i < colspan && currentCol+i < len(columnWidths); i++ {
				cellWidth += columnWidths[currentCol+i]
			}

			// Create a containing block for the cell with calculated width
			cellContainingBlock := Dimensions{
				Content: Rect{
					X:      currentX,
					Y:      box.Dimensions.Content.Y,
					Width:  cellWidth,
					Height: 0,
				},
			}
			cell.Layout(cellContainingBlock)

			// Update position for next cell
			// CSS 2.1 §17.6.1: Add cell width plus border-spacing
			currentX += cell.marginBox().Width + borderSpacing
			currentCol += colspan

			// Track maximum height
			if cell.marginBox().Height > maxHeight {
				maxHeight = cell.marginBox().Height
			}
		}
	}

	// Set row height to maximum cell height
	box.Dimensions.Content.Height = maxHeight

	// If row has explicit height, use that instead
	if height := styles["height"]; height != "" {
		if h := parseLength(height, 0); h >= 0 {
			box.Dimensions.Content.Height = h
		}
	}
}

// layoutWithColumns lays out a table row with a specified column count.
// CSS 2.1 §17.5.2: Table width algorithms
func (box *LayoutBox) layoutWithColumns(containingBlock Dimensions, numColumns int) {
	styles := box.StyledNode.Styles

	// Calculate position
	box.Dimensions.Margin.Top = parseLengthOr0(styles["margin-top"], containingBlock.Content.Width)
	box.Dimensions.Margin.Bottom = parseLengthOr0(styles["margin-bottom"], containingBlock.Content.Width)
	box.Dimensions.Padding.Top = parseLengthOr0(styles["padding-top"], containingBlock.Content.Width)
	box.Dimensions.Padding.Bottom = parseLengthOr0(styles["padding-bottom"], containingBlock.Content.Width)
	box.Dimensions.Border.Top = parseLengthOr0(styles["border-top-width"], containingBlock.Content.Width)
	box.Dimensions.Border.Bottom = parseLengthOr0(styles["border-bottom-width"], containingBlock.Content.Width)

	// Position row
	box.Dimensions.Content.X = containingBlock.Content.X
	box.Dimensions.Content.Y = containingBlock.Content.Y + containingBlock.Content.Height +
		box.Dimensions.Margin.Top + box.Dimensions.Border.Top + box.Dimensions.Padding.Top
	box.Dimensions.Content.Width = containingBlock.Content.Width

	// Handle empty rows (e.g., spacer rows) - apply explicit height and return
	if len(box.Children) == 0 || numColumns == 0 {
		applyExplicitRowHeight(box, styles)
		return
	}

	// Calculate width per column
	// CSS 2.1 §17.5.2.1: In the fixed table layout algorithm
	columnWidth := containingBlock.Content.Width / float64(numColumns)

	// Layout each cell horizontally
	currentX := box.Dimensions.Content.X
	maxHeight := 0.0

	for _, cell := range box.Children {
		if cell.BoxType == TableCellBox {
			colspan := getColspan(cell)

			// Calculate cell width based on colspan
			cellWidth := columnWidth * float64(colspan)

			// Create a containing block for the cell with calculated width
			cellContainingBlock := Dimensions{
				Content: Rect{
					X:      currentX,
					Y:      box.Dimensions.Content.Y,
					Width:  cellWidth,
					Height: 0,
				},
			}
			cell.Layout(cellContainingBlock)

			// Update position for next cell
			currentX += cell.marginBox().Width

			// Track maximum height
			if cell.marginBox().Height > maxHeight {
				maxHeight = cell.marginBox().Height
			}
		}
	}

	// Set row height to maximum cell height
	// Note: maxHeight includes cell margins, padding, and borders since we use marginBox()
	// The row's content height encompasses the full height of its cells
	box.Dimensions.Content.Height = maxHeight

	// If row has explicit height, use that instead
	if height := styles["height"]; height != "" {
		if h := parseLength(height, 0); h >= 0 {
			box.Dimensions.Content.Height = h
		}
	}
}

// layoutTableCell lays out a table cell.
// CSS 2.1 §17.5.3 Table height algorithms
func (box *LayoutBox) layoutTableCell(containingBlock Dimensions) {
	styles := box.StyledNode.Styles

	// Parse width - if specified, use it; otherwise use the width from containing block
	width := parseLength(styles["width"], containingBlock.Content.Width)
	if width < 0 {
		width = containingBlock.Content.Width
	}

	// Padding
	paddingLeft := parseLengthOr0(styles["padding-left"], containingBlock.Content.Width)
	paddingRight := parseLengthOr0(styles["padding-right"], containingBlock.Content.Width)
	paddingTop := parseLengthOr0(styles["padding-top"], containingBlock.Content.Width)
	paddingBottom := parseLengthOr0(styles["padding-bottom"], containingBlock.Content.Width)

	// Border
	borderLeft := parseLengthOr0(styles["border-left-width"], containingBlock.Content.Width)
	borderRight := parseLengthOr0(styles["border-right-width"], containingBlock.Content.Width)
	borderTop := parseLengthOr0(styles["border-top-width"], containingBlock.Content.Width)
	borderBottom := parseLengthOr0(styles["border-bottom-width"], containingBlock.Content.Width)

	// Margin (typically 0 for table cells)
	marginLeft := parseLengthOr0(styles["margin-left"], containingBlock.Content.Width)
	marginRight := parseLengthOr0(styles["margin-right"], containingBlock.Content.Width)
	marginTop := parseLengthOr0(styles["margin-top"], containingBlock.Content.Width)
	marginBottom := parseLengthOr0(styles["margin-bottom"], containingBlock.Content.Width)

	// Calculate content width
	contentWidth := width - paddingLeft - paddingRight - borderLeft - borderRight - marginLeft - marginRight
	if contentWidth < 0 {
		contentWidth = 0
	}

	// Set dimensions
	box.Dimensions.Content.Width = contentWidth
	box.Dimensions.Content.X = containingBlock.Content.X + marginLeft + borderLeft + paddingLeft
	box.Dimensions.Content.Y = containingBlock.Content.Y + marginTop + borderTop + paddingTop

	box.Dimensions.Padding.Left = paddingLeft
	box.Dimensions.Padding.Right = paddingRight
	box.Dimensions.Padding.Top = paddingTop
	box.Dimensions.Padding.Bottom = paddingBottom

	box.Dimensions.Border.Left = borderLeft
	box.Dimensions.Border.Right = borderRight
	box.Dimensions.Border.Top = borderTop
	box.Dimensions.Border.Bottom = borderBottom

	box.Dimensions.Margin.Left = marginLeft
	box.Dimensions.Margin.Right = marginRight
	box.Dimensions.Margin.Top = marginTop
	box.Dimensions.Margin.Bottom = marginBottom

	// Layout children (cell content)
	for _, child := range box.Children {
		child.Layout(box.Dimensions)
		box.Dimensions.Content.Height += child.marginBox().Height
	}

	// If height is explicitly set, use that
	if height := styles["height"]; height != "" {
		if h := parseLength(height, 0); h >= 0 {
			box.Dimensions.Content.Height = h
		}
	}

	// Apply HTML align attribute for horizontal alignment
	// HTML 4.01 §11.3.2: The align attribute specifies horizontal alignment
	// Supported values: left, center, right
	// Note: CSS text-align property (§16.2) is NOT yet implemented. It would
	// affect inline content rendering within block containers, not child box
	// positioning like the HTML align attribute does here.
	if align := box.StyledNode.Node.GetAttribute("align"); align != "" {
		box.applyHorizontalAlignment(align)
	}

	// Apply HTML valign attribute for vertical alignment
	// HTML 4.01 §11.3.2: The valign attribute specifies vertical alignment
	// Supported values: top, middle, bottom
	if valign := box.StyledNode.Node.GetAttribute("valign"); valign != "" {
		box.applyVerticalAlignment(valign)
	}
}

// applyHorizontalAlignment adjusts child positions based on HTML align attribute.
// HTML 4.01 §11.3.2: The align attribute specifies horizontal alignment in table cells.
func (box *LayoutBox) applyHorizontalAlignment(align string) {
	align = strings.ToLower(strings.TrimSpace(align))

	if len(box.Children) == 0 {
		return
	}

	// Calculate the total width of all children
	totalChildWidth := 0.0
	for _, child := range box.Children {
		totalChildWidth += child.marginBox().Width
	}

	// Calculate available space
	availableSpace := box.Dimensions.Content.Width - totalChildWidth

	if availableSpace <= 0 {
		return // No space to align
	}

	var offset float64
	switch align {
	case "right":
		offset = availableSpace
	case "center":
		offset = availableSpace / 2.0
	case "left":
		// Default, no offset needed
		offset = 0
	default:
		return // Unknown alignment, do nothing
	}

	// Adjust X position of all children recursively
	for _, child := range box.Children {
		child.shiftX(offset)
	}
}

// shiftX recursively shifts this box and all its descendants by the given offset.
// This is used when applying horizontal alignment to ensure all nested elements
// are moved together.
func (box *LayoutBox) shiftX(offset float64) {
	box.Dimensions.Content.X += offset
	for _, child := range box.Children {
		child.shiftX(offset)
	}
}

// applyVerticalAlignment adjusts child positions based on HTML valign attribute.
// HTML 4.01 §11.3.2: The valign attribute specifies vertical alignment in table cells.
func (box *LayoutBox) applyVerticalAlignment(valign string) {
	valign = strings.ToLower(strings.TrimSpace(valign))

	if len(box.Children) == 0 {
		return
	}

	// Calculate the total height of all children
	totalChildHeight := 0.0
	for _, child := range box.Children {
		totalChildHeight += child.marginBox().Height
	}

	// Calculate available space
	availableSpace := box.Dimensions.Content.Height - totalChildHeight

	if availableSpace <= 0 {
		return // No space to align
	}

	var offset float64
	switch valign {
	case "bottom":
		offset = availableSpace
	case "middle":
		offset = availableSpace / 2.0
	case "top":
		// Default, no offset needed
		offset = 0
	default:
		return // Unknown alignment, do nothing
	}

	// Adjust Y position of all children recursively
	for _, child := range box.Children {
		child.shiftY(offset)
	}
}

// shiftY recursively shifts this box and all its descendants by the given offset.
// This is used when applying vertical alignment to ensure all nested elements
// are moved together.
func (box *LayoutBox) shiftY(offset float64) {
	box.Dimensions.Content.Y += offset
	for _, child := range box.Children {
		child.shiftY(offset)
	}
}

// layoutFlex lays out a flex container using the CSS Flexible Box Layout Module.
// Spec reference: CSS Flexible Box Layout Module Level 1
// https://www.w3.org/TR/css-flexbox-1/
//
// Currently supported:
// - display: flex (block-level flex container)
// - flex-direction: row (main axis is horizontal, left to right)
// - justify-content: flex-start | center | flex-end | space-between
//
// Not yet implemented (graceful degradation with warnings):
// - flex-direction: column, row-reverse, column-reverse
// - justify-content: space-around, space-evenly
// - align-items, align-content, align-self
// - flex-wrap property
// - flex, flex-grow, flex-shrink, flex-basis properties for items
// - order property
func (box *LayoutBox) layoutFlex(containingBlock Dimensions) {
	styles := box.StyledNode.Styles
	
	// Parse flex-direction (default: row)
	flexDirection := strings.TrimSpace(strings.ToLower(styles["flex-direction"]))
	if flexDirection == "" {
		flexDirection = "row"
	}
	
	// Warn about unsupported flex-direction values
	if flexDirection != "row" {
		log.Warnf("CSS3 Flexbox: flex-direction:%s not yet implemented, using 'row'", flexDirection)
		flexDirection = "row"
	}
	
	// Parse justify-content (default: flex-start)
	justifyContent := strings.TrimSpace(strings.ToLower(styles["justify-content"]))
	if justifyContent == "" {
		justifyContent = "flex-start"
	}
	
	// Warn about unsupported justify-content values
	supportedJustify := map[string]bool{
		"flex-start":    true,
		"center":        true,
		"flex-end":      true,
		"space-between": true,
	}
	if !supportedJustify[justifyContent] {
		log.Warnf("CSS3 Flexbox: justify-content:%s not yet implemented, using 'flex-start'", justifyContent)
		justifyContent = "flex-start"
	}
	
	// Warn about unsupported flex properties
	if alignItems := styles["align-items"]; alignItems != "" && alignItems != "stretch" {
		log.Warnf("CSS3 Flexbox: align-items:%s not yet implemented", alignItems)
	}
	if alignContent := styles["align-content"]; alignContent != "" && alignContent != "stretch" {
		log.Warnf("CSS3 Flexbox: align-content:%s not yet implemented", alignContent)
	}
	if flexWrap := styles["flex-wrap"]; flexWrap != "" && flexWrap != "nowrap" {
		log.Warnf("CSS3 Flexbox: flex-wrap:%s not yet implemented", flexWrap)
	}
	
	// Calculate flex container dimensions (similar to block)
	box.calculateBlockWidth(containingBlock)
	box.calculateBlockPosition(containingBlock)
	
	// Layout flex items with flex-direction: row
	box.layoutFlexRow(justifyContent)
	
	// Calculate height based on tallest item
	box.calculateBlockHeight()
}

// layoutFlexRow lays out flex items in a row with the given justify-content alignment.
// Implements the main-axis alignment for flex-direction: row.
func (box *LayoutBox) layoutFlexRow(justifyContent string) {
	if len(box.Children) == 0 {
		return
	}
	
	// First pass: layout all items to determine their dimensions
	// Flex items are laid out as block-level boxes to calculate their preferred sizes
	totalItemWidth := 0.0
	maxHeight := 0.0
	
	for _, child := range box.Children {
		// Create a containing block for the flex item
		// In flex layout, items have their own sizing rules
		itemCB := Dimensions{
			Content: Rect{
				X:      box.Dimensions.Content.X,
				Y:      box.Dimensions.Content.Y,
				Width:  box.Dimensions.Content.Width,
				Height: 0,
			},
		}
		
		// Layout the child as a block to get its dimensions
		child.Layout(itemCB)
		
		// Accumulate total width (including margins)
		totalItemWidth += child.marginBox().Width
		
		// Track maximum height for flex container
		itemHeight := child.marginBox().Height
		if itemHeight > maxHeight {
			maxHeight = itemHeight
		}
	}
	
	// Second pass: position items according to justify-content
	currentX := box.Dimensions.Content.X
	gap := 0.0
	
	// Calculate spacing based on justify-content
	availableSpace := box.Dimensions.Content.Width - totalItemWidth
	
	switch justifyContent {
	case "flex-start":
		// Items are packed at the start (default behavior, currentX already set)
	case "center":
		// Items are centered in the container
		currentX += availableSpace / 2.0
	case "flex-end":
		// Items are packed at the end
		currentX += availableSpace
	case "space-between":
		// Items are evenly distributed with space between them
		if len(box.Children) > 1 {
			gap = availableSpace / float64(len(box.Children)-1)
		}
	}
	
	// Position each flex item
	for i, child := range box.Children {
		// Adjust X position
		offsetX := currentX - child.Dimensions.Content.X + child.Dimensions.Margin.Left + child.Dimensions.Border.Left + child.Dimensions.Padding.Left
		child.shiftX(offsetX)
		
		// Move to next item position
		currentX += child.marginBox().Width
		
		// Add gap for space-between (except after last item)
		if justifyContent == "space-between" && i < len(box.Children)-1 {
			currentX += gap
		}
	}
	
	// Set flex container height to the tallest item
	box.Dimensions.Content.Height = maxHeight
}

// extractFontWeight extracts font-weight from CSS styles.
// CSS 2.1 §15.6: Parse font-weight
func extractFontWeight(styles map[string]string) string {
	fontWeight := styles["font-weight"]
	if fontWeight == "" {
		return "normal"
	}
	
	fontWeight = strings.TrimSpace(strings.ToLower(fontWeight))
	if fontWeight == "bold" || fontWeight == "bolder" {
		return "bold"
	}
	
	if weight, err := strconv.Atoi(fontWeight); err == nil && weight >= 600 {
		return "bold"
	}
	
	return "normal"
}

// extractFontStyle extracts font-style from CSS styles.
// CSS 2.1 §15.7: Parse font-style
func extractFontStyle(styles map[string]string) string {
	fontStyle := styles["font-style"]
	if fontStyle == "" {
		return "normal"
	}
	
	fontStyle = strings.TrimSpace(strings.ToLower(fontStyle))
	if fontStyle == "italic" || fontStyle == "oblique" {
		return "italic"
	}
	
	return "normal"
}
