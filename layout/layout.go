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

// Table layout constants
const (
	// baseFontHeight is the height in pixels of basicfont.Face7x13
	// This is used for text dimension calculations throughout the layout engine
	// CSS 2.1 §15.7: The default 'medium' font size is typically 16px, but we use
	// 13px to match the available basicfont.Face7x13 from golang.org/x/image/font/basicfont
	baseFontHeight = 13.0
	
	// maxColumnWidth is the maximum width any table column can have.
	// This prevents extremely wide content from creating unusable layouts.
	// CSS 2.1 §17.5.2.2 does not specify a maximum, but practical implementations
	// need limits to prevent performance issues. Set to 400px as a reasonable maximum.
	maxColumnWidth = 400.0
	
	// maxColspan is the maximum number of columns a cell can span.
	// HTML5 §4.9.9 specifies that user agents may choose to limit colspan
	// to prevent denial of service attacks. The recommended maximum is 1000.
	// See: https://html.spec.whatwg.org/multipage/tables.html#attributes-common-to-td-and-th-elements
	maxColspan = 1000
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

	// CSS 2.1 §16.6.1: Collapse whitespace for layout calculations
	// This ensures dimensions match what will actually be rendered
	text = collapseWhitespace(text)
	
	if text == "" {
		box.Dimensions.Content.Width = 0
		box.Dimensions.Content.Height = 0
		return
	}

	// Calculate text dimensions using basicfont.Face7x13 as base
	// Note: For basicfont.Face7x13, all characters have fixed width (Advance)
	// For more accurate measurement, we could use font.Drawer.MeasureString()
	// but basicfont is monospaced so character count * Advance is accurate
	face := basicfont.Face7x13
	
	// Get font size from styles (CSS 2.1 §15.7)
	fontSize := extractFontSize(box.StyledNode.Styles)
	scale := fontSize / baseFontHeight
	
	width := float64(len(text)*face.Advance) * scale
	height := float64(face.Height) * scale

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
		return baseFontHeight // Default font size
	}
	
	fontSize = strings.TrimSpace(strings.ToLower(fontSize))
	
	// Handle pixel values (e.g., "14px")
	if strings.HasSuffix(fontSize, "px") {
		fontSize = strings.TrimSuffix(fontSize, "px")
		if size, err := strconv.ParseFloat(fontSize, 64); err == nil && size > 0 {
			return size
		}
	}
	
	// Handle plain numbers (treat as pixels)
	if size, err := strconv.ParseFloat(fontSize, 64); err == nil && size > 0 {
		return size
	}
	
	// Handle named sizes (CSS 2.1 §15.7)
	namedSizes := map[string]float64{
		"xx-small": 9.0,
		"x-small":  10.0,
		"small":    12.0,
		"medium":   baseFontHeight,
		"large":    16.0,
		"x-large":  20.0,
		"xx-large": 24.0,
	}
	
	if size, ok := namedSizes[fontSize]; ok {
		return size
	}
	
	return baseFontHeight // Default font size
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
func (box *LayoutBox) layoutTable(containingBlock Dimensions) {
	// Calculate table width (similar to block)
	box.calculateBlockWidth(containingBlock)
	box.calculateBlockPosition(containingBlock)

	// Calculate the number of columns in the table
	// CSS 2.1 §17.2.1: The number of columns is determined by examining all rows
	numColumns := box.calculateTableColumns()
	
	// Calculate column widths based on content
	// CSS 2.1 §17.5.2.2: Auto table layout
	columnWidths := box.calculateColumnWidths(numColumns, box.Dimensions.Content.Width)

	// Layout table rows with column widths
	for _, row := range box.Children {
		if row.BoxType == TableRowBox {
			row.layoutWithColumnWidths(box.Dimensions, columnWidths)
			box.Dimensions.Content.Height += row.marginBox().Height
		}
	}

	// If height is explicitly set, use that
	box.calculateBlockHeight()
}

// calculateTableColumns calculates the number of columns in a table.
// CSS 2.1 §17.2.1: The table column count is determined by examining all rows.
// This implementation counts columns based on actual table cells and their colspan attributes.
// Note: This is a simplified implementation that doesn't support column groups or explicit
// column specifications via <col> elements, which are part of the full CSS 2.1 spec.
func (box *LayoutBox) calculateTableColumns() int {
	maxColumns := 0

	// Examine each row to find the maximum column count
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

	// Default to 1 if no columns found
	if maxColumns == 0 {
		maxColumns = 1
	}

	return maxColumns
}

// getColspan extracts the colspan attribute from a table cell.
// Returns 1 if no colspan attribute is present or if the value is invalid.
// CSS 2.1 §17.2.1: The colspan attribute specifies the number of columns spanned by a cell
// HTML5 recommends a maximum colspan of 1000 to prevent performance issues
func getColspan(cell *LayoutBox) int {
	if cell.StyledNode == nil || cell.StyledNode.Node == nil {
		return 1
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
// This implements a simplified auto table layout algorithm
// CSS 2.1 §17.5.2.2: Auto table layout
func (box *LayoutBox) calculateColumnWidths(numColumns int, tableWidth float64) []float64 {
	// Collect column content sizes from all rows
	columnMinWidths := make([]float64, numColumns)
	
	// Examine each row to estimate minimum column widths
	for _, row := range box.Children {
		if row.BoxType == TableRowBox {
			colIndex := 0
			for _, cell := range row.Children {
				if cell.BoxType == TableCellBox {
					colspan := getColspan(cell)
					
					// Estimate content width based on text content
					minWidth := box.estimateCellMinWidth(cell)
					
					// For single-column cells, update the column minimum
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
	
	// Calculate total minimum width
	totalMinWidth := 0.0
	for _, w := range columnMinWidths {
		totalMinWidth += w
	}
	
	// Allocate widths proportionally, ensuring minimums are met
	columnWidths := make([]float64, numColumns)
	if totalMinWidth > 0 && totalMinWidth < tableWidth {
		// Distribute extra space proportionally
		scale := tableWidth / totalMinWidth
		for i := range columnWidths {
			columnWidths[i] = columnMinWidths[i] * scale
		}
	} else if totalMinWidth > 0 {
		// Use minimum widths if they exceed table width
		for i := range columnWidths {
			columnWidths[i] = columnMinWidths[i]
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
	
	// Recursively estimate width from children
	for _, child := range layoutBox.Children {
		if child.StyledNode != nil && child.StyledNode.Node != nil {
			if child.StyledNode.Node.Type == dom.TextNode {
				// Estimate text width accounting for font size and whitespace collapsing
				// CSS 2.1 §16.6.1: Collapse whitespace for width calculations
				text := collapseWhitespace(child.StyledNode.Node.Data)
				fontSize := extractFontSize(child.StyledNode.Styles)
				scale := fontSize / baseFontHeight
				textWidth := float64(len(text)) * charWidth * scale
				width += textWidth
			} else {
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
func (box *LayoutBox) layoutWithColumnWidths(containingBlock Dimensions, columnWidths []float64) {
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

	if len(box.Children) == 0 {
		return
	}

	// Layout each cell horizontally using calculated widths
	currentX := box.Dimensions.Content.X
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
			currentX += cell.marginBox().Width
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

	if len(box.Children) == 0 || numColumns == 0 {
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
	
	// Adjust X position of all children
	for _, child := range box.Children {
		child.Dimensions.Content.X += offset
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
	
	// Adjust Y position of all children
	for _, child := range box.Children {
		child.Dimensions.Content.Y += offset
	}
}
