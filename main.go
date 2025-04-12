package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

// Config holds the configuration read from the .cfg file
type Config struct {
	Target         string
	ExtraArgs      string
	ExtraArgsOrder string
	HideTarget     bool
}

func main() {
	// 1. Determine own executable path
	exePath, err := getExecutablePath()
	if err != nil {
		showErrorMessageBox("Failed to determine executable path: " + err.Error())
		return
	}

	// Get directory and file name
	exeDir := filepath.Dir(exePath)
	exeFileName := filepath.Base(exePath)
	exeNameWithoutExt := strings.TrimSuffix(exeFileName, filepath.Ext(exeFileName))

	// 2.1. Read configuration file
	configPath := filepath.Join(exeDir, exeNameWithoutExt+".cfg")
	config, err := readConfig(configPath)
	if err != nil {
		// Check if the file doesn't exist
		if os.IsNotExist(err) {
			// Create default configuration file
			if err := createDefaultConfig(configPath); err != nil {
				showErrorMessageBox("Failed to create default configuration file: " + err.Error())
				return
			}

			// Open the config file in notepad
			cmd := exec.Command("notepad.exe", configPath)
			if err := cmd.Start(); err != nil {
				showErrorMessageBox("Failed to open new default configuration file in Notepad: " + err.Error())
				return
			}

			// Show message box and exit
			showInfoMessageBox("No configuration file found. A default configuration has been created. Please edit it to your needs and restart the application.")
			return
		}

		// Handle other errors
		showErrorMessageBox("Failed to read configuration file: " + err.Error())
		return
	}

	// Resolve target path if it's relative
	var targetPath string
	if filepath.IsAbs(config.Target) {
		targetPath = config.Target
	} else {
		targetPath = filepath.Join(exeDir, config.Target)
	}

	// 3. Check if target executable exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		showErrorMessageBox("Target executable not found: " + targetPath)
		return
	}

	// 4. Launch target executable with arguments
	// Parse extra args
	extraArgs := parseArgs(config.ExtraArgs)

	// Get received command line args (skip the first argument which is this executable's path)
	receivedArgs := os.Args[1:]

	// Combine arguments based on extraArgsOrder
	var allArgs []string
	if strings.ToLower(config.ExtraArgsOrder) == "before" {
		allArgs = append(extraArgs, receivedArgs...)
	} else {
		allArgs = append(receivedArgs, extraArgs...)
	}

	// Create command
	cmd := exec.Command(targetPath, allArgs...)

	// Hide window if configured
	if config.HideTarget {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}

	// Use the current working directory from which proxylauncher was launched
	// Working directory is inherited from the current process

	// Execute
	err = cmd.Run()
	if err != nil {
		showErrorMessageBox("Failed to execute target: " + err.Error())
	}
}

// getExecutablePath returns the full path of the current executable
func getExecutablePath() (string, error) {
	buffer := make([]uint16, windows.MAX_PATH)
	n, err := windows.GetModuleFileName(0, &buffer[0], windows.MAX_PATH)
	if err != nil {
		return "", err
	}
	return syscall.UTF16ToString(buffer[:n]), nil
}

// createDefaultConfig creates a default configuration file with comments
func createDefaultConfig(configPath string) error {
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the default configuration with comments
	lines := []string{
		"# Path to the target executable (absolute or relative to this config file's directory)",
		"target=",
		"",
		"# Additional arguments to pass to the target executable (can be empty)",
		"extraArgs=",
		"",
		"# Whether extra arguments come before or after the received command line arguments (valid values: before, after)",
		"extraArgsOrder=before",
		"",
		"# Whether to hide the target application's windows (valid values: true/yes/on, false/no/off)",
		"hideTarget=false",
	}

	// for _, line := range lines {
	// 	if _, err := file.WriteString(line + "\n"); err != nil {
	// 		return err
	// 	}
	// }

	_, err = file.WriteString(strings.Join(lines, "\n"))
	return err
}

// readConfig reads and parses the configuration file
func readConfig(configPath string) (Config, error) {
	config := Config{}

	file, err := os.Open(configPath)
	if err != nil {
		return config, err
	}
	defer file.Close()

	// Track key occurrences for duplicate detection
	keyCount := make(map[string]int)
	lineNumber := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip invalid lines
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Increment key count for duplicate detection
		keyCount[key]++
		if keyCount[key] > 1 {
			return config, fmt.Errorf("duplicate key %q found at line %d in configuration file", key, lineNumber)
		}

		switch strings.ToLower(key) {
		case "target":
			config.Target = value
		case "extraargs":
			config.ExtraArgs = value
		case "extraargsorder":
			config.ExtraArgsOrder = value
		case "hidetarget":
			config.HideTarget = slices.Contains([]string{"true", "yes", "on"}, strings.ToLower(value))
		}
	}

	if err := scanner.Err(); err != nil {
		return config, err
	}

	// Validate required fields
	if config.Target == "" {
		return config, fmt.Errorf("missing 'target' in configuration")
	}
	if config.ExtraArgs != "" && config.ExtraArgsOrder == "" {
		return config, fmt.Errorf("missing 'extraArgsOrder' in configuration when 'extraArgs' is specified")
	}
	if config.ExtraArgsOrder != "" && !slices.Contains([]string{"before", "after"}, strings.ToLower(config.ExtraArgsOrder)) {
		return config, fmt.Errorf("invalid 'extraArgsOrder' value (must be 'before' or 'after')")
	}

	return config, nil
}

// parseArgs parses a string of arguments, respecting quoted sections
func parseArgs(argsStr string) []string {
	if strings.TrimSpace(argsStr) == "" {
		return []string{}
	}

	var args []string
	var currentArg strings.Builder
	inQuotes := false

	for _, r := range argsStr {
		switch {
		case r == '"':
			currentArg.WriteRune(r)
			inQuotes = !inQuotes
		case ((r == ' ') || (r == '\t')) && !inQuotes:
			// End of argument
			if currentArg.Len() > 0 {
				args = append(args, currentArg.String())
				currentArg.Reset()
			}
		default:
			// Add character to current argument
			currentArg.WriteRune(r)
		}
	}

	// Add the last argument if any
	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	return args
}

// showErrorMessageBox displays an error message box
func showErrorMessageBox(message string) {
	title := "ProxyLauncher Error"
	flags := windows.MB_OK | windows.MB_ICONERROR
	windows.MessageBox(0, windows.StringToUTF16Ptr(message), windows.StringToUTF16Ptr(title), uint32(flags))
}

// showInfoMessageBox displays an information message box
func showInfoMessageBox(message string) {
	title := "ProxyLauncher Information"
	flags := windows.MB_OK | windows.MB_ICONINFORMATION
	windows.MessageBox(0, windows.StringToUTF16Ptr(message), windows.StringToUTF16Ptr(title), uint32(flags))
}
