package contract

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/eip3009"
)

// generateValidSignature creates a cryptographically valid EIP-3009 signature
// This is used for testing to ensure signatures pass verification
func generateValidSignature(
	privateKey *ecdsa.PrivateKey,
	from common.Address,
	to common.Address,
	value *big.Int,
	validAfter *big.Int,
	validBefore *big.Int,
	nonce [32]byte,
	chainID *big.Int,
	usdcContract common.Address,
) (v uint8, r, s *big.Int, err error) {
	// Create EIP-712 domain
	domain := &eip3009.EIP712Domain{
		Name:              "USD Coin",
		Version:           "2",
		ChainID:           chainID,
		VerifyingContract: usdcContract,
	}

	// Create message
	message := &eip3009.ReceiveWithAuthorizationMessage{
		From:        from,
		To:          to,
		Value:       value,
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
		Nonce:       nonce,
	}

	// Compute typed data hash
	typedDataHash, err := eip3009.TypedDataHash(domain, message)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to compute typed data hash: %w", err)
	}

	// Sign the hash
	signature, err := crypto.Sign(typedDataHash.Bytes(), privateKey)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to sign message: %w", err)
	}

	// Extract v, r, s from signature
	v = signature[64] + 27 // Ethereum uses 27/28 for v
	r = new(big.Int).SetBytes(signature[0:32])
	s = new(big.Int).SetBytes(signature[32:64])

	return v, r, s, nil
}

// createTestPrivateKeyAndAddress generates a test private key and derives its address
func createTestPrivateKeyAndAddress() (*ecdsa.PrivateKey, common.Address, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("failed to generate private key: %w", err)
	}

	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	address := crypto.PubkeyToAddress(*publicKey)

	return privateKey, address, nil
}
