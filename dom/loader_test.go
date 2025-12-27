package dom

import (
	"bytes"
	"image"
	_ "image/png"
	"testing"
)

func TestLoadFromDataURL(t *testing.T) {
	tests := []struct {
		name     string
		dataURL  string
		expected []byte
		wantErr  bool
	}{
		{
			name:     "base64 encoded text",
			dataURL:  "data:text/plain;base64,SGVsbG8sIFdvcmxkIQ==",
			expected: []byte("Hello, World!"),
			wantErr:  false,
		},
		{
			name:     "URL encoded SVG",
			dataURL:  "data:image/svg+xml,%3Csvg%20xmlns%3D%22http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg%22%3E%3C%2Fsvg%3E",
			expected: []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`),
			wantErr:  false,
		},
		{
			name:     "base64 encoded SVG",
			dataURL:  "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjwvc3ZnPg==",
			expected: []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`),
			wantErr:  false,
		},
		{
			name:    "invalid data URL - no comma",
			dataURL: "data:text/plain;base64",
			wantErr: true,
		},
		{
			name:    "invalid base64",
			dataURL: "data:text/plain;base64,!!!invalid!!!",
			wantErr: true,
		},
	}

	loader := NewResourceLoader("")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loader.LoadResource(tt.dataURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !bytes.Equal(got, tt.expected) {
				t.Errorf("LoadResource() = %v, want %v", string(got), string(tt.expected))
			}
		})
	}
}

func TestLoadFromDataURL_PNG(t *testing.T) {
	// A minimal 1x1 red PNG image (base64 encoded)
	// This is a valid PNG with IHDR, IDAT, and IEND chunks
	dataURL := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8DwHwAFBQIAX8jx0gAAAABJRU5ErkJggg=="

	loader := NewResourceLoader("")
	data, err := loader.LoadResource(dataURL)
	if err != nil {
		t.Fatalf("LoadResource() failed: %v", err)
	}

	// Verify it's a valid PNG by decoding it
	_, _, err = image.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("Failed to decode PNG from data URL: %v", err)
	}
}

func TestIsDataURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"data:text/plain;base64,SGVsbG8=", true},
		{"data:image/svg+xml,%3Csvg%3E", true},
		{"http://example.com/image.png", false},
		{"https://example.com/image.png", false},
		{"/path/to/file.png", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isDataURL(tt.input); got != tt.want {
				t.Errorf("isDataURL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
