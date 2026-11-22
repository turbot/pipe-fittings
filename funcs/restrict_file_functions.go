package funcs

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/turbot/pipe-fittings/v2/app_specific"
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

// errOutsideFolder is the reason a confined file function gives for a path outside its folder
const errOutsideFolder = "path is outside the mod folder"

// RestrictFileFunctionsEnabled reports whether the app's RESTRICT_FILE_FUNCTIONS env var
// (for example POWERPIPE_RESTRICT_FILE_FUNCTIONS) is set to a true value, as strconv.ParseBool reads it.
// It is off when the var is unset, empty or not a boolean, and when the app has not set its env var keys.
func RestrictFileFunctionsEnabled() bool {
	key := app_specific.EnvRestrictFileFunctions
	if key == "" {
		return false
	}
	enabled, err := strconv.ParseBool(os.Getenv(key))
	return err == nil && enabled
}

// confineFileFunctions replaces the file functions in ctxFuncs with versions that only read inside baseDir,
// with symlinks resolved, so that a mod cannot read a file outside its own folder.
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
// refused when it lies outside baseDir, and the inner function is called with the resolved path, so the
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

// confineFileSetFunc wraps fileset: the folder argument must resolve inside baseDir, the pattern may not
// climb out of it, and any match that resolves outside baseDir (through a symlink) is dropped.
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

			root, err := resolveRoot(baseDir)
			if err != nil {
				return cty.UnknownVal(retType), function.NewArgError(0, confineError("fileset", path, pathMarks, err))
			}
			resolved, err := confinePath(baseDir, path)
			if err != nil {
				return cty.UnknownVal(retType), function.NewArgError(0, confineError("fileset", path, pathMarks, err))
			}
			// the pattern is joined to the folder by fileset itself; a pattern that climbs out with ".."
			// would glob (and walk) folders outside the mod folder
			if !isWithin(root, filepath.Join(resolved, pattern)) {
				return cty.UnknownVal(retType), function.NewArgError(1, confineError("fileset", pattern, patternMarks, errors.New(errOutsideFolder)))
			}

			result, err := inner.Call([]cty.Value{cty.StringVal(resolved).WithMarks(pathMarks), args[1]})
			if err != nil {
				return result, err
			}

			matches, resultMarks := result.Unmark()
			if !matches.IsKnown() || matches.IsNull() || matches.LengthInt() == 0 {
				return result, nil
			}
			var kept []cty.Value
			for it := matches.ElementIterator(); it.Next(); {
				_, match := it.Element()
				full := filepath.Join(resolved, filepath.FromSlash(match.AsString()))
				target, err := filepath.EvalSymlinks(full)
				// a match that can no longer be resolved, or that resolves outside the folder, is left out
				if err != nil || !isWithin(root, target) {
					continue
				}
				kept = append(kept, match)
			}
			if len(kept) == 0 {
				return cty.SetValEmpty(cty.String).WithMarks(resultMarks), nil
			}
			return cty.SetVal(kept).WithMarks(resultMarks), nil
		},
	})
}

// confinePath returns path as the file functions read it (~ expanded, relative to baseDir, cleaned),
// absolute and with every symlink resolved, or an error when that lies outside baseDir.
func confinePath(baseDir, path string) (string, error) {
	root, err := resolveRoot(baseDir)
	if err != nil {
		return "", err
	}

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

	resolved, err := resolveExisting(abs)
	if err != nil {
		return "", err
	}
	if !isWithin(root, resolved) {
		return "", errors.New(errOutsideFolder)
	}
	return resolved, nil
}

// resolveRoot returns baseDir made absolute with its symlinks resolved. It fails, so the file functions
// fail closed, when the folder cannot be resolved.
func resolveRoot(baseDir string) (string, error) {
	abs, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("cannot resolve the mod folder: %w", err)
	}
	root, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("cannot resolve the mod folder: %w", err)
	}
	return root, nil
}

// resolveExisting resolves the symlinks in an absolute, clean path. When the path does not exist, the
// deepest folder above it that does is resolved and the missing part appended, so a missing file can still
// be checked against the folder (and reported missing by the function itself).
func resolveExisting(path string) (string, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolved, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	parent := filepath.Dir(path)
	if parent == path {
		return "", err
	}
	resolvedParent, err := resolveExisting(parent)
	if err != nil {
		return "", err
	}
	return filepath.Join(resolvedParent, filepath.Base(path)), nil
}

// isWithin reports whether path, once cleaned, is root or lies under it
func isWithin(root, path string) bool {
	rel, err := filepath.Rel(root, filepath.Clean(path))
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
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

// confineError formats a refusal as `file("/etc/passwd"): path is outside the mod folder`, hiding a
// path that is marked sensitive
func confineError(name, path string, pathMarks cty.ValueMarks, err error) error {
	shown := fmt.Sprintf("%q", path)
	if marks.Has(cty.DynamicVal.WithMarks(pathMarks), marks.Sensitive) {
		shown = "(sensitive value)"
	}
	return fmt.Errorf("%s(%s): %w", name, shown, err)
}
