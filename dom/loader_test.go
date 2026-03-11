package dom

import (
	"bytes"
	"image"
	_ "image/png"
	"os"
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

func TestIsURL(t *testing.T) {
tests := []struct {
input string
want  bool
}{
{"http://example.com", true},
{"https://example.com", true},
{"ftp://example.com", false},
{"data:text/plain,hello", false},
{"/home/user/file.txt", false},
{"", false},
{"relative/path", false},
}

for _, tt := range tests {
t.Run(tt.input, func(t *testing.T) {
if got := isURL(tt.input); got != tt.want {
t.Errorf("isURL(%q) = %v, want %v", tt.input, got, tt.want)
}
})
}
}

func TestNewResourceLoader(t *testing.T) {
loader := NewResourceLoader("http://example.com")
if loader.BaseURL != "http://example.com" {
t.Errorf("Expected BaseURL 'http://example.com', got %q", loader.BaseURL)
}
}

func TestLoadResourceFromFile(t *testing.T) {
// Create a temporary file for testing
tmpFile := t.TempDir() + "/test_loader_resource.txt"
content := []byte("hello world")
if err := os.WriteFile(tmpFile, content, 0644); err != nil {
t.Fatalf("Failed to create temp file: %v", err)
}

loader := NewResourceLoader("")
data, err := loader.LoadResource(tmpFile)
if err != nil {
t.Fatalf("LoadResource failed: %v", err)
}
if string(data) != "hello world" {
t.Errorf("Expected 'hello world', got %q", string(data))
}
}

func TestLoadResourceAsString(t *testing.T) {
tmpFile := t.TempDir() + "/test_loader_string.txt"
content := []byte("test content")
if err := os.WriteFile(tmpFile, content, 0644); err != nil {
t.Fatalf("Failed to create temp file: %v", err)
}

loader := NewResourceLoader("")
str, err := loader.LoadResourceAsString(tmpFile)
if err != nil {
t.Fatalf("LoadResourceAsString failed: %v", err)
}
if str != "test content" {
t.Errorf("Expected 'test content', got %q", str)
}
}

func TestLoadResourceAsStringError(t *testing.T) {
loader := NewResourceLoader("")
_, err := loader.LoadResourceAsString("/nonexistent/path/file.txt")
if err == nil {
t.Error("Expected error for nonexistent file")
}
}

func TestLoadResourceFromDataURL(t *testing.T) {
loader := NewResourceLoader("")
data, err := loader.LoadResource("data:text/plain;base64,SGVsbG8=")
if err != nil {
t.Fatalf("LoadResource data URL failed: %v", err)
}
if string(data) != "Hello" {
t.Errorf("Expected 'Hello', got %q", string(data))
}
}

func TestLoadFromDataURLInvalidNoComma(t *testing.T) {
_, err := loadFromDataURL("data:text/plain;base64")
if err == nil {
t.Error("Expected error for data URL without comma")
}
}

func TestLoadResourceFileNotFound(t *testing.T) {
loader := NewResourceLoader("")
_, err := loader.LoadResource("/nonexistent/file.txt")
if err == nil {
t.Error("Expected error for nonexistent file")
}
}
