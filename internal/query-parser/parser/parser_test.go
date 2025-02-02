package parser

import (
	"fmt"
	"testing"

	"github.com/saferwall/advanced-search/lexer"
	"github.com/saferwall/advanced-search/token"
)

func TestSimpleComparison(t *testing.T) {
	tests := []struct {
		input     string
		wantLeft  string
		wantOp    token.TokenType
		wantRight string
	}{
		{"type=pe", "type", token.ASSIGN, "pe"},
		{"size>1000", "size", token.GT, "1000"},
		{"name!=test.exe", "name", token.NOT_EQ, "test.exe"},
		{"fs<=2023-01-01", "fs", token.LE, "2023-01-01"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			var tokens []*token.Token
			for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
				tokCopy := tok
				tokens = append(tokens, &tokCopy)
			}

			p := New(tokens)
			expr, err := p.ParseExpression()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			compExpr, ok := expr.(*ComparisonExpression)
			if !ok {
				t.Fatalf("expected ComparisonExpression, got %T", expr)
			}

			if compExpr.Left != tt.wantLeft {
				t.Errorf("wrong left value: got %q, want %q", compExpr.Left, tt.wantLeft)
			}
			if compExpr.Operator.Type != tt.wantOp {
				t.Errorf("wrong operator: got %q, want %q", compExpr.Operator.Type, tt.wantOp)
			}
			if compExpr.Right != tt.wantRight {
				t.Errorf("wrong right value: got %q, want %q", compExpr.Right, tt.wantRight)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		input      string
		wantErrMsg string
	}{
		{"type", "expected operator after type"},
		{"type=", "expected value after operator"},
		{"", "unexpected end of input"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			var tokens []*token.Token
			for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
				tokCopy := tok
				tokens = append(tokens, &tokCopy)
			}

			p := New(tokens)
			_, err := p.ParseExpression()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tt.wantErrMsg {
				t.Errorf("wrong error message: got %q, want %q", err.Error(), tt.wantErrMsg)
			}
		})
	}
}

func TestAndOrPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected Expression
	}{
		{
			"type=pe AND tag=upx OR size>1000",
			&BinaryExpression{
				Left: &BinaryExpression{
					Left: &ComparisonExpression{
						Left:     "type",
						Operator: &token.Token{Type: token.ASSIGN, Literal: "="},
						Right:    "pe",
					},
					Operator: &token.Token{Type: token.AND, Literal: "AND"},
					Right: &ComparisonExpression{
						Left:     "tag",
						Operator: &token.Token{Type: token.ASSIGN, Literal: "="},
						Right:    "upx",
					},
				},
				Operator: &token.Token{Type: token.OR, Literal: "OR"},
				Right: &ComparisonExpression{
					Left:     "size",
					Operator: &token.Token{Type: token.GT, Literal: ">"},
					Right:    "1000",
				},
			},
		},
		{
			"type=pe OR tag=upx size>1000",
			&BinaryExpression{
				Left: &ComparisonExpression{
					Left:     "type",
					Operator: &token.Token{Type: token.ASSIGN, Literal: "="},
					Right:    "pe",
				},
				Operator: &token.Token{Type: token.OR, Literal: "OR"},
				Right: &BinaryExpression{
					Left: &ComparisonExpression{
						Left:     "tag",
						Operator: &token.Token{Type: token.ASSIGN, Literal: "="},
						Right:    "upx",
					},
					Operator: &token.Token{Type: token.AND, Literal: "AND"},
					Right: &ComparisonExpression{
						Left:     "size",
						Operator: &token.Token{Type: token.GT, Literal: ">"},
						Right:    "1000",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			var tokens []*token.Token
			for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
				tokCopy := tok
				tokens = append(tokens, &tokCopy)
			}

			p := New(tokens)
			expr, err := p.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			equal, errMsg := compareExpressionsWithErrors(expr, tt.expected)
			if !equal {
				t.Errorf("wrong expression: %s", errMsg)
			}
		})
	}
}

func compareExpressions(a, b Expression) bool {
	switch a := a.(type) {
	case *BinaryExpression:
		b, ok := b.(*BinaryExpression)
		if !ok {
			return false
		}
		return compareExpressions(a.Left, b.Left) &&
			a.Operator.Type == b.Operator.Type &&
			compareExpressions(a.Right, b.Right)
	case *ComparisonExpression:
		b, ok := b.(*ComparisonExpression)
		if !ok {
			return false
		}
		return a.Left == b.Left &&
			a.Operator.Type == b.Operator.Type &&
			a.Right == b.Right
	default:
		return false
	}
}

func compareExpressionsWithErrors(a, b Expression) (bool, string) {
	switch a := a.(type) {
	case *BinaryExpression:
		b, ok := b.(*BinaryExpression)
		if !ok {
			return false, fmt.Sprintf("expected BinaryExpression, got %T", b)
		}
		leftEqual, leftErr := compareExpressionsWithErrors(a.Left, b.Left)
		if !leftEqual {
			return false, fmt.Sprintf("left expressions not equal: %s", leftErr)
		}
		if a.Operator.Type != b.Operator.Type {
			return false, fmt.Sprintf("operators not equal: got %v, want %v", a.Operator.Type, b.Operator.Type)
		}
		rightEqual, rightErr := compareExpressionsWithErrors(a.Right, b.Right)
		if !rightEqual {
			return false, fmt.Sprintf("right expressions not equal: %s", rightErr)
		}
		return true, ""
	case *ComparisonExpression:
		b, ok := b.(*ComparisonExpression)
		if !ok {
			return false, fmt.Sprintf("expected ComparisonExpression, got %T", b)
		}
		if a.Left != b.Left {
			return false, fmt.Sprintf("left values not equal: got %v, want %v", a.Left, b.Left)
		}
		if a.Operator.Type != b.Operator.Type {
			return false, fmt.Sprintf("operators not equal: got %v, want %v", a.Operator.Type, b.Operator.Type)
		}
		if a.Right != b.Right {
			return false, fmt.Sprintf("right values not equal: got %v, want %v", a.Right, b.Right)
		}
		return true, ""
	default:
		return false, fmt.Sprintf("unexpected expression type: %T", a)
	}
}
