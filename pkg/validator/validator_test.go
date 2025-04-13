package validator

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/pkg/logger"
)

// TestValidateCommand tests the ValidateCommand function with various scenarios.
func TestValidateCommand(t *testing.T) {
	// Setup test config
	cfg := &config.ShellCommandConfig{
		AllowedDirectories: []string{"/home", "/tmp"},
		AllowCommands: []config.AllowCommand{
			{Command: "ls"}, // Command with no subcommand restrictions
			{Command: "git", SubCommands: []string{"pull", "push", "status"}},                                // Command with allowed subcommands
			{Command: "docker", DenySubCommands: []string{"rm", "exec", "run"}},                              // Command with denied subcommands
			{Command: "npm", SubCommands: []string{"install", "update"}, DenySubCommands: []string{"audit"}}, // Command with both allowed and denied subcommands
		},
		DenyCommands: []config.DenyCommand{
			{Command: "rm", Message: "Remove command is not allowed"},
			{Command: "sudo"}, // With default error message
		},
		DefaultErrorMessage: "Command not allowed by security policy",
		BlockLogPath:        "", // Don't write to a log file in tests
	}

	// Create a logger with a buffer
	var logBuffer bytes.Buffer
	log := logger.NewWithWriter(&logBuffer)

	// Create the validator
	v := New(cfg, log)

	// Test cases
	tests := []struct {
		name    string
		cmd     string
		args    []string
		allowed bool
		message string
	}{
		// Test allowed commands
		{name: "AllowedCommand", cmd: "ls", args: []string{"-la"}, allowed: true, message: ""},

		// Test denied commands
		{name: "ExplicitlyDeniedCommand", cmd: "rm", args: []string{"-rf", "/tmp"}, allowed: false, message: "command \"rm\" is denied: Remove command is not allowed"},
		{name: "DeniedCommandWithDefaultMessage", cmd: "sudo", args: []string{"apt-get", "update"}, allowed: false, message: "command \"sudo\" is denied: Command not allowed by security policy"},
		{name: "UnlistedCommand", cmd: "wget", args: []string{"https://example.com"}, allowed: false, message: "command \"wget\" is not permitted: Command not allowed by security policy"},

		// Test subcommand handling
		{name: "AllowedSubcommand", cmd: "git", args: []string{"pull"}, allowed: true, message: ""},
		{name: "DisallowedSubcommand", cmd: "git", args: []string{"fetch"}, allowed: false, message: "subcommand \"fetch\" is not allowed for command \"git\""},
		{name: "DeniedSubcommand", cmd: "docker", args: []string{"rm"}, allowed: false, message: "subcommand \"rm\" is denied for command \"docker\""},
		{name: "AllowedDockerSubcommand", cmd: "docker", args: []string{"ps"}, allowed: true, message: ""},

		// Test command with both allowed and denied subcommands
		{name: "NpmWithAllowedSubcommand", cmd: "npm", args: []string{"install"}, allowed: true, message: ""},
		{name: "NpmWithDeniedSubcommand", cmd: "npm", args: []string{"audit"}, allowed: false, message: "subcommand \"audit\" is not allowed for command \"npm\""},
		{name: "NpmWithDisallowedSubcommand", cmd: "npm", args: []string{"run"}, allowed: false, message: "subcommand \"run\" is not allowed for command \"npm\""},

		// Test edge cases
		{name: "EmptyCommand", cmd: "", args: []string{}, allowed: false, message: "command \"\" is not permitted: Command not allowed by security policy"},
		{name: "AllowedCommandWithNoArgs", cmd: "ls", args: []string{}, allowed: true, message: ""},
		{name: "CommandWithAllowedSubcommandsNoArgs", cmd: "git", args: []string{}, allowed: true, message: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset log buffer for each test
			logBuffer.Reset()

			gotAllowed, gotMessage := v.ValidateCommand(tt.cmd, tt.args)
			if gotAllowed != tt.allowed {
				t.Errorf("ValidateCommand() allowed = %v, want %v", gotAllowed, tt.allowed)
			}
			if gotMessage != tt.message {
				t.Errorf("ValidateCommand() message = %q, want %q", gotMessage, tt.message)
			}
		})
	}
}

// TestCommandLogging tests the command logging functionality.
func TestCommandLogging(t *testing.T) {
	// Create a temporary directory for the log file
	tempDir := t.TempDir()

	// Create log file path
	logPath := filepath.Join(tempDir, "blocked.log")

	// Setup test config with log path
	cfg := &config.ShellCommandConfig{
		AllowedDirectories:  []string{"/home", "/tmp"},
		AllowCommands:       []config.AllowCommand{{Command: "ls"}},
		DenyCommands:        []config.DenyCommand{{Command: "rm"}},
		DefaultErrorMessage: "Command not allowed",
		BlockLogPath:        logPath,
	}

	// Create a logger
	var logBuffer bytes.Buffer
	log := logger.NewWithWriter(&logBuffer)

	// Create the validator
	v := New(cfg, log)

	// Test blocked command to trigger logging
	v.ValidateCommand("rm", []string{"-rf", "/tmp"})

	// Check if log file was created and contains the expected content
	logContent, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logStr := string(logContent)
	if !strings.Contains(logStr, "[BLOCKED] Command: rm [-rf /tmp]") {
		t.Errorf("Expected blocked command log entry, got: %s", logStr)
	}
}

// TestLogBlockedCommandError tests error handling in logBlockedCommand.
func TestLogBlockedCommandError(t *testing.T) {
	// Create a non-existent directory path
	testDir := "/non-existent-dir-" + tempDirSuffix()

	// Setup test config with invalid log path
	cfg := &config.ShellCommandConfig{
		AllowedDirectories:  []string{"/home", "/tmp"},
		AllowCommands:       []config.AllowCommand{{Command: "ls"}},
		DenyCommands:        []config.DenyCommand{{Command: "rm"}},
		DefaultErrorMessage: "Command not allowed",
		BlockLogPath:        filepath.Join(testDir, "blocked.log"), // Path that likely can't be written to
	}

	// Create a logger with a buffer to capture error logs
	var logBuffer bytes.Buffer
	log := logger.NewWithWriter(&logBuffer)

	// Create the validator
	v := New(cfg, log)

	// Test blocked command to trigger logging attempt
	v.ValidateCommand("rm", []string{"-rf", "/tmp"})

	// Check if error was logged
	if !strings.Contains(logBuffer.String(), "Failed to create directory for block log") {
		// This is a bit tricky since we're testing with a non-existent path that might
		// actually be writable on some systems. If the test fails, it might be because
		// the non-existent directory was creatable.
		if _, err := os.Stat(testDir); os.IsNotExist(err) {
			t.Errorf("Expected error log about directory creation, got: %s", logBuffer.String())
		}
	}
}

// Helper to generate a unique temp directory suffix.
func tempDirSuffix() string {
	return filepath.Base(os.TempDir()) + "-" + filepath.Base(filepath.Join("validator", "test"))
}

// TestNoLogPathSet tests that no logging occurs when BlockLogPath is empty.
func TestNoLogPathSet(t *testing.T) {
	// Setup test config with no log path
	cfg := &config.ShellCommandConfig{
		AllowedDirectories:  []string{"/home", "/tmp"},
		AllowCommands:       []config.AllowCommand{{Command: "ls"}},
		DenyCommands:        []config.DenyCommand{{Command: "rm"}},
		DefaultErrorMessage: "Command not allowed",
		BlockLogPath:        "", // Empty log path
	}

	// Create a logger with a buffer to capture logs
	var logBuffer bytes.Buffer
	log := logger.NewWithWriter(&logBuffer)

	// Create the validator
	v := New(cfg, log)

	// Test blocked command
	v.ValidateCommand("rm", []string{"-rf", "/tmp"})

	// Verify no errors about log file creation were logged
	if strings.Contains(logBuffer.String(), "Failed to create directory for block log") {
		t.Errorf("Unexpected log message about log directory: %s", logBuffer.String())
	}

	if strings.Contains(logBuffer.String(), "Failed to open block log file") {
		t.Errorf("Unexpected log message about log file: %s", logBuffer.String())
	}

	if strings.Contains(logBuffer.String(), "Failed to write to block log file") {
		t.Errorf("Unexpected log message about writing to log: %s", logBuffer.String())
	}
}
