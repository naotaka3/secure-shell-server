package config

import (
	"encoding/json"
	"testing"
)

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()

	// Verify default allowed commands are present
	expectedCommands := []string{"ls", "cat", "echo"}
	for _, cmd := range expectedCommands {
		if !cfg.IsCommandAllowed(cmd) {
			t.Errorf("IsCommandAllowed(%q) = false, want true", cmd)
		}
	}

	// Verify disallowed commands
	disallowedCommands := []string{"wget", "curl", "cp"}
	for _, cmd := range disallowedCommands {
		if cfg.IsCommandAllowed(cmd) {
			t.Errorf("IsCommandAllowed(%q) = true, want false", cmd)
		}
	}

	// Verify default execution time
	if cfg.MaxExecutionTime != 30 {
		t.Errorf("MaxExecutionTime = %d, want %d", cfg.MaxExecutionTime, 30)
	}
}

func TestIsCommandAllowed(t *testing.T) {
	cfg := NewDefaultConfig()

	tests := []struct {
		name    string
		cmd     string
		allowed bool
	}{
		{"allowed - ls", "ls", true},
		{"allowed - echo", "echo", true},
		{"allowed - cat", "cat", true},
		{"denied - rm", "rm", false},
		{"denied - empty", "", false},
		{"denied - nonexistent", "nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cfg.IsCommandAllowed(tt.cmd); got != tt.allowed {
				t.Errorf("IsCommandAllowed() = %v, want %v", got, tt.allowed)
			}
		})
	}
}

func TestMixedCommandFormats(t *testing.T) {
	const configJSON = `{
		"allowedDirectories": ["/home", "/tmp"],
		"allowCommands": [
			"ls",
			{"command": "git", "subCommands": ["status", "pull"], "denySubCommands": ["push"]},
			"cat"
		],
		"denyCommands": [
			"rm",
			{"command": "sudo", "message": "Elevated privileges not allowed"},
			"vi"
		],
		"defaultErrorMessage": "Command not allowed",
		"maxExecutionTime": 60
	}`

	var config ShellCommandConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Verify allow commands count
	if len(config.AllowCommands) != 3 {
		t.Errorf("Expected 3 allow commands, got %d", len(config.AllowCommands))
	}

	// Verify a string-only allow command
	if config.AllowCommands[0].Command != "ls" {
		t.Errorf("Expected first command to be 'ls', got %q", config.AllowCommands[0].Command)
	}

	// Verify a structured allow command
	if config.AllowCommands[1].Command != "git" {
		t.Errorf("Expected second command to be 'git', got %q", config.AllowCommands[1].Command)
	}
	if len(config.AllowCommands[1].SubCommands) != 2 {
		t.Errorf("Expected 2 subcommands for 'git', got %d", len(config.AllowCommands[1].SubCommands))
	}
	if len(config.AllowCommands[1].DenySubCommands) != 1 {
		t.Errorf("Expected 1 deny subcommand for 'git', got %d", len(config.AllowCommands[1].DenySubCommands))
	}

	// Verify deny commands count
	if len(config.DenyCommands) != 3 {
		t.Errorf("Expected 3 deny commands, got %d", len(config.DenyCommands))
	}

	// Verify a string-only deny command
	if config.DenyCommands[0].Command != "rm" {
		t.Errorf("Expected first deny command to be 'rm', got %q", config.DenyCommands[0].Command)
	}
	if config.DenyCommands[0].Message != "" {
		t.Errorf("Expected empty message for 'rm', got %q", config.DenyCommands[0].Message)
	}

	// Verify a structured deny command
	if config.DenyCommands[1].Command != "sudo" {
		t.Errorf("Expected second deny command to be 'sudo', got %q", config.DenyCommands[1].Command)
	}
	if config.DenyCommands[1].Message != "Elevated privileges not allowed" {
		t.Errorf("Expected specific message for 'sudo', got %q", config.DenyCommands[1].Message)
	}
}
