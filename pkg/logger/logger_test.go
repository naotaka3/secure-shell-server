package logger

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogger_LogCommandAttempt(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewWithWriter(buf)

	tests := []struct {
		name    string
		cmd     string
		args    []string
		allowed bool
		wantMsg string
	}{
		{
			name:    "allowed command",
			cmd:     "ls",
			args:    []string{"-l"},
			allowed: true,
			wantMsg: "ALLOWED",
		},
		{
			name:    "blocked command",
			cmd:     "rm",
			args:    []string{"-rf", "/"},
			allowed: false,
			wantMsg: "BLOCKED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset buffer
			buf.Reset()

			// Call the function
			logger.LogCommandAttempt(tt.cmd, tt.args, tt.allowed)

			// Check if the output contains the expected message
			if !strings.Contains(buf.String(), tt.wantMsg) {
				t.Errorf("LogCommandAttempt() output = %v, want to contain %v", buf.String(), tt.wantMsg)
			}

			// Check if the output contains the command
			if !strings.Contains(buf.String(), tt.cmd) {
				t.Errorf("LogCommandAttempt() output = %v, want to contain %v", buf.String(), tt.cmd)
			}
		})
	}
}

func TestLogger_LogErrorf(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewWithWriter(buf)

	tests := []struct {
		name    string
		format  string
		args    []interface{}
		wantMsg string
	}{
		{
			name:    "simple error",
			format:  "Error: %s",
			args:    []interface{}{"test error"},
			wantMsg: "[ERROR] Error: test error",
		},
		{
			name:    "complex error",
			format:  "Multiple values: %d, %s",
			args:    []interface{}{42, "test"},
			wantMsg: "[ERROR] Multiple values: 42, test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset buffer
			buf.Reset()

			// Call the function
			logger.LogErrorf(tt.format, tt.args...)

			// Check if the output contains the expected message
			if !strings.Contains(buf.String(), tt.wantMsg) {
				t.Errorf("LogErrorf() output = %v, want to contain %v", buf.String(), tt.wantMsg)
			}
		})
	}
}

func TestLogger_LogInfof(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewWithWriter(buf)

	tests := []struct {
		name    string
		format  string
		args    []interface{}
		wantMsg string
	}{
		{
			name:    "simple info",
			format:  "Info: %s",
			args:    []interface{}{"test info"},
			wantMsg: "[INFO] Info: test info",
		},
		{
			name:    "complex info",
			format:  "Multiple values: %d, %s",
			args:    []interface{}{42, "test"},
			wantMsg: "[INFO] Multiple values: 42, test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset buffer
			buf.Reset()

			// Call the function
			logger.LogInfof(tt.format, tt.args...)

			// Check if the output contains the expected message
			if !strings.Contains(buf.String(), tt.wantMsg) {
				t.Errorf("LogInfof() output = %v, want to contain %v", buf.String(), tt.wantMsg)
			}
		})
	}
}

func TestNewWithPath(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	// t.TempDir() manages cleanup automatically, so no need for defer os.RemoveAll()

	tests := []struct {
		name          string
		path          string
		expectError   bool
		expectLogging bool
		message       string
	}{
		{
			name:          "with valid path",
			path:          filepath.Join(tmpDir, "test.log"),
			expectError:   false,
			expectLogging: true,
			message:       "Test log message",
		},
		{
			name:          "with empty path",
			path:          "",
			expectError:   false,
			expectLogging: false,
			message:       "This should not be logged",
		},
		{
			name:          "with invalid path",
			path:          "/invalid/path/that/does/not/exist/test.log",
			expectError:   true,
			expectLogging: false,
			message:       "This should not be logged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create logger with path
			logger, err := NewWithPath(tt.path)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// If no error, test logging
			if err == nil {
				// Write a log message
				logger.LogInfo(tt.message)

				// Close the logger to flush file content
				logger.Close()

				// Check if log file exists and contains the message
				if tt.expectLogging {
					fileContent, err := os.ReadFile(tt.path)
					if err != nil {
						t.Errorf("Failed to read log file: %v", err)
					}

					if !strings.Contains(string(fileContent), tt.message) {
						t.Errorf("Log file does not contain expected message. Got: %s", string(fileContent))
					}
				} else if tt.path != "" {
					// Check that file is empty or doesn't exist for non-logging case
					// (but only if a path was specified)
					if _, err := os.Stat(tt.path); err == nil {
						fileContent, err := os.ReadFile(tt.path)
						if err != nil {
							t.Errorf("Failed to read log file: %v", err)
						}

						if strings.Contains(string(fileContent), tt.message) {
							t.Errorf("Log file contains message when it shouldn't: %s", string(fileContent))
						}
					}
				}
			}
		})
	}
}

func TestNew_NoOutput(_ *testing.T) {
	// Test that New() creates a logger that doesn't output anything
	logger := New()

	// There's no direct way to check if logs are being discarded,
	// but we can confirm that the logger can be used without errors
	logger.LogInfo("This should be discarded")
	logger.LogErrorf("This error should be discarded: %s", "test")
	logger.LogCommandAttempt("test", []string{"arg1", "arg2"}, true)

	// If we reached here without errors, the test passed
}
