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
		expected [4]uint8 // R, G, B, A
	}{
		// CSS 2.1 Basic colors
		{"black", "black", [4]uint8{0, 0, 0, 255}},
		{"white", "white", [4]uint8{255, 255, 255, 255}},
		{"red", "red", [4]uint8{255, 0, 0, 255}},
		{"green", "green", [4]uint8{0, 128, 0, 255}},
		{"blue", "blue", [4]uint8{0, 0, 255, 255}},
		{"yellow", "yellow", [4]uint8{255, 255, 0, 255}},
		{"cyan", "cyan", [4]uint8{0, 255, 255, 255}},
		{"magenta", "magenta", [4]uint8{255, 0, 255, 255}},
		{"silver", "silver", [4]uint8{192, 192, 192, 255}},
		{"gray", "gray", [4]uint8{128, 128, 128, 255}},
		{"grey", "grey", [4]uint8{128, 128, 128, 255}},
		{"maroon", "maroon", [4]uint8{128, 0, 0, 255}},
		{"purple", "purple", [4]uint8{128, 0, 128, 255}},
		{"lime", "lime", [4]uint8{0, 255, 0, 255}},
		{"olive", "olive", [4]uint8{128, 128, 0, 255}},
		{"navy", "navy", [4]uint8{0, 0, 128, 255}},
		{"teal", "teal", [4]uint8{0, 128, 128, 255}},
		{"orange", "orange", [4]uint8{255, 165, 0, 255}},

		// Extended colors (sample)
		{"lightgray", "lightgray", [4]uint8{211, 211, 211, 255}},
		{"darkgray", "darkgray", [4]uint8{169, 169, 169, 255}},
		{"pink", "pink", [4]uint8{255, 192, 203, 255}},
		{"brown", "brown", [4]uint8{165, 42, 42, 255}},
		{"coral", "coral", [4]uint8{255, 127, 80, 255}},

		// Hex colors - 6 digit
		{"hex_000000", "#000000", [4]uint8{0, 0, 0, 255}},
		{"hex_ffffff", "#ffffff", [4]uint8{255, 255, 255, 255}},
		{"hex_ff0000", "#ff0000", [4]uint8{255, 0, 0, 255}},
		{"hex_00ff00", "#00ff00", [4]uint8{0, 255, 0, 255}},
		{"hex_0000ff", "#0000ff", [4]uint8{0, 0, 255, 255}},
		{"hex_123456", "#123456", [4]uint8{0x12, 0x34, 0x56, 255}},
		{"hex_abcdef", "#abcdef", [4]uint8{0xab, 0xcd, 0xef, 255}},

		// Hex colors - 3 digit (shorthand)
		{"hex_000", "#000", [4]uint8{0, 0, 0, 255}},
		{"hex_fff", "#fff", [4]uint8{255, 255, 255, 255}},
		{"hex_f00", "#f00", [4]uint8{255, 0, 0, 255}},
		{"hex_0f0", "#0f0", [4]uint8{0, 255, 0, 255}},
		{"hex_00f", "#00f", [4]uint8{0, 0, 255, 255}},
		{"hex_abc", "#abc", [4]uint8{0xaa, 0xbb, 0xcc, 255}},

		// Case insensitivity
		{"uppercase_RED", "RED", [4]uint8{255, 0, 0, 255}},
		{"uppercase_BLUE", "BLUE", [4]uint8{0, 0, 255, 255}},
		{"mixed_BlUe", "BlUe", [4]uint8{0, 0, 255, 255}},

		// Whitespace handling
		{"whitespace_before", "  red", [4]uint8{255, 0, 0, 255}},
		{"whitespace_after", "blue  ", [4]uint8{0, 0, 255, 255}},
		{"whitespace_both", "  green  ", [4]uint8{0, 128, 0, 255}},

		// Unknown colors default to black
		{"unknown", "unknowncolor", [4]uint8{0, 0, 0, 255}},
		{"empty", "", [4]uint8{0, 0, 0, 255}},
		{"invalid_hex", "#gg0000", [4]uint8{0, 0, 0, 255}},
		{"invalid_hex_length", "#12345", [4]uint8{0, 0, 0, 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseColor(tt.input)
			if result.R != tt.expected[0] || result.G != tt.expected[1] ||
			   result.B != tt.expected[2] || result.A != tt.expected[3] {
				t.Errorf("ParseColor(%q) = RGBA(%d, %d, %d, %d), expected RGBA(%d, %d, %d, %d)",
					tt.input, result.R, result.G, result.B, result.A,
					tt.expected[0], tt.expected[1], tt.expected[2], tt.expected[3])
			}
		})
	}
}
