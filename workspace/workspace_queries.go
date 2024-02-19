package workspace

import (
	"fmt"
	"github.com/turbot/pipe-fittings/schema"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/parse"
)

// ResolveQueryAndArgsFromSQLString attempts to resolve 'arg' to a query and query args
func (w *Workspace) ResolveQueryAndArgsFromSQLString(sqlString string) (modconfig.QueryProvider, *modconfig.QueryArgs, error) {
	var err error

	// 1) check if this is a resource
	// if this looks like a named query provider invocation, parse the sql string for arguments
	resource, args, err := w.extractQueryProviderFromQueryString(sqlString)
	if err != nil {
		return nil, nil, err
	}

	if resource != nil {
		slog.Debug("query string is a query provider resource", "resourceName", resource.Name())

		return resource, args, nil
	}

	//// 3) just use the query string as is and assume it is valid SQL
	//return &modconfig.ResolvedQuery{RawSQL: sqlString, ExecuteSQL: sqlString}, nil, nil
	return nil, nil, fmt.Errorf("could not resolve query from SQL string: %s", sqlString)

}

// ResolveQueryFromQueryProvider resolves the query for the given QueryProvider
func (w *Workspace) ResolveQueryFromQueryProvider(queryProvider modconfig.QueryProvider, runtimeArgs *modconfig.QueryArgs) (*modconfig.ResolvedQuery, error) {
	slog.Debug("ResolveQueryFromQueryProvider", "resourceName", queryProvider.Name())

	query := queryProvider.GetQuery()
	sql := queryProvider.GetSQL()

	params := queryProvider.GetParams()

	// merge the base args with the runtime args
	var err error
	runtimeArgs, err = modconfig.MergeArgs(queryProvider, runtimeArgs)
	if err != nil {
		return nil, err
	}

	// determine the source for the query
	// - this will either be the control itself or any named query the control refers to
	// either via its SQL proper ty (passing a query name) or Query property (using a reference to a query object)

	// if a query is provided, use that to resolve the sql
	if query != nil {
		return w.ResolveQueryFromQueryProvider(query, runtimeArgs)
	}

	// must have sql is there is no query
	if sql == nil {
		return nil, fmt.Errorf("%s does not define  either a 'sql' property or a 'query' property\n", queryProvider.Name())
	}

	queryProviderSQL := typehelpers.SafeString(sql)
	slog.Debug("control defines inline SQL")

	// if the SQL refers to a named query, this is the same as if the 'Query' property is set
	if namedQueryProvider, ok := w.GetQueryProvider(queryProviderSQL); ok {
		// in this case, it is NOT valid for the query provider to define its own Param definitions
		if params != nil {
			return nil, fmt.Errorf("%s has an 'SQL' property which refers to %s, so it cannot define 'param' blocks", queryProvider.Name(), namedQueryProvider.Name())
		}
		return w.ResolveQueryFromQueryProvider(namedQueryProvider, runtimeArgs)
	}

	// so the  sql is NOT a named query
	return queryProvider.GetResolvedQuery(runtimeArgs)

}

// try to treat the input string as a file name and if it exists, return its contents
func (w *Workspace) getQueryFromFile(input string) (*modconfig.ResolvedQuery, bool, error) {
	// get absolute filename
	path, err := filepath.Abs(input)
	if err != nil {
		//nolint:golint,nilerr // if this gives any error, return not exist
		return nil, false, nil
	}

	// does it exist?
	if _, err := os.Stat(path); err != nil {
		//nolint:golint,nilerr // if this gives any error, return not exist (we may get a not found or a path too long for example)
		return nil, false, nil
	}

	// read file
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, true, err
	}

	res := &modconfig.ResolvedQuery{
		RawSQL:     string(fileBytes),
		ExecuteSQL: string(fileBytes),
	}
	return res, true, nil
}

// does the input look like a resource which can be executed as a query
// Note: if anything fails just return nil values
func (w *Workspace) extractQueryProviderFromQueryString(input string) (modconfig.QueryProvider, *modconfig.QueryArgs, error) {
	// can we extract a resource name from the string
	parsedResourceName := extractResourceNameFromQuery(input)
	if parsedResourceName == nil {
		return nil, nil, nil
	}
	// ok we managed to extract a resource name - does this resource exist?
	resource, ok := w.GetResource(parsedResourceName)
	if !ok {
		return nil, nil, nil
	}

	//- is the resource a query provider, and if so does it have a query?
	queryProvider, ok := resource.(modconfig.QueryProvider)
	if !ok {
		return nil, nil, fmt.Errorf("%s cannot be executed as a query", input)
	}

	_, args, err := parse.ParseQueryInvocation(input)
	if err != nil {
		return nil, nil, err
	}
	// success
	return queryProvider, args, nil
}

func extractResourceNameFromQuery(input string) *modconfig.ParsedResourceName {
	// remove parameters from the input string before calling ParseResourceName
	// as parameters may break parsing
	openBracketIdx := strings.Index(input, "(")
	if openBracketIdx != -1 {
		input = input[:openBracketIdx]
	}
	parsedName, err := modconfig.ParseResourceName(input)
	// do not bubble error up, just return nil parsed name
	// it is expected that this function may fail if a raw query is passed to it
	if err != nil {
		return nil
	}
	return parsedName
}

func queryLooksLikeExecutableResource(input string) (string, bool) {
	// remove parameters from the input string before calling ParseResourceName
	// as parameters may break parsing
	openBracketIdx := strings.Index(input, "(")
	if openBracketIdx != -1 {
		input = input[:openBracketIdx]
	}
	parsedName, err := modconfig.ParseResourceName(input)
	if err == nil && helpers.StringSliceContains(schema.QueryProviderBlocks, parsedName.ItemType) {
		return parsedName.ToResourceName(), true
	}
	// do not bubble error up, just return false
	return "", false

}
