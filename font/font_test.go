package font

import (
	"testing"

	"golang.org/x/image/font/opentype"
)

func TestLoadGoFonts(t *testing.T) {
	// Test that fonts load successfully
	err := LoadGoFonts()
	if err != nil {
		t.Fatalf("LoadGoFonts() failed: %v", err)
	}

	// Call again to test idempotency
	err = LoadGoFonts()
	if err != nil {
		t.Errorf("LoadGoFonts() second call failed: %v", err)
	}

	// Verify fonts are loaded
	if goRegularFont == nil {
		t.Error("goRegularFont is nil after LoadGoFonts()")
	}
	if goBoldFont == nil {
		t.Error("goBoldFont is nil after LoadGoFonts()")
	}
	if goItalicFont == nil {
		t.Error("goItalicFont is nil after LoadGoFonts()")
	}
	if goBoldItalicFont == nil {
		t.Error("goBoldItalicFont is nil after LoadGoFonts()")
	}
}

func TestSelectFont(t *testing.T) {
	tests := []struct {
		name     string
		style    Style
		wantFont string // "regular", "bold", "italic", "bolditalic"
	}{
		{
			name:     "regular font",
			style:    Style{Weight: "normal", Style: "normal"},
			wantFont: "regular",
		},
		{
			name:     "bold font",
			style:    Style{Weight: "bold", Style: "normal"},
			wantFont: "bold",
		},
		{
			name:     "italic font",
			style:    Style{Weight: "normal", Style: "italic"},
			wantFont: "italic",
		},
		{
			name:     "bold italic font",
			style:    Style{Weight: "bold", Style: "italic"},
			wantFont: "bolditalic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			font := SelectFont(tt.style)
			if font == nil {
				t.Fatalf("SelectFont() returned nil")
			}

			// Verify we got the correct font by comparing pointers
			var expectedFont *opentype.Font
			switch tt.wantFont {
			case "regular":
				expectedFont = goRegularFont
			case "bold":
				expectedFont = goBoldFont
			case "italic":
				expectedFont = goItalicFont
			case "bolditalic":
				expectedFont = goBoldItalicFont
			}

			if font != expectedFont {
				t.Errorf("SelectFont() returned wrong font, got %p, want %p", font, expectedFont)
			}
		})
	}
}

func TestMeasureText(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		style  Style
		wantW  bool // true if width should be > 0
		wantH  bool // true if height should be > 0
	}{
		{
			name:  "empty string",
			text:  "",
			style: Style{Size: 13.0, Weight: "normal", Style: "normal"},
			wantW: false,
			wantH: false,
		},
		{
			name:  "simple text",
			text:  "Hello",
			style: Style{Size: 13.0, Weight: "normal", Style: "normal"},
			wantW: true,
			wantH: true,
		},
		{
			name:  "bold text",
			text:  "Bold",
			style: Style{Size: 16.0, Weight: "bold", Style: "normal"},
			wantW: true,
			wantH: true,
		},
		{
			name:  "italic text",
			text:  "Italic",
			style: Style{Size: 14.0, Weight: "normal", Style: "italic"},
			wantW: true,
			wantH: true,
		},
		{
			name:  "large text",
			text:  "Large",
			style: Style{Size: 24.0, Weight: "normal", Style: "normal"},
			wantW: true,
			wantH: true,
		},
		{
			name:  "single character",
			text:  "X",
			style: Style{Size: 13.0, Weight: "normal", Style: "normal"},
			wantW: true,
			wantH: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, height := MeasureText(tt.text, tt.style)

			if tt.wantW && width <= 0 {
				t.Errorf("MeasureText(%q) width = %v, expected > 0", tt.text, width)
			}
			if !tt.wantW && width != 0 {
				t.Errorf("MeasureText(%q) width = %v, expected 0", tt.text, width)
			}
			if tt.wantH && height <= 0 {
				t.Errorf("MeasureText(%q) height = %v, expected > 0", tt.text, height)
			}
			if !tt.wantH && height != 0 {
				t.Errorf("MeasureText(%q) height = %v, expected 0", tt.text, height)
			}

			// Verify that larger font size produces larger dimensions for non-empty text
			if tt.text != "" && tt.style.Size > 0 {
				smallStyle := Style{Size: 10.0, Weight: "normal", Style: "normal"}
				largeStyle := Style{Size: 20.0, Weight: "normal", Style: "normal"}

				smallW, smallH := MeasureText(tt.text, smallStyle)
				largeW, largeH := MeasureText(tt.text, largeStyle)

				if largeW <= smallW {
					t.Errorf("Expected larger font to have larger width, got small=%v large=%v", smallW, largeW)
				}
				if largeH <= smallH {
					t.Errorf("Expected larger font to have larger height, got small=%v large=%v", smallH, largeH)
				}
			}
		})
	}
}

func TestMeasureTextConsistency(t *testing.T) {
	// Test that measuring the same text multiple times gives consistent results
	text := "Test Text"
	style := Style{Size: 14.0, Weight: "normal", Style: "normal"}

	w1, h1 := MeasureText(text, style)
	w2, h2 := MeasureText(text, style)

	if w1 != w2 || h1 != h2 {
		t.Errorf("MeasureText() inconsistent results: (%v,%v) vs (%v,%v)", w1, h1, w2, h2)
	}
}

func TestMeasureTextLengthScaling(t *testing.T) {
	// Test that longer text has larger width
	shortText := "Hi"
	longText := "Hello World"
	style := Style{Size: 13.0, Weight: "normal", Style: "normal"}

	shortW, _ := MeasureText(shortText, style)
	longW, _ := MeasureText(longText, style)

	if longW <= shortW {
		t.Errorf("Expected longer text to have larger width, got short=%v long=%v", shortW, longW)
	}
}
