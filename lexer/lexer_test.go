package lexer

import (
	"testing"

	"github.com/hilli/cooklang/token"
)

func TestNextToken(t *testing.T) {
	input := `=@a~{}#c{}(),/5;`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.SECTION, "="},
		{token.INGREDIENT, "@"},
		{token.IDENT, "a"},
		{token.COOKTIME, "~"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.COOKWARE, "#"},
		{token.IDENT, "c"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.COMMA, ","},
		{token.DIVIDE, "/"},
		{token.IDENT, "5"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	l := New(input)

	for _, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("expected type %q, got %q", tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("expected literal %q, got %q", tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestYAMLFrontmatter(t *testing.T) {
	input := `---
title: A recipe
tags:
  - recipe
  - cooking
---
Cook the @shrimp{1} for ~{3%minutes} in a #pot`

	l := New(input)

	// First token should be the YAML frontmatter
	tok := l.NextToken()
	if tok.Type != token.YAML_FRONTMATTER {
		t.Fatalf("expected YAML_FRONTMATTER, got %q", tok.Type)
	}

	expectedYAML := `title: A recipe
tags:
  - recipe
  - cooking
`
	if tok.Literal != expectedYAML {
		t.Fatalf("expected YAML content %q, got %q", expectedYAML, tok.Literal)
	}

	// Next should be recipe content tokens
	tok = l.NextToken()
	// Skip any newlines after the frontmatter
	if tok.Type == token.NEWLINE {
		tok = l.NextToken()
	}
	if tok.Type != token.IDENT || tok.Literal != "Cook" {
		t.Fatalf("expected Cook token, got type %q, literal %q", tok.Type, tok.Literal)
	}
}

func TestWithoutYAMLFrontmatter(t *testing.T) {
	input := `Cook the @shrimp{1} for ~{3%minutes} in a #pot`

	l := New(input)

	// Should start directly with recipe content, no frontmatter
	tok := l.NextToken()
	if tok.Type != token.IDENT || tok.Literal != "Cook" {
		t.Fatalf("expected Cook token, got type %q, literal %q", tok.Type, tok.Literal)
	}

	// Skip whitespace and get "the"
	tok = l.NextToken()
	if tok.Type == token.WHITESPACE {
		tok = l.NextToken()
	}
	if tok.Type != token.IDENT || tok.Literal != "the" {
		t.Fatalf("expected 'the' token, got type %q, literal %q", tok.Type, tok.Literal)
	}

	// Skip whitespace and check for ingredient token "@"
	tok = l.NextToken()
	if tok.Type == token.WHITESPACE {
		tok = l.NextToken()
	}
	if tok.Type != token.INGREDIENT {
		t.Fatalf("expected INGREDIENT token, got %q", tok.Type)
	}
}

func TestDashesNotYAMLFrontmatter(t *testing.T) {
	input := `Cook for 5-7 minutes --- this is not frontmatter`

	l := New(input)

	// Should treat --- as ILLEGAL since it's not at start of line
	tok := l.NextToken() // "Cook"
	if tok.Type != token.IDENT || tok.Literal != "Cook" {
		t.Fatalf("expected Cook token, got type %q, literal %q", tok.Type, tok.Literal)
	}
	// Skip to the dashes
	for tok.Type != token.DASH && tok.Type != token.EOF {
		tok = l.NextToken()
	}

	// Should find DASH token for the first dash (since --- is not at start of line)
	if tok.Type != token.DASH {
		t.Fatalf("expected DASH token for dash not at start of line, got %q", tok.Type)
	}
}

// TestNewlineVariants tests that different newline styles are all tokenized as NEWLINE
func TestNewlineVariants(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []token.TokenType
	}{
		{
			name:     "Unix LF",
			input:    "a\nb",
			expected: []token.TokenType{token.IDENT, token.NEWLINE, token.IDENT, token.EOF},
		},
		{
			name:     "Windows CRLF",
			input:    "a\r\nb",
			expected: []token.TokenType{token.IDENT, token.NEWLINE, token.IDENT, token.EOF},
		},
		{
			name:     "Old Mac CR",
			input:    "a\rb",
			expected: []token.TokenType{token.IDENT, token.NEWLINE, token.IDENT, token.EOF},
		},
		{
			name:     "Double Unix LF",
			input:    "a\n\nb",
			expected: []token.TokenType{token.IDENT, token.NEWLINE, token.NEWLINE, token.IDENT, token.EOF},
		},
		{
			name:     "Double Windows CRLF",
			input:    "a\r\n\r\nb",
			expected: []token.TokenType{token.IDENT, token.NEWLINE, token.NEWLINE, token.IDENT, token.EOF},
		},
		{
			name:     "Double Old Mac CR",
			input:    "a\r\rb",
			expected: []token.TokenType{token.IDENT, token.NEWLINE, token.NEWLINE, token.IDENT, token.EOF},
		},
		{
			name:     "Mixed CRLF and LF",
			input:    "a\r\nb\nc",
			expected: []token.TokenType{token.IDENT, token.NEWLINE, token.IDENT, token.NEWLINE, token.IDENT, token.EOF},
		},
		{
			name:     "Triple blank lines (Unix)",
			input:    "a\n\n\nb",
			expected: []token.TokenType{token.IDENT, token.NEWLINE, token.NEWLINE, token.NEWLINE, token.IDENT, token.EOF},
		},
		{
			name:     "Triple blank lines (Windows)",
			input:    "a\r\n\r\n\r\nb",
			expected: []token.TokenType{token.IDENT, token.NEWLINE, token.NEWLINE, token.NEWLINE, token.IDENT, token.EOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, expected := range tt.expected {
				tok := l.NextToken()
				if tok.Type != expected {
					t.Errorf("token[%d]: expected %s, got %s (literal: %q)", i, expected, tok.Type, tok.Literal)
				}
			}
		})
	}
}

// TestNewlineLiteralNormalization tests that all newline variants produce normalized "\n" literals
func TestNewlineLiteralNormalization(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Unix LF", "a\nb"},
		{"Windows CRLF", "a\r\nb"},
		{"Old Mac CR", "a\rb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			l.NextToken() // skip 'a'
			tok := l.NextToken()
			if tok.Type != token.NEWLINE {
				t.Fatalf("expected NEWLINE, got %s", tok.Type)
			}
			if tok.Literal != "\n" {
				t.Errorf("expected newline literal to be normalized to \"\\n\", got %q", tok.Literal)
			}
		})
	}
}
