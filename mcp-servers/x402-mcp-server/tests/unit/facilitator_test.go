package unit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/config"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/eip3009"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/facilitator"
)

// TestFacilitatorClient_ConstructRequest tests HTTP POST request body construction
func TestFacilitatorClient_ConstructRequest(t *testing.T) {
	cfg := &config.Config{
		Networks: map[string]config.NetworkConfig{
			"base": {
				ChainID:        8453,
				USDCContract:   "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
				FacilitatorURL: "https://api.cdp.coinbase.com/x402/base",
			},
		},
	}

	client := facilitator.NewClient(cfg, 5*time.Second)

	// Create test authorization
	auth := &eip3009.EIP3009Authorization{
		From:        "0x1111111111111111111111111111111111111111",
		To:          "0x2222222222222222222222222222222222222222",
		Value:       "50000",
		ValidAfter:  1700000000,
		ValidBefore: 1700003600,
		Nonce:       "0x0000000000000000000000000000000000000000000000000000000000000001",
		V:           27,
		R:           "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		S:           "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
	}

	// Construct request
	requestBody, err := client.BuildSettlementRequest(auth, "base")
	if err != nil {
		t.Fatalf("Failed to construct request: %v", err)
	}

	// Verify request body is valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(requestBody, &parsed); err != nil {
		t.Fatalf("Request body is not valid JSON: %v", err)
	}

	// Verify required fields exist
	requiredFields := []string{"from", "to", "value", "validAfter", "validBefore", "nonce", "v", "r", "s"}
	for _, field := range requiredFields {
		if _, exists := parsed[field]; !exists {
			t.Errorf("Request body missing required field: %s", field)
		}
	}

	t.Logf("Request body: %s", string(requestBody))
}

// TestFacilitatorClient_SuccessResponse tests parsing successful settlement response
func TestFacilitatorClient_SuccessResponse(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	defer server.Close()

	// Create client with mock server URL
	cfg := &config.Config{
		Networks: map[string]config.NetworkConfig{
			"base": {
				ChainID:        8453,
				USDCContract:   "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
				FacilitatorURL: server.URL,
			},
		},
	}

	client := facilitator.NewClient(cfg, 5*time.Second)

	auth := &eip3009.EIP3009Authorization{
		From:        "0x1111111111111111111111111111111111111111",
		To:          "0x2222222222222222222222222222222222222222",
		Value:       "50000",
		ValidAfter:  1700000000,
		ValidBefore: 1700003600,
		Nonce:       "0x0000000000000000000000000000000000000000000000000000000000000001",
		V:           27,
		R:           "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		S:           "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
	}

	// Submit settlement
	response, err := client.SubmitSettlement(auth, "base")
	if err != nil {
		t.Fatalf("Settlement submission failed: %v", err)
	}

	// Verify response
	if response.Status != "settled" {
		t.Errorf("Expected status 'settled', got '%s'", response.Status)
	}

	if response.TxHash != "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890" {
		t.Errorf("Unexpected tx_hash: %s", response.TxHash)
	}

	if response.BlockNumber != 12345678 {
		t.Errorf("Expected block_number 12345678, got %d", response.BlockNumber)
	}

	t.Logf("Success response: status=%s, tx_hash=%s, block=%d",
		response.Status, response.TxHash, response.BlockNumber)
}

// TestFacilitatorClient_ErrorResponse tests handling of 400 Bad Request
func TestFacilitatorClient_ErrorResponse_BadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status": "failed",
			"error":  "invalid signature",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &config.Config{
		Networks: map[string]config.NetworkConfig{
			"base": {
				ChainID:        8453,
				USDCContract:   "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
				FacilitatorURL: server.URL,
			},
		},
	}

	client := facilitator.NewClient(cfg, 5*time.Second)

	auth := &eip3009.EIP3009Authorization{
		From:        "0x1111111111111111111111111111111111111111",
		To:          "0x2222222222222222222222222222222222222222",
		Value:       "50000",
		ValidAfter:  1700000000,
		ValidBefore: 1700003600,
		Nonce:       "0x0000000000000000000000000000000000000000000000000000000000000001",
		V:           27,
		R:           "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		S:           "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
	}

	response, err := client.SubmitSettlement(auth, "base")
	if err != nil {
		// Error is acceptable for bad request
		t.Logf("Correctly returned error for 400 Bad Request: %v", err)
		return
	}

	// If no error, verify response indicates failure
	if response.Status != "failed" {
		t.Errorf("Expected status 'failed', got '%s'", response.Status)
	}

	if response.Error == "" {
		t.Error("Expected error message for failed status")
	}

	t.Logf("Error response: status=%s, error=%s", response.Status, response.Error)
}

// TestFacilitatorClient_ErrorResponse_ServerError tests handling of 500 Server Error
func TestFacilitatorClient_ErrorResponse_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	cfg := &config.Config{
		Networks: map[string]config.NetworkConfig{
			"base": {
				ChainID:        8453,
				USDCContract:   "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
				FacilitatorURL: server.URL,
			},
		},
	}

	client := facilitator.NewClient(cfg, 5*time.Second)

	auth := &eip3009.EIP3009Authorization{
		From:        "0x1111111111111111111111111111111111111111",
		To:          "0x2222222222222222222222222222222222222222",
		Value:       "50000",
		ValidAfter:  1700000000,
		ValidBefore: 1700003600,
		Nonce:       "0x0000000000000000000000000000000000000000000000000000000000000001",
		V:           27,
		R:           "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		S:           "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
	}

	_, err := client.SubmitSettlement(auth, "base")
	if err == nil {
		t.Error("Expected error for 500 Server Error")
	}

	t.Logf("Correctly returned error for 500 Server Error: %v", err)
}

// TestFacilitatorClient_Timeout tests handling of request timeout
func TestFacilitatorClient_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay longer than client timeout
		time.Sleep(6 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.Config{
		Networks: map[string]config.NetworkConfig{
			"base": {
				ChainID:        8453,
				USDCContract:   "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
				FacilitatorURL: server.URL,
			},
		},
	}

	// Create client with 5-second timeout
	client := facilitator.NewClient(cfg, 5*time.Second)

	auth := &eip3009.EIP3009Authorization{
		From:        "0x1111111111111111111111111111111111111111",
		To:          "0x2222222222222222222222222222222222222222",
		Value:       "50000",
		ValidAfter:  1700000000,
		ValidBefore: 1700003600,
		Nonce:       "0x0000000000000000000000000000000000000000000000000000000000000001",
		V:           27,
		R:           "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		S:           "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
	}

	start := time.Now()
	_, err := client.SubmitSettlement(auth, "base")
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

// TestFacilitatorClient_IdempotencyCache tests caching of settlement results
func TestFacilitatorClient_IdempotencyCache(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	defer server.Close()

	cfg := &config.Config{
		Networks: map[string]config.NetworkConfig{
			"base": {
				ChainID:        8453,
				USDCContract:   "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
				FacilitatorURL: server.URL,
			},
		},
		Cache: config.CacheConfig{
			SettlementTTLMinutes: 10,
		},
	}

	client := facilitator.NewClient(cfg, 5*time.Second)

	auth := &eip3009.EIP3009Authorization{
		From:        "0x1111111111111111111111111111111111111111",
		To:          "0x2222222222222222222222222222222222222222",
		Value:       "50000",
		ValidAfter:  1700000000,
		ValidBefore: 1700003600,
		Nonce:       "0x0000000000000000000000000000000000000000000000000000000000000001",
		V:           27,
		R:           "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		S:           "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
	}

	// First call - should hit facilitator
	response1, err := client.SubmitSettlement(auth, "base")
	if err != nil {
		t.Fatalf("First settlement failed: %v", err)
	}

	// Second call with same nonce - should return cached result
	response2, err := client.SubmitSettlement(auth, "base")
	if err != nil {
		t.Fatalf("Second settlement failed: %v", err)
	}

	// Verify both responses match
	if response1.TxHash != response2.TxHash {
		t.Error("Cached response tx_hash mismatch")
	}

	if response1.BlockNumber != response2.BlockNumber {
		t.Error("Cached response block_number mismatch")
	}

	// Verify facilitator was only called once
	if callCount != 1 {
		t.Errorf("Expected 1 facilitator call, got %d (caching failed)", callCount)
	}

	t.Logf("Idempotency test passed: facilitator called %d time(s)", callCount)
}

// TestFacilitatorClient_PendingResponse tests handling of pending settlement
func TestFacilitatorClient_PendingResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":      "pending",
			"retry_after": 30,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &config.Config{
		Networks: map[string]config.NetworkConfig{
			"base": {
				ChainID:        8453,
				USDCContract:   "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
				FacilitatorURL: server.URL,
			},
		},
	}

	client := facilitator.NewClient(cfg, 5*time.Second)

	auth := &eip3009.EIP3009Authorization{
		From:        "0x1111111111111111111111111111111111111111",
		To:          "0x2222222222222222222222222222222222222222",
		Value:       "50000",
		ValidAfter:  1700000000,
		ValidBefore: 1700003600,
		Nonce:       "0x0000000000000000000000000000000000000000000000000000000000000001",
		V:           27,
		R:           "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		S:           "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
	}

	response, err := client.SubmitSettlement(auth, "base")
	if err != nil {
		t.Fatalf("Pending settlement request failed: %v", err)
	}

	if response.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", response.Status)
	}

	if response.RetryAfter != 30 {
		t.Errorf("Expected retry_after=30, got %d", response.RetryAfter)
	}

	t.Logf("Pending response: status=%s, retry_after=%d", response.Status, response.RetryAfter)
}
