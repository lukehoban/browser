// Package layout implements the CSS 2.1 visual formatting model.
// It converts styled nodes into a tree of layout boxes with computed dimensions.
//
// Spec references:
// - CSS 2.1 §8 Box model: https://www.w3.org/TR/CSS21/box.html
// - CSS 2.1 §9 Visual formatting model: https://www.w3.org/TR/CSS21/visuren.html
// - CSS 2.1 §10 Visual formatting model details: https://www.w3.org/TR/CSS21/visudet.html
package layout

import (
	"strconv"
	"strings"

	"github.com/lukehoban/browser/dom"
	"github.com/lukehoban/browser/style"
	"golang.org/x/image/font/basicfont"
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
	// CSS 2.1 §16.6.1: Whitespace-only text nodes should not affect layout
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
			return nil // Don't create a box for whitespace-only text
		}
	}

	// Determine box type based on display property or HTML element
	// CSS 2.1 §17.2.1: The table element generates a principal table box
	boxType := BlockBox
	display := styledNode.Styles["display"]
	
	// If no explicit display property, infer from HTML element
	if display == "" && styledNode.Node != nil && styledNode.Node.Type == dom.ElementNode {
		switch styledNode.Node.Data {
		case "table":
			display = "table"
		case "tr":
			display = "table-row"
		case "td", "th":
			display = "table-cell"
		}
	}
	
	switch display {
	case "inline":
		boxType = InlineBox
	case "none":
		return nil // Don't create a box
	case "table":
		boxType = TableBox
	case "table-row":
		boxType = TableRowBox
	case "table-cell":
		boxType = TableCellBox
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
func (box *LayoutBox) Layout(containingBlock Dimensions) {
	switch box.BoxType {
	case BlockBox:
		box.layoutBlock(containingBlock)
	case InlineBox:
		// Simplified: treat inline as block for now
		box.layoutBlock(containingBlock)
	case TableBox:
		box.layoutTable(containingBlock)
	case TableRowBox:
		box.layoutTableRow(containingBlock)
	case TableCellBox:
		box.layoutTableCell(containingBlock)
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
	for _, child := range box.Children {
		child.Layout(box.Dimensions)
		// Update height to include child
		box.Dimensions.Content.Height += child.marginBox().Height
	}
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

	// Calculate text dimensions using basicfont.Face7x13
	// Note: For basicfont.Face7x13, all characters have fixed width (Advance)
	// For more accurate measurement, we could use font.Drawer.MeasureString()
	// but basicfont is monospaced so character count * Advance is accurate
	face := basicfont.Face7x13
	width := float64(len(text) * face.Advance)
	height := float64(face.Height)

	// Position the text node
	box.Dimensions.Content.X = containingBlock.Content.X
	box.Dimensions.Content.Y = containingBlock.Content.Y + containingBlock.Content.Height
	box.Dimensions.Content.Width = width
	box.Dimensions.Content.Height = height
}

// layoutTable lays out a table element.
// CSS 2.1 §17.5 Visual layout of table contents
func (box *LayoutBox) layoutTable(containingBlock Dimensions) {
	// Calculate table width (similar to block)
	box.calculateBlockWidth(containingBlock)
	box.calculateBlockPosition(containingBlock)

	// Layout table rows
	for _, row := range box.Children {
		if row.BoxType == TableRowBox {
			row.Layout(box.Dimensions)
			box.Dimensions.Content.Height += row.marginBox().Height
		}
	}

	// If height is explicitly set, use that
	box.calculateBlockHeight()
}

// layoutTableRow lays out a table row.
// CSS 2.1 §17.5.3 Table height algorithms
func (box *LayoutBox) layoutTableRow(containingBlock Dimensions) {
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

	// Calculate number of cells and distribute width
	numCells := len(box.Children)
	if numCells == 0 {
		return
	}

	// Simple algorithm: distribute width equally among cells
	// CSS 2.1 §17.5.2.1: In the fixed table layout algorithm
	cellWidth := containingBlock.Content.Width / float64(numCells)

	// Layout each cell horizontally
	currentX := box.Dimensions.Content.X
	maxHeight := 0.0

	for _, cell := range box.Children {
		if cell.BoxType == TableCellBox {
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
}
