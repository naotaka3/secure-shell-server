package runner

import (
	"bytes"
	"strings"
	"testing"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/pkg/logger"
	"github.com/shimizu1995/secure-shell-server/pkg/validator"
)

func TestSafeRunner_Run(t *testing.T) {
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
		args    []string
		wantErr bool
	}{
		{
			name:    "allowed command",
			args:    []string{"echo", "hello"},
			wantErr: false,
		},
		{
			name:    "disallowed command",
			args:    []string{"rm", "-rf", "/"},
			wantErr: true,
		},
		{
			name:    "empty args",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the output buffers
			stdout.Reset()
			stderr.Reset()

			ctx := t.Context()
			err := safeRunner.Run(ctx, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

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

func TestSafeRunner_RunScriptFile(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the output buffers
			stdout.Reset()
			stderr.Reset()

			ctx := t.Context()
			reader := strings.NewReader(tt.script)
			err := safeRunner.RunScriptFile(ctx, reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunScriptFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
