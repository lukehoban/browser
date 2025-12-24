package render

import (
	"testing"
)

func TestParseFontSize(t *testing.T) {
	tests := []struct {
		input      string
		parentSize float64
		expected   float64
	}{
		{"", 16.0, 16.0},                  // Default
		{"medium", 16.0, 16.0},            // Named size
		{"16px", 16.0, 16.0},              // Pixels
		{"12pt", 16.0, 16.0},              // Points (12pt ≈ 16px at 96 DPI)
		{"1.5em", 16.0, 24.0},             // Em (relative to parent)
		{"2em", 10.0, 20.0},               // Em (relative to parent)
		{"small", 16.0, 13.0},             // Named size
		{"large", 16.0, 18.0},             // Named size
		{"x-large", 16.0, 24.0},           // Named size
		{"xx-large", 16.0, 32.0},          // Named size
		{"smaller", 16.0, 13.28},          // Relative (0.83 * 16)
		{"larger", 16.0, 19.2},            // Relative (1.2 * 16)
		{"20", 16.0, 20.0},                // Plain number
		{"invalid", 16.0, 16.0},           // Invalid defaults to 16
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseFontSize(tt.input, tt.parentSize)
			// Allow small floating point differences
			if result < tt.expected-0.1 || result > tt.expected+0.1 {
				t.Errorf("ParseFontSize(%q, %.1f) = %.2f, expected %.2f",
					tt.input, tt.parentSize, result, tt.expected)
			}
		})
	}
}

func TestGetFace(t *testing.T) {
	fm := NewFontManager()

	// Test sans-serif
	face := fm.GetFace("sans-serif", 16.0, "normal")
	if face == nil {
		t.Error("expected sans-serif face, got nil")
	}

	// Test monospace
	face = fm.GetFace("monospace", 14.0, "normal")
	if face == nil {
		t.Error("expected monospace face, got nil")
	}

	// Test bold
	face = fm.GetFace("sans-serif", 16.0, "bold")
	if face == nil {
		t.Error("expected bold face, got nil")
	}

	// Test cache - second call should return cached face
	face1 := fm.GetFace("sans-serif", 16.0, "normal")
	face2 := fm.GetFace("sans-serif", 16.0, "normal")
	if face1 != face2 {
		t.Error("expected cached face to be returned")
	}
}

func TestMeasureString(t *testing.T) {
	fm := NewFontManager()
	face := fm.GetFace("sans-serif", 16.0, "normal")
	if face == nil {
		t.Fatal("failed to get font face")
	}

	// Measure a simple string
	width := MeasureString(face, "Hello")
	if width == 0 {
		t.Error("expected non-zero width for 'Hello'")
	}

	// Longer string should be wider
	width2 := MeasureString(face, "Hello World")
	if width2 <= width {
		t.Error("expected 'Hello World' to be wider than 'Hello'")
	}

	// Empty string should have zero width
	width3 := MeasureString(face, "")
	if width3 != 0 {
		t.Error("expected zero width for empty string")
	}
}
