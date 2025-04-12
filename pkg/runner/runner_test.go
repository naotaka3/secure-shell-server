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
			name:    "allowed command",
			command: "echo hello\nls -l",
			wantErr: false,
		},
		{
			name:    "disallowed command",
			command: "echo hello\nrm -rf /",
			wantErr: true,
		},
		{
			name:    "syntax error",
			command: "echo 'unclosed string",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the output buffers
			stdout.Reset()
			stderr.Reset()

			ctx := t.Context()
			err := safeRunner.RunCommand(ctx, tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
