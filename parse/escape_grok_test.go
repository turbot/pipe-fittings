package parse

// import (
// 	"testing"
// )

// func Test_escapeGrokArgs(t *testing.T) {
// 	type args struct {
// 		fileData                 []byte
// 		disableTemplatesForProps []string
// 	}
// 	tests := []struct {
// 		name      string
// 		args      args
// 		wantBytes []byte
// 		wantError bool
// 	}{
// 		{
// 			name: "single grok func call",
// 			args: args{
// 				disableTemplatesForProps: []string{"layout"},
// 				fileData: []byte(`format "custom" "c1" {
//   layout = grok(%{TIMESTAMP_ISO8601:tp_timestamp} - %{NUMBER:event_id} - %{WORD:user} - $begin:math:display$%{DATA:location}$end:math:display$ "%{DATA:message}" %{WORD:severity} "%{DATA:additional_info}")
// }`),
// 			},
// 			wantBytes: []byte(`format "custom" "c1" {
//   layout = grok(%%{TIMESTAMP_ISO8601:tp_timestamp} - %%{NUMBER:event_id} - %%{WORD:user} - $begin:math:display$%%{DATA:location}$end:math:display$ "%%{DATA:message}" %%{WORD:severity} "%%{DATA:additional_info}")
// }`),
// 		},
// 		{
// 			name: "db log",
// 			args: args{
// 				disableTemplatesForProps: []string{"layout"},
// 				fileData: []byte(`
// format "custom" "db_log" {
//   layout = grok(%{TIMESTAMP_ISO8601:timestamp} %{WORD:log_level} %{DATA:database} $begin:math:display$%{NUMBER:query_id}$end:math:display$ %{GREEDYDATA:query})
// }`),
// 			},
// 			wantBytes: []byte(`
// format "custom" "db_log" {
//   layout = grok(%%{TIMESTAMP_ISO8601:timestamp} %%{WORD:log_level} %%{DATA:database} $begin:math:display$%%{NUMBER:query_id}$end:math:display$ %%{GREEDYDATA:query})
// }`),
// 		},
// 		{
// 			name: "multiple grok func calls",
// 			args: args{
// 				disableTemplatesForProps: []string{"layout"},
// 				fileData: []byte(`format "custom" "web_log" {
//   layout = grok(%{IPORHOST:client_ip} - %{DATA:ident} - %{DATA:user} $begin:math:display$%{HTTPDATE:timestamp}$end:math:display$ "%{WORD:method} %{URIPATHPARAM:request} HTTP/%{NUMBER:http_version}" %{NUMBER:status_code} %{NUMBER:bytes_sent})
// }
// format "custom" "db_log" {
//   layout = grok(%{TIMESTAMP_ISO8601:timestamp} %{WORD:log_level} %{DATA:database} $begin:math:display$%{NUMBER:query_id}$end:math:display$ %{GREEDYDATA:query})
// }`),
// 			},
// 			wantBytes: []byte(`format "custom" "web_log" {
//   layout = grok(%%{IPORHOST:client_ip} - %%{DATA:ident} - %%{DATA:user} $begin:math:display$%%{HTTPDATE:timestamp}$end:math:display$ "%%{WORD:method} %%{URIPATHPARAM:request} HTTP/%%{NUMBER:http_version}" %%{NUMBER:status_code} %%{NUMBER:bytes_sent})
// }
// format "custom" "db_log" {
//   layout = grok(%%{TIMESTAMP_ISO8601:timestamp} %%{WORD:log_level} %%{DATA:database} $begin:math:display$%%{NUMBER:query_id}$end:math:display$ %%{GREEDYDATA:query})
// }`),
// 		},
// 		{
// 			name: "multiple grok func calls with extra HCL blocks",
// 			args: args{
// 				disableTemplatesForProps: []string{"layout"},
// 				fileData: []byte(`partition "my_dynamic_log" "test" {
//   source "file" {
//     format = format.custom.c1
//     paths = ["/Users/kai/tailpipe_data/dynamic_logs"]
//     file_layout = ".log$"
//   }
// }

// table "my_dynamic_log" {
//   format = format.custom.c1
// }

// format "custom" "network_log" {
//   layout = grok(%{IPV4:src_ip}:%{NUMBER:src_port} -> %{IPV4:dest_ip}:%{NUMBER:dest_port} %{WORD:protocol} %{NUMBER:bytes})
// }`),
// 			},
// 			wantBytes: []byte(`partition "my_dynamic_log" "test" {
//   source "file" {
//     format = format.custom.c1
//     paths = ["/Users/kai/tailpipe_data/dynamic_logs"]
//     file_layout = ".log$"
//   }
// }

// table "my_dynamic_log" {
//   format = format.custom.c1
// }

// format "custom" "network_log" {
//   layout = grok(%%{IPV4:src_ip}:%%{NUMBER:src_port} -> %%{IPV4:dest_ip}:%%{NUMBER:dest_port} %%{WORD:protocol} %%{NUMBER:bytes})
// }`),
// 		},
// 		{
// 			name: "complex case with multiple Grok calls and non-Grok content",
// 			args: args{
// 				disableTemplatesForProps: []string{"layout"},
// 				fileData: []byte(`partition "logs_partition" "test" {
//   source "syslog" {
//     format = format.custom.syslog
//     retention_days = 30
//   }
// }

// format "custom" "system_event" {
//   layout = grok(%{TIMESTAMP_ISO8601:event_time} - %{DATA:source} - $begin:math:display$%{DATA:category}$end:math:display$ %{DATA:event_name} %{DATA:severity} "%{GREEDYDATA:details}")
// }

// format "delimited" "csv" {
//   delimiter = "\n"
//   header
// }`),
// 			},
// 			wantBytes: []byte(`partition "logs_partition" "test" {
//   source "syslog" {
//     format = format.custom.syslog
//     retention_days = 30
//   }
// }

// format "custom" "system_event" {
//   layout = grok(%%{TIMESTAMP_ISO8601:event_time} - %%{DATA:source} - $begin:math:display$%%{DATA:category}$end:math:display$ %%{DATA:event_name} %%{DATA:severity} "%%{GREEDYDATA:details}")
// }

// format "delimited" "csv" {
//   delimiter = "\n"
//   header
// }`),
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			escapedData := escapeGrokArgs(tt.args.fileData, "testfile.tpc")

// 			if got := escapedData; string(got) != string(tt.wantBytes) {
// 				t.Errorf("escapeGrokArgs() = \n%v\n, want \n%v\n", string(got), string(tt.wantBytes))
// 			}
// 		})
// 	}
// }
