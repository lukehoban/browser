package main

import (
"testing"
)

func TestIsURL(t *testing.T) {
tests := []struct {
input    string
expected bool
}{
{"http://example.com", true},
{"https://example.com", true},
{"https://news.ycombinator.com/", true},
{"file.html", false},
{"test/file.html", false},
{"/absolute/path/file.html", false},
{"ftp://example.com", false},
}

for _, tt := range tests {
result := isURL(tt.input)
if result != tt.expected {
t.Errorf("isURL(%q) = %v, want %v", tt.input, result, tt.expected)
}
}
}
