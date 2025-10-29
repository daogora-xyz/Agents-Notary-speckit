package crypto

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
)

// GeneratePrivateKey generates a new secp256k1 private key
func GeneratePrivateKey() (*btcec.PrivateKey, error) {
	privKey, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}
	return privKey, nil
}

// Sign signs the given message hash with the private key
// The message should be a 32-byte hash (e.g., SHA-256)
// Returns a DER-encoded signature
func Sign(privKey *btcec.PrivateKey, messageHash []byte) (*ecdsa.Signature, error) {
	if len(messageHash) != 32 {
		return nil, fmt.Errorf("message hash must be 32 bytes (got %d)", len(messageHash))
	}

	// Sign using RFC 6979 deterministic signing
	signature := ecdsa.Sign(privKey, messageHash)
	return signature, nil
}

// Verify verifies a signature against a public key and message hash
// Returns true if the signature is valid, false otherwise
func Verify(pubKey *btcec.PublicKey, messageHash []byte, signature *ecdsa.Signature) bool {
	if len(messageHash) != 32 {
		return false
	}

	return signature.Verify(messageHash, pubKey)
}

// ParseSignature parses a DER-encoded signature
func ParseSignature(sigBytes []byte) (*ecdsa.Signature, error) {
	signature, err := ecdsa.ParseDERSignature(sigBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse signature: %w", err)
	}
	return signature, nil
}

// SerializePublicKey serializes a public key to compressed format (33 bytes)
func SerializePublicKey(pubKey *btcec.PublicKey) []byte {
	return pubKey.SerializeCompressed()
}

// ParsePublicKey parses a serialized public key (compressed or uncompressed)
func ParsePublicKey(pubKeyBytes []byte) (*btcec.PublicKey, error) {
	pubKey, err := btcec.ParsePubKey(pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	return pubKey, nil
}
