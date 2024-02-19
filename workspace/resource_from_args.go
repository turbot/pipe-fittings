package workspace

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"strings"

	"github.com/turbot/pipe-fittings/modconfig"
)

// GetResourcesFromArgs retrieves queries from args
//
// For each arg check if it is a named query or a file, before falling back to treating it as sql
func GetResourcesFromArgs[T modconfig.ModTreeItem](args []string, w *Workspace) ([]modconfig.ModTreeItem, map[string]*modconfig.QueryArgs, error) {
	utils.LogTime("execute.GetResourcesFromArgs start")
	defer utils.LogTime("execute.GetResourcesFromArgs end")

	var targets []modconfig.ModTreeItem
	var queryArgsMap = map[string]*modconfig.QueryArgs{}
	for _, arg := range args {
		target, args, err := resolveResourceAndArgsFromSQLString[T](arg, w)
		if err != nil {
			return nil, nil, err
		}
		if target == nil {
			continue
		}
		targets = append(targets, target)
		if args != nil {
			queryArgsMap[target.GetUnqualifiedName()] = args
		}
	}
	return targets, queryArgsMap, nil
}

// resolveResourceAndArgsFromSQLString attempts to resolve 'arg' to a query and query args
func resolveResourceAndArgsFromSQLString[T modconfig.ModTreeItem](sqlString string, w *Workspace) (modconfig.ModTreeItem, *modconfig.QueryArgs, error) {
	var err error

	// 1) check if this is a resource
	// if this looks like a named query provider invocation, parse the sql string for arguments
	resource, args, err := extractResourceFromQueryString[T](sqlString, w)
	if err != nil {
		return nil, nil, err
	}

	if resource != nil {
		// success
		return resource, args, nil
	}

	// 2) if the target type is query, just use the query string as is and assume it is valid SQL
	if utils.GetGenericTypeName[T]() == schema.BlockTypeQuery {
		// TODO KAI  check whethe the sqlString looks like a resource name and if so, DO NOT create a new query (which will fail)
		q := createQueryResourceForCommandLineQuery(sqlString, w.Mod)

		// add to the workspace mod so the dashboard execution code can find it
		if err := w.Mod.AddResource(q); err != nil {
			return nil, nil, err
		}

		return q, nil, nil
	}

	// failed to resolve
	return nil, nil, nil
}

// does the input look like a resource which can be executed as a query
// Note: if anything fails just return nil values
func extractResourceFromQueryString[T modconfig.ModTreeItem](input string, w *Workspace) (modconfig.ModTreeItem, *modconfig.QueryArgs, error) {
	// can we extract a resource name from the string
	parsedResourceName, err := extractResourceNameFromQuery[T](input)
	if err != nil {
		return nil, nil, err
	}
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
		return nil, nil, sperr.New("target '%s' is not of the expected type '%s'", resource.GetUnqualifiedName(), typeName)
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
func createQueryResourceForCommandLineQuery(queryString string, mod *modconfig.Mod) *modconfig.Query {
	// build name
	shortName := "command_line_query"

	// this is NOT a named query - create the query using RawSql
	q := modconfig.NewQuery(&hcl.Block{Type: schema.BlockTypeQuery}, mod, shortName).(*modconfig.Query)
	q.SQL = utils.ToStringPointer(queryString)

	// add empty metadata
	q.SetMetadata(&modconfig.ResourceMetadata{})

	// return the new resource
	return q
}

// attempt top extra a resource name of the given type from the input string
// look at string up the the first open bracket
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

	// if the typo eis query, do not bubble error up, just return nil parsed name
	// it is expected that this function may fail if a raw query is passed to it
	if err != nil && resourceType == "query" {
		return nil, nil
	}

	return parsedName, err
}
