package backend

import (
	"errors"
	"time"
)

const (
	DefaultMaxConnLifeTime = 10 * time.Minute
	DefaultMaxConnIdleTime = 1 * time.Minute
	DefaultMaxOpenConns    = 10
)

var ErrInvalidConfig = errors.New("invalid config")

type BackendConfig struct {
	MaxConnLifeTime  time.Duration
	MaxConnIdleTime  time.Duration
	MaxOpenConns     int
	SearchPathConfig SearchPathConfig
	Filters          *DatabaseFilters
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

// WithFilter is a BackendOption that sets the filters for the backend.
func WithFilter(f *DatabaseFilters) BackendOption {
	return func(c *BackendConfig) {
		c.Filters = f
	}
}
