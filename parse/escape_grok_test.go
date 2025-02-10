package parse

import (
	"testing"
)

func Test_escapeGrokArgs(t *testing.T) {
	type args struct {
		fileData                 []byte
		disableTemplatesForProps []string
	}
	tests := []struct {
		name      string
		args      args
		wantBytes []byte
		wantError bool
	}{
		{
			name: "single grok func call",
			args: args{
				disableTemplatesForProps: []string{"layout"},
				fileData: []byte(`format "custom" "c1" {
  layout = grok(%{TIMESTAMP_ISO8601:time_local} - %{NUMBER:event_id} - %{WORD:user} - \[%{DATA:location}\] "%{DATA:message}" %{WORD:severity} "%{DATA:additional_info}")
}
`),
			},
			wantBytes: []byte(`format "custom" "c1" {
  layout ="%%{TIMESTAMP_ISO8601:time_local} - %%{NUMBER:event_id} - %%{WORD:user} - \\[%%{DATA:location}\\] \"%%{DATA:message}\" %%{WORD:severity} \"%%{DATA:additional_info}\""
}
`),
		},
		{
			name: "single grok func call with white space at end",
			args: args{
				disableTemplatesForProps: []string{"layout"},
				fileData: []byte(`format "custom" "c1" {
  layout = grok(%{TIMESTAMP_ISO8601:time_local})                   
}
`),
			},
			wantBytes: []byte(`format "custom" "c1" {
  layout ="%%{TIMESTAMP_ISO8601:time_local}"
}
`),
		},

		{
			name: "single grok func call with white space before bracket end",
			args: args{
				disableTemplatesForProps: []string{"layout"},
				fileData: []byte(`format "custom" "c1" {
  layout = grok(%{TIMESTAMP_ISO8601:time_local}    )           
}
`),
			},
			wantBytes: []byte(`format "custom" "c1" {
  layout ="%%{TIMESTAMP_ISO8601:time_local}    "
}
`),
		},

		{
			name: "single grok func call with brackets and quotes ",
			args: args{
				disableTemplatesForProps: []string{"layout"},
				fileData: []byte(`format "custom" "c1" {
  layout = grok(%{TIMESTAMP_ISO8601:time_local} "static" \[%{DATA:location}\] (other))           
}
`),
			},
			wantBytes: []byte(`format "custom" "c1" {
  layout ="%%{TIMESTAMP_ISO8601:time_local} \"static\" \\[%%{DATA:location}\\] (other)"
}
`),
		},

		{
			name: "multiple grok func calls",
			args: args{
				disableTemplatesForProps: []string{"layout"},
				fileData: []byte(`format "custom" "c1" {
  layout = grok(%{TIMESTAMP_ISO8601:time_local} - %{NUMBER:event_id} - %{WORD:user} - \[%{DATA:location}\] "%{DATA:message}" %{WORD:severity} "%{DATA:additional_info}")
}
format "custom" "c2" {
  layout = grok(%{TIMESTAMP_ISO8601:time_local} - %{NUMBER:event_id} - %{WORD:user} - \[%{DATA:location}\] "%{DATA:message}" %{WORD:severity} "%{DATA:additional_info}")
}
`),
			},
			wantBytes: []byte(`format "custom" "c1" {
  layout ="%%{TIMESTAMP_ISO8601:time_local} - %%{NUMBER:event_id} - %%{WORD:user} - \\[%%{DATA:location}\\] \"%%{DATA:message}\" %%{WORD:severity} \"%%{DATA:additional_info}\""
}
format "custom" "c2" {
  layout ="%%{TIMESTAMP_ISO8601:time_local} - %%{NUMBER:event_id} - %%{WORD:user} - \\[%%{DATA:location}\\] \"%%{DATA:message}\" %%{WORD:severity} \"%%{DATA:additional_info}\""
}
`),
		},
		{
			name: "multiple grok func calls with additional blocks",
			args: args{
				disableTemplatesForProps: []string{"layout"},
				fileData: []byte(`partition "my_dynamic_log" "test"{
  source "file"  {
    format = format.custom.c1
    paths = ["/Users/kai/tailpipe_data/dynamic_logs"]
    file_layout   = ".log$"
  }
}

table  "my_dynamic_log" {
  format = format.custom.c1
  
  column "tp_timestamp"{
    source = "time_local"
  }
}

format "custom" "c1" {
  layout = grok(%{TIMESTAMP_ISO8601:time_local} - %{NUMBER:event_id} - %{WORD:user} - \[%{DATA:location}\] "%{DATA:message}" %{WORD:severity} "%{DATA:additional_info}")
}
format "custom" "c2" {
  layout = grok(%{TIMESTAMP_ISO8601:time_local} - %{NUMBER:event_id} - %{WORD:user} - \[%{DATA:location}\] "%{DATA:message}" %{WORD:severity} "%{DATA:additional_info}")
}
`),
			},
			wantBytes: []byte(`partition "my_dynamic_log" "test"{
  source "file"  {
    format = format.custom.c1
    paths = ["/Users/kai/tailpipe_data/dynamic_logs"]
    file_layout   = ".log$"
  }
}

table  "my_dynamic_log" {
  format = format.custom.c1
  
  column "tp_timestamp"{
    source = "time_local"
  }
}

format "custom" "c1" {
  layout ="%%{TIMESTAMP_ISO8601:time_local} - %%{NUMBER:event_id} - %%{WORD:user} - \\[%%{DATA:location}\\] \"%%{DATA:message}\" %%{WORD:severity} \"%%{DATA:additional_info}\""
}
format "custom" "c2" {
  layout ="%%{TIMESTAMP_ISO8601:time_local} - %%{NUMBER:event_id} - %%{WORD:user} - \\[%%{DATA:location}\\] \"%%{DATA:message}\" %%{WORD:severity} \"%%{DATA:additional_info}\""
}
`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			escapedData := GrokEscape(tt.args.fileData, "testfile.tpc")

			if got := escapedData; string(got) != string(tt.wantBytes) {
				t.Errorf("GrokEscape() = \n%v\n, want \n%v\n", string(got), string(tt.wantBytes))
			}
		})
	}
}
