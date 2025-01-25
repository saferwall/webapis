package token

import "strings"

type TokenType string
type ValueType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// literals
	INT   = "INT"
	IDENT = "IDENT"
	DATE  = "DATE"
	UNIT  = "UNIT"

	// Operators
	ASSIGN = "="

	LT = "<"
	GT = ">"
	LE = "<="
	GE = ">="

	EQ     = "=="
	NOT_EQ = "!="

	// Keywords
	AND = "AND"
	OR  = "OR"

	LPAREN = "("
	RPAREN = ")"
)

var TypeEnum = []string{"pe", "elf", "macho", "txt"}
var ExtensionEnum = []string{"dll", "exe", "ps1"}

var keywords = map[string]TokenType{
	"or":  OR,
	"and": AND,
}

var sizeUnits = map[string]TokenType{
	"kb": UNIT,
	"mb": UNIT,
	"gb": UNIT,
	"tb": UNIT,
}

// Update LookupIdent to return both token type and value type
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[strings.ToLower(ident)]; ok {
		return tok
	}
	if tok, ok := sizeUnits[strings.ToLower(ident)]; ok {
		return tok
	}
	return IDENT
}
