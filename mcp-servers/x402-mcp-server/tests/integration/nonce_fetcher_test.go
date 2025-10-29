package integration

import (
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/rpc"
)

// TestNonceFetcher_RealRPC tests against Base Sepolia testnet
// Skipped by default unless RPC_TEST_ENABLED=1 is set
func TestNonceFetcher_RealRPC(t *testing.T) {
	if os.Getenv("RPC_TEST_ENABLED") != "1" {
		t.Skip("Skipping RPC integration test (set RPC_TEST_ENABLED=1 to run)")
	}

	// Base Sepolia public RPC
	rpcURL := "https://sepolia.base.org"

	fetcher, err := rpc.NewNonceFetcher(rpcURL)
	if err != nil {
		t.Fatalf("Failed to create nonce fetcher: %v", err)
	}
	defer fetcher.Close()

	// Verify connection
	if err := fetcher.VerifyConnection(); err != nil {
		t.Fatalf("Connection verification failed: %v", err)
	}

	// Get chain ID
	chainID, err := fetcher.GetChainID()
	if err != nil {
		t.Fatalf("Failed to get chain ID: %v", err)
	}

	// Base Sepolia chain ID is 84532
	expectedChainID := big.NewInt(84532)
	if chainID.Cmp(expectedChainID) != 0 {
		t.Errorf("Expected chain ID %s, got %s", expectedChainID.String(), chainID.String())
	}

	t.Logf("Connected to Base Sepolia (chain ID: %s)", chainID.String())
}

func TestNonceFetcher_GetNonce_KnownAddress(t *testing.T) {
	if os.Getenv("RPC_TEST_ENABLED") != "1" {
		t.Skip("Skipping RPC integration test (set RPC_TEST_ENABLED=1 to run)")
	}

	rpcURL := "https://sepolia.base.org"

	fetcher, err := rpc.NewNonceFetcher(rpcURL)
	if err != nil {
		t.Fatalf("Failed to create nonce fetcher: %v", err)
	}
	defer fetcher.Close()

	// Use a known address (e.g., zero address or USDC contract)
	// Zero address will always have nonce 0
	zeroAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")

	nonce, err := fetcher.GetNonce(zeroAddress)
	if err != nil {
		t.Fatalf("Failed to get nonce: %v", err)
	}

	// Zero address should have nonce 0
	expectedNonce := big.NewInt(0)
	if nonce.Cmp(expectedNonce) != 0 {
		t.Errorf("Expected nonce %s for zero address, got %s", expectedNonce.String(), nonce.String())
	}

	t.Logf("Nonce for zero address: %s", nonce.String())
}

func TestNonceFetcher_GetNonce_ActiveAddress(t *testing.T) {
	if os.Getenv("RPC_TEST_ENABLED") != "1" {
		t.Skip("Skipping RPC integration test (set RPC_TEST_ENABLED=1 to run)")
	}

	rpcURL := "https://sepolia.base.org"

	fetcher, err := rpc.NewNonceFetcher(rpcURL)
	if err != nil {
		t.Fatalf("Failed to create nonce fetcher: %v", err)
	}
	defer fetcher.Close()

	// Base Sepolia USDC contract (has transaction history)
	usdcContract := common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e")

	nonce, err := fetcher.GetNonce(usdcContract)
	if err != nil {
		t.Fatalf("Failed to get nonce: %v", err)
	}

	// Contract address should have nonce 1 (deployed via contract creation)
	expectedNonce := big.NewInt(1)
	if nonce.Cmp(expectedNonce) != 0 {
		t.Logf("WARNING: Expected nonce %s for USDC contract, got %s (may have changed)",
			expectedNonce.String(), nonce.String())
	}

	t.Logf("Nonce for USDC contract: %s", nonce.String())
}

func TestNonceFetcher_GetNonceAt_HistoricalBlock(t *testing.T) {
	if os.Getenv("RPC_TEST_ENABLED") != "1" {
		t.Skip("Skipping RPC integration test (set RPC_TEST_ENABLED=1 to run)")
	}

	rpcURL := "https://sepolia.base.org"

	fetcher, err := rpc.NewNonceFetcher(rpcURL)
	if err != nil {
		t.Fatalf("Failed to create nonce fetcher: %v", err)
	}
	defer fetcher.Close()

	// Get nonce at genesis block
	zeroAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")
	blockNumber := big.NewInt(0)

	nonce, err := fetcher.GetNonceAt(zeroAddress, blockNumber)
	if err != nil {
		t.Fatalf("Failed to get historical nonce: %v", err)
	}

	// Should always be 0
	expectedNonce := big.NewInt(0)
	if nonce.Cmp(expectedNonce) != 0 {
		t.Errorf("Expected nonce %s at genesis, got %s", expectedNonce.String(), nonce.String())
	}

	t.Logf("Historical nonce at block %s: %s", blockNumber.String(), nonce.String())
}

func TestNonceFetcher_InvalidRPC(t *testing.T) {
	// Test with invalid RPC URL
	_, err := rpc.NewNonceFetcher("http://invalid-rpc-url-that-does-not-exist.com")
	if err == nil {
		t.Error("Expected error for invalid RPC URL")
	}

	t.Logf("Got expected error: %v", err)
}

func TestNonceFetcher_RetryLogic(t *testing.T) {
	if os.Getenv("RPC_TEST_ENABLED") != "1" {
		t.Skip("Skipping RPC integration test (set RPC_TEST_ENABLED=1 to run)")
	}

	// Test with a valid RPC to ensure retry logic doesn't break normal operation
	rpcURL := "https://sepolia.base.org"

	fetcher, err := rpc.NewNonceFetcher(rpcURL)
	if err != nil {
		t.Fatalf("Failed to create nonce fetcher: %v", err)
	}
	defer fetcher.Close()

	zeroAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")

	// This should succeed on first attempt
	nonce, err := fetcher.GetNonce(zeroAddress)
	if err != nil {
		t.Fatalf("Failed to get nonce: %v", err)
	}

	if nonce == nil {
		t.Error("Expected non-nil nonce")
	}

	t.Logf("Retry logic test passed with nonce: %s", nonce.String())
}

func TestNonceFetcher_Concurrent(t *testing.T) {
	if os.Getenv("RPC_TEST_ENABLED") != "1" {
		t.Skip("Skipping RPC integration test (set RPC_TEST_ENABLED=1 to run)")
	}

	rpcURL := "https://sepolia.base.org"

	fetcher, err := rpc.NewNonceFetcher(rpcURL)
	if err != nil {
		t.Fatalf("Failed to create nonce fetcher: %v", err)
	}
	defer fetcher.Close()

	// Test concurrent nonce fetches
	zeroAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")

	results := make(chan error, 5)

	for i := 0; i < 5; i++ {
		go func() {
			_, err := fetcher.GetNonce(zeroAddress)
			results <- err
		}()
	}

	// Collect results
	for i := 0; i < 5; i++ {
		err := <-results
		if err != nil {
			t.Errorf("Concurrent fetch %d failed: %v", i, err)
		}
	}

	t.Log("Concurrent nonce fetches successful")
}
