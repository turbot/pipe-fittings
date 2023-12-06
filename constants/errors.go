package constants

import "log/slog"

const (
	LevelTrace = slog.Level(-8)
)

const (
	// A consistent detail message for all "not a valid identifier" diagnostics.
	BadIdentifierDetail = "A name must start with a letter or underscore and may contain only letters, digits, underscores, and dashes."
	BadDependsOn        = "Invalid depends_on"
)
