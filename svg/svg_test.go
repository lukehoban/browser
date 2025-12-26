// Package svg provides tests for SVG parsing and rasterization.
package svg

import (
	"image"
	"image/color"
	"testing"
)

// TestParseTriangleSVG tests parsing the HN vote arrow triangle.svg
func TestParseTriangleSVG(t *testing.T) {
	// HN's triangle.svg for vote arrows
	svgData := []byte(`<svg height="32" viewBox="0 0 32 16" width="32" xmlns="http://www.w3.org/2000/svg"><path d="m2 27 14-29 14 29z" fill="#999"/></svg>`)

	parsed, err := Parse(svgData)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if parsed == nil {
		t.Fatal("Parse returned nil")
	}

	// Check viewBox
	if parsed.ViewBox == nil || len(parsed.ViewBox) != 4 {
		t.Fatal("ViewBox not parsed correctly")
	}
	if parsed.ViewBox[0] != 0 || parsed.ViewBox[1] != 0 || parsed.ViewBox[2] != 32 || parsed.ViewBox[3] != 16 {
		t.Errorf("ViewBox = %v, want [0 0 32 16]", parsed.ViewBox)
	}

	// Check that we have exactly one path
	if len(parsed.Paths) != 1 {
		t.Errorf("len(Paths) = %d, want 1", len(parsed.Paths))
	}

	// Check fill color (#999 = RGB 153,153,153)
	if parsed.Paths[0].FillColor != (color.RGBA{153, 153, 153, 255}) {
		t.Errorf("FillColor = %v, want {153, 153, 153, 255}", parsed.Paths[0].FillColor)
	}

	// Check that we have 4 points (triangle + close point)
	if len(parsed.Paths[0].Points) != 4 {
		t.Errorf("len(Points) = %d, want 4", len(parsed.Paths[0].Points))
	}

	// Check the triangle vertices
	// m2 27 = moveto (2, 27)
	// 14 -29 = relative lineto (16, -2)
	// 14 29 = relative lineto (30, 27)
	// z = closepath back to (2, 27)
	expectedPoints := [][2]float64{{2, 27}, {16, -2}, {30, 27}, {2, 27}}
	for i, expected := range expectedPoints {
		if i >= len(parsed.Paths[0].Points) {
			break
		}
		got := parsed.Paths[0].Points[i]
		if got[0] != expected[0] || got[1] != expected[1] {
			t.Errorf("Point[%d] = %v, want %v", i, got, expected)
		}
	}
}

// TestParseY18SVG tests parsing the HN y18.svg logo
func TestParseY18SVG(t *testing.T) {
	// HN's y18.svg logo
	svgData := []byte(`<svg height="18" viewBox="4 4 188 188" width="18" xmlns="http://www.w3.org/2000/svg"><path d="m4 4h188v188h-188z" fill="#f60"/><path d="m73.2521756 45.01 22.7478244 47.39130083 22.7478244-47.39130083h19.56569631l-34.32352071 64.48661468v41.49338532h-15.98v-41.49338532l-34.32352071-64.48661468z" fill="#fff"/></svg>`)

	parsed, err := Parse(svgData)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if parsed == nil {
		t.Fatal("Parse returned nil")
	}

	// Check viewBox
	if parsed.ViewBox == nil || len(parsed.ViewBox) != 4 {
		t.Fatal("ViewBox not parsed correctly")
	}
	if parsed.ViewBox[0] != 4 || parsed.ViewBox[1] != 4 || parsed.ViewBox[2] != 188 || parsed.ViewBox[3] != 188 {
		t.Errorf("ViewBox = %v, want [4 4 188 188]", parsed.ViewBox)
	}

	// Check that we have exactly two paths (orange background + white Y)
	if len(parsed.Paths) != 2 {
		t.Errorf("len(Paths) = %d, want 2", len(parsed.Paths))
	}

	// Check first path (orange background square)
	orangePath := parsed.Paths[0]
	// #f60 = RGB 255,102,0
	if orangePath.FillColor != (color.RGBA{255, 102, 0, 255}) {
		t.Errorf("Path[0].FillColor = %v, want {255, 102, 0, 255}", orangePath.FillColor)
	}
	// Should have 5 points for the square (4 corners + close)
	if len(orangePath.Points) != 5 {
		t.Errorf("Path[0] len(Points) = %d, want 5", len(orangePath.Points))
	}

	// Check second path (white Y letter)
	whitePath := parsed.Paths[1]
	// #fff = RGB 255,255,255
	if whitePath.FillColor != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("Path[1].FillColor = %v, want {255, 255, 255, 255}", whitePath.FillColor)
	}
	// Y letter path should have 10 points
	if len(whitePath.Points) != 10 {
		t.Errorf("Path[1] len(Points) = %d, want 10", len(whitePath.Points))
	}
}

// TestRenderTriangle tests that the triangle renders correctly
func TestRenderTriangle(t *testing.T) {
	// HN's triangle.svg
	svgData := []byte(`<svg height="32" viewBox="0 0 32 16" width="32" xmlns="http://www.w3.org/2000/svg"><path d="m2 27 14-29 14 29z" fill="#999"/></svg>`)

	parsed, err := Parse(svgData)
	if err != nil || parsed == nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Render to 10x10 (typical vote arrow size)
	width, height := 10, 10
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with white
	white := color.RGBA{255, 255, 255, 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, white)
		}
	}

	// Render the triangle
	rasterizer := Rasterizer{Width: width, Height: height}
	for _, path := range parsed.Paths {
		transformed := TransformPoints(path.Points, parsed.ViewBox, width, height)
		fillFunc := func(x, y int, col color.RGBA) {
			img.Set(x, y, col)
		}
		rasterizer.FillPolygon(transformed, fillFunc, path.FillColor)
	}

	// Check that some pixels were filled (not all white)
	hasGray := false
	gray := color.RGBA{153, 153, 153, 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if img.At(x, y) == gray {
				hasGray = true
				break
			}
		}
	}
	if !hasGray {
		t.Error("Triangle was not rendered - no gray pixels found")
	}

	// The triangle should be rendered in the middle-top area
	// Check that there are gray pixels in the center area
	centerHasGray := false
	for y := 0; y < height/2; y++ {
		for x := width/4; x < 3*width/4; x++ {
			if img.At(x, y) == gray {
				centerHasGray = true
				break
			}
		}
	}
	if !centerHasGray {
		t.Error("Triangle not centered properly - no gray pixels in center-top area")
	}
}

// TestRenderY18 tests that the Y18 logo renders correctly with both paths
func TestRenderY18(t *testing.T) {
	// HN's y18.svg
	svgData := []byte(`<svg height="18" viewBox="4 4 188 188" width="18" xmlns="http://www.w3.org/2000/svg"><path d="m4 4h188v188h-188z" fill="#f60"/><path d="m73.2521756 45.01 22.7478244 47.39130083 22.7478244-47.39130083h19.56569631l-34.32352071 64.48661468v41.49338532h-15.98v-41.49338532l-34.32352071-64.48661468z" fill="#fff"/></svg>`)

	parsed, err := Parse(svgData)
	if err != nil || parsed == nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Render to 36x36 for better visibility
	width, height := 36, 36
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with gray background to detect what's rendered
	gray := color.RGBA{128, 128, 128, 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, gray)
		}
	}

	// Render all paths
	rasterizer := Rasterizer{Width: width, Height: height}
	for _, path := range parsed.Paths {
		transformed := TransformPoints(path.Points, parsed.ViewBox, width, height)
		fillFunc := func(x, y int, col color.RGBA) {
			img.Set(x, y, col)
		}
		rasterizer.FillPolygon(transformed, fillFunc, path.FillColor)
	}

	// Check that orange pixels exist (background square)
	orange := color.RGBA{255, 102, 0, 255}
	hasOrange := false
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if img.At(x, y) == orange {
				hasOrange = true
				break
			}
		}
	}
	if !hasOrange {
		t.Error("Y18 background not rendered - no orange pixels found")
	}

	// Check that white pixels exist (Y letter)
	white := color.RGBA{255, 255, 255, 255}
	hasWhite := false
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if img.At(x, y) == white {
				hasWhite = true
				break
			}
		}
	}
	if !hasWhite {
		t.Error("Y letter not rendered - no white pixels found")
	}
}

// TestTransformPointsUniformScale tests that uniform scaling is applied
func TestTransformPointsUniformScale(t *testing.T) {
	// A square in viewBox coordinates
	points := [][2]float64{{0, 0}, {100, 0}, {100, 100}, {0, 100}}
	viewBox := []float64{0, 0, 100, 100}

	// Transform to a non-square target (wider than tall)
	transformed := TransformPoints(points, viewBox, 200, 100)

	// With uniform scaling (meet), the shape should be centered and maintain aspect ratio
	// Scale = min(200/100, 100/100) = 1.0
	// OffsetX = (200 - 100*1) / 2 = 50
	// OffsetY = (100 - 100*1) / 2 = 0

	expected := [][2]float64{{50, 0}, {150, 0}, {150, 100}, {50, 100}}
	for i, exp := range expected {
		got := transformed[i]
		if got[0] != exp[0] || got[1] != exp[1] {
			t.Errorf("Point[%d] = %v, want %v", i, got, exp)
		}
	}
}

// TestPathCommandH tests horizontal lineto command
func TestPathCommandH(t *testing.T) {
	svgData := []byte(`<svg viewBox="0 0 100 100"><path d="M0 0 h50 v50 h-50 z" fill="#000"/></svg>`)

	parsed, err := Parse(svgData)
	if err != nil || parsed == nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(parsed.Paths) != 1 {
		t.Fatalf("Expected 1 path, got %d", len(parsed.Paths))
	}

	// Should create a 50x50 square
	// M0 0 = (0, 0)
	// h50 = (50, 0)
	// v50 = (50, 50)
	// h-50 = (0, 50)
	// z = back to (0, 0)
	expected := [][2]float64{{0, 0}, {50, 0}, {50, 50}, {0, 50}, {0, 0}}

	if len(parsed.Paths[0].Points) != len(expected) {
		t.Fatalf("Expected %d points, got %d", len(expected), len(parsed.Paths[0].Points))
	}

	for i, exp := range expected {
		got := parsed.Paths[0].Points[i]
		if got[0] != exp[0] || got[1] != exp[1] {
			t.Errorf("Point[%d] = %v, want %v", i, got, exp)
		}
	}
}

// TestPathCommandV tests vertical lineto command
func TestPathCommandV(t *testing.T) {
	svgData := []byte(`<svg viewBox="0 0 100 100"><path d="M10 10 V60 H60 V10 z" fill="#000"/></svg>`)

	parsed, err := Parse(svgData)
	if err != nil || parsed == nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(parsed.Paths) != 1 {
		t.Fatalf("Expected 1 path, got %d", len(parsed.Paths))
	}

	// M10 10 = (10, 10)
	// V60 = (10, 60)
	// H60 = (60, 60)
	// V10 = (60, 10)
	// z = back to (10, 10)
	expected := [][2]float64{{10, 10}, {10, 60}, {60, 60}, {60, 10}, {10, 10}}

	if len(parsed.Paths[0].Points) != len(expected) {
		t.Fatalf("Expected %d points, got %d", len(expected), len(parsed.Paths[0].Points))
	}

	for i, exp := range expected {
		got := parsed.Paths[0].Points[i]
		if got[0] != exp[0] || got[1] != exp[1] {
			t.Errorf("Point[%d] = %v, want %v", i, got, exp)
		}
	}
}

// TestMultiplePaths tests parsing SVG with multiple path elements
func TestMultiplePaths(t *testing.T) {
	svgData := []byte(`<svg viewBox="0 0 100 100"><path d="M0 0 L50 0 L50 50 L0 50 z" fill="#f00"/><path d="M50 50 L100 50 L100 100 L50 100 z" fill="#00f"/></svg>`)

	parsed, err := Parse(svgData)
	if err != nil || parsed == nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(parsed.Paths) != 2 {
		t.Fatalf("Expected 2 paths, got %d", len(parsed.Paths))
	}

	// First path should be red
	if parsed.Paths[0].FillColor != (color.RGBA{255, 0, 0, 255}) {
		t.Errorf("Path[0].FillColor = %v, want red", parsed.Paths[0].FillColor)
	}

	// Second path should be blue
	if parsed.Paths[1].FillColor != (color.RGBA{0, 0, 255, 255}) {
		t.Errorf("Path[1].FillColor = %v, want blue", parsed.Paths[1].FillColor)
	}
}
