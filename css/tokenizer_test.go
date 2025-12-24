package css

import "testing"

func TestTokenizerIdent(t *testing.T) {
	tokenizer := NewTokenizer("color")
	token := tokenizer.Next()

	if token.Type != IdentToken {
		t.Errorf("Expected IdentToken, got %v", token.Type)
	}
	if token.Value != "color" {
		t.Errorf("Expected 'color', got %v", token.Value)
	}
}

func TestTokenizerString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"double quotes", `"hello"`, "hello"},
		{"single quotes", `'world'`, "world"},
		{"with spaces", `"hello world"`, "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenizer := NewTokenizer(tt.input)
			token := tokenizer.Next()

			if token.Type != StringToken {
				t.Errorf("Expected StringToken, got %v", token.Type)
			}
			if token.Value != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, token.Value)
			}
		})
	}
}

func TestTokenizerNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"integer", "42", "42"},
		{"decimal", "3.14", "3.14"},
		{"with px unit", "10px", "10px"},
		{"with em unit", "1.5em", "1.5em"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenizer := NewTokenizer(tt.input)
			token := tokenizer.Next()

			if token.Type != NumberToken {
				t.Errorf("Expected NumberToken, got %v", token.Type)
			}
			if token.Value != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, token.Value)
			}
		})
	}
}

func TestTokenizerHash(t *testing.T) {
	tokenizer := NewTokenizer("#header")
	token := tokenizer.Next()

	if token.Type != HashToken {
		t.Errorf("Expected HashToken, got %v", token.Type)
	}
	if token.Value != "header" {
		t.Errorf("Expected 'header', got %v", token.Value)
	}
}

func TestTokenizerDot(t *testing.T) {
	tokenizer := NewTokenizer(".container")
	token := tokenizer.Next()

	if token.Type != DotToken {
		t.Errorf("Expected DotToken, got %v", token.Type)
	}

	// Next should be ident
	token = tokenizer.Next()
	if token.Type != IdentToken {
		t.Errorf("Expected IdentToken, got %v", token.Type)
	}
	if token.Value != "container" {
		t.Errorf("Expected 'container', got %v", token.Value)
	}
}

func TestTokenizerPunctuation(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{":", ColonToken},
		{";", SemicolonToken},
		{",", CommaToken},
		{"{", LeftBraceToken},
		{"}", RightBraceToken},
		{"(", LeftParenToken},
		{")", RightParenToken},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokenizer := NewTokenizer(tt.input)
			token := tokenizer.Next()

			if token.Type != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, token.Type)
			}
		})
	}
}

func TestTokenizerComment(t *testing.T) {
	tokenizer := NewTokenizer("/* comment */ color")
	token := tokenizer.Next()

	// Comment should be skipped
	if token.Type != IdentToken {
		t.Errorf("Expected IdentToken after comment, got %v", token.Type)
	}
	if token.Value != "color" {
		t.Errorf("Expected 'color', got %v", token.Value)
	}
}

func TestTokenizerCSSRule(t *testing.T) {
	input := "div { color: red; }"
	tokenizer := NewTokenizer(input)

	expectedTokens := []struct {
		tokenType TokenType
		value     string
	}{
		{IdentToken, "div"},
		{WhitespaceToken, " "},
		{LeftBraceToken, "{"},
		{WhitespaceToken, " "},
		{IdentToken, "color"},
		{ColonToken, ":"},
		{WhitespaceToken, " "},
		{IdentToken, "red"},
		{SemicolonToken, ";"},
		{WhitespaceToken, " "},
		{RightBraceToken, "}"},
	}

	for i, expected := range expectedTokens {
		token := tokenizer.Next()
		if token.Type != expected.tokenType {
			t.Errorf("Token %d: expected type %v, got %v", i, expected.tokenType, token.Type)
		}
		if token.Value != expected.value {
			t.Errorf("Token %d: expected value %v, got %v", i, expected.value, token.Value)
		}
	}
}
