package contract

import (
	"bytes"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/config"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/logger"
	x402server "github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/server"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/tools"
)

// TestSettlePayment_ToolSchema validates the tool schema per contract
func TestSettlePayment_ToolSchema(t *testing.T) {
	cfg := createTestConfigForSettlement()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create the tool
	tool := tools.NewSettlePaymentTool(srv)

	// Validate tool name
	expectedName := "settle_payment"
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

// TestSettlePayment_Execute_SuccessfulSettlement tests successful settlement flow
func TestSettlePayment_Execute_SuccessfulSettlement(t *testing.T) {
	// Create mock facilitator server
	facilitator := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		// Return success response
		response := map[string]interface{}{
			"status":       "settled",
			"tx_hash":      "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"block_number": 12345678,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer facilitator.Close()

	cfg := createTestConfigForSettlement()
	baseNet := cfg.Networks["base"]
	baseNet.FacilitatorURL = facilitator.URL
	cfg.Networks["base"] = baseNet
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tool := tools.NewSettlePaymentTool(srv)

	// Generate valid signature for testing
	privateKey, fromAddr, err := createTestPrivateKeyAndAddress()
	if err != nil {
		t.Fatalf("Failed to create test private key: %v", err)
	}

	toAddr := common.HexToAddress("0x2222222222222222222222222222222222222222")
	value := big.NewInt(50000)
	now := time.Now().Unix()
	validAfter := big.NewInt(now - 3600)
	validBefore := big.NewInt(now + 3600)
	var nonce [32]byte
	copy(nonce[:], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})

	chainID := big.NewInt(8453) // Base mainnet
	usdcContract := common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913")

	v, r, s, err := generateValidSignature(privateKey, fromAddr, toAddr, value, validAfter, validBefore, nonce, chainID, usdcContract)
	if err != nil {
		t.Fatalf("Failed to generate valid signature: %v", err)
	}

	// Prepare input with valid authorization
	input := map[string]interface{}{
		"authorization": map[string]interface{}{
			"from":        fromAddr.Hex(),
			"to":          toAddr.Hex(),
			"value":       "50000",
			"validAfter":  float64(validAfter.Int64()),
			"validBefore": float64(validBefore.Int64()),
			"nonce":       common.BytesToHash(nonce[:]).Hex(),
			"v":           float64(v),
			"r":           common.BytesToHash(r.Bytes()).Hex(),
			"s":           common.BytesToHash(s.Bytes()).Hex(),
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
	if _, exists := resultMap["status"]; !exists {
		t.Error("Result missing 'status' field")
	}

	if _, exists := resultMap["tx_hash"]; !exists {
		t.Error("Result missing 'tx_hash' field")
	}

	if _, exists := resultMap["block_number"]; !exists {
		t.Error("Result missing 'block_number' field")
	}

	// Verify settlement success
	status, ok := resultMap["status"].(string)
	if !ok {
		t.Fatal("status should be string")
	}

	if status != "settled" {
		t.Errorf("Expected status 'settled', got '%s'", status)
	}

	txHash, ok := resultMap["tx_hash"].(string)
	if !ok {
		t.Fatal("tx_hash should be string")
	}

	if txHash != "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890" {
		t.Errorf("Unexpected tx_hash: %s", txHash)
	}

	t.Logf("Successful settlement result: %+v", resultMap)
}

// TestSettlePayment_Execute_FacilitatorError tests facilitator error handling
func TestSettlePayment_Execute_FacilitatorError(t *testing.T) {
	// Create mock facilitator that returns error
	facilitator := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status": "failed",
			"error":  "insufficient balance",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
	}))
	defer facilitator.Close()

	cfg := createTestConfigForSettlement()
	baseNet := cfg.Networks["base"]
	baseNet.FacilitatorURL = facilitator.URL
	cfg.Networks["base"] = baseNet
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tool := tools.NewSettlePaymentTool(srv)

	// Generate valid signature for testing
	privateKey, fromAddr, err := createTestPrivateKeyAndAddress()
	if err != nil {
		t.Fatalf("Failed to create test private key: %v", err)
	}

	toAddr := common.HexToAddress("0x2222222222222222222222222222222222222222")
	value := big.NewInt(50000)
	now := time.Now().Unix()
	validAfter := big.NewInt(now - 3600)
	validBefore := big.NewInt(now + 3600)
	var nonce [32]byte
	copy(nonce[:], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2})

	chainID := big.NewInt(8453) // Base mainnet
	usdcContract := common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913")

	v, r, s, err := generateValidSignature(privateKey, fromAddr, toAddr, value, validAfter, validBefore, nonce, chainID, usdcContract)
	if err != nil {
		t.Fatalf("Failed to generate valid signature: %v", err)
	}

	input := map[string]interface{}{
		"authorization": map[string]interface{}{
			"from":        fromAddr.Hex(),
			"to":          toAddr.Hex(),
			"value":       "50000",
			"validAfter":  float64(validAfter.Int64()),
			"validBefore": float64(validBefore.Int64()),
			"nonce":       common.BytesToHash(nonce[:]).Hex(),
			"v":           float64(v),
			"r":           common.BytesToHash(r.Bytes()).Hex(),
			"s":           common.BytesToHash(s.Bytes()).Hex(),
		},
		"network": "base",
	}

	result, err := tool.Execute(input)
	if err != nil {
		// Error is acceptable for facilitator failures
		t.Logf("Tool correctly returned error for facilitator failure: %v", err)
		return
	}

	// If no error, verify result indicates failure
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result should be a map")
	}

	status, _ := resultMap["status"].(string)
	if status != "failed" {
		t.Errorf("Expected status 'failed', got '%s'", status)
	}

	if _, exists := resultMap["error"]; !exists {
		t.Error("Result missing 'error' field for failed status")
	}

	t.Logf("Facilitator error result: %+v", resultMap)
}

// TestSettlePayment_Execute_Idempotency tests idempotency via nonce caching
func TestSettlePayment_Execute_Idempotency(t *testing.T) {
	callCount := 0

	// Create mock facilitator that counts calls
	facilitator := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := map[string]interface{}{
			"status":       "settled",
			"tx_hash":      "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"block_number": 12345678,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer facilitator.Close()

	cfg := createTestConfigForSettlement()
	baseNet := cfg.Networks["base"]
	baseNet.FacilitatorURL = facilitator.URL
	cfg.Networks["base"] = baseNet
	cfg.Cache.SettlementTTLMinutes = 10
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tool := tools.NewSettlePaymentTool(srv)

	// Generate valid signature for testing
	privateKey, fromAddr, err := createTestPrivateKeyAndAddress()
	if err != nil {
		t.Fatalf("Failed to create test private key: %v", err)
	}

	toAddr := common.HexToAddress("0x2222222222222222222222222222222222222222")
	value := big.NewInt(50000)
	now := time.Now().Unix()
	validAfter := big.NewInt(now - 3600)
	validBefore := big.NewInt(now + 3600)
	var nonce [32]byte
	copy(nonce[:], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3})

	chainID := big.NewInt(8453) // Base mainnet
	usdcContract := common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913")

	v, r, s, err := generateValidSignature(privateKey, fromAddr, toAddr, value, validAfter, validBefore, nonce, chainID, usdcContract)
	if err != nil {
		t.Fatalf("Failed to generate valid signature: %v", err)
	}

	input := map[string]interface{}{
		"authorization": map[string]interface{}{
			"from":        fromAddr.Hex(),
			"to":          toAddr.Hex(),
			"value":       "50000",
			"validAfter":  float64(validAfter.Int64()),
			"validBefore": float64(validBefore.Int64()),
			"nonce":       common.BytesToHash(nonce[:]).Hex(),
			"v":           float64(v),
			"r":           common.BytesToHash(r.Bytes()).Hex(),
			"s":           common.BytesToHash(s.Bytes()).Hex(),
		},
		"network": "base",
	}

	// First call - should hit facilitator
	result1, err := tool.Execute(input)
	if err != nil {
		t.Fatalf("First settlement failed: %v", err)
	}

	// Second call with same nonce - should return cached result
	result2, err := tool.Execute(input)
	if err != nil {
		t.Fatalf("Second settlement failed: %v", err)
	}

	// Verify both results match
	result1Map := result1.(map[string]interface{})
	result2Map := result2.(map[string]interface{})

	if result1Map["tx_hash"] != result2Map["tx_hash"] {
		t.Error("Cached result tx_hash mismatch")
	}

	// Verify facilitator was only called once
	if callCount != 1 {
		t.Errorf("Expected 1 facilitator call, got %d (idempotency failed)", callCount)
	}

	t.Logf("Idempotency test passed: facilitator called %d time(s)", callCount)
}

// TestSettlePayment_Execute_NetworkTimeout tests handling of slow facilitator
func TestSettlePayment_Execute_NetworkTimeout(t *testing.T) {
	// Create mock facilitator that delays
	facilitator := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(6 * time.Second) // Exceeds 5-second timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer facilitator.Close()

	cfg := createTestConfigForSettlement()
	baseNet := cfg.Networks["base"]
	baseNet.FacilitatorURL = facilitator.URL
	cfg.Networks["base"] = baseNet
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tool := tools.NewSettlePaymentTool(srv)

	// Generate valid signature for testing
	privateKey, fromAddr, err := createTestPrivateKeyAndAddress()
	if err != nil {
		t.Fatalf("Failed to create test private key: %v", err)
	}

	toAddr := common.HexToAddress("0x2222222222222222222222222222222222222222")
	value := big.NewInt(50000)
	now := time.Now().Unix()
	validAfter := big.NewInt(now - 3600)
	validBefore := big.NewInt(now + 3600)
	var nonce [32]byte
	copy(nonce[:], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4})

	chainID := big.NewInt(8453) // Base mainnet
	usdcContract := common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913")

	v, r, s, err := generateValidSignature(privateKey, fromAddr, toAddr, value, validAfter, validBefore, nonce, chainID, usdcContract)
	if err != nil {
		t.Fatalf("Failed to generate valid signature: %v", err)
	}

	input := map[string]interface{}{
		"authorization": map[string]interface{}{
			"from":        fromAddr.Hex(),
			"to":          toAddr.Hex(),
			"value":       "50000",
			"validAfter":  float64(validAfter.Int64()),
			"validBefore": float64(validBefore.Int64()),
			"nonce":       common.BytesToHash(nonce[:]).Hex(),
			"v":           float64(v),
			"r":           common.BytesToHash(r.Bytes()).Hex(),
			"s":           common.BytesToHash(s.Bytes()).Hex(),
		},
		"network": "base",
	}

	start := time.Now()
	_, err = tool.Execute(input)
	duration := time.Since(start)

	if err == nil {
		t.Error("Expected timeout error")
	}

	// Verify timeout occurred within reasonable time
	if duration > 6*time.Second {
		t.Errorf("Timeout took too long: %v (expected ~5s)", duration)
	}

	t.Logf("Correctly timed out after %v: %v", duration, err)
}

// TestSettlePayment_JSONOutput tests that output is valid JSON
func TestSettlePayment_JSONOutput(t *testing.T) {
	// Create mock facilitator
	facilitator := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":       "settled",
			"tx_hash":      "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"block_number": 12345678,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer facilitator.Close()

	cfg := createTestConfigForSettlement()
	baseNet := cfg.Networks["base"]
	baseNet.FacilitatorURL = facilitator.URL
	cfg.Networks["base"] = baseNet
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tool := tools.NewSettlePaymentTool(srv)

	// Generate valid signature for testing
	privateKey, fromAddr, err := createTestPrivateKeyAndAddress()
	if err != nil {
		t.Fatalf("Failed to create test private key: %v", err)
	}

	toAddr := common.HexToAddress("0x2222222222222222222222222222222222222222")
	value := big.NewInt(50000)
	now := time.Now().Unix()
	validAfter := big.NewInt(now - 3600)
	validBefore := big.NewInt(now + 3600)
	var nonce [32]byte
	copy(nonce[:], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5})

	chainID := big.NewInt(8453) // Base mainnet
	usdcContract := common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913")

	v, r, s, err := generateValidSignature(privateKey, fromAddr, toAddr, value, validAfter, validBefore, nonce, chainID, usdcContract)
	if err != nil {
		t.Fatalf("Failed to generate valid signature: %v", err)
	}

	input := map[string]interface{}{
		"authorization": map[string]interface{}{
			"from":        fromAddr.Hex(),
			"to":          toAddr.Hex(),
			"value":       "50000",
			"validAfter":  float64(validAfter.Int64()),
			"validBefore": float64(validBefore.Int64()),
			"nonce":       common.BytesToHash(nonce[:]).Hex(),
			"v":           float64(v),
			"r":           common.BytesToHash(r.Bytes()).Hex(),
			"s":           common.BytesToHash(s.Bytes()).Hex(),
		},
		"network": "base",
	}

	result, err := tool.Execute(input)
	if err != nil {
		t.Fatalf("Tool execution failed: %v", err)
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

// createTestConfigForSettlement creates test configuration for settlement tests
func createTestConfigForSettlement() *config.Config {
	return &config.Config{
		Networks: map[string]config.NetworkConfig{
			"base": {
				ChainID:        8453,
				USDCContract:   "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
				FacilitatorURL: "https://api.cdp.coinbase.com/x402/base",
				RPCURL:         "https://mainnet.base.org",
				PayeeAddress:   "0x2222222222222222222222222222222222222222",
			},
			"base-sepolia": {
				ChainID:        84532,
				USDCContract:   "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				FacilitatorURL: "https://api.cdp.coinbase.com/x402/base-sepolia",
				RPCURL:         "https://sepolia.base.org",
				PayeeAddress:   "0x2222222222222222222222222222222222222222",
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
