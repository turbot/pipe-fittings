package connection

import (
	"context"
	"github.com/turbot/pipe-fittings/constants"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/zclconf/go-cty/cty"
)

const TailpipeConnectionType = "tailpipe"

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

type TailpipeConnection struct {
	ConnectionImpl

	From *string `cty:"from" hcl:"from"`
	To   *string `cty:"to" hcl:"to"`
	// if an option is passed to GetConnectionString, it may override the From and To values
	OverrideFrom *string
	OverrideTo   *string

	// store the current db filename, along with the db options used to create it
	// when GetConnecitonString is called, if we have a db filename already _and the options are the same_
	// then just return the existing filename
	// if we do not have a filename, or the options are different, create a new filename

}

func NewTailpipeConnection(shortName string, declRange hcl.Range) PipelingConnection {
	return &TailpipeConnection{
		ConnectionImpl: NewConnectionImpl(TailpipeConnectionType, shortName, declRange),
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

	return nil
}

func (c *TailpipeConnection) GetConnectionString(opts ...ConnectionStringOpt) string {
	for _, opt := range opts {
		opt(c)
	}
	// TODO HACK
	return constants.DefaultSteampipeConnectionString

	// Invoke the "tailpipe connect" shell command and capture output
	cmd := exec.Command("tailpipe", "connect")

	// Run the command and get the output
	output, err := cmd.Output()
	if err != nil {
		// Handle the error, e.g., by returning an empty string or a specific error message
		return "Error executing command"
	}

	// Convert output to string, trim whitespace, and return as connection string
	connectionString := strings.TrimSpace(string(output))
	return connectionString
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

// SetTimeRange sets the time range for the connection
func (c *TailpipeConnection) SetTimeRange(from, to *time.Time) {
	if from != nil {
		fromStr := from.Format(time.RFC3339)
		c.OverrideFrom = &fromStr
	}
	if to != nil {
		toStr := to.Format(time.RFC3339)
		c.OverrideTo = &toStr
	}
}
