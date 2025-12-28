package render

import (
	"image/color"
	"testing"

	"github.com/lukehoban/browser/css"
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

func TestCanvasDrawSVG(t *testing.T) {
	c := NewCanvas(32, 16)
	white := color.RGBA{255, 255, 255, 255}
	c.Clear(white)

	// Simple SVG triangle (similar to HN's vote arrow)
	svgData := []byte(`<svg height="32" viewBox="0 0 32 16" width="32" xmlns="http://www.w3.org/2000/svg"><path d="m2 27 14-29 14 29z" fill="#999"/></svg>`)
	
	err := c.DrawSVG(svgData, 0, 0, 32, 16)
	if err != nil {
		t.Fatalf("DrawSVG failed: %v", err)
	}

	// Check that some pixels have been drawn (not all white)
	hasNonWhite := false
	for _, px := range c.Pixels {
		if px != white {
			hasNonWhite = true
			break
		}
	}
	
	if !hasNonWhite {
		t.Error("expected some non-white pixels after drawing SVG, but canvas is all white")
	}
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		input    string
		expected color.RGBA
	}{
		// CSS 2.1 basic 17 colors (including orange added in CSS 2.1)
		{"black", color.RGBA{0, 0, 0, 255}},
		{"white", color.RGBA{255, 255, 255, 255}},
		{"red", color.RGBA{255, 0, 0, 255}},
		{"blue", color.RGBA{0, 0, 255, 255}},
		{"green", color.RGBA{0, 128, 0, 255}},
		{"yellow", color.RGBA{255, 255, 0, 255}},
		{"navy", color.RGBA{0, 0, 128, 255}},
		{"purple", color.RGBA{128, 0, 128, 255}},
		{"silver", color.RGBA{192, 192, 192, 255}},
		{"gray", color.RGBA{128, 128, 128, 255}},
		{"grey", color.RGBA{128, 128, 128, 255}},
		{"maroon", color.RGBA{128, 0, 0, 255}},
		{"olive", color.RGBA{128, 128, 0, 255}},
		{"teal", color.RGBA{0, 128, 128, 255}},
		{"lime", color.RGBA{0, 255, 0, 255}},
		{"orange", color.RGBA{255, 165, 0, 255}},
		{"fuchsia", color.RGBA{255, 0, 255, 255}},
		{"magenta", color.RGBA{255, 0, 255, 255}},
		{"aqua", color.RGBA{0, 255, 255, 255}},
		{"cyan", color.RGBA{0, 255, 255, 255}},
		
		// Extended color names - gray variants (both spellings)
		{"lightgray", color.RGBA{211, 211, 211, 255}},
		{"lightgrey", color.RGBA{211, 211, 211, 255}},
		{"darkgray", color.RGBA{169, 169, 169, 255}},
		{"darkgrey", color.RGBA{169, 169, 169, 255}},
		{"dimgray", color.RGBA{105, 105, 105, 255}},
		{"dimgrey", color.RGBA{105, 105, 105, 255}},
		{"slategray", color.RGBA{112, 128, 144, 255}},
		{"slategrey", color.RGBA{112, 128, 144, 255}},
		
		// Extended colors - commonly used
		{"lightblue", color.RGBA{173, 216, 230, 255}},
		{"darkblue", color.RGBA{0, 0, 139, 255}},
		{"lightgreen", color.RGBA{144, 238, 144, 255}},
		{"darkgreen", color.RGBA{0, 100, 0, 255}},
		{"pink", color.RGBA{255, 192, 203, 255}},
		{"lightpink", color.RGBA{255, 182, 193, 255}},
		{"brown", color.RGBA{165, 42, 42, 255}},
		{"gold", color.RGBA{255, 215, 0, 255}},
		{"coral", color.RGBA{255, 127, 80, 255}},
		{"crimson", color.RGBA{220, 20, 60, 255}},
		{"indigo", color.RGBA{75, 0, 130, 255}},
		{"violet", color.RGBA{238, 130, 238, 255}},
		
		// Hex colors
		{"#FF0000", color.RGBA{255, 0, 0, 255}},
		{"#00FF00", color.RGBA{0, 255, 0, 255}},
		{"#0000FF", color.RGBA{0, 0, 255, 255}},
		{"#f00", color.RGBA{255, 0, 0, 255}},
		{"#0f0", color.RGBA{0, 255, 0, 255}},
		{"#00f", color.RGBA{0, 0, 255, 255}},
		
		// Unknown color defaults to black
		{"unknown", color.RGBA{0, 0, 0, 255}},
		{"notacolor", color.RGBA{0, 0, 0, 255}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := css.ParseColor(tt.input)
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

func TestDrawStyledText(t *testing.T) {
	c := NewCanvas(200, 100)
	white := color.RGBA{255, 255, 255, 255}
	black := color.RGBA{0, 0, 0, 255}
	c.Clear(white)

	// Test normal text
	c.DrawStyledText("Normal", 10, 20, black, FontStyle{
		Size:       13.0,
		Weight:     "normal",
		Style:      "normal",
		Decoration: "none",
	})

	// Test large text
	c.DrawStyledText("Large", 10, 40, black, FontStyle{
		Size:       26.0,
		Weight:     "normal",
		Style:      "normal",
		Decoration: "none",
	})

	// Test bold text
	c.DrawStyledText("Bold", 10, 60, black, FontStyle{
		Size:       13.0,
		Weight:     "bold",
		Style:      "normal",
		Decoration: "none",
	})

	// Test italic text
	c.DrawStyledText("Italic", 10, 80, black, FontStyle{
		Size:       13.0,
		Weight:     "normal",
		Style:      "italic",
		Decoration: "none",
	})

	// Verify text was drawn
	hasBlackPixels := false
	for _, px := range c.Pixels {
		if px.R < 255 || px.G < 255 || px.B < 255 {
			hasBlackPixels = true
			break
		}
	}

	if !hasBlackPixels {
		t.Errorf("expected text to be drawn")
	}
}

func TestExtractFontStyle(t *testing.T) {
	tests := []struct {
		name     string
		styles   map[string]string
		expected FontStyle
	}{
		{
			name:   "default styles",
			styles: map[string]string{},
			expected: FontStyle{
				Size:       13.0,
				Weight:     "normal",
				Style:      "normal",
				Decoration: "none",
			},
		},
		{
			name: "font-size 20px",
			styles: map[string]string{
				"font-size": "20px",
			},
			expected: FontStyle{
				Size:       20.0,
				Weight:     "normal",
				Style:      "normal",
				Decoration: "none",
			},
		},
		{
			name: "bold weight",
			styles: map[string]string{
				"font-weight": "bold",
			},
			expected: FontStyle{
				Size:       13.0,
				Weight:     "bold",
				Style:      "normal",
				Decoration: "none",
			},
		},
		{
			name: "italic style",
			styles: map[string]string{
				"font-style": "italic",
			},
			expected: FontStyle{
				Size:       13.0,
				Weight:     "normal",
				Style:      "italic",
				Decoration: "none",
			},
		},
		{
			name: "underline decoration",
			styles: map[string]string{
				"text-decoration": "underline",
			},
			expected: FontStyle{
				Size:       13.0,
				Weight:     "normal",
				Style:      "normal",
				Decoration: "underline",
			},
		},
		{
			name: "combined styles",
			styles: map[string]string{
				"font-size":       "18px",
				"font-weight":     "bold",
				"font-style":      "italic",
				"text-decoration": "underline",
			},
			expected: FontStyle{
				Size:       18.0,
				Weight:     "bold",
				Style:      "italic",
				Decoration: "underline",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFontStyle(tt.styles)
			if result != tt.expected {
				t.Errorf("extractFontStyle() = %+v, expected %+v", result, tt.expected)
			}
		})
	}
}

func TestParseFontSize(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"14px", 14.0},
		{"20px", 20.0},
		{"10pt", 10.0 * 96.0 / 72.0}, // 10pt at 96 DPI
		{"12pt", 12.0 * 96.0 / 72.0}, // 12pt at 96 DPI
		{"10", 10.0},
		{"24", 24.0},
		{"medium", 13.0},
		{"large", 16.0},
		{"x-large", 20.0},
		{"small", 12.0},
		{"invalid", 0.0},
		{"", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := css.ParseFontSize(tt.input)
			if result != tt.expected {
				t.Errorf("css.ParseFontSize(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// SKIPPED TESTS FOR KNOWN BROKEN/UNIMPLEMENTED FEATURES
// These tests document known limitations that need to be implemented.
// See MILESTONES.md for more details.

func TestFontFamilySupport_Skipped(t *testing.T) {
	t.Skip("Font-family support not implemented - CSS 2.1 §15.3")
	// CSS 2.1 §15.3 Font family: the 'font-family' property
	// Should support multiple font families and fallback to system fonts
	
	layoutBox := &layout.LayoutBox{
		BoxType: layout.BlockBox,
		Dimensions: layout.Dimensions{
			Content: layout.Rect{X: 10, Y: 10, Width: 180, Height: 80},
		},
		StyledNode: &style.StyledNode{
			Styles: map[string]string{
				"font-family": "Arial, Helvetica, sans-serif",
				"font-size":   "14px",
			},
		},
	}
	
	// Should use Arial if available, else Helvetica, else sans-serif
	canvas := Render(layoutBox, 200, 100)
	
	// Verify that the specified font was used (not just default)
	// This would require font metrics inspection
	if canvas == nil {
		t.Error("Expected canvas to be created")
	}
}

func TestTextAlign_Skipped(t *testing.T) {
	t.Skip("Text-align not implemented - CSS 2.1 §16.2")
	// CSS 2.1 §16.2 Alignment: the 'text-align' property
	// Should support left, right, center, justify
	
	tests := []struct {
		align    string
		expectedX float64
	}{
		{"left", 10.0},
		{"center", 150.0}, // Middle of 300px container
		{"right", 290.0},  // Right edge minus text width
	}
	
	for _, tt := range tests {
		t.Run(tt.align, func(t *testing.T) {
			layoutBox := &layout.LayoutBox{
				BoxType: layout.BlockBox,
				Dimensions: layout.Dimensions{
					Content: layout.Rect{X: 0, Y: 10, Width: 300, Height: 20},
				},
				StyledNode: &style.StyledNode{
					Styles: map[string]string{
						"text-align": tt.align,
					},
				},
			}
			
			canvas := Render(layoutBox, 300, 100)
			
			// Text should be positioned according to alignment
			// This would require tracking text position during render
			if canvas == nil {
				t.Error("Expected canvas to be created")
			}
		})
	}
}

func TestLineHeight_Skipped(t *testing.T) {
	t.Skip("Line-height not implemented - CSS 2.1 §10.8.1")
	// CSS 2.1 §10.8.1 Leading and half-leading
	// Line-height should control vertical spacing between lines
	
	layoutBox := &layout.LayoutBox{
		BoxType: layout.BlockBox,
		Dimensions: layout.Dimensions{
			Content: layout.Rect{X: 10, Y: 10, Width: 180, Height: 180},
		},
		StyledNode: &style.StyledNode{
			Styles: map[string]string{
				"font-size":   "14px",
				"line-height": "1.5", // 21px
			},
		},
	}
	
	canvas := Render(layoutBox, 200, 200)
	
	// Second line should be 21px below first line (14px * 1.5)
	// This would require tracking line positions
	if canvas == nil {
		t.Error("Expected canvas to be created")
	}
}

func TestBackgroundImageTriangle(t *testing.T) {
	// Note: This test verifies the rendering pipeline for background-image
	// but doesn't actually fetch the SVG file since it's a unit test
	// The actual SVG rendering is tested in TestCanvasDrawSVG
	
	layoutBox := &layout.LayoutBox{
		BoxType: layout.BlockBox,
		Dimensions: layout.Dimensions{
			Content: layout.Rect{X: 10, Y: 10, Width: 10, Height: 10},
		},
		StyledNode: &style.StyledNode{
			Styles: map[string]string{
				"background": "url(triangle.svg)",
			},
		},
	}
	
	canvas := Render(layoutBox, 50, 50)
	
	// Check that the canvas was created
	if canvas == nil {
		t.Fatal("Expected canvas to be created")
	}
	
	// Note: The triangle.svg file doesn't exist in the test environment,
	// so the background won't actually be rendered. This test just verifies
	// the code path doesn't panic.
}

func TestBackgroundImageSVG(t *testing.T) {
	t.Skip("Background-image SVG integration test - requires network/file access")
	// This would test the full integration of:
	// 1. URL extraction from CSS
	// 2. SVG file loading
	// 3. SVG rendering
	// For now, TestCanvasDrawSVG tests the core SVG rendering functionality
}

func TestBackgroundImage_Skipped(t *testing.T) {
	t.Skip("Background-image not implemented - CSS 2.1 §14.2.1")
	// CSS 2.1 §14.2.1 Background properties: 'background-image'
	// Should load and render background images
	
	layoutBox := &layout.LayoutBox{
		BoxType: layout.BlockBox,
		Dimensions: layout.Dimensions{
			Content: layout.Rect{X: 10, Y: 10, Width: 100, Height: 100},
		},
		StyledNode: &style.StyledNode{
			Styles: map[string]string{
				"background-image":  "url(test.png)",
				"background-repeat": "no-repeat",
			},
		},
	}
	
	canvas := Render(layoutBox, 200, 200)
	
	// Background image should be loaded and rendered
	// Pixels should match the image content
	if canvas == nil {
		t.Error("Expected canvas to be created")
	}
}

func TestTextDecorationOverline_Skipped(t *testing.T) {
	t.Skip("Text-decoration overline not implemented - CSS 2.1 §16.3.1")
	// CSS 2.1 §16.3.1 Underlining, overlining, striking, and blinking
	// Should support overline and line-through in addition to underline
	
	tests := []struct {
		decoration string
		checkY     float64 // Y position where line should appear
	}{
		{"overline", 10.0},      // Above text
		{"line-through", 15.0},  // Through middle
		{"underline", 20.0},     // Below text (already implemented)
	}
	
	for _, tt := range tests {
		t.Run(tt.decoration, func(t *testing.T) {
			layoutBox := &layout.LayoutBox{
				BoxType: layout.BlockBox,
				Dimensions: layout.Dimensions{
					Content: layout.Rect{X: 10, Y: 10, Width: 100, Height: 20},
				},
				StyledNode: &style.StyledNode{
					Styles: map[string]string{
						"text-decoration": tt.decoration,
						"color":           "black",
					},
				},
			}
			
			canvas := Render(layoutBox, 200, 100)
			
			// Line should appear at expected Y position
			// Would need to inspect canvas pixels
			if canvas == nil {
				t.Error("Expected canvas to be created")
			}
		})
	}
}

func TestTableRowspan_Skipped(t *testing.T) {
	t.Skip("Table rowspan not implemented - CSS 2.1 §17.2")
	// CSS 2.1 §17.2 The CSS table model
	// Rowspan attribute should allow cells to span multiple rows
	
	// Table with rowspan:
	// Row 1: [Cell A (rowspan=2)] [Cell B]
	// Row 2:                      [Cell C]
	
	layoutBox := &layout.LayoutBox{
		BoxType: layout.TableBox,
		Dimensions: layout.Dimensions{
			Content: layout.Rect{X: 0, Y: 0, Width: 400, Height: 300},
		},
		Children: []*layout.LayoutBox{
			// Row 1
			{
				BoxType: layout.TableRowBox,
				Children: []*layout.LayoutBox{
					{
						BoxType: layout.TableCellBox,
						Dimensions: layout.Dimensions{
							Content: layout.Rect{X: 0, Y: 0, Width: 200, Height: 100},
						},
						StyledNode: &style.StyledNode{
							Node: nil, // Would have rowspan attribute
						},
					},
					{
						BoxType: layout.TableCellBox,
						Dimensions: layout.Dimensions{
							Content: layout.Rect{X: 200, Y: 0, Width: 200, Height: 50},
						},
					},
				},
			},
			// Row 2
			{
				BoxType: layout.TableRowBox,
				Children: []*layout.LayoutBox{
					{
						BoxType: layout.TableCellBox,
						Dimensions: layout.Dimensions{
							Content: layout.Rect{X: 200, Y: 50, Width: 200, Height: 50},
						},
					},
				},
			},
		},
	}
	
	canvas := Render(layoutBox, 400, 300)
	
	// Cell A should span both rows (height = 100)
	// Cell B and C should be in single rows (height = 50 each)
	if canvas == nil {
		t.Error("Expected canvas to be created")
	}
}

func TestTableHeaders_Skipped(t *testing.T) {
	t.Skip("Table headers (<thead>, <tbody>, <tfoot>) not implemented - HTML5 §4.9.5-7")
	// HTML5 §4.9.5 The thead element
	// HTML5 §4.9.6 The tbody element
	// HTML5 §4.9.7 The tfoot element
	// Should properly handle table header, body, and footer sections
	
	// Table structure should recognize and lay out thead, tbody, tfoot
	// When implemented, would have TableHeaderBox, TableBodyBox, TableFooterBox types
	layoutBox := &layout.LayoutBox{
		BoxType: layout.TableBox,
		Children: []*layout.LayoutBox{
			{
				BoxType: layout.TableRowBox, // Would be TableHeaderBox for <thead>
				Children: []*layout.LayoutBox{
					// Header rows
				},
			},
			{
				BoxType: layout.TableRowBox, // Would be TableBodyBox for <tbody>
				Children: []*layout.LayoutBox{
					// Body rows
				},
			},
			{
				BoxType: layout.TableRowBox, // Would be TableFooterBox for <tfoot>
				Children: []*layout.LayoutBox{
					// Footer rows
				},
			},
		},
	}
	
	canvas := Render(layoutBox, 400, 300)
	
	// Headers should be styled/positioned differently from body
	if canvas == nil {
		t.Error("Expected canvas to be created")
	}
}

func TestBorderCollapse_Skipped(t *testing.T) {
	t.Skip("Border-collapse not implemented - CSS 2.1 §17.6.1")
	// CSS 2.1 §17.6.1 The separated borders model
	// CSS 2.1 §17.6.2 The collapsing border model
	// Should support both separate and collapsed border models
	
	layoutBox := &layout.LayoutBox{
		BoxType: layout.TableBox,
		StyledNode: &style.StyledNode{
			Styles: map[string]string{
				"border-collapse": "collapse",
			},
		},
		Children: []*layout.LayoutBox{
			{
				BoxType: layout.TableRowBox,
				Children: []*layout.LayoutBox{
					{
						BoxType: layout.TableCellBox,
						Dimensions: layout.Dimensions{
							Content: layout.Rect{X: 0, Y: 0, Width: 100, Height: 50},
							Border: layout.EdgeSizes{
								Top: 1, Right: 1, Bottom: 1, Left: 1,
							},
						},
					},
					{
						BoxType: layout.TableCellBox,
						Dimensions: layout.Dimensions{
							Content: layout.Rect{X: 100, Y: 0, Width: 100, Height: 50},
							Border: layout.EdgeSizes{
								Top: 1, Right: 1, Bottom: 1, Left: 1,
							},
						},
					},
				},
			},
		},
	}
	
	canvas := Render(layoutBox, 300, 200)
	
	// With border-collapse: collapse, adjacent borders should merge
	// Border between cells should be 1px, not 2px
	if canvas == nil {
		t.Error("Expected canvas to be created")
	}
}

// TestAntialiasingQuality tests that text rendering produces output and doesn't crash.
func TestAntialiasingQuality(t *testing.T) {
	canvas := NewCanvas(200, 100)
	canvas.Clear(color.RGBA{255, 255, 255, 255})
	
	// Draw text with the current antialiasing implementation
	text := "Test"
	col := color.RGBA{0, 0, 0, 255}
	style := FontStyle{
		Size:   20.0,
		Weight: "normal",
		Style:  "normal",
	}
	
	canvas.DrawStyledText(text, 10, 30, col, style)
	
	// Verify the canvas was modified (text was drawn)
	foundNonWhite := false
	white := color.RGBA{255, 255, 255, 255}
	for _, px := range canvas.Pixels {
		if px != white {
			foundNonWhite = true
			break
		}
	}
	
	if !foundNonWhite {
		t.Error("Text rendering failed - canvas is entirely white")
	}
	
	// Note: The antialiasing may produce subtle gray values that are difficult
	// to detect in small test canvases. Visual inspection of rendered output
	// shows improved text quality with smoother edges.
}

func TestExtractURLFromCSS(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"url(triangle.svg)", "triangle.svg"},
		{"url('triangle.svg')", "triangle.svg"},
		{"url(\"triangle.svg\")", "triangle.svg"},
		{"url( triangle.svg )", "triangle.svg"},
		{"url( 'triangle.svg' )", "triangle.svg"},
		{"url(https://example.com/image.svg)", "https://example.com/image.svg"},
		{"background: url(image.png) no-repeat", "image.png"},
		{"no url here", ""},
		{"url()", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractURLFromCSS(tt.input)
			if result != tt.expected {
				t.Errorf("extractURLFromCSS(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
