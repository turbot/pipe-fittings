package constants

// viper config keys
const (
	ConfigKeyVersion = "main.version"
	ConfigKeyCommit  = "main.commit"
	ConfigKeyDate    = "main.date"
	ConfigKeyBuiltBy = "main.builtBy"

	ConfigKeyInteractive                 = "interactive"
	ConfigKeyActiveCommand               = "cmd"
	ConfigKeyActiveCommandArgs           = "cmd_args"
	ConfigInteractiveVariables           = "interactive_var"
	ConfigKeyIsTerminalTTY               = "is_terminal"
	ConfigKeyServerSearchPath            = "server-search-path"
	ConfigKeyServerSearchPathPrefix      = "server-search-path-prefix"
	ConfigKeyBypassHomeDirModfileWarning = "bypass-home-dir-modfile-warning" //nolint: gosec // not credentials
)
