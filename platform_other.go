//go:build !windows
// +build !windows

package main

import (
	"os/exec"
)

// hideTargetWindow is a no-op on non-Windows platforms
// since hiding windows is not supported on these platforms
func hideTargetWindow(cmd *exec.Cmd) {
	// Intentionally empty - hiding windows is only supported on Windows
	// This function exists to provide a consistent API across platforms
}
