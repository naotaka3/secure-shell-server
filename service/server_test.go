package service_test

import (
	"testing"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/service"
)

func TestNewServer(t *testing.T) {
	// Create a test configuration
	cfg := config.NewDefaultConfig()

	// Create a test server
	server := service.NewServer(cfg, 8080)

	// Just a basic test to ensure server creation works
	if server == nil {
		t.Fatal("Failed to create server")
	}
}
