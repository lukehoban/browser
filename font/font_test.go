package font

import (
	"testing"
)

func TestLoadGoFonts(t *testing.T) {
	err := LoadGoFonts()
	if err != nil {
		t.Fatalf("LoadGoFonts() failed: %v", err)
	}

	// Test that calling LoadGoFonts multiple times doesn't cause errors
	err = LoadGoFonts()
	if err != nil {
		t.Fatalf("LoadGoFonts() second call failed: %v", err)
	}

	// Verify fonts are loaded
	if goRegularFont == nil {
		t.Error("goRegularFont should not be nil after LoadGoFonts")
	}
	if goBoldFont == nil {
		t.Error("goBoldFont should not be nil after LoadGoFonts")
	}
	if goItalicFont == nil {
		t.Error("goItalicFont should not be nil after LoadGoFonts")
	}
	if goBoldItalicFont == nil {
		t.Error("goBoldItalicFont should not be nil after LoadGoFonts")
	}
}

func TestSelectFont(t *testing.T) {
	tests := []struct {
		name     string
		style    Style
		wantNil  bool
	}{
		{
			name: "regular_font",
			style: Style{
				Size:   14.0,
				Weight: "normal",
				Style:  "normal",
			},
			wantNil: false,
		},
		{
			name: "bold_font",
			style: Style{
				Size:   14.0,
				Weight: "bold",
				Style:  "normal",
			},
			wantNil: false,
		},
		{
			name: "italic_font",
			style: Style{
				Size:   14.0,
				Weight: "normal",
				Style:  "italic",
			},
			wantNil: false,
		},
		{
			name: "bold_italic_font",
			style: Style{
				Size:   14.0,
				Weight: "bold",
				Style:  "italic",
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			font := SelectFont(tt.style)
			if tt.wantNil && font != nil {
				t.Errorf("SelectFont() = %v, expected nil", font)
			}
			if !tt.wantNil && font == nil {
				t.Errorf("SelectFont() = nil, expected non-nil font")
			}
		})
	}
}

func TestMeasureText(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		style      Style
		wantZero   bool
	}{
		{
			name:     "empty_string",
			text:     "",
			style:    Style{Size: 14.0, Weight: "normal", Style: "normal"},
			wantZero: true,
		},
		{
			name:     "regular_text",
			text:     "Hello, World!",
			style:    Style{Size: 14.0, Weight: "normal", Style: "normal"},
			wantZero: false,
		},
		{
			name:     "bold_text",
			text:     "Bold Text",
			style:    Style{Size: 16.0, Weight: "bold", Style: "normal"},
			wantZero: false,
		},
		{
			name:     "italic_text",
			text:     "Italic Text",
			style:    Style{Size: 12.0, Weight: "normal", Style: "italic"},
			wantZero: false,
		},
		{
			name:     "bold_italic_text",
			text:     "Bold Italic",
			style:    Style{Size: 18.0, Weight: "bold", Style: "italic"},
			wantZero: false,
		},
		{
			name:     "small_font",
			text:     "Small",
			style:    Style{Size: 10.0, Weight: "normal", Style: "normal"},
			wantZero: false,
		},
		{
			name:     "large_font",
			text:     "Large",
			style:    Style{Size: 24.0, Weight: "normal", Style: "normal"},
			wantZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, height := MeasureText(tt.text, tt.style)

			if tt.wantZero {
				if width != 0 || height != 0 {
					t.Errorf("MeasureText(%q, %v) = (%v, %v), expected (0, 0)", tt.text, tt.style, width, height)
				}
			} else {
				if width <= 0 {
					t.Errorf("MeasureText(%q, %v) width = %v, expected > 0", tt.text, tt.style, width)
				}
				if height <= 0 {
					t.Errorf("MeasureText(%q, %v) height = %v, expected > 0", tt.text, tt.style, height)
				}
			}
		})
	}
}

func TestMeasureTextSizeScaling(t *testing.T) {
	// Test that text measured at different sizes scales appropriately
	text := "Test"
	style1 := Style{Size: 10.0, Weight: "normal", Style: "normal"}
	style2 := Style{Size: 20.0, Weight: "normal", Style: "normal"}

	width1, height1 := MeasureText(text, style1)
	width2, height2 := MeasureText(text, style2)

	// Larger font should produce larger dimensions
	if width2 <= width1 {
		t.Errorf("MeasureText with larger font should have larger width: %v vs %v", width2, width1)
	}
	if height2 <= height1 {
		t.Errorf("MeasureText with larger font should have larger height: %v vs %v", height2, height1)
	}
}

func TestMeasureTextConsistency(t *testing.T) {
	// Test that measuring the same text multiple times gives the same result
	text := "Consistency Test"
	style := Style{Size: 14.0, Weight: "normal", Style: "normal"}

	width1, height1 := MeasureText(text, style)
	width2, height2 := MeasureText(text, style)

	if width1 != width2 {
		t.Errorf("MeasureText should be consistent: width %v != %v", width1, width2)
	}
	if height1 != height2 {
		t.Errorf("MeasureText should be consistent: height %v != %v", height1, height2)
	}
}
