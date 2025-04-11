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
	var exitCode int = run()
	os.Exit(exitCode)
}

func run() int {
	// Define command-line flags
	cmdToExec := flag.String("cmd", "", "Command to execute")
	scriptFile := flag.String("file", "", "Script file to execute")
	scriptStr := flag.String("script", "", "Script string to execute")
	allowedCommands := flag.String("allow", "ls,echo,cat", "Comma-separated list of allowed commands")
	maxTime := flag.Int("timeout", config.DefaultExecutionTimeout, "Maximum execution time in seconds")
	workingDir := flag.String("dir", "", "Working directory for command execution")

	flag.Parse()

	// Create logger
	log := logger.New()

	// Create config with allowed commands
	cfg := &config.ShellConfig{
		AllowedCommands: make(map[string]bool),
		RestrictedEnv: map[string]string{
			"PATH": "/usr/bin:/bin",
		},
		WorkingDir:       *workingDir,
		MaxExecutionTime: *maxTime,
	}

	// Parse and add allowed commands
	for _, cmd := range strings.Split(*allowedCommands, ",") {
		cmd = strings.TrimSpace(cmd)
		if cmd != "" {
			cfg.AddAllowedCommand(cmd)
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
	case *cmdToExec != "":
		// Execute a single command
		args := strings.Fields(*cmdToExec)
		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, "Error: Empty command provided\n")
			return 1
		}

		err = safeRunner.Run(ctx, args)

	case *scriptFile != "":
		// Execute a script file
		var file *os.File
		var fileErr error
		file, fileErr = os.Open(*scriptFile)
		if fileErr != nil {
			fmt.Fprintf(os.Stderr, "Error opening script file: %v\n", fileErr)
			return 1
		}
		// Close the file when we're done
		defer file.Close()

		err = safeRunner.RunScriptFile(ctx, file)

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
