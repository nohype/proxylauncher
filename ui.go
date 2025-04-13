// Package main provides the ProxyLauncher utility
package main

import (
	"github.com/ncruces/zenity"
)

// UI related function variables for easy mocking in tests
var showErrorMessageFunc func(message string) error = func(message string) error {
	return zenity.Error(message, zenity.Title("ProxyLauncher Error"))
}

var showInfoMessageFunc func(message string) error = func(message string) error {
	return zenity.Info(message, zenity.Title("ProxyLauncher Information"))
}

// showErrorMessageBox displays an error message box
func showErrorMessageBox(message string) {
	_ = showErrorMessageFunc(message)
}

// showInfoMessageBox displays an information message box
func showInfoMessageBox(message string) {
	_ = showInfoMessageFunc(message)
}
