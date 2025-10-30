package tools

import (
	"fmt"
	"time"

	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/server"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/x402"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// CreatePaymentRequirementTool implements the create_payment_requirement MCP tool
type CreatePaymentRequirementTool struct {
	server *server.Server
}

// NewCreatePaymentRequirementTool creates a new create_payment_requirement tool
func NewCreatePaymentRequirementTool(srv *server.Server) *CreatePaymentRequirementTool {
	return &CreatePaymentRequirementTool{
		server: srv,
	}
}

// Name returns the tool name
func (t *CreatePaymentRequirementTool) Name() string {
	return "create_payment_requirement"
}

// Description returns the tool description
func (t *CreatePaymentRequirementTool) Description() string {
	return "Generate x402-compliant payment requirement per official Coinbase x402 specification. Returns complete payment requirement with resource URL, description, and payment details."
}

// Schema returns the JSON schema for the tool's input
func (t *CreatePaymentRequirementTool) Schema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"amount": map[string]interface{}{
				"type":        "string",
				"description": "Payment amount in USDC atomic units (6 decimals). Example: '50000' = 0.05 USDC",
				"pattern":     "^[1-9][0-9]*$",
			},
			"network": map[string]interface{}{
				"type":        "string",
				"description": "Blockchain network for payment",
				"enum":        []interface{}{"base", "base-sepolia", "arbitrum"},
			},
			"resource": map[string]interface{}{
				"type":        "string",
				"description": "URL of the resource being paid for (e.g., certification endpoint)",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Human-readable description of what the payment is for",
			},
			"mime_type": map[string]interface{}{
				"type":        "string",
				"description": "MIME type of the resource response (default: application/json)",
				"default":     "application/json",
			},
		},
		"required": []interface{}{"amount", "network"},
	}
}

// Execute executes the tool with the given arguments
func (t *CreatePaymentRequirementTool) Execute(args map[string]interface{}) (interface{}, error) {
	// Extract required fields
	amount, ok := args["amount"].(string)
	if !ok {
		return nil, fmt.Errorf("amount must be a string")
	}

	network, ok := args["network"].(string)
	if !ok {
		return nil, fmt.Errorf("network must be a string")
	}

	// Extract optional resource with default
	resource, ok := args["resource"].(string)
	if !ok || resource == "" {
		resource = "https://api.example.com/resource"
	}

	// Extract optional description with default
	description, ok := args["description"].(string)
	if !ok || description == "" {
		description = "Payment requirement"
	}

	// Extract optional mime_type with default
	mimeType, ok := args["mime_type"].(string)
	if !ok || mimeType == "" {
		mimeType = "application/json"
	}

	// Get network configuration
	cfg := t.server.GetConfig()
	networkCfg, exists := cfg.Networks[network]
	if !exists {
		return nil, fmt.Errorf("unsupported network: %s", network)
	}

	// Create payment requirement with 24-hour validity
	paymentReq, err := x402.NewPaymentRequirement(
		amount,
		network,
		networkCfg.PayeeAddress,
		networkCfg.USDCContract,
		resource,
		description,
		mimeType,
		24*time.Hour,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment requirement: %w", err)
	}

	// Log the operation
	logger := t.server.GetLogger()
	logger.Info("Created payment requirement", map[string]interface{}{
		"network":     network,
		"amount":      amount,
		"resource":    resource,
		"description": description,
		"nonce":       paymentReq.Nonce,
	})

	// Return as map for MCP
	return paymentReq.ToMap(), nil
}

// Register registers the tool with the MCP server
func (t *CreatePaymentRequirementTool) Register(mcpServer *mcpserver.MCPServer) error {
	if mcpServer == nil {
		return fmt.Errorf("MCP server is nil")
	}

	// For now, registration will be handled externally
	// The mcp-go API requires different registration approach
	return nil
}
