package parse

import (
	"strings"
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
			name: "in middle of string",
			args: args{
				disableTemplatesForProps: []string{"layout"},
				fileData: []byte(`format "custom" "c1" {
  layout = "something __BACKTICK__ escaped __BACKTICK__ something else"
}
`),
			},
			wantBytes: []byte(`format "custom" "c1" {
  layout = "something __BACKTICK__ escaped __BACKTICK__ something else"
}
`),
		},
		{
			name: "not in attribute",
			args: args{
				disableTemplatesForProps: []string{"layout"},
				fileData: []byte(`format "custom" "c1" {
  layout __BACKTICK__= "something escaped __BACKTICK__ something else"
}
`),
			},
			wantBytes: []byte(`format "custom" "c1" {
  layout __BACKTICK__= "something escaped __BACKTICK__ something else"
}
`),
		},
		{
			name: "halfway through string",
			args: args{
				disableTemplatesForProps: []string{"layout"},
				fileData: []byte(`format "custom" "c1" {
  layout = "something __BACKTICK__ escaped something else"
}
`),
			},
			wantBytes: []byte(`format "custom" "c1" {
  layout = "something __BACKTICK__ escaped something else"
}
`),
		},
		{
			name: "single grok func call",
			args: args{
				disableTemplatesForProps: []string{"layout"},
				fileData: []byte(`format "custom" "c1" {
  layout = __BACKTICK__%{TIMESTAMP_ISO8601:time_local} - %{NUMBER:event_id} - %{WORD:user} - \[%{DATA:location}\] "%{DATA:message}" %{WORD:severity} "%{DATA:additional_info}"__BACKTICK__
}
`),
			},
			wantBytes: []byte(`format "custom" "c1" {
  layout = "%%{TIMESTAMP_ISO8601:time_local} - %%{NUMBER:event_id} - %%{WORD:user} - \\[%%{DATA:location}\\] \"%%{DATA:message}\" %%{WORD:severity} \"%%{DATA:additional_info}\""
}
`),
		},
		{
			name: "single grok func call with white space at end",
			args: args{
				disableTemplatesForProps: []string{"layout"},
				fileData: []byte(`format "custom" "c1" {
  layout = __BACKTICK__%{TIMESTAMP_ISO8601:time_local}__BACKTICK__                   
}
`),
			},
			wantBytes: []byte(`format "custom" "c1" {
  layout = "%%{TIMESTAMP_ISO8601:time_local}"                   
}
`),
		},

		{
			name: "single grok func call with white space before backtick",
			args: args{
				disableTemplatesForProps: []string{"layout"},
				fileData: []byte(`format "custom" "c1" {
  layout = __BACKTICK__%{TIMESTAMP_ISO8601:time_local}    __BACKTICK__
}
`),
			},
			wantBytes: []byte(`format "custom" "c1" {
  layout = "%%{TIMESTAMP_ISO8601:time_local}    "
}
`),
		},

		{
			name: "single grok func call with brackets and quotes ",
			args: args{
				disableTemplatesForProps: []string{"layout"},
				fileData: []byte(`format "custom" "c1" {
  layout = __BACKTICK__%{TIMESTAMP_ISO8601:time_local} "static" \[%{DATA:location}\] (other)__BACKTICK__
}
`),
			},
			wantBytes: []byte(`format "custom" "c1" {
  layout = "%%{TIMESTAMP_ISO8601:time_local} \"static\" \\[%%{DATA:location}\\] (other)"
}
`),
		},

		{
			name: "multiple grok func calls",
			args: args{
				disableTemplatesForProps: []string{"layout"},
				fileData: []byte(`format "custom" "c1" {
  layout = __BACKTICK__%{TIMESTAMP_ISO8601:time_local} - %{NUMBER:event_id} - %{WORD:user} - \[%{DATA:location}\] "%{DATA:message}" %{WORD:severity} "%{DATA:additional_info}"__BACKTICK__
}
format "custom" "c2" {
  layout = __BACKTICK__%{TIMESTAMP_ISO8601:time_local} - %{NUMBER:event_id} - %{WORD:user} - \[%{DATA:location}\] "%{DATA:message}" %{WORD:severity} "%{DATA:additional_info}"__BACKTICK__
}
`),
			},
			wantBytes: []byte(`format "custom" "c1" {
  layout = "%%{TIMESTAMP_ISO8601:time_local} - %%{NUMBER:event_id} - %%{WORD:user} - \\[%%{DATA:location}\\] \"%%{DATA:message}\" %%{WORD:severity} \"%%{DATA:additional_info}\""
}
format "custom" "c2" {
  layout = "%%{TIMESTAMP_ISO8601:time_local} - %%{NUMBER:event_id} - %%{WORD:user} - \\[%%{DATA:location}\\] \"%%{DATA:message}\" %%{WORD:severity} \"%%{DATA:additional_info}\""
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
  layout = __BACKTICK__%{TIMESTAMP_ISO8601:time_local} - %{NUMBER:event_id} - %{WORD:user} - \[%{DATA:location}\] "%{DATA:message}" %{WORD:severity} "%{DATA:additional_info}"__BACKTICK__
}
format "custom" "c2" {
  layout = __BACKTICK__%{TIMESTAMP_ISO8601:time_local} - %{NUMBER:event_id} - %{WORD:user} - \[%{DATA:location}\] "%{DATA:message}" %{WORD:severity} "%{DATA:additional_info}"__BACKTICK__
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
  layout = "%%{TIMESTAMP_ISO8601:time_local} - %%{NUMBER:event_id} - %%{WORD:user} - \\[%%{DATA:location}\\] \"%%{DATA:message}\" %%{WORD:severity} \"%%{DATA:additional_info}\""
}
format "custom" "c2" {
  layout = "%%{TIMESTAMP_ISO8601:time_local} - %%{NUMBER:event_id} - %%{WORD:user} - \\[%%{DATA:location}\\] \"%%{DATA:message}\" %%{WORD:severity} \"%%{DATA:additional_info}\""
}
`),
		},
		{
			name: "arg with opening bracket",
			args: args{
				disableTemplatesForProps: []string{"layout"},
				fileData: []byte(`format "custom" "vpc_log_default" {
  layout = __BACKTICK__(?:%{TIMESTAMP_ISO8601:timestamp}\s+)?%{NUMBER:version} %{NUMBER:account_id} %{DATA:interface_id} (?:%{IP:srcaddr}|-) (?:%{IP:dstaddr}|-) (?:%{NUMBER:srcport}|-) (?:%{NUMBER:dstport}|-) (?:%{NUMBER:protocol}|-) (?:%{NUMBER:packets}|-) (?:%{NUMBER:bytes}|-) %{NUMBER:start_time} %{NUMBER:end_time} %{DATA:action} %{WORD:log_status}__BACKTICK__
}
`),
			},
			wantBytes: []byte(`format "custom" "vpc_log_default" {
  layout = "(?:%%{TIMESTAMP_ISO8601:timestamp}\\s+)?%%{NUMBER:version} %%{NUMBER:account_id} %%{DATA:interface_id} (?:%%{IP:srcaddr}|-) (?:%%{IP:dstaddr}|-) (?:%%{NUMBER:srcport}|-) (?:%%{NUMBER:dstport}|-) (?:%%{NUMBER:protocol}|-) (?:%%{NUMBER:packets}|-) (?:%%{NUMBER:bytes}|-) %%{NUMBER:start_time} %%{NUMBER:end_time} %%{DATA:action} %%{WORD:log_status}"
}
`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileData := strings.ReplaceAll(string(tt.args.fileData), "__BACKTICK__", "`")
			want := strings.ReplaceAll(string(tt.wantBytes), "__BACKTICK__", "`")
			got, _ := GrokEscape([]byte(fileData), "testfile.tpc")

			if len(got) != len(want) {
				t.Errorf("GrokEscape() = \n%v\n, want \n%v\n", string(got), want)
			}

			for i := 0; i < len(got); i++ {
				if got[i] != []byte(want)[i] {
					t.Errorf("GrokEscape() = \n%v\n, want \n%v\n", string(got), want)
					return
				}
			}
			if string(got) != want {
				t.Errorf("GrokEscape() = \n%v\n, want \n%v\n", string(got), want)
			}
		})
	}
}
