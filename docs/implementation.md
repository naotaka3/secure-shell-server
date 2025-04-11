# Secure Shell Server - Implementation Guide

## Architecture

The Secure Shell Server is built with a modular architecture following Go's best practices. The main components are:

1. **Config Package**: Manages the configuration settings including allowed commands, environment variables, and execution parameters.

2. **Validator Package**: Parses and validates shell commands against the allowed commands list using the `mvdan.cc/sh/v3` package.

3. **Runner Package**: Executes validated commands in a secure manner, enforcing restrictions and timeouts.

4. **Logger Package**: Provides structured logging for auditing and debugging purposes.

## Component Details

### Config Package

The `config` package defines the `ShellConfig` struct which holds all configuration settings:

```go
type ShellConfig struct {
	AllowedCommands map[string]bool
	RestrictedEnv   map[string]string
	WorkingDir      string
	MaxExecutionTime int
}
```

This package provides methods to manage the configuration, such as adding or removing allowed commands and setting environment variables.

### Validator Package

The `validator` package uses the `mvdan.cc/sh/v3/syntax` package to parse shell scripts and validate them against the allowed commands list:

```go
func (v *CommandValidator) ValidateScript(script string) (bool, error) {
	parser := syntax.NewParser()
	prog, err := parser.Parse(strings.NewReader(script), "")
	if err != nil {
		return false, fmt.Errorf("parse error: %w", err)
	}

	valid := true
	err = nil

	syntax.Walk(prog, func(node syntax.Node) bool {
		if call, ok := node.(*syntax.CallExpr); ok && len(call.Args) > 0 {
			// Extract the command name
			if word, ok := call.Args[0].(*syntax.Word); ok && len(word.Parts) > 0 {
				if lit, ok := word.Parts[0].(*syntax.Lit); ok {
					cmd := lit.Value
					if !v.config.IsCommandAllowed(cmd) {
						err = fmt.Errorf("command %q is not permitted", cmd)
						valid = false
						return false
					}
				}
			}
		}
		return true
	})

	return valid, err
}
```

The validator traverses the abstract syntax tree (AST) of the script and checks each command against the allowlist.

### Runner Package

The `runner` package implements the `SafeRunner` struct which executes validated commands:

```go
func (r *SafeRunner) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command provided")
	}

	cmd := args[0]
	if !r.config.IsCommandAllowed(cmd) {
		r.logger.LogCommandAttempt(cmd, args[1:], false)
		return fmt.Errorf("command %q is not permitted", cmd)
	}

	// Create a timeout context
	if r.config.MaxExecutionTime > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(r.config.MaxExecutionTime)*time.Second)
		defer cancel()
		ctx = timeoutCtx
	}

	// Execute the command
	command := exec.CommandContext(ctx, cmd, args[1:]...)
	command.Env = buildEnv(r.config.RestrictedEnv)
	command.Dir = r.config.WorkingDir
	command.Stdout = r.stdout
	command.Stderr = r.stderr

	return command.Run()
}
```

The runner enforces security by:
- Verifying commands against the allowlist
- Setting execution timeouts
- Restricting environment variables
- Setting working directory
- Capturing command output

### Logger Package

The `logger` package provides structured logging:

```go
func (l *Logger) LogCommandAttempt(cmd string, args []string, allowed bool) {
	status := "ALLOWED"
	if !allowed {
		status = "BLOCKED"
	}

	timestamp := time.Now().Format(time.RFC3339)
	l.logger.Printf("%s [%s] Command: %s %v\n", timestamp, status, cmd, args)
}
```

This logging system records all command attempts, both allowed and blocked, for security auditing.

## Security Implementation

### Command Validation

The Secure Shell Server uses a multi-layered approach to command validation:

1. **Parsing**: The `mvdan.cc/sh/v3/syntax` package parses shell scripts into ASTs, allowing precise analysis of command structures.

2. **Allowlist Checking**: Each command is checked against the allowlist before execution.

3. **Execution Control**: Commands are executed using Go's `os/exec` package with context-based timeouts and environment restrictions.

### Environment Restrictions

The Secure Shell Server restricts the execution environment in multiple ways:

1. **Limited Environment Variables**: Only specified environment variables are passed to commands.

2. **Working Directory Restriction**: Commands are executed in a specified working directory.

3. **Execution Timeouts**: Commands are terminated if they exceed the configured timeout.

## Design Patterns

The Secure Shell Server implements several design patterns:

1. **Dependency Injection**: Components like `config`, `validator`, and `logger` are injected into the `runner`, making the code more testable and maintainable.

2. **Builder Pattern**: The configuration can be built incrementally using methods like `AddAllowedCommand` and `SetEnv`.

3. **Strategy Pattern**: Different validation and execution strategies can be implemented by modifying the validator and runner components.

## Testing Strategy

The Secure Shell Server includes comprehensive tests:

1. **Unit Tests**: Each package has unit tests that verify individual component functionality.

2. **Integration Tests**: The cmd/secure-shell package includes integration tests that verify the complete command execution flow.

3. **Security Tests**: Specific tests ensure that disallowed commands are properly blocked.

## Extension Points

The Secure Shell Server is designed to be extensible:

1. **Custom Validators**: You can implement custom validation logic by modifying the `validator` package.

2. **Alternative Runners**: You can implement different execution strategies by modifying the `runner` package.

3. **Enhanced Logging**: You can extend the `logger` package to support additional logging formats or destinations.

## Performance Considerations

1. **Parsing Overhead**: The parsing and validation of shell scripts adds some overhead compared to direct command execution.

2. **Memory Usage**: The AST representation of large scripts may consume significant memory.

3. **Execution Context**: Using context-based timeouts ensures that hung commands don't consume resources indefinitely.

## Concurrency Model

The Secure Shell Server supports concurrent execution with some considerations:

1. **Thread Safety**: The core components (`config`, `validator`, `runner`, `logger`) are designed to be thread-safe.

2. **Context Usage**: All execution methods accept a context parameter, allowing proper cancellation and timeout handling.

## Future Enhancements

1. **Advanced AST Analysis**: Enhance the validator to detect and block more sophisticated harmful patterns.

2. **Sandboxing**: Integrate with container technologies to provide additional isolation.

3. **Network Controls**: Add controls for network access during command execution.

4. **Resource Limits**: Implement resource usage limits (CPU, memory) for executed commands.

## Libraries and Dependencies

1. **mvdan.cc/sh/v3**: Used for parsing and executing shell scripts.

2. **context**: Used for timeout and cancellation support.

3. **os/exec**: Used for command execution.

## Contributing Code

When contributing to the Secure Shell Server, please follow these guidelines:

1. Follow the Go coding standards and project-specific conventions.

2. Include unit tests for all new functionality.

3. Ensure that security is a primary consideration in all changes.

4. Document all public APIs and significant implementation details.

## Deployment Considerations

When deploying the Secure Shell Server, consider:

1. **Permission Model**: Run the server with the minimum necessary system permissions.

2. **Logging Configuration**: Ensure logs are written to a secure, persistent storage.

3. **Configuration Management**: Use secure methods to manage and deploy the allowed commands list.

4. **Monitoring**: Set up monitoring for blocked command attempts.