package parser

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/saferwall/saferwall-api/internal/query-parser/token"
)

// AST node types
type Node interface {
	TokenLiteral() string
}

type Expression interface {
	Node
	expressionNode()
}

// Binary expression (e.g. type=pe AND tag=upx)
type BinaryExpression struct {
	Left     Expression
	Operator *token.Token
	Right    Expression
}

func (be *BinaryExpression) expressionNode()      {}
func (be *BinaryExpression) TokenLiteral() string { return be.Operator.Literal }

// Comparison expression (e.g. type=pe)
type ComparisonExpression struct {
	Left     string
	Operator *token.Token
	Right    string
}

func (ce *ComparisonExpression) expressionNode()      {}
func (ce *ComparisonExpression) TokenLiteral() string { return ce.Operator.Literal }

// Parser structure
type Parser struct {
	tokens  []*token.Token
	current int
}

func New(tokens []*token.Token) *Parser {
	return &Parser{
		tokens:  tokens,
		current: 0,
	}
}

func (p *Parser) eatToken() *token.Token {
	if p.current >= len(p.tokens) {
		return &token.Token{Type: token.EOF, Literal: ""}
	}

	tok := p.tokens[p.current]
	p.current++
	return tok
}

func (p *Parser) peekToken() *token.Token {
	if p.current >= len(p.tokens) {
		return &token.Token{Type: token.EOF, Literal: ""}
	}

	return p.tokens[p.current]
}

func (p *Parser) match(tokenType token.TokenType) bool {
	if p.current >= len(p.tokens) {
		return false
	}

	return p.tokens[p.current].Type == tokenType
}

func (p *Parser) Parse() (Expression, error) {
	return p.ParseExpression()
}

func (p *Parser) ParseExpression() (Expression, error) {
	return p.ParseOr()
}

func (p *Parser) ParseOr() (Expression, error) {
	expr, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.match(token.OR) {
		operator := p.eatToken()

		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}

		expr = &BinaryExpression{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) parseAnd() (Expression, error) {
	var expr Expression

	if p.match(token.LPAREN) {
		p.eatToken()
		exp, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}

		if !p.match(token.RPAREN) {
			return nil, fmt.Errorf("expected closing parenthesis")
		}
		p.eatToken() // eat closing parenthesis

		expr = exp
	} else {
		exp, err := p.parseComparison()
		if err != nil {
			return nil, err
		}

		expr = exp
	}

	for p.match(token.AND) {
		operator := p.eatToken()

		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}

		expr = &BinaryExpression{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}

	// two expressions separated by a space are implicitly ANDed
	for p.match(token.IDENT) {
		operator := token.Token{Type: token.AND, Literal: "AND"}

		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}

		expr = &BinaryExpression{
			Left:     expr,
			Operator: &operator,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) parseComparison() (Expression, error) {
	if p.current >= len(p.tokens) {
		return nil, fmt.Errorf("unexpected end of input")
	}

	left := p.eatToken()
	left.Literal = p.applyUnitIfExist(*left)

	if p.current >= len(p.tokens) {
		return nil, fmt.Errorf("expected operator after %s", left.Literal)
	}

	operator := p.eatToken()

	if p.current >= len(p.tokens) {
		return nil, fmt.Errorf("expected value after operator")
	}

	right := p.eatToken()
	right.Literal = p.applyUnitIfExist(*right)

	return &ComparisonExpression{
		Left:     left.Literal,
		Operator: operator,
		Right:    right.Literal,
	}, nil
}

func (p *Parser) applyUnitIfExist(ident token.Token) string {

	if ident.Type != token.INT || !p.match(token.UNIT) {
		return ident.Literal
	}

	unit := p.eatToken()

	num, err := strconv.Atoi(ident.Literal)
	if err != nil {
		// since the lexer always returns valid numbers
		// Atoi would fail only if the number exceeds int32
		return strconv.Itoa(math.MaxInt32)
	}

	units := map[string]int{
		"kb": 1000,
		"mb": 1000_000,
		"gb": 1000_000_000,
		"tb": 1000_000_000_000,
	}
	return strconv.Itoa(num * units[strings.ToLower(unit.Literal)])
}
