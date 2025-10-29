package main

import (
	"fmt"
	"os"

	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/config"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/logger"
	x402server "github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/server"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/tools"
	"github.com/mark3labs/mcp-go/server"
)

const (
	serverName    = "x402-payment-mcp-server"
	serverVersion = "0.1.0"
	configPath    = "config.yaml"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid config: %v\n", err)
		os.Exit(1)
	}

	// Initialize structured logger
	logLevel := logger.INFO
	if cfg.Logging.Level == "DEBUG" {
		logLevel = logger.DEBUG
	} else if cfg.Logging.Level == "WARN" {
		logLevel = logger.WARN
	} else if cfg.Logging.Level == "ERROR" {
		logLevel = logger.ERROR
	}

	log := logger.New(logLevel, os.Stderr)
	log.Info("Starting x402 Payment MCP Server", map[string]interface{}{
		"version": serverVersion,
		"config":  configPath,
	})

	// Create MCP server instance
	mcpServer := server.NewMCPServer(
		serverName,
		serverVersion,
		server.WithStdio(),
	)

	// Initialize x402 server with tools
	x402Server, err := x402server.NewServer(cfg, log)
	if err != nil {
		log.Error("Failed to initialize x402 server", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}

	// Create and add tools
	createPaymentTool := tools.NewCreatePaymentRequirementTool(x402Server)
	if err := x402Server.AddTool(createPaymentTool); err != nil {
		log.Error("Failed to add create_payment_requirement tool", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}

	verifyPaymentTool := tools.NewVerifyPaymentTool(x402Server)
	if err := x402Server.AddTool(verifyPaymentTool); err != nil {
		log.Error("Failed to add verify_payment tool", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}

	// Register tools with MCP server
	if err := x402Server.RegisterTools(mcpServer); err != nil {
		log.Error("Failed to register tools", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}

	log.Info("Server initialized successfully", map[string]interface{}{
		"tools_registered": x402Server.ToolCount(),
	})

	// Start server (blocking)
	if err := mcpServer.Serve(); err != nil {
		log.Error("Server error", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}
}
