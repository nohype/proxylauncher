package main

import (
	"os"
	"time"
)

func main() {
	// Log execution to a temp file for test validation
	tempFile := os.Getenv("TEST_OUTPUT_FILE")
	if tempFile == "" {
		tempFile = "test-gui-output.txt"
	}
	f, err := os.Create(tempFile)
	if err == nil {
		f.WriteString("Simulated GUI app running\n")
		f.Close()
	}

	// Simulate a running GUI application by sleeping
	// On Windows, hideTarget can still apply to this process via SysProcAttr.HideWindow
	time.Sleep(5 * time.Second)
	os.Exit(0)
}
