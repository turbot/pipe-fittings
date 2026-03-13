//go:build darwin || linux
// +build darwin linux

package utils

import (
	"os"
	"testing"
)

func TestFindProcess_CurrentProcess(t *testing.T) {
	pid := os.Getpid()
	proc, err := FindProcess(pid)
	if err != nil {
		t.Fatalf("FindProcess(%d) returned error: %v", pid, err)
	}
	if proc == nil {
		t.Fatalf("FindProcess(%d) returned nil process for current process", pid)
	}
}

func TestPidExists_CurrentProcess(t *testing.T) {
	pid := os.Getpid()
	if !PidExists(pid) {
		t.Fatalf("PidExists(%d) returned false for current process", pid)
	}
}

func TestFindProcess_NonexistentPid(t *testing.T) {
	proc, err := FindProcess(-1)
	if err != nil {
		t.Fatalf("FindProcess(-1) returned error: %v", err)
	}
	if proc != nil {
		t.Fatalf("FindProcess(-1) returned non-nil process, expected nil")
	}
}

func TestPidExists_NonexistentPid(t *testing.T) {
	if PidExists(-1) {
		t.Fatalf("PidExists(-1) returned true, expected false")
	}
}
