package contract

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/config"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/logger"
	x402server "github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/server"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/tools"
)

// TestCreatePaymentRequirement_ToolSchema validates the tool conforms to the contract
func TestCreatePaymentRequirement_ToolSchema(t *testing.T) {
	cfg := createTestConfigForPayment()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create the tool
	tool := tools.NewCreatePaymentRequirementTool(srv)

	// Validate tool name
	expectedName := "create_payment_requirement"
	if tool.Name() != expectedName {
		t.Errorf("Expected tool name %s, got %s", expectedName, tool.Name())
	}

	// Validate description exists and is non-empty
	if tool.Description() == "" {
		t.Error("Tool description should not be empty")
	}

	// Validate schema
	schema := tool.Schema()
	if schema == nil {
		t.Fatal("Tool schema should not be nil")
	}

	// Validate schema is a map
	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		t.Fatal("Schema should be a map")
	}

	// Validate schema has required top-level fields
	if schemaMap["type"] != "object" {
		t.Error("Schema type should be 'object'")
	}

	// Validate properties
	props, ok := schemaMap["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have properties")
	}

	// Validate 'amount' property
	amountProp, ok := props["amount"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have 'amount' property")
	}

	if amountProp["type"] != "string" {
		t.Error("amount property should be string type")
	}

	// Validate 'network' property
	networkProp, ok := props["network"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have 'network' property")
	}

	if networkProp["type"] != "string" {
		t.Error("network property should be string type")
	}

	// Validate enum values for network
	enum, ok := networkProp["enum"].([]interface{})
	if !ok {
		t.Fatal("network should have enum values")
	}

	expectedNetworks := []string{"base", "base-sepolia", "arbitrum"}
	if len(enum) != len(expectedNetworks) {
		t.Errorf("Expected %d network options, got %d", len(expectedNetworks), len(enum))
	}

	// Validate required fields
	required, ok := schemaMap["required"].([]interface{})
	if !ok {
		t.Fatal("Schema should have required fields")
	}

	if len(required) != 2 {
		t.Errorf("Expected 2 required fields, got %d", len(required))
	}
}

// TestCreatePaymentRequirement_Execute_ValidInput tests tool execution with valid input
func TestCreatePaymentRequirement_Execute_ValidInput(t *testing.T) {
	cfg := createTestConfigForPayment()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tool := tools.NewCreatePaymentRequirementTool(srv)

	// Prepare input
	input := map[string]interface{}{
		"amount":  "50000",
		"network": "base",
	}

	// Execute tool
	result, err := tool.Execute(input)
	if err != nil {
		t.Fatalf("Tool execution failed: %v", err)
	}

	// Validate result is not empty
	if result == nil {
		t.Fatal("Tool result should not be nil")
	}

	// Parse result as map
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result should be a map")
	}

	// Validate all required output fields are present
	requiredFields := []string{
		"x402_version",
		"scheme",
		"network",
		"maxAmountRequired",
		"payee",
		"valid_until",
		"nonce",
		"asset",
	}

	for _, field := range requiredFields {
		if _, exists := resultMap[field]; !exists {
			t.Errorf("Result missing required field: %s", field)
		}
	}

	// Validate specific values
	if resultMap["x402_version"] != float64(1) { // JSON numbers are float64
		t.Errorf("Expected x402_version 1, got %v", resultMap["x402_version"])
	}

	if resultMap["scheme"] != "exact" {
		t.Errorf("Expected scheme 'exact', got %v", resultMap["scheme"])
	}

	if resultMap["network"] != "base" {
		t.Errorf("Expected network 'base', got %v", resultMap["network"])
	}

	if resultMap["maxAmountRequired"] != "50000" {
		t.Errorf("Expected maxAmountRequired '50000', got %v", resultMap["maxAmountRequired"])
	}

	t.Logf("Tool result: %+v", resultMap)
}

// TestCreatePaymentRequirement_Execute_InvalidAmount tests error handling for invalid amount
func TestCreatePaymentRequirement_Execute_InvalidAmount(t *testing.T) {
	cfg := createTestConfigForPayment()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tool := tools.NewCreatePaymentRequirementTool(srv)

	testCases := []struct {
		amount      string
		description string
	}{
		{"0", "zero amount"},
		{"-1000", "negative amount"},
		{"abc", "non-numeric"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			input := map[string]interface{}{
				"amount":  tc.amount,
				"network": "base",
			}

			_, err := tool.Execute(input)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tc.description)
			}
		})
	}
}

// TestCreatePaymentRequirement_Execute_InvalidNetwork tests error handling for unsupported network
func TestCreatePaymentRequirement_Execute_InvalidNetwork(t *testing.T) {
	cfg := createTestConfigForPayment()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tool := tools.NewCreatePaymentRequirementTool(srv)

	input := map[string]interface{}{
		"amount":  "50000",
		"network": "ethereum", // Unsupported
	}

	_, err = tool.Execute(input)
	if err == nil {
		t.Error("Expected error for unsupported network, got nil")
	}
}

// TestCreatePaymentRequirement_JSONOutput tests that output is valid JSON
func TestCreatePaymentRequirement_JSONOutput(t *testing.T) {
	cfg := createTestConfigForPayment()
	log := logger.New(logger.DEBUG, &bytes.Buffer{})

	srv, err := x402server.NewServer(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tool := tools.NewCreatePaymentRequirementTool(srv)

	input := map[string]interface{}{
		"amount":  "100000",
		"network": "base-sepolia",
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

// createTestConfigForPayment creates a config for payment testing
func createTestConfigForPayment() *config.Config {
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
			"arbitrum": {
				ChainID:        42161,
				USDCContract:   "0xaf88d065e77c8cC2239327C5EDb3A432268e5831",
				FacilitatorURL: "https://api.cdp.coinbase.com",
				RPCURL:         "https://arb1.arbitrum.io/rpc",
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
