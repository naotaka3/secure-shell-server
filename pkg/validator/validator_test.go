package validator

import (
	"strings"
	"testing"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/pkg/logger"
)

func TestValidator_ValidateScript(t *testing.T) {
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
		script  string
		wantOk  bool
		wantErr bool
	}{
		{
			name:    "allowed command",
			script:  "ls -l",
			wantOk:  true,
			wantErr: false,
		},
		{
			name:    "disallowed command",
			script:  "rm -rf /",
			wantOk:  false,
			wantErr: true,
		},
		{
			name:    "multiple allowed commands",
			script:  "ls -l; echo hello",
			wantOk:  true,
			wantErr: false,
		},
		{
			name:    "mixed allowed/disallowed commands",
			script:  "ls -l; rm -rf /",
			wantOk:  false,
			wantErr: true,
		},
		{
			name:    "syntax error",
			script:  "ls -l; echo 'unclosed string",
			wantOk:  false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOk, err := validator.ValidateScript(tt.script)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOk != tt.wantOk {
				t.Errorf("ValidateScript() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestValidator_ValidateScriptWithSubcommands(t *testing.T) {
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
		script      string
		wantOk      bool
		wantErr     bool
		errContains string
	}{
		{
			name:        "allowed simple command",
			script:      "ls -l",
			wantOk:      true,
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "allowed command with allowed subcommand",
			script:      "git status",
			wantOk:      true,
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "allowed command with disallowed subcommand",
			script:      "git push",
			wantOk:      false,
			wantErr:     true,
			errContains: "subcommand \"push\" is not allowed",
		},
		{
			name:        "allowed command with denied subcommand",
			script:      "cp -r /src /dest",
			wantOk:      false,
			wantErr:     true,
			errContains: "subcommand \"-r\" is denied",
		},
		{
			name:        "explicitly denied command with message",
			script:      "rm -rf /",
			wantOk:      false,
			wantErr:     true,
			errContains: "Remove command is not allowed",
		},
		{
			name:        "explicitly denied command without message",
			script:      "mv /src /dest",
			wantOk:      false,
			wantErr:     true,
			errContains: "Command not allowed by security policy",
		},
		{
			name:        "command not in allow list",
			script:      "wget http://example.com",
			wantOk:      false,
			wantErr:     true,
			errContains: "Command not allowed by security policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOk, err := validator.ValidateScript(tt.script)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOk != tt.wantOk {
				t.Errorf("ValidateScript() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if tt.errContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateScript() error = %v, should contain %v", err, tt.errContains)
				}
			}
		})
	}
}
