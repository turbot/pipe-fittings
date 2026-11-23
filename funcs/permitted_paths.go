package funcs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar"
	"github.com/turbot/pipe-fittings/v2/permitted_paths"
	"github.com/turbot/terraform-components/lang/marks"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// singlePathFileFunctions are the context functions whose first argument is the path of one file they read
// or stat. fileset, which takes a folder and a glob pattern, is confined separately.
var singlePathFileFunctions = []string{
	"file",
	"fileexists",
	"filebase64",
	"filebase64sha256",
	"filebase64sha512",
	"filemd5",
	"filesha1",
	"filesha256",
	"filesha512",
}

// confineFileFunctions replaces the file functions in ctxFuncs with versions that only read inside the folders
// the app's PERMITTED_ROOT_PATHS setting names (see the permitted_paths package), with symlinks resolved, so
// that a mod cannot read a file outside them. Relative paths are read from baseDir, as before.
func confineFileFunctions(ctxFuncs map[string]function.Function, baseDir string) {
	for _, name := range singlePathFileFunctions {
		if inner, ok := ctxFuncs[name]; ok {
			ctxFuncs[name] = confineSinglePathFunc(name, baseDir, inner)
		}
	}
	if inner, ok := ctxFuncs["fileset"]; ok {
		ctxFuncs["fileset"] = confineFileSetFunc(baseDir, inner)
	}
}

// confineSinglePathFunc wraps a function whose first argument is a file path: the path is resolved and
// refused when it lies outside the permitted folders, and the inner function is called with the resolved path, so the
// file it opens is the one that was checked.
func confineSinglePathFunc(name, baseDir string, inner function.Function) function.Function {
	return function.New(&function.Spec{
		Description: inner.Description(),
		Params:      inner.Params(),
		VarParam:    inner.VarParam(),
		Type:        inner.ReturnTypeForValues,
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			pathArg, pathMarks := args[0].Unmark()
			path := pathArg.AsString()

			resolved, err := confinePath(baseDir, path)
			if err != nil {
				return cty.UnknownVal(retType), function.NewArgError(0, confineError(name, path, pathMarks, err))
			}

			innerArgs := append([]cty.Value{cty.StringVal(resolved).WithMarks(pathMarks)}, args[1:]...)
			return inner.Call(innerArgs)
		},
	})
}

// confineFileSetFunc replaces fileset: the folder argument must resolve inside the permitted folders, the
// pattern may not climb out of them, and a match that resolves outside (through a symlink) is left out before
// it is looked at, so neither the result nor an error says anything about a file outside. Otherwise it does
// what terraform-components' fileset does.
func confineFileSetFunc(baseDir string, inner function.Function) function.Function {
	return function.New(&function.Spec{
		Description: inner.Description(),
		Params:      inner.Params(),
		VarParam:    inner.VarParam(),
		Type:        inner.ReturnTypeForValues,
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			pathArg, pathMarks := args[0].Unmark()
			path := pathArg.AsString()
			patternArg, patternMarks := args[1].Unmark()
			pattern := patternArg.AsString()
			allMarks := []cty.ValueMarks{pathMarks, patternMarks}

			resolved, err := confinePath(baseDir, path)
			if err != nil {
				return cty.UnknownVal(retType), function.NewArgError(0, confineError("fileset", path, pathMarks, err))
			}
			// a pattern that climbs out with ".." would glob (and walk) folders outside the permitted folders
			fullPattern := filepath.Join(resolved, pattern)
			if !permitted_paths.Within(fullPattern) {
				outside := &permitted_paths.OutsideError{Path: pattern, Setting: permitted_paths.Setting()}
				return cty.UnknownVal(retType), function.NewArgError(1, confineError("fileset", pattern, patternMarks, outside))
			}

			matches, err := doublestar.Glob(fullPattern)
			if err != nil {
				return cty.UnknownVal(retType), fmt.Errorf("failed to glob pattern %s: %w", redact(pattern, allMarks...), err)
			}

			var matchVals []cty.Value
			for _, match := range permitted_paths.Filter(matches) {
				fi, err := os.Stat(match)
				if err != nil {
					return cty.UnknownVal(retType), fmt.Errorf("failed to stat %s: %w", redact(match, allMarks...), err)
				}
				if !fi.Mode().IsRegular() {
					continue
				}
				rel, err := filepath.Rel(resolved, match)
				if err != nil {
					return cty.UnknownVal(retType), fmt.Errorf("failed to trim path of match %s: %w", redact(match, allMarks...), err)
				}
				// forward slashes, for cross-system compatibility, as fileset does
				matchVals = append(matchVals, cty.StringVal(filepath.ToSlash(rel)))
			}

			if len(matchVals) == 0 {
				return cty.SetValEmpty(cty.String).WithMarks(allMarks...), nil
			}
			return cty.SetVal(matchVals).WithMarks(allMarks...), nil
		},
	})
}

// confinePath returns path as the file functions read it (~ expanded, relative to baseDir, cleaned),
// absolute and with every symlink resolved, or an error when that lies outside the permitted folders.
func confinePath(baseDir, path string) (string, error) {
	expanded, err := expandHome(path)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(expanded) {
		expanded = filepath.Join(baseDir, expanded)
	}
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", err
	}
	return permitted_paths.Check(abs)
}

// expandHome expands a leading ~ the way the file functions do (github.com/mitchellh/go-homedir), so the
// check sees the same path they would open
func expandHome(path string) (string, error) {
	if path == "" || path[0] != '~' {
		return path, nil
	}
	if len(path) > 1 && path[1] != '/' && path[1] != '\\' {
		return "", errors.New("cannot expand user-specific home dir")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to expand ~: %w", err)
	}
	return filepath.Join(home, path[1:]), nil
}

// confineError formats a refusal as `file("/etc/passwd"): path is outside the folders
// POWERPIPE_PERMITTED_ROOT_PATHS permits`, hiding a path that is marked sensitive
func confineError(name, path string, pathMarks cty.ValueMarks, err error) error {
	var outside *permitted_paths.OutsideError
	if errors.As(err, &outside) {
		err = errors.New(outside.Reason())
	}
	return fmt.Errorf("%s(%s): %w", name, redact(path, pathMarks), err)
}

// redact quotes a path for an error message, or hides it when any of its marks is sensitive, as
// terraform-components' file functions do
func redact(path string, pathMarks ...cty.ValueMarks) string {
	if marks.Has(cty.DynamicVal.WithMarks(pathMarks...), marks.Sensitive) {
		return "(sensitive value)"
	}
	return fmt.Sprintf("%q", path)
}
