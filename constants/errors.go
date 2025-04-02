package constants

import "log/slog"

const (
	LogLevelTrace = slog.Level(-8)
	LogLevelOff   = slog.Level(-16)
)

const (
	// A consistent detail message for all "not a valid identifier" diagnostics.
	BadIdentifierDetail       = "A name must start with a letter or underscore and may contain only letters, digits, underscores, and dashes."
	BadDependsOn              = "Invalid depends_on"
	MissingVariableWarning    = "Unresolved variable: "
	InvalidRelativeTimeFormat = `Invalid relative time format

Supported formats:
  • T-2Y   (2 years ago)
  • T-10m  (10 months ago)
  • T-10W  (10 weeks ago)
  • T-180d (180 days ago)
  • T-9H   (9 hours ago)
  • T-10M  (10 minutes ago)
`
)
