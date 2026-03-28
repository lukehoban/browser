package css

import "testing"

func TestParseFontSize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		// Pixel values
		{"pixels_14", "14px", 14.0},
		{"pixels_20", "20px", 20.0},
		{"pixels_zero", "0px", 0.0}, // 0 is not > 0
		
		// Point values (1pt = 96/72 pixels at 96 DPI)
		{"points_10", "10pt", 10.0 * 96.0 / 72.0},
		{"points_12", "12pt", 12.0 * 96.0 / 72.0},
		{"points_7", "7pt", 7.0 * 96.0 / 72.0},
		{"points_8.5", "8.5pt", 8.5 * 96.0 / 72.0},
		
		// Plain numbers (treated as pixels)
		{"number_10", "10", 10.0},
		{"number_24", "24", 24.0},
		
		// Named sizes
		{"named_xx-small", "xx-small", 9.0},
		{"named_x-small", "x-small", 10.0},
		{"named_small", "small", 12.0},
		{"named_medium", "medium", 13.0},
		{"named_large", "large", 16.0},
		{"named_x-large", "x-large", 20.0},
		{"named_xx-large", "xx-large", 24.0},
		
		// Case insensitivity
		{"uppercase_PX", "14PX", 14.0},
		{"uppercase_PT", "10PT", 10.0 * 96.0 / 72.0},
		{"mixed_Medium", "Medium", 13.0},
		
		// Invalid values
		{"invalid_text", "invalid", 0.0},
		{"empty_string", "", 0.0},
		{"negative", "-10px", 0.0}, // negative not > 0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseFontSize(tt.input)
			if result != tt.expected {
				t.Errorf("ParseFontSize(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBaseFontHeight(t *testing.T) {
	// Ensure the constant is set correctly
	if BaseFontHeight != 13.0 {
		t.Errorf("BaseFontHeight = %v, expected 13.0", BaseFontHeight)
	}
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		r, g, b  uint8
	}{
		// CSS 2.1 basic named colors
		{"black", "black", 0, 0, 0},
		{"white", "white", 255, 255, 255},
		{"red", "red", 255, 0, 0},
		{"green", "green", 0, 128, 0},
		{"blue", "blue", 0, 0, 255},
		{"yellow", "yellow", 255, 255, 0},
		{"navy", "navy", 0, 0, 128},
		{"purple", "purple", 128, 0, 128},
		{"silver", "silver", 192, 192, 192},
		{"gray", "gray", 128, 128, 128},
		{"grey", "grey", 128, 128, 128},
		{"maroon", "maroon", 128, 0, 0},
		{"olive", "olive", 128, 128, 0},
		{"teal", "teal", 0, 128, 128},
		{"lime", "lime", 0, 255, 0},
		{"orange", "orange", 255, 165, 0},
		{"fuchsia", "fuchsia", 255, 0, 255},
		{"magenta", "magenta", 255, 0, 255},
		{"aqua", "aqua", 0, 255, 255},
		{"cyan", "cyan", 0, 255, 255},
		// Extended colors
		{"lightgray", "lightgray", 211, 211, 211},
		{"lightgrey", "lightgrey", 211, 211, 211},
		{"darkgray", "darkgray", 169, 169, 169},
		{"darkgreen", "darkgreen", 0, 100, 0},
		{"pink", "pink", 255, 192, 203},
		{"gold", "gold", 255, 215, 0},
		{"brown", "brown", 165, 42, 42},
		{"coral", "coral", 255, 127, 80},
		{"crimson", "crimson", 220, 20, 60},
		{"indigo", "indigo", 75, 0, 130},
		// Case insensitivity
		{"uppercase RED", "RED", 255, 0, 0},
		{"mixed Blue", "Blue", 0, 0, 255},
		// Hex colors - 6 digit
		{"#FF0000", "#FF0000", 255, 0, 0},
		{"#00FF00", "#00FF00", 0, 255, 0},
		{"#0000FF", "#0000FF", 0, 0, 255},
		{"#FFFFFF", "#FFFFFF", 255, 255, 255},
		{"#000000", "#000000", 0, 0, 0},
		{"#2196F3", "#2196F3", 33, 150, 243},
		// Hex colors - 3 digit shorthand
		{"#f00", "#f00", 255, 0, 0},
		{"#0f0", "#0f0", 0, 255, 0},
		{"#00f", "#00f", 0, 0, 255},
		{"#fff", "#fff", 255, 255, 255},
		{"#000", "#000", 0, 0, 0},
		// Unknown defaults to black
		{"unknown", "unknown", 0, 0, 0},
		{"empty", "", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseColor(tt.input)
			if result.R != tt.r || result.G != tt.g || result.B != tt.b {
				t.Errorf("ParseColor(%q) = {R:%d,G:%d,B:%d}, expected {R:%d,G:%d,B:%d}",
					tt.input, result.R, result.G, result.B, tt.r, tt.g, tt.b)
			}
			if result.A != 255 {
				t.Errorf("ParseColor(%q) alpha = %d, expected 255", tt.input, result.A)
			}
		})
	}
}

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		r, g, b uint8
	}{
		{"6-digit red", "#FF0000", 255, 0, 0},
		{"6-digit green", "#00FF00", 0, 255, 0},
		{"6-digit blue", "#0000FF", 0, 0, 255},
		{"6-digit mixed", "#4CAF50", 76, 175, 80},
		{"3-digit red", "#f00", 255, 0, 0},
		{"3-digit green", "#0f0", 0, 255, 0},
		{"3-digit blue", "#00f", 0, 0, 255},
		{"3-digit white", "#fff", 255, 255, 255},
		{"3-digit black", "#000", 0, 0, 0},
		{"3-digit gray", "#999", 153, 153, 153},
		// Invalid length returns zero color
		{"invalid length", "#1234", 0, 0, 0},
		{"empty after #", "#", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseHexColor(tt.input)
			if result.R != tt.r || result.G != tt.g || result.B != tt.b {
				t.Errorf("parseHexColor(%q) = {R:%d,G:%d,B:%d}, expected {R:%d,G:%d,B:%d}",
					tt.input, result.R, result.G, result.B, tt.r, tt.g, tt.b)
			}
		})
	}
}
