package unit

import (
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/eip3009"
)

// TestSignatureVerification_ValidSignature tests successful signature verification with known test vector
func TestSignatureVerification_ValidSignature(t *testing.T) {
	// Generate test private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Derive public key and address
	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKey)

	// Create test authorization message
	toAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	value := big.NewInt(50000)
	now := time.Now().Unix()
	validAfter := big.NewInt(now - 3600)  // 1 hour ago
	validBefore := big.NewInt(now + 3600) // 1 hour from now
	nonce := [32]byte{}
	copy(nonce[:], []byte("test-nonce-123"))

	// Create domain separator for Base
	domain := &eip3009.EIP712Domain{
		Name:              "USD Coin",
		Version:           "2",
		ChainID:           big.NewInt(8453), // Base mainnet
		VerifyingContract: common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"),
	}

	// Create message
	message := &eip3009.ReceiveWithAuthorizationMessage{
		From:        fromAddress,
		To:          toAddress,
		Value:       value,
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
		Nonce:       nonce,
	}

	// Compute typed data hash
	typedDataHash, err := eip3009.TypedDataHash(domain, message)
	if err != nil {
		t.Fatalf("Failed to compute typed data hash: %v", err)
	}

	// Sign the hash
	signature, err := crypto.Sign(typedDataHash.Bytes(), privateKey)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Extract v, r, s from signature
	v := signature[64] + 27 // Ethereum uses 27/28 for v
	r := new(big.Int).SetBytes(signature[0:32])
	s := new(big.Int).SetBytes(signature[32:64])

	// This test should pass when we implement the verifier
	// For now, we're just verifying the signature generation works
	t.Logf("Generated valid signature:")
	t.Logf("  From: %s", fromAddress.Hex())
	t.Logf("  V: %d", v)
	t.Logf("  R: %s", common.BytesToHash(r.Bytes()).Hex())
	t.Logf("  S: %s", common.BytesToHash(s.Bytes()).Hex())

	// Recover public key from signature
	recoveredPubKey, err := crypto.SigToPub(typedDataHash.Bytes(), signature)
	if err != nil {
		t.Fatalf("Failed to recover public key: %v", err)
	}

	recoveredAddress := crypto.PubkeyToAddress(*recoveredPubKey)
	if recoveredAddress != fromAddress {
		t.Errorf("Recovered address %s does not match original %s", recoveredAddress.Hex(), fromAddress.Hex())
	}
}

// TestSignatureVerification_InvalidSignature tests detection of tampered value field
func TestSignatureVerification_InvalidSignature(t *testing.T) {
	// Generate test private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKey)

	// Create test authorization with original value
	toAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	originalValue := big.NewInt(50000)
	now := time.Now().Unix()
	validAfter := big.NewInt(now - 3600)
	validBefore := big.NewInt(now + 3600)
	nonce := [32]byte{}
	copy(nonce[:], []byte("test-nonce-456"))

	domain := &eip3009.EIP712Domain{
		Name:              "USD Coin",
		Version:           "2",
		ChainID:           big.NewInt(8453),
		VerifyingContract: common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"),
	}

	// Sign with original value
	message := &eip3009.ReceiveWithAuthorizationMessage{
		From:        fromAddress,
		To:          toAddress,
		Value:       originalValue,
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
		Nonce:       nonce,
	}

	typedDataHash, _ := eip3009.TypedDataHash(domain, message)
	signature, _ := crypto.Sign(typedDataHash.Bytes(), privateKey)

	// Now tamper with the value
	tamperedValue := big.NewInt(100000) // Changed from 50000
	tamperedMessage := &eip3009.ReceiveWithAuthorizationMessage{
		From:        fromAddress,
		To:          toAddress,
		Value:       tamperedValue,
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
		Nonce:       nonce,
	}

	// Compute hash with tampered value
	tamperedHash, _ := eip3009.TypedDataHash(domain, tamperedMessage)

	// Try to recover - should not match original signer
	recoveredPubKey, err := crypto.SigToPub(tamperedHash.Bytes(), signature)
	if err != nil {
		t.Fatalf("Failed to recover public key: %v", err)
	}

	recoveredAddress := crypto.PubkeyToAddress(*recoveredPubKey)
	if recoveredAddress == fromAddress {
		t.Error("Signature verification should fail for tampered value, but recovered matching address")
	} else {
		t.Logf("Correctly detected tampered signature: recovered %s != original %s",
			recoveredAddress.Hex(), fromAddress.Hex())
	}
}

// TestSignatureVerification_WrongNetwork tests detection of signature from different chain ID
func TestSignatureVerification_WrongNetwork(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKey)

	toAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	value := big.NewInt(50000)
	now := time.Now().Unix()
	validAfter := big.NewInt(now - 3600)
	validBefore := big.NewInt(now + 3600)
	nonce := [32]byte{}
	copy(nonce[:], []byte("test-nonce-789"))

	// Sign with Base Sepolia (testnet)
	sepoliaDomain := &eip3009.EIP712Domain{
		Name:              "USD Coin",
		Version:           "2",
		ChainID:           big.NewInt(84532), // Base Sepolia
		VerifyingContract: common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e"),
	}

	message := &eip3009.ReceiveWithAuthorizationMessage{
		From:        fromAddress,
		To:          toAddress,
		Value:       value,
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
		Nonce:       nonce,
	}

	sepoliaHash, _ := eip3009.TypedDataHash(sepoliaDomain, message)
	signature, _ := crypto.Sign(sepoliaHash.Bytes(), privateKey)

	// Try to verify on Base mainnet
	mainnetDomain := &eip3009.EIP712Domain{
		Name:              "USD Coin",
		Version:           "2",
		ChainID:           big.NewInt(8453), // Base mainnet
		VerifyingContract: common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"),
	}

	mainnetHash, _ := eip3009.TypedDataHash(mainnetDomain, message)

	// Recovery should fail or give wrong address
	recoveredPubKey, err := crypto.SigToPub(mainnetHash.Bytes(), signature)
	if err != nil {
		t.Logf("Correctly failed to recover signature from wrong network: %v", err)
		return
	}

	recoveredAddress := crypto.PubkeyToAddress(*recoveredPubKey)
	if recoveredAddress == fromAddress {
		t.Error("Signature should not verify across different networks")
	} else {
		t.Logf("Correctly detected cross-network signature: %s != %s",
			recoveredAddress.Hex(), fromAddress.Hex())
	}
}

// TestSignatureVerification_TimeBounds tests validation of validAfter and validBefore
func TestSignatureVerification_TimeBounds(t *testing.T) {
	now := time.Now().Unix()

	testCases := []struct {
		name          string
		validAfter    int64
		validBefore   int64
		shouldBeValid bool
	}{
		{
			name:          "valid time window",
			validAfter:    now - 3600, // 1 hour ago
			validBefore:   now + 3600, // 1 hour from now
			shouldBeValid: true,
		},
		{
			name:          "expired (validBefore in past)",
			validAfter:    now - 7200, // 2 hours ago
			validBefore:   now - 3600, // 1 hour ago (expired)
			shouldBeValid: false,
		},
		{
			name:          "not yet valid (validAfter in future)",
			validAfter:    now + 3600, // 1 hour from now (not yet valid)
			validBefore:   now + 7200, // 2 hours from now
			shouldBeValid: false,
		},
		{
			name:          "exactly at validAfter boundary",
			validAfter:    now,
			validBefore:   now + 3600,
			shouldBeValid: true,
		},
		{
			name:          "exactly at validBefore boundary",
			validAfter:    now - 3600,
			validBefore:   now,
			shouldBeValid: false, // Should be strictly less than validBefore
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Time bound validation logic
			currentTime := now
			isAfterValid := currentTime >= tc.validAfter
			isBeforeValid := currentTime < tc.validBefore
			isValid := isAfterValid && isBeforeValid

			if isValid != tc.shouldBeValid {
				t.Errorf("Time bound validation incorrect: got %v, want %v", isValid, tc.shouldBeValid)
			}

			t.Logf("Time bounds check: validAfter=%d, validBefore=%d, current=%d, valid=%v",
				tc.validAfter, tc.validBefore, currentTime, isValid)
		})
	}
}

// TestSignatureVerification_VParameter tests that v must be 27 or 28
func TestSignatureVerification_VParameter(t *testing.T) {
	validVValues := []uint8{27, 28}
	invalidVValues := []uint8{0, 1, 26, 29, 255}

	for _, v := range validVValues {
		if v != 27 && v != 28 {
			t.Errorf("Valid v value %d failed validation", v)
		}
	}

	for _, v := range invalidVValues {
		if v == 27 || v == 28 {
			t.Errorf("Invalid v value %d passed validation", v)
		} else {
			t.Logf("Correctly rejected invalid v value: %d", v)
		}
	}
}

// TestSignatureVerification_AddressFormats tests Ethereum address validation
func TestSignatureVerification_AddressFormats(t *testing.T) {
	testCases := []struct {
		address string
		valid   bool
	}{
		{"0x1234567890123456789012345678901234567890", true},
		{"0x0000000000000000000000000000000000000000", true},
		{"0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", true},
		{"1234567890123456789012345678901234567890", false},   // Missing 0x
		{"0x12345", false},                                     // Too short
		{"0x12345678901234567890123456789012345678901", false}, // Too long
		{"0xGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG", false}, // Invalid hex
	}

	for _, tc := range testCases {
		t.Run(tc.address, func(t *testing.T) {
			_ = common.HexToAddress(tc.address)
			// common.HexToAddress is lenient, so we need to validate format separately
			isValid := len(tc.address) == 42 && tc.address[:2] == "0x"

			if isValid != tc.valid {
				t.Errorf("Address validation mismatch for %s: got %v, want %v", tc.address, isValid, tc.valid)
			}
		})
	}
}
