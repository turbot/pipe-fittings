package constants

import (
	"github.com/turbot/go-kit/helpers"
)

// Original Steampipe values
const (
	PluginExtension   = ".plugin"
	SqlExtension      = ".sql"
	MarkdownExtension = ".md"

	JsonExtension     = ".json"
	CsvExtension      = ".csv"
	TextExtension     = ".txt"
	SnapshotExtension = ".pps"
	TokenExtension    = ".tptt"
	PipelineExtension = ".fp"
)

var YamlExtensions = []string{".yml", ".yaml"}

func IsYamlExtension(ext string) bool {
	return helpers.StringSliceContains(YamlExtensions, ext)
}
