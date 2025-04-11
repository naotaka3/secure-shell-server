package validator

import (
	"fmt"
	"io"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/pkg/logger"
)

// CommandValidator validates shell commands.
type CommandValidator struct {
	config *config.ShellConfig
	logger *logger.Logger
}

// New creates a new CommandValidator.
func New(config *config.ShellConfig, logger *logger.Logger) *CommandValidator {
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
					if !v.config.IsCommandAllowed(cmd) {
						validationErr = fmt.Errorf("command %q is not permitted", cmd)
						valid = false
						v.logger.LogCommandAttempt(cmd, extractArgs(call.Args), false)
						return false
					}
					v.logger.LogCommandAttempt(cmd, extractArgs(call.Args), true)
				}
			}
		}
		return true
	})

	return valid, validationErr
}

// ValidateScriptFile validates a shell script file.
func (v *CommandValidator) ValidateScriptFile(r io.Reader) (bool, error) {
	parser := syntax.NewParser()
	prog, err := parser.Parse(r, "")
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
					if !v.config.IsCommandAllowed(cmd) {
						validationErr = fmt.Errorf("command %q is not permitted", cmd)
						valid = false
						v.logger.LogCommandAttempt(cmd, extractArgs(call.Args), false)
						return false
					}
					v.logger.LogCommandAttempt(cmd, extractArgs(call.Args), true)
				}
			}
		}
		return true
	})

	return valid, validationErr
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
