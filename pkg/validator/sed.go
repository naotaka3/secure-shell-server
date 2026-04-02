package validator

import (
	"fmt"
	"os"
	"strings"
)

// SedValidator handles specific sed/gsed command validation.
type SedValidator struct{}

// NewSedValidator creates a new SedValidator.
func NewSedValidator() *SedValidator {
	return &SedValidator{}
}

// IsSedCommand checks if the command is a sed variant.
func IsSedCommand(cmd string) bool {
	switch cmd {
	case "sed", "gsed":
		return true
	}
	return false
}

// ValidateSedArgs checks sed arguments for dangerous patterns that could execute commands.
// The 'e' command/flag in sed executes the pattern space (or replacement) as a shell command.
// Returns (hasDanger, description).
func (s *SedValidator) ValidateSedArgs(args []string) (bool, string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// -e and -f take a value argument
		if arg == "-e" || arg == "--expression" {
			if i+1 < len(args) {
				i++
				if hasDangerousSedPattern(args[i]) {
					return true, "sed expression contains dangerous 'e' command that executes shell commands"
				}
			}
			continue
		}

		if arg == "-f" || arg == "--file" {
			if i+1 < len(args) {
				i++
				content, err := os.ReadFile(args[i])
				if err != nil {
					return true, fmt.Sprintf("cannot read sed script file %q for validation", args[i])
				}
				if hasDangerousSedPattern(string(content)) {
					return true, "sed script file contains dangerous 'e' command that executes shell commands"
				}
			}
			continue
		}

		// Skip other flags
		if strings.HasPrefix(arg, "-") {
			// Handle combined flags like -ie or -ne (inline with e)
			// Check if 'e' appears as a standalone sed flag (not -e for expression)
			continue
		}

		// First non-flag argument is the sed script
		if hasDangerousSedPattern(arg) {
			return true, "sed script contains dangerous 'e' command that executes shell commands"
		}

		// After the script, remaining args are input files
		break
	}

	return false, ""
}

// hasDangerousSedPattern checks if a sed script contains the dangerous 'e' command or flag.
func hasDangerousSedPattern(script string) bool {
	// Check for standalone 'e' command: executes the pattern space as a shell command
	// In sed, a bare 'e' on its own line or after semicolon executes pattern space
	if containsSedECommand(script) {
		return true
	}

	// Check for 'e' flag on substitution: s/pattern/replacement/e
	if containsSedSubstitutionEFlag(script) {
		return true
	}

	return false
}

// containsSedSubstitutionEFlag checks for the 'e' flag on sed substitution commands.
// It parses s-commands like s/foo/bar/e, s|foo|bar|ge, s,foo,bar,ei etc.
func containsSedSubstitutionEFlag(script string) bool {
	commands := splitSedCommands(script)
	for _, cmd := range commands {
		trimmed := strings.TrimSpace(cmd)

		// Strip optional address prefix (line number, /pattern/, etc.)
		trimmed = skipSedAddress(trimmed)

		// Check if this is a substitution command
		if len(trimmed) < 2 || trimmed[0] != 's' {
			continue
		}

		// The character after 's' is the delimiter
		delimiter := trimmed[1]
		flags := extractSedSubstitutionFlags(trimmed[2:], delimiter)

		// Check if 'e' is among the flags
		if strings.ContainsRune(flags, 'e') {
			return true
		}
	}
	return false
}

// skipSedAddress strips a leading sed address (line numbers, /regex/) from a command.
func skipSedAddress(cmd string) string {
	if len(cmd) == 0 {
		return cmd
	}

	i := 0

	// Skip line number address like "1,5"
	for i < len(cmd) && (cmd[i] >= '0' && cmd[i] <= '9' || cmd[i] == ',' || cmd[i] == '$' || cmd[i] == '~') {
		i++
	}

	// Skip /regex/ address
	if i < len(cmd) && cmd[i] == '/' {
		i++ // skip opening /
		for i < len(cmd) && cmd[i] != '/' {
			if cmd[i] == '\\' && i+1 < len(cmd) {
				i++ // skip escaped char
			}
			i++
		}
		if i < len(cmd) {
			i++ // skip closing /
		}
	}

	return cmd[i:]
}

// extractSedSubstitutionFlags parses the body of a sed s-command after the opening delimiter
// and returns the flags string. E.g., for input `foo/bar/ge` with delimiter '/', returns "ge".
func extractSedSubstitutionFlags(body string, delimiter byte) string {
	// We need to find the 2nd occurrence of the delimiter to get past pattern and replacement
	count := 0
	i := 0
	for i < len(body) {
		if body[i] == '\\' && i+1 < len(body) {
			i += 2 // skip escaped character
			continue
		}
		if body[i] == delimiter {
			count++
			if count == 2 {
				// Everything after this delimiter is flags
				return body[i+1:]
			}
		}
		i++
	}
	return ""
}

// filterSedNonPathArgs returns only the input file arguments from sed args,
// skipping flags, flag values, and the sed script itself.
// This prevents sed expressions like 's/foo/bar/' from being validated as file paths.
func filterSedNonPathArgs(args []string) []string {
	var result []string
	scriptFound := false

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// -e and -f take a value argument — skip both
		if arg == "-e" || arg == "--expression" || arg == "-f" || arg == "--file" {
			if i+1 < len(args) {
				i++ // skip the value
			}
			continue
		}

		// Skip other flags
		if strings.HasPrefix(arg, "-") {
			continue
		}

		// First non-flag argument is the sed script (unless -e was used) — skip it
		if !scriptFound {
			scriptFound = true
			continue
		}

		// Remaining arguments are input files — keep them for path validation
		result = append(result, arg)
	}

	return result
}

// containsSedECommand checks for the standalone 'e' command in sed scripts.
// The 'e' command (without arguments) executes the contents of pattern space as a shell command.
// The 'e command' form executes the given command.
func containsSedECommand(script string) bool {
	// Split by semicolons and newlines to check individual commands
	commands := splitSedCommands(script)
	for _, cmd := range commands {
		trimmed := strings.TrimSpace(cmd)

		// Standalone 'e' command
		if trimmed == "e" {
			return true
		}

		// 'e' followed by a command to execute: 'e date', 'e ls -la'
		if strings.HasPrefix(trimmed, "e ") || strings.HasPrefix(trimmed, "e\t") {
			return true
		}
	}

	return false
}

// splitSedCommands splits a sed script into individual commands by semicolons and newlines.
func splitSedCommands(script string) []string {
	var commands []string
	// Split by semicolons first
	parts := strings.Split(script, ";")
	for _, part := range parts {
		// Then split by newlines
		lines := strings.Split(part, "\n")
		commands = append(commands, lines...)
	}
	return commands
}
