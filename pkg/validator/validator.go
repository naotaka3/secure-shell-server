package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mvdan.cc/sh/v3/syntax"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/pkg/logger"
)

const (
	// DirPermissions represents the permission bits for directories.
	DirPermissions = 0o755
	// FilePermissions represents the permission bits for files.
	FilePermissions = 0o644
)

// CommandValidator validates shell commands.
type CommandValidator struct {
	config *config.ShellCommandConfig
	logger *logger.Logger
}

// New creates a new CommandValidator.
func New(config *config.ShellCommandConfig, logger *logger.Logger) *CommandValidator {
	return &CommandValidator{
		config: config,
		logger: logger,
	}
}

// ValidateScript validates a shell script string.
func (v *CommandValidator) ValidateScript(script string) (bool, error) {
	parser := syntax.NewParser()
	prog, err := parser.Parse(strings.NewReader(script), "")
	if err != nil {
		v.logger.LogErrorf("Parse error: %v", err)
		return false, fmt.Errorf("parse error: %w", err)
	}

	valid := true
	var validationErr error

	syntax.Walk(prog, func(node syntax.Node) bool {
		if call, ok := node.(*syntax.CallExpr); ok && len(call.Args) > 0 {
			// Extract the command name from the first argument
			word := call.Args[0]
			if len(word.Parts) > 0 {
				if lit, ok := word.Parts[0].(*syntax.Lit); ok {
					cmd := lit.Value
					args := extractArgs(call.Args)

					allowed, errMsg := v.validateCommand(cmd, args)
					if !allowed {
						validationErr = fmt.Errorf("%s", errMsg)
						valid = false
						v.logger.LogCommandAttempt(cmd, args, false)

						// If a block log path is specified, log the blocked command
						if v.config.BlockLogPath != "" {
							v.logBlockedCommand(cmd, args, errMsg)
						}

						return false
					}
					v.logger.LogCommandAttempt(cmd, args, true)
				}
			}
		}
		return true
	})

	return valid, validationErr
}

// IsDirectoryAllowed checks if a given directory is allowed to run commands in.
func (v *CommandValidator) IsDirectoryAllowed(dir string) (bool, string) {
	// If the directory is empty, use the working directory from the config
	if dir == "" {
		dir = v.config.WorkingDir
	}

	// Check if the directory is in the allowed directories list or is a subdirectory of an allowed directory
	for _, allowedDir := range v.config.AllowedDirectories {
		if strings.HasPrefix(dir, allowedDir) {
			return true, ""
		}
	}

	return false, fmt.Sprintf("directory %q is not allowed: %s", dir, v.config.DefaultErrorMessage)
}

// ValidateCommandInDirectory checks if a command is allowed to run in a specific directory.
func (v *CommandValidator) ValidateCommandInDirectory(cmd string, args []string, dir string) (bool, error) {
	// First check if the directory is allowed
	allowed, message := v.IsDirectoryAllowed(dir)
	if !allowed {
		return false, fmt.Errorf("%s", message)
	}

	// Then check if the command itself is allowed
	allowed, message = v.validateCommand(cmd, args)
	if !allowed {
		return false, fmt.Errorf("%s", message)
	}

	return true, nil
}

// validateCommand checks if a command is allowed based on the configuration.
func (v *CommandValidator) validateCommand(cmd string, args []string) (bool, string) {
	// Check if the command is explicitly denied
	if denied, message := v.isCommandExplicitlyDenied(cmd); denied {
		return false, message
	}

	// Check if the command is explicitly allowed
	for _, allowed := range v.config.AllowCommands {
		if allowed.Command == cmd {
			// If there are no subcommands specified, the command is allowed without restrictions
			if len(allowed.SubCommands) == 0 && len(allowed.DenySubCommands) == 0 {
				return true, ""
			}

			// Check subcommand permissions
			if allowed, message := v.checkSubCommandPermissions(cmd, args, allowed); !allowed {
				return false, message
			}

			return true, ""
		}
	}

	// If command was not found in the allow list, it's denied
	return false, fmt.Sprintf("command %q is not permitted: %s", cmd, v.config.DefaultErrorMessage)
}

// isCommandExplicitlyDenied checks if a command is explicitly denied in the configuration.
func (v *CommandValidator) isCommandExplicitlyDenied(cmd string) (bool, string) {
	for _, denied := range v.config.DenyCommands {
		if denied.Command == cmd {
			message := v.config.DefaultErrorMessage
			if denied.Message != "" {
				message = denied.Message
			}
			return true, fmt.Sprintf("command %q is denied: %s", cmd, message)
		}
	}
	return false, ""
}

// checkSubCommandPermissions checks if the subcommand is allowed for the specified command.
func (v *CommandValidator) checkSubCommandPermissions(cmd string, args []string, allowed config.AllowCommand) (bool, string) {
	// If there are subcommands specified, check if the first argument matches any of them
	if len(allowed.SubCommands) > 0 && len(args) > 0 {
		subCommandAllowed := false
		for _, subCmd := range allowed.SubCommands {
			if args[0] == subCmd {
				subCommandAllowed = true
				break
			}
		}

		if !subCommandAllowed {
			return false, fmt.Sprintf("subcommand %q is not allowed for command %q", args[0], cmd)
		}
	}

	// If there are denied subcommands specified, check if the first argument matches any of them
	if len(allowed.DenySubCommands) > 0 && len(args) > 0 {
		for _, deniedSubCmd := range allowed.DenySubCommands {
			if args[0] == deniedSubCmd {
				return false, fmt.Sprintf("subcommand %q is denied for command %q", args[0], cmd)
			}
		}
	}

	return true, ""
}

// logBlockedCommand logs blocked commands to the specified file.
func (v *CommandValidator) logBlockedCommand(cmd string, args []string, reason string) {
	if v.config.BlockLogPath == "" {
		return
	}

	// Ensure the directory exists
	dir := filepath.Dir(v.config.BlockLogPath)
	if err := os.MkdirAll(dir, DirPermissions); err != nil {
		v.logger.LogErrorf("Failed to create directory for block log: %v", err)
		return
	}

	// Open the log file in append mode
	f, err := os.OpenFile(v.config.BlockLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, FilePermissions)
	if err != nil {
		v.logger.LogErrorf("Failed to open block log file: %v", err)
		return
	}
	defer f.Close()

	// Create log entry
	timestamp := time.Now().Format(time.RFC3339)
	logEntry := fmt.Sprintf("%s [BLOCKED] Command: %s %v, Reason: %s\n", timestamp, cmd, args, reason)

	// Write to log file
	if _, err := f.WriteString(logEntry); err != nil {
		v.logger.LogErrorf("Failed to write to block log file: %v", err)
	}
}

// extractArgs extracts command arguments as strings.
func extractArgs(args []*syntax.Word) []string {
	if len(args) <= 1 {
		return []string{}
	}

	result := make([]string, 0, len(args)-1)
	for i := 1; i < len(args); i++ {
		word := args[i]
		if len(word.Parts) > 0 {
			if lit, ok := word.Parts[0].(*syntax.Lit); ok {
				result = append(result, lit.Value)
			}
		}
	}

	return result
}
