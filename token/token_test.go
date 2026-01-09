package token

import "testing"

func TestLookupIdent(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  TokenType
	}{
		{"ingredient symbol", "@", INGREDIENT},
		{"cookware symbol", "#", COOKWARE},
		{"cooktime symbol", "~", COOKTIME},
		{"regular identifier", "flour", IDENT},
		{"empty string", "", IDENT},
		{"number string", "123", IDENT},
		{"mixed identifier", "flour123", IDENT},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LookupIdent(tt.input)
			if got != tt.want {
				t.Errorf("LookupIdent(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestTokenType(t *testing.T) {
	// Verify token type constants are unique and non-empty
	tokenTypes := []TokenType{
		ILLEGAL, EOF, PREAMPLE, YAML_FRONTMATTER,
		COMMENT, BLOCK_COMMENT, NOTE, SECTION, SECTION_HEADER,
		NEWLINE, WHITESPACE, IDENT, INT,
		COOKTIME, COOKWARE, INGREDIENT,
		COMMA, SEMICOLON, DIVIDE, PERCENT, DASH, PERIOD,
		LPAREN, RPAREN, LBRACE, RBRACE,
	}

	seen := make(map[TokenType]bool)
	for _, tt := range tokenTypes {
		if tt == "" {
			t.Errorf("Found empty token type")
		}
		if seen[tt] {
			t.Errorf("Duplicate token type: %v", tt)
		}
		seen[tt] = true
	}
}

func TestToken(t *testing.T) {
	tok := Token{
		Type:    INGREDIENT,
		Literal: "flour",
	}

	if tok.Type != INGREDIENT {
		t.Errorf("Token.Type = %v, want %v", tok.Type, INGREDIENT)
	}
	if tok.Literal != "flour" {
		t.Errorf("Token.Literal = %v, want %v", tok.Literal, "flour")
	}
}
