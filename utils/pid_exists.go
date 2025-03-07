//go:build darwin || linux
// +build darwin linux

package utils

import (
	"fmt"

	psutils "github.com/shirou/gopsutil/process"
)

// PidExists uses psutils.NewProcess, instead of signalling, since we have observed that
// signalling does not always work reliably when the destination of the signal
// is a child of the source of the signal - which may be the case then starting
// implicit services
func PidExists(targetPid int) bool {
	LogTime("utils.PidExists start")
	defer LogTime("utils.PidExists end")

	_, err := psutils.NewProcess(int32(targetPid)) //nolint: gosec	// pid will fit into int32
	return err == nil
}

// FindProcess tries to find the process with the given pid
// returns nil if the process could not be found
func FindProcess(targetPid int) (*psutils.Process, error) {
	LogTime("utils.FindProcess start")
	defer LogTime("utils.FindProcess end")

	pids, err := psutils.Pids()
	if err != nil {
		return nil, fmt.Errorf("failed to get pids")
	}
	for _, pid := range pids {
		if targetPid == int(pid) {
			//nolint: gosec	// target pdi will be 32 bit
			process, err := psutils.NewProcess(int32(targetPid))
			if err != nil {
				return nil, nil
			}

			status, err := process.Status()
			if err != nil {
				return nil, fmt.Errorf("failed to get status: %s", err.Error())
			}

			if status == "Z" {
				// this means that postgres went away, but the process itself is still a zombie.
				return nil, nil
			}
			return process, nil
		}
	}
	return nil, nil
}
