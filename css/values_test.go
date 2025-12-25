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
