package backend


import (
"strings"
"testing"
)

func TestEscapeLiteral(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: "'hello'",
		},
		{
			name:     "string with single quote",
			input:    "O'Reilly",
			expected: "'O''Reilly'",
		},
		{
			name:     "string with multiple single quotes",
			input:    "can't won't don't",
			expected: "'can''t won''t don''t'",
		},
		{
			name:     "string with double quotes",
			input:    `"quoted"`,
			expected: `'"quoted"'`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "''",
		},
		{
			name:     "string with backslash",
			input:    "path\\to\\file",
			expected: "'path\\to\\file'",
		},
		{
			name:     "string with newline",
			input:    "line1\nline2",
			expected: "'line1\nline2'",
		},
		{
			name:     "string with tab",
			input:    "col1\tcol2",
			expected: "'col1\tcol2'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeLiteral(tt.input)
			if result != tt.expected {
				t.Errorf("EscapeLiteral(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeDuckDBIdentifier(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		// Simple identifiers (unquoted)
		{
			name:     "simple identifier",
			input:    "my_table",
			expected: "my_table",
		},
		{
			name:     "identifier with numbers",
			input:    "table_123",
			expected: "table_123",
		},
		{
			name:     "identifier starting with underscore",
			input:    "_private_table",
			expected: "_private_table",
		},
		{
			name:     "reserved keyword (simple)",
			input:    "select",
			expected: "select",
		},
		{
			name:     "single character",
			input:    "a",
			expected: "a",
		},

		// Complex identifiers (quoted)
		{
			name:     "identifier with double quote",
			input:    `some"col`,
			expected: `"some""col"`,
		},
		{
			name:     "identifier with multiple quotes",
			input:    `"quoted"name`,
			expected: `"""quoted""name"`,
		},
		{
			name:     "identifier with spaces",
			input:    "table with spaces",
			expected: `"table with spaces"`,
		},
		{
			name:     "identifier with special characters",
			input:    "table-with-dashes",
			expected: `"table-with-dashes"`,
		},
		{
			name:     "identifier with dots",
			input:    "schema.table",
			expected: `"schema.table"`,
		},
		{
			name:     "identifier starting with number",
			input:    "123table",
			expected: `"123table"`,
		},
		{
			name:     "identifier with unicode",
			input:    "tëst_täble",
			expected: `"tëst_täble"`,
		},
		{
			name:     "identifier with mixed case and special chars",
			input:    "MyTable-123",
			expected: `"MyTable-123"`,
		},

		// Edge cases
		{
			name:        "empty identifier",
			input:       "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "identifier with NUL character",
			input:       "table\x00name",
			expected:    "",
			expectError: true,
		},
		{
			name:     "identifier with only underscores",
			input:    "___",
			expected: "___",
		},
		{
			name:     "identifier with newlines",
			input:    "table\nname",
			expected: `"table
name"`,
		},
		{
			name:     "identifier with tabs",
			input:    "table\tname",
			expected: `"table	name"`,
		},
		{
			name:     "very long identifier",
			input:    strings.Repeat("a", 100),
			expected: strings.Repeat("a", 100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizeDuckDBIdentifier(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("SanitizeDuckDBIdentifier(%q) expected error but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("SanitizeDuckDBIdentifier(%q) failed: %v", tt.input, err)
				return
			}

			if result != tt.expected {
				t.Errorf("SanitizeDuckDBIdentifier(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
