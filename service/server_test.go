package service_test

import (
	"testing"

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
