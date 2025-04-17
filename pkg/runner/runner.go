package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/pkg/limiter"
	"github.com/shimizu1995/secure-shell-server/pkg/logger"
	"github.com/shimizu1995/secure-shell-server/pkg/validator"
)

// SafeRunner executes shell commands securely.
type SafeRunner struct {
	config    *config.ShellCommandConfig
	validator *validator.CommandValidator
	logger    *logger.Logger
	stdout    io.Writer
	stderr    io.Writer
	// Output limiters to track truncation
	stdoutLimiter *limiter.OutputLimiter
	stderrLimiter *limiter.OutputLimiter
}

// New creates a new SafeRunner.
func New(config *config.ShellCommandConfig, validator *validator.CommandValidator, logger *logger.Logger) *SafeRunner {
	return &SafeRunner{
		config:        config,
		validator:     validator,
		logger:        logger,
		stdout:        os.Stdout,
		stderr:        os.Stderr,
		stdoutLimiter: nil,
		stderrLimiter: nil,
	}
}

// SetOutputs sets the stdout and stderr writers.
func (r *SafeRunner) SetOutputs(stdout, stderr io.Writer) {
	// If MaxOutputSize is set, wrap the writers with limiters
	if r.config.MaxOutputSize > 0 {
		r.stdoutLimiter = limiter.NewOutputLimiter(stdout, r.config.MaxOutputSize)
		r.stderrLimiter = limiter.NewOutputLimiter(stderr, r.config.MaxOutputSize)
		r.stdout = r.stdoutLimiter
		r.stderr = r.stderrLimiter
	} else {
		// Use the writers directly if no limit is set
		r.stdout = stdout
		r.stderr = stderr
		r.stdoutLimiter = nil
		r.stderrLimiter = nil
	}
}

// RunCommand runs a shell command in the specified working directory.
// It enforces security constraints by validating commands and file access.
// WasOutputTruncated returns whether stdout or stderr was truncated due to size limits.
func (r *SafeRunner) WasOutputTruncated() bool {
	if r.stdoutLimiter != nil && r.stdoutLimiter.WasTruncated() {
		return true
	}
	if r.stderrLimiter != nil && r.stderrLimiter.WasTruncated() {
		return true
	}
	return false
}

// GetTruncationStatus returns detailed information about which outputs were truncated.
func (r *SafeRunner) GetTruncationStatus() (stdoutTruncated bool, stderrTruncated bool) {
	stdoutTruncated = r.stdoutLimiter != nil && r.stdoutLimiter.WasTruncated()
	stderrTruncated = r.stderrLimiter != nil && r.stderrLimiter.WasTruncated()
	return
}

// GetTruncationDetails returns detailed information about truncation, including which
// outputs were truncated and how many bytes remained unwritten for each.
func (r *SafeRunner) GetTruncationDetails() (stdoutTruncated bool, stderrTruncated bool, stdoutRemainingBytes int, stderrRemainingBytes int) {
	stdoutTruncated = r.stdoutLimiter != nil && r.stdoutLimiter.WasTruncated()
	stderrTruncated = r.stderrLimiter != nil && r.stderrLimiter.WasTruncated()

	stdoutRemainingBytes = 0
	if stdoutTruncated {
		stdoutRemainingBytes = r.stdoutLimiter.GetRemainingBytes()
	}

	stderrRemainingBytes = 0
	if stderrTruncated {
		stderrRemainingBytes = r.stderrLimiter.GetRemainingBytes()
	}

	return
}

// RunCommand runs a shell command in the specified working directory.
// It enforces security constraints by validating commands and file access.
func (r *SafeRunner) RunCommand(ctx context.Context, command string, workingDir string) error {
	// Get absolute path of the working directory
	absWorkingDir, err := filepath.Abs(workingDir)
	if err != nil {
		r.logger.LogErrorf("Failed to get absolute path for working directory: %v", err)
		return fmt.Errorf("failed to get absolute path for working directory: %w", err)
	}

	// Validate that the working directory is allowed
	dirAllowed, dirMessage := r.validator.IsDirectoryAllowed(absWorkingDir)
	if !dirAllowed {
		r.logger.LogErrorf("Directory validation failed: %s", dirMessage)
		return fmt.Errorf("directory validation failed: %s", dirMessage)
	}

	// Parse the command
	parser := syntax.NewParser()
	prog, err := parser.Parse(strings.NewReader(command), "")
	if err != nil {
		r.logger.LogErrorf("Parse error: %v", err)
		return fmt.Errorf("parse error: %w", err)
	}

	// Create a timeout context if MaxExecutionTime is set
	if r.config.MaxExecutionTime > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(r.config.MaxExecutionTime)*time.Second)
		defer cancel()
		ctx = timeoutCtx
	}

	callFunc := func(_ context.Context, args []string) ([]string, error) {
		cmd := args[0]
		allowed, errMsg := r.validator.ValidateCommand(cmd, args[1:], absWorkingDir)
		if !allowed {
			r.logger.LogCommandAttempt(cmd, args[1:], false)
			return args, fmt.Errorf("%s", errMsg)
		}

		r.logger.LogCommandAttempt(cmd, args[1:], true)

		return args, nil
	}

	// Create a custom OpenHandler for security checks
	openHandler := func(ctx context.Context, path string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
		// Get absolute path of the file
		absPath, absErr := filepath.Abs(path)
		if absErr != nil {
			r.logger.LogErrorf("Failed to get absolute path for file %s: %v", path, absErr)
			return nil, &os.PathError{Op: "open", Path: path, Err: absErr}
		}

		// Check if file is in an allowed directory
		// First check if the file's directory is allowed
		fileDir := filepath.Dir(absPath)
		allowed, disallowedMessage := r.validator.IsDirectoryAllowed(fileDir)

		if !allowed {
			r.logger.LogErrorf("File access attempted outside allowed directories: %s", absPath)
			return nil, &os.PathError{
				Op:   "open",
				Path: path,
				Err:  fmt.Errorf("access denied: file is outside allowed directories: %s", disallowedMessage),
			}
		}

		// Delegate to the default open handler
		return interp.DefaultOpenHandler()(ctx, path, flag, perm)
	}

	// Create interpreter
	runner, err := interp.New(
		interp.CallHandler(callFunc),
		interp.StdIO(nil, r.stdout, r.stderr),
		interp.Env(nil),
		interp.Dir(absWorkingDir),
		interp.OpenHandler(openHandler),
	)
	if err != nil {
		r.logger.LogErrorf("Interpreter creation error: %v", err)
		return fmt.Errorf("interpreter creation error: %w", err)
	}

	err = runner.Run(ctx, prog)
	return err
}
