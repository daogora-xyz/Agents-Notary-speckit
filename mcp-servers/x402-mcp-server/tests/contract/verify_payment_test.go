package contract

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/config"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/eip3009"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/logger"
	x402server "github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/server"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/tools"
)

// TestVerifyPayment_ToolSchema validates the tool schema per contract
func TestVerifyPayment_ToolSchema(t *testing.T) {
	cfg := createTestConfigForVerification()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create the tool
	tool := tools.NewVerifyPaymentTool(srv)

	// Validate tool name
	expectedName := "verify_payment"
	if tool.Name() != expectedName {
		t.Errorf("Expected tool name %s, got %s", expectedName, tool.Name())
	}

	// Validate description exists
	if tool.Description() == "" {
		t.Error("Tool description should not be empty")
	}

	// Validate schema
	schema := tool.Schema()
	if schema == nil {
		t.Fatal("Tool schema should not be nil")
	}

	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		t.Fatal("Schema should be a map")
	}

	// Validate schema type
	if schemaMap["type"] != "object" {
		t.Error("Schema type should be 'object'")
	}

	// Validate properties exist
	props, ok := schemaMap["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have properties")
	}

	// Validate 'authorization' property
	authProp, ok := props["authorization"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have 'authorization' property")
	}

	if authProp["type"] != "object" {
		t.Error("authorization property should be object type")
	}

	// Validate 'network' property
	networkProp, ok := props["network"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have 'network' property")
	}

	if networkProp["type"] != "string" {
		t.Error("network property should be string type")
	}

	// Validate required fields
	required, ok := schemaMap["required"].([]string)
	if !ok {
		t.Fatal("Schema should have required fields")
	}

	if len(required) != 2 {
		t.Errorf("Expected 2 required fields, got %d", len(required))
	}
}

// TestVerifyPayment_Execute_ValidSignature tests tool execution with valid signature
func TestVerifyPayment_Execute_ValidSignature(t *testing.T) {
	cfg := createTestConfigForVerification()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tool := tools.NewVerifyPaymentTool(srv)

	// Generate test signature
	privateKey, _ := crypto.GenerateKey()
	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKey)

	toAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	value := big.NewInt(50000)
	now := time.Now().Unix()
	validAfter := big.NewInt(now - 3600)
	validBefore := big.NewInt(now + 3600)
	nonce := [32]byte{}
	copy(nonce[:], []byte("test-nonce"))

	domain := &eip3009.EIP712Domain{
		Name:              "USD Coin",
		Version:           "2",
		ChainID:           big.NewInt(8453),
		VerifyingContract: common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"),
	}

	message := &eip3009.ReceiveWithAuthorizationMessage{
		From:        fromAddress,
		To:          toAddress,
		Value:       value,
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
		Nonce:       nonce,
	}

	typedDataHash, _ := eip3009.TypedDataHash(domain, message)
	signature, _ := crypto.Sign(typedDataHash.Bytes(), privateKey)

	v := signature[64] + 27
	r := common.BytesToHash(signature[0:32]).Hex()
	s := common.BytesToHash(signature[32:64]).Hex()

	// Prepare input
	input := map[string]interface{}{
		"authorization": map[string]interface{}{
			"from":        fromAddress.Hex(),
			"to":          toAddress.Hex(),
			"value":       "50000",
			"validAfter":  float64(validAfter.Uint64()),
			"validBefore": float64(validBefore.Uint64()),
			"nonce":       common.BytesToHash(nonce[:]).Hex(),
			"v":           float64(v),
			"r":           r,
			"s":           s,
		},
		"network": "base",
	}

	// Execute tool
	result, err := tool.Execute(input)
	if err != nil {
		t.Fatalf("Tool execution failed: %v", err)
	}

	// Validate result
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result should be a map")
	}

	// Check required fields
	if _, exists := resultMap["is_valid"]; !exists {
		t.Error("Result missing 'is_valid' field")
	}

	if _, exists := resultMap["signer_address"]; !exists {
		t.Error("Result missing 'signer_address' field")
	}

	// Verify signature is valid
	isValid, ok := resultMap["is_valid"].(bool)
	if !ok {
		t.Fatal("is_valid should be boolean")
	}

	if !isValid {
		t.Errorf("Expected valid signature, got invalid. Error: %v", resultMap["error"])
	}

	// Verify recovered signer matches
	signerAddr, ok := resultMap["signer_address"].(string)
	if !ok {
		t.Fatal("signer_address should be string")
	}

	if signerAddr != fromAddress.Hex() {
		t.Errorf("Expected signer %s, got %s", fromAddress.Hex(), signerAddr)
	}

	t.Logf("Valid signature verification result: %+v", resultMap)
}

// TestVerifyPayment_Execute_InvalidSignature tests invalid signature detection
func TestVerifyPayment_Execute_InvalidSignature(t *testing.T) {
	cfg := createTestConfigForVerification()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tool := tools.NewVerifyPaymentTool(srv)

	// Create input with invalid signature (all zeros)
	now := time.Now().Unix()
	input := map[string]interface{}{
		"authorization": map[string]interface{}{
			"from":        "0x0000000000000000000000000000000000000001",
			"to":          "0x1234567890123456789012345678901234567890",
			"value":       "50000",
			"validAfter":  float64(now - 3600),
			"validBefore": float64(now + 3600),
			"nonce":       "0x0000000000000000000000000000000000000000000000000000000000000001",
			"v":           float64(27),
			"r":           "0x0000000000000000000000000000000000000000000000000000000000000001",
			"s":           "0x0000000000000000000000000000000000000000000000000000000000000001",
		},
		"network": "base",
	}

	// Execute tool
	result, err := tool.Execute(input)
	if err != nil {
		// It's OK if execution returns an error for invalid signature
		t.Logf("Tool correctly returned error for invalid signature: %v", err)
		return
	}

	// If no error, check result indicates invalid
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result should be a map")
	}

	isValid, ok := resultMap["is_valid"].(bool)
	if !ok {
		t.Fatal("is_valid should be boolean")
	}

	if isValid {
		t.Error("Expected invalid signature to be detected")
	}

	t.Logf("Invalid signature correctly detected: %+v", resultMap)
}

// TestVerifyPayment_Execute_ExpiredAuthorization tests expired time bounds
func TestVerifyPayment_Execute_ExpiredAuthorization(t *testing.T) {
	cfg := createTestConfigForVerification()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tool := tools.NewVerifyPaymentTool(srv)

	// Generate valid signature but with expired time bounds
	privateKey, _ := crypto.GenerateKey()
	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKey)

	toAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	value := big.NewInt(50000)
	now := time.Now().Unix()
	validAfter := big.NewInt(now - 7200)  // 2 hours ago
	validBefore := big.NewInt(now - 3600) // 1 hour ago (expired!)
	nonce := [32]byte{}
	copy(nonce[:], []byte("expired-test"))

	domain := &eip3009.EIP712Domain{
		Name:              "USD Coin",
		Version:           "2",
		ChainID:           big.NewInt(8453),
		VerifyingContract: common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"),
	}

	message := &eip3009.ReceiveWithAuthorizationMessage{
		From:        fromAddress,
		To:          toAddress,
		Value:       value,
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
		Nonce:       nonce,
	}

	typedDataHash, _ := eip3009.TypedDataHash(domain, message)
	signature, _ := crypto.Sign(typedDataHash.Bytes(), privateKey)

	v := signature[64] + 27
	r := common.BytesToHash(signature[0:32]).Hex()
	s := common.BytesToHash(signature[32:64]).Hex()

	input := map[string]interface{}{
		"authorization": map[string]interface{}{
			"from":        fromAddress.Hex(),
			"to":          toAddress.Hex(),
			"value":       "50000",
			"validAfter":  float64(validAfter.Uint64()),
			"validBefore": float64(validBefore.Uint64()),
			"nonce":       common.BytesToHash(nonce[:]).Hex(),
			"v":           float64(v),
			"r":           r,
			"s":           s,
		},
		"network": "base",
	}

	// Execute tool
	result, err := tool.Execute(input)
	if err != nil {
		t.Logf("Tool correctly returned error for expired authorization: %v", err)
		return
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result should be a map")
	}

	isValid, _ := resultMap["is_valid"].(bool)
	if isValid {
		t.Error("Expected expired authorization to be rejected")
	}

	t.Logf("Expired authorization correctly rejected: %+v", resultMap)
}

// TestVerifyPayment_JSONOutput tests that output is valid JSON
func TestVerifyPayment_JSONOutput(t *testing.T) {
	cfg := createTestConfigForVerification()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tool := tools.NewVerifyPaymentTool(srv)

	// Create test input with invalid signature
	now := time.Now().Unix()
	input := map[string]interface{}{
		"authorization": map[string]interface{}{
			"from":        "0x0000000000000000000000000000000000000001",
			"to":          "0x1234567890123456789012345678901234567890",
			"value":       "50000",
			"validAfter":  float64(now - 3600),
			"validBefore": float64(now + 3600),
			"nonce":       "0x0000000000000000000000000000000000000000000000000000000000000001",
			"v":           float64(27),
			"r":           "0x0000000000000000000000000000000000000000000000000000000000000001",
			"s":           "0x0000000000000000000000000000000000000000000000000000000000000001",
		},
		"network": "base",
	}

	result, err := tool.Execute(input)
	if err != nil {
		// Even on error, try to marshal
		t.Logf("Tool returned error: %v", err)
	}

	if result == nil {
		t.Skip("Tool returned nil result, skipping JSON test")
	}

	// Convert result to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal result to JSON: %v", err)
	}

	// Verify it's valid JSON by unmarshaling
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Result is not valid JSON: %v", err)
	}

	t.Logf("JSON output: %s", string(jsonData))
}

// createTestConfigForVerification creates test configuration
func createTestConfigForVerification() *config.Config {
	return &config.Config{
		Networks: map[string]config.NetworkConfig{
			"base": {
				ChainID:        8453,
				USDCContract:   "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
				FacilitatorURL: "https://api.cdp.coinbase.com",
				RPCURL:         "https://mainnet.base.org",
				PayeeAddress:   "0x1234567890123456789012345678901234567890",
			},
			"base-sepolia": {
				ChainID:        84532,
				USDCContract:   "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				FacilitatorURL: "https://api.cdp.coinbase.com",
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
