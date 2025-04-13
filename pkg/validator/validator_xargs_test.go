package validator

import (
	"bytes"
	"os"
	"testing"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/pkg/logger"
)

// TestValidateXargsCommand tests the validateXargsCommand function.
func TestValidateXargsCommand(t *testing.T) {
	// Create test validator
	v, _ := createXargsTestValidator(t)

	// Run the allowed commands tests
	testAllowedXargsCommands(t, v)

	// Run the disallowed commands tests
	testDisallowedXargsCommands(t, v)

	// Run the error cases tests
	testXargsErrorCases(t, v)
}

// createXargsTestValidator creates a validator for testing xargs commands.
func createXargsTestValidator(t *testing.T) (*CommandValidator, *bytes.Buffer) {
	// Create temporary directories for testing
	tempHomeDir := t.TempDir()
	tempWorkDir := t.TempDir()

	// Setup test config with temp directories
	cfg := &config.ShellCommandConfig{
		AllowedDirectories: []string{tempHomeDir, tempWorkDir},
		AllowCommands: []config.AllowCommand{
			{Command: "ls"},
			{Command: "grep"},
			{Command: "cat"},
			{Command: "echo"},
			{Command: "find"},
			{Command: "xargs"},
		},
		DenyCommands: []config.DenyCommand{
			{Command: "rm", Message: "Remove command is not allowed"},
			{Command: "sudo", Message: "Sudo is not allowed for security reasons"},
		},
		DefaultErrorMessage: "Command not allowed by security policy",
	}

	// Create a logger with a buffer
	var logBuffer bytes.Buffer
	log := logger.NewWithWriter(&logBuffer)

	// Create the validator
	v := New(cfg, log)

	return v, &logBuffer
}

// testAllowedXargsCommands tests cases where xargs executes allowed commands.
func testAllowedXargsCommands(t *testing.T, v *CommandValidator) {
	tests := []struct {
		name    string
		args    []string
		allowed bool
		message string
	}{
		{
			name:    "AllowedSimpleCommand",
			args:    []string{"echo", "test"},
			allowed: true,
			message: "",
		},
		{
			name:    "AllowedCommandWithArguments",
			args:    []string{"grep", "pattern", "file.txt"},
			allowed: true,
			message: "",
		},
		{
			name:    "AllowedWithFlags",
			args:    []string{"-n", "1", "ls", "-la"},
			allowed: true,
			message: "",
		},
	}

	runXargsValidationTests(t, v, tests)
}

// testDisallowedXargsCommands tests cases where xargs executes disallowed commands.
func testDisallowedXargsCommands(t *testing.T, v *CommandValidator) {
	tests := []struct {
		name    string
		args    []string
		allowed bool
		message string
	}{
		{
			name:    "DisallowedCommand",
			args:    []string{"rm", "-rf", "file.txt"},
			allowed: false,
			message: "xargs would execute disallowed command: command \"rm\" is denied: Remove command is not allowed",
		},
		{
			name:    "DisallowedWithExec",
			args:    []string{"-exec", "sudo", "apt-get", "update"},
			allowed: false,
			message: "xargs would execute disallowed command: command \"sudo\" is denied: Sudo is not allowed for security reasons",
		},
		{
			name:    "UnlistedCommand",
			args:    []string{"wget", "https://example.com"},
			allowed: false,
			message: "xargs would execute disallowed command: command \"wget\" is not permitted: Command not allowed by security policy",
		},
		{
			name:    "PathValidation",
			args:    []string{"cat", "/etc/passwd"},
			allowed: false,
			message: "xargs would execute disallowed command: path \"/etc/passwd\" is outside of allowed directories: Command not allowed by security policy",
		},
	}

	runXargsValidationTests(t, v, tests)
}

// testXargsErrorCases tests error cases when parsing xargs commands.
func testXargsErrorCases(t *testing.T, v *CommandValidator) {
	tests := []struct {
		name    string
		args    []string
		allowed bool
		message string
	}{
		{
			name:    "InvalidXargsCommand",
			args:    []string{"-n", "1", "-L", "1"},
			allowed: false,
			message: "unable to determine command to be executed by xargs",
		},
		{
			name:    "EmptyArguments",
			args:    []string{},
			allowed: false,
			message: "no arguments provided to xargs",
		},
		{
			name:    "MissingCommandAfterExec",
			args:    []string{"-exec"},
			allowed: false,
			message: "unable to determine command to be executed by xargs",
		},
	}

	runXargsValidationTests(t, v, tests)
}

// runXargsValidationTests runs the validation tests and checks the results.
func runXargsValidationTests(t *testing.T, v *CommandValidator, tests []struct {
	name    string
	args    []string
	allowed bool
	message string
},
) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use current working directory for test
			wd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get working directory: %v", err)
			}

			gotAllowed, gotMessage := v.validateXargsCommand(tt.args, wd)
			if gotAllowed != tt.allowed {
				t.Errorf("validateXargsCommand() allowed = %v, want %v", gotAllowed, tt.allowed)
			}
			if gotMessage != tt.message {
				t.Errorf("validateXargsCommand() message = %q, want %q", gotMessage, tt.message)
			}
		})
	}
}

// TestXargsWithValidateCommand tests xargs command validation using ValidateCommand directly.
func TestXargsWithValidateCommand(t *testing.T) {
	// Create temporary directories for testing
	tempHomeDir := t.TempDir()
	tempWorkDir := t.TempDir()

	// Setup test config with temp directories
	cfg := &config.ShellCommandConfig{
		AllowedDirectories: []string{tempHomeDir, tempWorkDir},
		AllowCommands: []config.AllowCommand{
			{Command: "ls"},
			{Command: "grep"},
			{Command: "xargs"},
			{Command: "echo"},
		},
		DenyCommands: []config.DenyCommand{
			{Command: "rm", Message: "Remove command is not allowed"},
		},
		DefaultErrorMessage: "Command not allowed by security policy",
	}

	// Create a logger with a buffer
	var logBuffer bytes.Buffer
	log := logger.NewWithWriter(&logBuffer)

	// Create the validator
	v := New(cfg, log)

	// Test cases
	tests := []struct {
		name    string
		args    []string
		allowed bool
		message string
	}{
		{
			name:    "XargsWithAllowedCommand",
			args:    []string{"echo", "hello"},
			allowed: true,
			message: "",
		},
		{
			name:    "XargsWithDisallowedCommand",
			args:    []string{"rm", "-rf", "file.txt"},
			allowed: false,
			message: "xargs would execute disallowed command: command \"rm\" is denied: Remove command is not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset log buffer for each test
			logBuffer.Reset()

			// Use current working directory for test
			wd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get working directory: %v", err)
			}

			// Call ValidateCommand directly with xargs
			gotAllowed, gotMessage := v.ValidateCommand("xargs", tt.args, wd)
			if gotAllowed != tt.allowed {
				t.Errorf("ValidateCommand() allowed = %v, want %v", gotAllowed, tt.allowed)
			}
			if gotMessage != tt.message {
				t.Errorf("ValidateCommand() message = %q, want %q", gotMessage, tt.message)
			}
		})
	}
}

// TestXargsWhenNotAllowed tests xargs validation when xargs is not in the allowed commands list.
func TestXargsWhenNotAllowed(t *testing.T) {
	// Create a configuration that doesn't include xargs in allowed commands
	cfg := &config.ShellCommandConfig{
		AllowedDirectories: []string{"/tmp"},
		AllowCommands: []config.AllowCommand{
			{Command: "ls"},
			{Command: "grep"},
			// xargs not included
		},
		DefaultErrorMessage: "Command not allowed by security policy",
	}

	// Create a logger with a buffer
	var logBuffer bytes.Buffer
	log := logger.NewWithWriter(&logBuffer)

	// Create the validator
	v := New(cfg, log)

	// Test xargs command
	wd, _ := os.Getwd()
	allowed, message := v.ValidateCommand("xargs", []string{"echo", "hello"}, wd)

	// Verify xargs is not allowed
	expectedMsg := "command \"xargs\" is not permitted: Command not allowed by security policy"
	if allowed || message != expectedMsg {
		t.Errorf("Expected xargs to be disallowed with message %q, got allowed=%v with message %q",
			expectedMsg, allowed, message)
	}
}

// TestXargsWhenExplicitlyDenied tests xargs validation when xargs is explicitly denied.
func TestXargsWhenExplicitlyDenied(t *testing.T) {
	// Create a configuration that explicitly denies xargs
	cfg := &config.ShellCommandConfig{
		AllowedDirectories: []string{"/tmp"},
		AllowCommands: []config.AllowCommand{
			{Command: "ls"},
			{Command: "grep"},
		},
		DenyCommands: []config.DenyCommand{
			{Command: "xargs", Message: "xargs is explicitly denied"},
		},
		DefaultErrorMessage: "Command not allowed by security policy",
	}

	// Create a logger with a buffer
	var logBuffer bytes.Buffer
	log := logger.NewWithWriter(&logBuffer)

	// Create the validator
	v := New(cfg, log)

	// Test xargs command
	wd, _ := os.Getwd()
	allowed, message := v.ValidateCommand("xargs", []string{"echo", "hello"}, wd)

	// Verify xargs is explicitly denied
	expectedMsg := "command \"xargs\" is denied: xargs is explicitly denied"
	if allowed || message != expectedMsg {
		t.Errorf("Expected xargs to be explicitly denied with message %q, got allowed=%v with message %q",
			expectedMsg, allowed, message)
	}
}
