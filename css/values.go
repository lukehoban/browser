// Package css provides CSS tokenization, parsing, and value utilities.
//
// This file contains CSS value parsing utilities used across the browser.
//
// Spec references:
// - CSS 2.1 §4.3 Values: https://www.w3.org/TR/CSS21/syndata.html#values
// - CSS 2.1 §15.7 Font size: https://www.w3.org/TR/CSS21/fonts.html#font-size-props
package css

import (
	"strconv"
	"strings"
)

// BaseFontHeight is the default 'medium' font size in pixels.
// CSS 2.1 §15.7: The initial value of font-size is 'medium'
// Using basicfont.Face7x13 as the reference (13px height)
const BaseFontHeight = 13.0

// ParseFontSize parses a CSS font-size value and returns the size in pixels.
// CSS 2.1 §15.7 Font size: the 'font-size' property
// Supports:
// - Pixel values (e.g., "14px")
// - Point values (e.g., "10pt") - converted at 96 DPI
// - Named sizes (e.g., "small", "medium", "large")
// - Plain numbers (treated as pixels)
// Returns 0 if the value cannot be parsed.
func ParseFontSize(value string) float64 {
	value = strings.TrimSpace(strings.ToLower(value))

	if value == "" {
		return 0
	}

	// CSS 2.1 §4.3.2: Pixel values (e.g., "14px")
	if strings.HasSuffix(value, "px") {
		numStr := strings.TrimSuffix(value, "px")
		if size, err := strconv.ParseFloat(numStr, 64); err == nil && size > 0 {
			return size
		}
		return 0
	}

	// CSS 2.1 §4.3.2: Point values (1pt = 1/72 inch, at 96dpi = 1.333... pixels)
	if strings.HasSuffix(value, "pt") {
		numStr := strings.TrimSuffix(value, "pt")
		if size, err := strconv.ParseFloat(numStr, 64); err == nil && size > 0 {
			return size * 96.0 / 72.0 // Convert points to pixels at 96 DPI
		}
		return 0
	}

	// Plain number (treat as pixels)
	if size, err := strconv.ParseFloat(value, 64); err == nil && size > 0 {
		return size
	}

	// CSS 2.1 §15.7: Named sizes
	namedSizes := map[string]float64{
		"xx-small": 9.0,
		"x-small":  10.0,
		"small":    12.0,
		"medium":   BaseFontHeight,
		"large":    16.0,
		"x-large":  20.0,
		"xx-large": 24.0,
	}

	if size, ok := namedSizes[value]; ok {
		return size
	}

	return 0
}
