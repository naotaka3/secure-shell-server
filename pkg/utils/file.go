// Package utils provides utility functions for the secure-shell-server.
package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultDirPermissions defines the default permissions for created directories.
const DefaultDirPermissions = 0o755

// EnsureLogDirectory ensures that the directory for the log file exists.
// If the directory doesn't exist, it attempts to create it.
// Returns an error if the directory creation fails.
func EnsureLogDirectory(logPath string) error {
	if logPath == "" {
		return nil // No log path specified, nothing to do
	}

	// Get the directory part of the log path
	dir := filepath.Dir(logPath)

	// Check if the directory exists
	_, err := os.Stat(dir)
	if err == nil {
		return nil // Directory exists, nothing to do
	}

	// If the error is not because the directory doesn't exist, return the error
	if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check log directory: %w", err)
	}

	// Directory doesn't exist, try to create it
	err = os.MkdirAll(dir, DefaultDirPermissions)
	if err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	return nil
}
