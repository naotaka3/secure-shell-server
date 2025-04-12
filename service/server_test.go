package service_test

import (
	"testing"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/service"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
		wantErr bool
	}{
		{
			name:    "simple command",
			input:   "ls -la",
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "command with multiple args",
			input:   "echo hello world",
			wantLen: 3,
			wantErr: false,
		},
		{
			name:    "empty command",
			input:   "",
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "command with spaces",
			input:   "  ls  -la  ",
			wantLen: 2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := service.TestParseCommand(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(args) != tt.wantLen {
				t.Errorf("parseCommand() returned %d args, want %d, args: %v", len(args), tt.wantLen, args)
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	// Create a test configuration
	cfg := config.NewDefaultConfig()

	// Create a test server
	server := service.NewServer(cfg, 8080)

	// Just a basic test to ensure server creation works
	if server == nil {
		t.Fatal("Failed to create server")
	}
}
