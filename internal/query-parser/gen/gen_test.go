package gen

import (
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
		validateFn  func(t *testing.T, q search.Query)
		errContains string
	}{
		{
			name:  "simple string equality",
			input: "type=pe",
			config: Config{
				"type": {},
			},
			validateFn: func(t *testing.T, q search.Query) {
				mq, ok := q.(*search.MatchQuery)
				assert.True(t, ok, "expected MatchQuery")
				assert.NotNil(t, mq)
			},
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
			validateFn: func(t *testing.T, q search.Query) {
				nq, ok := q.(*search.NumericRangeQuery)
				assert.True(t, ok, "expected NumericRangeQuery")
				assert.NotNil(t, nq)
			},
		},
		{
			name:  "date range with field mapping",
			input: "first_seen>=2023-01-01",
			config: Config{
				"first_seen": {
					Type: DATE,
				},
			},
			validateFn: func(t *testing.T, q search.Query) {
				nq, ok := q.(*search.NumericRangeQuery)
				assert.True(t, ok, "expected NumericRangeQuery")
				assert.NotNil(t, nq)
			},
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
			validateFn: func(t *testing.T, q search.Query) {
				dq, ok := q.(*search.DisjunctionQuery)
				assert.True(t, ok, "expected DisjunctionQuery")
				assert.NotNil(t, dq)

				assert.NotNil(t, q)
			},
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
			validateFn: func(t *testing.T, q search.Query) {
				dq, ok := q.(*search.DisjunctionQuery)
				assert.True(t, ok, "expected DisjunctionQuery")
				assert.NotNil(t, dq)
			},
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
			tt.validateFn(t, query)
		})
	}
}
