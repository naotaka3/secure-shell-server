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

// ExecCommand represents a command extracted from find -exec with its arguments.
type ExecCommand struct {
	Name string
	Args []string
}

// ParseFindExecArgs parses find command arguments to extract the command executed by -exec.
// Returns:
// - The extracted commands with their arguments.
// - Whether any commands were successfully extracted.
// - Error message if any.
func (f *FindParser) ParseFindExecArgs(args []string) ([]ExecCommand, bool, string) {
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
func extractExecCommands(args []string) []ExecCommand {
	var commands []ExecCommand
	for i := 0; i < len(args)-1; i++ {
		// Check for -exec or -execdir flags
		if args[i] == "-exec" || args[i] == "-execdir" {
			// Collect all arguments until \; or + is encountered
			j := i + 1
			var cmdParts []string
			for j < len(args) {
				// Check if we've reached the end of the command (\; or +)
				if args[j] == ";" || args[j] == "\\;" || args[j] == "+" {
					break
				}

				// Skip {} placeholder
				if args[j] != "{}" {
					cmdParts = append(cmdParts, args[j])
				}

				j++
			}

			if len(cmdParts) > 0 && !strings.HasPrefix(cmdParts[0], "{") {
				cmd := ExecCommand{
					Name: cmdParts[0],
				}
				if len(cmdParts) > 1 {
					cmd.Args = cmdParts[1:]
				}
				commands = append(commands, cmd)
			}

			// Skip to the end of this -exec clause
			i = j
		}
	}

	return commands
}
