package validator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/pkg/logger"
)

// TestValidateSedCommand tests sed command validation through ValidateCommand.
func TestValidateSedCommand(t *testing.T) {
	v, _ := createSedTestValidator(t)

	testAllowedSedCommands(t, v)
	testDisallowedSedCommands(t, v)
}

// createSedTestValidator creates a validator for testing sed commands.
func createSedTestValidator(t *testing.T) (*CommandValidator, *bytes.Buffer) {
	tempHomeDir := t.TempDir()
	tempWorkDir := t.TempDir()

	cfg := &config.ShellCommandConfig{
		AllowedDirectories: []string{tempHomeDir, tempWorkDir},
		AllowCommands: []config.AllowCommand{
			{Command: "ls"},
			{Command: "sed"},
			{Command: "gsed"},
			{Command: "echo"},
		},
		DenyCommands: []config.DenyCommand{
			{Command: "rm", Message: "Remove command is not allowed"},
		},
		DefaultErrorMessage: "Command not allowed by security policy",
	}

	var logBuffer bytes.Buffer
	log := logger.NewWithWriter(&logBuffer)
	v := New(cfg, log)

	return v, &logBuffer
}

// testAllowedSedCommands tests sed commands that should be allowed.
func testAllowedSedCommands(t *testing.T, v *CommandValidator) {
	tests := []struct {
		name    string
		cmd     string
		args    []string
		allowed bool
		message string
	}{
		{
			name:    "SimpleSubstitution",
			cmd:     "sed",
			args:    []string{"s/foo/bar/"},
			allowed: true,
			message: "",
		},
		{
			name:    "GlobalSubstitution",
			cmd:     "sed",
			args:    []string{"s/foo/bar/g"},
			allowed: true,
			message: "",
		},
		{
			name:    "DeleteLine",
			cmd:     "sed",
			args:    []string{"/pattern/d"},
			allowed: true,
			message: "",
		},
		{
			name:    "PrintLine",
			cmd:     "sed",
			args:    []string{"-n", "/pattern/p"},
			allowed: true,
			message: "",
		},
		{
			name:    "InPlaceEdit",
			cmd:     "sed",
			args:    []string{"-i", "s/old/new/g"},
			allowed: true,
			message: "",
		},
		{
			name:    "MultipleExpressions",
			cmd:     "sed",
			args:    []string{"-e", "s/foo/bar/g", "-e", "s/baz/qux/g"},
			allowed: true,
			message: "",
		},
		{
			name:    "GsedSimpleSubstitution",
			cmd:     "gsed",
			args:    []string{"s/foo/bar/g"},
			allowed: true,
			message: "",
		},
	}

	runSedValidationTests(t, v, tests)
}

// testDisallowedSedCommands tests sed commands that should be blocked.
func testDisallowedSedCommands(t *testing.T, v *CommandValidator) {
	tests := []struct {
		name    string
		cmd     string
		args    []string
		allowed bool
		message string
	}{
		{
			name:    "SubstitutionWithEFlag",
			cmd:     "sed",
			args:    []string{"s/pattern/replacement/e"},
			allowed: false,
			message: "sed command blocked: sed script contains dangerous 'e' command that executes shell commands",
		},
		{
			name:    "SubstitutionWithGEFlags",
			cmd:     "sed",
			args:    []string{"s/pattern/replacement/ge"},
			allowed: false,
			message: "sed command blocked: sed script contains dangerous 'e' command that executes shell commands",
		},
		{
			name:    "StandaloneECommand",
			cmd:     "sed",
			args:    []string{"e"},
			allowed: false,
			message: "sed command blocked: sed script contains dangerous 'e' command that executes shell commands",
		},
		{
			name:    "ECommandWithShellCommand",
			cmd:     "sed",
			args:    []string{"e date"},
			allowed: false,
			message: "sed command blocked: sed script contains dangerous 'e' command that executes shell commands",
		},
		{
			name:    "ECommandAfterSemicolon",
			cmd:     "sed",
			args:    []string{"s/foo/bar/;e"},
			allowed: false,
			message: "sed command blocked: sed script contains dangerous 'e' command that executes shell commands",
		},
		{
			name:    "ExpressionWithEFlag",
			cmd:     "sed",
			args:    []string{"-e", "s/foo/bar/e"},
			allowed: false,
			message: "sed command blocked: sed expression contains dangerous 'e' command that executes shell commands",
		},
		{
			name:    "ExpressionWithECommand",
			cmd:     "sed",
			args:    []string{"-e", "e date"},
			allowed: false,
			message: "sed command blocked: sed expression contains dangerous 'e' command that executes shell commands",
		},
		{
			name:    "GsedWithEFlag",
			cmd:     "gsed",
			args:    []string{"s/pattern/replacement/e"},
			allowed: false,
			message: "gsed command blocked: sed script contains dangerous 'e' command that executes shell commands",
		},
		{
			name:    "AlternateSeparatorWithEFlag",
			cmd:     "sed",
			args:    []string{"s|pattern|replacement|e"},
			allowed: false,
			message: "sed command blocked: sed script contains dangerous 'e' command that executes shell commands",
		},
	}

	runSedValidationTests(t, v, tests)
}

// runSedValidationTests runs sed validation tests.
func runSedValidationTests(t *testing.T, v *CommandValidator, tests []struct {
	name    string
	cmd     string
	args    []string
	allowed bool
	message string
},
) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get working directory: %v", err)
			}

			gotAllowed, gotMessage := v.ValidateCommand(tt.cmd, tt.args, wd)
			if gotAllowed != tt.allowed {
				t.Errorf("ValidateCommand() allowed = %v, want %v", gotAllowed, tt.allowed)
			}
			if gotMessage != tt.message {
				t.Errorf("ValidateCommand() message = %q, want %q", gotMessage, tt.message)
			}
		})
	}
}

// TestSedScriptFileValidation tests that sed -f validates the content of script files.
func TestSedScriptFileValidation(t *testing.T) {
	v, _ := createSedTestValidator(t)
	tempDir := t.TempDir()

	// Create a malicious sed script file with 'e' command
	maliciousFile := filepath.Join(tempDir, "evil.sed")
	if err := os.WriteFile(maliciousFile, []byte("e date\n"), 0o644); err != nil {
		t.Fatalf("Failed to create malicious sed script: %v", err)
	}

	// Create a malicious sed script with s///e flag
	maliciousSubstFile := filepath.Join(tempDir, "evil_subst.sed")
	if err := os.WriteFile(maliciousSubstFile, []byte("s/foo/bar/e\n"), 0o644); err != nil {
		t.Fatalf("Failed to create malicious sed script: %v", err)
	}

	// Create a safe sed script file
	safeFile := filepath.Join(tempDir, "safe.sed")
	if err := os.WriteFile(safeFile, []byte("s/foo/bar/g\n"), 0o644); err != nil {
		t.Fatalf("Failed to create safe sed script: %v", err)
	}

	tests := []struct {
		name    string
		cmd     string
		args    []string
		allowed bool
		message string
	}{
		{
			name:    "ScriptFileWithECommand",
			cmd:     "sed",
			args:    []string{"-f", maliciousFile},
			allowed: false,
			message: "sed command blocked: sed script file contains dangerous 'e' command that executes shell commands",
		},
		{
			name:    "ScriptFileWithSubstitutionEFlag",
			cmd:     "sed",
			args:    []string{"-f", maliciousSubstFile},
			allowed: false,
			message: "sed command blocked: sed script file contains dangerous 'e' command that executes shell commands",
		},
		{
			name:    "SafeScriptFile",
			cmd:     "sed",
			args:    []string{"-f", safeFile},
			allowed: true,
			message: "",
		},
		{
			name:    "NonExistentScriptFile",
			cmd:     "sed",
			args:    []string{"-f", filepath.Join(tempDir, "nonexistent.sed")},
			allowed: false,
			message: fmt.Sprintf("sed command blocked: cannot read sed script file %q for validation", filepath.Join(tempDir, "nonexistent.sed")),
		},
		{
			name:    "LongFormFileFlag",
			cmd:     "sed",
			args:    []string{"--file", maliciousFile},
			allowed: false,
			message: "sed command blocked: sed script file contains dangerous 'e' command that executes shell commands",
		},
		{
			name:    "GsedScriptFileWithECommand",
			cmd:     "gsed",
			args:    []string{"-f", maliciousFile},
			allowed: false,
			message: "gsed command blocked: sed script file contains dangerous 'e' command that executes shell commands",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get working directory: %v", err)
			}

			gotAllowed, gotMessage := v.ValidateCommand(tt.cmd, tt.args, wd)
			if gotAllowed != tt.allowed {
				t.Errorf("ValidateCommand() allowed = %v, want %v", gotAllowed, tt.allowed)
			}
			if gotMessage != tt.message {
				t.Errorf("ValidateCommand() message = %q, want %q", gotMessage, tt.message)
			}
		})
	}
}

// TestSedWhenNotAllowed tests sed validation when sed is not in the allowed commands list.
func TestSedWhenNotAllowed(t *testing.T) {
	cfg := &config.ShellCommandConfig{
		AllowedDirectories: []string{"/tmp"},
		AllowCommands: []config.AllowCommand{
			{Command: "ls"},
			{Command: "grep"},
		},
		DefaultErrorMessage: "Command not allowed by security policy",
	}

	var logBuffer bytes.Buffer
	log := logger.NewWithWriter(&logBuffer)
	v := New(cfg, log)

	wd, _ := os.Getwd()
	allowed, message := v.ValidateCommand("sed", []string{"s/foo/bar/"}, wd)

	expectedMsg := "command \"sed\" is not permitted: Command not allowed by security policy"
	if allowed || message != expectedMsg {
		t.Errorf("Expected sed to be disallowed with message %q, got allowed=%v with message %q",
			expectedMsg, allowed, message)
	}
}

// TestSedWhenExplicitlyDenied tests sed validation when sed is explicitly denied.
func TestSedWhenExplicitlyDenied(t *testing.T) {
	cfg := &config.ShellCommandConfig{
		AllowedDirectories: []string{"/tmp"},
		AllowCommands: []config.AllowCommand{
			{Command: "ls"},
		},
		DenyCommands: []config.DenyCommand{
			{Command: "sed", Message: "sed is explicitly denied"},
		},
		DefaultErrorMessage: "Command not allowed by security policy",
	}

	var logBuffer bytes.Buffer
	log := logger.NewWithWriter(&logBuffer)
	v := New(cfg, log)

	wd, _ := os.Getwd()
	allowed, message := v.ValidateCommand("sed", []string{"s/foo/bar/"}, wd)

	expectedMsg := "command \"sed\" is denied: sed is explicitly denied"
	if allowed || message != expectedMsg {
		t.Errorf("Expected sed to be explicitly denied with message %q, got allowed=%v with message %q",
			expectedMsg, allowed, message)
	}
}
