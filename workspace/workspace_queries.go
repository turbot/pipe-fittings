package workspace

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/modconfig"
)

// GetResourcesFromArgs retrieves queries from args
//
// For each arg check if it is a named query or a file, before falling back to treating it as sql
func GetResourcesFromArgs[T modconfig.ModTreeItem](args []string, w *Workspace) ([]modconfig.ModTreeItem, map[string]*modconfig.QueryArgs, error) {
	utils.LogTime("execute.GetResourcesFromArgs start")
	defer utils.LogTime("execute.GetResourcesFromArgs end")

	var res []modconfig.ModTreeItem
	var queryArgsMap = map[string]*modconfig.QueryArgs{}
	for _, arg := range args {
		queryProvider, args, err := ResolveResourceAndArgsFromSQLString[T](arg, w)
		if err != nil {
			return nil, nil, err
		}
		res = append(res, queryProvider)
		if args != nil {
			queryArgsMap[queryProvider.GetUnqualifiedName()] = args
		}
	}
	return res, queryArgsMap, nil
}

// ResolveResourceAndArgsFromSQLString attempts to resolve 'arg' to a query and query args
func ResolveResourceAndArgsFromSQLString[T modconfig.ModTreeItem](sqlString string, w *Workspace) (modconfig.ModTreeItem, *modconfig.QueryArgs, error) {
	var err error

	// 1) check if this is a resource
	// if this looks like a named query provider invocation, parse the sql string for arguments
	resource, args, err := extractResourceFromQueryString[T](sqlString, w)
	if err != nil {
		return nil, nil, err
	}

	if resource != nil {
		slog.Debug("query string is a query provider resource", "resourceName", resource.Name())
		return resource, args, nil
	}

	// 2) just use the query string as is and assume it is valid SQL
	q, err := w.createQueryResourceForCommandLineQuery(sqlString)
	if err != nil {
		return nil, nil, err
	}
	return q, nil, nil

}

// does the input look like a resource which can be executed as a query
// Note: if anything fails just return nil values
func extractResourceFromQueryString[T modconfig.ModTreeItem](input string, w *Workspace) (modconfig.ModTreeItem, *modconfig.QueryArgs, error) {
	// can we extract a resource name from the string
	parsedResourceName, err := extractResourceNameFromQuery[T](input)

	if parsedResourceName == nil {
		return nil, nil, nil
	}
	// ok we managed to extract a resource name - does this resource exist?
	resource, ok := w.GetResource(parsedResourceName)
	if !ok {
		return nil, nil, nil
	}

	//- is the resource a query provider, and if so does it have a query?
	target, ok := resource.(T)
	if !ok {
		typeName := utils.GetGenericTypeName[T]()
		return nil, nil, sperr.New("target '%s' is not of the expected type '%s'", target.GetUnqualifiedName(), typeName)
	}

	_, args, err := parse.ParseQueryInvocation(input)
	if err != nil {
		return nil, nil, err
	}

	// success
	return target, args, nil
}

// convert the given command line query into a query resource and add to workspace
// this is to allow us to use existing dashboard execution code
func (w *Workspace) createQueryResourceForCommandLineQuery(queryString string) (*modconfig.Query, error) {
	// build name
	shortName := "command_line_query"

	// this is NOT a named query - create the query using RawSql
	q := modconfig.NewQuery(&hcl.Block{Type: schema.BlockTypeQuery}, w.Mod, shortName).(*modconfig.Query)
	q.SQL = utils.ToStringPointer(queryString)

	// add empty metadata
	q.SetMetadata(&modconfig.ResourceMetadata{})

	// add to the workspace mod so the dashboard execution code can find it
	if err := w.Mod.AddResource(q); err != nil {
		return nil, err
	}
	// return the new resource
	return q, nil
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

func extractResourceNameFromQuery[T modconfig.ModTreeItem](input string) (*modconfig.ParsedResourceName, error) {
	resourceType := utils.GetGenericTypeName[T]()
	if resourceType == "variable" {
		// name of variable resources is "var."
		resourceType = "var"
	}

	// remove parameters from the input string before calling ParseResourceName
	// as parameters may break parsing
	openBracketIdx := strings.Index(input, "(")
	if openBracketIdx != -1 {
		input = input[:openBracketIdx]
	}

	// if there is not type specified, add it
	if !strings.Contains(input, ".") {
		input = resourceType + "." + input
	}

	parsedName, err := modconfig.ParseResourceName(input)
	// do not bubble error up, just return nil parsed name
	// it is expected that this function may fail if a raw query is passed to it
	if err != nil {
		return nil, nil
	}

	// ensure the resource type matches the expected type
	if parsedName.ItemType != resourceType {
		return nil, fmt.Errorf("invalid resource type %s - expected %s", parsedName.ItemType, resourceType)
	}
	return parsedName, nil
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
