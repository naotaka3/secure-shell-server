# Secure Shell Server Usage Guide

## Overview

The Secure Shell Server is a tool for executing shell commands securely. It ensures that only permitted commands are executed, helping to prevent unauthorized or potentially harmful operations.

## Key Features

- **Command Validation**: Only allows execution of commands in a predefined allowlist.
- **Script Validation**: Validates shell scripts before execution to ensure they only use allowed commands.
- **Environment Restrictions**: Limits access to system resources and environment variables.
- **Execution Timeouts**: Enforces time limits on command execution.
- **Detailed Logging**: Logs all command attempts for auditing and security purposes.

## Basic Usage

### Running a Simple Command

```bash
secure-shell-server run -cmd="ls -l"
```

This executes the `ls -l` command if `ls` is in the allowed commands list.

### Running a Script from a String

```bash
secure-shell-server run -script="echo 'Hello, World!'; ls -l"
```

This parses and validates the script, then executes it if all commands are allowed.

### Running a Script from a File

```bash
secure-shell-server run -file="/path/to/script.sh"
```

This reads, parses, and validates the script file, then executes it if all commands are allowed.

## Advanced Usage

### Custom Allowed Commands

By default, only a few basic commands like `ls`, `echo`, and `cat` are allowed. You can specify a custom list of allowed commands:

```bash
secure-shell-server run -cmd="grep pattern file.txt" -allow="ls,echo,cat,grep"
```

### Setting a Working Directory

You can specify a working directory for command execution:

```bash
secure-shell-server run -cmd="ls -l" -dir="/safe/directory"
```

### Execution Timeout

You can set a maximum execution time (in seconds) for commands:

```bash
secure-shell-server run -cmd="find / -type f -name '*.log'" -timeout=60
```

## Command-Line Options

- `-cmd="<command>"`: Command to execute.
- `-script="<script>"`: Script string to execute.
- `-file="<path>"`: Script file to execute.
- `-allow="cmd1,cmd2,..."`: Comma-separated list of allowed commands (default: "ls,echo,cat").
- `-timeout=<seconds>`: Maximum execution time in seconds (default: 30).
- `-dir="<path>"`: Working directory for command execution.

## Security Considerations

- Always run the Secure Shell Server with the minimum set of allowed commands necessary for your task.
- Use the working directory option to restrict file system access.
- Set appropriate timeouts to prevent long-running commands from consuming excessive resources.
- Review logs regularly to monitor command attempts.

## Error Handling

The Secure Shell Server provides clear error messages for various failure scenarios:

- **Disallowed Command**: "Error: command 'rm' is not permitted"
- **Syntax Error**: "Error: parse error: syntax error"
- **Timeout Error**: "Error: command execution error: context deadline exceeded"

## Integration with Other Tools

The Secure Shell Server can be integrated with other tools and systems:

### API Usage

You can also use the Secure Shell Server as a library in your Go applications:

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/pkg/logger"
	"github.com/shimizu1995/secure-shell-server/pkg/runner"
	"github.com/shimizu1995/secure-shell-server/pkg/validator"
)

func main() {
	// Create configuration
	cfg := config.NewDefaultConfig()

	// Create logger
	log := logger.New()

	// Create validator and runner
	validatorObj := validator.New(cfg, log)
	safeRunner := runner.New(cfg, validatorObj, log)

	// Execute a command
	args := []string{"ls", "-l"}
	err := safeRunner.Run(context.Background(), args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

### Web Service Integration

You can wrap the Secure Shell Server in a web service to provide secure command execution via HTTP endpoints:

```go
func handleCommandExecution(w http.ResponseWriter, r *http.Request) {
	// Parse request parameters
	cmd := r.FormValue("command")
	
	// Create secure runner components
	cfg := config.NewDefaultConfig()
	log := logger.New()
	validatorObj := validator.New(cfg, log)
	safeRunner := runner.New(cfg, validatorObj, log)
	
	// Capture command output
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	safeRunner.SetOutputs(stdout, stderr)
	
	// Execute the command
	args := strings.Fields(cmd)
	err := safeRunner.Run(r.Context(), args)
	
	// Return results
	response := map[string]interface{}{
		"stdout": stdout.String(),
		"stderr": stderr.String(),
		"error":  err != nil,
	}
	
	json.NewEncoder(w).Encode(response)
}
```

## Examples

### Example 1: Executing a File Listing

```bash
secure-shell-server run -cmd="ls -l /var/log"
```

If `ls` is in the allowed commands list, this will list the contents of `/var/log` with detailed information.

### Example 2: Running a Simple Script

```bash
secure-shell-server run -script="echo 'Starting script'; ls -l /var/log; echo 'Script completed'"
```

This will output:

```
Starting script
[listing of /var/log contents]
Script completed
```

### Example 3: Reading a File with Cat

```bash
secure-shell-server run -cmd="cat /etc/hostname"
```

If `cat` is in the allowed commands list, this will display the contents of `/etc/hostname`.

## Best Practices

1. **Principle of Least Privilege**: Only include commands in the allowlist that are absolutely necessary for the task at hand.

2. **Regular Auditing**: Regularly review command logs to identify potentially suspicious activities.

3. **Environment Isolation**: Run the Secure Shell Server in an isolated environment (e.g., a container) for an additional layer of security.

4. **Secure Configuration**: Store your allowed command lists in secure configuration files and load them at runtime rather than passing them as command-line arguments.

5. **Input Validation**: When integrating with other systems, ensure all input is properly validated before being passed to the Secure Shell Server.

## Troubleshooting

### Common Issues

1. **Command Not Found**: Ensure the command exists in the system's PATH environment variable.

2. **Permission Denied**: Check that the Secure Shell Server has appropriate permissions to execute the command.

3. **Command Not Allowed**: Verify that the command is included in the allowlist.

4. **Syntax Error**: Check the script for syntax errors, particularly unclosed quotes or missing semicolons.

5. **Timeout Error**: Increase the timeout value for long-running commands.

## Contributing

Contributions to the Secure Shell Server are welcome! Please see the CONTRIBUTING.md file for guidelines.

## License

This project is licensed under the terms found in the LICENSE file.