//go:build !unix

package modinstaller

// tryLockFile does nothing where there is no flock: the apps ship for Linux and macOS only, so a build for another
// system installs without the lock
func tryLockFile(uintptr) (bool, error) {
	return true, nil
}
