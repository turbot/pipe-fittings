package filepaths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveInside joins the relative path rel onto root and returns the result.
// It returns an error unless the result is a folder strictly inside root.
//
// rel is untrusted: it is built from a mod name, branch, tag or version read from a mod file or lock file.
// A rel is refused if it is empty, absolute (or has a volume name or leading separator),
// has a '..' element (with either kind of separator), holds a NUL byte, cleans to root itself,
// or goes through a symbolic link that already exists below root.
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

	// the text is fine, but a symlink on the way (a mod's repo can contain one) would still lead out of root:
	// refuse the path if any folder or file below root that already exists is a symlink
	current := cleanRoot
	for _, element := range strings.Split(check, string(filepath.Separator)) {
		current = filepath.Join(current, element)
		info, err := os.Lstat(current)
		if err != nil {
			// this and everything below it does not exist yet
			break
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("path '%s' goes through a symbolic link", rel)
		}
	}
	return resolved, nil
}
