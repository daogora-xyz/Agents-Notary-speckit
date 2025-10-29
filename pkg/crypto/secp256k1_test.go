package crypto

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecp256k1SignAndVerify(t *testing.T) {
	t.Run("sign and verify valid data", func(t *testing.T) {
		// Generate a new private key
		privKey, err := GeneratePrivateKey()
		require.NoError(t, err, "should generate private key")
		require.NotNil(t, privKey, "private key should not be nil")

		// Get public key
		pubKey := privKey.PubKey()
		require.NotNil(t, pubKey, "public key should not be nil")

		// Data to sign
		data := []byte("Hello, Circular Protocol!")
		dataHash := sha256.Sum256(data)

		// Sign the data
		signature, err := Sign(privKey, dataHash[:])
		require.NoError(t, err, "should sign data")
		require.NotNil(t, signature, "signature should not be nil")

		// Verify the signature
		valid := Verify(pubKey, dataHash[:], signature)
		assert.True(t, valid, "signature should be valid")
	})

	t.Run("verify fails with wrong public key", func(t *testing.T) {
		// Generate two key pairs
		privKey1, err := GeneratePrivateKey()
		require.NoError(t, err)

		privKey2, err := GeneratePrivateKey()
		require.NoError(t, err)

		pubKey2 := privKey2.PubKey()

		// Sign with privKey1
		data := []byte("test data")
		dataHash := sha256.Sum256(data)
		signature, err := Sign(privKey1, dataHash[:])
		require.NoError(t, err)

		// Verify with pubKey2 (should fail)
		valid := Verify(pubKey2, dataHash[:], signature)
		assert.False(t, valid, "signature should NOT be valid with wrong public key")
	})

	t.Run("verify fails with wrong data", func(t *testing.T) {
		privKey, err := GeneratePrivateKey()
		require.NoError(t, err)

		pubKey := privKey.PubKey()

		// Sign original data
		originalData := []byte("original data")
		originalHash := sha256.Sum256(originalData)
		signature, err := Sign(privKey, originalHash[:])
		require.NoError(t, err)

		// Verify with different data (should fail)
		differentData := []byte("different data")
		differentHash := sha256.Sum256(differentData)
		valid := Verify(pubKey, differentHash[:], signature)
		assert.False(t, valid, "signature should NOT be valid with different data")
	})

	t.Run("verify fails with tampered signature", func(t *testing.T) {
		privKey, err := GeneratePrivateKey()
		require.NoError(t, err)

		pubKey := privKey.PubKey()

		data := []byte("test data")
		dataHash := sha256.Sum256(data)
		signature, err := Sign(privKey, dataHash[:])
		require.NoError(t, err)

		// Tamper with signature
		signatureBytes := signature.Serialize()
		signatureBytes[0] ^= 0x01 // Flip one bit

		// Try to parse tampered signature
		tamperedSig, err := ParseSignature(signatureBytes)
		if err != nil {
			// If parsing fails, that's expected
			t.Log("Tampered signature failed to parse (expected)")
			return
		}

		// If parsing succeeded, verification should still fail
		valid := Verify(pubKey, dataHash[:], tamperedSig)
		assert.False(t, valid, "tampered signature should NOT be valid")
	})

	t.Run("signature serialization round-trip", func(t *testing.T) {
		privKey, err := GeneratePrivateKey()
		require.NoError(t, err)

		pubKey := privKey.PubKey()

		data := []byte("test data")
		dataHash := sha256.Sum256(data)
		signature, err := Sign(privKey, dataHash[:])
		require.NoError(t, err)

		// Serialize signature
		sigBytes := signature.Serialize()
		require.NotEmpty(t, sigBytes, "serialized signature should not be empty")

		// Deserialize signature
		deserializedSig, err := ParseSignature(sigBytes)
		require.NoError(t, err, "should deserialize signature")

		// Verify with deserialized signature
		valid := Verify(pubKey, dataHash[:], deserializedSig)
		assert.True(t, valid, "deserialized signature should be valid")
	})

	t.Run("public key serialization round-trip", func(t *testing.T) {
		privKey, err := GeneratePrivateKey()
		require.NoError(t, err)

		pubKey := privKey.PubKey()

		// Serialize public key
		pubKeyBytes := SerializePublicKey(pubKey)
		require.NotEmpty(t, pubKeyBytes, "serialized public key should not be empty")

		// Deserialize public key
		deserializedPubKey, err := ParsePublicKey(pubKeyBytes)
		require.NoError(t, err, "should deserialize public key")

		// Sign with original private key
		data := []byte("test data")
		dataHash := sha256.Sum256(data)
		signature, err := Sign(privKey, dataHash[:])
		require.NoError(t, err)

		// Verify with deserialized public key
		valid := Verify(deserializedPubKey, dataHash[:], signature)
		assert.True(t, valid, "signature should be valid with deserialized public key")
	})
}

func TestSecp256k1Performance(t *testing.T) {
	t.Run("signing performance <100ms per spec SC-006", func(t *testing.T) {
		privKey, err := GeneratePrivateKey()
		require.NoError(t, err)

		data := []byte("benchmark data")
		dataHash := sha256.Sum256(data)

		// Benchmark signing time
		iterations := 100
		totalDuration := int64(0)

		for i := 0; i < iterations; i++ {
			start := testing.Benchmark(func(b *testing.B) {
				_, _ = Sign(privKey, dataHash[:])
			})
			totalDuration += start.T.Nanoseconds()
		}

		avgDurationMs := float64(totalDuration) / float64(iterations) / 1000000.0
		t.Logf("Average signing time: %.2f ms", avgDurationMs)

		// Per spec SC-006: Crypto signing must complete in <100ms
		assert.Less(t, avgDurationMs, 100.0, "signing should take less than 100ms on average")
	})
}

func TestSecp256k1CircularProtocolCompatibility(t *testing.T) {
	t.Run("secp256k1 curve compatibility", func(t *testing.T) {
		// This test documents that we're using the same curve as Bitcoin/Ethereum/Circular
		privKey, err := GeneratePrivateKey()
		require.NoError(t, err)

		// The curve should be secp256k1
		// This is implicit in btcsuite/btcd/btcec/v2 but we document it here
		pubKey := privKey.PubKey()
		require.NotNil(t, pubKey)

		// Public key should be 33 bytes (compressed) or 65 bytes (uncompressed)
		pubKeyBytes := SerializePublicKey(pubKey)
		validLength := len(pubKeyBytes) == 33 || len(pubKeyBytes) == 65
		assert.True(t, validLength, "public key should be 33 (compressed) or 65 (uncompressed) bytes")

		t.Logf("Public key length: %d bytes", len(pubKeyBytes))
	})

	t.Run("signature determinism", func(t *testing.T) {
		// Note: btcsuite/btcd uses RFC 6979 for deterministic signatures
		// This means signing the same data with the same key produces the same signature
		privKey, err := GeneratePrivateKey()
		require.NoError(t, err)

		data := []byte("deterministic test")
		dataHash := sha256.Sum256(data)

		sig1, err := Sign(privKey, dataHash[:])
		require.NoError(t, err)

		sig2, err := Sign(privKey, dataHash[:])
		require.NoError(t, err)

		// Signatures should be identical (RFC 6979 deterministic signing)
		sig1Bytes := sig1.Serialize()
		sig2Bytes := sig2.Serialize()
		assert.Equal(t, sig1Bytes, sig2Bytes, "signatures should be deterministic")
	})
}
