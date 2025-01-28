package lexer

import (
	"regexp"

	"github.com/saferwall/saferwall-api/internal/query-parser/token"
)

// Define a regular expression for ISO date format
var isoDateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}(T\d{2}:\d{2}(:\d{2})?(Z|(\+|-)\d{2}:\d{2})?)?$`)

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.eatChar()
	return l
}

func (l *Lexer) eatChar() {
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
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.eatChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.eatChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	case '<':
		if l.peekChar() == '=' {
			l.eatChar()
			tok = token.Token{Type: token.LE, Literal: "<="}
		} else {
			tok = newToken(token.LT, l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			l.eatChar()
			tok = token.Token{Type: token.GE, Literal: ">="}
		} else {
			tok = newToken(token.GT, l.ch)
		}
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '"':
		tok.Literal = l.readString()
		if l.ch == '"' {
			tok.Type = token.IDENT
		} else {
			tok.Type = token.ILLEGAL
		}
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			literal := l.readNumber()
			if l.ch == '-' || l.ch == 'T' {
				literal += l.readDatePart()
				if IsISODate(literal) {
					tok.Type = token.DATE
					tok.Literal = literal
					return tok
				}
			}
			tok.Type = token.INT
			tok.Literal = literal
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.eatChar()
	return tok
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.eatChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.eatChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readDatePart() string {
	position := l.position
	for isDigit(l.ch) || l.ch == '-' || l.ch == 'T' || l.ch == ':' || l.ch == 'Z' || l.ch == '+' {
		l.eatChar()
	}
	return l.input[position:l.position]
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.eatChar()
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.eatChar()
	}
	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '.'
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

// Add a function to check if a string matches the ISO date format
func IsISODate(s string) bool {
	return isoDateRegex.MatchString(s)
}
