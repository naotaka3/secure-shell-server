package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

// IsDirectoryAllowed checks if a given directory is allowed to run commands in.
func (v *CommandValidator) IsDirectoryAllowed(dir string) (bool, string) {
	// If the directory is empty, it cannot be validated
	if dir == "" {
		return false, "empty directory path is not allowed"
	}

	// Check if the directory is in the allowed directories list or is a subdirectory of an allowed directory
	for _, allowedDir := range v.config.AllowedDirectories {
		if strings.HasPrefix(dir, allowedDir) {
			return true, ""
		}
	}

	return false, fmt.Sprintf("directory %q is not allowed: %s", dir, v.config.DefaultErrorMessage)
}

// IsPathInAllowedDirectory checks if a given path (absolute or relative) is within any of the allowed directories.
func (v *CommandValidator) IsPathInAllowedDirectory(path string, baseDir string) (bool, string) {
	// Handle empty path
	if path == "" {
		return false, "empty path is not allowed"
	}

	// Determine if the path is absolute or relative
	var absPath string
	var err error
	if filepath.IsAbs(path) {
		absPath = path
	} else {
		// For relative paths, join with the base directory
		absPath = filepath.Join(baseDir, path)
	}

	// Clean the path to resolve any . or .. components
	absPath = filepath.Clean(absPath)

	// Get absolute path to ensure proper comparison
	absPath, err = filepath.Abs(absPath)
	if err != nil {
		return false, fmt.Sprintf("failed to resolve absolute path: %v", err)
	}

	// Check if the resolved path is within any allowed directory
	for _, allowedDir := range v.config.AllowedDirectories {
		// Get absolute path of allowed directory for proper comparison
		allowedAbsDir, err := filepath.Abs(allowedDir)
		if err != nil {
			continue // Skip directories that can't be resolved
		}

		// Check if path is within the allowed directory
		if strings.HasPrefix(absPath, allowedAbsDir) {
			return true, ""
		}
	}

	return false, fmt.Sprintf("path %q is outside of allowed directories: %s", path, v.config.DefaultErrorMessage)
}

// isPathLike checks if an argument looks like a file path.
func (v *CommandValidator) isPathLike(arg string) bool {
	// Check if the argument contains path separators or starts with common path prefixes
	return strings.Contains(arg, string(os.PathSeparator)) ||
		strings.Contains(arg, "/") || // For Unix paths
		strings.Contains(arg, "\\") || // For Windows paths
		strings.HasPrefix(arg, "./") ||
		strings.HasPrefix(arg, "../") ||
		strings.HasPrefix(arg, "~") ||
		strings.HasPrefix(arg, ".")
}

// ValidateCommand checks if a command is allowed based on the configuration.
func (v *CommandValidator) ValidateCommand(cmd string, args []string, workDir string) (bool, string) {
	// Special handling for xargs command
	if cmd == "xargs" {
		return v.validateXargsCommand(args, workDir)
	}

	// Check if the command is explicitly denied
	if denied, message := v.isCommandExplicitlyDenied(cmd); denied {
		v.logBlockedCommand(cmd, args, message)
		return false, message
	}

	// Check if the command is explicitly allowed
	for _, allowed := range v.config.AllowCommands {
		if allowed.Command == cmd {
			// If there are no subcommands specified, the command is allowed without restrictions
			if len(allowed.SubCommands) == 0 && len(allowed.DenySubCommands) == 0 {
				// Check path-like arguments even for fully allowed commands
				return v.validatePathArguments(cmd, args, workDir)
			}

			// Check subcommand permissions
			if allowed, message := v.checkSubCommandPermissions(cmd, args, allowed); !allowed {
				return false, message
			}

			// If subcommand is allowed, also validate any path-like arguments
			return v.validatePathArguments(cmd, args, workDir)
		}
	}

	// If command was not found in the allow list, it's denied
	deniedMessage := fmt.Sprintf("command %q is not permitted: %s", cmd, v.config.DefaultErrorMessage)
	v.logBlockedCommand(cmd, args, deniedMessage)
	return false, deniedMessage
}

// validatePathArguments checks if any path-like arguments are within allowed directories.
func (v *CommandValidator) validatePathArguments(cmd string, args []string, workDir string) (bool, string) {
	for _, arg := range args {
		// Skip arguments that don't look like paths or that start with a dash (flags)
		if strings.HasPrefix(arg, "-") || !v.isPathLike(arg) {
			continue
		}

		// Validate the path argument
		allowed, message := v.IsPathInAllowedDirectory(arg, workDir)
		if !allowed {
			v.logBlockedCommand(cmd, args, message)
			return false, message
		}
	}

	return true, ""
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
			deniedMessage := fmt.Sprintf("subcommand %q is not allowed for command %q", args[0], cmd)
			v.logBlockedCommand(cmd, args, deniedMessage)
			return false, deniedMessage
		}
	}

	// If there are denied subcommands specified, check if the first argument matches any of them
	if len(allowed.DenySubCommands) > 0 && len(args) > 0 {
		for _, deniedSubCmd := range allowed.DenySubCommands {
			if args[0] == deniedSubCmd {
				deniedMessage := fmt.Sprintf("subcommand %q is denied for command %q", args[0], cmd)
				v.logBlockedCommand(cmd, args, deniedMessage)
				return false, deniedMessage
			}
		}
	}

	return true, ""
}

// validateXargsCommand checks if the command executed by xargs is allowed.
func (v *CommandValidator) validateXargsCommand(args []string, workDir string) (bool, string) {
	// First check if xargs itself is allowed
	if denied, message := v.isCommandExplicitlyDenied("xargs"); denied {
		v.logBlockedCommand("xargs", args, message)
		return false, message
	}

	// Check if xargs is explicitly allowed
	if !v.config.IsCommandAllowed("xargs") {
		deniedMessage := fmt.Sprintf("command %q is not permitted: %s", "xargs", v.config.DefaultErrorMessage)
		v.logBlockedCommand("xargs", args, deniedMessage)
		return false, deniedMessage
	}

	// Parse the xargs command to extract the actual command
	parser := NewXargsParser()
	xargsCmd, xargsArgs, valid, errMsg := parser.ParseXargsCommand(args)

	if !valid {
		v.logBlockedCommand("xargs", args, errMsg)
		return false, errMsg
	}

	// Now validate the command that xargs will execute
	allowed, message := v.ValidateCommand(xargsCmd, xargsArgs, workDir)
	if !allowed {
		// Add context that this is from an xargs command
		message = "xargs would execute disallowed command: " + message
		v.logBlockedCommand("xargs", args, message)
		return false, message
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
