package render

import (
	"image/color"
	"testing"

	"github.com/lukehoban/browser/layout"
	"github.com/lukehoban/browser/style"
)

func TestNewCanvas(t *testing.T) {
	c := NewCanvas(100, 50)

	if c.Width != 100 {
		t.Errorf("expected width 100, got %d", c.Width)
	}
	if c.Height != 50 {
		t.Errorf("expected height 50, got %d", c.Height)
	}
	if len(c.Pixels) != 5000 {
		t.Errorf("expected 5000 pixels, got %d", len(c.Pixels))
	}
}

func TestCanvasClear(t *testing.T) {
	c := NewCanvas(10, 10)
	white := color.RGBA{255, 255, 255, 255}
	c.Clear(white)

	for i, px := range c.Pixels {
		if px != white {
			t.Errorf("pixel %d expected white, got %v", i, px)
			break
		}
	}
}

func TestCanvasSetPixel(t *testing.T) {
	c := NewCanvas(10, 10)
	red := color.RGBA{255, 0, 0, 255}

	c.SetPixel(5, 5, red)

	// Check the pixel at (5, 5)
	if c.Pixels[5*10+5] != red {
		t.Errorf("expected red at (5,5), got %v", c.Pixels[5*10+5])
	}

	// Check bounds - these should not panic
	c.SetPixel(-1, 0, red)
	c.SetPixel(0, -1, red)
	c.SetPixel(10, 5, red)
	c.SetPixel(5, 10, red)
}

func TestCanvasFillRect(t *testing.T) {
	c := NewCanvas(20, 20)
	c.Clear(color.RGBA{255, 255, 255, 255})
	blue := color.RGBA{0, 0, 255, 255}

	c.FillRect(5, 5, 10, 10, blue)

	// Check inside the rectangle
	if c.Pixels[7*20+7] != blue {
		t.Errorf("expected blue inside rect, got %v", c.Pixels[7*20+7])
	}

	// Check outside the rectangle
	if c.Pixels[0*20+0].B != 255 || c.Pixels[0*20+0].R != 255 {
		t.Errorf("expected white outside rect, got %v", c.Pixels[0])
	}
}

func TestCanvasDrawRect(t *testing.T) {
	c := NewCanvas(30, 30)
	white := color.RGBA{255, 255, 255, 255}
	red := color.RGBA{255, 0, 0, 255}
	c.Clear(white)

	// Draw a rectangle outline at (5, 5) with size 20x20 and thickness 2
	c.DrawRect(5, 5, 20, 20, red, 2)

	// Check top border - pixel at (10, 5) should be red (inside top border)
	if c.Pixels[5*30+10] != red {
		t.Errorf("expected red at top border (10,5), got %v", c.Pixels[5*30+10])
	}

	// Check bottom border - pixel at (10, 23) should be red (inside bottom border, y=5+20-2=23)
	if c.Pixels[23*30+10] != red {
		t.Errorf("expected red at bottom border (10,23), got %v", c.Pixels[23*30+10])
	}

	// Check left border - pixel at (5, 15) should be red
	if c.Pixels[15*30+5] != red {
		t.Errorf("expected red at left border (5,15), got %v", c.Pixels[15*30+5])
	}

	// Check right border - pixel at (23, 15) should be red (x=5+20-2=23)
	if c.Pixels[15*30+23] != red {
		t.Errorf("expected red at right border (23,15), got %v", c.Pixels[15*30+23])
	}

	// Check center should be white (not filled)
	if c.Pixels[15*30+15] != white {
		t.Errorf("expected white at center (15,15), got %v", c.Pixels[15*30+15])
	}

	// Check outside should be white
	if c.Pixels[0*30+0] != white {
		t.Errorf("expected white outside rect (0,0), got %v", c.Pixels[0])
	}
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		input    string
		expected color.RGBA
	}{
		{"black", color.RGBA{0, 0, 0, 255}},
		{"white", color.RGBA{255, 255, 255, 255}},
		{"red", color.RGBA{255, 0, 0, 255}},
		{"blue", color.RGBA{0, 0, 255, 255}},
		{"#FF0000", color.RGBA{255, 0, 0, 255}},
		{"#00FF00", color.RGBA{0, 255, 0, 255}},
		{"#0000FF", color.RGBA{0, 0, 255, 255}},
		{"#f00", color.RGBA{255, 0, 0, 255}},
		{"#0f0", color.RGBA{0, 255, 0, 255}},
		{"#00f", color.RGBA{0, 0, 255, 255}},
		{"unknown", color.RGBA{0, 0, 0, 255}}, // defaults to black
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseColor(tt.input)
			if result != tt.expected {
				t.Errorf("parseColor(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRender(t *testing.T) {
	// Create a simple styled node
	styledNode := &style.StyledNode{
		Styles: map[string]string{
			"background": "blue",
			"width":      "100px",
			"height":     "50px",
		},
		Children: []*style.StyledNode{},
	}

	// Build layout tree
	box := &layout.LayoutBox{
		BoxType:    layout.BlockBox,
		StyledNode: styledNode,
		Dimensions: layout.Dimensions{
			Content: layout.Rect{
				X:      10,
				Y:      10,
				Width:  100,
				Height: 50,
			},
		},
		Children: []*layout.LayoutBox{},
	}

	// Render
	canvas := Render(box, 200, 100)

	if canvas.Width != 200 || canvas.Height != 100 {
		t.Errorf("canvas size unexpected: %dx%d", canvas.Width, canvas.Height)
	}

	// Check that the background was rendered (inside the box)
	blue := color.RGBA{0, 0, 255, 255}
	pixelInBox := canvas.Pixels[30*200+50] // y=30, x=50 should be in the box
	if pixelInBox != blue {
		t.Errorf("expected blue inside box, got %v", pixelInBox)
	}
}

func TestToImage(t *testing.T) {
	c := NewCanvas(10, 10)
	red := color.RGBA{255, 0, 0, 255}
	c.Clear(red)

	img := c.ToImage()

	if img.Bounds().Dx() != 10 || img.Bounds().Dy() != 10 {
		t.Errorf("image size unexpected: %v", img.Bounds())
	}

	r, g, b, a := img.At(5, 5).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 || a>>8 != 255 {
		t.Errorf("expected red, got rgba(%d, %d, %d, %d)", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestDrawText(t *testing.T) {
	c := NewCanvas(100, 50)
	white := color.RGBA{255, 255, 255, 255}
	black := color.RGBA{0, 0, 0, 255}
	c.Clear(white)

	// Draw text "Hello"
	c.DrawText("Hello", 10, 20, black)

	// Check that some pixels are black (text was drawn)
	// We can't check exact positions since font rendering is complex,
	// but we can verify that not all pixels are white anymore
	hasBlackPixels := false
	for _, px := range c.Pixels {
		if px.R < 255 || px.G < 255 || px.B < 255 {
			hasBlackPixels = true
			break
		}
	}

	if !hasBlackPixels {
		t.Errorf("expected text to be drawn (some non-white pixels), but canvas is all white")
	}
}
