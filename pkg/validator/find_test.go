package validator

import (
	"reflect"
	"testing"
)

// TestParseFindExecArgs tests the ParseFindExecArgs function.
func TestParseFindExecArgs(t *testing.T) {
	parser := NewFindParser()

	// Test basic command parsing
	testBasicFindExecCommands(t, parser)

	// Test multiple exec commands
	testMultipleExecCommands(t, parser)

	// Test edge cases
	testFindExecEdgeCases(t, parser)
}

// testBasicFindExecCommands tests basic find -exec command parsing.
func testBasicFindExecCommands(t *testing.T, parser *FindParser) {
	tests := []struct {
		name       string
		args       []string
		wantCmds   []string
		wantValid  bool
		wantErrMsg string
	}{
		{
			name:       "SimpleExec",
			args:       []string{"-name", "*.txt", "-exec", "echo", "{}", "\\;"},
			wantCmds:   []string{"echo"},
			wantValid:  true,
			wantErrMsg: "",
		},
		{
			name:       "ExecWithPath",
			args:       []string{"-type", "f", "-exec", "/bin/rm", "{}", "\\;"},
			wantCmds:   []string{"/bin/rm"},
			wantValid:  true,
			wantErrMsg: "",
		},
		{
			name:       "ExecWithPlusTerminator",
			args:       []string{"-name", "*.log", "-exec", "gzip", "{}", "+"},
			wantCmds:   []string{"gzip"},
			wantValid:  true,
			wantErrMsg: "",
		},
		{
			name:       "ExecdirCommand",
			args:       []string{"-type", "f", "-execdir", "chmod", "+x", "{}", "\\;"},
			wantCmds:   []string{"chmod"},
			wantValid:  true,
			wantErrMsg: "",
		},
	}

	runFindParserTests(t, parser, tests)
}

// testMultipleExecCommands tests parsing of find with multiple -exec clauses.
func testMultipleExecCommands(t *testing.T, parser *FindParser) {
	tests := []struct {
		name       string
		args       []string
		wantCmds   []string
		wantValid  bool
		wantErrMsg string
	}{
		{
			name: "MultipleExecs",
			args: []string{
				"-type", "f", 
				"-name", "*.txt", 
				"-exec", "grep", "pattern", "{}", "\\;", 
				"-exec", "cp", "{}", "/backup/", "\\;",
			},
			wantCmds:   []string{"grep", "cp"},
			wantValid:  true,
			wantErrMsg: "",
		},
		{
			name: "MixedExecAndExecdir",
			args: []string{
				"-type", "f", 
				"-exec", "ls", "-la", "{}", "\\;", 
				"-execdir", "chmod", "644", "{}", "\\;",
			},
			wantCmds:   []string{"ls", "chmod"},
			wantValid:  true,
			wantErrMsg: "",
		},
	}

	runFindParserTests(t, parser, tests)
}

// testFindExecEdgeCases tests edge cases and error conditions for find -exec parsing.
func testFindExecEdgeCases(t *testing.T, parser *FindParser) {
	tests := []struct {
		name       string
		args       []string
		wantCmds   []string
		wantValid  bool
		wantErrMsg string
	}{
		{
			name:       "NoArguments",
			args:       []string{},
			wantCmds:   nil,
			wantValid:  false,
			wantErrMsg: "",
		},
		{
			name:       "NoExecCommand",
			args:       []string{"-name", "*.txt", "-type", "f"},
			wantCmds:   nil,
			wantValid:  false,
			wantErrMsg: "",
		},
		{
			name:       "ExecWithNoCommand",
			args:       []string{"-name", "*.txt", "-exec"},
			wantCmds:   nil,
			wantValid:  false,
			wantErrMsg: "",
		},
	}

	runFindParserTests(t, parser, tests)
}

// runFindParserTests runs the find parser tests and checks the results.
func runFindParserTests(t *testing.T, parser *FindParser, tests []struct {
	name       string
	args       []string
	wantCmds   []string
	wantValid  bool
	wantErrMsg string
},
) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmds, gotValid, gotErrMsg := parser.ParseFindExecArgs(tt.args)

			if !reflect.DeepEqual(gotCmds, tt.wantCmds) {
				t.Errorf("ParseFindExecArgs() commands = %v, want %v", gotCmds, tt.wantCmds)
			}

			if gotValid != tt.wantValid {
				t.Errorf("ParseFindExecArgs() valid = %v, want %v", gotValid, tt.wantValid)
			}

			if gotErrMsg != tt.wantErrMsg {
				t.Errorf("ParseFindExecArgs() errMsg = %v, want %v", gotErrMsg, tt.wantErrMsg)
			}
		})
	}
}

// TestFilterFindSpecialArgs tests the FilterFindSpecialArgs function
func TestFilterFindSpecialArgs(t *testing.T) {
	parser := NewFindParser()
	
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "NoSpecialArgs",
			args:     []string{"-name", "*.txt", "-type", "f"},
			expected: []string{"-name", "*.txt", "-type", "f"},
		},
		{
			name:     "WithSemicolon",
			args:     []string{"-name", "*.txt", "-exec", "echo", "{}" , ";"},
			expected: []string{"-name", "*.txt", "-exec", "echo", "{}"},
		},
		{
			name:     "WithEscapedSemicolon",
			args:     []string{"-name", "*.txt", "-exec", "cat", "{}", "\\;"},
			expected: []string{"-name", "*.txt", "-exec", "cat", "{}"},
		},
		{
			name:     "WithPlus",
			args:     []string{"-name", "*.txt", "-exec", "grep", "pattern", "{}", "+"},
			expected: []string{"-name", "*.txt", "-exec", "grep", "pattern", "{}"},
		},
		{
			name:     "MultipleSpecialArgs",
			args:     []string{"-name", "*.txt", "-exec", "echo", "{}", ";", "-exec", "cat", "{}", "\\;"},
			expected: []string{"-name", "*.txt", "-exec", "echo", "{}", "-exec", "cat", "{}"},
		},
		{
			name:     "EmptyArgs",
			args:     []string{},
			expected: []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.FilterFindSpecialArgs(tt.args)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("FilterFindSpecialArgs() = %v, want %v", result, tt.expected)
			}
		})
	}
}