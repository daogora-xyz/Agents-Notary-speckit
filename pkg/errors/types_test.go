package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationError(t *testing.T) {
	t.Run("create validation error", func(t *testing.T) {
		err := NewValidationError("invalid_field", "field must not be empty")
		require.NotNil(t, err)
		assert.Equal(t, "validation error on 'invalid_field': field must not be empty", err.Error())
	})

	t.Run("validation error is error type", func(t *testing.T) {
		err := NewValidationError("test_field", "test message")
		var validationErr *ValidationError
		ok := errors.As(err, &validationErr)
		assert.True(t, ok, "should be ValidationError type")
		assert.Equal(t, "test_field", validationErr.Field)
		assert.Equal(t, "test message", validationErr.Message)
	})

	t.Run("wrap validation error", func(t *testing.T) {
		originalErr := errors.New("original error")
		validationErr := NewValidationError("test_field", "test message")
		wrappedErr := WrapValidationError(validationErr, originalErr)

		assert.Contains(t, wrappedErr.Error(), "test message")
		assert.True(t, errors.Is(wrappedErr, originalErr), "should preserve original error in chain")
	})
}

func TestNetworkError(t *testing.T) {
	t.Run("create network error", func(t *testing.T) {
		err := NewNetworkError("ethereum", "connection timeout")
		require.NotNil(t, err)
		assert.Equal(t, "network error on 'ethereum': connection timeout", err.Error())
	})

	t.Run("network error is error type", func(t *testing.T) {
		err := NewNetworkError("polygon", "test message")
		var networkErr *NetworkError
		ok := errors.As(err, &networkErr)
		assert.True(t, ok, "should be NetworkError type")
		assert.Equal(t, "polygon", networkErr.Network)
		assert.Equal(t, "test message", networkErr.Message)
	})

	t.Run("wrap network error", func(t *testing.T) {
		originalErr := errors.New("connection refused")
		networkErr := NewNetworkError("ethereum", "RPC call failed")
		wrappedErr := WrapNetworkError(networkErr, originalErr)

		assert.Contains(t, wrappedErr.Error(), "RPC call failed")
		assert.True(t, errors.Is(wrappedErr, originalErr), "should preserve original error in chain")
	})

	t.Run("network error with retryable flag", func(t *testing.T) {
		err := NewNetworkError("ethereum", "temporary failure")
		err.(*NetworkError).Retryable = true

		var networkErr *NetworkError
		errors.As(err, &networkErr)
		assert.True(t, networkErr.Retryable, "should mark error as retryable")
	})
}

func TestBlockchainError(t *testing.T) {
	t.Run("create blockchain error", func(t *testing.T) {
		err := NewBlockchainError("circular", "transaction failed", "txhash123")
		require.NotNil(t, err)
		assert.Equal(t, "blockchain error on 'circular': transaction failed (tx: txhash123)", err.Error())
	})

	t.Run("blockchain error without tx hash", func(t *testing.T) {
		err := NewBlockchainError("circular", "insufficient balance", "")
		assert.Equal(t, "blockchain error on 'circular': insufficient balance", err.Error())
	})

	t.Run("blockchain error is error type", func(t *testing.T) {
		err := NewBlockchainError("ethereum", "test message", "0xabc123")
		var blockchainErr *BlockchainError
		ok := errors.As(err, &blockchainErr)
		assert.True(t, ok, "should be BlockchainError type")
		assert.Equal(t, "ethereum", blockchainErr.Chain)
		assert.Equal(t, "test message", blockchainErr.Message)
		assert.Equal(t, "0xabc123", blockchainErr.TxHash)
	})

	t.Run("wrap blockchain error", func(t *testing.T) {
		originalErr := errors.New("RPC error")
		blockchainErr := NewBlockchainError("circular", "tx submission failed", "txhash")
		wrappedErr := WrapBlockchainError(blockchainErr, originalErr)

		assert.Contains(t, wrappedErr.Error(), "tx submission failed")
		assert.True(t, errors.Is(wrappedErr, originalErr), "should preserve original error in chain")
	})
}

func TestPaymentError(t *testing.T) {
	t.Run("create payment error", func(t *testing.T) {
		err := NewPaymentError("payment_nonce_123", "signature verification failed")
		require.NotNil(t, err)
		assert.Equal(t, "payment error for nonce 'payment_nonce_123': signature verification failed", err.Error())
	})

	t.Run("payment error is error type", func(t *testing.T) {
		err := NewPaymentError("nonce_abc", "test message")
		var paymentErr *PaymentError
		ok := errors.As(err, &paymentErr)
		assert.True(t, ok, "should be PaymentError type")
		assert.Equal(t, "nonce_abc", paymentErr.PaymentNonce)
		assert.Equal(t, "test message", paymentErr.Message)
	})

	t.Run("wrap payment error", func(t *testing.T) {
		originalErr := errors.New("EIP-3009 verification failed")
		paymentErr := NewPaymentError("nonce_123", "invalid signature")
		wrappedErr := WrapPaymentError(paymentErr, originalErr)

		assert.Contains(t, wrappedErr.Error(), "invalid signature")
		assert.True(t, errors.Is(wrappedErr, originalErr), "should preserve original error in chain")
	})
}

func TestErrorWrapping(t *testing.T) {
	t.Run("error chain preserves all errors", func(t *testing.T) {
		baseErr := errors.New("base error")
		networkErr := NewNetworkError("ethereum", "network failure")
		wrappedErr := WrapNetworkError(networkErr, baseErr)

		// Should be able to unwrap to base error
		assert.True(t, errors.Is(wrappedErr, baseErr))

		// Should be able to extract NetworkError from chain
		var ne *NetworkError
		ok := errors.As(wrappedErr, &ne)
		assert.True(t, ok)
		assert.Equal(t, "ethereum", ne.Network)
	})

	t.Run("multiple wrapping levels", func(t *testing.T) {
		err1 := errors.New("level 1")
		err2 := NewValidationError("field", "level 2")
		wrappedErr2 := WrapValidationError(err2, err1)

		err3 := NewNetworkError("ethereum", "level 3")
		wrappedErr3 := WrapNetworkError(err3, wrappedErr2)

		// Should be able to find errors at any level
		assert.True(t, errors.Is(wrappedErr3, err1))

		var ve *ValidationError
		assert.True(t, errors.As(wrappedErr3, &ve))

		var ne *NetworkError
		assert.True(t, errors.As(wrappedErr3, &ne))
	})
}

func TestErrorCategorization(t *testing.T) {
	t.Run("distinguish error types", func(t *testing.T) {
		validationErr := NewValidationError("field", "invalid")
		networkErr := NewNetworkError("ethereum", "timeout")
		blockchainErr := NewBlockchainError("circular", "failed", "tx123")
		paymentErr := NewPaymentError("nonce", "rejected")

		// Each should only match its own type
		var ve *ValidationError
		assert.True(t, errors.As(validationErr, &ve))
		assert.False(t, errors.As(networkErr, &ve))

		var ne *NetworkError
		assert.True(t, errors.As(networkErr, &ne))
		assert.False(t, errors.As(blockchainErr, &ne))

		var be *BlockchainError
		assert.True(t, errors.As(blockchainErr, &be))
		assert.False(t, errors.As(paymentErr, &be))

		var pe *PaymentError
		assert.True(t, errors.As(paymentErr, &pe))
		assert.False(t, errors.As(validationErr, &pe))
	})
}
