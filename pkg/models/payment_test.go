package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaymentValidation(t *testing.T) {
	t.Run("valid payment", func(t *testing.T) {
		payment := &Payment{
			RequestID:    "req_test_12345",
			PaymentNonce: "nonce_abc123",
			FromAddress:  "0x1234567890abcdef1234567890abcdef12345678",
			ToAddress:    "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			AmountUSDC:   "10.50",
			Network:      NetworkEthereum,
			EVMTxHash:    "0xabcdef1234567890",
			Status:       PaymentStatusAuthorized,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := payment.Validate()
		assert.NoError(t, err, "valid payment should pass validation")
	})

	t.Run("missing request_id", func(t *testing.T) {
		payment := &Payment{
			RequestID:    "", // Missing
			PaymentNonce: "nonce_abc123",
			FromAddress:  "0x1234567890abcdef1234567890abcdef12345678",
			ToAddress:    "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			AmountUSDC:   "10.50",
			Network:      NetworkEthereum,
			Status:       PaymentStatusAuthorized,
		}

		err := payment.Validate()
		require.Error(t, err, "payment without request_id should fail validation")
		assert.Contains(t, err.Error(), "request_id")
	})

	t.Run("missing payment_nonce", func(t *testing.T) {
		payment := &Payment{
			RequestID:    "req_test_12345",
			PaymentNonce: "", // Missing
			FromAddress:  "0x1234567890abcdef1234567890abcdef12345678",
			ToAddress:    "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			AmountUSDC:   "10.50",
			Network:      NetworkEthereum,
			Status:       PaymentStatusAuthorized,
		}

		err := payment.Validate()
		require.Error(t, err, "payment without payment_nonce should fail validation")
		assert.Contains(t, err.Error(), "payment_nonce")
	})

	t.Run("missing from_address", func(t *testing.T) {
		payment := &Payment{
			RequestID:    "req_test_12345",
			PaymentNonce: "nonce_abc123",
			FromAddress:  "", // Missing
			ToAddress:    "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			AmountUSDC:   "10.50",
			Network:      NetworkEthereum,
			Status:       PaymentStatusAuthorized,
		}

		err := payment.Validate()
		require.Error(t, err, "payment without from_address should fail validation")
		assert.Contains(t, err.Error(), "from_address")
	})

	t.Run("missing to_address", func(t *testing.T) {
		payment := &Payment{
			RequestID:    "req_test_12345",
			PaymentNonce: "nonce_abc123",
			FromAddress:  "0x1234567890abcdef1234567890abcdef12345678",
			ToAddress:    "", // Missing
			AmountUSDC:   "10.50",
			Network:      NetworkEthereum,
			Status:       PaymentStatusAuthorized,
		}

		err := payment.Validate()
		require.Error(t, err, "payment without to_address should fail validation")
		assert.Contains(t, err.Error(), "to_address")
	})

	t.Run("invalid amount (empty)", func(t *testing.T) {
		payment := &Payment{
			RequestID:    "req_test_12345",
			PaymentNonce: "nonce_abc123",
			FromAddress:  "0x1234567890abcdef1234567890abcdef12345678",
			ToAddress:    "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			AmountUSDC:   "", // Invalid
			Network:      NetworkEthereum,
			Status:       PaymentStatusAuthorized,
		}

		err := payment.Validate()
		require.Error(t, err, "payment with empty amount should fail validation")
		assert.Contains(t, err.Error(), "amount_usdc")
	})

	t.Run("invalid amount (zero)", func(t *testing.T) {
		payment := &Payment{
			RequestID:    "req_test_12345",
			PaymentNonce: "nonce_abc123",
			FromAddress:  "0x1234567890abcdef1234567890abcdef12345678",
			ToAddress:    "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			AmountUSDC:   "0", // Invalid
			Network:      NetworkEthereum,
			Status:       PaymentStatusAuthorized,
		}

		err := payment.Validate()
		require.Error(t, err, "payment with zero amount should fail validation")
		assert.Contains(t, err.Error(), "amount_usdc")
	})

	t.Run("invalid amount (negative)", func(t *testing.T) {
		payment := &Payment{
			RequestID:    "req_test_12345",
			PaymentNonce: "nonce_abc123",
			FromAddress:  "0x1234567890abcdef1234567890abcdef12345678",
			ToAddress:    "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			AmountUSDC:   "-10.50", // Invalid
			Network:      NetworkEthereum,
			Status:       PaymentStatusAuthorized,
		}

		err := payment.Validate()
		require.Error(t, err, "payment with negative amount should fail validation")
		assert.Contains(t, err.Error(), "amount_usdc")
	})

	t.Run("invalid network", func(t *testing.T) {
		payment := &Payment{
			RequestID:    "req_test_12345",
			PaymentNonce: "nonce_abc123",
			FromAddress:  "0x1234567890abcdef1234567890abcdef12345678",
			ToAddress:    "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			AmountUSDC:   "10.50",
			Network:      "invalid_network", // Invalid
			Status:       PaymentStatusAuthorized,
		}

		err := payment.Validate()
		require.Error(t, err, "payment with invalid network should fail validation")
		assert.Contains(t, err.Error(), "network")
	})

	t.Run("all valid networks", func(t *testing.T) {
		validNetworks := []Network{
			NetworkEthereum,
			NetworkPolygon,
			NetworkBase,
			NetworkArbitrum,
			NetworkOptimism,
		}

		for _, network := range validNetworks {
			payment := &Payment{
				RequestID:    "req_test_12345",
				PaymentNonce: "nonce_abc123",
				FromAddress:  "0x1234567890abcdef1234567890abcdef12345678",
				ToAddress:    "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				AmountUSDC:   "10.50",
				Network:      network,
				Status:       PaymentStatusAuthorized,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			err := payment.Validate()
			assert.NoError(t, err, "network %s should be valid", network)
		}
	})

	t.Run("all valid statuses", func(t *testing.T) {
		validStatuses := []PaymentStatus{
			PaymentStatusPending,
			PaymentStatusAuthorized,
			PaymentStatusSettled,
			PaymentStatusFailed,
		}

		for _, status := range validStatuses {
			payment := &Payment{
				RequestID:    "req_test_12345",
				PaymentNonce: "nonce_abc123",
				FromAddress:  "0x1234567890abcdef1234567890abcdef12345678",
				ToAddress:    "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				AmountUSDC:   "10.50",
				Network:      NetworkEthereum,
				Status:       status,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			err := payment.Validate()
			assert.NoError(t, err, "status %s should be valid", status)
		}
	})

	t.Run("evm_tx_hash optional when pending", func(t *testing.T) {
		payment := &Payment{
			RequestID:    "req_test_12345",
			PaymentNonce: "nonce_abc123",
			FromAddress:  "0x1234567890abcdef1234567890abcdef12345678",
			ToAddress:    "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			AmountUSDC:   "10.50",
			Network:      NetworkEthereum,
			EVMTxHash:    "", // Optional when pending
			Status:       PaymentStatusPending,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := payment.Validate()
		assert.NoError(t, err, "evm_tx_hash should be optional when payment is pending")
	})
}
