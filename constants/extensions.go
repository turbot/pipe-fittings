package constants

import (
	"slices"

	"github.com/turbot/pipe-fittings/v2/app_specific"
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

func ConnectionConfigExtension() []string {
	return append(YamlExtensions, app_specific.ConfigExtension, JsonExtension)
}

func IsYamlExtension(ext string) bool {
	return slices.Contains(YamlExtensions, ext)
}
