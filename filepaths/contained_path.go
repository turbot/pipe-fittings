package filepaths

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ResolveInside joins the relative path rel onto root and returns the result.
// It returns an error unless the result is a folder strictly inside root.
//
// rel is untrusted: it is built from a mod name, branch, tag or version read from a mod file or lock file.
// A rel is refused if it is empty, absolute (or has a volume name or leading separator),
// has a '..' element (with either kind of separator), holds a NUL byte, or cleans to root itself.
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
