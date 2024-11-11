package connection

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/zclconf/go-cty/cty"
)

const TailpipeConnectionType = "tailpipe"

type TailpipeConnection struct {
	ConnectionImpl

	From *string `cty:"from" hcl:"from"`
	To   *string `cty:"to" hcl:"to"`
	// if an option is passed to GetConnectionString, it may override the From and To values
	OverrideFilters *TailpipeDatabaseFilters

	// store a maps of connection strings, keyed by the filters used to create the db
	// this is to avoid creating a new connection string each time GetConnectionString is called, unless
	connectionStrings map[string]string
}

func NewTailpipeConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &TailpipeConnection{
		ConnectionImpl:    NewConnectionImpl(TailpipeConnectionType, shortName, declRange),
		connectionStrings: make(map[string]string),
	}
}

func (c *TailpipeConnection) GetConnectionType() string {
	return TailpipeConnectionType
}

func (c *TailpipeConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &TailpipeConnection{ConnectionImpl: c.ConnectionImpl})
	}

	// if pipes is nil, are able to get a connection string, so there is nothing to so
	return c, nil
}

func (c *TailpipeConnection) Validate() hcl.Diagnostics {
	// todo validate From and To
	return nil
}

func (c *TailpipeConnection) GetConnectionString(opts ...ConnectionStringOpt) (string, error) {
	for _, opt := range opts {
		opt(c)
	}
	args := []string{"connect"}

	// resolve the filters
	filters := c.getFilters()
	if from := filters.From; from != nil {
		args = append(args, "--from", from.Format(time.RFC3339))
	}
	if to := filters.To; to != nil {
		args = append(args, "--to", to.Format(time.RFC3339))
	}

	// see if we already have a connection string for these filters
	filterKey := filters.String()
	if connectionString, ok := c.connectionStrings[filterKey]; ok {
		slog.Info("GetConnectionString cache hit", "connectionString", connectionString)
		return connectionString, nil
	}

	slog.Debug("TailpipeConnection.GetConnectionString", "args", args)

	// Invoke the "tailpipe connect" shell command and capture output
	cmd := exec.Command("tailpipe", args...)

	// Run the command and get the output
	filename, err := cmd.Output()
	if err != nil {
		// Handle the error, e.g., by returning an empty string or a specific error message
		return "", fmt.Errorf("TailpipeConnection failed to get connection string: %w", err)
	}
	// Convert output to string, trim whitespace, and return as connection string
	connectionString := fmt.Sprintf("duckdb://%s", strings.TrimSpace(string(filename)))

	// add to cache
	c.connectionStrings[filterKey] = connectionString

	slog.Info("GetConnectionString return4ed from tailpipe", "args", args, "connectionString", connectionString)

	return connectionString, nil
}

func (c *TailpipeConnection) GetEnv() map[string]cty.Value {
	return map[string]cty.Value{}
}
func (c *TailpipeConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*TailpipeConnection)
	if !ok {
		return false
	}

	if (c.From == nil) != (other.From == nil) {
		return false
	}
	if c.From != nil && *c.From != *other.From {
		return false
	}
	if (c.To == nil) != (other.To == nil) {
		return false
	}
	if c.To != nil && *c.To != *other.To {
		return false
	}

	return c.GetConnectionImpl().Equals(other.GetConnectionImpl())
}

func (c *TailpipeConnection) CtyValue() (cty.Value, error) {
	return ctyValueForConnection(c)
}

func (c *TailpipeConnection) setFilters(f *TailpipeDatabaseFilters) {
	c.OverrideFilters = f
}

// resolve the active filters, either from the connection or the override
func (c *TailpipeConnection) getFilters() *TailpipeDatabaseFilters {
	var res = &TailpipeDatabaseFilters{}
	if c.From != nil {
		// we have already validated the time format
		from, _ := time.Parse(time.RFC3339, *c.From)
		res.From = &from
	}
	if c.To != nil {
		// we have already validated the time format
		to, _ := time.Parse(time.RFC3339, *c.To)
		res.To = &to
	}

	// if we have overrides, use them
	if c.OverrideFilters != nil {
		if c.OverrideFilters.From != nil {
			if from := res.From; from == nil || from.Before(*c.OverrideFilters.From) {
				res.From = c.OverrideFilters.From
			}
		}
		if overrideTo := c.OverrideFilters.To; overrideTo != nil {
			if to := res.To; to == nil || to.Before(*overrideTo) {
				res.To = overrideTo
			}
		}
	}
	// TODO partitions and indexes

	return res
}

// IsDynamic implements the DynamicConnectionStringProvider interface
// indicating that the connection string may change
func (c *TailpipeConnection) IsDynamic() {}

// WithFilter is a ConnectionStringOpt that sets the filters for the connection
// it currently only supports TailpipeConnection
// TODO K maybe we need a FilterableConnection interface
func WithFilter(f *TailpipeDatabaseFilters) ConnectionStringOpt {
	return func(c ConnectionStringProvider) {
		// if this connection supports time range, set it
		if c, ok := c.(*TailpipeConnection); ok {
			c.setFilters(f)
		}
	}
}

type TailpipeDatabaseFilters struct {
	// partition wildcards
	Partitions []string
	// the indexes to include
	Indexes []string
	// the data range
	From *time.Time
	To   *time.Time
}

func (o *TailpipeDatabaseFilters) Equals(other *TailpipeDatabaseFilters) bool {
	if (o == nil) != (other == nil) ||
		!slices.Equal(o.Partitions, other.Partitions) ||
		!slices.Equal(o.Indexes, other.Indexes) ||
		(o.From == nil) != (other.From == nil) ||
		o.From != nil && !o.From.Equal(*other.From) ||
		(o.To == nil) != (other.To == nil) ||
		o.To != nil && !o.To.Equal(*other.To) {
		return false
	}

	return true
}

func (o *TailpipeDatabaseFilters) String() string {
	var str strings.Builder
	if len(o.Partitions) > 0 {
		str.WriteString("partitions: ")
		str.WriteString(strings.Join(o.Partitions, ","))
	}
	if len(o.Indexes) > 0 {
		str.WriteString("indexes: ")
		str.WriteString(strings.Join(o.Indexes, ","))
	}
	if o.From != nil {
		str.WriteString("from: ")
		str.WriteString(o.From.String())
	}
	if o.To != nil {
		str.WriteString("to: ")
		str.WriteString(o.To.String())
	}
	return str.String()
}
