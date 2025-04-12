// Secure Shell Server - A tool to execute shell commands securely.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/pkg/logger"
	"github.com/shimizu1995/secure-shell-server/pkg/runner"
	"github.com/shimizu1995/secure-shell-server/pkg/validator"
	"github.com/shimizu1995/secure-shell-server/service"
)

func main() {
	exitCode := run()
	os.Exit(exitCode)
}

func run() int {
	// Print usage information by default
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Secure Shell Server - A tool to execute shell commands securely.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [command]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  help        Display this help message\n")
		fmt.Fprintf(os.Stderr, "  version     Display version information\n")
		fmt.Fprintf(os.Stderr, "  run         Run the secure shell command executor\n")
		fmt.Fprintf(os.Stderr, "  server      Start the MCP server\n\n")
		fmt.Fprintf(os.Stderr, "For more information, use '%s [command] --help'\n", os.Args[0])
	}

	// Minimum number of arguments required
	const minArgs = 2
	if len(os.Args) < minArgs {
		flag.Usage()
		return 1
	}

	// Process commands
	switch os.Args[1] {
	case "version":
		fmt.Println("Secure Shell Server v0.1.0")
		return 0

	case "help":
		flag.Usage()
		return 0

	case "run":
		// Use the same signature as cmd/secure-shell/main.go
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...)

		// Define command-line flags
		cmdToExec := flag.String("cmd", "", "Command to execute")
		scriptFile := flag.String("file", "", "Script file to execute")
		scriptStr := flag.String("script", "", "Script string to execute")
		allowedCommands := flag.String("allow", "ls,echo,cat", "Comma-separated list of allowed commands")
		maxTime := flag.Int("timeout", config.DefaultExecutionTimeout, "Maximum execution time in seconds")
		workingDir := flag.String("dir", "", "Working directory for command execution")

		// Parse flags
		flag.Parse()

		// Configure and run the secure shell
		return runSecureShell(cmdToExec, scriptFile, scriptStr, allowedCommands, maxTime, workingDir)

	case "server":
		return runServer()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		flag.Usage()
		return 1
	}
}

// runServer handles the server subcommand.
func runServer() int {
	// Use a different flag set to avoid conflicts with the run command
	serverFlagSet := flag.NewFlagSet("server", flag.ExitOnError)

	// Default port for HTTP server
	const defaultPort = 8080

	// Define server-specific flags
	port := serverFlagSet.Int("port", defaultPort, "Port to listen on")
	configFile := serverFlagSet.String("config", "", "Path to configuration file")
	stdio := serverFlagSet.Bool("stdio", false, "Use stdin/stdout for MCP communication")

	// Parse the server flags
	if err := serverFlagSet.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		return 1
	}

	// Get configuration
	var cfg *config.ShellCommandConfig
	var err error

	if *configFile != "" {
		// Load configuration from file
		cfg, err = config.LoadConfigFromFile(*configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
			return 1
		}
	} else {
		// Use default configuration
		cfg = config.NewDefaultConfig()
	}

	// Create server
	mcpServer := service.NewServer(cfg, *port)

	// Start the server using stdio or HTTP
	if *stdio {
		fmt.Println("Starting MCP server using stdin/stdout...")
		if err := mcpServer.ServeStdio(); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			return 1
		}
	} else {
		fmt.Printf("Starting MCP server on port %d...\n", *port)
		if err := mcpServer.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			return 1
		}
	}

	return 0
}

func runSecureShell(cmdToExec, scriptFile, scriptStr, _ *string, maxTime *int, workingDir *string) int {
	// Create a configuration object
	cfg := config.NewDefaultConfig()
	cfg.WorkingDir = *workingDir
	cfg.MaxExecutionTime = *maxTime

	// Create logger
	log := logger.New()

	// Create validator and runner
	validatorObj := validator.New(cfg, log)
	safeRunner := runner.New(cfg, validatorObj, log)

	// Execute the requested operation
	var err error

	switch {
	case *cmdToExec != "":
		// Execute a single command
		args := flag.Args()
		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, "Error: Empty command provided\n")
			return 1
		}

		err = safeRunner.Run(context.Background(), args)

	case *scriptFile != "":
		// Execute a script file
		var file *os.File
		var fileErr error
		file, fileErr = os.Open(*scriptFile)
		if fileErr != nil {
			fmt.Fprintf(os.Stderr, "Error opening script file: %v\n", fileErr)
			return 1
		}
		defer file.Close()

		err = safeRunner.RunScriptFile(context.Background(), file)

	case *scriptStr != "":
		// Execute a script string
		err = safeRunner.RunScript(context.Background(), *scriptStr)

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
