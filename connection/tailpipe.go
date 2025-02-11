package connection

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/turbot/go-kit/helpers"
)

const TailpipeConnectionType = "tailpipe"

type TailpipeConnectResponse struct {
	DatabaseFilepath string `json:"database_filepath,omitempty"`
	Error            string `json:"error,omitempty"`
}

type TailpipeConnection struct {
	ConnectionImpl

	From       *string   `cty:"from" hcl:"from"`
	To         *string   `cty:"to" hcl:"to"`
	Indexes    *[]string `cty:"indexes" hcl:"indexes"`
	Partitions *[]string `cty:"partitions" hcl:"partitions"`

	// if an option is passed to GetConnectionString, it may override the From, To, Indexes or Partitions values
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
	// TODO #validate validate From and To https://github.com/turbot/powerpipe/issues/645
	return nil
}

func (c *TailpipeConnection) GetConnectionString(opts ...ConnectionStringOpt) (string, error) {
	for _, opt := range opts {
		opt(c)
	}
	args := []string{"connect", "--output", "json"}

	// resolve the filters
	filters := c.getFilters()
	if from := filters.From; from != nil {
		args = append(args, "--from", from.Format(time.RFC3339))
	}
	if to := filters.To; to != nil {
		args = append(args, "--to", to.Format(time.RFC3339))
	}

	if len(filters.Indexes) > 0 {
		args = append(args, "--index", fmt.Sprintf("\"%s\"", strings.Join(filters.Indexes, ",")))
	}

	if len(filters.Partitions) > 0 {
		args = append(args, "--partition", fmt.Sprintf("\"%s\"", strings.Join(filters.Partitions, ",")))
	}

	// see if we already have a connection string for these filters
	filterKey := filters.String()
	if connectionString, ok := c.connectionStrings[filterKey]; ok {
		return connectionString, nil
	}

	slog.Debug("TailpipeConnection.GetConnectionString cache miss, calling tailpipe", "args", args)

	// Invoke the "tailpipe connect" shell command and capture output
	cmd := exec.Command("tailpipe", args...)

	// Run the command and get the output
	op, err := cmd.Output()

	if err != nil {
		// Handle the error, e.g., by returning an empty string or a specific error message
		return "", fmt.Errorf("TailpipeConnection failed to get connection string: %w", err)
	}

	res := TailpipeConnectResponse{}
	err = json.Unmarshal(op, &res)
	if err != nil {
		return "", fmt.Errorf("'tailpipe connect' returned invalid response: %w", err)
	}

	if res.Error != "" {
		return "", fmt.Errorf("'tailpipe connect' returned an error: %s", res.Error)
	}

	// Convert output to string, trim whitespace, and return as connection string
	connectionString := fmt.Sprintf("duckdb://%s", strings.TrimSpace(res.DatabaseFilepath))

	// add to cache
	c.connectionStrings[filterKey] = connectionString

	slog.Info("GetConnectionString returned from tailpipe", "args", args, "connectionString", connectionString)

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

	if c.Indexes == nil && other.Indexes != nil {
		return false
	}

	if c.Indexes != nil && other.Indexes == nil {
		return false
	}

	if c.Indexes != nil && other.Indexes != nil && !slices.Equal(*c.Indexes, *other.Indexes) {
		return false
	}

	if c.Partitions == nil && other.Partitions != nil {
		return false
	}

	if c.Partitions != nil && other.Partitions == nil {
		return false
	}

	if c.Partitions != nil && other.Partitions != nil && !slices.Equal(*c.Partitions, *other.Partitions) {
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

	if c.Indexes != nil && len(*c.Indexes) > 0 {
		res.Indexes = *c.Indexes
	}

	if c.Partitions != nil && len(*c.Partitions) > 0 {
		res.Partitions = *c.Partitions
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

		if len(c.OverrideFilters.Indexes) > 0 {
			res.Indexes = c.OverrideFilters.Indexes
		}

		if len(c.OverrideFilters.Partitions) > 0 {
			res.Partitions = c.OverrideFilters.Partitions
		}
	}

	return res
}

// IsDynamic implements the DynamicConnectionStringProvider interface
// indicating that the connection string may change
func (c *TailpipeConnection) IsDynamic() {}

// WithFilter is a ConnectionStringOpt that sets the filters for the connection
// it currently only supports TailpipeConnection
func WithFilter(f *TailpipeDatabaseFilters) ConnectionStringOpt {
	return func(c ConnectionStringProvider) {

		// if this connection supports filter, set it
		type filterSetter interface {
			setFilters(f *TailpipeDatabaseFilters)
		}
		if setter, ok := c.(filterSetter); ok {
			setter.setFilters(f)
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
