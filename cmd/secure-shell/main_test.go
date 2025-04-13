package main

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// Integration tests for the main executable
// Note: These tests require the binary to be built beforehand

func TestMainCommand(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Check if binary exists
	if _, err := exec.LookPath("./secure-shell"); err != nil {
		t.Skip("skipping integration test: binary not found, run 'make build' first")
	}

	tests := []struct {
		name        string
		args        []string
		wantSuccess bool
		wantOutput  string
	}{
		{
			name:        "script with allowed commands",
			args:        []string{"-script", "echo hello\nls"},
			wantSuccess: true,
			wantOutput:  "hello",
		},
		{
			name:        "script with disallowed commands",
			args:        []string{"-script", "echo hello\nrm -rf /"},
			wantSuccess: false,
			wantOutput:  "command \"rm\" is denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a command with timeout
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			// Run the compiled binary
			cmd := exec.CommandContext(ctx, "./secure-shell")
			cmd.Args = append([]string{"./secure-shell"}, tt.args...)

			// Capture stdout and stderr
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			// Run the command
			err := cmd.Run()
			success := err == nil

			// Check if success status matches expected
			if success != tt.wantSuccess {
				t.Errorf("Command success = %v, want %v\nStdout: %s\nStderr: %s\nError: %v",
					success, tt.wantSuccess, stdout.String(), stderr.String(), err)
			}

			// Check if output contains expected string
			combinedOutput := stdout.String() + stderr.String()
			if !strings.Contains(combinedOutput, tt.wantOutput) {
				t.Errorf("Output does not contain %q\nStdout: %s\nStderr: %s",
					tt.wantOutput, stdout.String(), stderr.String())
			}
		})
	}
}
