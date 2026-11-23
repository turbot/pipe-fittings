package filepaths

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ResolveInside joins the relative path rel onto root and returns the result.
// It returns an error unless the result is a folder strictly inside root.
//
// rel is untrusted: it is built from a mod name, branch, tag or version read from a mod file or lock file.
// A rel is refused if it is empty, absolute (or has a volume name or leading separator),
// has a '..' element (with either kind of separator), holds a NUL byte, or cleans to root itself.
// The check is on the text of the path; ResolveInsideNoSymlinks also looks at the disk.
func ResolveInside(root, rel string) (string, error) {
	if rel == "" {
		return "", fmt.Errorf("path is empty")
	}
	if strings.ContainsRune(rel, 0) {
		return "", fmt.Errorf("path '%s' contains a NUL byte", rel)
	}
	if filepath.IsAbs(rel) || filepath.VolumeName(rel) != "" || strings.HasPrefix(rel, "/") || strings.HasPrefix(rel, `\`) {
		return "", fmt.Errorf("path '%s' is absolute", rel)
	}
	for _, element := range strings.FieldsFunc(rel, func(r rune) bool { return r == '/' || r == '\\' }) {
		if element == ".." {
			return "", fmt.Errorf("path '%s' contains '..'", rel)
		}
	}

	cleanRoot := filepath.Clean(root)
	resolved := filepath.Join(cleanRoot, rel)
	// belt and braces: confirm the cleaned result really is below root
	check, err := filepath.Rel(cleanRoot, resolved)
	if err != nil || check == "." || check == ".." || strings.HasPrefix(check, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path '%s' is not inside '%s'", rel, root)
	}

	return resolved, nil
}

// ResolveInsideNoSymlinks is ResolveInside which also refuses a path going through a symbolic link that already
// exists below root. Use it for a folder which is about to be written to or deleted and where root holds nothing
// but what was just cloned: a mod's repo can contain a symlink which would lead out of root.
// (A caller reading a root a person has set up, such as the mods folder, uses ResolveInside, as the person may
// have linked a mod folder in on purpose.)
func ResolveInsideNoSymlinks(root, rel string) (string, error) {
	resolved, err := ResolveInside(root, rel)
	if err != nil {
		return "", err
	}
	cleanRoot := filepath.Clean(root)
	check, err := filepath.Rel(cleanRoot, resolved)
	if err != nil {
		return "", err
	}
	current := cleanRoot
	for _, element := range strings.Split(check, string(filepath.Separator)) {
		current = filepath.Join(current, element)
		info, err := os.Lstat(current)
		if errors.Is(err, fs.ErrNotExist) {
			// this and everything below it does not exist yet
			break
		}
		if err != nil {
			return "", fmt.Errorf("cannot check path '%s': %w", rel, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("path '%s' goes through a symbolic link", rel)
		}
	}
	return resolved, nil
}
