package lexer

import (
	"unicode"

	"github.com/hilli/cooklang/token"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte // Only supports ASCII
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.ch {
	case '=': // = or ==
		tok = newToken(token.SECTION, l.ch)
	case '-':
		if l.peekChar() == '-' && l.peekCharAt(2) == '-' {
			// Check if we're at the beginning of the input or after a newline
			if l.position == 0 || (l.position > 0 && l.input[l.position-1] == '\n') {
				return l.readYAMLFrontmatter()
			}
		}
		tok = newToken(token.DASH, l.ch)
	case '%':
		tok = newToken(token.PERCENT, l.ch)
	case '.':
		tok = newToken(token.PERIOD, l.ch)
	case '/':
		tok = newToken(token.DIVIDE, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '@':
		tok = newToken(token.INGREDIENT, l.ch)
	case '#':
		tok = newToken(token.COOKWARE, l.ch)
	case '~':
		tok = newToken(token.COOKTIME, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifyer()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readIdentifyer() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch))
}

func isDigit(ch byte) bool {
	return unicode.IsDigit(rune(ch))
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) peekCharAt(offset int) byte {
	pos := l.readPosition + offset - 1
	if pos >= len(l.input) {
		return 0
	} else {
		return l.input[pos]
	}
}

func (l *Lexer) readYAMLFrontmatter() token.Token {
	// Skip the opening ---
	l.readChar() // skip first -
	l.readChar() // skip second -
	l.readChar() // skip third -

	// Skip any whitespace after opening ---
	for l.ch == ' ' || l.ch == '\t' {
		l.readChar()
	}

	// Must have newline after opening ---
	if l.ch != '\n' && l.ch != '\r' {
		return newToken(token.ILLEGAL, '-')
	}

	if l.ch == '\r' && l.peekChar() == '\n' {
		l.readChar() // skip \r
	}
	if l.ch == '\n' {
		l.readChar() // skip \n
	}

	// Read the YAML content until we find closing ---
	start := l.position
	for {
		if l.ch == 0 {
			// EOF without closing ---
			return newToken(token.ILLEGAL, '-')
		}

		// Check for closing ---
		if l.ch == '-' && l.peekChar() == '-' && l.peekCharAt(2) == '-' {
			// Make sure it's at the start of a line
			if l.position == 0 || l.input[l.position-1] == '\n' {
				break
			}
		}

		l.readChar()
	}

	// Extract the YAML content
	yamlContent := l.input[start:l.position]

	// Skip the closing ---
	l.readChar() // skip first -
	l.readChar() // skip second -
	l.readChar() // skip third -

	return token.Token{
		Type:    token.YAML_FRONTMATTER,
		Literal: yamlContent,
	}
}
