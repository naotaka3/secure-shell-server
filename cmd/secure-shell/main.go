package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/pkg/logger"
	"github.com/shimizu1995/secure-shell-server/pkg/runner"
	"github.com/shimizu1995/secure-shell-server/pkg/validator"
)

func main() {
	exitCode := run()
	os.Exit(exitCode)
}

func run() int {
	// Define command-line flags
	scriptStr := flag.String("script", "", "Script string to execute")
	allowedCommands := flag.String("allow", "ls,echo,cat", "Comma-separated list of allowed commands")
	maxTime := flag.Int("timeout", config.DefaultExecutionTimeout, "Maximum execution time in seconds")
	workingDir := flag.String("dir", "", "Working directory for command execution")

	flag.Parse()

	// Create logger
	log := logger.New()

	// Create config with allowed commands
	cfg := config.NewDefaultConfig()
	cfg.WorkingDir = *workingDir
	cfg.MaxExecutionTime = *maxTime

	// Clear the default allowed commands and add the ones from command line
	cfg.AllowCommands = nil

	// Parse and add allowed commands
	for _, cmd := range strings.Split(*allowedCommands, ",") {
		cmd = strings.TrimSpace(cmd)
		if cmd != "" {
			cfg.AllowCommands = append(cfg.AllowCommands, config.AllowCommand{Command: cmd})
		}
	}

	// Create validator and runner
	validatorObj := validator.New(cfg, log)
	safeRunner := runner.New(cfg, validatorObj, log)

	// Create a context with timeout for the entire execution
	ctx := context.Background()
	var cancel context.CancelFunc
	if *maxTime > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(*maxTime)*time.Second)
		defer cancel()
	}

	// Execute the requested operation
	var err error

	switch {
	case *scriptStr != "":
		// Execute a script string
		err = safeRunner.RunScript(ctx, *scriptStr)

	default:
		fmt.Fprintf(os.Stderr, "Error: No command or script specified\n")
		flag.Usage()
		return 1
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	return 0
}
