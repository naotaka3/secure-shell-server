# Secure Shell Server Examples

This directory contains examples demonstrating how to use the Secure Shell Server.

## Example Scripts

### example.sh

A basic shell script containing both allowed and disallowed commands. This script can be used to demonstrate how the Secure Shell Server validates and executes commands.

#### Contents

- Echo commands (allowed by default)
- List files with ls (allowed by default)
- Display file content with cat (allowed by default)
- Remove a file with rm (should be blocked by default)

#### Running the Example

```bash
# From the project root directory
$ secure-shell-server run -file="examples/example.sh"
```

Expected output:

```
Starting script execution...
Files in the current directory:
[ls output]
Content of example.sh:
[content of the script]
Attempting to execute 'rm' command (should be blocked):
Error: command "rm" is not permitted
```

Notice that the script execution stops when it encounters the disallowed `rm` command.

### Running with Custom Allowed Commands

To allow the `rm` command in the example script:

```bash
$ secure-shell-server run -file="examples/example.sh" -allow="ls,echo,cat,rm"
```

Now the entire script should execute without errors.

## API Usage Example

The `examples` directory also demonstrates how to use the Secure Shell Server as a library in your Go code.

### api_example.go

This example shows how to integrate the Secure Shell Server into your Go applications:

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

	// Add additional allowed commands if needed
	cfg.AddAllowedCommand("grep")

	// Create logger
	log := logger.New()

	// Create validator and runner
	validatorObj := validator.New(cfg, log)
	safeRunner := runner.New(cfg, validatorObj, log)

	// Execute a command
	args := []string{"ls", "-l"}
	fmt.Printf("Executing command: %v\n", args)
	err := safeRunner.Run(context.Background(), args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Execute a script
	script := "echo 'Hello, World!'; ls -l"
	fmt.Printf("\nExecuting script: %s\n", script)
	err = safeRunner.RunScript(context.Background(), script)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Try to execute a disallowed command
	disallowedArgs := []string{"rm", "-rf", "/tmp/test"}
	fmt.Printf("\nAttempting to execute disallowed command: %v\n", disallowedArgs)
	err = safeRunner.Run(context.Background(), disallowedArgs)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	} else {
		fmt.Println("Error: Command was allowed but should have been blocked")
		os.Exit(1)
	}
}
```

## Additional Examples

You can create additional example scripts in this directory to test different command combinations and security scenarios.