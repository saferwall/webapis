package lexer

import (
	"testing"

	"github.com/saferwall/saferwall-api/internal/query-parser/token"
)

var tests = []struct {
	input    string
	expected []token.Token
}{
	{
		input: `size >= 1000kb type = pe fs < 2020-12-31 name = "example" positives != 0`,
		expected: []token.Token{
			{Type: token.IDENT, Literal: "size"},
			{Type: token.GE, Literal: ">="},
			{Type: token.INT, Literal: "1000"},
			{Type: token.UNIT, Literal: "kb"},
			{Type: token.IDENT, Literal: "type"},
			{Type: token.ASSIGN, Literal: "="},
			{Type: token.IDENT, Literal: "pe"},
			{Type: token.IDENT, Literal: "fs"},
			{Type: token.LT, Literal: "<"},
			{Type: token.DATE, Literal: "2020-12-31"},
			{Type: token.IDENT, Literal: "name"},
			{Type: token.ASSIGN, Literal: "="},
			{Type: token.IDENT, Literal: "example"},
			{Type: token.IDENT, Literal: "positives"},
			{Type: token.NOT_EQ, Literal: "!="},
			{Type: token.INT, Literal: "0"},
			{Type: token.EOF, Literal: ""},
		},
	},
	{
		input: `type = elf and fs >= 2021-01-01T00:00:00Z or positives < 5`,
		expected: []token.Token{
			{Type: token.IDENT, Literal: "type"},
			{Type: token.ASSIGN, Literal: "="},
			{Type: token.IDENT, Literal: "elf"},
			{Type: token.AND, Literal: "and"},
			{Type: token.IDENT, Literal: "fs"},
			{Type: token.GE, Literal: ">="},
			{Type: token.DATE, Literal: "2021-01-01T00:00:00Z"},
			{Type: token.OR, Literal: "or"},
			{Type: token.IDENT, Literal: "positives"},
			{Type: token.LT, Literal: "<"},
			{Type: token.INT, Literal: "5"},
			{Type: token.EOF, Literal: ""},
		},
	},
	{
		input: `extension = dll or (type = macho and positives > 10)`,
		expected: []token.Token{
			{Type: token.IDENT, Literal: "extension"},
			{Type: token.ASSIGN, Literal: "="},
			{Type: token.IDENT, Literal: "dll"},
			{Type: token.OR, Literal: "or"},
			{Type: token.LPAREN, Literal: "("},
			{Type: token.IDENT, Literal: "type"},
			{Type: token.ASSIGN, Literal: "="},
			{Type: token.IDENT, Literal: "macho"},
			{Type: token.AND, Literal: "and"},
			{Type: token.IDENT, Literal: "positives"},
			{Type: token.GT, Literal: ">"},
			{Type: token.INT, Literal: "10"},
			{Type: token.RPAREN, Literal: ")"},
			{Type: token.EOF, Literal: ""},
		},
	},
}

func TestNextToken(t *testing.T) {
	for i, tt := range tests {

		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			t.Parallel()

			for j, expectedToken := range tt.expected {
				tok := l.NextToken()

				if tok.Type != expectedToken.Type {
					t.Fatalf("tests[%d][%d] - token type wrong. expected=%q, got=%q",
						i, j, expectedToken.Type, tok.Type)
				}

				if tok.Literal != expectedToken.Literal {
					t.Fatalf("tests[%d][%d] - literal wrong. expected=%q, got=%q",
						i, j, expectedToken.Literal, tok.Literal)
				}
			}
		})
	}
}
