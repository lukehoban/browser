// Package render implements font management and text rendering.
// Supports basic CSS 2.1 font properties: font-family, font-size, font-weight.
//
// Spec references:
// - CSS 2.1 §15 Fonts: https://www.w3.org/TR/CSS21/fonts.html
// - CSS 2.1 §15.3 Font family: https://www.w3.org/TR/CSS21/fonts.html#font-family-prop
// - CSS 2.1 §15.7 Font size: https://www.w3.org/TR/CSS21/fonts.html#font-size-prop
package render

import (
	"strconv"
	"strings"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// FontManager manages font loading and caching.
type FontManager struct {
	cache map[string]font.Face
	mu    sync.RWMutex
}

// NewFontManager creates a new font manager.
func NewFontManager() *FontManager {
	return &FontManager{
		cache: make(map[string]font.Face),
	}
}

// GetFace returns a font face for the given family, size, and weight.
// CSS 2.1 §15.3 Font family and §15.7 Font size
func (fm *FontManager) GetFace(family string, size float64, weight string) font.Face {
	// Create a cache key
	key := family + ":" + strconv.FormatFloat(size, 'f', 1, 64) + ":" + weight

	// Check cache first
	fm.mu.RLock()
	if face, ok := fm.cache[key]; ok {
		fm.mu.RUnlock()
		return face
	}
	fm.mu.RUnlock()

	// Parse the font
	var ttfData []byte
	familyLower := strings.ToLower(strings.TrimSpace(family))

	// CSS 2.1 §15.3: Font family matching
	// Support generic font families: serif, sans-serif, monospace
	switch familyLower {
	case "sans-serif", "arial", "helvetica", "":
		// Default to sans-serif
		if weight == "bold" || weight == "700" || weight == "800" || weight == "900" {
			ttfData = gobold.TTF
		} else {
			ttfData = goregular.TTF
		}
	case "monospace", "courier", "courier new":
		ttfData = gomono.TTF
	case "serif", "times", "times new roman":
		// For now, fall back to sans-serif for serif fonts
		// Could add goserif package if available
		ttfData = goregular.TTF
	default:
		// Unknown font family, default to sans-serif
		ttfData = goregular.TTF
	}

	// Parse TTF data
	f, err := opentype.Parse(ttfData)
	if err != nil {
		// Fall back to a default face on error
		return nil
	}

	// Create face with specified size
	// CSS 2.1 §15.7: Font size in CSS points (1 point = 1/72 inch)
	// We use DPI of 72 to match CSS points to pixels for simplicity
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size: size,
		DPI:  72,
	})
	if err != nil {
		return nil
	}

	// Cache the face
	fm.mu.Lock()
	fm.cache[key] = face
	fm.mu.Unlock()

	return face
}

// GetFaceMetrics returns metrics for a font face.
// CSS 2.1 §15.5: Font metrics
func GetFaceMetrics(face font.Face) font.Metrics {
	return face.Metrics()
}

// MeasureString measures the width of a string in the given font face.
// CSS 2.1 §16.2: Alignment properties and §10.3.1: Inline content width
func MeasureString(face font.Face, text string) fixed.Int26_6 {
	var width fixed.Int26_6
	for _, r := range text {
		advance, ok := face.GlyphAdvance(r)
		if !ok {
			// Use a default advance if glyph not found
			advance = face.Metrics().Height / 2
		}
		width += advance
	}
	return width
}

// ParseFontSize parses a CSS font-size value and returns the size in pixels.
// CSS 2.1 §15.7 Font size: https://www.w3.org/TR/CSS21/fonts.html#font-size-prop
// Supports: px, pt, em, named sizes (small, medium, large, etc.)
func ParseFontSize(value string, parentSize float64) float64 {
	value = strings.TrimSpace(strings.ToLower(value))

	// Default font size if not specified
	// CSS 2.1 §15.7: Medium is the initial value
	if value == "" || value == "medium" {
		return 16.0
	}

	// Named font sizes
	// CSS 2.1 §15.7: Absolute size keywords
	switch value {
	case "xx-small":
		return 9.0
	case "x-small":
		return 10.0
	case "small":
		return 13.0
	case "large":
		return 18.0
	case "x-large":
		return 24.0
	case "xx-large":
		return 32.0
	case "smaller":
		// CSS 2.1 §15.7: Relative size keywords
		return parentSize * 0.83
	case "larger":
		return parentSize * 1.2
	}

	// Pixel values
	if strings.HasSuffix(value, "px") {
		if size, err := strconv.ParseFloat(value[:len(value)-2], 64); err == nil {
			return size
		}
	}

	// Point values (1pt = 1.333px at 96 DPI, but we use 1:1 for simplicity)
	// CSS 2.1 §4.3.2: pt is a physical unit
	if strings.HasSuffix(value, "pt") {
		if size, err := strconv.ParseFloat(value[:len(value)-2], 64); err == nil {
			// Convert points to pixels (72 points = 96 pixels at standard DPI)
			return size * 96.0 / 72.0
		}
	}

	// Em values (relative to parent font size)
	// CSS 2.1 §4.3.2: em is relative to font size
	if strings.HasSuffix(value, "em") {
		if size, err := strconv.ParseFloat(value[:len(value)-2], 64); err == nil {
			return size * parentSize
		}
	}

	// Plain number (assume pixels)
	if size, err := strconv.ParseFloat(value, 64); err == nil {
		return size
	}

	// Default to 16px if parsing fails
	return 16.0
}
