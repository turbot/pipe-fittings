package funcs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bmatcuk/doublestar"
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

// confineFileSetFunc replaces fileset: the folder argument must resolve inside baseDir, the pattern may not
// climb out of it, and a match that resolves outside baseDir (through a symlink) is left out before it is
// looked at, so neither the result nor an error says anything about a file outside. Otherwise it does what
// terraform-components' fileset does.
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

			root, err := resolveRoot(baseDir)
			if err != nil {
				return cty.UnknownVal(retType), function.NewArgError(0, confineError("fileset", path, pathMarks, err))
			}
			resolved, err := confinePath(baseDir, path)
			if err != nil {
				return cty.UnknownVal(retType), function.NewArgError(0, confineError("fileset", path, pathMarks, err))
			}
			// a pattern that climbs out with ".." would glob (and walk) folders outside the mod folder
			fullPattern := filepath.Join(resolved, pattern)
			if !isWithin(root, fullPattern) {
				return cty.UnknownVal(retType), function.NewArgError(1, confineError("fileset", pattern, patternMarks, errors.New(errOutsideFolder)))
			}

			matches, err := doublestar.Glob(fullPattern)
			if err != nil {
				return cty.UnknownVal(retType), fmt.Errorf("failed to glob pattern %s: %w", redact(pattern, allMarks...), err)
			}

			var matchVals []cty.Value
			for _, match := range matches {
				target, err := resolvePath(match)
				if err != nil || !isWithin(root, target) {
					continue
				}
				fi, err := os.Stat(target)
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

	// refuse a path that is outside before touching it, so its lookup error cannot tell a mod what exists there
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}
	if !isWithin(absBase, abs) && !isWithin(root, abs) {
		return "", errors.New(errOutsideFolder)
	}

	resolved, err := resolvePath(abs)
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

// maxSymlinks bounds the symlinks resolvePath follows, as the OS does, so a symlink loop ends
const maxSymlinks = 255

// resolvePath resolves the symlinks in an absolute, clean path one part at a time, following each symlink
// by its target whether or not that target exists. Once a part cannot be looked up (it is missing, or a
// folder on the way is not a folder or cannot be searched), the rest is appended as written: the OS cannot
// follow anything past that part either. So a missing file inside the folder can still be checked (and is
// reported missing by the function itself), and a symlink that points outside is refused whether or not its
// target exists or can be reached.
func resolvePath(path string) (string, error) {
	sep := string(filepath.Separator)
	vol := filepath.VolumeName(path)
	resolved := vol + sep
	pending := strings.Split(path[len(vol):], sep)
	links := 0
	for len(pending) > 0 {
		part := pending[0]
		pending = pending[1:]
		switch part {
		case "", ".":
			continue
		case "..":
			resolved = filepath.Dir(resolved)
			continue
		}
		next := filepath.Join(resolved, part)
		fi, err := os.Lstat(next)
		if err != nil {
			return filepath.Join(append([]string{next}, pending...)...), nil
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			resolved = next
			continue
		}
		links++
		if links > maxSymlinks {
			return "", fmt.Errorf("too many levels of symbolic links")
		}
		target, err := os.Readlink(next)
		if err != nil {
			return "", err
		}
		if filepath.IsAbs(target) {
			targetVol := filepath.VolumeName(target)
			resolved = targetVol + sep
			target = target[len(targetVol):]
		}
		// a relative target is read from the folder holding the symlink, which is resolved
		pending = append(strings.Split(target, sep), pending...)
	}
	return resolved, nil
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
