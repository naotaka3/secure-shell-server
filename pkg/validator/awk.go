package validator

import (
	"fmt"
	"os"
	"strings"
)

const (
	// flagLongFile is the long form of the -f flag used by awk and sed.
	flagLongFile = "--file"

	// pipeCoprocessLen is the length of the "|&" coprocess operator.
	pipeCoprocessLen = 2
)

// AwkValidator handles specific awk/gawk/mawk/nawk command validation.
type AwkValidator struct{}

// NewAwkValidator creates a new AwkValidator.
func NewAwkValidator() *AwkValidator {
	return &AwkValidator{}
}

// dangerousAwkPatterns contains patterns that indicate command execution in awk scripts.
var dangerousAwkPatterns = []string{
	"system(",   // system() function executes shell commands
	"| getline", // pipe from command via getline (with space)
	"|getline",  // pipe from command via getline (without space)
	"|&",        // two-way pipe to coprocess
	"@load",     // gawk extension loading (arbitrary native code)
}

// IsAwkCommand checks if the command is an awk variant.
func IsAwkCommand(cmd string) bool {
	switch cmd {
	case "awk", "gawk", "mawk", "nawk":
		return true
	}
	return false
}

// ValidateAwkArgs checks awk arguments for dangerous patterns that could execute commands.
// It inspects both inline scripts and -e/--source scripts.
// Returns (hasDanger, description).
func (a *AwkValidator) ValidateAwkArgs(args []string) (bool, string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// -e/--source takes an awk script as value — must validate it
		if isAwkScriptFlag(arg) {
			if i+1 < len(args) {
				i++
				if hasDangerousAwkPattern(args[i]) {
					return true, "awk script contains dangerous command execution pattern"
				}
			}
			continue
		}

		// Check -f/--file flag: read and validate script file contents
		if isAwkFileFlag(arg) {
			if i+1 < len(args) {
				i++
				content, err := os.ReadFile(args[i])
				if err != nil {
					return true, fmt.Sprintf("cannot read awk script file %q for validation", args[i])
				}
				if hasDangerousAwkPattern(string(content)) {
					return true, "awk script file contains dangerous command execution pattern"
				}
			}
			continue
		}

		// Skip flags that take a non-script value argument
		if isAwkNonScriptFlagWithValue(arg) {
			i++ // skip the next argument (the value)
			continue
		}

		// Skip flags
		if strings.HasPrefix(arg, "-") {
			continue
		}

		// This is likely the awk script (first non-flag argument)
		if hasDangerousAwkPattern(arg) {
			return true, "awk script contains dangerous command execution pattern"
		}

		// After the script, remaining args are input files — stop checking
		break
	}

	return false, ""
}

// isAwkScriptFlag returns true if the flag takes an awk script as its value.
// These values must be validated for dangerous patterns.
func isAwkScriptFlag(flag string) bool {
	switch flag {
	case "-e", "--source":
		return true
	}
	return false
}

// isFlagWithAwkValue returns true if the awk flag takes a subsequent value argument.
// This includes script flags, file flags, and non-script flags.
func isFlagWithAwkValue(flag string) bool {
	return isAwkScriptFlag(flag) || isAwkFileFlag(flag) || isAwkNonScriptFlagWithValue(flag)
}

// isAwkFileFlag returns true if the flag is -f/--file (script file flag).
func isAwkFileFlag(flag string) bool {
	switch flag {
	case "-f", flagLongFile:
		return true
	}
	return false
}

// isAwkNonScriptFlagWithValue returns true if the flag takes a non-script value argument.
func isAwkNonScriptFlagWithValue(flag string) bool {
	switch flag {
	case "-v", "--assign",
		"-F", "--field-separator",
		"-o", "--pretty-print":
		return true
	}
	return false
}

// hasDangerousAwkPattern checks if an awk script string contains dangerous patterns.
func hasDangerousAwkPattern(script string) bool {
	lower := strings.ToLower(script)

	for _, pattern := range dangerousAwkPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	// Check for pipe operator used for command output: print ... | "command"
	// This catches patterns like: print "data" | "sh"
	if containsAwkPipeOutput(script) {
		return true
	}

	// Check for getline < "file" pattern which reads from arbitrary files
	if containsAwkGetlineFromFile(script) {
		return true
	}

	return false
}

// Safe: getline (from stdin), getline var (from stdin).
func containsAwkGetlineFromFile(script string) bool {
	idx := 0
	for idx < len(script) {
		glIdx := strings.Index(script[idx:], "getline")
		if glIdx == -1 {
			break
		}
		glIdx += idx

		// Look for '<' after "getline" (possibly with variable name and whitespace in between)
		rest := script[glIdx+len("getline"):]
		// Skip optional whitespace and variable name, then look for '<'
		i := 0
		// Skip whitespace
		for i < len(rest) && (rest[i] == ' ' || rest[i] == '\t') {
			i++
		}
		// Skip optional variable name (identifier)
		if i < len(rest) && isAwkIdentStart(rest[i]) {
			i++
			for i < len(rest) && isAwkIdentChar(rest[i]) {
				i++
			}
			// Skip whitespace after variable name
			for i < len(rest) && (rest[i] == ' ' || rest[i] == '\t') {
				i++
			}
		}
		// Check for '<'
		if i < len(rest) && rest[i] == '<' {
			return true
		}

		idx = glIdx + len("getline")
	}
	return false
}

func isAwkIdentStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isAwkIdentChar(c byte) bool {
	return isAwkIdentStart(c) || (c >= '0' && c <= '9')
}

// filterAwkNonPathArgs returns only the input file arguments from awk args,
// skipping flags, flag values, and the awk script itself.
// This prevents awk scripts like '{print $1}' from being validated as file paths.
func filterAwkNonPathArgs(args []string) []string {
	var result []string
	scriptFound := false

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Skip flags that take a value argument
		if isFlagWithAwkValue(arg) {
			i++ // skip the next argument (the value)
			continue
		}

		// Skip other flags
		if strings.HasPrefix(arg, "-") {
			continue
		}

		// First non-flag argument is the awk script — skip it
		if !scriptFound {
			scriptFound = true
			continue
		}

		// Remaining arguments are input files — keep them for path validation
		result = append(result, arg)
	}

	return result
}

// containsAwkPipeOutput detects awk pipe-to-command patterns like: print ... | "cmd"
// In awk, `print "x" | "cmd"` pipes output to a shell command.
func containsAwkPipeOutput(script string) bool {
	// Look for pipe followed by a quoted string (command)
	// This is a heuristic: in awk, `| "something"` means pipe to command
	idx := 0
	for idx < len(script) {
		pipeIdx := strings.Index(script[idx:], "|")
		if pipeIdx == -1 {
			break
		}
		pipeIdx += idx

		// Skip |& which is already caught by dangerousAwkPatterns
		if pipeIdx+1 < len(script) && script[pipeIdx+1] == '&' {
			idx = pipeIdx + pipeCoprocessLen
			continue
		}

		// Check if there's a quoted string after the pipe (possibly with whitespace)
		rest := strings.TrimSpace(script[pipeIdx+1:])
		if len(rest) > 0 && (rest[0] == '"' || rest[0] == '\'') {
			return true
		}

		idx = pipeIdx + 1
	}

	return false
}
