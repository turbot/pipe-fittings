package connection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/v2/backend"
	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/zclconf/go-cty/cty"
)

const TailpipeConnectionType = "tailpipe"

type TailpipeConnectResponse struct {
	DatabaseFilepath string `json:"database_filepath,omitempty"`
	InitScriptPath   string `json:"init_script_path,omitempty"`
	Error            string `json:"error,omitempty"`
}

// TailpipeConnection represents a connection to a tailpipe database
// It uses the `tailpipe connect` command to get a connection string
// The connection string is cached based on the filters used to create the database
// so that subsequent calls with the same filters do not require calling the command again
type TailpipeConnection struct {
	ConnectionImpl

	From       *string   `cty:"from" hcl:"from"`
	To         *string   `cty:"to" hcl:"to"`
	Indexes    *[]string `cty:"indexes" hcl:"indexes"`
	Partitions *[]string `cty:"partitions" hcl:"partitions"`

	// if an option is passed to GetConnectionString, it may override the From, To, Indexes or Partitions values
	OverrideFilters *backend.DatabaseFilters

	// store a maps of connection strings, keyed by the filters used to create the db
	// this is to avoid creating a new connection string each time GetConnectionString is called
	// NOTE: these connection strings are in fact paths to init scripts. These scripts will be deleted by the client
	//  when the client is closed, to avoid buildup of files,
	//  so we need to handle the case that we have a connection string in our map for a missing file
	// (in which case we clear the map entry and return a cache miss)
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
	return nil
}

// GetConnectionString implements the ConnectionStringProvider interface
// it calls the `tailpipe connect` command to get a connection string
// it caches the connection string based on the filters used to create the database
// so that subsequent calls with the same filters do not require calling the command again
// it supports the following ConnectionStringOpt options:
// - WithFilter: to override the filters used to create the database
// NOTE: the connection string returned is either of duckdb://<db path> or duckdbinit://<init script location>
// The format depends on the tailpipe version
//   - for <= v0.6.x it is duckdb://<db path>
//   - for >= v0.7.x it is duckdbinit://<init script location>
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

	if connectionString, ok := c.getCachedConnectionString(filterKey); ok {
		return connectionString, nil
	}

	slog.Debug("TailpipeConnection.GetConnectionString cache miss, calling tailpipe connect", "args", args)

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

	// buil a connection string based on the response
	var connectionString string
	if res.DatabaseFilepath != "" {
		// for tailpipe up to v0.6.x, the response contains DatabaseFilepath
		// - use duckdb:// prefix so we create a DuckDBBackend
		connectionString = fmt.Sprintf("%s%s", backend.DuckDBConnectionStringPrefix, strings.TrimSpace(res.DatabaseFilepath))
	} else if res.InitScriptPath != "" {
		// for tailpipe v0.7.x and later, the response contains InitScriptPath
		// - use duckdbinit:// prefix so we create a DuckDBBackend
		connectionString = fmt.Sprintf("%s%s", backend.DuckDBInitConnectionStringPrefix, strings.TrimSpace(res.InitScriptPath))
	}

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

func (c *TailpipeConnection) setFilters(f *backend.DatabaseFilters) {
	c.OverrideFilters = f
}

// resolve the active filters, either from the connection or the override
func (c *TailpipeConnection) getFilters() *backend.DatabaseFilters {
	var res = &backend.DatabaseFilters{}
	if c.From != nil {
		// we have already validated the time format
		from, _ := parseTime(*c.From, time.Now())
		res.From = &from
	}
	if c.To != nil {
		// we have already validated the time format
		to, _ := parseTime(*c.To, time.Now())
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

// getCachedConnectionString checks if we have a cached connection string for the given filter key
// if we do, check whether the underlying file still exists
// NOTE: these connection strings are in fact paths to init scripts. These scripts will be deleted by the client
//
//	when the client is closed, to avoid buildup of files,
//	so we need to handle the case that we have a connection string in our map for a missing file
//
// (in which case we clear the map entry and return a cache miss)
func (c *TailpipeConnection) getCachedConnectionString(filterKey string) (string, bool) {
	connectionString, ok := c.connectionStrings[filterKey]
	// if we have no hit, return
	if !ok {
		return "", false
	}

	// so we have a hit - check if the file exists - extract the filepath from the connection string

	// connection string might be either duckdb://<db path> or duckdbinit://<init script location>
	var filePath string
	switch {
	case backend.IsDuckDBConnectionString(connectionString):
		filePath = strings.TrimPrefix(connectionString, backend.DuckDBConnectionStringPrefix)
	case backend.IsDuckDBInitConnectionString(connectionString):
		filePath = strings.TrimPrefix(connectionString, backend.DuckDBInitConnectionStringPrefix)
	default:
		// unknown format - return as is - let the backend code handle it
		return connectionString, true
	}

	// so we have a filepath, check it exists
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		slog.Info("TailpipeConnection.getCachedConnectionString: cached connection string file does not exist, removing from cache", "file", filePath)
		// file does not exist - remove from cache and return miss
		delete(c.connectionStrings, filterKey)
		return "", false
	}
	// file exists - return the connection string
	return connectionString, true
}

// WithFilter is a ConnectionStringOpt that sets the filters for the connection
// it currently only supports TailpipeConnection
func WithFilter(f *backend.DatabaseFilters) ConnectionStringOpt {
	return func(c ConnectionStringProvider) {

		// if this connection supports filter, set it
		type filterSetter interface {
			setFilters(f *backend.DatabaseFilters)
		}
		if setter, ok := c.(filterSetter); ok {
			setter.setFilters(f)
		}
	}
}

// This is a duplicate of the function in parse/time.go. We have to duplicate it since we are not
// able to import the package due to circular dependencies.
// The alternative would be to move the function to a different package, but that would mean a breaking
// change for all users of the function.
// TODO: this is a temporary tactical solution, we will eventually split the pipe-fittings repo into
// two separate repos: one for the turbot IP code and one for the utilities code. At that point we can
// move the whole parse package to the new repo, use that and remove the duplicate code.
// https://github.com/turbot/pipe-fittings/issues/716

// parseTime parses a time string into a time.Time object.
func parseTime(input string, now time.Time) (time.Time, error) {
	// short-circuit if time is relative
	if strings.HasPrefix(input, "T-") {
		return parseRelativeTime(input, now)
	}

	// Handle absolute time formats using go-kit helpers.ParseTime
	t, err := helpers.ParseTime(input)
	if err != nil {
		// TODO #error improve the error message to link to docs for supported formats: https://github.com/turbot/pipe-fittings/issues/639
		return time.Time{}, err
	}

	// normalize to UTC
	return t.UTC(), nil
}

// parseRelativeTime parses relative time strings.
func parseRelativeTime(input string, now time.Time) (time.Time, error) {
	if len(input) < 3 || !strings.HasPrefix(input, "T-") {
		return time.Time{}, errors.New(constants.InvalidRelativeTimeFormat)
	}

	// Extract the value and unit
	relative := input[2:]
	unit := relative[len(relative)-1]
	value, err := strconv.Atoi(relative[:len(relative)-1])
	if err != nil {
		return time.Time{}, errors.New(constants.InvalidRelativeTimeFormat)
	}

	// Calculate the resulting time
	switch unit {
	case 'Y': // Years
		return now.AddDate(-value, 0, 0), nil
	case 'm': // Months
		return now.AddDate(0, -value, 0), nil
	case 'W': // Weeks
		return now.AddDate(0, 0, -value*7), nil
	case 'd': // Days
		return now.AddDate(0, 0, -value), nil
	case 'H': // Hours
		return now.Add(time.Duration(-value) * time.Hour), nil
	case 'M': // Minutes
		return now.Add(time.Duration(-value) * time.Minute), nil
	default:
		return time.Time{}, errors.New(constants.InvalidRelativeTimeFormat)
	}
}
