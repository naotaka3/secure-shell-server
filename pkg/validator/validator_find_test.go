package validator

import (
	"bytes"
	"os"
	"testing"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/pkg/logger"
)

// TestValidateFindCommand tests the find command validation with -exec clause.
func TestValidateFindCommand(t *testing.T) {
	// Create test validator
	v, _ := createFindTestValidator(t)

	// Run the allowed commands tests
	testAllowedFindExecCommands(t, v)

	// Run the disallowed commands tests
	testDisallowedFindExecCommands(t, v)

	// Run the error cases tests
	testFindExecErrorCases(t, v)
}

// createFindTestValidator creates a validator for testing find commands.
func createFindTestValidator(t *testing.T) (*CommandValidator, *bytes.Buffer) {
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
			{Command: "chmod"},
			{Command: "cp"},
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

// testAllowedFindExecCommands tests cases where find executes allowed commands.
func testAllowedFindExecCommands(t *testing.T, v *CommandValidator) {
	tests := []struct {
		name    string
		args    []string
		allowed bool
		message string
	}{
		{
			name:    "FindWithAllowedExec",
			args:    []string{"-type", "f", "-name", "*.txt", "-exec", "echo", "{}", "\\;"},
			allowed: true,
			message: "",
		},
		{
			name:    "FindWithMultipleAllowedExecs",
			args:    []string{"-type", "f", "-exec", "grep", "pattern", "{}", "\\;", "-exec", "cat", "{}", "\\;"},
			allowed: true,
			message: "",
		},
		{
			name:    "FindWithPlus",
			args:    []string{"-name", "*.txt", "-exec", "echo", "{}", "+"},
			allowed: true,
			message: "",
		},
		{
			name:    "FindWithExecdir",
			args:    []string{"-type", "f", "-execdir", "ls", "-la", "{}", "\\;"},
			allowed: true,
			message: "",
		},
	}

	runFindValidationTests(t, v, tests)
}

// testDisallowedFindExecCommands tests cases where find executes disallowed commands.
func testDisallowedFindExecCommands(t *testing.T, v *CommandValidator) {
	tests := []struct {
		name    string
		args    []string
		allowed bool
		message string
	}{
		{
			name:    "FindWithDisallowedExec",
			args:    []string{"-type", "f", "-exec", "rm", "-f", "{}", "\\;"},
			allowed: false,
			message: "find command contains disallowed -exec: command \"rm\" is denied: Remove command is not allowed",
		},
		{
			name:    "FindWithDisallowedAndAllowedExecs",
			args:    []string{"-type", "f", "-exec", "echo", "{}", "\\;", "-exec", "sudo", "chmod", "777", "{}", "\\;"},
			allowed: false,
			message: "find command contains disallowed -exec: command \"sudo\" is denied: Sudo is not allowed for security reasons",
		},
		{
			name:    "FindWithUnlistedExecCommand",
			args:    []string{"-type", "f", "-exec", "wget", "https://example.com", "\\;"},
			allowed: false,
			message: "find command contains disallowed -exec: command \"wget\" is not permitted: Command not allowed by security policy",
		},
	}

	runFindValidationTests(t, v, tests)
}

// testFindExecErrorCases tests error cases for find command validation.
func testFindExecErrorCases(t *testing.T, v *CommandValidator) {
	tests := []struct {
		name    string
		args    []string
		allowed bool
		message string
	}{
		{
			name:    "FindWithNoArgs",
			args:    []string{},
			allowed: true, // find with no args is actually valid, it lists current directory
			message: "",
		},
		{
			name:    "FindWithNoExec",
			args:    []string{"-type", "f", "-name", "*.txt"},
			allowed: true, // find without -exec is allowed
			message: "",
		},
		{
			name:    "FindWithIncompleteExec",
			args:    []string{"-type", "f", "-exec"},
			allowed: true, // Parsing will just not find any commands, validator doesn't check command format
			message: "",
		},
	}

	runFindValidationTests(t, v, tests)
}

// runFindValidationTests runs the validation tests and checks the results.
func runFindValidationTests(t *testing.T, v *CommandValidator, tests []struct {
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

			gotAllowed, gotMessage := v.ValidateCommand("find", tt.args, wd)
			if gotAllowed != tt.allowed {
				t.Errorf("ValidateCommand() allowed = %v, want %v", gotAllowed, tt.allowed)
			}
			if gotMessage != tt.message {
				t.Errorf("ValidateCommand() message = %q, want %q", gotMessage, tt.message)
			}
		})
	}
}

// TestFindWhenNotAllowed tests find validation when find is not in the allowed commands list.
func TestFindWhenNotAllowed(t *testing.T) {
	// Create a configuration that doesn't include find in allowed commands
	cfg := &config.ShellCommandConfig{
		AllowedDirectories: []string{"/tmp"},
		AllowCommands: []config.AllowCommand{
			{Command: "ls"},
			{Command: "grep"},
			// find not included
		},
		DefaultErrorMessage: "Command not allowed by security policy",
	}

	// Create a logger with a buffer
	var logBuffer bytes.Buffer
	log := logger.NewWithWriter(&logBuffer)

	// Create the validator
	v := New(cfg, log)

	// Test find command
	wd, _ := os.Getwd()
	allowed, message := v.ValidateCommand("find", []string{"-type", "f", "-exec", "echo", "{}", "\\;"}, wd)

	// Verify find is not allowed
	expectedMsg := "command \"find\" is not permitted: Command not allowed by security policy"
	if allowed || message != expectedMsg {
		t.Errorf("Expected find to be disallowed with message %q, got allowed=%v with message %q",
			expectedMsg, allowed, message)
	}
}

// TestFindWhenExplicitlyDenied tests find validation when find is explicitly denied.
func TestFindWhenExplicitlyDenied(t *testing.T) {
	// Create a configuration that explicitly denies find
	cfg := &config.ShellCommandConfig{
		AllowedDirectories: []string{"/tmp"},
		AllowCommands: []config.AllowCommand{
			{Command: "ls"},
			{Command: "grep"},
		},
		DenyCommands: []config.DenyCommand{
			{Command: "find", Message: "find is explicitly denied"},
		},
		DefaultErrorMessage: "Command not allowed by security policy",
	}

	// Create a logger with a buffer
	var logBuffer bytes.Buffer
	log := logger.NewWithWriter(&logBuffer)

	// Create the validator
	v := New(cfg, log)

	// Test find command
	wd, _ := os.Getwd()
	allowed, message := v.ValidateCommand("find", []string{"-type", "f", "-exec", "echo", "{}", "\\;"}, wd)

	// Verify find is explicitly denied
	expectedMsg := "command \"find\" is denied: find is explicitly denied"
	if allowed || message != expectedMsg {
		t.Errorf("Expected find to be explicitly denied with message %q, got allowed=%v with message %q",
			expectedMsg, allowed, message)
	}
}
