package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/pkg/logger"
	"github.com/shimizu1995/secure-shell-server/pkg/runner"
	"github.com/shimizu1995/secure-shell-server/pkg/validator"
)

// createRunTool creates the run tool for executing shell commands.
func createRunTool() mcp.Tool {
	return mcp.NewTool("run",
		mcp.WithDescription("Run a shell command in the current working directory.\n"+
			"Use change_directory to set the working directory before running commands."),
		mcp.WithString("command",
			mcp.Required(),
			mcp.Description("Command to execute"),
		),
	)
}

// createChangeDirectoryTool creates the change_directory tool for setting the working directory.
func createChangeDirectoryTool() mcp.Tool {
	return mcp.NewTool("change_directory",
		mcp.WithDescription("Set the working directory for subsequent commands.\n"+
			"Must be called before running any commands. The directory must be within allowed paths."),
		mcp.WithString("directory",
			mcp.Required(),
			mcp.Description("The directory to set as the working directory."),
		),
	)
}

// Server is the MCP server for secure shell execution.
type Server struct {
	config    *config.ShellCommandConfig
	validator *validator.CommandValidator
	runner    *runner.SafeRunner
	logger    *logger.Logger
	mcpServer *server.MCPServer
	port      int
	// Mutex to protect shared resources (config, runner, validator) during command execution
	cmdMutex sync.Mutex
	// workingDir holds the session's current working directory. Empty means not yet set.
	workingDir string
}

// NewServer creates a new MCP server instance.
func NewServer(cfg *config.ShellCommandConfig, port int, logPath string) (*Server, error) {
	// Create logger with optional path
	loggerObj, err := logger.NewWithPath(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	validatorObj := validator.New(cfg, loggerObj)
	runnerObj := runner.New(cfg, validatorObj, loggerObj)

	mcpServer := server.NewMCPServer(
		"Secure Shell Server",
		"1.0.0",
		server.WithLogging(),
		server.WithRecovery(),
	)

	return &Server{
		config:    cfg,
		validator: validatorObj,
		runner:    runnerObj,
		logger:    loggerObj,
		mcpServer: mcpServer,
		port:      port,
	}, nil
}

// Start initializes and starts the MCP server.
func (s *Server) Start() error {
	// Register tools
	s.mcpServer.AddTool(createRunTool(), s.HandleRunCommand)
	s.mcpServer.AddTool(createChangeDirectoryTool(), s.HandleChangeDirectory)

	// Start the server
	address := fmt.Sprintf(":%d", s.port)
	s.logger.LogInfof("Starting MCP server on %s", address)

	// Create HTTP server to serve the MCP server
	handler := http.NewServeMux()
	handler.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// TODO: Implement proper HTTP handler for MCP
		_, err := w.Write([]byte("MCP server running"))
		if err != nil {
			s.logger.LogErrorf("Failed to write response: %v", err)
		}
	}))

	// Timeout constants
	const (
		readTimeoutSeconds  = 10
		writeTimeoutSeconds = 10
	)

	// Create a server with timeouts
	server := &http.Server{
		Addr:         address,
		Handler:      handler,
		ReadTimeout:  readTimeoutSeconds * time.Second,
		WriteTimeout: writeTimeoutSeconds * time.Second,
	}

	return server.ListenAndServe()
}

// HandleChangeDirectory handles the change_directory tool execution.
func (s *Server) HandleChangeDirectory(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	directory, ok := request.Params.Arguments["directory"].(string)
	if !ok || directory == "" {
		return mcp.NewToolResultError("Directory parameter must be a non-empty string"), nil
	}

	absDir, err := filepath.Abs(directory)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to resolve directory path: %v", err)), nil
	}

	// Validate against allowed directories
	allowed, msg := s.validator.IsDirectoryAllowed(absDir)
	if !allowed {
		return mcp.NewToolResultError(msg), nil
	}

	// Verify directory exists
	info, err := os.Stat(absDir)
	if err != nil || !info.IsDir() {
		return mcp.NewToolResultError("Directory does not exist: " + absDir), nil
	}

	s.cmdMutex.Lock()
	s.workingDir = absDir
	s.cmdMutex.Unlock()

	s.logger.LogInfof("Working directory changed to: %s", absDir)
	return mcp.NewToolResultText("Working directory set to: " + absDir), nil
}

// HandleRunCommand handles the run tool execution.
func (s *Server) HandleRunCommand(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	commandStr, ok := request.Params.Arguments["command"].(string)
	if !ok || commandStr == "" {
		return mcp.NewToolResultError("Command parameter must be a non-empty string"), nil
	}

	s.cmdMutex.Lock()
	defer s.cmdMutex.Unlock()

	if s.workingDir == "" {
		return mcp.NewToolResultError("No working directory set. Use change_directory to set a working directory first."), nil
	}

	s.logger.LogInfof("Command attempt: %s in directory: %s", commandStr, s.workingDir)

	outputBuffer := new(strings.Builder)
	s.runner.SetOutputs(outputBuffer, outputBuffer)

	err := s.runner.RunCommand(ctx, commandStr, s.workingDir)
	if err != nil {
		s.logger.LogErrorf("Command execution failed: %v", err)
		output := fmt.Sprintf("Error: %v\n%s", err, outputBuffer.String())
		return mcp.NewToolResultError(output), nil
	}

	return mcp.NewToolResultText(outputBuffer.String()), nil
}

// ServeStdio starts an MCP server using stdin/stdout for communication.
func (s *Server) ServeStdio() error {
	// Register tools
	s.mcpServer.AddTool(createRunTool(), s.HandleRunCommand)
	s.mcpServer.AddTool(createChangeDirectoryTool(), s.HandleChangeDirectory)

	// Start the server using stdio
	s.logger.LogInfof("Starting MCP server using stdin/stdout")
	return server.ServeStdio(s.mcpServer)
}
