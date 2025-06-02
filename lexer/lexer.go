package lexer

import (
	"unicode"

	"github.com/hilli/cooklang/token"
)

type Lexer struct {
	input         string
	position      int
	readPosition  int
	ch            byte          // Only supports ASCII
	tokenBuffer   []token.Token // Buffer for putback tokens
	documentStart bool          // True if we're still at the very beginning of the document
}

func New(input string) *Lexer {
	l := &Lexer{input: input, documentStart: true}
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
	// Check buffer first
	if len(l.tokenBuffer) > 0 {
		tok := l.tokenBuffer[0]
		l.tokenBuffer = l.tokenBuffer[1:]
		return tok
	}

	var tok token.Token

	l.skipWhitespace()

	// If we encounter any non-whitespace content, we're no longer at document start
	if l.documentStart && l.ch != 0 && l.ch != '\n' && !(l.ch == '-' && l.peekChar() == '-' && l.peekCharAt(2) == '-' && l.position == 0) {
		l.documentStart = false
	}

	switch l.ch {
	case '\n':
		tok = newToken(token.NEWLINE, l.ch)
		l.readChar()
		return tok
	case '=': // = or ==
		tok = newToken(token.SECTION, l.ch)
	case '-':
		if l.peekChar() == '-' {
			if l.peekCharAt(2) == '-' {
				// Check if we're at the very beginning of the document
				if l.documentStart && l.position == 0 {
					return l.readYAMLFrontmatter()
				}
			} else {
				// Single line comment starting with --
				// Comments should only be recognized at start of line or after whitespace
				if l.position == 0 || (l.position > 0 && (l.input[l.position-1] == '\n' || l.input[l.position-1] == ' ' || l.input[l.position-1] == '\t')) {
					return l.readComment()
				}
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
		// Only treat as INGREDIENT if immediately followed by a letter, underscore, or digit
		if isLetter(l.peekChar()) || l.peekChar() == '_' || isDigit(l.peekChar()) {
			tok = newToken(token.INGREDIENT, l.ch)
		} else {
			// Treat as regular text if followed by whitespace or other characters
			tok = newToken(token.ILLEGAL, l.ch)
		}
	case '#':
		// Only treat as COOKWARE if immediately followed by a letter, underscore, or digit
		if isLetter(l.peekChar()) || l.peekChar() == '_' || isDigit(l.peekChar()) {
			tok = newToken(token.COOKWARE, l.ch)
		} else {
			// Treat as regular text if followed by whitespace or other characters
			tok = newToken(token.ILLEGAL, l.ch)
		}
	case '~':
		// Only treat as COOKTIME if immediately followed by a letter, underscore, or opening brace
		if isLetter(l.peekChar()) || l.peekChar() == '_' || l.peekChar() == '{' {
			tok = newToken(token.COOKTIME, l.ch)
		} else {
			// Treat as regular text if followed by whitespace, digit, or other characters
			tok = newToken(token.ILLEGAL, l.ch)
		}
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
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
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

func (l *Lexer) readComment() token.Token {
	// Skip the opening --
	l.readChar() // skip first -
	l.readChar() // skip second -

	// Read the comment content until end of line or EOF
	start := l.position
	for l.ch != '\n' && l.ch != '\r' && l.ch != 0 {
		l.readChar()
	}

	// Extract the comment content
	commentContent := l.input[start:l.position]

	// If we stopped at a newline, consume it to prevent extra step creation
	if l.ch == '\r' && l.peekChar() == '\n' {
		l.readChar() // skip \r
		l.readChar() // skip \n
	} else if l.ch == '\n' {
		l.readChar() // skip \n
	}

	return token.Token{
		Type:    token.COMMENT,
		Literal: commentContent,
	}
}

// PeekToken returns the next token without advancing the lexer position
func (l *Lexer) PeekToken() token.Token {
	// Save current state
	savedPosition := l.position
	savedReadPosition := l.readPosition
	savedCh := l.ch

	// Get next token
	tok := l.NextToken()

	// Restore state
	l.position = savedPosition
	l.readPosition = savedReadPosition
	l.ch = savedCh

	return tok
}

// PutBackToken puts a token back into the buffer to be returned by the next NextToken call
func (l *Lexer) PutBackToken(tok token.Token) {
	// Add to the beginning of the buffer
	l.tokenBuffer = append([]token.Token{tok}, l.tokenBuffer...)
}
