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
  • T-10M  (10 months ago)
  • T-10W  (10 weeks ago)
  • T-180d (180 days ago)
  • T-9H   (9 hours ago)
  • T-10m  (10 minutes ago)
`
	InvalidTimeFormat = `Invalid time format

Supported formats:
 • 2024-01-06
 • 2006-01-06T15:04:05
 • 2006-01-06T15:04:05.000
 • 2006-01-06T15:04:05Z07:00
`
)
