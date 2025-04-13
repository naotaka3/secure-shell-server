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
		name        string
		command     string
		workingDir  string
		allowedDirs []string
		wantErr     bool
	}{
		{
			name:        "allowed command in allowed directory",
			command:     "echo hello\nls -l",
			workingDir:  "/tmp",
			allowedDirs: []string{"/tmp", "/home"},
			wantErr:     false,
		},
		{
			name:        "allowed command in disallowed directory",
			command:     "echo hello\nls -l",
			workingDir:  "/etc",
			allowedDirs: []string{"/tmp", "/home"},
			wantErr:     true,
		},
		{
			name:        "disallowed command",
			command:     "echo hello\nrm -rf /",
			workingDir:  "/tmp",
			allowedDirs: []string{"/tmp"},
			wantErr:     true,
		},
		{
			name:        "syntax error",
			command:     "echo 'unclosed string",
			workingDir:  "/tmp",
			allowedDirs: []string{"/tmp"},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new config and runner for each test case
			testCfg := config.NewDefaultConfig()
			testCfg.AllowedDirectories = tt.allowedDirs
			testLog := logger.New()
			testValidator := validator.New(testCfg, testLog)
			testRunner := New(testCfg, testValidator, testLog)

			// Reset the output buffers
			stdout.Reset()
			stderr.Reset()
			testRunner.SetOutputs(stdout, stderr)

			ctx := t.Context()
			err := testRunner.RunCommand(ctx, tt.command, tt.workingDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
