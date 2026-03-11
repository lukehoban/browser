package css

import (
	"image/color"
	"testing"
)

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
		expected color.RGBA
	}{
		// CSS 2.1 basic named colors
		{"named_black", "black", color.RGBA{0, 0, 0, 255}},
		{"named_white", "white", color.RGBA{255, 255, 255, 255}},
		{"named_red", "red", color.RGBA{255, 0, 0, 255}},
		{"named_green", "green", color.RGBA{0, 128, 0, 255}},
		{"named_blue", "blue", color.RGBA{0, 0, 255, 255}},
		{"named_orange", "orange", color.RGBA{255, 165, 0, 255}},
		{"named_silver", "silver", color.RGBA{192, 192, 192, 255}},
		{"named_gray", "gray", color.RGBA{128, 128, 128, 255}},
		{"named_grey", "grey", color.RGBA{128, 128, 128, 255}},
		{"named_maroon", "maroon", color.RGBA{128, 0, 0, 255}},
		{"named_purple", "purple", color.RGBA{128, 0, 128, 255}},
		{"named_fuchsia", "fuchsia", color.RGBA{255, 0, 255, 255}},
		{"named_lime", "lime", color.RGBA{0, 255, 0, 255}},
		{"named_olive", "olive", color.RGBA{128, 128, 0, 255}},
		{"named_yellow", "yellow", color.RGBA{255, 255, 0, 255}},
		{"named_navy", "navy", color.RGBA{0, 0, 128, 255}},
		{"named_teal", "teal", color.RGBA{0, 128, 128, 255}},
		{"named_aqua", "aqua", color.RGBA{0, 255, 255, 255}},

		// Extended named colors
		{"named_coral", "coral", color.RGBA{255, 127, 80, 255}},
		{"named_tomato", "tomato", color.RGBA{255, 99, 71, 255}},
		{"named_lightgray", "lightgray", color.RGBA{211, 211, 211, 255}},
		{"named_darkblue", "darkblue", color.RGBA{0, 0, 139, 255}},

		// Case insensitivity
		{"case_Black", "Black", color.RGBA{0, 0, 0, 255}},
		{"case_RED", "RED", color.RGBA{255, 0, 0, 255}},
		{"case_WhItE", "WhItE", color.RGBA{255, 255, 255, 255}},

		// Hex colors - 6 digit
		{"hex6_black", "#000000", color.RGBA{0, 0, 0, 255}},
		{"hex6_white", "#ffffff", color.RGBA{255, 255, 255, 255}},
		{"hex6_red", "#ff0000", color.RGBA{255, 0, 0, 255}},
		{"hex6_green", "#00ff00", color.RGBA{0, 255, 0, 255}},
		{"hex6_blue", "#0000ff", color.RGBA{0, 0, 255, 255}},
		{"hex6_mixed", "#336699", color.RGBA{51, 102, 153, 255}},
		{"hex6_upper", "#FF0000", color.RGBA{255, 0, 0, 255}},

		// Hex colors - 3 digit
		{"hex3_black", "#000", color.RGBA{0, 0, 0, 255}},
		{"hex3_white", "#fff", color.RGBA{255, 255, 255, 255}},
		{"hex3_red", "#f00", color.RGBA{255, 0, 0, 255}},
		{"hex3_mixed", "#369", color.RGBA{51, 102, 153, 255}},

		// Whitespace handling
		{"whitespace_leading", "  red  ", color.RGBA{255, 0, 0, 255}},

		// Invalid/unknown values default to black
		{"invalid_unknown", "unknowncolor", color.RGBA{0, 0, 0, 255}},
		{"invalid_empty", "", color.RGBA{0, 0, 0, 255}},
		{"invalid_hash_only", "#", color.RGBA{0, 0, 0, 255}},
		{"invalid_hex_2", "#ab", color.RGBA{0, 0, 0, 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseColor(tt.input)
			if result != tt.expected {
				t.Errorf("ParseColor(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected color.RGBA
	}{
		{"full_6digit", "#abcdef", color.RGBA{171, 205, 239, 255}},
		{"full_3digit", "#ace", color.RGBA{170, 204, 238, 255}},
		{"no_hash_6", "abcdef", color.RGBA{171, 205, 239, 255}},
		{"no_hash_3", "ace", color.RGBA{170, 204, 238, 255}},
		{"invalid_len_4", "#abcd", color.RGBA{0, 0, 0, 255}},
		{"invalid_len_1", "#a", color.RGBA{0, 0, 0, 255}},
		{"empty", "", color.RGBA{0, 0, 0, 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseHexColor(tt.input)
			if result != tt.expected {
				t.Errorf("parseHexColor(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseFontSizeEm(t *testing.T) {
	// Additional edge cases for ParseFontSize
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"whitespace_around", "  14px  ", 14.0},
		{"zero_no_unit", "0", 0.0},
		{"fractional_px", "12.5px", 12.5},
		{"fractional_pt", "10.5pt", 10.5 * 96.0 / 72.0},
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
