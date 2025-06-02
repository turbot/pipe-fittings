package plugin

import (
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/v2/hclhelpers"
	"github.com/turbot/pipe-fittings/v2/ociinstaller"
)

const (
	LimiterSourceConfig     = "config"
	LimiterSourcePlugin     = "plugin"
	LimiterStatusActive     = "active"
	LimiterStatusOverridden = "overridden"
)

// TODO what are db tags for https://github.com/turbot/pipe-fittings/v2/issues/615
type RateLimiter struct {
	Name            string                 `hcl:"name,label" db:"name" cty:"name"`
	BucketSize      *int64                 `hcl:"bucket_size,optional" db:"bucket_size" cty:"bucket_size"`
	FillRate        *float32               `hcl:"fill_rate,optional" db:"fill_rate" cty:"fill_rate"`
	MaxConcurrency  *int64                 `hcl:"max_concurrency,optional" db:"max_concurrency" cty:"max_concurrency"`
	Scope           []string               `hcl:"scope,optional" db:"scope" cty:"scope"`
	Where           *string                `hcl:"where,optional" db:"where" cty:"where"`
	Plugin          string                 `db:"plugin" cty:"plugin"`
	PluginInstance  string                 `db:"plugin_instance" cty:"plugin_instance"`
	FileName        *string                `db:"file_name" json:"-" cty:"file_name"`
	StartLineNumber *int                   `db:"start_line_number" json:"-" cty:"start_line_number"`
	EndLineNumber   *int                   `db:"end_line_number" json:"-" cty:"end_line_number"`
	Status          string                 `db:"status" cty:"status"`
	Source          string                 `db:"source_type" cty:"source"`
	ImageRef        *ociinstaller.ImageRef `db:"-" json:"-" cty:"-"`
}

func (l *RateLimiter) OnDecoded(block *hcl.Block) {
	limiterRange := hclhelpers.BlockRange(block)
	l.FileName = &limiterRange.Filename
	l.StartLineNumber = &limiterRange.Start.Line
	l.EndLineNumber = &limiterRange.End.Line
	if l.Scope == nil {
		l.Scope = []string{}
	}
	l.Status = LimiterStatusActive
	l.Source = LimiterSourceConfig
}

func (l *RateLimiter) scopeString() string {
	scope := l.Scope
	sort.Strings(scope)
	return strings.Join(scope, "'")
}

func (l *RateLimiter) Equals(other *RateLimiter) bool {
	return l.Name == other.Name &&
		pointersHaveSameValue(l.BucketSize, other.BucketSize) &&
		pointersHaveSameValue(l.FillRate, other.FillRate) &&
		pointersHaveSameValue(l.MaxConcurrency, other.MaxConcurrency) &&
		pointersHaveSameValue(l.Where, other.Where) &&
		l.scopeString() == other.scopeString() &&
		l.Plugin == other.Plugin &&
		l.PluginInstance == other.PluginInstance &&
		l.Source == other.Source
}

func (l *RateLimiter) SetPlugin(plugin *Plugin) {
	l.PluginInstance = plugin.Instance
	l.SetPluginImageRef(plugin.Alias)
}

func (l *RateLimiter) SetPluginImageRef(alias string) {
	l.ImageRef = ociinstaller.NewImageRef(alias)
	l.Plugin = l.ImageRef.DisplayImageRef()

}

func pointersHaveSameValue[T comparable](l, r *T) bool {
	if l == nil {
		return r == nil
	}
	if r == nil {
		return false
	}
	return *l == *r
}
