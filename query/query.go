package query

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/turbot/pipe-fittings/v2/error_helpers"
)

// GetQueriesFromArgs retrieves queries from args
//
// For each arg check if it is a sql query or a file
func GetQueriesFromArgs(args []string) ([]string, error) {

	var queries = make([]string, len(args))
	for idx, arg := range args {
		resolvedQuery, err := resolveQueryFromSQLString(arg)
		if err != nil {
			return nil, err
		}
		if resolvedQuery != "" {
			queries[idx] = resolvedQuery
		}
	}
	return queries, nil
}

// resolveQueryFromSQLString determines whether the given string is filename or raw sql
func resolveQueryFromSQLString(sqlString string) (string, error) {
	var err error

	// 1) is this a file
	// get absolute filename
	filePath, err := filepath.Abs(sqlString)
	if err != nil {
		return "", fmt.Errorf("%s", err.Error())
	}
	fileQuery, fileExists, err := getQueryFromFile(filePath)
	if err != nil {
		return "", fmt.Errorf("%s", err.Error())
	}
	if fileExists {
		if fileQuery == "" {
			error_helpers.ShowWarning(fmt.Sprintf("file '%s' does not contain any data", filePath))
		}
		return fileQuery, nil
	}
	// the argument cannot be resolved as an existing file
	// if it has a sql suffix (i.e we believe the user meant to specify a file) return a file not found error
	if strings.HasSuffix(strings.ToLower(sqlString), ".sql") {
		return "", fmt.Errorf("file '%s' does not exist", filePath)
	}

	// 2) just use the query string as is and assume it is valid SQL
	return sqlString, nil
}

// try to treat the input string as a file name and if it exists, return its contents
func getQueryFromFile(input string) (string, bool, error) {
	// get absolute filename
	path, err := filepath.Abs(input)
	if err != nil {
		//nolint:golint,nilerr // if this gives any error, return not exist
		return "", false, nil
	}

	// does it exist?
	if _, err := os.Stat(path); err != nil {
		//nolint:golint,nilerr // if this gives any error, return not exist (we may get a not found or a path too long for example)
		return "", false, nil
	}

	// read file
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return "", true, err
	}

	return string(fileBytes), true, nil
}
