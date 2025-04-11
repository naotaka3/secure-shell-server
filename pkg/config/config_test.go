package config

import (
	"testing"
)

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()

	// Verify default allowed commands
	expectedCommands := map[string]bool{
		"ls":   true,
		"echo": true,
		"cat":  true,
	}

	for cmd, expected := range expectedCommands {
		if cfg.IsCommandAllowed(cmd) != expected {
			t.Errorf("IsCommandAllowed(%q) = %v, want %v", cmd, cfg.IsCommandAllowed(cmd), expected)
		}
	}

	// Verify disallowed commands
	disallowedCommands := []string{"rm", "mv", "cp"}
	for _, cmd := range disallowedCommands {
		if cfg.IsCommandAllowed(cmd) {
			t.Errorf("IsCommandAllowed(%q) = true, want false", cmd)
		}
	}

	// Verify default environment variables
	if cfg.RestrictedEnv["PATH"] != "/usr/bin:/bin" {
		t.Errorf("RestrictedEnv[PATH] = %q, want %q", cfg.RestrictedEnv["PATH"], "/usr/bin:/bin")
	}

	// Verify default execution time
	if cfg.MaxExecutionTime != 30 {
		t.Errorf("MaxExecutionTime = %d, want %d", cfg.MaxExecutionTime, 30)
	}
}

func TestShellConfig_AddAllowedCommand(t *testing.T) {
	cfg := NewDefaultConfig()

	// Add a new command
	cfg.AddAllowedCommand("grep")

	// Verify the command was added
	if !cfg.IsCommandAllowed("grep") {
		t.Errorf("IsCommandAllowed(\"grep\") = false, want true")
	}
}

func TestShellConfig_RemoveAllowedCommand(t *testing.T) {
	cfg := NewDefaultConfig()

	// Verify a command is allowed
	if !cfg.IsCommandAllowed("ls") {
		t.Errorf("IsCommandAllowed(\"ls\") = false, want true")
	}

	// Remove the command
	cfg.RemoveAllowedCommand("ls")

	// Verify the command was removed
	if cfg.IsCommandAllowed("ls") {
		t.Errorf("IsCommandAllowed(\"ls\") = true, want false")
	}
}

func TestShellConfig_SetEnv(t *testing.T) {
	cfg := NewDefaultConfig()

	// Set a new environment variable
	cfg.SetEnv("HOME", "/home/user")

	// Verify the environment variable was set
	if cfg.RestrictedEnv["HOME"] != "/home/user" {
		t.Errorf("RestrictedEnv[HOME] = %q, want %q", cfg.RestrictedEnv["HOME"], "/home/user")
	}

	// Update an existing environment variable
	cfg.SetEnv("PATH", "/usr/local/bin:/usr/bin:/bin")

	// Verify the environment variable was updated
	if cfg.RestrictedEnv["PATH"] != "/usr/local/bin:/usr/bin:/bin" {
		t.Errorf("RestrictedEnv[PATH] = %q, want %q", cfg.RestrictedEnv["PATH"], "/usr/local/bin:/usr/bin:/bin")
	}
}

func TestShellConfig_IsCommandAllowed(t *testing.T) {
	cfg := NewDefaultConfig()

	tests := []struct {
		name    string
		cmd     string
		allowed bool
	}{
		{"allowed - ls", "ls", true},
		{"allowed - echo", "echo", true},
		{"allowed - cat", "cat", true},
		{"disallowed - rm", "rm", false},
		{"disallowed - empty", "", false},
		{"disallowed - nonexistent", "nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cfg.IsCommandAllowed(tt.cmd); got != tt.allowed {
				t.Errorf("IsCommandAllowed() = %v, want %v", got, tt.allowed)
			}
		})
	}
}
