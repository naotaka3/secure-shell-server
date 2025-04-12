package validator

import (
	"strings"
	"testing"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/pkg/logger"
)

func TestValidator_ValidateCommand(t *testing.T) {
	cfg := &config.ShellCommandConfig{
		AllowedDirectories: []string{"/home", "/tmp"},
		AllowCommands: []config.AllowCommand{
			{Command: "ls"},
			{Command: "echo"},
			{Command: "cat"},
		},
		DenyCommands:        []config.DenyCommand{{Command: "rm", Message: "Remove command is not allowed"}},
		DefaultErrorMessage: "Command not allowed by security policy",
	}

	log := logger.New()
	validator := New(cfg, log)

	tests := []struct {
		name    string
		command string
		wantOk  bool
		wantErr bool
	}{
		{
			name:    "allowed command",
			command: "ls -l",
			wantOk:  true,
			wantErr: false,
		},
		{
			name:    "disallowed command",
			command: "rm -rf /",
			wantOk:  false,
			wantErr: true,
		},
		{
			name:    "multiple allowed commands",
			command: "ls -l; echo hello",
			wantOk:  true,
			wantErr: false,
		},
		{
			name:    "mixed allowed/disallowed commands",
			command: "ls -l; rm -rf /",
			wantOk:  false,
			wantErr: true,
		},
		{
			name:    "syntax error",
			command: "ls -l; echo 'unclosed string",
			wantOk:  false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOk, err := validator.ValidateCommand(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOk != tt.wantOk {
				t.Errorf("ValidateCommand() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestValidator_ValidateCommandWithSubcommands(t *testing.T) {
	cfg := &config.ShellCommandConfig{
		AllowedDirectories: []string{"/home", "/tmp"},
		AllowCommands: []config.AllowCommand{
			{Command: "ls"},
			{Command: "git", SubCommands: []string{"status", "pull", "commit"}},
			{Command: "cp", DenySubCommands: []string{"-r", "--recursive"}},
		},
		DenyCommands: []config.DenyCommand{
			{Command: "rm", Message: "Remove command is not allowed"},
			{Command: "mv"},
		},
		DefaultErrorMessage: "Command not allowed by security policy",
	}

	log := logger.New()
	validator := New(cfg, log)

	tests := []struct {
		name        string
		command     string
		wantOk      bool
		wantErr     bool
		errContains string
	}{
		{
			name:        "allowed simple command",
			command:     "ls -l",
			wantOk:      true,
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "allowed command with allowed subcommand",
			command:     "git status",
			wantOk:      true,
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "allowed command with disallowed subcommand",
			command:     "git push",
			wantOk:      false,
			wantErr:     true,
			errContains: "subcommand \"push\" is not allowed",
		},
		{
			name:        "allowed command with denied subcommand",
			command:     "cp -r /src /dest",
			wantOk:      false,
			wantErr:     true,
			errContains: "subcommand \"-r\" is denied",
		},
		{
			name:        "explicitly denied command with message",
			command:     "rm -rf /",
			wantOk:      false,
			wantErr:     true,
			errContains: "Remove command is not allowed",
		},
		{
			name:        "explicitly denied command without message",
			command:     "mv /src /dest",
			wantOk:      false,
			wantErr:     true,
			errContains: "Command not allowed by security policy",
		},
		{
			name:        "command not in allow list",
			command:     "wget http://example.com",
			wantOk:      false,
			wantErr:     true,
			errContains: "Command not allowed by security policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOk, err := validator.ValidateCommand(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOk != tt.wantOk {
				t.Errorf("ValidateCommand() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if tt.errContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateCommand() error = %v, should contain %v", err, tt.errContains)
				}
			}
		})
	}
}
