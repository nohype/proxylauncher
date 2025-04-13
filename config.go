// Package main provides the ProxyLauncher utility
package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

// Configuration holds all settings for ProxyLauncher
type Configuration struct {
	Target         string
	ExtraArgs      string
	ExtraArgsOrder string
	HideTarget     bool
}

// loadConfig loads and validates the configuration from a file
func loadConfig(path string) (*Configuration, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %v", err)
	}
	defer file.Close()

	config, err := parseConfig(file)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	// Validate target exists and is an executable file
	if !fileExistsFunc(config.Target) {
		return nil, fmt.Errorf("target executable not found: %s", config.Target)
	}

	return config, nil
}

// parseConfig reads and parses the configuration file
func parseConfig(reader *os.File) (*Configuration, error) {
	config := &Configuration{}
	scanner := bufio.NewScanner(reader)

	// Track seen keys to detect duplicates
	seenKeys := make(map[string]bool)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip invalid lines
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Check for duplicate keys
		if seenKeys[key] {
			return nil, fmt.Errorf("duplicate key in config: %s", key)
		}
		seenKeys[key] = true

		// Remove quotes if present
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}

		switch strings.ToLower(key) {
		case "target":
			config.Target = value
		case "extraargs":
			config.ExtraArgs = value
		case "extraargsorder":
			lowerValue := strings.ToLower(value)
			if !slices.Contains([]string{"before", "after"}, lowerValue) {
				return nil, fmt.Errorf("invalid extraArgsOrder value %q, must be 'before' or 'after'", value)
			}
			config.ExtraArgsOrder = lowerValue
		case "hidetarget":
			lowerValue := strings.ToLower(value)
			if slices.Contains([]string{"true", "yes", "on"}, lowerValue) {
				config.HideTarget = true
			} else if slices.Contains([]string{"false", "no", "off"}, lowerValue) {
				config.HideTarget = false
			} else {
				return nil, fmt.Errorf("invalid hideTarget value %q, must be 'true/yes/on' or 'false/no/off'", value)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Validate configuration
	if config.Target == "" {
		return nil, fmt.Errorf("target executable not specified in config")
	}

	// Make extraArgsOrder required only if extraArgs is non-empty (per memory 12f6245f)
	if config.ExtraArgs != "" && config.ExtraArgsOrder == "" {
		return nil, fmt.Errorf("extraArgsOrder must be specified when extraArgs is set")
	}

	return config, nil
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

	_, err = file.WriteString(strings.Join(lines, "\n") + "\n")
	return err
}
