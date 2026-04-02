package service_test

import (
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/service"
)

func TestNewServer(t *testing.T) {
	// Create a test configuration
	cfg := config.NewDefaultConfig()

	// Test with empty log path
	t.Run("with empty log path", func(t *testing.T) {
		server, err := service.NewServer(cfg, 8080, "")
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		if server == nil {
			t.Fatal("Server is nil")
		}
	})

	// Test with valid log path
	t.Run("with valid log path", func(t *testing.T) {
		// Create a temporary log file
		tmpDir := t.TempDir()
		logPath := tmpDir + "/server.log"

		server, err := service.NewServer(cfg, 8080, logPath)
		if err != nil {
			t.Fatalf("Failed to create server with log path: %v", err)
		}

		if server == nil {
			t.Fatal("Server is nil with log path")
		}
	})

	// Test with invalid log path
	t.Run("with invalid log path", func(t *testing.T) {
		// Use a path that shouldn't be writable
		logPath := "/nonexistent/directory/that/should/not/exist/server.log"

		_, err := service.NewServer(cfg, 8080, logPath)

		// This should fail
		if err == nil {
			t.Fatal("Expected error for invalid log path, but got nil")
		}
	})
}

// helper to build a CallToolRequest.
func makeToolRequest(args map[string]interface{}) mcp.CallToolRequest {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	return req
}

func newTestServer(t *testing.T) (*service.Server, string) {
	t.Helper()
	tmpDir := t.TempDir()

	cfg := &config.ShellCommandConfig{
		AllowedDirectories: []string{tmpDir},
		AllowCommands: []config.AllowCommand{
			{Command: "echo"},
			{Command: "pwd"},
		},
		DenyCommands:        []config.DenyCommand{},
		DefaultErrorMessage: "Command not allowed",
		MaxExecutionTime:    10,
		MaxOutputSize:       1024,
	}

	srv, err := service.NewServer(cfg, 0, "")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	return srv, tmpDir
}

func TestChangeDirectory(t *testing.T) {
	srv, tmpDir := newTestServer(t)
	ctx := t.Context()

	t.Run("allowed path succeeds", func(t *testing.T) {
		result, err := srv.HandleChangeDirectory(ctx, makeToolRequest(map[string]interface{}{
			"directory": tmpDir,
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertToolSuccess(t, result, "Working directory set to")
	})

	t.Run("disallowed path fails", func(t *testing.T) {
		result, err := srv.HandleChangeDirectory(ctx, makeToolRequest(map[string]interface{}{
			"directory": "/usr/local",
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertToolError(t, result, "not allowed")
	})

	t.Run("nonexistent path fails", func(t *testing.T) {
		result, err := srv.HandleChangeDirectory(ctx, makeToolRequest(map[string]interface{}{
			"directory": tmpDir + "/nonexistent",
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertToolError(t, result, "does not exist")
	})

	t.Run("empty string fails", func(t *testing.T) {
		result, err := srv.HandleChangeDirectory(ctx, makeToolRequest(map[string]interface{}{
			"directory": "",
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertToolError(t, result, "non-empty string")
	})
}

func TestRunCommand(t *testing.T) {
	srv, tmpDir := newTestServer(t)
	ctx := t.Context()

	t.Run("run without setting directory returns error", func(t *testing.T) {
		result, err := srv.HandleRunCommand(ctx, makeToolRequest(map[string]interface{}{
			"commands": []interface{}{"echo hello"},
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertToolError(t, result, "No working directory set")
	})

	t.Run("single command succeeds", func(t *testing.T) {
		_, _ = srv.HandleChangeDirectory(ctx, makeToolRequest(map[string]interface{}{
			"directory": tmpDir,
		}))
		result, err := srv.HandleRunCommand(ctx, makeToolRequest(map[string]interface{}{
			"commands": []interface{}{"echo hello"},
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertToolSuccess(t, result, "hello")
	})

	t.Run("invalid mode returns error", func(t *testing.T) {
		result, err := srv.HandleRunCommand(ctx, makeToolRequest(map[string]interface{}{
			"commands": []interface{}{"echo hello"},
			"mode":     "invalid",
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertToolError(t, result, "Mode must be")
	})

	t.Run("empty commands array fails", func(t *testing.T) {
		result, err := srv.HandleRunCommand(ctx, makeToolRequest(map[string]interface{}{
			"commands": []interface{}{},
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertToolError(t, result, "non-empty array")
	})

	t.Run("empty string in commands fails", func(t *testing.T) {
		result, err := srv.HandleRunCommand(ctx, makeToolRequest(map[string]interface{}{
			"commands": []interface{}{""},
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertToolError(t, result, "non-empty string")
	})

	t.Run("directory persists across calls", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			result, err := srv.HandleRunCommand(ctx, makeToolRequest(map[string]interface{}{
				"commands": []interface{}{"echo persist"},
			}))
			if err != nil {
				t.Fatalf("run %d: unexpected error: %v", i, err)
			}
			assertToolSuccess(t, result, "persist")
		}
	})
}

func TestRunCommandMultiple(t *testing.T) {
	srv, tmpDir := newTestServer(t)
	ctx := t.Context()
	_, _ = srv.HandleChangeDirectory(ctx, makeToolRequest(map[string]interface{}{
		"directory": tmpDir,
	}))

	t.Run("parallel default", func(t *testing.T) {
		result, err := srv.HandleRunCommand(ctx, makeToolRequest(map[string]interface{}{
			"commands": []interface{}{"echo aaa", "echo bbb"},
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := extractText(result)
		if !strings.Contains(text, "aaa") || !strings.Contains(text, "bbb") {
			t.Fatalf("expected both outputs, got: %s", text)
		}
	})

	t.Run("serial mode", func(t *testing.T) {
		result, err := srv.HandleRunCommand(ctx, makeToolRequest(map[string]interface{}{
			"commands": []interface{}{"echo first", "echo second"},
			"mode":     "serial",
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := extractText(result)
		if !strings.Contains(text, "first") || !strings.Contains(text, "second") {
			t.Fatalf("expected both outputs, got: %s", text)
		}
	})

	t.Run("serial stops on first error", func(t *testing.T) {
		result, err := srv.HandleRunCommand(ctx, makeToolRequest(map[string]interface{}{
			"commands": []interface{}{"rm forbidden", "echo should_not_run"},
			"mode":     "serial",
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := extractText(result)
		if strings.Contains(text, "should_not_run") {
			t.Fatalf("serial mode should have stopped on first error, got: %s", text)
		}
		assertToolError(t, result, "Error")
	})
}

func assertToolError(t *testing.T, result *mcp.CallToolResult, contains string) {
	t.Helper()
	if !result.IsError {
		t.Fatalf("expected error result, got success")
	}
	text := extractText(result)
	if !strings.Contains(text, contains) {
		t.Fatalf("expected error containing %q, got: %s", contains, text)
	}
}

func assertToolSuccess(t *testing.T, result *mcp.CallToolResult, contains string) {
	t.Helper()
	if result.IsError {
		t.Fatalf("expected success, got error: %s", extractText(result))
	}
	text := extractText(result)
	if !strings.Contains(text, contains) {
		t.Fatalf("expected output containing %q, got: %s", contains, text)
	}
}

func extractText(result *mcp.CallToolResult) string {
	var sb strings.Builder
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			sb.WriteString(tc.Text)
		}
	}
	return sb.String()
}
