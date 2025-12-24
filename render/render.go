// Package render implements the rendering engine for the browser.
// It converts a layout tree into a visual representation as a PNG image.
//
// Spec references:
// - CSS 2.1 §14 Colors and backgrounds
// - CSS 2.1 §8 Box model
package render

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"net/http"
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

// DrawText draws text at the given position with the given color.
// CSS 2.1 §16 Text
func (c *Canvas) DrawText(text string, x, y int, col color.RGBA) {
	// Use basicfont.Face7x13 as a simple built-in font
	face := basicfont.Face7x13
	
	// Calculate the bounding box for the text
	width := len(text) * face.Advance
	height := face.Height
	
	// Create a temporary image for just the text
	textImg := image.NewRGBA(image.Rect(0, 0, width, height))
	
	// Create a drawer for the text
	drawer := &font.Drawer{
		Dst:  textImg,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.Point26_6{X: 0, Y: fixed.I(face.Ascent)},
	}
	
	// Draw the text
	drawer.DrawString(text)
	
	// Copy only the text pixels to the canvas
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			px := x + dx
			py := y - face.Ascent + dy // Adjust for baseline
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
// Network loading follows standard HTTP/HTTPS protocols for remote images.
func (c *Canvas) LoadImage(path string) (image.Image, error) {
	// Check cache first
	if img, ok := c.ImageCache[path]; ok {
		return img, nil
	}

	var img image.Image
	var err error

	// Check if path is a URL
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		// Fetch from network
		img, err = loadImageFromURL(path)
	} else {
		// Load from file
		img, err = loadImageFromFile(path)
	}

	if err != nil {
		return nil, err
	}

	// Cache the image
	c.ImageCache[path] = img

	return img, nil
}

// loadImageFromFile loads an image from a local file
func loadImageFromFile(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	return img, err
}

// loadImageFromURL loads an image from a URL
func loadImageFromURL(urlStr string) (image.Image, error) {
	resp, err := http.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	img, _, err := image.Decode(resp.Body)
	return img, err
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

	// Render the text at the box's position
	// Add a small vertical offset to position text at baseline
	x := int(box.Dimensions.Content.X)
	y := int(box.Dimensions.Content.Y) + 13 // basicfont.Face7x13 height is 13 pixels

	canvas.DrawText(text, x, y, textColor)
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

