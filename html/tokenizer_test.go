package html

import "testing"

func TestTokenizerText(t *testing.T) {
	input := "Hello, World!"
	tokenizer := NewTokenizer(input)

	token, ok := tokenizer.Next()
	if !ok {
		t.Fatal("Expected token")
	}
	if token.Type != TextToken {
		t.Errorf("Expected TextToken, got %v", token.Type)
	}
	if token.Data != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got %v", token.Data)
	}
}

func TestTokenizerSimpleTag(t *testing.T) {
	input := "<div>"
	tokenizer := NewTokenizer(input)

	token, ok := tokenizer.Next()
	if !ok {
		t.Fatal("Expected token")
	}
	if token.Type != StartTagToken {
		t.Errorf("Expected StartTagToken, got %v", token.Type)
	}
	if token.Data != "div" {
		t.Errorf("Expected tag name 'div', got %v", token.Data)
	}
}

func TestTokenizerEndTag(t *testing.T) {
	input := "</div>"
	tokenizer := NewTokenizer(input)

	token, ok := tokenizer.Next()
	if !ok {
		t.Fatal("Expected token")
	}
	if token.Type != EndTagToken {
		t.Errorf("Expected EndTagToken, got %v", token.Type)
	}
	if token.Data != "div" {
		t.Errorf("Expected tag name 'div', got %v", token.Data)
	}
}

func TestTokenizerSelfClosingTag(t *testing.T) {
	input := "<br />"
	tokenizer := NewTokenizer(input)

	token, ok := tokenizer.Next()
	if !ok {
		t.Fatal("Expected token")
	}
	if token.Type != SelfClosingTagToken {
		t.Errorf("Expected SelfClosingTagToken, got %v", token.Type)
	}
	if token.Data != "br" {
		t.Errorf("Expected tag name 'br', got %v", token.Data)
	}
}

func TestTokenizerAttributes(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedID    string
		expectedClass string
	}{
		{
			name:          "double quoted attributes",
			input:         `<div id="main" class="container">`,
			expectedID:    "main",
			expectedClass: "container",
		},
		{
			name:          "single quoted attributes",
			input:         `<div id='main' class='container'>`,
			expectedID:    "main",
			expectedClass: "container",
		},
		{
			name:          "unquoted attributes",
			input:         `<div id=main class=container>`,
			expectedID:    "main",
			expectedClass: "container",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenizer := NewTokenizer(tt.input)
			token, ok := tokenizer.Next()

			if !ok {
				t.Fatal("Expected token")
			}
			if token.Type != StartTagToken {
				t.Errorf("Expected StartTagToken, got %v", token.Type)
			}
			if token.Attributes["id"] != tt.expectedID {
				t.Errorf("Expected id='%v', got '%v'", tt.expectedID, token.Attributes["id"])
			}
			if token.Attributes["class"] != tt.expectedClass {
				t.Errorf("Expected class='%v', got '%v'", tt.expectedClass, token.Attributes["class"])
			}
		})
	}
}

func TestTokenizerComment(t *testing.T) {
	input := "<!-- This is a comment -->"
	tokenizer := NewTokenizer(input)

	token, ok := tokenizer.Next()
	if !ok {
		t.Fatal("Expected token")
	}
	if token.Type != CommentToken {
		t.Errorf("Expected CommentToken, got %v", token.Type)
	}
	if token.Data != " This is a comment " {
		t.Errorf("Expected ' This is a comment ', got %v", token.Data)
	}
}

func TestTokenizerDoctype(t *testing.T) {
	input := "<!DOCTYPE html>"
	tokenizer := NewTokenizer(input)

	token, ok := tokenizer.Next()
	if !ok {
		t.Fatal("Expected token")
	}
	if token.Type != DoctypeToken {
		t.Errorf("Expected DoctypeToken, got %v", token.Type)
	}
}

func TestTokenizerMultipleTokens(t *testing.T) {
	input := "<html><body>Hello</body></html>"
	tokenizer := NewTokenizer(input)

	expectedTokens := []struct {
		tokenType TokenType
		data      string
	}{
		{StartTagToken, "html"},
		{StartTagToken, "body"},
		{TextToken, "Hello"},
		{EndTagToken, "body"},
		{EndTagToken, "html"},
	}

	for i, expected := range expectedTokens {
		token, ok := tokenizer.Next()
		if !ok {
			t.Fatalf("Expected token %d", i)
		}
		if token.Type != expected.tokenType {
			t.Errorf("Token %d: expected type %v, got %v", i, expected.tokenType, token.Type)
		}
		if token.Data != expected.data {
			t.Errorf("Token %d: expected data '%v', got '%v'", i, expected.data, token.Data)
		}
	}
}
