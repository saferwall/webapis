package gen

import (
	"encoding/json"
	"testing"

	"github.com/couchbase/gocb/v2/search"
	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		config      Config
		wantErr     bool
		wanted      search.Query
		errContains string
	}{
		{
			name:  "simple string equality",
			input: "type=pe",
			config: Config{
				"type": {},
			},
			wanted: search.NewMatchQuery("pe").Field("type"),
		},
		{
			name:  "numeric comparison with field mapping",
			input: "size>1000",
			config: Config{
				"size": {
					Type:  NUMBER,
					Field: "file_size",
				},
			},
			wanted: search.NewNumericRangeQuery().Field("file_size").Min(float32(1000), false),
		},
		{
			name:  "date range with field mapping",
			input: "first_seen>=2023-01-01",
			config: Config{
				"first_seen": {
					Type: DATE,
				},
			},
			wanted: search.NewNumericRangeQuery().Field("first_seen").Min(float32(1672531200), true), // 2023-01-01 00:00:00 UTC
		},
		{
			name:  "complex boolean expression",
			input: "type=pe AND size>1000 OR first_seen>=2023-01-01",
			config: Config{
				"type": {},
				"size": {
					Type: NUMBER,
				},
				"first_seen": {
					Type: DATE,
				},
			},
			wanted: search.NewDisjunctionQuery(
				search.NewConjunctionQuery(
					search.NewMatchQuery("pe").Field("type"),
					search.NewNumericRangeQuery().Field("size").Min(float32(1000), false),
				),
				search.NewNumericRangeQuery().Field("first_seen").Min(float32(1672531200), true),
			),
		},
		{
			name:  "field group search",
			input: "engines=malware",
			config: Config{
				"engines": {
					FieldGroup: []string{
						"multiav.last_scan.avast.output",
						"multiav.last_scan.mcafee.output",
					},
				},
			},
			wanted: search.NewDisjunctionQuery(
				search.NewMatchQuery("malware").Field("multiav.last_scan.avast.output"),
				search.NewMatchQuery("malware").Field("multiav.last_scan.mcafee.output"),
			),
		},
		{
			name:  "invalid date format",
			input: "first_seen>=invalid-date",
			config: Config{
				"first_seen": {
					Type: DATE,
				},
			},
			wantErr:     true,
			errContains: "unsupported type for field",
		},
		{
			name:  "invalid numeric value",
			input: "size>not-a-number",
			config: Config{
				"size": {
					Type: NUMBER,
				},
			},
			wantErr:     true,
			errContains: "unsupported type for field",
		},
		{
			name:  "multiple field groups",
			input: "all_engines=trojan",
			config: Config{
				"all_engines": {
					FieldGroup: []string{
						"multiav.last_scan.avast.output",
						"multiav.last_scan.mcafee.output",
						"multiav.last_scan.kaspersky.output",
						"multiav.last_scan.symantec.output",
					},
				},
			},
			wanted: search.NewDisjunctionQuery(
				search.NewMatchQuery("trojan").Field("multiav.last_scan.avast.output"),
				search.NewMatchQuery("trojan").Field("multiav.last_scan.mcafee.output"),
				search.NewMatchQuery("trojan").Field("multiav.last_scan.kaspersky.output"),
				search.NewMatchQuery("trojan").Field("multiav.last_scan.symantec.output"),
			),
		},
		{
			name:  "not equals operator",
			input: "type!=pe",
			config: Config{
				"type": {},
			},
			wanted: search.NewBooleanQuery().MustNot(search.NewMatchQuery("pe").Field("type")),
		},
		{
			name:  "complex nested expression with parentheses",
			input: "(type=pe AND size>1000) OR (first_seen>=2023-01-01 AND engines=malware)",
			config: Config{
				"type": {},
				"size": {
					Type: NUMBER,
				},
				"first_seen": {
					Type: DATE,
				},
				"engines": {
					FieldGroup: []string{
						"multiav.last_scan.avast.output",
						"multiav.last_scan.mcafee.output",
					},
				},
			},
			wanted: search.NewDisjunctionQuery(
				search.NewConjunctionQuery(
					search.NewMatchQuery("pe").Field("type"),
					search.NewNumericRangeQuery().Field("size").Min(float32(1000), false),
				),
				search.NewConjunctionQuery(
					search.NewNumericRangeQuery().Field("first_seen").Min(float32(1672531200), true),
					search.NewDisjunctionQuery(
						search.NewMatchQuery("malware").Field("multiav.last_scan.avast.output"),
						search.NewMatchQuery("malware").Field("multiav.last_scan.mcafee.output"),
					),
				),
			),
		},
		{
			name:    "empty input",
			input:   "",
			config:  Config{"type": {}},
			wantErr: true,
		},
		{
			name:    "whitespace only input",
			input:   "   ",
			config:  Config{"type": {}},
			wantErr: true,
		},
		{
			name:  "multiple numeric comparisons",
			input: "size>=1000 AND size<=5000",
			config: Config{
				"size": {
					Type: NUMBER,
				},
			},
			wanted: search.NewConjunctionQuery(
				search.NewNumericRangeQuery().Field("size").Min(float32(1000), true),
				search.NewNumericRangeQuery().Field("size").Max(float32(5000), true),
			),
		},
		{
			name:  "multiple date comparisons",
			input: "first_seen>=2023-01-01 AND first_seen<=2023-12-31",
			config: Config{
				"first_seen": {
					Type: DATE,
				},
			},
			wanted: search.NewConjunctionQuery(
				search.NewNumericRangeQuery().Field("first_seen").Min(float32(1672531200), true), // 2023-01-01
				search.NewNumericRangeQuery().Field("first_seen").Max(float32(1703980800), true), // 2023-12-31 23:59:59
			),
		},
		{
			name:  "invalid operator",
			input: "type>pe",
			config: Config{
				"type": {},
			},
			wantErr: true,
		},
		{
			name:  "missing field in config",
			input: "unknown_field=value",
			config: Config{
				"type": {},
			},
			wanted: search.NewMatchQuery("value").Field("unknown_field"),
		},
		{
			name:  "multiple field groups with complex expression",
			input: "(engines=trojan OR engines=virus) AND size>1000",
			config: Config{
				"engines": {
					FieldGroup: []string{
						"multiav.last_scan.avast.output",
						"multiav.last_scan.mcafee.output",
					},
				},
				"size": {
					Type: NUMBER,
				},
			},
			wanted: search.NewConjunctionQuery(
				search.NewDisjunctionQuery(
					search.NewDisjunctionQuery(
						search.NewMatchQuery("trojan").Field("multiav.last_scan.avast.output"),
						search.NewMatchQuery("trojan").Field("multiav.last_scan.mcafee.output"),
					),
					search.NewDisjunctionQuery(
						search.NewMatchQuery("virus").Field("multiav.last_scan.avast.output"),
						search.NewMatchQuery("virus").Field("multiav.last_scan.mcafee.output"),
					),
				),
				search.NewNumericRangeQuery().Field("size").Min(float32(1000), false),
			),
		},
		{
			name:  "field alias",
			input: "fs>=2023-01-01",
			config: Config{
				"fs": {
					Type:  DATE,
					Field: "first_seen",
				},
			},
			wanted: search.NewNumericRangeQuery().Field("first_seen").Min(float32(1672531200), true),
		},
		{
			name:  "test value with year only",
			input: "fs >= 2024",
			config: Config{
				"fs": {
					Type:  DATE,
					Field: "first_seen",
				},
			},
			wanted: search.NewNumericRangeQuery().Field("first_seen").Min(float32(1704067200), true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := Generate(tt.input, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			current, err := json.Marshal(query)
			assert.NoError(t, err)
			wanted, err := json.Marshal(tt.wanted)
			assert.NoError(t, err)
			assert.Equal(t, string(wanted), string(current), "query mismatch")
		})
	}
}
