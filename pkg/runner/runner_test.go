package runner

import (
	"bytes"
	"testing"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/pkg/logger"
	"github.com/shimizu1995/secure-shell-server/pkg/validator"
)

func TestSafeRunner_RunScript(t *testing.T) {
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
		script  string
		wantErr bool
	}{
		{
			name:    "allowed script",
			script:  "echo hello\nls -l",
			wantErr: false,
		},
		{
			name:    "disallowed script",
			script:  "echo hello\nrm -rf /",
			wantErr: true,
		},
		{
			name:    "syntax error",
			script:  "echo 'unclosed string",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the output buffers
			stdout.Reset()
			stderr.Reset()

			ctx := t.Context()
			err := safeRunner.RunScript(ctx, tt.script)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
