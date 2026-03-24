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
		expectedR uint8
		expectedG uint8
		expectedB uint8
		expectedA uint8
	}{
		// Basic CSS 2.1 named colors
		{"black", "black", 0, 0, 0, 255},
		{"white", "white", 255, 255, 255, 255},
		{"red", "red", 255, 0, 0, 255},
		{"green", "green", 0, 128, 0, 255},
		{"blue", "blue", 0, 0, 255, 255},
		{"yellow", "yellow", 255, 255, 0, 255},
		{"cyan", "cyan", 0, 255, 255, 255},
		{"magenta", "magenta", 255, 0, 255, 255},
		{"orange", "orange", 255, 165, 0, 255},

		// Gray variants
		{"gray", "gray", 128, 128, 128, 255},
		{"grey", "grey", 128, 128, 128, 255},
		{"silver", "silver", 192, 192, 192, 255},
		{"lightgray", "lightgray", 211, 211, 211, 255},
		{"darkgray", "darkgray", 169, 169, 169, 255},

		// Extended colors
		{"hotpink", "hotpink", 255, 105, 180, 255},
		{"cornflowerblue", "cornflowerblue", 100, 149, 237, 255},
		{"lavender", "lavender", 230, 230, 250, 255},

		// Hex colors - 6 digit
		{"hex_red", "#ff0000", 255, 0, 0, 255},
		{"hex_green", "#00ff00", 0, 255, 0, 255},
		{"hex_blue", "#0000ff", 0, 0, 255, 255},
		{"hex_white", "#ffffff", 255, 255, 255, 255},
		{"hex_black", "#000000", 0, 0, 0, 255},
		{"hex_custom", "#123456", 0x12, 0x34, 0x56, 255},

		// Hex colors - 3 digit
		{"hex3_red", "#f00", 255, 0, 0, 255},
		{"hex3_green", "#0f0", 0, 255, 0, 255},
		{"hex3_blue", "#00f", 0, 0, 255, 255},
		{"hex3_white", "#fff", 255, 255, 255, 255},
		{"hex3_black", "#000", 0, 0, 0, 255},
		{"hex3_custom", "#abc", 0xaa, 0xbb, 0xcc, 255},

		// Case insensitivity
		{"uppercase_RED", "RED", 255, 0, 0, 255},
		{"uppercase_Blue", "Blue", 0, 0, 255, 255},
		{"uppercase_HEX", "#FF0000", 255, 0, 0, 255},
		{"mixedcase", "HotPink", 255, 105, 180, 255},

		// Whitespace handling
		{"leading_space", "  red", 255, 0, 0, 255},
		{"trailing_space", "blue  ", 0, 0, 255, 255},
		{"both_spaces", "  green  ", 0, 128, 0, 255},

		// Invalid/unknown colors (should default to black)
		{"unknown", "unknowncolor", 0, 0, 0, 255},
		{"empty", "", 0, 0, 0, 255},
		{"invalid_hex", "#gggggg", 0, 0, 0, 255},
		{"short_hex", "#ff", 0, 0, 0, 255},
		{"long_hex", "#fffffff", 0, 0, 0, 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseColor(tt.input)
			if result.R != tt.expectedR || result.G != tt.expectedG ||
			   result.B != tt.expectedB || result.A != tt.expectedA {
				t.Errorf("ParseColor(%q) = RGBA{%d, %d, %d, %d}, expected RGBA{%d, %d, %d, %d}",
					tt.input, result.R, result.G, result.B, result.A,
					tt.expectedR, tt.expectedG, tt.expectedB, tt.expectedA)
			}
		})
	}
}

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expectedR uint8
		expectedG uint8
		expectedB uint8
	}{
		// 6-digit hex colors
		{"red_6digit", "#ff0000", 255, 0, 0},
		{"green_6digit", "#00ff00", 0, 255, 0},
		{"blue_6digit", "#0000ff", 0, 0, 255},
		{"custom_6digit", "#a1b2c3", 0xa1, 0xb2, 0xc3},

		// 3-digit hex colors
		{"red_3digit", "#f00", 255, 0, 0},
		{"green_3digit", "#0f0", 0, 255, 0},
		{"blue_3digit", "#00f", 0, 0, 255},
		{"custom_3digit", "#abc", 0xaa, 0xbb, 0xcc},

		// Edge cases
		{"all_zero", "#000", 0, 0, 0},
		{"all_f", "#fff", 255, 255, 255},
		{"without_hash", "ff0000", 255, 0, 0},

		// Invalid lengths (should return 0, 0, 0)
		{"too_short", "#f", 0, 0, 0},
		{"too_long", "#fffffff", 0, 0, 0},
		{"four_digit", "#ffff", 0, 0, 0},
		{"empty", "", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseHexColor(tt.input)
			if result.R != tt.expectedR || result.G != tt.expectedG || result.B != tt.expectedB {
				t.Errorf("parseHexColor(%q) = RGB{%d, %d, %d}, expected RGB{%d, %d, %d}",
					tt.input, result.R, result.G, result.B,
					tt.expectedR, tt.expectedG, tt.expectedB)
			}
			// Alpha should always be 255
			if result.A != 255 {
				t.Errorf("parseHexColor(%q) alpha = %d, expected 255", tt.input, result.A)
			}
		})
	}
}
