package validator

import (
	"strings"
	"testing"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/pkg/logger"
)

func TestValidator_ValidateScript(t *testing.T) {
	cfg := config.NewDefaultConfig()
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

func TestValidator_ValidateScriptFile(t *testing.T) {
	cfg := config.NewDefaultConfig()
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.script)
			gotOk, err := validator.ValidateScriptFile(reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateScriptFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOk != tt.wantOk {
				t.Errorf("ValidateScriptFile() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
