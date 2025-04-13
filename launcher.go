// Package main provides the ProxyLauncher utility
package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

// Launcher handles launching target applications
type Launcher struct {
	Config    *Configuration
	DebugMode bool
}

// NewLauncher creates a new launcher instance
func NewLauncher(config *Configuration) *Launcher {
	return &Launcher{
		Config:    config,
		DebugMode: os.Getenv("PROXYLAUNCHER_DEBUG") == "true", // Keep for testing only
	}
}

// Launch starts the target application with configured settings
func (l *Launcher) Launch() error {
	// Parse extraArgs if present
	args := []string{}
	if l.Config.ExtraArgs != "" {
		args = parseArgs(l.Config.ExtraArgs)
	}

	// Get received command line args (skip the first argument which is this executable's path)
	receivedArgs := os.Args[1:]

	// Combine arguments based on extraArgsOrder
	var allArgs []string
	if strings.ToLower(l.Config.ExtraArgsOrder) == "before" {
		allArgs = append(args, receivedArgs...)
	} else {
		allArgs = append(receivedArgs, args...)
	}

	// Prepare the command using our mockable execCommand
	cmd := execCommand(l.Config.Target, allArgs...)

	// Redirect I/O
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Hide window if configured, platform-specific
	if l.Config.HideTarget {
		hideTargetWindow(cmd)
	}

	// Execute
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to execute target: %v", err)
	}

	return nil
}

// parseArgs splits a string into command line arguments, respecting quoted sections
func parseArgs(argsStr string) []string {
	var args []string
	var currentArg strings.Builder
	inQuotes := false

	for _, r := range argsStr {
		switch {
		case r == '"':
			// Toggle quote state without adding quote to the argument
			inQuotes = !inQuotes
		case unicode.IsSpace(r) && !inQuotes:
			// End of an argument (only if not inside quotes)
			if currentArg.Len() > 0 {
				args = append(args, currentArg.String())
				currentArg.Reset()
			}
		default:
			// Add the character to the current argument
			currentArg.WriteRune(r)
		}
	}

	// Add the last argument if any
	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	return args
}
