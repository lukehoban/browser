// Package render implements the rendering engine for the browser.
// It converts a layout tree into a visual representation as a PNG image.
//
// Spec references:
// - CSS 2.1 §14 Colors and backgrounds
// - CSS 2.1 §8 Box model
package render

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"
	"strings"

	"github.com/lukehoban/browser/layout"
)

// Canvas represents the rendering surface.
type Canvas struct {
	Width  int
	Height int
	Pixels []color.RGBA
}

// NewCanvas creates a new canvas with the given dimensions.
func NewCanvas(width, height int) *Canvas {
	return &Canvas{
		Width:  width,
		Height: height,
		Pixels: make([]color.RGBA, width*height),
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
	defer file.Close()

	return png.Encode(file, c.ToImage())
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
