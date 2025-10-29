package tools

import (
	"fmt"
	"time"

	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/eip3009"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/facilitator"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/server"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// SettlePaymentTool implements the settle_payment MCP tool
type SettlePaymentTool struct {
	server            *server.Server
	verifier          *eip3009.SignatureVerifier
	facilitatorClient *facilitator.Client
}

// NewSettlePaymentTool creates a new settle_payment tool
func NewSettlePaymentTool(srv *server.Server) *SettlePaymentTool {
	return &SettlePaymentTool{
		server:            srv,
		verifier:          eip3009.NewSignatureVerifier(srv.GetConfig()),
		facilitatorClient: facilitator.NewClient(srv.GetConfig(), 5*time.Second),
	}
}

// Name returns the tool name
func (t *SettlePaymentTool) Name() string {
	return "settle_payment"
}

// Description returns the tool description
func (t *SettlePaymentTool) Description() string {
	return "Submit verified EIP-3009 payment authorization to x402 facilitator for on-chain settlement. Returns settlement status (settled/pending/failed) with transaction details. Implements idempotency caching to prevent duplicate submissions."
}

// Schema returns the JSON schema for the tool's input
func (t *SettlePaymentTool) Schema() interface{} {
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
				"description": "Blockchain network for settlement",
				"enum":        []string{"base", "base-sepolia", "arbitrum"},
			},
		},
		"required": []string{"authorization", "network"},
	}
}

// Execute executes the tool with the given arguments
func (t *SettlePaymentTool) Execute(args map[string]interface{}) (interface{}, error) {
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

	logger := t.server.GetLogger()
	logger.Info("Settling payment authorization", map[string]interface{}{
		"network": network,
		"from":    auth.From,
		"to":      auth.To,
		"value":   auth.Value,
		"nonce":   auth.Nonce,
	})

	// Step 1: Verify signature before settlement (FR-011 requirement)
	verifyResult, err := t.verifier.VerifyAuthorization(auth, network)
	if err != nil {
		logger.Error("Signature verification failed before settlement", map[string]interface{}{
			"error":   err.Error(),
			"network": network,
			"from":    auth.From,
		})
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	if !verifyResult.IsValid {
		logger.Warn("Invalid signature - refusing settlement", map[string]interface{}{
			"network": network,
			"from":    auth.From,
			"error":   verifyResult.Error,
		})
		return map[string]interface{}{
			"status": "failed",
			"error":  fmt.Sprintf("invalid signature: %s", verifyResult.Error),
		}, nil
	}

	logger.Info("Signature verified successfully, submitting to facilitator", map[string]interface{}{
		"network":        network,
		"signer_address": verifyResult.SignerAddress,
	})

	// Step 2: Submit to facilitator
	startTime := time.Now()
	result, err := t.facilitatorClient.SubmitSettlement(auth, network)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		logger.Error("Facilitator submission failed", map[string]interface{}{
			"error":       err.Error(),
			"network":     network,
			"from":        auth.From,
			"duration_ms": duration,
		})
		return nil, fmt.Errorf("facilitator submission failed: %w", err)
	}

	// Log result
	logContext := map[string]interface{}{
		"network":     network,
		"status":      result.Status,
		"duration_ms": duration,
		"from":        auth.From,
		"nonce":       auth.Nonce,
	}

	if result.Status == "settled" {
		logContext["tx_hash"] = result.TxHash
		logContext["block_number"] = result.BlockNumber
		logger.Info("Payment settled successfully", logContext)
	} else if result.Status == "pending" {
		logContext["retry_after"] = result.RetryAfter
		logger.Info("Payment settlement pending", logContext)
	} else {
		logContext["error"] = result.Error
		logger.Warn("Payment settlement failed", logContext)
	}

	// Return facilitator response
	return result.ToMap(), nil
}

// parseAuthorization converts the input map to an EIP3009Authorization struct
func (t *SettlePaymentTool) parseAuthorization(authMap map[string]interface{}) (*eip3009.EIP3009Authorization, error) {
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
func (t *SettlePaymentTool) Register(mcpServer *mcpserver.MCPServer) error {
	if mcpServer == nil {
		return fmt.Errorf("MCP server is nil")
	}

	// For now, registration will be handled externally
	// The mcp-go API requires different registration approach
	return nil
}
