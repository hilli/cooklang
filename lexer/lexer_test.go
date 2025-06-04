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
		{token.INT, "5"},
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
