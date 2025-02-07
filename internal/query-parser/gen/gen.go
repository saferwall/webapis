package gen

import (
	"fmt"
	"strconv"
	"time"

	"github.com/couchbase/gocb/v2/search"
	"github.com/saferwall/saferwall-api/internal/query-parser/lexer"
	"github.com/saferwall/saferwall-api/internal/query-parser/parser"
	"github.com/saferwall/saferwall-api/internal/query-parser/token"
)

type Type int

// Key is the identifier which can map to one or multiple fields
// if none is provided it's assumed that the field is the identifier.
// if Type is not provided it's assumed to be a string.
type Config map[string]struct {
	Type       Type
	Field      string
	FieldGroup []string
}

const (
	NUMBER Type = iota
	DATE
)

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

	for _, v := range cfg {
		if v.Field != "" && len(v.FieldGroup) != 0 {
			panic("config can not have path and path group at the same time")
		}
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
	if len(config[expr.Left].FieldGroup) != 0 {
		queries := []search.Query{}
		for _, field := range config[expr.Left].FieldGroup {
			query, err := buildComparisonQuery(field, expr.Right, expr.Operator.Type, config[expr.Left].Type)
			if err != nil {
				return nil, err
			}

			queries = append(queries, query)
		}
		return search.NewDisjunctionQuery(queries...), nil
	}

	identifier := expr.Left
	field := identifier
	if v, ok := config[identifier]; ok && v.Field != "" {
		field = v.Field
	}

	valueType := config[expr.Left].Type

	return buildComparisonQuery(field, expr.Right, expr.Operator.Type, valueType)
}

func buildComparisonQuery(field, value string, operator token.TokenType, valueType Type) (search.Query, error) {
	switch operator {
	case token.ASSIGN:
		// NOTE: might need to support term match query
		return search.NewMatchQuery(value).Field(field), nil
	case token.NOT_EQ:
		return search.NewBooleanQuery().MustNot(search.NewMatchQuery(value).Field(field)), nil
	case token.GT, token.GE, token.LT, token.LE:
		return buildRangeQuery(field, value, operator, valueType)
	default:
		return nil, fmt.Errorf("unsupported comparison operator: %s", operator)
	}
}

func buildRangeQuery(field, value string, operator token.TokenType, valueType Type) (search.Query, error) {
	isInclusive := operator == token.GE || operator == token.LE
	switch operator {
	case token.GT, token.GE:
		switch valueType {
		case NUMBER:
			v, err := strconv.ParseFloat(value, 32)
			if err != nil {
				return nil, fmt.Errorf("unsupported type for field: %s", field)
			}
			return search.NewNumericRangeQuery().Field(field).Min(float32(v), isInclusive), nil
		case DATE:
			timestamp, err := parseDate(value)
			if err != nil {
				return nil, fmt.Errorf("unsupported type for field: %s", field)
			}
			return search.NewNumericRangeQuery().Field(field).Min(float32(timestamp), isInclusive), nil
		default:
			return search.NewTermRangeQuery(field).Min(value, isInclusive), nil
		}

	case token.LT, token.LE:
		switch valueType {
		case NUMBER:
			num, err := strconv.ParseFloat(value, 32)
			if err != nil {
				return nil, fmt.Errorf("unsupported type for field: %s", field)
			}
			return search.NewNumericRangeQuery().Field(field).Max(float32(num), isInclusive), nil
		case DATE:
			timestamp, err := parseDate(value)
			if err != nil {
				return nil, fmt.Errorf("unsupported type for field: %s", field)
			}
			return search.NewNumericRangeQuery().Field(field).Max(float32(timestamp), isInclusive), nil
		default:
			return search.NewTermRangeQuery(field).Max(value, isInclusive), nil
		}
	}

	return nil, fmt.Errorf("unsupported range operator: %s", operator)
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

func parseDate(date string) (int64, error) {
	// Try parsing various formats
	formats := []string{
		"2006",
		"2006-01",
		"2006-01-02",                // ISO date
		"2006-01-02T15:04:05Z07:00", // RFC3339
		"2006-01-02T15:04:05Z",      // RFC3339 without timezone
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, date); err == nil {
			return t.Unix(), nil
		}
	}

	return 0, fmt.Errorf("unable to parse date: %s", date)
}
