package validator

import (
	"testing"
)

// TestIsSedCommand tests the IsSedCommand function.
func TestIsSedCommand(t *testing.T) {
	tests := []struct {
		cmd  string
		want bool
	}{
		{"sed", true},
		{"gsed", true},
		{"awk", false},
		{"grep", false},
		{"cat", false},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			if got := IsSedCommand(tt.cmd); got != tt.want {
				t.Errorf("IsSedCommand(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}

// TestValidateSedArgs tests the SedValidator.ValidateSedArgs function.
func TestValidateSedArgs(t *testing.T) {
	v := NewSedValidator()

	// Safe patterns
	testSafeSedPatterns(t, v)

	// Dangerous patterns
	testDangerousSedPatterns(t, v)
}

// testSafeSedPatterns tests sed scripts that should be allowed.
func testSafeSedPatterns(t *testing.T, v *SedValidator) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "SimpleSubstitution",
			args: []string{"s/foo/bar/", "input.txt"},
		},
		{
			name: "GlobalSubstitution",
			args: []string{"s/foo/bar/g", "input.txt"},
		},
		{
			name: "CaseInsensitiveSubstitution",
			args: []string{"s/foo/bar/gi", "input.txt"},
		},
		{
			name: "DeleteLine",
			args: []string{"/pattern/d", "input.txt"},
		},
		{
			name: "PrintLine",
			args: []string{"-n", "/pattern/p", "input.txt"},
		},
		{
			name: "InPlaceEdit",
			args: []string{"-i", "s/old/new/g", "input.txt"},
		},
		{
			name: "MultipleExpressions",
			args: []string{"-e", "s/foo/bar/g", "-e", "s/baz/qux/g", "input.txt"},
		},
		{
			name: "LineRange",
			args: []string{"1,5s/foo/bar/", "input.txt"},
		},
		{
			name: "AppendText",
			args: []string{"/pattern/a\\newtext", "input.txt"},
		},
		{
			name: "AlternateSeparator",
			args: []string{"s|foo|bar|g", "input.txt"},
		},
		{
			name: "WithFileFlag",
			args: []string{"-f", "script.sed", "input.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasDanger, desc := v.ValidateSedArgs(tt.args)
			if hasDanger {
				t.Errorf("ValidateSedArgs(%v) unexpectedly flagged as dangerous: %s", tt.args, desc)
			}
		})
	}
}

// testDangerousSedPatterns tests sed scripts that should be blocked.
func testDangerousSedPatterns(t *testing.T, v *SedValidator) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "SubstitutionWithEFlag",
			args: []string{"s/pattern/replacement/e", "input.txt"},
		},
		{
			name: "SubstitutionWithEFlagAndGlobal",
			args: []string{"s/pattern/replacement/ge", "input.txt"},
		},
		{
			name: "SubstitutionWithEFlagAlternateSeparator",
			args: []string{"s|pattern|replacement|e", "input.txt"},
		},
		{
			name: "StandaloneECommand",
			args: []string{"e", "input.txt"},
		},
		{
			name: "ECommandWithArgument",
			args: []string{"e date", "input.txt"},
		},
		{
			name: "ECommandAfterSemicolon",
			args: []string{"s/foo/bar/;e", "input.txt"},
		},
		{
			name: "ECommandWithExpressionFlag",
			args: []string{"-e", "s/foo/bar/e", "input.txt"},
		},
		{
			name: "ECommandInExpression",
			args: []string{"-e", "e date", "input.txt"},
		},
		{
			name: "SubstitutionWithCommaSeparatorEFlag",
			args: []string{"s,pattern,replacement,e", "input.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasDanger, _ := v.ValidateSedArgs(tt.args)
			if !hasDanger {
				t.Errorf("ValidateSedArgs(%v) should have been flagged as dangerous", tt.args)
			}
		})
	}
}
