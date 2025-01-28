package filter

import (
	"sort"
	"testing"
)

func TestSqlFilterSatisfied(t *testing.T) {
	testCases := []struct {
		filter             string
		values             map[string]string
		expected           bool
		expectedFieldNames []string
		err                string
	}{
		// Comparisons
		{
			filter:             "connection = 'foo'",
			values:             map[string]string{"connection": "foo"},
			expected:           true,
			expectedFieldNames: []string{"connection"},
		},
		{
			filter:             "connection = 'foo'",
			values:             map[string]string{"connection": "bar"},
			expected:           false,
			expectedFieldNames: []string{"connection"},
		},
		{
			filter:             "connection != 'foo'",
			values:             map[string]string{"connection": "foo"},
			expected:           false,
			expectedFieldNames: []string{"connection"},
		},
		{
			filter:             "connection != 'foo'",
			values:             map[string]string{"connection": "bar"},
			expected:           true,
			expectedFieldNames: []string{"connection"},
		},
		// IN operator
		{
			filter:             "connection in ('foo','bar')",
			values:             map[string]string{"connection": "bar"},
			expected:           true,
			expectedFieldNames: []string{"connection"},
		},
		{
			filter:             "connection in ('foo','bar')",
			values:             map[string]string{"connection": "other"},
			expected:           false,
			expectedFieldNames: []string{"connection"},
		},
		{
			filter:             "connection not in ('foo','bar')",
			values:             map[string]string{"connection": "other"},
			expected:           true,
			expectedFieldNames: []string{"connection"},
		},
		// LIKE operator
		{
			filter:             "connection like 'fo_'",
			values:             map[string]string{"connection": "foo"},
			expected:           true,
			expectedFieldNames: []string{"connection"},
		},
		{
			filter:             "connection like 'fo_'",
			values:             map[string]string{"connection": "bar"},
			expected:           false,
			expectedFieldNames: []string{"connection"},
		},
		{
			filter:             "connection like 'f%'",
			values:             map[string]string{"connection": "foo"},
			expected:           true,
			expectedFieldNames: []string{"connection"},
		},
		{
			filter:             "connection like '%ob%'",
			values:             map[string]string{"connection": "foobar"},
			expected:           true,
			expectedFieldNames: []string{"connection"},
		},
		// NOT LIKE operator
		{
			filter:             "connection not like 'fo_'",
			values:             map[string]string{"connection": "foo"},
			expected:           false,
			expectedFieldNames: []string{"connection"},
		},
		{
			filter:             "connection not like 'fo_'",
			values:             map[string]string{"connection": "bar"},
			expected:           true,
			expectedFieldNames: []string{"connection"},
		},
		// Complex queries (AND, OR)
		{
			filter:             "connection in ('foo','bar') and connection = 'foo'",
			values:             map[string]string{"connection": "foo"},
			expected:           true,
			expectedFieldNames: []string{"connection"},
		},
		{
			filter:             "connection in ('foo','bar') or connection = 'baz'",
			values:             map[string]string{"connection": "baz"},
			expected:           true,
			expectedFieldNames: []string{"connection"},
		},
		{
			filter:             "connection in ('foo','bar') or connection2 = 'baz'",
			values:             map[string]string{"connection": "no", "connection2": "baz"},
			expected:           true,
			expectedFieldNames: []string{"connection", "connection2"},
		},
	}

	for _, testCase := range testCases {
		// Capture testCase to avoid closure issues in t.Run
		tc := testCase

		t.Run(tc.filter, func(t *testing.T) {
			// Parse the filter
			f, err := NewSqlFilter(tc.filter)
			if tc.err != "" {
				if err == nil || err.Error() != tc.err {
					t.Errorf("parseWhere(%v) err: %v, want %s", tc.filter, err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			// Check if the filter is satisfied
			satisfiesFilter := f.Satisfied(tc.values)

			// Get fieldNames
			fieldNames, err := f.GetFieldNames()
			if err != nil {
				t.Fatal(err)
			}

			// Verify fieldNames
			if len(fieldNames) != len(tc.expectedFieldNames) {
				t.Errorf("sqlFilterSatisfied(%v, %v) fieldNames length want %v, got %v", tc.filter, tc.values, len(tc.expectedFieldNames), len(fieldNames))
			} else {
				sort.Strings(fieldNames)
				sort.Strings(tc.expectedFieldNames)
				for i, v := range fieldNames {
					if v != tc.expectedFieldNames[i] {
						t.Errorf("sqlFilterSatisfied(%v, %v) fieldNames[%d] want %v, got %v", tc.filter, tc.values, i, tc.expectedFieldNames[i], v)
					}
				}
			}

			// Verify satisfiesFilter
			if satisfiesFilter != tc.expected {
				t.Errorf("sqlFilterSatisfied(%v, %v) want %v, got %v", tc.filter, tc.values, tc.expected, satisfiesFilter)
			}
		})
	}
}
