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
			{Command: "cat"},
			{Command: "echo"},
			{Command: "grep"},
			{Command: "find"},
			{Command: "git", SubCommands: []string{"status", "log", "diff"}, DenySubCommands: []string{"push", "commit"}},
			{Command: "docker", DenySubCommands: []string{"rm", "exec", "run"}},                              // Command with denied subcommands
			{Command: "npm", SubCommands: []string{"install", "update"}, DenySubCommands: []string{"audit"}}, // Command with both allowed and denied subcommands
		},
		DenyCommands: []config.DenyCommand{
			{Command: "rm", Message: "Remove command is not allowed"},
			{Command: "sudo", Message: "Sudo is not allowed for security reasons"}, // With custom error message
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
		// Test additional allowed commands
		{name: "LsCommand", cmd: "ls", args: []string{"-la"}, allowed: true, message: ""},
		{name: "EchoCommand", cmd: "echo", args: []string{"hello"}, allowed: true, message: ""},
		{name: "CatCommand", cmd: "cat", args: []string{"/tmp/file.txt"}, allowed: true, message: ""},
		{name: "GrepCommand", cmd: "grep", args: []string{"pattern", "file.txt"}, allowed: true, message: ""},
		{name: "FindCommand", cmd: "find", args: []string{".", "-name", "*.txt"}, allowed: true, message: ""},

		// Test denied commands
		{name: "ExplicitlyDeniedCommand", cmd: "rm", args: []string{"-rf", "/tmp"}, allowed: false, message: "command \"rm\" is denied: Remove command is not allowed"},
		{name: "DeniedCommandWithCustomMessage", cmd: "sudo", args: []string{"apt-get", "update"}, allowed: false, message: "command \"sudo\" is denied: Sudo is not allowed for security reasons"},
		{name: "UnlistedCommand", cmd: "wget", args: []string{"https://example.com"}, allowed: false, message: "command \"wget\" is not permitted: Command not allowed by security policy"},
		{name: "ChmodNotInAllowList", cmd: "chmod", args: []string{"777", "file.txt"}, allowed: false, message: "command \"chmod\" is not permitted: Command not allowed by security policy"},

		// Test git-specific subcommands
		{name: "GitStatusAllowed", cmd: "git", args: []string{"status"}, allowed: true, message: ""},
		{name: "GitLogAllowed", cmd: "git", args: []string{"log"}, allowed: true, message: ""},
		{name: "GitDiffAllowed", cmd: "git", args: []string{"diff"}, allowed: true, message: ""},
		{name: "GitPushDenied", cmd: "git", args: []string{"push"}, allowed: false, message: "subcommand \"push\" is not allowed for command \"git\""},
		{name: "GitCommitDenied", cmd: "git", args: []string{"commit"}, allowed: false, message: "subcommand \"commit\" is not allowed for command \"git\""},
		{name: "GitCloneNotAllowed", cmd: "git", args: []string{"clone", "https://github.com/example/repo.git"}, allowed: false, message: "subcommand \"clone\" is not allowed for command \"git\""},

		// Test docker subcommand handling
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
