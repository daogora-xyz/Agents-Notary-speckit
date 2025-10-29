package eip3009

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/config"
)

// SignatureVerifier handles EIP-3009 signature verification
type SignatureVerifier struct {
	config *config.Config
}

// NewSignatureVerifier creates a new signature verifier
func NewSignatureVerifier(cfg *config.Config) *SignatureVerifier {
	return &SignatureVerifier{
		config: cfg,
	}
}

// VerifyAuthorization performs complete signature verification including:
// - Input validation
// - EIP-712 domain matching
// - Signature recovery via secp256k1 ECDSA
// - Time bound validation
// - Signer address verification
func (v *SignatureVerifier) VerifyAuthorization(
	auth *EIP3009Authorization,
	network string,
) (*VerifyPaymentOutput, error) {
	// Step 1: Input validation
	if err := auth.Validate(); err != nil {
		return &VerifyPaymentOutput{
			IsValid: false,
			Error:   fmt.Sprintf("validation failed: %v", err),
		}, nil
	}

	// Step 2: Get network configuration
	networkCfg, exists := v.config.Networks[network]
	if !exists {
		return &VerifyPaymentOutput{
			IsValid: false,
			Error:   fmt.Sprintf("unsupported network: %s", network),
		}, nil
	}

	// Step 3: Time bound validation
	currentTime := time.Now().Unix()
	if currentTime < int64(auth.ValidAfter) {
		return &VerifyPaymentOutput{
			IsValid: false,
			Error:   fmt.Sprintf("authorization not yet valid: current=%d, validAfter=%d", currentTime, auth.ValidAfter),
		}, nil
	}
	if currentTime >= int64(auth.ValidBefore) {
		return &VerifyPaymentOutput{
			IsValid: false,
			Error:   fmt.Sprintf("authorization expired: current=%d, validBefore=%d", currentTime, auth.ValidBefore),
		}, nil
	}

	// Step 4: Create EIP-712 domain separator
	domain := &EIP712Domain{
		Name:              v.config.EIP712.DomainName,
		Version:           v.config.EIP712.DomainVersion,
		ChainID:           big.NewInt(int64(networkCfg.ChainID)),
		VerifyingContract: common.HexToAddress(networkCfg.USDCContract),
	}

	// Step 5: Convert authorization to message
	message, err := auth.ToMessage()
	if err != nil {
		return &VerifyPaymentOutput{
			IsValid: false,
			Error:   fmt.Sprintf("failed to convert authorization: %v", err),
		}, nil
	}

	// Step 6: Compute EIP-712 typed data hash
	typedDataHash, err := TypedDataHash(domain, message)
	if err != nil {
		return &VerifyPaymentOutput{
			IsValid: false,
			Error:   fmt.Sprintf("failed to compute typed data hash: %v", err),
		}, nil
	}

	// Step 7: Get signature bytes
	signature, err := auth.GetSignature()
	if err != nil {
		return &VerifyPaymentOutput{
			IsValid: false,
			Error:   fmt.Sprintf("failed to parse signature: %v", err),
		}, nil
	}

	// Step 8: Recover public key from signature
	recoveredPubKey, err := crypto.SigToPub(typedDataHash.Bytes(), signature)
	if err != nil {
		return &VerifyPaymentOutput{
			IsValid: false,
			Error:   fmt.Sprintf("failed to recover public key: %v", err),
		}, nil
	}

	// Step 9: Derive signer address from recovered public key
	signerAddress := crypto.PubkeyToAddress(*recoveredPubKey)

	// Step 10: Verify signer matches 'from' address
	expectedFrom := common.HexToAddress(auth.From)
	if signerAddress != expectedFrom {
		return &VerifyPaymentOutput{
			IsValid:       false,
			SignerAddress: signerAddress.Hex(),
			Error:         fmt.Sprintf("signer mismatch: expected %s, got %s", expectedFrom.Hex(), signerAddress.Hex()),
		}, nil
	}

	// All checks passed
	return &VerifyPaymentOutput{
		IsValid:       true,
		SignerAddress: signerAddress.Hex(),
	}, nil
}

// VerifyDomain checks if the domain separator matches the network configuration
// This is a helper function for domain matching validation
func (v *SignatureVerifier) VerifyDomain(network string) (*EIP712Domain, error) {
	networkCfg, exists := v.config.Networks[network]
	if !exists {
		return nil, fmt.Errorf("unsupported network: %s", network)
	}

	domain := &EIP712Domain{
		Name:              v.config.EIP712.DomainName,
		Version:           v.config.EIP712.DomainVersion,
		ChainID:           big.NewInt(int64(networkCfg.ChainID)),
		VerifyingContract: common.HexToAddress(networkCfg.USDCContract),
	}

	return domain, nil
}

// RecoverSigner is a helper function to recover the signer address from a signature
// without performing full verification. Useful for debugging.
func (v *SignatureVerifier) RecoverSigner(
	auth *EIP3009Authorization,
	network string,
) (common.Address, error) {
	// Get network configuration
	networkCfg, exists := v.config.Networks[network]
	if !exists {
		return common.Address{}, fmt.Errorf("unsupported network: %s", network)
	}

	// Create domain
	domain := &EIP712Domain{
		Name:              v.config.EIP712.DomainName,
		Version:           v.config.EIP712.DomainVersion,
		ChainID:           big.NewInt(int64(networkCfg.ChainID)),
		VerifyingContract: common.HexToAddress(networkCfg.USDCContract),
	}

	// Convert to message
	message, err := auth.ToMessage()
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to convert authorization: %w", err)
	}

	// Compute hash
	typedDataHash, err := TypedDataHash(domain, message)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to compute hash: %w", err)
	}

	// Get signature
	signature, err := auth.GetSignature()
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to parse signature: %w", err)
	}

	// Recover public key
	recoveredPubKey, err := crypto.SigToPub(typedDataHash.Bytes(), signature)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to recover public key: %w", err)
	}

	return crypto.PubkeyToAddress(*recoveredPubKey), nil
}

// ValidateTimeBounds checks if the authorization is within valid time bounds
func ValidateTimeBounds(validAfter, validBefore uint64) error {
	currentTime := uint64(time.Now().Unix())

	if currentTime < validAfter {
		return fmt.Errorf("authorization not yet valid: current=%d, validAfter=%d", currentTime, validAfter)
	}

	if currentTime >= validBefore {
		return fmt.Errorf("authorization expired: current=%d, validBefore=%d", currentTime, validBefore)
	}

	return nil
}
