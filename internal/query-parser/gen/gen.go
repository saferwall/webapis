package gen

import (
	"fmt"
	"strconv"

	"github.com/couchbase/gocb/v2/search"
	"github.com/saferwall/advanced-search/gen"
	"github.com/saferwall/advanced-search/parser"
	"github.com/saferwall/advanced-search/token"
	"github.com/saferwall/saferwall-api/internal/query-parser/lexer"
)

func Generate(input string) (search.Query, error) {
	l := lexer.New(input)
	var tokens []*token.Token
	for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
		tokCopy := tok
		tokens = append(tokens, &tokCopy)
	}

	p := parser.New(tokens)
	expr, err := p.Parse()
	if err != nil {
		return nil, err
	}
	result, err := gen.GenerateCouchbaseFTS(expr)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GenerateCouchbaseFTS(expr parser.Expression) (search.Query, error) {
	switch e := expr.(type) {
	case *parser.BinaryExpression:
		return generateBinaryCouchbase(e)
	case *parser.ComparisonExpression:
		return generateComparisonCouchbase(e)
	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

func generateBinaryCouchbase(expr *parser.BinaryExpression) (search.Query, error) {
	left, err := GenerateCouchbaseFTS(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := GenerateCouchbaseFTS(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case token.AND:
		return search.NewConjunctionQuery(left, right), nil

	case token.OR:
		return search.NewDisjunctionQuery(left, right), nil
	default:
		return nil, fmt.Errorf("unsupported operator type: %T", expr.Operator.Type)
	}
}

func generateComparisonCouchbase(expr *parser.ComparisonExpression) (search.Query, error) {
	// NOTE: might need to support term match query

	switch expr.Operator.Type {
	case token.ASSIGN:
		return search.NewMatchQuery(expr.Right).Field(expr.Left), nil
	case token.NOT_EQ:
		return search.NewBooleanQuery().MustNot(search.NewMatchQuery(expr.Right).Field(expr.Left)), nil
	case token.GT, token.GE, token.LT, token.LE:
		return generateRangeQuery(expr)
	default:
		return nil, fmt.Errorf("unsupported comparison operator: %s", expr.Operator.Type)
	}
}

func generateRangeQuery(expr *parser.ComparisonExpression) (search.Query, error) {
	switch expr.Operator.Type {
	case token.GT, token.GE:
		isInclusive := expr.Operator.Type == token.GE
		if v, ok := isValidF32(expr.Right); ok {
			return search.NewNumericRangeQuery().Field(expr.Left).Min(v, isInclusive), nil
		} else if lexer.IsISODate(expr.Right) {
			return search.NewDateRangeQuery().Field(expr.Left).Start(expr.Right, isInclusive), nil
		} else {
			return search.NewTermRangeQuery(expr.Left).Min(expr.Right, isInclusive), nil
		}

	case token.LT, token.LE:
		isInclusive := expr.Operator.Type == token.LE
		if v, ok := isValidF32(expr.Right); ok {
			return search.NewNumericRangeQuery().Field(expr.Left).Max(v, isInclusive), nil
		} else if lexer.IsISODate(expr.Right) {
			return search.NewDateRangeQuery().Field(expr.Left).End(expr.Right, isInclusive), nil
		} else {
			return search.NewTermRangeQuery(expr.Left).Max(expr.Right, isInclusive), nil
		}
	}

	return nil, fmt.Errorf("unsupported range operator: %s", expr.Operator.Type)
}

func isValidF32(s string) (float32, bool) {
	// Attempt to parse the string as a float64
	value, err := strconv.ParseFloat(s, 32)
	// Check for parsing errors and ensure it fits in float32 range
	if err == nil {
		// Convert the value to float32 and back to float64 for precision check
		f32Value := float32(value)
		if float64(f32Value) == value {
			return f32Value, true
		}
	}
	return 0, false
}
