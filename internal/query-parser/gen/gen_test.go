package gen

/*
func TestGenerate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		expected search.Query
	}{
		{
			name:    "simple equality",
			input:   "size=100",
			wantErr: false,
			expected: search.NewMatchQuery("100").
				Field("size"),
		},
		{
			name:    "simple inequality",
			input:   "size!=100",
			wantErr: false,
			expected: search.NewBooleanQuery().
				MustNot(search.NewMatchQuery("100").Field("size")),
		},
		{
			name:    "numeric greater than",
			input:   "size>100",
			wantErr: false,
			expected: search.NewNumericRangeQuery().
				Field("size").
				Min(float32(100), false),
		},
		{
			name:    "numeric greater than or equal",
			input:   "size>=100",
			wantErr: false,
			expected: search.NewNumericRangeQuery().
				Field("size").
				Min(float32(100), true),
		},
		{
			name:    "date comparison",
			input:   "created>2024-01-01",
			wantErr: false,
			expected: search.NewDateRangeQuery().
				Field("created").
				Start("2024-01-01", false),
		},
		{
			name:    "string range",
			input:   "name>alice",
			wantErr: false,
			expected: search.NewTermRangeQuery("name").
				Min("alice", false),
		},
		{
			name:    "AND operation",
			input:   "size=100 AND name=test",
			wantErr: false,
			expected: search.NewConjunctionQuery(
				search.NewMatchQuery("100").Field("size"),
				search.NewMatchQuery("test").Field("name"),
			),
		},
		{
			name:    "OR operation",
			input:   "size=100 OR name=test",
			wantErr: false,
			expected: search.NewDisjunctionQuery(
				search.NewMatchQuery("100").Field("size"),
				search.NewMatchQuery("test").Field("name"),
			),
		},
		{
			name:    "invalid syntax",
			input:   "size=",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := Generate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !reflect.DeepEqual(query, tt.expected) {
				t.Errorf("\nexpected: %#v\ngot: %#v", tt.expected, query)
			}
		})
	}
}
*/
