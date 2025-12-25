package render

import (
	"bytes"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
	"strconv"
	"strings"

	"github.com/lukehoban/browser/dom"
	"github.com/lukehoban/browser/layout"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Font rendering constants
const (
	// CSS 2.1 §15.7: Default 'medium' font size (using basicfont.Face7x13)
	baseFontHeight = 13.0
	
	// CSS 2.1 §15.7: Slant factor for synthetic italic rendering
	italicSlantFactor = 0.2
	
	// CSS 2.1 §16.3.1: Distance below baseline for underlines
	underlineOffset = 2.0
)

// defaultFontStyle represents CSS 2.1 initial values for font properties
var defaultFontStyle = FontStyle{
	Size:       baseFontHeight,
	Weight:     "normal",
	Style:      "normal",
	Decoration: "none",
}

// Canvas represents the rendering surface.
type Canvas struct {
	Width      int
	Height     int
	Pixels     []color.RGBA
	ImageCache map[string]image.Image // Cache for loaded images
}

// NewCanvas creates a new canvas with the given dimensions.
func NewCanvas(width, height int) *Canvas {
	return &Canvas{
		Width:      width,
		Height:     height,
		Pixels:     make([]color.RGBA, width*height),
		ImageCache: make(map[string]image.Image),
	}
}

// Clear fills the canvas with a background color.
func (c *Canvas) Clear(bg color.RGBA) {
	for i := range c.Pixels {
		c.Pixels[i] = bg
	}
}

// SetPixel sets a pixel at the given coordinates.
func (c *Canvas) SetPixel(x, y int, col color.RGBA) {
	if x >= 0 && x < c.Width && y >= 0 && y < c.Height {
		c.Pixels[y*c.Width+x] = col
	}
}

// FillRect fills a rectangle with the given color.
// CSS 2.1 §14.2 The background
func (c *Canvas) FillRect(x, y, width, height int, col color.RGBA) {
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			c.SetPixel(x+dx, y+dy, col)
		}
	}
}

// DrawRect draws a rectangle outline with the given color and thickness.
// CSS 2.1 §8.5 Border properties
func (c *Canvas) DrawRect(x, y, width, height int, col color.RGBA, thickness int) {
	// Top border
	c.FillRect(x, y, width, thickness, col)
	// Bottom border
	c.FillRect(x, y+height-thickness, width, thickness, col)
	// Left border
	c.FillRect(x, y, thickness, height, col)
	// Right border
	c.FillRect(x+width-thickness, y, thickness, height, col)
}

// FontStyle represents text rendering options.
// CSS 2.1 §16 Text and §15 Fonts
type FontStyle struct {
	Size           float64 // Font size in pixels
	Weight         string  // Font weight: "normal" or "bold"
	Style          string  // Font style: "normal" or "italic"
	Decoration     string  // Text decoration: "none" or "underline"
}

// DrawText draws text at the given position with the given color.
// CSS 2.1 §16 Text
// Uses default font style for backward compatibility with code that doesn't specify font properties.
func (c *Canvas) DrawText(text string, x, y int, col color.RGBA) {
	c.DrawStyledText(text, x, y, col, defaultFontStyle)
}

// DrawStyledText draws text with font styling at the given position.
// CSS 2.1 §15 Fonts and §16 Text
// CSS 2.1 §15.6 Font boldness: font-weight
// CSS 2.1 §15.7 Font style: font-style
// CSS 2.1 §16.3.1 Underlining, overlining, striking: text-decoration
func (c *Canvas) DrawStyledText(text string, x, y int, col color.RGBA, style FontStyle) {
	baseFace := basicfont.Face7x13 // 7x13 pixel font
	
	// Calculate scale factor based on desired font size
	scale := style.Size / baseFontHeight
	if scale <= 0 {
		scale = 1.0
	}
	
	// Calculate the dimensions for scaled text
	baseWidth := len(text) * baseFace.Advance
	baseHeight := baseFace.Height
	
	scaledWidth := int(float64(baseWidth) * scale)
	scaledHeight := int(float64(baseHeight) * scale)
	
	// Create temporary image for the base text
	baseImg := image.NewRGBA(image.Rect(0, 0, baseWidth, baseHeight))
	
	drawer := &font.Drawer{
		Dst:  baseImg,
		Src:  image.NewUniform(col),
		Face: baseFace,
		Dot:  fixed.Point26_6{X: 0, Y: fixed.I(baseFace.Ascent)},
	}
	
	drawer.DrawString(text)
	
	// CSS 2.1 §15.6: Apply bold effect by drawing with offset
	if style.Weight == "bold" {
		drawer.Dot = fixed.Point26_6{X: fixed.I(1), Y: fixed.I(baseFace.Ascent)}
		drawer.DrawString(text)
	}
	
	// Scale image if needed
	var textImg *image.RGBA
	if scale != 1.0 {
		textImg = scaleImage(baseImg, scaledWidth, scaledHeight)
	} else {
		textImg = baseImg
	}
	
	// Calculate adjusted baseline offset for scaled text
	baselineOffset := int(float64(baseFace.Ascent) * scale)
	
	// Copy text pixels to canvas with italic slant if needed
	for dy := 0; dy < scaledHeight; dy++ {
		for dx := 0; dx < scaledWidth; dx++ {
			// CSS 2.1 §15.7: Apply italic slant
			slant := 0
			if style.Style == "italic" {
				slant = int(float64(scaledHeight-dy) * italicSlantFactor)
			}
			
			px := x + dx + slant
			py := y - baselineOffset + dy
			
			if px >= 0 && px < c.Width && py >= 0 && py < c.Height {
				r, g, b, a := textImg.At(dx, dy).RGBA()
				if a > 0 {
					c.Pixels[py*c.Width+px] = color.RGBA{
						R: uint8(r >> 8),
						G: uint8(g >> 8),
						B: uint8(b >> 8),
						A: uint8(a >> 8),
					}
				}
			}
		}
	}
	
	// CSS 2.1 §16.3.1: Draw underline if needed
	if style.Decoration == "underline" {
		underlineY := y + int(underlineOffset*scale)
		underlineThickness := max(1, int(scale*0.5))
		c.FillRect(x, underlineY, scaledWidth, underlineThickness, col)
	}
}

// scaleImage scales an image using nearest-neighbor interpolation.
func scaleImage(src *image.RGBA, newWidth, newHeight int) *image.RGBA {
	if newWidth <= 0 || newHeight <= 0 {
		return src
	}
	
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	bounds := src.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()
	
	for dy := 0; dy < newHeight; dy++ {
		for dx := 0; dx < newWidth; dx++ {
			// Map destination pixel to source pixel
			srcX := bounds.Min.X + (dx * srcWidth / newWidth)
			srcY := bounds.Min.Y + (dy * srcHeight / newHeight)
			dst.Set(dx, dy, src.At(srcX, srcY))
		}
	}
	
	return dst
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// DrawImage draws an image onto the canvas at the specified position.
// HTML5 §4.8.2 The img element
func (c *Canvas) DrawImage(img image.Image, x, y, width, height int) {
	// Validate dimensions to prevent division by zero
	if width <= 0 || height <= 0 {
		return
	}

	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	// Draw the image with simple nearest-neighbor scaling
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			// Map destination pixel to source pixel
			srcX := bounds.Min.X + (dx * srcWidth / width)
			srcY := bounds.Min.Y + (dy * srcHeight / height)

			// Get source color and convert to RGBA
			col := img.At(srcX, srcY)
			r, g, b, a := col.RGBA()

			// Convert from 16-bit to 8-bit color
			rgba := color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			}

			// Handle alpha blending with existing pixel
			destX := x + dx
			destY := y + dy

			if rgba.A == 255 {
				c.SetPixel(destX, destY, rgba)
			} else if rgba.A > 0 {
				// Check bounds before accessing pixel array directly
				if destX >= 0 && destX < c.Width && destY >= 0 && destY < c.Height {
					// Simple alpha blending
					existing := c.Pixels[destY*c.Width+destX]
					alpha := float64(rgba.A) / 255.0
					blended := color.RGBA{
						R: uint8(float64(rgba.R)*alpha + float64(existing.R)*(1-alpha)),
						G: uint8(float64(rgba.G)*alpha + float64(existing.G)*(1-alpha)),
						B: uint8(float64(rgba.B)*alpha + float64(existing.B)*(1-alpha)),
						A: 255,
					}
					c.SetPixel(destX, destY, blended)
				}
			}
		}
	}
}

// LoadImage loads an image from an absolute file path or URL.
// Supports PNG, JPEG, and GIF formats.
// The path should be already resolved (absolute) before calling this method.
//
// Uses dom.ResourceLoader for consistent resource fetching across the browser.
func (c *Canvas) LoadImage(path string) (image.Image, error) {
	// Check cache first
	if img, ok := c.ImageCache[path]; ok {
		return img, nil
	}

	// Use DOM resource loader to fetch the image data
	loader := dom.NewResourceLoader("")
	data, err := loader.LoadResource(path)
	if err != nil {
		return nil, err
	}

	// Decode the image from the loaded data
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	// Cache the image
	c.ImageCache[path] = img

	return img, nil
}

// ToImage converts the canvas to an image.Image.
func (c *Canvas) ToImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			img.Set(x, y, c.Pixels[y*c.Width+x])
		}
	}
	return img
}

// SavePNG saves the canvas as a PNG file.
func (c *Canvas) SavePNG(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	if err := png.Encode(file, c.ToImage()); err != nil {
		_ = file.Close()
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return nil
}

// Render renders a layout tree to a canvas.
func Render(root *layout.LayoutBox, width, height int) *Canvas {
	canvas := NewCanvas(width, height)
	// Default white background
	canvas.Clear(color.RGBA{255, 255, 255, 255})

	renderLayoutBox(canvas, root)

	return canvas
}

// renderLayoutBox renders a single layout box and its children.
func renderLayoutBox(canvas *Canvas, box *layout.LayoutBox) {
	renderBackground(canvas, box)
	renderBorders(canvas, box)
	renderText(canvas, box)
	renderImage(canvas, box)

	for _, child := range box.Children {
		renderLayoutBox(canvas, child)
	}
}

// renderBackground renders the background of a layout box.
// CSS 2.1 §14.2 The background
func renderBackground(canvas *Canvas, box *layout.LayoutBox) {
	if box.StyledNode == nil {
		return
	}

	bg := box.StyledNode.Styles["background"]
	if bg == "" {
		bg = box.StyledNode.Styles["background-color"]
	}

	if bg != "" && bg != "transparent" && bg != "none" {
		bgColor := parseColor(bg)
		borderBox := box.Dimensions.Content
		borderBox.X -= box.Dimensions.Padding.Left
		borderBox.Y -= box.Dimensions.Padding.Top
		borderBox.Width += box.Dimensions.Padding.Left + box.Dimensions.Padding.Right
		borderBox.Height += box.Dimensions.Padding.Top + box.Dimensions.Padding.Bottom

		canvas.FillRect(
			int(borderBox.X),
			int(borderBox.Y),
			int(borderBox.Width),
			int(borderBox.Height),
			bgColor,
		)
	}
}

// renderBorders renders the borders of a layout box.
// CSS 2.1 §8.5 Border properties
// CSS 2.1 §8.5.3: Borders only render if border-style is not 'none' or absent
func renderBorders(canvas *Canvas, box *layout.LayoutBox) {
	if box.StyledNode == nil {
		return
	}

	styles := box.StyledNode.Styles

	// CSS 2.1 §8.5.3: Check if border-style is set and not 'none'
	borderStyle := styles["border-style"]
	if borderStyle == "" || borderStyle == "none" {
		return
	}

	borderColor := parseColor(styles["border-color"])

	// Get the padding box coordinates (borders are drawn around it)
	paddingBox := box.Dimensions.Content
	paddingBox.X -= box.Dimensions.Padding.Left
	paddingBox.Y -= box.Dimensions.Padding.Top
	paddingBox.Width += box.Dimensions.Padding.Left + box.Dimensions.Padding.Right
	paddingBox.Height += box.Dimensions.Padding.Top + box.Dimensions.Padding.Bottom

	// Draw top border
	if box.Dimensions.Border.Top > 0 {
		canvas.FillRect(
			int(paddingBox.X-box.Dimensions.Border.Left),
			int(paddingBox.Y-box.Dimensions.Border.Top),
			int(paddingBox.Width+box.Dimensions.Border.Left+box.Dimensions.Border.Right),
			int(box.Dimensions.Border.Top),
			borderColor,
		)
	}

	// Draw bottom border
	if box.Dimensions.Border.Bottom > 0 {
		canvas.FillRect(
			int(paddingBox.X-box.Dimensions.Border.Left),
			int(paddingBox.Y+paddingBox.Height),
			int(paddingBox.Width+box.Dimensions.Border.Left+box.Dimensions.Border.Right),
			int(box.Dimensions.Border.Bottom),
			borderColor,
		)
	}

	// Draw left border
	if box.Dimensions.Border.Left > 0 {
		canvas.FillRect(
			int(paddingBox.X-box.Dimensions.Border.Left),
			int(paddingBox.Y-box.Dimensions.Border.Top),
			int(box.Dimensions.Border.Left),
			int(paddingBox.Height+box.Dimensions.Border.Top+box.Dimensions.Border.Bottom),
			borderColor,
		)
	}

	// Draw right border
	if box.Dimensions.Border.Right > 0 {
		canvas.FillRect(
			int(paddingBox.X+paddingBox.Width),
			int(paddingBox.Y-box.Dimensions.Border.Top),
			int(box.Dimensions.Border.Right),
			int(paddingBox.Height+box.Dimensions.Border.Top+box.Dimensions.Border.Bottom),
			borderColor,
		)
	}
}

// renderText renders the text content of a layout box.
// CSS 2.1 §16 Text
func renderText(canvas *Canvas, box *layout.LayoutBox) {
	if box.StyledNode == nil || box.StyledNode.Node == nil {
		return
	}

	// Only render text for text nodes
	if box.StyledNode.Node.Type != dom.TextNode {
		return
	}

	// Get the text content
	text := box.StyledNode.Node.Data
	if text == "" {
		return
	}

	// CSS 2.1 §16.6.1: Whitespace processing
	// Collapse sequences of whitespace (spaces, tabs, newlines) into a single space
	// This is the default behavior for normal text (not pre-formatted)
	text = collapseWhitespace(text)
	
	if text == "" {
		return
	}

	// Get text color from styles (default to black)
	textColor := parseColor(box.StyledNode.Styles["color"])
	if textColor == (color.RGBA{0, 0, 0, 0}) {
		textColor = color.RGBA{0, 0, 0, 255} // Default to black
	}

	// Extract font properties from styles
	fontStyle := extractFontStyle(box.StyledNode.Styles)
	
	// Render the text at the box's position
	// Add a vertical offset to position text at baseline
	x := int(box.Dimensions.Content.X)
	y := int(box.Dimensions.Content.Y) + int(fontStyle.Size)

	canvas.DrawStyledText(text, x, y, textColor, fontStyle)
}

// extractFontStyle extracts font styling properties from CSS styles.
// CSS 2.1 §15 Fonts and §16 Text
func extractFontStyle(styles map[string]string) FontStyle {
	fontStyle := FontStyle{
		Size:       13.0, // Default font size
		Weight:     "normal",
		Style:      "normal",
		Decoration: "none",
	}
	
	// Parse font-size (CSS 2.1 §15.7)
	if fontSize := styles["font-size"]; fontSize != "" {
		if size := parseFontSize(fontSize); size > 0 {
			fontStyle.Size = size
		}
	}
	
	// CSS 2.1 §15.6: Parse font-weight
	if fontWeight := styles["font-weight"]; fontWeight != "" {
		fontWeight = strings.TrimSpace(strings.ToLower(fontWeight))
		if fontWeight == "bold" || fontWeight == "bolder" {
			fontStyle.Weight = "bold"
		} else if weight, err := strconv.Atoi(fontWeight); err == nil && weight >= 600 {
			fontStyle.Weight = "bold"
		}
	}
	
	// CSS 2.1 §15.7: Parse font-style
	if fontStyleVal := styles["font-style"]; fontStyleVal != "" {
		fontStyleVal = strings.TrimSpace(strings.ToLower(fontStyleVal))
		if fontStyleVal == "italic" || fontStyleVal == "oblique" {
			fontStyle.Style = "italic"
		}
	}
	
	// CSS 2.1 §16.3.1: Parse text-decoration
	if textDecoration := styles["text-decoration"]; textDecoration != "" {
		textDecoration = strings.TrimSpace(strings.ToLower(textDecoration))
		if strings.Contains(textDecoration, "underline") {
			fontStyle.Decoration = "underline"
		}
	}
	
	return fontStyle
}

// parseFontSize parses a CSS font-size value and returns the size in pixels.
// CSS 2.1 §15.7 Font size: the 'font-size' property
func parseFontSize(value string) float64 {
	value = strings.TrimSpace(strings.ToLower(value))
	
	if strings.HasSuffix(value, "px") {
		value = strings.TrimSuffix(value, "px")
		if size, err := strconv.ParseFloat(value, 64); err == nil {
			return size
		}
	}
	
	if size, err := strconv.ParseFloat(value, 64); err == nil {
		return size
	}
	
	// CSS 2.1 §15.7: Named sizes
	namedSizes := map[string]float64{
		"xx-small": 9.0,
		"x-small":  10.0,
		"small":    12.0,
		"medium":   13.0,
		"large":    16.0,
		"x-large":  20.0,
		"xx-large": 24.0,
	}
	
	if size, ok := namedSizes[value]; ok {
		return size
	}
	
	return 0
}

// parseColor parses a CSS color value and returns a color.RGBA.
// CSS 2.1 §4.3.6: Supports basic color names and hex colors
// Extended color keywords from CSS Color Module Level 3
func parseColor(value string) color.RGBA {
	value = strings.TrimSpace(strings.ToLower(value))

	// CSS 2.1 §4.3.6: Named colors (16 basic colors)
	// CSS Color Module Level 3: Extended color keywords (147 colors)
	namedColors := map[string]color.RGBA{
		// CSS 2.1 Basic 16 colors (SVG 1.0 color keywords)
		"black":   {0, 0, 0, 255},
		"silver":  {192, 192, 192, 255},
		"gray":    {128, 128, 128, 255},
		"grey":    {128, 128, 128, 255},
		"white":   {255, 255, 255, 255},
		"maroon":  {128, 0, 0, 255},
		"red":     {255, 0, 0, 255},
		"purple":  {128, 0, 128, 255},
		"fuchsia": {255, 0, 255, 255},
		"magenta": {255, 0, 255, 255},
		"green":   {0, 128, 0, 255},
		"lime":    {0, 255, 0, 255},
		"olive":   {128, 128, 0, 255},
		"yellow":  {255, 255, 0, 255},
		"navy":    {0, 0, 128, 255},
		"blue":    {0, 0, 255, 255},
		"teal":    {0, 128, 128, 255},
		"aqua":    {0, 255, 255, 255},
		"cyan":    {0, 255, 255, 255},
		"orange":  {255, 165, 0, 255},
		
		// Extended colors commonly used in web pages
		"lightgray":      {211, 211, 211, 255},
		"lightgrey":      {211, 211, 211, 255},
		"darkgray":       {169, 169, 169, 255},
		"darkgrey":       {169, 169, 169, 255},
		"dimgray":        {105, 105, 105, 255},
		"dimgrey":        {105, 105, 105, 255},
		"lightslategray": {119, 136, 153, 255},
		"lightslategrey": {119, 136, 153, 255},
		"slategray":      {112, 128, 144, 255},
		"slategrey":      {112, 128, 144, 255},
		"darkslategray":  {47, 79, 79, 255},
		"darkslategrey":  {47, 79, 79, 255},
		
		"aliceblue":            {240, 248, 255, 255},
		"antiquewhite":         {250, 235, 215, 255},
		"aquamarine":           {127, 255, 212, 255},
		"azure":                {240, 255, 255, 255},
		"beige":                {245, 245, 220, 255},
		"bisque":               {255, 228, 196, 255},
		"blanchedalmond":       {255, 235, 205, 255},
		"blueviolet":           {138, 43, 226, 255},
		"brown":                {165, 42, 42, 255},
		"burlywood":            {222, 184, 135, 255},
		"cadetblue":            {95, 158, 160, 255},
		"chartreuse":           {127, 255, 0, 255},
		"chocolate":            {210, 105, 30, 255},
		"coral":                {255, 127, 80, 255},
		"cornflowerblue":       {100, 149, 237, 255},
		"cornsilk":             {255, 248, 220, 255},
		"crimson":              {220, 20, 60, 255},
		"darkblue":             {0, 0, 139, 255},
		"darkcyan":             {0, 139, 139, 255},
		"darkgoldenrod":        {184, 134, 11, 255},
		"darkgreen":            {0, 100, 0, 255},
		"darkkhaki":            {189, 183, 107, 255},
		"darkmagenta":          {139, 0, 139, 255},
		"darkolivegreen":       {85, 107, 47, 255},
		"darkorange":           {255, 140, 0, 255},
		"darkorchid":           {153, 50, 204, 255},
		"darkred":              {139, 0, 0, 255},
		"darksalmon":           {233, 150, 122, 255},
		"darkseagreen":         {143, 188, 143, 255},
		"darkslateblue":        {72, 61, 139, 255},
		"darkturquoise":        {0, 206, 209, 255},
		"darkviolet":           {148, 0, 211, 255},
		"deeppink":             {255, 20, 147, 255},
		"deepskyblue":          {0, 191, 255, 255},
		"dodgerblue":           {30, 144, 255, 255},
		"firebrick":            {178, 34, 34, 255},
		"floralwhite":          {255, 250, 240, 255},
		"forestgreen":          {34, 139, 34, 255},
		"gainsboro":            {220, 220, 220, 255},
		"ghostwhite":           {248, 248, 255, 255},
		"gold":                 {255, 215, 0, 255},
		"goldenrod":            {218, 165, 32, 255},
		"greenyellow":          {173, 255, 47, 255},
		"honeydew":             {240, 255, 240, 255},
		"hotpink":              {255, 105, 180, 255},
		"indianred":            {205, 92, 92, 255},
		"indigo":               {75, 0, 130, 255},
		"ivory":                {255, 255, 240, 255},
		"khaki":                {240, 230, 140, 255},
		"lavender":             {230, 230, 250, 255},
		"lavenderblush":        {255, 240, 245, 255},
		"lawngreen":            {124, 252, 0, 255},
		"lemonchiffon":         {255, 250, 205, 255},
		"lightblue":            {173, 216, 230, 255},
		"lightcoral":           {240, 128, 128, 255},
		"lightcyan":            {224, 255, 255, 255},
		"lightgoldenrodyellow": {250, 250, 210, 255},
		"lightgreen":           {144, 238, 144, 255},
		"lightpink":            {255, 182, 193, 255},
		"lightsalmon":          {255, 160, 122, 255},
		"lightseagreen":        {32, 178, 170, 255},
		"lightskyblue":         {135, 206, 250, 255},
		"lightsteelblue":       {176, 196, 222, 255},
		"lightyellow":          {255, 255, 224, 255},
		"limegreen":            {50, 205, 50, 255},
		"linen":                {250, 240, 230, 255},
		"mediumaquamarine":     {102, 205, 170, 255},
		"mediumblue":           {0, 0, 205, 255},
		"mediumorchid":         {186, 85, 211, 255},
		"mediumpurple":         {147, 112, 219, 255},
		"mediumseagreen":       {60, 179, 113, 255},
		"mediumslateblue":      {123, 104, 238, 255},
		"mediumspringgreen":    {0, 250, 154, 255},
		"mediumturquoise":      {72, 209, 204, 255},
		"mediumvioletred":      {199, 21, 133, 255},
		"midnightblue":         {25, 25, 112, 255},
		"mintcream":            {245, 255, 250, 255},
		"mistyrose":            {255, 228, 225, 255},
		"moccasin":             {255, 228, 181, 255},
		"navajowhite":          {255, 222, 173, 255},
		"oldlace":              {253, 245, 230, 255},
		"olivedrab":            {107, 142, 35, 255},
		"orangered":            {255, 69, 0, 255},
		"orchid":               {218, 112, 214, 255},
		"palegoldenrod":        {238, 232, 170, 255},
		"palegreen":            {152, 251, 152, 255},
		"paleturquoise":        {175, 238, 238, 255},
		"palevioletred":        {219, 112, 147, 255},
		"papayawhip":           {255, 239, 213, 255},
		"peachpuff":            {255, 218, 185, 255},
		"peru":                 {205, 133, 63, 255},
		"pink":                 {255, 192, 203, 255},
		"plum":                 {221, 160, 221, 255},
		"powderblue":           {176, 224, 230, 255},
		"rosybrown":            {188, 143, 143, 255},
		"royalblue":            {65, 105, 225, 255},
		"saddlebrown":          {139, 69, 19, 255},
		"salmon":               {250, 128, 114, 255},
		"sandybrown":           {244, 164, 96, 255},
		"seagreen":             {46, 139, 87, 255},
		"seashell":             {255, 245, 238, 255},
		"sienna":               {160, 82, 45, 255},
		"skyblue":              {135, 206, 235, 255},
		"slateblue":            {106, 90, 205, 255},
		"snow":                 {255, 250, 250, 255},
		"springgreen":          {0, 255, 127, 255},
		"steelblue":            {70, 130, 180, 255},
		"tan":                  {210, 180, 140, 255},
		"thistle":              {216, 191, 216, 255},
		"tomato":               {255, 99, 71, 255},
		"turquoise":            {64, 224, 208, 255},
		"violet":               {238, 130, 238, 255},
		"wheat":                {245, 222, 179, 255},
		"whitesmoke":           {245, 245, 245, 255},
		"yellowgreen":          {154, 205, 50, 255},
	}

	if col, ok := namedColors[value]; ok {
		return col
	}

	// Hex colors (#RGB or #RRGGBB)
	if strings.HasPrefix(value, "#") {
		return parseHexColor(value)
	}

	// Default to black
	return color.RGBA{0, 0, 0, 255}
}

// parseHexColor parses a hex color string (#RGB or #RRGGBB).
func parseHexColor(hex string) color.RGBA {
	hex = strings.TrimPrefix(hex, "#")

	var r, g, b uint8

	switch len(hex) {
	case 3: // #RGB
		if rr, err := strconv.ParseUint(string(hex[0])+string(hex[0]), 16, 8); err == nil {
			r = uint8(rr)
		}
		if gg, err := strconv.ParseUint(string(hex[1])+string(hex[1]), 16, 8); err == nil {
			g = uint8(gg)
		}
		if bb, err := strconv.ParseUint(string(hex[2])+string(hex[2]), 16, 8); err == nil {
			b = uint8(bb)
		}
	case 6: // #RRGGBB
		if rr, err := strconv.ParseUint(hex[0:2], 16, 8); err == nil {
			r = uint8(rr)
		}
		if gg, err := strconv.ParseUint(hex[2:4], 16, 8); err == nil {
			g = uint8(gg)
		}
		if bb, err := strconv.ParseUint(hex[4:6], 16, 8); err == nil {
			b = uint8(bb)
		}
	}

	return color.RGBA{r, g, b, 255}
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

// renderImage renders an image element if present.
// HTML5 §4.8.2 The img element
func renderImage(canvas *Canvas, box *layout.LayoutBox) {
	if box.StyledNode == nil || box.StyledNode.Node == nil {
		return
	}

	// Check if this is an img element
	if box.StyledNode.Node.Data != "img" {
		return
	}

	// Get the src attribute
	src := box.StyledNode.Node.GetAttribute("src")
	if src == "" {
		return
	}

	// Load the image
	img, err := canvas.LoadImage(src)
	if err != nil {
		// Silently fail if image can't be loaded
		return
	}

	// Render the image in the content box
	canvas.DrawImage(
		img,
		int(box.Dimensions.Content.X),
		int(box.Dimensions.Content.Y),
		int(box.Dimensions.Content.Width),
		int(box.Dimensions.Content.Height),
	)
}
