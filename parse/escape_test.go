package parse

import (
	"testing"
)

func Test_parseHclFileWithGrokProperties(t *testing.T) {
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
			name: "single target property",
			args: args{
				disableTemplatesForProps: []string{"file_layout"},
				fileData: []byte(`partition "aws_cloudtrail_log" "fs" {
    source "file_system"  {
        file_layout = "AWSLogs/%{WORD:org}/%{WORD:account_id}/CloudTrail/%{NOTSPACE:file_name}.%{WORD:ext}"
        paths = ["/Users/kai/tailpipe_data/flaws_cloudtrail_logs"]
    }
}`),
			},
			wantBytes: []byte(`partition "aws_cloudtrail_log" "fs" {
    source "file_system"  {
        file_layout = "AWSLogs/%%{WORD:org}/%%{WORD:account_id}/CloudTrail/%%{NOTSPACE:file_name}.%%{WORD:ext}"
        paths = ["/Users/kai/tailpipe_data/flaws_cloudtrail_logs"]
    }
}`),
		},
		{
			name: "already escaped",
			args: args{
				disableTemplatesForProps: []string{"file_layout"},
				fileData: []byte(`partition "aws_cloudtrail_log" "fs" {
    source "file_system"  {
        file_layout = "AWSLogs/%%{WORD:org}/%%{WORD:account_id}/CloudTrail/%%{NOTSPACE:file_name}.%%{WORD:ext}"
        paths = ["/Users/kai/tailpipe_data/flaws_cloudtrail_logs"]
    }
}`),
			},
			wantBytes: []byte(`partition "aws_cloudtrail_log" "fs" {
    source "file_system"  {
        file_layout = "AWSLogs/%%{WORD:org}/%%{WORD:account_id}/CloudTrail/%%{NOTSPACE:file_name}.%%{WORD:ext}"
        paths = ["/Users/kai/tailpipe_data/flaws_cloudtrail_logs"]
    }
}`),
		},
		{
			name: "mix of escaped and not escaped",
			args: args{
				disableTemplatesForProps: []string{"file_layout"},
				fileData: []byte(`partition "aws_cloudtrail_log" "fs" {
    source "file_system"  {
        file_layout = "AWSLogs/%%{WORD:org}/%{WORD:account_id}/CloudTrail/%%{NOTSPACE:file_name}.%{WORD:ext}"
        paths = ["/Users/kai/tailpipe_data/flaws_cloudtrail_logs"]
    }
}`),
			},
			wantBytes: []byte(`partition "aws_cloudtrail_log" "fs" {
    source "file_system"  {
        file_layout = "AWSLogs/%%{WORD:org}/%%{WORD:account_id}/CloudTrail/%%{NOTSPACE:file_name}.%%{WORD:ext}"
        paths = ["/Users/kai/tailpipe_data/flaws_cloudtrail_logs"]
    }
}`),
		},
		{
			name: "non-target property has template expression",
			args: args{
				disableTemplatesForProps: []string{"file_layout"},
				fileData: []byte(`partition "aws_cloudtrail_log" "fs" {
    source "file_system"  {
        file_layout = "AWSLogs/%{WORD:org}/%{WORD:account_id}/CloudTrail/%{NOTSPACE:file_name}.%{WORD:ext}"
        paths = %{if var.foo != "a"}foo%{else}unnamed%{ endif }
    }
}`),
			},
			// this parsing should fail
			wantError: true,
			wantBytes: []byte(`partition "aws_cloudtrail_log" "fs" {
    source "file_system"  {
        file_layout = "AWSLogs/%%{WORD:org}/%%{WORD:account_id}/CloudTrail/%%{NOTSPACE:file_name}.%%{WORD:ext}"
        paths = %{if var.foo != "a"}foo%{else}unnamed%{ endif }
    }
}`),
		},
		{
			name: "target property not present",
			args: args{
				disableTemplatesForProps: []string{"file_layout"},
				fileData: []byte(`partition "aws_cloudtrail_log" "fs" {
    source "file_system"  {
        paths = ["/Users/kai/tailpipe_data/flaws_cloudtrail_logs"]
    }
}`),
			},
			wantBytes: []byte(`partition "aws_cloudtrail_log" "fs" {
    source "file_system"  {
        paths = ["/Users/kai/tailpipe_data/flaws_cloudtrail_logs"]
    }
}`),
		},
		{
			name: "target property has no template expression",
			args: args{
				disableTemplatesForProps: []string{"file_layout"},
				fileData: []byte(`partition "aws_cloudtrail_log" "fs" {
    source "file_system"  {
        file_layout = "AWSLogs/CloudTrail"
        paths = ["/Users/kai/tailpipe_data/flaws_cloudtrail_logs"]
    }
}`),
			},
			wantBytes: []byte(`partition "aws_cloudtrail_log" "fs" {
    source "file_system"  {
        file_layout = "AWSLogs/CloudTrail"
        paths = ["/Users/kai/tailpipe_data/flaws_cloudtrail_logs"]
    }
}`),
		},
		{
			name: "multiple instance of target property ",
			args: args{
				disableTemplatesForProps: []string{"file_layout"},
				fileData: []byte(`partition "aws_cloudtrail_log" "fs" {
    source "file_system"  {
        file_layout = "AWSLogs/%{WORD:org}/%{WORD:account_id}/CloudTrail/%{NOTSPACE:file_name}.%{WORD:ext}"
        paths = ["/Users/kai/tailpipe_data/flaws_cloudtrail_logs"]
    }
}
partition "aws_cloudtrail_log" "fs_short" {
    source "file_system"  {
        file_layout = "%{WORD:base_name}.json.gz$"
        paths = ["/Users/kai/tailpipe_data/flaws_cloudtrail_logs 2"]
    }
}`),
			},
			wantBytes: []byte(`partition "aws_cloudtrail_log" "fs" {
    source "file_system"  {
        file_layout = "AWSLogs/%%{WORD:org}/%%{WORD:account_id}/CloudTrail/%%{NOTSPACE:file_name}.%%{WORD:ext}"
        paths = ["/Users/kai/tailpipe_data/flaws_cloudtrail_logs"]
    }
}
partition "aws_cloudtrail_log" "fs_short" {
    source "file_system"  {
        file_layout = "%%{WORD:base_name}.json.gz$"
        paths = ["/Users/kai/tailpipe_data/flaws_cloudtrail_logs 2"]
    }
}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			escapedData, diags := EscapeTemplateTokens(tt.args.fileData, "testfile.tpc", tt.args.disableTemplatesForProps)

			if diags.HasErrors() {
				if !tt.wantError {
					t.Errorf("parseHclFile() unexpected error: %v", diags)
					return
				}
			} else if tt.wantError {
				t.Errorf("parseHclFile() expected error, got none")
			}
			if got := escapedData; string(got) != string(tt.wantBytes) {
				t.Errorf("parseHclFile() = %v, want %v", string(got), string(tt.wantBytes))
			}
		})
	}
}
