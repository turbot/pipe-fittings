package constants

import "github.com/turbot/go-kit/helpers"

// Original Steampipe values
const (
	PluginExtension   = ".plugin"
	ConfigExtension   = ".spc"
	SqlExtension      = ".sql"
	MarkdownExtension = ".md"

	JsonExtension        = ".json"
	CsvExtension         = ".csv"
	TextExtension        = ".txt"
	SnapshotExtension    = ".sps"
	TokenExtension       = ".tptt"
	LegacyTokenExtension = ".sptt"
	PipelineExtension    = ".fp"
)

var YamlExtensions = []string{".yml", ".yaml"}

var ConnectionConfigExtensions = append(YamlExtensions, ConfigExtension, JsonExtension)

func IsYamlExtension(ext string) bool {
	return helpers.StringSliceContains(YamlExtensions, ext)
}
