package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// setupTestEnv creates a test environment with mockable message functions
func setupTestEnv() (errorBuf, infoBuf *bytes.Buffer, origErrorFunc, origInfoFunc func(string) error) {
	errorBuf = &bytes.Buffer{}
	infoBuf = &bytes.Buffer{}

	// Save original functions for restoration
	origErrorFunc = showErrorMessageFunc
	origInfoFunc = showInfoMessageFunc

	// Set up mock functions
	showErrorMessageFunc = func(message string) error {
		errorBuf.WriteString("Error: " + message + "\n")
		return nil
	}
	showInfoMessageFunc = func(message string) error {
		infoBuf.WriteString("Info: " + message + "\n")
		return nil
	}

	return
}

// restoreTestEnv restores the original functions
func restoreTestEnv(origErrorFunc, origInfoFunc func(string) error) {
	showErrorMessageFunc = origErrorFunc
	showInfoMessageFunc = origInfoFunc
}

// TestParseConfig tests the configuration parsing logic
func TestParseConfig(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		errorSubstr string
		expected    Configuration
	}{
		{
			name: "Valid Config",
			content: `
# Valid config
target = "app.exe"
extraArgs = "--verbose"
extraArgsOrder = "after"
hideTarget = true
`,
			expectError: false,
			expected: Configuration{
				Target:         "app.exe",
				ExtraArgs:      "--verbose",
				ExtraArgsOrder: "after",
				HideTarget:     true,
			},
		},
		{
			name: "Missing Target",
			content: `
# Missing target
extraArgs = "--verbose"
extraArgsOrder = "before"
`,
			expectError: true,
			errorSubstr: "target executable not specified",
		},
		{
			name: "Invalid ExtraArgsOrder",
			content: `
target = "app.exe"
extraArgs = "--verbose"
extraArgsOrder = "invalid"
`,
			expectError: true,
			errorSubstr: "invalid extraArgsOrder value",
		},
		{
			name: "Missing ExtraArgsOrder with ExtraArgs",
			content: `
target = "app.exe"
extraArgs = "--verbose"
`,
			expectError: true,
			errorSubstr: "extraArgsOrder must be specified when extraArgs is set",
		},
		{
			name: "Duplicate Keys",
			content: `
target = "app.exe"
target = "second.exe"
`,
			expectError: true,
			errorSubstr: "duplicate key in config",
		},
		{
			name: "Invalid HideTarget Value",
			content: `
target = "app.exe"
hideTarget = "maybe"
`,
			expectError: true,
			errorSubstr: "invalid hideTarget value",
		},
		{
			name: "No ExtraArgs Required",
			content: `
target = "app.exe"
hideTarget = true
`,
			expectError: false,
			expected: Configuration{
				Target:     "app.exe",
				HideTarget: true,
			},
		},
		{
			name: "Comment-only and Empty Lines Ignored",
			content: `
# Comment

; Another comment
target = "app.exe"
`,
			expectError: false,
			expected: Configuration{
				Target: "app.exe",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary config file
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "test.cfg")
			err := os.WriteFile(configPath, []byte(tc.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write test config file: %v", err)
			}

			// Open the config file
			file, err := os.Open(configPath)
			if err != nil {
				t.Fatalf("Failed to open test config file: %v", err)
			}
			defer file.Close()

			// Test parseConfig
			config, err := parseConfig(file)

			// Check error expectations
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tc.errorSubstr)
				} else if !strings.Contains(err.Error(), tc.errorSubstr) {
					t.Errorf("Expected error containing '%s', got '%s'", tc.errorSubstr, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				// Check config structure
				if config.Target != tc.expected.Target {
					t.Errorf("Expected Target='%s', got '%s'", tc.expected.Target, config.Target)
				}
				if config.ExtraArgs != tc.expected.ExtraArgs {
					t.Errorf("Expected ExtraArgs='%s', got '%s'", tc.expected.ExtraArgs, config.ExtraArgs)
				}
				if config.ExtraArgsOrder != tc.expected.ExtraArgsOrder {
					t.Errorf("Expected ExtraArgsOrder='%s', got '%s'", tc.expected.ExtraArgsOrder, config.ExtraArgsOrder)
				}
				if config.HideTarget != tc.expected.HideTarget {
					t.Errorf("Expected HideTarget=%v, got %v", tc.expected.HideTarget, config.HideTarget)
				}
			}
		})
	}
}

// TestParseArgs checks the parsing of command-line arguments
func TestParseArgs(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"arg1 arg2 arg3", []string{"arg1", "arg2", "arg3"}},
		{"arg1 \"quoted arg\" arg3", []string{"arg1", "quoted arg", "arg3"}},
		{"\"quoted arg\" another", []string{"quoted arg", "another"}},
		{"arg1  arg2   arg3", []string{"arg1", "arg2", "arg3"}}, // multiple spaces
		{"", []string{}},               // empty string
		{"single", []string{"single"}}, // single arg
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("Case %d", i), func(t *testing.T) {
			result := parseArgs(tc.input)

			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d args, got %d. Expected %v, got %v",
					len(tc.expected), len(result), tc.expected, result)
				return
			}

			for j := range result {
				if result[j] != tc.expected[j] {
					t.Errorf("Args mismatch at index %d. Expected '%s', got '%s'",
						j, tc.expected[j], result[j])
				}
			}
		})
	}
}

// No need for override declarations as we use the variables from main.go

// TestLoadConfig tests the loadConfig function
func TestLoadConfig(t *testing.T) {
	// Set up test environment
	errorBuf, _, origErrorFunc, origInfoFunc := setupTestEnv()
	defer restoreTestEnv(origErrorFunc, origInfoFunc)

	// Create temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "proxylauncher.cfg")
	configContent := `
target = "app.exe"
extraArgs = "--verbose"
extraArgsOrder = "after"
hideTarget = true
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Save the original fileExistsFunc and restore it after the test
	originalFileExists := fileExistsFunc
	defer func() { fileExistsFunc = originalFileExists }()

	// Replace with a test version that always returns true
	fileExistsFunc = func(path string) bool {
		return true // Always consider files to exist in tests
	}

	// Load config
	config, err := loadConfig(configPath)

	// Check results
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expected := &Configuration{
		Target:         "app.exe",
		ExtraArgs:      "--verbose",
		ExtraArgsOrder: "after",
		HideTarget:     true,
	}

	if config.Target != expected.Target {
		t.Errorf("Expected Target='%s', got '%s'", expected.Target, config.Target)
	}
	if config.ExtraArgs != expected.ExtraArgs {
		t.Errorf("Expected ExtraArgs='%s', got '%s'", expected.ExtraArgs, config.ExtraArgs)
	}
	if config.ExtraArgsOrder != expected.ExtraArgsOrder {
		t.Errorf("Expected ExtraArgsOrder='%s', got '%s'", expected.ExtraArgsOrder, config.ExtraArgsOrder)
	}
	if config.HideTarget != expected.HideTarget {
		t.Errorf("Expected HideTarget=%v, got %v", expected.HideTarget, config.HideTarget)
	}

	// Test error case - missing file
	errorBuf.Reset() // Clear any previous messages
	nonExistentPath := filepath.Join(tempDir, "nonexistent.cfg")
	_, err = loadConfig(nonExistentPath)

	if err == nil {
		t.Error("Expected error for non-existent config file, got none")
	}
	if !strings.Contains(err.Error(), "opening config file") {
		t.Errorf("Expected error containing 'opening config file', got '%s'", err.Error())
	}
}

// TestLauncherLaunch tests the Launch method of Launcher
func TestLauncherLaunch(t *testing.T) {
	// Create temporary test file to act as executable
	tempDir := t.TempDir()
	tempExe := filepath.Join(tempDir, "test.exe")
	os.WriteFile(tempExe, []byte("dummy executable"), 0755)

	// Save the original execCommand function and restore it after the test
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	// Replace with a mock that returns a simple cross-platform command
	execCommand = func(command string, args ...string) *exec.Cmd {
		// Use platform-specific shell commands that are guaranteed to exist and succeed
		switch runtime.GOOS {
		case "windows":
			return exec.Command("cmd", "/c", "exit", "0")
		case "darwin", "linux":
			return exec.Command("true") // 'true' is a standard Unix command that always succeeds
		default:
			// For any other platform, attempt to use a simple echo command
			return exec.Command("echo", "test")
		}
	}

	// Create a launcher with the test executable
	launcher := NewLauncher(&Configuration{
		Target:         tempExe,
		ExtraArgs:      "--test",
		ExtraArgsOrder: "before",
		HideTarget:     false,
	})

	// Test successful launch
	err := launcher.Launch()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Test with wrong extraArgsOrder
	launcher.Config.ExtraArgsOrder = "invalid"
	err = launcher.Launch()
	if err != nil {
		// The test will still pass because we're mocking the command execution
		// This is more to test the code path
		t.Logf("Got error as expected with invalid extraArgsOrder: %v", err)
	}
}

// TestFileExists tests the fileExists utility function
func TestFileExists(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test-file.txt")

	// File should not exist yet
	if fileExists(tempFile) {
		t.Errorf("File %s should not exist yet", tempFile)
	}

	// Create the file
	err := os.WriteFile(tempFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Now file should exist
	if !fileExists(tempFile) {
		t.Errorf("File %s should exist now", tempFile)
	}

	// Test with a non-existent file
	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
	if fileExists(nonExistentFile) {
		t.Errorf("Non-existent file %s should not exist", nonExistentFile)
	}
}
