package gen

import (
	"fmt"
	"strconv"

	"github.com/couchbase/gocb/v2/search"
	"github.com/saferwall/saferwall-api/internal/query-parser/lexer"
	"github.com/saferwall/saferwall-api/internal/query-parser/parser"
	"github.com/saferwall/saferwall-api/internal/query-parser/token"
)

type Config map[string]struct {
	Type  string
	Alias string
}

var config Config

func Generate(input string, cfg Config) (search.Query, error) {
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

	config = cfg
	result, err := GenerateCouchbaseFTS(expr)
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
	field := expr.Left
	if value, ok := config[expr.Left]; ok {
		field = value.Alias
	}

	switch expr.Operator.Type {
	case token.ASSIGN:
		return search.NewMatchQuery(expr.Right).Field(field), nil
	case token.NOT_EQ:
		return search.NewBooleanQuery().MustNot(search.NewMatchQuery(expr.Right).Field(field)), nil
	case token.GT, token.GE, token.LT, token.LE:
		return generateRangeQuery(expr)
	default:
		return nil, fmt.Errorf("unsupported comparison operator: %s", expr.Operator.Type)
	}
}

func generateRangeQuery(expr *parser.ComparisonExpression) (search.Query, error) {
	field := expr.Left
	if v, ok := config[expr.Left]; ok {
		field = v.Alias
	}
	value := expr.Right

	t := config[expr.Left].Type

	isInclusive := expr.Operator.Type == token.GE || expr.Operator.Type == token.LE
	switch expr.Operator.Type {
	case token.GT, token.GE:
		switch t {
		case "number":
			v, err := strconv.ParseFloat(value, 32)
			if err != nil {
				return nil, fmt.Errorf("unsupported type for field: %s", field)
			}
			return search.NewNumericRangeQuery().Field(field).Min(float32(v), isInclusive), nil
		case "date":
			return search.NewDateRangeQuery().Field(field).Start(expr.Right, isInclusive), nil
		default:
			return search.NewTermRangeQuery(field).Min(expr.Right, isInclusive), nil
		}

	case token.LT, token.LE:
		switch t {
		case "number":
			v, err := strconv.ParseFloat(value, 32)
			if err != nil {
				return nil, fmt.Errorf("unsupported type for field: %s", field)
			}
			return search.NewNumericRangeQuery().Field(field).Max(float32(v), isInclusive), nil
		case "date":
			return search.NewDateRangeQuery().Field(field).End(expr.Right, isInclusive), nil
		default:
			return search.NewTermRangeQuery(field).Max(expr.Right, isInclusive), nil
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
