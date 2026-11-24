//go:build unix

package modinstaller

import (
	"errors"

	"golang.org/x/sys/unix"
)

// tryLockFile asks for an exclusive lock on the open file without waiting. It returns false when another process
// holds the lock. The operating system drops the lock when the file is closed or the process ends.
func tryLockFile(fd uintptr) (bool, error) {
	err := unix.Flock(int(fd), unix.LOCK_EX|unix.LOCK_NB)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, unix.EWOULDBLOCK) {
		return false, nil
	}
	return false, err
}
