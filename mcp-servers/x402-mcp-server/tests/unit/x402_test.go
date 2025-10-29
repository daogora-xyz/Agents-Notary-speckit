package unit

import (
	"testing"
	"time"

	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/x402"
)

// TestPaymentRequirement_Generate_Base tests payment requirement generation for Base network
func TestPaymentRequirement_Generate_Base(t *testing.T) {
	// Create payment requirement with 50000 USDC atomic units (0.05 USDC)
	req, err := x402.NewPaymentRequirement(
		"50000",
		"base",
		"0x1234567890123456789012345678901234567890", // Payee address
		"0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", // USDC contract (Base)
		time.Hour*24, // 24 hour validity
	)

	if err != nil {
		t.Fatalf("NewPaymentRequirement failed: %v", err)
	}

	// Validate x402 version
	if req.X402Version != 1 {
		t.Errorf("Expected x402_version 1, got %d", req.X402Version)
	}

	// Validate scheme
	if req.Scheme != "exact" {
		t.Errorf("Expected scheme 'exact', got %s", req.Scheme)
	}

	// Validate network
	if req.Network != "base" {
		t.Errorf("Expected network 'base', got %s", req.Network)
	}

	// Validate amount
	if req.MaxAmountRequired != "50000" {
		t.Errorf("Expected maxAmountRequired '50000', got %s", req.MaxAmountRequired)
	}

	// Validate payee address format
	if req.Payee != "0x1234567890123456789012345678901234567890" {
		t.Errorf("Expected payee address, got %s", req.Payee)
	}

	// Validate asset (USDC contract)
	if req.Asset != "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913" {
		t.Errorf("Expected Base USDC contract, got %s", req.Asset)
	}

	// Validate valid_until is in the future
	validUntil, err := time.Parse(time.RFC3339, req.ValidUntil)
	if err != nil {
		t.Fatalf("Failed to parse valid_until: %v", err)
	}

	if validUntil.Before(time.Now()) {
		t.Error("valid_until should be in the future")
	}

	// Validate nonce is hex-encoded
	if len(req.Nonce) < 3 || req.Nonce[:2] != "0x" {
		t.Errorf("Expected nonce to be hex-encoded (0x-prefixed), got %s", req.Nonce)
	}
}

// TestPaymentRequirement_NonceUniqueness tests that multiple calls generate different nonces
func TestPaymentRequirement_NonceUniqueness(t *testing.T) {
	payee := "0x1234567890123456789012345678901234567890"
	asset := "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
	validity := time.Hour * 24

	// Generate first payment requirement
	req1, err := x402.NewPaymentRequirement("100000", "base", payee, asset, validity)
	if err != nil {
		t.Fatalf("First NewPaymentRequirement failed: %v", err)
	}

	// Small delay to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Generate second payment requirement
	req2, err := x402.NewPaymentRequirement("100000", "base", payee, asset, validity)
	if err != nil {
		t.Fatalf("Second NewPaymentRequirement failed: %v", err)
	}

	// Nonces should be different
	if req1.Nonce == req2.Nonce {
		t.Errorf("Expected different nonces, both were: %s", req1.Nonce)
	}

	t.Logf("Nonce 1: %s", req1.Nonce)
	t.Logf("Nonce 2: %s", req2.Nonce)
}

// TestPaymentRequirement_InvalidNetwork tests error handling for unsupported networks
func TestPaymentRequirement_InvalidNetwork(t *testing.T) {
	_, err := x402.NewPaymentRequirement(
		"50000",
		"ethereum", // Unsupported network
		"0x1234567890123456789012345678901234567890",
		"0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
		time.Hour*24,
	)

	if err == nil {
		t.Error("Expected error for unsupported network, got nil")
	}

	t.Logf("Got expected error: %v", err)
}

// TestPaymentRequirement_InvalidAmount tests error handling for invalid amounts
func TestPaymentRequirement_InvalidAmount(t *testing.T) {
	testCases := []struct {
		amount      string
		description string
	}{
		{"0", "zero amount"},
		{"-1000", "negative amount"},
		{"abc", "non-numeric amount"},
		{"", "empty amount"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			_, err := x402.NewPaymentRequirement(
				tc.amount,
				"base",
				"0x1234567890123456789012345678901234567890",
				"0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
				time.Hour*24,
			)

			if err == nil {
				t.Errorf("Expected error for %s, got nil", tc.description)
			}
		})
	}
}

// TestPaymentRequirement_InvalidAddress tests error handling for malformed addresses
func TestPaymentRequirement_InvalidAddress(t *testing.T) {
	testCases := []struct {
		address     string
		description string
	}{
		{"0x123", "too short"},
		{"1234567890123456789012345678901234567890", "missing 0x prefix"},
		{"0xZZZZ567890123456789012345678901234567890", "invalid hex characters"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			_, err := x402.NewPaymentRequirement(
				"50000",
				"base",
				tc.address, // Invalid payee address
				"0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
				time.Hour*24,
			)

			if err == nil {
				t.Errorf("Expected error for address %s, got nil", tc.description)
			}
		})
	}
}

// TestPaymentRequirement_MultipleNetworks tests payment requirement generation for all supported networks
func TestPaymentRequirement_MultipleNetworks(t *testing.T) {
	networks := map[string]string{
		"base":         "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
		"base-sepolia": "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		"arbitrum":     "0xaf88d065e77c8cC2239327C5EDb3A432268e5831",
	}

	payee := "0x1234567890123456789012345678901234567890"
	validity := time.Hour * 24

	for network, usdcContract := range networks {
		t.Run(network, func(t *testing.T) {
			req, err := x402.NewPaymentRequirement(
				"100000",
				network,
				payee,
				usdcContract,
				validity,
			)

			if err != nil {
				t.Fatalf("NewPaymentRequirement failed for %s: %v", network, err)
			}

			if req.Network != network {
				t.Errorf("Expected network %s, got %s", network, req.Network)
			}

			if req.Asset != usdcContract {
				t.Errorf("Expected asset %s, got %s", usdcContract, req.Asset)
			}
		})
	}
}

// TestPaymentRequirement_ToJSON tests JSON serialization
func TestPaymentRequirement_ToJSON(t *testing.T) {
	req, err := x402.NewPaymentRequirement(
		"50000",
		"base",
		"0x1234567890123456789012345678901234567890",
		"0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
		time.Hour*24,
	)

	if err != nil {
		t.Fatalf("NewPaymentRequirement failed: %v", err)
	}

	jsonData, err := req.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Verify JSON contains required fields
	jsonStr := string(jsonData)
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
		if !contains(jsonStr, field) {
			t.Errorf("JSON missing required field: %s", field)
		}
	}

	t.Logf("Generated JSON: %s", jsonStr)
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
