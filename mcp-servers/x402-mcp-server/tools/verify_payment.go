package tools

import (
	"fmt"

	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/eip3009"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/server"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// VerifyPaymentTool implements the verify_payment MCP tool
type VerifyPaymentTool struct {
	server   *server.Server
	verifier *eip3009.SignatureVerifier
}

// NewVerifyPaymentTool creates a new verify_payment tool
func NewVerifyPaymentTool(srv *server.Server) *VerifyPaymentTool {
	return &VerifyPaymentTool{
		server:   srv,
		verifier: eip3009.NewSignatureVerifier(srv.GetConfig()),
	}
}

// Name returns the tool name
func (t *VerifyPaymentTool) Name() string {
	return "verify_payment"
}

// Description returns the tool description
func (t *VerifyPaymentTool) Description() string {
	return "Verify EIP-3009 payment authorization signature using secp256k1 ECDSA recovery. Validates signature authenticity, time bounds, and EIP-712 domain matching for blockchain payment verification."
}

// Schema returns the JSON schema for the tool's input
func (t *VerifyPaymentTool) Schema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"authorization": map[string]interface{}{
				"type":        "object",
				"description": "EIP-3009 receiveWithAuthorization parameters",
				"properties": map[string]interface{}{
					"from": map[string]interface{}{
						"type":        "string",
						"description": "Payer address (0x-prefixed hex)",
						"pattern":     "^0x[a-fA-F0-9]{40}$",
					},
					"to": map[string]interface{}{
						"type":        "string",
						"description": "Payee address (0x-prefixed hex)",
						"pattern":     "^0x[a-fA-F0-9]{40}$",
					},
					"value": map[string]interface{}{
						"type":        "string",
						"description": "Amount in USDC atomic units (6 decimals)",
						"pattern":     "^[1-9][0-9]*$",
					},
					"validAfter": map[string]interface{}{
						"type":        "integer",
						"description": "Unix timestamp (seconds) after which the authorization is valid",
					},
					"validBefore": map[string]interface{}{
						"type":        "integer",
						"description": "Unix timestamp (seconds) before which the authorization is valid",
					},
					"nonce": map[string]interface{}{
						"type":        "string",
						"description": "Unique nonce as 32-byte hex string (0x-prefixed)",
						"pattern":     "^0x[a-fA-F0-9]{64}$",
					},
					"v": map[string]interface{}{
						"type":        "integer",
						"description": "ECDSA recovery parameter (27 or 28)",
						"enum":        []int{27, 28},
					},
					"r": map[string]interface{}{
						"type":        "string",
						"description": "ECDSA signature r component as 32-byte hex string",
						"pattern":     "^0x[a-fA-F0-9]{64}$",
					},
					"s": map[string]interface{}{
						"type":        "string",
						"description": "ECDSA signature s component as 32-byte hex string",
						"pattern":     "^0x[a-fA-F0-9]{64}$",
					},
				},
				"required": []string{"from", "to", "value", "validAfter", "validBefore", "nonce", "v", "r", "s"},
			},
			"network": map[string]interface{}{
				"type":        "string",
				"description": "Blockchain network for verification",
				"enum":        []string{"base", "base-sepolia", "arbitrum"},
			},
		},
		"required": []string{"authorization", "network"},
	}
}

// Execute executes the tool with the given arguments
func (t *VerifyPaymentTool) Execute(args map[string]interface{}) (interface{}, error) {
	// Extract network
	network, ok := args["network"].(string)
	if !ok {
		return nil, fmt.Errorf("network must be a string")
	}

	// Extract authorization object
	authMap, ok := args["authorization"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("authorization must be an object")
	}

	// Parse authorization fields
	auth, err := t.parseAuthorization(authMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse authorization: %w", err)
	}

	// Log verification attempt
	logger := t.server.GetLogger()
	logger.Info("Verifying payment authorization", map[string]interface{}{
		"network": network,
		"from":    auth.From,
		"to":      auth.To,
		"value":   auth.Value,
		"nonce":   auth.Nonce,
	})

	// Verify the authorization
	result, err := t.verifier.VerifyAuthorization(auth, network)
	if err != nil {
		logger.Error("Verification failed", map[string]interface{}{
			"error":   err.Error(),
			"network": network,
			"from":    auth.From,
		})
		return nil, fmt.Errorf("verification error: %w", err)
	}

	// Log result
	if result.IsValid {
		logger.Info("Signature verified successfully", map[string]interface{}{
			"network":        network,
			"signer_address": result.SignerAddress,
			"from":           auth.From,
		})
	} else {
		logger.Info("Signature verification failed", map[string]interface{}{
			"network": network,
			"error":   result.Error,
			"from":    auth.From,
		})
	}

	// Return as map for MCP
	return result.ToMap(), nil
}

// parseAuthorization converts the input map to an EIP3009Authorization struct
func (t *VerifyPaymentTool) parseAuthorization(authMap map[string]interface{}) (*eip3009.EIP3009Authorization, error) {
	// Extract required string fields
	from, ok := authMap["from"].(string)
	if !ok {
		return nil, fmt.Errorf("from must be a string")
	}

	to, ok := authMap["to"].(string)
	if !ok {
		return nil, fmt.Errorf("to must be a string")
	}

	value, ok := authMap["value"].(string)
	if !ok {
		return nil, fmt.Errorf("value must be a string")
	}

	nonce, ok := authMap["nonce"].(string)
	if !ok {
		return nil, fmt.Errorf("nonce must be a string")
	}

	r, ok := authMap["r"].(string)
	if !ok {
		return nil, fmt.Errorf("r must be a string")
	}

	s, ok := authMap["s"].(string)
	if !ok {
		return nil, fmt.Errorf("s must be a string")
	}

	// Extract uint64 fields (JSON numbers come as float64)
	validAfterFloat, ok := authMap["validAfter"].(float64)
	if !ok {
		return nil, fmt.Errorf("validAfter must be a number")
	}
	validAfter := uint64(validAfterFloat)

	validBeforeFloat, ok := authMap["validBefore"].(float64)
	if !ok {
		return nil, fmt.Errorf("validBefore must be a number")
	}
	validBefore := uint64(validBeforeFloat)

	// Extract v (could be float64 or int)
	var v uint8
	switch vVal := authMap["v"].(type) {
	case float64:
		v = uint8(vVal)
	case int:
		v = uint8(vVal)
	default:
		return nil, fmt.Errorf("v must be a number")
	}

	// Validate v is 27 or 28
	if v != 27 && v != 28 {
		return nil, fmt.Errorf("v must be 27 or 28, got %d", v)
	}

	return &eip3009.EIP3009Authorization{
		From:        from,
		To:          to,
		Value:       value,
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
		Nonce:       nonce,
		V:           v,
		R:           r,
		S:           s,
	}, nil
}

// Register registers the tool with the MCP server
func (t *VerifyPaymentTool) Register(mcpServer *mcpserver.MCPServer) error {
	if mcpServer == nil {
		return fmt.Errorf("MCP server is nil")
	}

	// For now, registration will be handled externally
	// The mcp-go API requires different registration approach
	return nil
}
