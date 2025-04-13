package validator

import (
	"strings"
)

// XargsParser handles specific xargs command validation.
type XargsParser struct{}

// NewXargsParser creates a new XargsParser.
func NewXargsParser() *XargsParser {
	return &XargsParser{}
}

// - Error message if any.
func (x *XargsParser) ParseXargsCommand(args []string) (string, []string, bool, string) {
	if len(args) == 0 {
		return "", nil, false, "no arguments provided to xargs"
	}

	// Try to extract command with explicit exec flag first
	execCmd, execArgs, found := findExecCommand(args)
	if found {
		return execCmd, execArgs, true, ""
	}

	// If no exec command, try to find command after flags
	execCmd, execArgs, found = findCommandAfterFlags(args)
	if found {
		return execCmd, execArgs, true, ""
	}

	return "", nil, false, "unable to determine command to be executed by xargs"
}

// findExecCommand looks for -exec or --exec flag and extracts the command that follows.
func findExecCommand(args []string) (string, []string, bool) {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "-exec" || args[i] == "--exec" {
			execCmd := args[i+1]
			var execArgs []string
			if i+2 < len(args) {
				execArgs = args[i+2:]
			}
			return execCmd, execArgs, true
		}
	}
	return "", nil, false
}

// findCommandAfterFlags tries to find the first non-flag argument, skipping any flag options.
func findCommandAfterFlags(args []string) (string, []string, bool) {
	i := 0
	for i < len(args) {
		// If not a flag, assume it's the command
		if !strings.HasPrefix(args[i], "-") {
			execCmd := args[i]
			var execArgs []string
			if i+1 < len(args) {
				execArgs = args[i+1:]
			}
			return execCmd, execArgs, true
		}

		// Handle flags that take an additional argument
		arg := args[i]
		i++ // Move to next arg by default

		if isFlagWithArg(arg) && i < len(args) {
			// Skip the next argument which is the value for this flag
			i++
		}
	}

	return "", nil, false
}

// isFlagWithArg checks if a flag takes an additional argument.
func isFlagWithArg(flag string) bool {
	// Common xargs flags that take values
	flagsWithArgs := []string{
		"-a", "--arg-file",
		"-E", "--eof",
		"--max-args", "-n",
		"--max-chars", "-s",
		"--max-lines", "-L",
		"--max-procs", "-P",
	}

	// Check direct matches
	for _, f := range flagsWithArgs {
		if flag == f {
			return true
		}
	}

	// Handle special case for -I and -i flags
	if flag == "-i" || flag == "-I" {
		return true
	}

	// Handle flags with attached values like -I{} or -i{}
	if (strings.HasPrefix(flag, "-i") || strings.HasPrefix(flag, "-I")) && len(flag) > 2 {
		return false // Has value attached, no need to skip next arg
	}

	return false
}
