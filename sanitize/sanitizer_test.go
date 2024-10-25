package sanitize

import (
	"testing"
)

func TestSanitizer_SanitizeString(t *testing.T) {
	tests := []struct {
		name  string
		opts  SanitizerOptions
		input string
		want  string
	}{
		{
			name: "field value replacement",
			opts: SanitizerOptions{
				ExcludeFields: []string{"password"},
			},
			input: `{"password":"foo"}`,
			want:  `{"password":"` + RedactedStr + `"}`,
		},
		{
			name: "full replacement",
			opts: SanitizerOptions{
				ExcludePatterns: []string{"mypass([0-9]*)"},
			},
			input: `{"password":"mypass12345"}`,
			want:  `{"password":"` + RedactedStr + `"}`,
		},
		{
			name: "full replacement github",
			opts: SanitizerOptions{
				ExcludePatterns: []string{"(?m)(ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9_]{36}"},
			},
			input: `key = ghp_abcdfyocz0uxyzyO9Xn2Estui2kv12aaabgd`,
			want:  `key = ` + RedactedStr,
		},
		{
			name: "form_url",
			opts: SanitizerOptions{
				ExcludeFields: []string{"form_url"},
			},
			input: `{"form_url":"https://example.com/form?token=1234abcd"}`,
			want:  `{"form_url":"` + RedactedStr + `"}`,
		},
		// The database connection string is also redacted by the Basic Auth redaction, it will actually redact more than the
		// plain db redaction
		// {
		// 	name: "database connection string",
		// 	opts: SanitizerOptions{
		// 		ImportCodeMatchers: false,
		// 	},
		// 	input: `{"connection":"mysql://user:1234abcd@localhost:3306/db"}`,
		// 	want:  `{"connection":"mysql://user:` + RedactedStr + `@localhost:3306/db"}`,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSanitizer(tt.opts)
			if got := s.SanitizeString(tt.input); got != tt.want {
				t.Errorf("SanitizeString() = %v, want %v", got, tt.want)
			}
		})
	}
}
