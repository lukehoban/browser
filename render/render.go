// Package render implements the rendering engine for the browser.
// It converts a layout tree into a visual representation as a PNG image.
//
// Spec references:
// - CSS 2.1 §14 Colors and backgrounds
// - CSS 2.1 §8 Box model
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
func (c *Canvas) DrawText(text string, x, y int, col color.RGBA) {
	// Use default font style for backward compatibility
	c.DrawStyledText(text, x, y, col, FontStyle{
		Size:       13.0,
		Weight:     "normal",
		Style:      "normal",
		Decoration: "none",
	})
}

// DrawStyledText draws text with font styling at the given position.
// CSS 2.1 §15 Fonts and §16 Text
// CSS 2.1 §15.6 Font boldness: font-weight
// CSS 2.1 §15.7 Font style: font-style
// CSS 2.1 §16.3.1 Underlining, overlining, striking: text-decoration
func (c *Canvas) DrawStyledText(text string, x, y int, col color.RGBA, style FontStyle) {
	// Use basicfont.Face7x13 as a simple built-in font
	// This is a 7x13 pixel font, where 7 is advance and 13 is height
	baseFace := basicfont.Face7x13
	
	// Calculate scale factor based on desired font size
	// basicfont.Face7x13 has height of 13 pixels
	scale := style.Size / 13.0
	if scale <= 0 {
		scale = 1.0
	}
	
	// Calculate the dimensions for scaled text
	baseWidth := len(text) * baseFace.Advance
	baseHeight := baseFace.Height
	
	scaledWidth := int(float64(baseWidth) * scale)
	scaledHeight := int(float64(baseHeight) * scale)
	
	// Create a temporary image for the base text
	baseImg := image.NewRGBA(image.Rect(0, 0, baseWidth, baseHeight))
	
	// Create a drawer for the text
	drawer := &font.Drawer{
		Dst:  baseImg,
		Src:  image.NewUniform(col),
		Face: baseFace,
		Dot:  fixed.Point26_6{X: 0, Y: fixed.I(baseFace.Ascent)},
	}
	
	// Draw the base text
	drawer.DrawString(text)
	
	// Apply bold effect by drawing text with slight offset
	// CSS 2.1 §15.6: Synthetic bold can be created by drawing twice with offset
	if style.Weight == "bold" {
		drawer.Dot = fixed.Point26_6{X: fixed.I(1), Y: fixed.I(baseFace.Ascent)}
		drawer.DrawString(text)
	}
	
	// Create scaled image if needed
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
			// Apply italic slant: shift pixels based on vertical position
			// CSS 2.1 §15.7: Synthetic italic can be created by slanting
			slant := 0
			if style.Style == "italic" {
				// Slant by about 15 degrees (tan(15°) ≈ 0.27)
				// Shift more at the top, less at the bottom
				slant = int(float64(scaledHeight-dy) * 0.2)
			}
			
			px := x + dx + slant
			py := y - baselineOffset + dy
			
			if px >= 0 && px < c.Width && py >= 0 && py < c.Height {
				r, g, b, a := textImg.At(dx, dy).RGBA()
				// Only copy non-transparent pixels (text)
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
	
	// Draw underline if needed
	// CSS 2.1 §16.3.1: Underline is drawn below the baseline
	if style.Decoration == "underline" {
		underlineY := y + int(float64(baseFace.Descent)*scale/2)
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
	
	// Parse font-weight (CSS 2.1 §15.6)
	if fontWeight := styles["font-weight"]; fontWeight != "" {
		fontWeight = strings.TrimSpace(strings.ToLower(fontWeight))
		if fontWeight == "bold" || fontWeight == "bolder" || fontWeight >= "600" {
			fontStyle.Weight = "bold"
		}
	}
	
	// Parse font-style (CSS 2.1 §15.7)
	if fontStyleVal := styles["font-style"]; fontStyleVal != "" {
		fontStyleVal = strings.TrimSpace(strings.ToLower(fontStyleVal))
		if fontStyleVal == "italic" || fontStyleVal == "oblique" {
			fontStyle.Style = "italic"
		}
	}
	
	// Parse text-decoration (CSS 2.1 §16.3.1)
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
	
	// Handle pixel values (e.g., "14px")
	if strings.HasSuffix(value, "px") {
		value = strings.TrimSuffix(value, "px")
		if size, err := strconv.ParseFloat(value, 64); err == nil {
			return size
		}
	}
	
	// Handle plain numbers (treat as pixels)
	if size, err := strconv.ParseFloat(value, 64); err == nil {
		return size
	}
	
	// Handle named sizes (CSS 2.1 §15.7)
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
	
	return 0 // Return 0 for invalid values (will use default)
}

// parseColor parses a CSS color value and returns a color.RGBA.
// Supports basic color names and hex colors.
// CSS 2.1 §4.3.6 Colors
func parseColor(value string) color.RGBA {
	value = strings.TrimSpace(strings.ToLower(value))

	// Named colors (CSS 2.1 §4.3.6)
	namedColors := map[string]color.RGBA{
		"black":   {0, 0, 0, 255},
		"white":   {255, 255, 255, 255},
		"red":     {255, 0, 0, 255},
		"green":   {0, 128, 0, 255},
		"blue":    {0, 0, 255, 255},
		"yellow":  {255, 255, 0, 255},
		"cyan":    {0, 255, 255, 255},
		"magenta": {255, 0, 255, 255},
		"gray":    {128, 128, 128, 255},
		"grey":    {128, 128, 128, 255},
		"silver":  {192, 192, 192, 255},
		"maroon":  {128, 0, 0, 255},
		"navy":    {0, 0, 128, 255},
		"olive":   {128, 128, 0, 255},
		"purple":  {128, 0, 128, 255},
		"teal":    {0, 128, 128, 255},
		"orange":  {255, 165, 0, 255},
		"aqua":    {0, 255, 255, 255},
		"fuchsia": {255, 0, 255, 255},
		"lime":    {0, 255, 0, 255},
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
