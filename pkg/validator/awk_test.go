package validator

import (
	"testing"
)

// TestIsAwkCommand tests the IsAwkCommand function.
func TestIsAwkCommand(t *testing.T) {
	tests := []struct {
		cmd  string
		want bool
	}{
		{"awk", true},
		{"gawk", true},
		{"mawk", true},
		{"nawk", true},
		{"sed", false},
		{"grep", false},
		{"cat", false},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			if got := IsAwkCommand(tt.cmd); got != tt.want {
				t.Errorf("IsAwkCommand(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}

// TestValidateAwkArgs tests the AwkValidator.ValidateAwkArgs function.
func TestValidateAwkArgs(t *testing.T) {
	v := NewAwkValidator()

	// Safe patterns
	testSafeAwkPatterns(t, v)

	// Dangerous patterns
	testDangerousAwkPatterns(t, v)
}

// testSafeAwkPatterns tests awk scripts that should be allowed.
func testSafeAwkPatterns(t *testing.T, v *AwkValidator) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "SimplePrint",
			args: []string{"{print $1}"},
		},
		{
			name: "PrintWithFieldSeparator",
			args: []string{"-F", ":", "{print $1}"},
		},
		{
			name: "PatternMatch",
			args: []string{"/error/{print $0}", "logfile.txt"},
		},
		{
			name: "BeginEndBlock",
			args: []string{"BEGIN{sum=0} {sum+=$1} END{print sum}"},
		},
		{
			name: "VariableAssignment",
			args: []string{"-v", "threshold=10", "$1 > threshold {print $0}"},
		},
		{
			name: "MathOperations",
			args: []string{"{print $1 * 2}"},
		},
		{
			name: "StringFunctions",
			args: []string{"{print substr($0, 1, 10)}"},
		},
		{
			name: "ConditionalPrint",
			args: []string{"NR > 5 && NR < 10 {print}"},
		},
		{
			name: "PipeLiteralWithoutQuotes",
			args: []string{"{gsub(/|/, \"-\"); print}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasDanger, desc := v.ValidateAwkArgs(tt.args)
			if hasDanger {
				t.Errorf("ValidateAwkArgs(%v) unexpectedly flagged as dangerous: %s", tt.args, desc)
			}
		})
	}
}

// testDangerousAwkPatterns tests awk scripts that should be blocked.
func testDangerousAwkPatterns(t *testing.T, v *AwkValidator) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "SystemCall",
			args: []string{`BEGIN { system("rm -rf /") }`},
		},
		{
			name: "SystemCallLowerCase",
			args: []string{`{system("cat /etc/passwd")}`},
		},
		{
			name: "PipeToCommand",
			args: []string{`{print $0 | "sh"}`},
		},
		{
			name: "PipeToCommandWithSpace",
			args: []string{`{print $0 | "mail -s subject user@example.com"}`},
		},
		{
			name: "GetlineFromPipe",
			args: []string{`{"date" | getline current_date; print current_date}`},
		},
		{
			name: "TwoWayPipe",
			args: []string{`BEGIN { print "hello" |& "/bin/sh" }`},
		},
		{
			name: "SystemWithVariable",
			args: []string{`{cmd="ls"; system(cmd)}`},
		},
		{
			name: "PipeWithSingleQuote",
			args: []string{`{print $0 | 'sh'}`},
		},
		{
			name: "SystemViaSourceFlag",
			args: []string{"-e", `{system("id")}`},
		},
		{
			name: "SystemViaLongSourceFlag",
			args: []string{"--source", `BEGIN{system("whoami")}`},
		},
		{
			name: "PipeViaSourceFlag",
			args: []string{"-e", `{print $0 | "sh"}`},
		},
		{
			name: "GawkLoadExtension",
			args: []string{`@load "evil"; BEGIN{run()}`},
		},
		{
			name: "GetlineWithoutSpace",
			args: []string{`{"date"|getline d; print d}`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasDanger, _ := v.ValidateAwkArgs(tt.args)
			if !hasDanger {
				t.Errorf("ValidateAwkArgs(%v) should have been flagged as dangerous", tt.args)
			}
		})
	}
}
