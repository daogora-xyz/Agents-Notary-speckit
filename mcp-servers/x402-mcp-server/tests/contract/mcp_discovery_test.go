package contract

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/config"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/logger"
	x402server "github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/server"
	"github.com/mark3labs/mcp-go/server"
)

// TestMCPServer_Initialize verifies the MCP server can be created
func TestMCPServer_Initialize(t *testing.T) {
	cfg := createTestConfig()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if srv == nil {
		t.Fatal("Server should not be nil")
	}

	if srv.ToolCount() != 0 {
		t.Logf("Note: Server has %d tools (expected 0 in Phase 2)", srv.ToolCount())
	}
}

// TestMCPServer_ToolRegistration verifies tool registration framework
func TestMCPServer_ToolRegistration(t *testing.T) {
	cfg := createTestConfig()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create MCP server instance
	mcpServer := server.NewMCPServer(
		"test-server",
		"0.1.0",
	)

	// Register tools (should not error even with 0 tools)
	err = srv.RegisterTools(mcpServer)
	if err != nil {
		t.Errorf("Tool registration failed: %v", err)
	}
}

// TestMCPServer_AddToolDuplicate verifies duplicate tool prevention
func TestMCPServer_AddToolDuplicate(t *testing.T) {
	cfg := createTestConfig()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create mock tool
	mockTool := &MockTool{
		name:        "test_tool",
		description: "Test tool",
	}

	// First add should succeed
	err = srv.AddTool(mockTool)
	if err != nil {
		t.Fatalf("First AddTool failed: %v", err)
	}

	// Second add should fail (duplicate name)
	err = srv.AddTool(mockTool)
	if err == nil {
		t.Error("Expected error for duplicate tool, got nil")
	}
}

// TestMCPServer_AddToolNil verifies nil tool rejection
func TestMCPServer_AddToolNil(t *testing.T) {
	cfg := createTestConfig()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Adding nil tool should fail
	err = srv.AddTool(nil)
	if err == nil {
		t.Error("Expected error for nil tool, got nil")
	}
}

// TestMCPServer_ToolDiscovery verifies tools list request format
func TestMCPServer_ToolDiscovery(t *testing.T) {
	cfg := createTestConfig()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Add mock tool
	mockTool := &MockTool{
		name:        "create_payment_requirement",
		description: "Create a payment requirement for x402 protocol",
	}

	err = srv.AddTool(mockTool)
	if err != nil {
		t.Fatalf("AddTool failed: %v", err)
	}

	// Verify tool count
	if srv.ToolCount() != 1 {
		t.Errorf("Expected 1 tool, got %d", srv.ToolCount())
	}

	// In real MCP, tools/list would return JSON-RPC response
	// For now, verify the structure is correct
	expectedName := "create_payment_requirement"
	if mockTool.Name() != expectedName {
		t.Errorf("Expected tool name %s, got %s", expectedName, mockTool.Name())
	}
}

// TestMCPServer_ConfigAccess verifies config is accessible
func TestMCPServer_ConfigAccess(t *testing.T) {
	cfg := createTestConfig()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	retrievedConfig := srv.GetConfig()
	if retrievedConfig == nil {
		t.Fatal("Config should not be nil")
	}

	if retrievedConfig.EIP712.DomainName != "USD Coin" {
		t.Errorf("Expected domain name 'USD Coin', got %s", retrievedConfig.EIP712.DomainName)
	}
}

// TestMCPServer_LoggerAccess verifies logger is accessible
func TestMCPServer_LoggerAccess(t *testing.T) {
	cfg := createTestConfig()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	retrievedLogger := srv.GetLogger()
	if retrievedLogger == nil {
		t.Fatal("Logger should not be nil")
	}
}

// TestMCPServer_CacheAccess verifies cache is accessible
func TestMCPServer_CacheAccess(t *testing.T) {
	cfg := createTestConfig()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	cache := srv.GetCache()
	if cache == nil {
		t.Fatal("Cache should not be nil")
	}

	// Verify cache is functional
	cache.Set("test_key", "test_value")
	val, found := cache.Get("test_key")
	if !found {
		t.Error("Expected to find test_key in cache")
	}

	if val != "test_value" {
		t.Errorf("Expected 'test_value', got %v", val)
	}
}

// MockTool implements the Tool interface for testing
type MockTool struct {
	name        string
	description string
}

func (m *MockTool) Name() string {
	return m.name
}

func (m *MockTool) Description() string {
	return m.description
}

func (m *MockTool) Schema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"test_param": map[string]interface{}{
				"type":        "string",
				"description": "Test parameter",
			},
		},
		"required": []string{"test_param"},
	}
}

func (m *MockTool) Register(s *server.MCPServer) error {
	// Mock registration - just verify server is not nil
	if s == nil {
		return fmt.Errorf("MCP server is nil")
	}
	return nil
}

// createTestConfig creates a minimal config for testing
func createTestConfig() *config.Config {
	return &config.Config{
		Networks: map[string]config.NetworkConfig{
			"base": {
				ChainID:        8453,
				USDCContract:   "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
				FacilitatorURL: "https://api.cdp.coinbase.com",
				RPCURL:         "https://mainnet.base.org",
				PayeeAddress:   "0x1234567890123456789012345678901234567890",
			},
		},
		EIP712: config.EIP712Config{
			DomainName:    "USD Coin",
			DomainVersion: "2",
		},
		Logging: config.LoggingConfig{
			Level:  "DEBUG",
			Format: "json",
		},
		Cache: config.CacheConfig{
			SettlementTTLMinutes: 10,
		},
	}
}
