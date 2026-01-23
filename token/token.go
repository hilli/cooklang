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

	COMMENT        = "-- "
	BLOCK_COMMENT  = "[- -]"
	NOTE           = ">"
	SECTION        = "="
	SECTION_HEADER = "SECTION_HEADER"
	NEWLINE        = "NEWLINE"
	WHITESPACE     = "WHITESPACE"

	IDENT = "IDENT"
	INT   = "INT"

	COOKTIME            = "~"
	COOKWARE            = "#"
	INGREDIENT          = "@"
	OPTIONAL_INGREDIENT = "@?"

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
