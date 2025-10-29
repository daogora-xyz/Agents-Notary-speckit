package server

import (
	"fmt"
	"time"

	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/cache"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/config"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/logger"
	"github.com/mark3labs/mcp-go/server"
)

// Server represents the x402 MCP server instance
type Server struct {
	config *config.Config
	logger *logger.Logger
	cache  *cache.TTLCache
	tools  []Tool
}

// Tool represents an MCP tool handler
type Tool interface {
	Name() string
	Description() string
	Schema() interface{}
	Register(s *server.MCPServer) error
}

// NewServer creates a new x402 server instance
func NewServer(cfg *config.Config, log *logger.Logger) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if log == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// Initialize cache with configured TTL
	cacheTTL := time.Duration(cfg.Cache.SettlementTTLMinutes) * time.Minute
	settlementCache := cache.NewTTLCache(cacheTTL)

	srv := &Server{
		config: cfg,
		logger: log,
		cache:  settlementCache,
		tools:  make([]Tool, 0),
	}

	// Initialize tools (will be added in subsequent phases)
	if err := srv.initializeTools(); err != nil {
		return nil, fmt.Errorf("failed to initialize tools: %w", err)
	}

	return srv, nil
}

// initializeTools sets up all available MCP tools
func (s *Server) initializeTools() error {
	s.logger.Debug("Initializing MCP tools", nil)

	// Import tools dynamically to avoid import cycles
	// For now, tools will be registered externally via AddTool

	// TODO: Add tools in subsequent phases
	// - verify_payment_authorization (Phase 4)
	// - settle_payment (Phase 5)
	// - get_payment_status (Phase 6)

	s.logger.Info("Tools initialized", map[string]interface{}{
		"count": len(s.tools),
	})

	return nil
}

// RegisterTools registers all tools with the MCP server
func (s *Server) RegisterTools(mcpServer *server.MCPServer) error {
	if mcpServer == nil {
		return fmt.Errorf("mcp server cannot be nil")
	}

	s.logger.Info("Registering tools with MCP server", map[string]interface{}{
		"tool_count": len(s.tools),
	})

	for _, tool := range s.tools {
		if err := tool.Register(mcpServer); err != nil {
			return fmt.Errorf("failed to register tool %s: %w", tool.Name(), err)
		}

		s.logger.Debug("Registered tool", map[string]interface{}{
			"tool": tool.Name(),
		})
	}

	return nil
}

// ToolCount returns the number of registered tools
func (s *Server) ToolCount() int {
	return len(s.tools)
}

// GetConfig returns the server configuration
func (s *Server) GetConfig() *config.Config {
	return s.config
}

// GetLogger returns the server logger
func (s *Server) GetLogger() *logger.Logger {
	return s.logger
}

// GetCache returns the settlement cache
func (s *Server) GetCache() *cache.TTLCache {
	return s.cache
}

// AddTool adds a tool to the server's tool registry
func (s *Server) AddTool(tool Tool) error {
	if tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}

	// Check for duplicate tool names
	for _, existingTool := range s.tools {
		if existingTool.Name() == tool.Name() {
			return fmt.Errorf("tool with name %s already registered", tool.Name())
		}
	}

	s.tools = append(s.tools, tool)

	s.logger.Debug("Added tool", map[string]interface{}{
		"tool":        tool.Name(),
		"description": tool.Description(),
	})

	return nil
}
