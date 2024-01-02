package db_client

import (
	"database/sql"
	"time"
)

type PoolOverrides struct {
	PoolSize        int
	PoolMaxLifeTime time.Duration
	PoolMaxIdleTime time.Duration
}

// applies the values in the given config if they are non-zero in PoolOverrides
func (c PoolOverrides) apply(db *sql.DB) {
	if c.PoolSize > 0 {
		db.SetMaxOpenConns(c.PoolSize)
	}
	if c.PoolMaxLifeTime > 0 {
		db.SetConnMaxLifetime(c.PoolMaxLifeTime)
	}
	if c.PoolMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(c.PoolMaxIdleTime)
	}
}

type clientConfig struct {
	userPoolSettings       PoolOverrides
	managementPoolSettings PoolOverrides
	searchPathPrefix       []string
	searchPathSuffix       []string
	searchPath             []string
}

type ClientOption func(*clientConfig)

func WithUserPoolOverride(s PoolOverrides) ClientOption {
	return func(c *clientConfig) {
		c.userPoolSettings = s
	}
}

func WithManagementPoolOverride(s PoolOverrides) ClientOption {
	return func(c *clientConfig) {
		c.managementPoolSettings = s
	}
}

func WithSearchPath(searchPath []string) ClientOption {
	return func(c *clientConfig) {
		c.searchPath = searchPath
	}
}
func WithSearchPathPrefix(searchPathPrefix []string) ClientOption {
	return func(c *clientConfig) {
		c.searchPathPrefix = searchPathPrefix
	}
}
func WithSearchPathSuffix(searchPathSuffix []string) ClientOption {
	return func(c *clientConfig) {
		c.searchPathPrefix = searchPathSuffix
	}
}
