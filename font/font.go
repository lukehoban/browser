// Package font provides font loading and text measurement utilities.
// This package is shared between layout and render to ensure consistent
// text dimensions during layout calculation and rendering.
//
// Spec references:
// - CSS 2.1 §15 Fonts
package font

import (
	"sync"

	"github.com/lukehoban/browser/css"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gobolditalic"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

var (
	// Global font cache to avoid reloading fonts
	// The Go fonts are embedded in the binary and always available
	goRegularFont    *opentype.Font
	goBoldFont       *opentype.Font
	goItalicFont     *opentype.Font
	goBoldItalicFont *opentype.Font
	fontOnce         sync.Once
	fontErr          error
)

// Style represents font styling properties.
// CSS 2.1 §15 Fonts
type Style struct {
	Size       float64 // Font size in pixels
	Weight     string  // Font weight: "normal" or "bold"
	Style      string  // Font style: "normal" or "italic"
	Decoration string  // Text decoration: "none" or "underline"
}

// LoadGoFonts loads the built-in Go fonts from the golang.org/x/image/font/gofont packages.
// These fonts are embedded in the binary and always available.
// The Go fonts are high-quality, open-source, sans-serif fonts designed for the Go project.
// See https://blog.golang.org/go-fonts for details.
// This function is safe to call multiple times - fonts are loaded only once.
func LoadGoFonts() error {
	fontOnce.Do(func() {
		var err error
		
		// Load Go Regular font (default)
		goRegularFont, err = opentype.Parse(goregular.TTF)
		if err != nil {
			fontErr = err
			return
		}
		
		// Load Go Bold font
		goBoldFont, err = opentype.Parse(gobold.TTF)
		if err != nil {
			fontErr = err
			return
		}
		
		// Load Go Italic font
		goItalicFont, err = opentype.Parse(goitalic.TTF)
		if err != nil {
			fontErr = err
			return
		}
		
		// Load Go Bold Italic font
		goBoldItalicFont, err = opentype.Parse(gobolditalic.TTF)
		if err != nil {
			fontErr = err
			return
		}
	})
	
	return fontErr
}

// SelectFont selects the appropriate font based on weight and style.
// Returns the selected font, or nil if fonts are not loaded.
func SelectFont(style Style) *opentype.Font {
	// Ensure fonts are loaded
	if err := LoadGoFonts(); err != nil {
		return nil
	}
	
	if style.Weight == "bold" && style.Style == "italic" {
		return goBoldItalicFont
	} else if style.Weight == "bold" {
		return goBoldFont
	} else if style.Style == "italic" {
		return goItalicFont
	}
	return goRegularFont
}

// MeasureText measures the dimensions of text using TrueType fonts.
// This provides accurate width and height for both layout calculations and rendering.
// Returns (width, height) in pixels.
//
// This is the single source of truth for text dimensions, ensuring layout and
// rendering use the same measurements.
func MeasureText(text string, style Style) (float64, float64) {
	if text == "" {
		return 0, 0
	}
	
	// Load and select the appropriate font
	selectedFont := SelectFont(style)
	if selectedFont == nil {
		// Fallback to basicfont dimensions with scaling
		face := basicfont.Face7x13
		scale := style.Size / css.BaseFontHeight
		width := float64(len(text)*face.Advance) * scale
		height := float64(face.Height) * scale
		return width, height
	}
	
	// Create face with proper DPI and size
	face, err := opentype.NewFace(selectedFont, &opentype.FaceOptions{
		Size:    style.Size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		// Fallback to basicfont dimensions with scaling
		basicFace := basicfont.Face7x13
		scale := style.Size / css.BaseFontHeight
		width := float64(len(text)*basicFace.Advance) * scale
		height := float64(basicFace.Height) * scale
		return width, height
	}
	defer face.Close()
	
	// Measure text using font drawer
	metrics := face.Metrics()
	drawer := &font.Drawer{
		Face: face,
	}
	
	width := drawer.MeasureString(text).Ceil()
	// Use line-height for height (ascent + descent gives the font's natural line height)
	// CSS 2.1 §10.8.1: line-height initial value is "normal", typically 1.2
	height := (metrics.Ascent + metrics.Descent).Ceil()
	
	return float64(width), float64(height)
}
