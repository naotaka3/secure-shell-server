package validator

import (
	"os"
	"path/filepath"
	"testing"
)

// scriptFile describes a temporary script file for testing.
type scriptFile struct {
	name    string
	content string
}

// validationTestCase describes a single validation test case.
type validationTestCase struct {
	name    string
	cmd     string
	args    []string
	allowed bool
	message string
}

// createScriptFiles creates temporary script files in dir and returns a map of name to path.
func createScriptFiles(t *testing.T, dir string, files []scriptFile) map[string]string {
	t.Helper()
	paths := make(map[string]string, len(files))
	for _, f := range files {
		p := filepath.Join(dir, f.name)
		if err := os.WriteFile(p, []byte(f.content), 0o600); err != nil {
			t.Fatalf("Failed to create script file %s: %v", f.name, err)
		}
		paths[f.name] = p
	}
	return paths
}

// runValidationTestCases runs a slice of validation test cases against the given validator.
func runValidationTestCases(t *testing.T, v *CommandValidator, tests []validationTestCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get working directory: %v", err)
			}

			gotAllowed, gotMessage := v.ValidateCommand(tt.cmd, tt.args, wd)
			if gotAllowed != tt.allowed {
				t.Errorf("ValidateCommand() allowed = %v, want %v", gotAllowed, tt.allowed)
			}
			if gotMessage != tt.message {
				t.Errorf("ValidateCommand() message = %q, want %q", gotMessage, tt.message)
			}
		})
	}
}
