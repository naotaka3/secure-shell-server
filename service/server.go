package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/pkg/logger"
	"github.com/shimizu1995/secure-shell-server/pkg/runner"
	"github.com/shimizu1995/secure-shell-server/pkg/validator"
)

// Server is the MCP server for secure shell execution.
type Server struct {
	config    *config.ShellCommandConfig
	validator *validator.CommandValidator
	runner    *runner.SafeRunner
	logger    *logger.Logger
	mcpServer *server.MCPServer
	port      int
}

// NewServer creates a new MCP server instance.
func NewServer(cfg *config.ShellCommandConfig, port int) *Server {
	loggerObj := logger.New()
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
	}
}

// Start initializes and starts the MCP server.
func (s *Server) Start() error {
	// Register the run_command tool
	runCommandTool := mcp.NewTool("run_command",
		mcp.WithDescription("Run shell commands in specific directories (only within allowed paths).\nThe \"directory\" parameter sets the working directory automatically; \"cd\" command isn't needed."),
		mcp.WithString("command",
			mcp.Required(),
			mcp.Description("Command to execute"),
		),
		mcp.WithString("directory",
			mcp.Required(),
			mcp.Description("Working directory to execute the command in."),
		),
	)

	// Add the tool handler
	s.mcpServer.AddTool(runCommandTool, s.handleRunCommand)

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

// handleRunCommand handles the run_command tool execution.
func (s *Server) handleRunCommand(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	commandStr, ok := request.Params.Arguments["command"].(string)
	if !ok || commandStr == "" {
		return mcp.NewToolResultError("Command parameter must be a non-empty string"), nil
	}

	directory, ok := request.Params.Arguments["directory"].(string)
	if !ok || directory == "" {
		return mcp.NewToolResultError("Directory parameter must be a non-empty string"), nil
	}

	// Log the command attempt
	s.logger.LogInfof("Command attempt: %s in directory: %s", commandStr, directory)

	// Parse the command into args
	args, err := parseCommand(commandStr)
	if err != nil {
		s.logger.LogErrorf("Failed to parse command: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse command: %v", err)), nil
	}

	if len(args) == 0 {
		return mcp.NewToolResultError("No command provided"), nil
	}

	// Validate the command and directory
	allowed, err := s.validator.ValidateCommandInDirectory(args[0], args[1:], directory)
	if !allowed || err != nil {
		s.logger.LogErrorf("Command validation failed: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("Command validation failed: %v", err)), nil
	}

	// Create a buffer to capture the output
	outputBuffer := new(strings.Builder)

	// Set the working directory in the config
	cfgCopy := *s.config
	cfgCopy.WorkingDir = directory

	// Create a new runner with the updated config and output capture
	tempValidator := validator.New(&cfgCopy, s.logger)
	tempRunner := runner.New(&cfgCopy, tempValidator, s.logger)
	tempRunner.SetOutputs(outputBuffer, outputBuffer)

	// Execute the command
	err = tempRunner.Run(ctx, args)
	if err != nil {
		s.logger.LogErrorf("Command execution failed: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("Command execution failed: %v", err)), nil
	}

	// Return the command output
	return mcp.NewToolResultText(outputBuffer.String()), nil
}

// parseCommand splits a command string into arguments.
func parseCommand(cmd string) ([]string, error) {
	// Return early for empty commands
	if strings.TrimSpace(cmd) == "" {
		return nil, errors.New("empty command")
	}

	// Simple splitting by space for now
	// This could be enhanced to handle quotes and other special characters
	return strings.Fields(cmd), nil
}

// TestHandleRunCommand is a wrapper for handleRunCommand for testing.
func (s *Server) TestHandleRunCommand(ctx context.Context, cmd string, dir string) (*mcp.CallToolResult, error) {
	// Create a mock request with the necessary structure
	request := mcp.CallToolRequest{}
	// Set the arguments directly
	request.Params.Arguments = map[string]interface{}{
		"command":   cmd,
		"directory": dir,
	}

	result, err := s.handleRunCommand(ctx, request)
	// Convert err to a string for testing
	if err != nil {
		return nil, err
	}

	// For testing, convert the result to a simpler structure
	return result, nil
}

// TestParseCommand is a wrapper for parseCommand for testing.
func TestParseCommand(cmd string) ([]string, error) {
	return parseCommand(cmd)
}

// ServeStdio starts an MCP server using stdin/stdout for communication.
func (s *Server) ServeStdio() error {
	// Register the run_command tool
	runCommandTool := mcp.NewTool("run_command",
		mcp.WithDescription("Run shell commands in specific directories (only within allowed paths).\nThe \"directory\" parameter sets the working directory automatically; \"cd\" command isn't needed."),
		mcp.WithString("command",
			mcp.Required(),
			mcp.Description("Command to execute"),
		),
		mcp.WithString("directory",
			mcp.Required(),
			mcp.Description("Working directory to execute the command in."),
		),
	)

	// Add the tool handler
	s.mcpServer.AddTool(runCommandTool, s.handleRunCommand)

	// Start the server using stdio
	s.logger.LogInfof("Starting MCP server using stdin/stdout")
	return server.ServeStdio(s.mcpServer)
}
