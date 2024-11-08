package backend

import (
	"errors"
	"fmt"
	"time"

	"github.com/turbot/pipe-fittings/utils"
)

const (
	DefaultMaxConnLifeTime = 10 * time.Minute
	DefaultMaxConnIdleTime = 1 * time.Minute
	DefaultMaxOpenConns    = 10
)

var ErrInvalidConfig = errors.New("invalid config")

type PoolConfig struct {
	MaxConnLifeTime time.Duration
	MaxConnIdleTime time.Duration
	MaxOpenConns    int
}

type SearchPathConfig struct {
	SearchPath       []string
	SearchPathPrefix []string
}

func (c SearchPathConfig) Empty() bool {
	return len(c.SearchPath) == 0 && len(c.SearchPathPrefix) == 0
}

func (c SearchPathConfig) String() string {
	if c.Empty() {
		return ""
	}
	if len(c.SearchPath) > 0 {
		return fmt.Sprintf("search_path=%v", c.SearchPath)
	}
	return fmt.Sprintf("search_path_prefix=%v", c.SearchPathPrefix)
}

type BackendConfig struct {
	MaxConnLifeTime  time.Duration
	MaxConnIdleTime  time.Duration
	MaxOpenConns     int
	SearchPathConfig SearchPathConfig
	TimeRange        utils.TimeRange
}

func NewBackendConfig(opts []BackendOption) *BackendConfig {
	c := &BackendConfig{
		MaxConnLifeTime: DefaultMaxConnLifeTime,
		MaxConnIdleTime: DefaultMaxConnIdleTime,
		MaxOpenConns:    DefaultMaxOpenConns,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type BackendOption func(*BackendConfig)

func WithConfig(other *BackendConfig) BackendOption {
	return func(c *BackendConfig) {
		c.SearchPathConfig = other.SearchPathConfig
		c.MaxConnLifeTime = other.MaxConnLifeTime
		c.MaxConnIdleTime = other.MaxConnIdleTime
		c.MaxOpenConns = other.MaxOpenConns
	}
}

// WithSearchPathConfig sets the search path to use when connecting to the database.
// If a prefix is also set, the search path will be resolved to the first matching
// schema in the search path. Only applies if the backend is postgres
func WithSearchPathConfig(config SearchPathConfig) BackendOption {
	return func(c *BackendConfig) {
		c.SearchPathConfig = config
	}
}

func WithTimeRange(tr utils.TimeRange) BackendOption {
	return func(c *BackendConfig) {
		c.TimeRange = tr
	}
}
