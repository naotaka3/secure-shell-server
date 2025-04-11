package logger

import (
	"bytes"
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
