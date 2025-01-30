package constants

// Default<xxx>WorkspaceContent is the content of the sample workspaces config file (workspaces.xpc.sample),
// that is created if it does not exist

const DefaultSteampipeWorkspaceContent = `
#
# For detailed descriptions, see the reference documentation
# at https://steampipe.io/docs/reference/config-files/workspace
#

# workspace "all_options" {
#   pipes_host         = "pipes.turbot.com"
#   pipes_token        = "spt_999faketoken99999999_111faketoken1111111111111"
#   install_dir        = "~/steampipe2"
#   query_timeout      = 300
#   snapshot_location  = "acme/dev"
#   workspace_database = "local" 
#   search_path        = "aws,aws_1,aws_2,gcp,gcp_1,gcp_2,slack,github"
#   search_path_prefix = "aws_all"
#   max_parallel       = 5
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
#     timing       = on   # off, on, verbose
#   }
# }
`

const DefaultPowerpipeWorkspaceContent = `
#
# For detailed descriptions, see the reference documentation
# at https://powerpipe.io/docs/reference/config-files/workspace
#

# workspace "all_options" {
# 
#   # Dashboard Server Options
#   listen              = "network"
#   port                = 9033
#   watch               = true
# 
#   # General Options
#   telemetry           = "info"
#   update_check        = true
#   log_level           = "info"
#   memory_max_mb       = "1024"
#   input               = true
#
#   benchmark_timeout   = 300
#   dashboard_timeout   = 300
# 
#   # Pipes Integration Options
#   cloud_host          = "pipes.turbot.com"
#   cloud_token         = "tpt_999faketoken99999999_111faketoken1111111111111"
#   snapshot_location   = "acme/dev"
# 
#   # DB Settings
#   database            = "postgres://steampipe@127.0.0.1:9193/steampipe"
#   query_timeout       = 300
#   max_parallel        = 5
# 
#   # Search Path Settings (Postgres-specific) 
#   search_path         = "aws,aws_1,aws_2,gcp,gcp_1,gcp_2,slack,github"
#   search_path_prefix  = "aws_all"
# 
#   # Output Options
#   output              = "csv"  
#   progress            = true
#   header              = true
#   separator           = ","
#   timing              = true
# }
`

const DefaultFlowpipeWorkspaceContent = `
#
# For detailed descriptions, see the reference documentation
# at https://flowpipe.io/docs/reference/config-files/workspace
#

# workspace "all_options" {
#  output              = "table"         # Default output format; one of: table, yaml, json (default table)
#
#  watch               = true
#  input               = true
#
#  host        = "http://localhost:7103"  # unset means "serverless" - run from pwd/mod-location
#
#   port        = 7103
#   listen      = "local"   # 'local' or 'network' (future - support postgres listen_addresses style)
#
#   log_level     = "info" # trace, debug, info, warn, error
#   memory_max_mb = "1024" # the maximum memory to allow the CLI process in MB
#  }
# }
`

const DefaultTailpipeWorkspaceContent = `
#
# For detailed descriptions, see the reference documentation
# at https://tailpipe.io/docs/reference/config-files/workspace
#
# workspace "all_options" {
#
#  # General Options
#  log_level     = "info" # trace, debug, info, warn, error
#  update_check  = true  	# check for updates on startup
#  
# }
`
