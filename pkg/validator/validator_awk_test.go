package validator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/pkg/logger"
)

// TestValidateAwkCommand tests awk command validation through ValidateCommand.
func TestValidateAwkCommand(t *testing.T) {
	v := createAwkTestValidator(t)

	testAllowedAwkCommands(t, v)
	testDisallowedAwkCommands(t, v)
}

// createAwkTestValidator creates a validator for testing awk commands.
func createAwkTestValidator(t *testing.T) *CommandValidator {
	tempHomeDir := t.TempDir()
	tempWorkDir := t.TempDir()

	cfg := &config.ShellCommandConfig{
		AllowedDirectories: []string{tempHomeDir, tempWorkDir},
		AllowCommands: []config.AllowCommand{
			{Command: "ls"},
			{Command: "awk"},
			{Command: "gawk"},
			{Command: "echo"},
		},
		DenyCommands: []config.DenyCommand{
			{Command: "rm", Message: "Remove command is not allowed"},
		},
		DefaultErrorMessage: "Command not allowed by security policy",
	}

	log := logger.NewWithWriter(io.Discard)
	v := New(cfg, log)

	return v
}

// testAllowedAwkCommands tests awk commands that should be allowed.
func testAllowedAwkCommands(t *testing.T, v *CommandValidator) {
	tests := []struct {
		name    string
		cmd     string
		args    []string
		allowed bool
		message string
	}{
		{
			name:    "SimplePrint",
			cmd:     "awk",
			args:    []string{"{print $1}"},
			allowed: true,
			message: "",
		},
		{
			name:    "FieldSeparator",
			cmd:     "awk",
			args:    []string{"-F", ":", "{print $1}"},
			allowed: true,
			message: "",
		},
		{
			name:    "PatternMatch",
			cmd:     "awk",
			args:    []string{"/error/{print $0}"},
			allowed: true,
			message: "",
		},
		{
			name:    "GawkSimplePrint",
			cmd:     "gawk",
			args:    []string{"{print NR, $0}"},
			allowed: true,
			message: "",
		},
		{
			name:    "BeginEndBlock",
			cmd:     "awk",
			args:    []string{"BEGIN{sum=0} {sum+=$1} END{print sum}"},
			allowed: true,
			message: "",
		},
	}

	runAwkValidationTests(t, v, tests)
}

// testDisallowedAwkCommands tests awk commands that should be blocked.
func testDisallowedAwkCommands(t *testing.T, v *CommandValidator) {
	tests := []struct {
		name    string
		cmd     string
		args    []string
		allowed bool
		message string
	}{
		{
			name:    "SystemCall",
			cmd:     "awk",
			args:    []string{`BEGIN { system("rm -rf /") }`},
			allowed: false,
			message: "awk command blocked: awk script contains dangerous command execution pattern",
		},
		{
			name:    "PipeToCommand",
			cmd:     "awk",
			args:    []string{`{print $0 | "sh"}`},
			allowed: false,
			message: "awk command blocked: awk script contains dangerous command execution pattern",
		},
		{
			name:    "GetlineFromPipe",
			cmd:     "awk",
			args:    []string{`{"date" | getline d; print d}`},
			allowed: false,
			message: "awk command blocked: awk script contains dangerous command execution pattern",
		},
		{
			name:    "TwoWayPipe",
			cmd:     "awk",
			args:    []string{`BEGIN { print "hello" |& "/bin/sh" }`},
			allowed: false,
			message: "awk command blocked: awk script contains dangerous command execution pattern",
		},
		{
			name:    "GawkSystemCall",
			cmd:     "gawk",
			args:    []string{`{system("id")}`},
			allowed: false,
			message: "gawk command blocked: awk script contains dangerous command execution pattern",
		},
		{
			name:    "SystemViaSourceFlag",
			cmd:     "awk",
			args:    []string{"-e", `{system("id")}`},
			allowed: false,
			message: "awk command blocked: awk script contains dangerous command execution pattern",
		},
		{
			name:    "SystemViaLongSourceFlag",
			cmd:     "awk",
			args:    []string{"--source", `BEGIN{system("whoami")}`},
			allowed: false,
			message: "awk command blocked: awk script contains dangerous command execution pattern",
		},
	}

	runAwkValidationTests(t, v, tests)
}

// runAwkValidationTests runs awk validation tests.
func runAwkValidationTests(t *testing.T, v *CommandValidator, tests []struct {
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

//nolint:dupl // test data for awk script file validation; structurally similar to sed but different commands and messages
func TestAwkScriptFileValidation(t *testing.T) {
	v := createAwkTestValidator(t)
	tempDir := t.TempDir()

	paths := createScriptFiles(t, tempDir, []scriptFile{
		{name: "evil_system.awk", content: "BEGIN { system(\"id\") }\n"},
		{name: "evil_getline.awk", content: "{\"date\" | getline d; print d}\n"},
		{name: "safe.awk", content: "{print $1}\n"},
	})

	nonExistent := filepath.Join(tempDir, "nonexistent.awk")

	runValidationTestCases(t, v, []validationTestCase{
		{
			name:    "ScriptFileWithSystemCall",
			cmd:     "awk",
			args:    []string{"-f", paths["evil_system.awk"]},
			allowed: false,
			message: "awk command blocked: awk script file contains dangerous command execution pattern",
		},
		{
			name:    "ScriptFileWithPipeGetline",
			cmd:     "awk",
			args:    []string{"-f", paths["evil_getline.awk"]},
			allowed: false,
			message: "awk command blocked: awk script file contains dangerous command execution pattern",
		},
		{
			name:    "SafeScriptFile",
			cmd:     "awk",
			args:    []string{"-f", paths["safe.awk"]},
			allowed: true,
			message: "",
		},
		{
			name:    "NonExistentScriptFile",
			cmd:     "awk",
			args:    []string{"-f", nonExistent},
			allowed: false,
			message: fmt.Sprintf("awk command blocked: cannot read awk script file %q for validation", nonExistent),
		},
		{
			name:    "LongFormFileFlag",
			cmd:     "awk",
			args:    []string{"--file", paths["evil_system.awk"]},
			allowed: false,
			message: "awk command blocked: awk script file contains dangerous command execution pattern",
		},
		{
			name:    "GawkScriptFileWithSystemCall",
			cmd:     "gawk",
			args:    []string{"-f", paths["evil_system.awk"]},
			allowed: false,
			message: "gawk command blocked: awk script file contains dangerous command execution pattern",
		},
	})
}

// TestAwkWhenNotAllowed tests awk validation when awk is not in the allowed commands list.
func TestAwkWhenNotAllowed(t *testing.T) {
	cfg := &config.ShellCommandConfig{
		AllowedDirectories: []string{"/tmp"},
		AllowCommands: []config.AllowCommand{
			{Command: "ls"},
			{Command: "grep"},
		},
		DefaultErrorMessage: "Command not allowed by security policy",
	}

	log := logger.NewWithWriter(io.Discard)
	v := New(cfg, log)

	wd, _ := os.Getwd()
	allowed, message := v.ValidateCommand("awk", []string{"{print $1}"}, wd)

	expectedMsg := "command \"awk\" is not permitted: Command not allowed by security policy"
	if allowed || message != expectedMsg {
		t.Errorf("Expected awk to be disallowed with message %q, got allowed=%v with message %q",
			expectedMsg, allowed, message)
	}
}

// TestAwkWhenExplicitlyDenied tests awk validation when awk is explicitly denied.
func TestAwkWhenExplicitlyDenied(t *testing.T) {
	cfg := &config.ShellCommandConfig{
		AllowedDirectories: []string{"/tmp"},
		AllowCommands: []config.AllowCommand{
			{Command: "ls"},
		},
		DenyCommands: []config.DenyCommand{
			{Command: "awk", Message: "awk is explicitly denied"},
		},
		DefaultErrorMessage: "Command not allowed by security policy",
	}

	log := logger.NewWithWriter(io.Discard)
	v := New(cfg, log)

	wd, _ := os.Getwd()
	allowed, message := v.ValidateCommand("awk", []string{"{print $1}"}, wd)

	expectedMsg := "command \"awk\" is denied: awk is explicitly denied"
	if allowed || message != expectedMsg {
		t.Errorf("Expected awk to be explicitly denied with message %q, got allowed=%v with message %q",
			expectedMsg, allowed, message)
	}
}
