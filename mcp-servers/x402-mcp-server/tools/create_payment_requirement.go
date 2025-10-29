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
	return "Generate x402-compliant payment requirement for blockchain certification payment. Returns payment details including amount, payee address, validity period, and blockchain nonce."
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
				"enum":        []string{"base", "base-sepolia", "arbitrum"},
			},
		},
		"required": []string{"amount", "network"},
	}
}

// Execute executes the tool with the given arguments
func (t *CreatePaymentRequirementTool) Execute(args map[string]interface{}) (interface{}, error) {
	// Extract amount
	amount, ok := args["amount"].(string)
	if !ok {
		return nil, fmt.Errorf("amount must be a string")
	}

	// Extract network
	network, ok := args["network"].(string)
	if !ok {
		return nil, fmt.Errorf("network must be a string")
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
		24*time.Hour,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment requirement: %w", err)
	}

	// Log the operation
	logger := t.server.GetLogger()
	logger.Info("Created payment requirement", map[string]interface{}{
		"network": network,
		"amount":  amount,
		"nonce":   paymentReq.Nonce,
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
