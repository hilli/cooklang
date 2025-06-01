package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	PREAMPLE         = "---"
	YAML_FRONTMATTER = "YAML_FRONTMATTER"

	COMMENT = "-- "
	SECTION = "="
	NEWLINE = "NEWLINE"

	IDENT = "IDENT"
	INT   = "INT"

	COOKTIME   = "~"
	COOKWARE   = "#"
	INGREDIENT = "@"

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"
	DIVIDE    = "/"
	PERCENT   = "%"
	DASH      = "-"
	PERIOD    = "."

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"
)

var keywords = map[string]TokenType{
	"@": INGREDIENT,
	"#": COOKWARE,
	"~": COOKTIME,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
