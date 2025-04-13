//go:build windows
// +build windows

package main

import (
	"os/exec"
	"syscall"
)

// hideTargetWindow configures a command to hide its window when executed
// This is only effective on Windows
func hideTargetWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
