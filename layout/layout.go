// Package layout implements the CSS 2.1 visual formatting model.
// It converts styled nodes into a tree of layout boxes with computed dimensions.
//
// Spec references:
// - CSS 2.1 §8 Box model: https://www.w3.org/TR/CSS21/box.html
// - CSS 2.1 §9 Visual formatting model: https://www.w3.org/TR/CSS21/visuren.html
// - CSS 2.1 §10 Visual formatting model details: https://www.w3.org/TR/CSS21/visudet.html
// - CSS 2.1 §17 Tables: https://www.w3.org/TR/CSS21/tables.html
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
//
// Not yet implemented (would log warnings if encountered):
// - Floats (CSS 2.1 §9.5)
// - Positioning schemes: absolute, relative, fixed (CSS 2.1 §9.3)
// - Inline layout with line wrapping (CSS 2.1 §9.4.2 - partial)
// - Z-index and stacking contexts (CSS 2.1 §9.9)
// - Table rowspan (CSS 2.1 §17.2)
// - Border-collapse model (CSS 2.1 §17.6.2)
// - Flexbox (CSS3)
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

	// Determine box type from the display property.
	// CSS 2.1 §9.2: Display property controls box generation.
	// All default display values come from the user-agent stylesheet (style/useragent.go),
	// so the cascade handles element → display mapping per CSS 2.1 §6.4.
	boxType := BlockBox
	display := styledNode.Styles["display"]

	// CSS 2.1 §9.3: Positioning schemes - not yet implemented
	if position := styledNode.Styles["position"]; position != "" && position != "static" {
		log.Warnf("CSS 2.1 §9.3: position:%s not yet implemented (only 'static' positioning supported)", position)
	}

	// CSS 2.1 §9.5: Floats - not yet implemented
	if float := styledNode.Styles["float"]; float != "" && float != "none" {
		log.Warnf("CSS 2.1 §9.5: float:%s not yet implemented", float)
	}

	// CSS3: Flexbox/Grid - not yet implemented, fall back to block
	if display == "flex" || display == "inline-flex" {
		log.Warnf("CSS3 Flexbox: display:%s not yet implemented, treating as block", display)
		display = "block"
	}
	if display == "grid" || display == "inline-grid" {
		log.Warnf("CSS3 Grid: display:%s not yet implemented, treating as block", display)
		display = "block"
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
		box.layoutInlineBox(containingBlock)
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
	cw := containingBlock.Content.Width

	width := parseLength(styles["width"], cw)
	margin, padding, border := parseAllBoxEdges(styles, cw)

	// Calculate total width
	total := margin.Left + margin.Right + border.Left + border.Right +
		padding.Left + padding.Right + width

	// CSS 2.1 §10.3.3: over-constrained, solve for margin-right
	if width >= 0 && total > cw {
		margin.Right = cw - width - margin.Left -
			border.Left - border.Right - padding.Left - padding.Right
	}

	// If width is auto, calculate it
	if width < 0 {
		width = cw - margin.Left - margin.Right -
			border.Left - border.Right - padding.Left - padding.Right
		if width < 0 {
			width = 0
		}
	}

	box.Dimensions.Content.Width = width
	box.Dimensions.Padding = padding
	box.Dimensions.Border = border
	box.Dimensions.Margin = margin
}

// calculateBlockPosition calculates the position of a block box.
// CSS 2.1 §10.6.3 Block-level non-replaced elements in normal flow
func (box *LayoutBox) calculateBlockPosition(containingBlock Dimensions) {
	// Note: margin, padding, border edges are already fully parsed in calculateBlockWidth.
	// This method only computes the content box position.

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

// layoutInlineRun lays out a sequence of inline children horizontally with baseline alignment.
// CSS 2.1 §9.4.2 Inline formatting contexts: inline-level boxes are laid out in horizontal line boxes.
// CSS 2.1 §10.8: Line height and baseline alignment
// Returns (endX, maxHeight) - the final X position and the tallest element height.
func layoutInlineRun(children []*LayoutBox, startX, startY, containerWidth, containerX float64) (float64, float64) {
	if len(children) == 0 {
		return startX, 0
	}

	currentX := startX
	maxBaseline := 0.0
	maxHeight := 0.0

	// First pass: layout all children and find the maximum baseline
	for i, child := range children {
		inlineCB := Dimensions{
			Content: Rect{
				X:      currentX,
				Y:      startY,
				Width:  math.Max(0, containerWidth-(currentX-containerX)),
				Height: 0,
			},
		}

		child.Layout(inlineCB)

		child.Dimensions.Content.X = currentX + child.Dimensions.Margin.Left + child.Dimensions.Border.Left + child.Dimensions.Padding.Left
		child.Dimensions.Content.Y = startY + child.Dimensions.Margin.Top + child.Dimensions.Border.Top + child.Dimensions.Padding.Top

		currentX += child.marginBox().Width

		// CSS 2.1 §16.4: Add word-spacing between adjacent inline elements
		if i < len(children)-1 {
			currentX += calculateWordSpacing(child)
		}

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
		baselineOffset := maxBaseline - getBaseline(child)
		if baselineOffset > 0 {
			child.shiftY(baselineOffset)
		}
	}

	return currentX, maxHeight
}

// layoutInlineChildren lays out a run of inline-level children within this block box.
// CSS 2.1 §9.4.2 Inline formatting contexts
func (box *LayoutBox) layoutInlineChildren(children []*LayoutBox) {
	if len(children) == 0 {
		return
	}

	_, maxHeight := layoutInlineRun(children,
		box.Dimensions.Content.X,
		box.Dimensions.Content.Y+box.Dimensions.Content.Height,
		box.Dimensions.Content.Width, box.Dimensions.Content.X)

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
	// Parse box model edges
	margin, padding, border := parseAllBoxEdges(box.StyledNode.Styles, containingBlock.Content.Width)
	box.Dimensions.Margin = margin
	box.Dimensions.Padding = padding
	box.Dimensions.Border = border

	// Position content relative to containing block
	box.Dimensions.Content.X = containingBlock.Content.X +
		margin.Left + border.Left + padding.Left
	box.Dimensions.Content.Y = containingBlock.Content.Y +
		margin.Top + border.Top + padding.Top

	// Layout inline children using shared logic
	endX, maxHeight := layoutInlineRun(box.Children,
		box.Dimensions.Content.X, box.Dimensions.Content.Y,
		containingBlock.Content.Width, containingBlock.Content.X)

	box.Dimensions.Content.Width = endX - box.Dimensions.Content.X
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

// parseEdges parses all four edges (top, right, bottom, left) for a CSS box model property.
// CSS 2.1 §8.1: The box model defines margin, padding, and border edges.
// prefix should be "margin", "padding", or "border" (with "-width" suffix for borders).
func parseEdges(styles map[string]string, prefix string, referenceWidth float64) EdgeSizes {
	suffix := ""
	if prefix == "border" {
		prefix = "border"
		suffix = "-width"
	}
	return EdgeSizes{
		Top:    parseLengthOr0(styles[prefix+"-top"+suffix], referenceWidth),
		Right:  parseLengthOr0(styles[prefix+"-right"+suffix], referenceWidth),
		Bottom: parseLengthOr0(styles[prefix+"-bottom"+suffix], referenceWidth),
		Left:   parseLengthOr0(styles[prefix+"-left"+suffix], referenceWidth),
	}
}

// parseAllBoxEdges parses margin, padding, and border edges from styles.
// CSS 2.1 §8.1: Box dimensions
func parseAllBoxEdges(styles map[string]string, referenceWidth float64) (margin, padding, border EdgeSizes) {
	margin = parseEdges(styles, "margin", referenceWidth)
	padding = parseEdges(styles, "padding", referenceWidth)
	border = parseEdges(styles, "border", referenceWidth)
	return
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

	// Get font style from CSS styles (CSS 2.1 §15 Fonts)
	// Uses shared font.ExtractStyle to ensure layout and rendering agree
	fontStyle := font.ExtractStyle(box.StyledNode.Styles)
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

// collapseWhitespace delegates to css.CollapseWhitespace.
// CSS 2.1 §16.6.1: The white-space property
func collapseWhitespace(text string) string {
	return css.CollapseWhitespace(text)
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

// layoutTableRow lays out a table row with equal-width columns.
// CSS 2.1 §17.5.3 Table height algorithms
// CSS 2.1 §17.5.2.1: Fixed table layout uses equal column widths as fallback
func (box *LayoutBox) layoutTableRow(containingBlock Dimensions) {
	numColumns := 0
	for _, cell := range box.Children {
		if cell.BoxType == TableCellBox {
			numColumns += getColspan(cell)
		}
	}
	if numColumns == 0 {
		numColumns = len(box.Children)
	}
	if numColumns == 0 {
		numColumns = 1
	}

	// Build equal-width column array and delegate to layoutWithColumnWidths
	columnWidth := containingBlock.Content.Width / float64(numColumns)
	columnWidths := make([]float64, numColumns)
	for i := range columnWidths {
		columnWidths[i] = columnWidth
	}
	box.layoutWithColumnWidths(containingBlock, columnWidths, 0)
}

// layoutWithColumnWidths lays out a table row with pre-calculated column widths.
// CSS 2.1 §17.5.2.2: Auto table layout
func (box *LayoutBox) layoutWithColumnWidths(containingBlock Dimensions, columnWidths []float64, borderSpacing float64) {
	styles := box.StyledNode.Styles

	// Parse vertical box model edges for row positioning
	margin, padding, border := parseAllBoxEdges(styles, containingBlock.Content.Width)
	box.Dimensions.Margin = margin
	box.Dimensions.Padding = padding
	box.Dimensions.Border = border

	// Position row
	box.Dimensions.Content.X = containingBlock.Content.X
	box.Dimensions.Content.Y = containingBlock.Content.Y + containingBlock.Content.Height +
		margin.Top + border.Top + padding.Top
	box.Dimensions.Content.Width = containingBlock.Content.Width

	// Handle empty rows (e.g., spacer rows)
	if len(box.Children) == 0 {
		applyExplicitRowHeight(box, styles)
		return
	}

	// Layout each cell horizontally using calculated widths
	// CSS 2.1 §17.6.1: border-spacing adds space between cells
	currentX := box.Dimensions.Content.X + borderSpacing
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

			cellContainingBlock := Dimensions{
				Content: Rect{
					X:      currentX,
					Y:      box.Dimensions.Content.Y,
					Width:  cellWidth,
					Height: 0,
				},
			}
			cell.Layout(cellContainingBlock)

			currentX += cell.marginBox().Width + borderSpacing
			currentCol += colspan

			if cell.marginBox().Height > maxHeight {
				maxHeight = cell.marginBox().Height
			}
		}
	}

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
	cw := containingBlock.Content.Width

	// Parse width - if specified, use it; otherwise use the width from containing block
	width := parseLength(styles["width"], cw)
	if width < 0 {
		width = cw
	}

	// Parse all box model edges
	margin, padding, border := parseAllBoxEdges(styles, cw)
	box.Dimensions.Margin = margin
	box.Dimensions.Padding = padding
	box.Dimensions.Border = border

	// Calculate content width
	contentWidth := width - padding.Left - padding.Right - border.Left - border.Right - margin.Left - margin.Right
	if contentWidth < 0 {
		contentWidth = 0
	}

	box.Dimensions.Content.Width = contentWidth
	box.Dimensions.Content.X = containingBlock.Content.X + margin.Left + border.Left + padding.Left
	box.Dimensions.Content.Y = containingBlock.Content.Y + margin.Top + border.Top + padding.Top

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

