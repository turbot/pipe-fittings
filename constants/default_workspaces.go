package constants

// DefaultWorkspaceContent is the content of the sample workspaces config file(workspaces.spc.sample),
// that is created if it does not exist
const DefaultSteampipeWorkspaceContent = `
#
# For detailed descriptions, see the reference documentation
# at https://steampipe.io/docs/reference/config-files/workspace
#

# workspace "all_options" {
#   cloud_host         = "pipes.turbot.com"
#   cloud_token        = "spt_999faketoken99999999_111faketoken1111111111111"
#   install_dir        = "~/steampipe2"
#   query_timeout      = 300
#   snapshot_location  = "acme/dev"
#   workspace_database = "local" 
#   search_path        = "aws,aws_1,aws_2,gcp,gcp_1,gcp_2,slack,github"
#   search_path_prefix = "aws_all"
#   watch              = true
#   max_parallel       = 5
#   introspection      = false
#   input              = true
#   progress           = true
#   theme              = "dark"  # light, dark, plain 
#   cache              = true
#   cache_ttl          = 300
# 
# 
#   options "query" {
#     autocomplete = true
#     header       = true    # true, false
#     multi        = false   # true, false
#     output       = "table" # json, csv, table, line
#     separator    = ","     # any single char
#     timing       = true   # true, false
#   }
# 
#   options "check" {
#     header    = true    # true, false
#     output    = "text"  # brief, csv, html, json, md, text, snapshot or none (default "text")
#     separator = ","     # any single char
#     timing    = true    # true, false
#   }
#   
#   options "dashboard" {
#     browser = true
#   }
# }
`
const DefaultPowerpipeWorkspaceContent = `
#
# For detailed descriptions, see the reference documentation
# at https://steampipe.io/docs/reference/config-files/workspace
#

# workspace "all_options" {
#   cloud_host         = "pipes.turbot.com"
#   cloud_token        = "spt_999faketoken99999999_111faketoken1111111111111"
#   install_dir        = "~/steampipe2"
#   query_timeout      = 300
#   snapshot_location  = "acme/dev"
#   workspace_database = "local" 
#   search_path        = "aws,aws_1,aws_2,gcp,gcp_1,gcp_2,slack,github"
#   search_path_prefix = "aws_all"
#   watch              = true
#   max_parallel       = 5
#   introspection      = false
#   input              = true
#   progress           = true
#   theme              = "dark"  # light, dark, plain 
#   cache              = true
#   cache_ttl          = 300
# 
# 
#   options "query" {
#     autocomplete = true
#     header       = true    # true, false
#     multi        = false   # true, false
#     output       = "table" # json, csv, table, line
#     separator    = ","     # any single char
#     timing       = true   # true, false
#   }
# 
#   options "check" {
#     header    = true    # true, false
#     output    = "text"  # brief, csv, html, json, md, text, snapshot or none (default "text")
#     separator = ","     # any single char
#     timing    = true    # true, false
#   }
#   
#   options "dashboard" {
#     browser = true
#   }
# }
`
const DefaultFlowpipeWorkspaceContent = `
#
# For detailed descriptions, see the reference documentation
# at https://steampipe.io/docs/reference/config-files/workspace
#

# workspace "all_options" {
#  max_parallel        = 5
#  output              = "table"         # Default output format; one of: table, yaml, json (default table)
#
#  watch               = true
#  input               = true
#  progress            = true  # is the the flag to enable the real-time "friendly" pipeline view ?
#
#  host        = "http://localhost:7103"  # unset means "serverless" - run from pwd/mod-location
#  insecure    = false   # Skip TLS verification (default false)
#
#options "server" {
#    port        = 7103
#    listen      = "local"   # 'local' or 'network' (future - support postgres listen_addresses style)
#  }
#
#  # should these be "global" options instead of workspace options?
#  options "general" {
#    update_check  = "true" # true, false
#    telemetry     = "info" # info, none
#    log_level     = "info" # trace, debug, info, warn, error
#    memory_max_mb = "1024" # the maximum memory to allow the CLI process in MB
#  }
}
`
