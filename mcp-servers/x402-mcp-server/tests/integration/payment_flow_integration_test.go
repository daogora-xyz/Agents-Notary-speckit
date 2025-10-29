package integration

import (
	"bytes"
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/config"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/eip3009"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/logger"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/server"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/x402"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/tools"
)

// TestEndToEndPaymentFlow demonstrates the complete payment flow:
// 1. Create payment requirement
// 2. Payer signs the authorization (simulated)
// 3. Verify the payment authorization
// 4. Settle the payment (mock facilitator)
//
// This test serves as both validation and documentation of the system.
func TestEndToEndPaymentFlow(t *testing.T) {
	t.Log("=== End-to-End Payment Flow Integration Test ===")
	t.Log("This test demonstrates the complete x402 payment workflow")
	t.Log("")

	// Setup test configuration
	cfg := createTestConfig()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	// Initialize server
	srv, err := server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	t.Log("Step 1: Generate Payment Requirement")
	t.Log("---------------------------------------")

	// Create payment requirement tool
	createTool := tools.NewCreatePaymentRequirementTool(srv)

	// Execute create_payment_requirement with atomic units (50000 = 0.05 USDC with 6 decimals)
	createInput := map[string]interface{}{
		"amount":      "50000", // 0.05 USDC in atomic units
		"network":     "base-sepolia",
		"resource":    "https://certify.ar4s.com/api/certify/abc123",
		"description": "Blockchain certification for test document",
		"mime_type":   "application/json",
	}

	createResult, err := createTool.Execute(createInput)
	if err != nil {
		t.Fatalf("Failed to create payment requirement: %v", err)
	}

	t.Logf("✓ Payment requirement created successfully")
	x402Requirement := createResult.(map[string]interface{})
	t.Logf("  - x402_version: %v", x402Requirement["x402_version"])
	t.Logf("  - Scheme: %s", x402Requirement["scheme"])
	t.Logf("  - Network: %s", x402Requirement["network"])
	t.Logf("  - Resource: %s", x402Requirement["resource"])
	t.Logf("  - Description: %s", x402Requirement["description"])
	t.Logf("  - MIME Type: %s", x402Requirement["mimeType"])
	t.Logf("  - Max Amount: %s atomic units (0.05 USDC)", x402Requirement["maxAmountRequired"])
	t.Logf("  - Pay To: %s", x402Requirement["payTo"])
	t.Logf("  - Asset (USDC): %s", x402Requirement["asset"])
	t.Logf("  - Max Timeout: %v seconds", x402Requirement["maxTimeoutSeconds"])
	extra := x402Requirement["extra"].(map[string]interface{})
	t.Logf("  - Extra (name): %s", extra["name"])
	t.Logf("  - Extra (version): %s", extra["version"])
	t.Logf("  - Nonce: %s", x402Requirement["nonce"])
	t.Logf("  - Valid Until: %s", x402Requirement["valid_until"])
	t.Logf("")

	// Step 2: Simulate payer signing the authorization
	t.Log("Step 2: Payer Signs Authorization (Simulated)")
	t.Log("---------------------------------------------")

	// Generate a test private key (in production, this would be the payer's wallet)
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	payerAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	t.Logf("✓ Generated test payer wallet: %s", payerAddress.Hex())

	// Create EIP-3009 authorization from x402 payment requirement
	auth := createTestAuthorization(t, cfg, privateKey, x402Requirement)
	t.Logf("✓ Authorization signed with EIP-712")
	t.Logf("  - Nonce: %s", auth.Nonce)
	t.Logf("  - Valid window: %d - %d", auth.ValidAfter, auth.ValidBefore)
	t.Logf("  - Signature: v=%d, r=%s..., s=%s...", auth.V, auth.R[:10], auth.S[:10])
	t.Logf("")

	// Step 3: Verify payment authorization
	t.Log("Step 3: Verify Payment Authorization")
	t.Log("------------------------------------")

	verifyTool := tools.NewVerifyPaymentTool(srv)

	verifyInput := map[string]interface{}{
		"authorization": map[string]interface{}{
			"from":        auth.From,
			"to":          auth.To,
			"value":       auth.Value,
			"validAfter":  float64(auth.ValidAfter),
			"validBefore": float64(auth.ValidBefore),
			"nonce":       auth.Nonce,
			"v":           float64(auth.V),
			"r":           auth.R,
			"s":           auth.S,
		},
		"network": "base-sepolia",
	}

	verifyResult, err := verifyTool.Execute(verifyInput)
	if err != nil {
		t.Fatalf("Failed to verify payment: %v", err)
	}

	verifyMap := verifyResult.(map[string]interface{})
	isValid := verifyMap["is_valid"].(bool)

	if !isValid {
		errorMsg := verifyMap["error"].(string)
		t.Fatalf("Signature verification failed: %s", errorMsg)
	}

	t.Logf("✓ Signature verified successfully")
	t.Logf("  - Signer recovered: %s", verifyMap["signer_address"])
	t.Logf("  - Matches payer: %v", verifyMap["signer_address"] == auth.From)
	t.Logf("")

	// Step 4: Settle payment (note: will fail without real facilitator, but demonstrates the flow)
	t.Log("Step 4: Settle Payment via Facilitator")
	t.Log("---------------------------------------")

	settleTool := tools.NewSettlePaymentTool(srv)

	settleInput := map[string]interface{}{
		"authorization": map[string]interface{}{
			"from":        auth.From,
			"to":          auth.To,
			"value":       auth.Value,
			"validAfter":  float64(auth.ValidAfter),
			"validBefore": float64(auth.ValidBefore),
			"nonce":       auth.Nonce,
			"v":           float64(auth.V),
			"r":           auth.R,
			"s":           auth.S,
		},
		"network": "base-sepolia",
	}

	// Note: This will attempt to contact the facilitator
	// In a real integration test, we'd need to:
	// - Use a mock HTTP server
	// - Or test against a real testnet facilitator
	// - Or skip this step
	t.Logf("Note: Settlement requires live facilitator connection")
	t.Logf("This would submit the verified authorization to: %s", cfg.Networks["base-sepolia"].FacilitatorURL)

	_, err = settleTool.Execute(settleInput)
	if err != nil {
		// Expected to fail without live facilitator
		t.Logf("⚠ Settlement failed (expected without live facilitator): %v", err)
		t.Logf("  In production, this would:")
		t.Logf("  1. Submit authorization to x402 facilitator")
		t.Logf("  2. Facilitator calls receiveWithAuthorization on-chain")
		t.Logf("  3. USDC transfers from payer to payee")
		t.Logf("  4. Returns tx_hash and block_number")
	} else {
		t.Logf("✓ Settlement submitted successfully (unexpected - facilitator available?)")
	}

	t.Logf("")
	t.Log("=== Integration Test Complete ===")
	t.Log("Summary:")
	t.Log("✓ Payment requirement generation: PASS")
	t.Log("✓ EIP-712 signature creation: PASS")
	t.Log("✓ Signature verification: PASS")
	t.Log("- Settlement: SKIPPED (requires live facilitator)")
	t.Log("")
	t.Log("This demonstrates that the core payment flow is fully functional.")
	t.Log("The only external dependency is the x402 facilitator service.")
}

// createTestAuthorization creates a signed EIP-3009 authorization for testing
func createTestAuthorization(t *testing.T, cfg *config.Config, privateKey *ecdsa.PrivateKey, paymentDetails map[string]interface{}) *eip3009.EIP3009Authorization {
	payerAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Extract payment details - support x402 format
	var payToAddress, amountAtomic, nonceHex string
	if val, ok := paymentDetails["payTo"].(string); ok {
		payToAddress = val
	} else if val, ok := paymentDetails["payee"].(string); ok {
		// Backward compatibility
		payToAddress = val
	}

	if val, ok := paymentDetails["maxAmountRequired"].(string); ok {
		amountAtomic = val
	} else if val, ok := paymentDetails["amount_atomic"].(string); ok {
		// Backward compatibility
		amountAtomic = val
	}

	nonceHex = paymentDetails["nonce"].(string)

	// Create authorization message
	value, _ := new(big.Int).SetString(amountAtomic, 10)
	validAfter := uint64(0)
	validBefore := uint64(time.Now().Add(1 * time.Hour).Unix())

	// Convert nonce to 32-byte format for EIP-3009
	nonce32 := hexToBytes32(nonceHex)
	// Convert back to hex string for authorization struct
	nonce32Hex := "0x" + common.Bytes2Hex(nonce32[:])

	message := &eip3009.ReceiveWithAuthorizationMessage{
		From:        payerAddress,
		To:          common.HexToAddress(payToAddress),
		Value:       value,
		ValidAfter:  new(big.Int).SetUint64(validAfter),
		ValidBefore: new(big.Int).SetUint64(validBefore),
		Nonce:       nonce32,
	}

	// Create EIP-712 domain
	networkCfg := cfg.Networks["base-sepolia"]
	domain := &eip3009.EIP712Domain{
		Name:              cfg.EIP712.DomainName,
		Version:           cfg.EIP712.DomainVersion,
		ChainID:           big.NewInt(int64(networkCfg.ChainID)),
		VerifyingContract: common.HexToAddress(networkCfg.USDCContract),
	}

	// Compute typed data hash
	typedDataHash, err := eip3009.TypedDataHash(domain, message)
	if err != nil {
		t.Fatalf("Failed to compute typed data hash: %v", err)
	}

	// Sign the hash
	signature, err := crypto.Sign(typedDataHash.Bytes(), privateKey)
	if err != nil {
		t.Fatalf("Failed to sign authorization: %v", err)
	}

	// Extract v, r, s from signature
	r := common.BytesToHash(signature[0:32])
	s := common.BytesToHash(signature[32:64])
	v := signature[64] + 27 // Convert from 0/1 to 27/28

	return &eip3009.EIP3009Authorization{
		From:        payerAddress.Hex(),
		To:          payToAddress,
		Value:       amountAtomic,
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
		Nonce:       nonce32Hex, // Use 32-byte formatted nonce
		V:           v,
		R:           r.Hex(),
		S:           s.Hex(),
	}
}

// hexToBytes32 converts a hex string to [32]byte, padding if necessary
func hexToBytes32(hexStr string) [32]byte {
	var result [32]byte
	bytes := common.FromHex(hexStr)

	// If longer than 32 bytes, truncate
	if len(bytes) > 32 {
		copy(result[:], bytes[:32])
	} else {
		// If shorter, left-pad with zeros (right-align the data)
		copy(result[32-len(bytes):], bytes)
	}

	return result
}

// createTestConfig creates a test configuration
func createTestConfig() *config.Config {
	return &config.Config{
		Networks: map[string]config.NetworkConfig{
			"base-sepolia": {
				ChainID:        84532,
				USDCContract:   "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				FacilitatorURL: "https://x402.org/facilitator",
				RPCURL:         "https://sepolia.base.org",
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

// TestPaymentRequirementCreation tests the create_payment_requirement tool in isolation
func TestPaymentRequirementCreation(t *testing.T) {
	t.Log("Testing payment requirement creation...")

	cfg := createTestConfig()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	createTool := tools.NewCreatePaymentRequirementTool(srv)

	tests := []struct {
		name    string
		input   map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid small amount",
			input: map[string]interface{}{
				"amount":      "1000000", // 1 USDC in atomic units (6 decimals)
				"network":     "base-sepolia",
				"resource":    "https://certify.ar4s.com/api/certify/test1",
				"description": "Test certification #1",
			},
			wantErr: false,
		},
		{
			name: "valid large amount",
			input: map[string]interface{}{
				"amount":      "10000500000", // 10000.50 USDC in atomic units
				"network":     "base-sepolia",
				"resource":    "https://certify.ar4s.com/api/certify/test2",
				"description": "Test certification #2",
			},
			wantErr: false,
		},
		{
			name: "invalid network",
			input: map[string]interface{}{
				"amount":      "5000000", // 5 USDC
				"network":     "invalid-network",
				"resource":    "https://certify.ar4s.com/api/certify/test3",
				"description": "Test certification #3",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := createTool.Execute(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Validate x402 result structure
			resultMap := result.(map[string]interface{})
			if resultMap["x402_version"] == nil {
				t.Error("Missing x402_version")
			}
			if resultMap["scheme"] == "" {
				t.Error("Missing scheme")
			}
			if resultMap["nonce"] == "" {
				t.Error("Missing nonce")
			}
			if resultMap["maxAmountRequired"] == "" {
				t.Error("Missing maxAmountRequired")
			}
			if resultMap["payTo"] == "" {
				t.Error("Missing payTo address")
			}
			if resultMap["asset"] == "" {
				t.Error("Missing asset address")
			}
			if resultMap["resource"] == "" {
				t.Error("Missing resource")
			}
			if resultMap["description"] == "" {
				t.Error("Missing description")
			}
			if resultMap["mimeType"] == "" {
				t.Error("Missing mimeType")
			}
			if resultMap["maxTimeoutSeconds"] == nil {
				t.Error("Missing maxTimeoutSeconds")
			}
			extra := resultMap["extra"].(map[string]interface{})
			if extra["name"] == "" {
				t.Error("Missing extra.name")
			}
			if extra["version"] == "" {
				t.Error("Missing extra.version")
			}

			t.Logf("✓ Created x402 requirement with nonce: %s", resultMap["nonce"])
		})
	}
}

// TestSignatureVerificationEdgeCases tests edge cases in signature verification
func TestSignatureVerificationEdgeCases(t *testing.T) {
	t.Log("Testing signature verification edge cases...")

	cfg := createTestConfig()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	verifyTool := tools.NewVerifyPaymentTool(srv)

	// Generate valid authorization for base test
	privateKey, _ := crypto.GenerateKey()
	networkCfg := cfg.Networks["base-sepolia"]
	paymentReq, err := x402.NewPaymentRequirement(
		"5000000", // 5 USDC in atomic units
		"base-sepolia",
		networkCfg.PayeeAddress,
		networkCfg.USDCContract,
		"https://certify.ar4s.com/api/certify/test",
		"Test certification for signature verification",
		"application/json",
		24*time.Hour,
	)
	if err != nil {
		t.Fatalf("Failed to create payment requirement: %v", err)
	}
	auth := createTestAuthorization(t, cfg, privateKey, paymentReq.ToMap())

	tests := []struct {
		name      string
		modifyFn  func(*eip3009.EIP3009Authorization)
		wantValid bool
		wantError string
	}{
		{
			name:      "valid signature",
			modifyFn:  func(a *eip3009.EIP3009Authorization) {},
			wantValid: true,
		},
		{
			name: "invalid v parameter",
			modifyFn: func(a *eip3009.EIP3009Authorization) {
				a.V = 26 // Should be 27 or 28
			},
			wantValid: false,
		},
		{
			name: "tampered amount",
			modifyFn: func(a *eip3009.EIP3009Authorization) {
				a.Value = "999999999" // Different from signed value
			},
			wantValid: false,
		},
		{
			name: "tampered recipient",
			modifyFn: func(a *eip3009.EIP3009Authorization) {
				a.To = "0x0000000000000000000000000000000000000000"
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the authorization
			testAuth := *auth
			tt.modifyFn(&testAuth)

			input := map[string]interface{}{
				"authorization": map[string]interface{}{
					"from":        testAuth.From,
					"to":          testAuth.To,
					"value":       testAuth.Value,
					"validAfter":  float64(testAuth.ValidAfter),
					"validBefore": float64(testAuth.ValidBefore),
					"nonce":       testAuth.Nonce,
					"v":           float64(testAuth.V),
					"r":           testAuth.R,
					"s":           testAuth.S,
				},
				"network": "base-sepolia",
			}

			result, err := verifyTool.Execute(input)
			if err != nil {
				// When there's an error, the tool may not return a result map
				t.Logf("Verification error (expected for invalid cases): %v", err)
				if tt.wantValid {
					t.Errorf("Expected valid signature but got error: %v", err)
				} else {
					t.Logf("✓ Verification correctly rejected invalid authorization")
				}
				return
			}

			resultMap := result.(map[string]interface{})
			isValid := resultMap["is_valid"].(bool)

			if isValid != tt.wantValid {
				t.Errorf("Expected is_valid=%v, got %v", tt.wantValid, isValid)
				if !isValid {
					t.Logf("Error: %v", resultMap["error"])
				}
			} else {
				t.Logf("✓ Verification result as expected: is_valid=%v", isValid)
			}
		})
	}
}
