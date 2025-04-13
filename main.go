// ProxyLauncher - A utility to launch target applications with configurable options
package main

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Function variables that can be overridden in tests
var (
	// execCommand is a variable wrapping exec.Command for testing
	execCommand = exec.Command

	// fileExistsFunc is a function to check if a file exists
	fileExistsFunc = func(path string) bool {
		info, err := os.Stat(path)
		if err != nil {
			return false
		}
		// Check that it exists and is a regular file (not a directory)
		return !info.IsDir()
	}
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	// Determine config path (defaults to executable directory)
	cfgPath := *configPath
	if cfgPath == "" {
		execPath, err := os.Executable()
		if err != nil {
			showErrorMessageBox("Failed to determine executable path: " + err.Error())
			return
		}
		cfgPath = filepath.Join(filepath.Dir(execPath), "proxylauncher.cfg")
	}

	// Check if config file exists
	if !fileExists(cfgPath) {
		// Create default config file
		if err := createDefaultConfig(cfgPath); err != nil {
			showErrorMessageBox("Failed to create default configuration file: " + err.Error())
			return
		}

		// Open the config file with the system default editor
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command("notepad.exe", cfgPath)
		case "darwin": // macOS
			cmd = exec.Command("open", cfgPath)
		default: // Linux and others
			cmd = exec.Command("xdg-open", cfgPath)
		}

		if err := cmd.Start(); err != nil {
			showErrorMessageBox("Failed to open new default configuration file: " + err.Error())
			return
		}

		// Show message box and exit
		showInfoMessageBox("No configuration file found. A default configuration has been created. Please edit it to your needs and restart the application.")
		return
	}

	// Load configuration
	config, err := loadConfig(cfgPath)
	if err != nil {
		showErrorMessageBox(err.Error())
		return
	}

	// Create launcher
	launcher := NewLauncher(config)

	// Launch target
	if err := launcher.Launch(); err != nil {
		showErrorMessageBox(err.Error())
	}
}

// fileExists checks if a file exists and is a regular file (not a directory)
func fileExists(path string) bool {
	return fileExistsFunc(path)
}
