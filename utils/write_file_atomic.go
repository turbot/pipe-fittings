package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFileAtomic replaces the file at path with data, so a reader sees the old file or the new one and never half
// of either.
//
// The data goes to a temporary file in the same folder, which is then renamed into place. The file keeps the
// permissions it had; a file that does not exist yet gets defaultPerm. When path is a symlink the link is kept and
// the file it points to is replaced.
func WriteFileAtomic(path string, data []byte, defaultPerm os.FileMode) (err error) {
	target := path
	if resolved, resolveErr := filepath.EvalSymlinks(path); resolveErr == nil {
		target = resolved
	}

	perm := defaultPerm
	if info, statErr := os.Stat(target); statErr == nil {
		perm = info.Mode().Perm()
	}

	tmp, err := os.CreateTemp(filepath.Dir(target), filepath.Base(target)+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		if err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()

	if _, err = tmp.Write(data); err != nil {
		return fmt.Errorf("writing %s: %w", target, err)
	}
	if err = tmp.Chmod(perm); err != nil {
		return err
	}
	if err = tmp.Sync(); err != nil {
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, target)
}
