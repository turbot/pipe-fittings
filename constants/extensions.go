package constants

import (
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/app_specific"
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

var ConnectionConfigExtensions = append(YamlExtensions, app_specific.ConfigExtension, JsonExtension)

func IsYamlExtension(ext string) bool {
	return helpers.StringSliceContains(YamlExtensions, ext)
}
