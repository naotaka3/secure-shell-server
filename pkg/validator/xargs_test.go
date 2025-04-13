package validator

import (
	"reflect"
	"testing"
)

// TestParseXargsCommand tests the ParseXargsCommand function.
func TestParseXargsCommand(t *testing.T) {
	parser := NewXargsParser()

	// Test basic command parsing
	testSimpleCommands(t, parser)

	// Test flag handling
	testFlagHandling(t, parser)

	// Test edge cases
	testEdgeCases(t, parser)
}

// testSimpleCommands tests basic command parsing without flags.
func testSimpleCommands(t *testing.T, parser *XargsParser) {
	tests := []struct {
		name       string
		args       []string
		wantCmd    string
		wantArgs   []string
		wantValid  bool
		wantErrMsg string
	}{
		{
			name:       "SimpleCommand",
			args:       []string{"echo", "hello"},
			wantCmd:    "echo",
			wantArgs:   []string{"hello"},
			wantValid:  true,
			wantErrMsg: "",
		},
		{
			name:       "CommandWithMultipleArgs",
			args:       []string{"grep", "pattern", "file.txt"},
			wantCmd:    "grep",
			wantArgs:   []string{"pattern", "file.txt"},
			wantValid:  true,
			wantErrMsg: "",
		},
	}

	runParserTests(t, parser, tests)
}

// testFlagHandling tests parsing of commands with various flags.
func testFlagHandling(t *testing.T, parser *XargsParser) {
	tests := []struct {
		name       string
		args       []string
		wantCmd    string
		wantArgs   []string
		wantValid  bool
		wantErrMsg string
	}{
		{
			name:       "WithFlags",
			args:       []string{"-L", "1", "echo", "line"},
			wantCmd:    "echo",
			wantArgs:   []string{"line"},
			wantValid:  true,
			wantErrMsg: "",
		},
		{
			name:       "WithLongFlags",
			args:       []string{"--max-args", "1", "echo", "file"},
			wantCmd:    "echo",
			wantArgs:   []string{"file"},
			wantValid:  true,
			wantErrMsg: "",
		},
		{
			name:       "WithExec",
			args:       []string{"-exec", "grep", "pattern", "file.txt"},
			wantCmd:    "grep",
			wantArgs:   []string{"pattern", "file.txt"},
			wantValid:  true,
			wantErrMsg: "",
		},
		{
			name:       "WithLongExec",
			args:       []string{"--exec", "cp", "src", "dest"},
			wantCmd:    "cp",
			wantArgs:   []string{"src", "dest"},
			wantValid:  true,
			wantErrMsg: "",
		},
		{
			name:       "WithReplacementFlag",
			args:       []string{"-I", "{}", "cp", "{}", "/backup/"},
			wantCmd:    "cp",
			wantArgs:   []string{"{}", "/backup/"},
			wantValid:  true,
			wantErrMsg: "",
		},
		{
			name:       "WithReplacementFlagNoSpace",
			args:       []string{"-I{}", "mv", "{}", "/archive/"},
			wantCmd:    "mv",
			wantArgs:   []string{"{}", "/archive/"},
			wantValid:  true,
			wantErrMsg: "",
		},
	}

	runParserTests(t, parser, tests)
}

// testEdgeCases tests edge cases and error conditions.
func testEdgeCases(t *testing.T, parser *XargsParser) {
	tests := []struct {
		name       string
		args       []string
		wantCmd    string
		wantArgs   []string
		wantValid  bool
		wantErrMsg string
	}{
		{
			name:       "NoArguments",
			args:       []string{},
			wantCmd:    "",
			wantArgs:   nil,
			wantValid:  false,
			wantErrMsg: "no arguments provided to xargs",
		},
		{
			name:       "OnlyFlags",
			args:       []string{"-L", "1", "-n", "1"},
			wantCmd:    "",
			wantArgs:   nil,
			wantValid:  false,
			wantErrMsg: "unable to determine command to be executed by xargs",
		},
		{
			name:       "MissingExecCommand",
			args:       []string{"-exec"},
			wantCmd:    "",
			wantArgs:   nil,
			wantValid:  false,
			wantErrMsg: "unable to determine command to be executed by xargs",
		},
	}

	runParserTests(t, parser, tests)
}

// runParserTests runs the parser tests and checks the results.
func runParserTests(t *testing.T, parser *XargsParser, tests []struct {
	name       string
	args       []string
	wantCmd    string
	wantArgs   []string
	wantValid  bool
	wantErrMsg string
},
) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotArgs, gotValid, gotErrMsg := parser.ParseXargsCommand(tt.args)

			if gotCmd != tt.wantCmd {
				t.Errorf("ParseXargsCommand() command = %v, want %v", gotCmd, tt.wantCmd)
			}

			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("ParseXargsCommand() args = %v, want %v", gotArgs, tt.wantArgs)
			}

			if gotValid != tt.wantValid {
				t.Errorf("ParseXargsCommand() valid = %v, want %v", gotValid, tt.wantValid)
			}

			if gotErrMsg != tt.wantErrMsg {
				t.Errorf("ParseXargsCommand() errMsg = %v, want %v", gotErrMsg, tt.wantErrMsg)
			}
		})
	}
}
