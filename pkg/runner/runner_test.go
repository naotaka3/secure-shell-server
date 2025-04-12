package runner

import (
	"bytes"
	"testing"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/pkg/logger"
	"github.com/shimizu1995/secure-shell-server/pkg/validator"
)

func TestSafeRunner_RunCommand(t *testing.T) {
	cfg := config.NewDefaultConfig()
	log := logger.New()
	validatorObj := validator.New(cfg, log)
	safeRunner := New(cfg, validatorObj, log)

	// Use bytes.Buffer for capturing output
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	safeRunner.SetOutputs(stdout, stderr)

	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{
			name:    "allowed command echo",
			command: "echo hello",
			wantErr: false,
		},
		{
			name:    "allowed command ls",
			command: "ls -l",
			wantErr: false,
		},
		{
			name:    "disallowed command",
			command: "rm -rf /",
			wantErr: true,
		},
		{
			name:    "empty command",
			command: "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the output buffers
			stdout.Reset()
			stderr.Reset()

			err := safeRunner.RunCommand(t.Context(), tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For successful commands, verify that we got some output
			if !tt.wantErr && tt.name == "allowed command echo" {
				if stdout.String() == "" {
					t.Errorf("Expected output for command %q, got empty output", tt.command)
				}
			}
		})
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		wantCmd  string
		wantArgs []string
	}{
		{
			name:     "simple command",
			command:  "ls",
			wantCmd:  "ls",
			wantArgs: nil,
		},
		{
			name:     "command with args",
			command:  "ls -la /tmp",
			wantCmd:  "ls",
			wantArgs: []string{"-la", "/tmp"},
		},
		{
			name:     "command with multiple spaces",
			command:  "echo   hello   world",
			wantCmd:  "echo",
			wantArgs: []string{"hello", "world"},
		},
		{
			name:     "empty command",
			command:  "",
			wantCmd:  "",
			wantArgs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args := parseCommand(tt.command)
			if cmd != tt.wantCmd {
				t.Errorf("parseCommand() cmd = %v, want %v", cmd, tt.wantCmd)
			}

			// Check args
			if len(args) != len(tt.wantArgs) {
				t.Errorf("parseCommand() args length = %v, want %v", len(args), len(tt.wantArgs))
				return
			}

			for i, arg := range args {
				if arg != tt.wantArgs[i] {
					t.Errorf("parseCommand() arg[%d] = %v, want %v", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}
