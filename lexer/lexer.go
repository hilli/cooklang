package lexer

import (
	"unicode"
	"unicode/utf8"

	"github.com/hilli/cooklang/token"
)

type Lexer struct {
	input         string
	position      int
	readPosition  int
	ch            rune          // Now supports Unicode
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
		l.position = l.readPosition // Update position even at EOF
	} else {
		r, size := utf8.DecodeRuneInString(l.input[l.readPosition:])
		if r == utf8.RuneError {
			l.ch = 0
		} else {
			l.ch = r
		}
		l.position = l.readPosition
		l.readPosition += size
	}
}

func (l *Lexer) NextToken() token.Token {
	// Check buffer first
	if len(l.tokenBuffer) > 0 {
		tok := l.tokenBuffer[0]
		l.tokenBuffer = l.tokenBuffer[1:]
		return tok
	}

	var tok token.Token

	// Handle whitespace as tokens instead of skipping
	if l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		return l.readWhitespace()
	}

	// If we encounter any non-whitespace content, we're no longer at document start
	// nolint: staticcheck
	if l.documentStart && !(l.ch == 0 || l.ch == '\n' || (l.ch == '-' && l.peekChar() == '-' && l.peekCharAt(1) == '-' && l.position == 0)) {
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
			if l.peekCharAt(1) == '-' {
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
		// Only treat as INGREDIENT if immediately followed by an identifier character or underscore
		if isIdentifierChar(l.peekChar()) || l.peekChar() == '_' {
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
		if isIdentifierChar(l.ch) {
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

func newToken(tokenType token.TokenType, ch rune) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
		if r == utf8.RuneError {
			return 0
		}
		return r
	}
}

func (l *Lexer) readWhitespace() token.Token {
	position := l.position
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
	return token.Token{Type: token.WHITESPACE, Literal: l.input[position:l.position]}
}

func (l *Lexer) readIdentifyer() string {
	position := l.position
	for isIdentifierChar(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}

// isIdentifierChar checks if a character can be part of an identifier
// This includes letters, digits, emojis, and certain punctuation like hyphens
// but excludes Cooklang special tokens like @, #, ~, =, etc.
func isIdentifierChar(ch rune) bool {
	if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
		return true
	}

	// Allow emojis (which are symbols) but exclude specific Cooklang tokens
	if unicode.IsSymbol(ch) {
		// Exclude these specific symbols that are Cooklang tokens
		switch ch {
		case '@', '#', '~', '=':
			return false
		default:
			return true
		}
	}

	// Allow hyphens for names like "7-inch nonstick frying pan"
	if ch == '-' {
		return true
	}

	return false
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) peekCharAt(offset int) rune {
	// This is more complex with UTF-8, so we'll decode from current position
	pos := l.readPosition
	for i := 0; i < offset && pos < len(l.input); i++ {
		_, size := utf8.DecodeRuneInString(l.input[pos:])
		if size == 0 {
			return 0
		}
		pos += size
	}

	if pos >= len(l.input) {
		return 0
	}

	r, _ := utf8.DecodeRuneInString(l.input[pos:])
	if r == utf8.RuneError {
		return 0
	}
	return r
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
		if l.ch == '-' && l.peekChar() == '-' && l.peekCharAt(1) == '-' {
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

	// Skip any whitespace immediately after --
	for l.ch == ' ' || l.ch == '\t' {
		l.readChar()
	}

	// Read the comment content until end of line or EOF
	start := l.position
	for l.ch != '\n' && l.ch != '\r' && l.ch != 0 {
		l.readChar()
	}

	// Extract the comment content
	commentContent := l.input[start:l.position]

	// Don't consume the newline - let normal token processing handle it
	// This allows newlines after comments to be converted to spaces properly

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
