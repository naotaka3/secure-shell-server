package config

import (
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
