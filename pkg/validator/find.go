package validator

import (
	"strings"
)

// FindParser handles specific find command validation.
type FindParser struct{}

// NewFindParser creates a new FindParser.
func NewFindParser() *FindParser {
	return &FindParser{}
}

// ParseFindExecArgs parses find command arguments to extract the command executed by -exec.
// Returns:
// - The extracted command.
// - The arguments for the command.
// - Whether a command was successfully extracted.
// - Error message if any.
func (f *FindParser) ParseFindExecArgs(args []string) ([]string, bool, string) {
	// No args is valid for find (lists current directory)
	if len(args) == 0 {
		return nil, false, ""
	}

	// Find all -exec or -execdir arguments and their associated commands
	commands := extractExecCommands(args)
	if len(commands) == 0 {
		return nil, false, ""
	}

	return commands, true, ""
}

// FilterFindSpecialArgs removes find's special arguments like \; and + from the args list
// to prevent them from being interpreted as paths during validation.
func (f *FindParser) FilterFindSpecialArgs(args []string) []string {
	filtered := make([]string, 0, len(args))
	for _, arg := range args {
		// Skip find's special terminators
		if arg == ";" || arg == "\\;" || arg == "+" {
			continue
		}
		filtered = append(filtered, arg)
	}
	return filtered
}

// extractExecCommands extracts all commands that follow -exec or -execdir in find arguments.
func extractExecCommands(args []string) []string {
	var commands []string
	for i := 0; i < len(args)-1; i++ {
		// Check for -exec or -execdir flags
		if args[i] == "-exec" || args[i] == "-execdir" {
			// Extract all arguments until \; or + is encountered
			j := i + 1
			for j < len(args) {
				// Check if we've reached the end of the command (\; or +)
				if args[j] == ";" || args[j] == "\\;" || args[j] == "+" {
					break
				}

				// Add the command or argument to our list
				if j == i+1 && !strings.HasPrefix(args[j], "{") {
					// This is the command itself, add it to our commands list
					commands = append(commands, args[j])
				}

				j++
			}

			// Skip to the end of this -exec clause
			i = j
		}
	}

	return commands
}
