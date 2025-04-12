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

	// Execute a script
	script := "echo 'Hello, World!'; ls -l"
	fmt.Printf("\nExecuting script: %s\n", script)
	err := safeRunner.RunCommand(context.Background(), script)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Try to execute a disallowed command
	disallowedScript := "rm -rf /tmp/test"
	fmt.Printf("\nAttempting to execute disallowed script: %s\n", disallowedScript)
	err = safeRunner.RunCommand(context.Background(), disallowedScript)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	} else {
		fmt.Println("Error: Command was allowed but should have been blocked")
		os.Exit(1)
	}
}
