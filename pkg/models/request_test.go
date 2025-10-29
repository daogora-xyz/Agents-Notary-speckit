package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCertificationRequestValidation(t *testing.T) {
	t.Run("valid certification request", func(t *testing.T) {
		req := &CertificationRequest{
			RequestID:     "req_test_12345",
			ClientID:      "client_abc",
			DataHash:      "0x1234567890abcdef",
			DataSizeBytes: 1024,
			Status:        StatusPending,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		err := req.Validate()
		assert.NoError(t, err, "valid request should pass validation")
	})

	t.Run("missing request_id", func(t *testing.T) {
		req := &CertificationRequest{
			RequestID:     "", // Missing
			ClientID:      "client_abc",
			DataHash:      "0x1234567890abcdef",
			DataSizeBytes: 1024,
			Status:        StatusPending,
		}

		err := req.Validate()
		require.Error(t, err, "request without request_id should fail validation")
		assert.Contains(t, err.Error(), "request_id")
	})

	t.Run("missing client_id", func(t *testing.T) {
		req := &CertificationRequest{
			RequestID:     "req_test_12345",
			ClientID:      "", // Missing
			DataHash:      "0x1234567890abcdef",
			DataSizeBytes: 1024,
			Status:        StatusPending,
		}

		err := req.Validate()
		require.Error(t, err, "request without client_id should fail validation")
		assert.Contains(t, err.Error(), "client_id")
	})

	t.Run("missing data_hash", func(t *testing.T) {
		req := &CertificationRequest{
			RequestID:     "req_test_12345",
			ClientID:      "client_abc",
			DataHash:      "", // Missing
			DataSizeBytes: 1024,
			Status:        StatusPending,
		}

		err := req.Validate()
		require.Error(t, err, "request without data_hash should fail validation")
		assert.Contains(t, err.Error(), "data_hash")
	})

	t.Run("invalid data size (zero)", func(t *testing.T) {
		req := &CertificationRequest{
			RequestID:     "req_test_12345",
			ClientID:      "client_abc",
			DataHash:      "0x1234567890abcdef",
			DataSizeBytes: 0, // Invalid
			Status:        StatusPending,
		}

		err := req.Validate()
		require.Error(t, err, "request with zero data size should fail validation")
		assert.Contains(t, err.Error(), "data_size_bytes")
	})

	t.Run("invalid data size (negative)", func(t *testing.T) {
		req := &CertificationRequest{
			RequestID:     "req_test_12345",
			ClientID:      "client_abc",
			DataHash:      "0x1234567890abcdef",
			DataSizeBytes: -100, // Invalid
			Status:        StatusPending,
		}

		err := req.Validate()
		require.Error(t, err, "request with negative data size should fail validation")
		assert.Contains(t, err.Error(), "data_size_bytes")
	})

	t.Run("invalid status", func(t *testing.T) {
		req := &CertificationRequest{
			RequestID:     "req_test_12345",
			ClientID:      "client_abc",
			DataHash:      "0x1234567890abcdef",
			DataSizeBytes: 1024,
			Status:        "invalid_status", // Invalid
		}

		err := req.Validate()
		require.Error(t, err, "request with invalid status should fail validation")
		assert.Contains(t, err.Error(), "status")
	})

	t.Run("all valid statuses", func(t *testing.T) {
		validStatuses := []RequestStatus{
			StatusPending,
			StatusPaymentReceived,
			StatusCertifying,
			StatusCompleted,
			StatusFailed,
		}

		for _, status := range validStatuses {
			req := &CertificationRequest{
				RequestID:     "req_test_12345",
				ClientID:      "client_abc",
				DataHash:      "0x1234567890abcdef",
				DataSizeBytes: 1024,
				Status:        status,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}

			err := req.Validate()
			assert.NoError(t, err, "status %s should be valid", status)
		}
	})
}

func TestCertificationRequestStatusTransitions(t *testing.T) {
	t.Run("can transition to payment_received", func(t *testing.T) {
		req := &CertificationRequest{
			RequestID:     "req_test_12345",
			ClientID:      "client_abc",
			DataHash:      "0x1234567890abcdef",
			DataSizeBytes: 1024,
			Status:        StatusPending,
		}

		req.Status = StatusPaymentReceived
		err := req.Validate()
		assert.NoError(t, err)
	})

	t.Run("can transition to certifying", func(t *testing.T) {
		req := &CertificationRequest{
			RequestID:     "req_test_12345",
			ClientID:      "client_abc",
			DataHash:      "0x1234567890abcdef",
			DataSizeBytes: 1024,
			Status:        StatusPaymentReceived,
		}

		req.Status = StatusCertifying
		err := req.Validate()
		assert.NoError(t, err)
	})

	t.Run("can transition to completed", func(t *testing.T) {
		req := &CertificationRequest{
			RequestID:     "req_test_12345",
			ClientID:      "client_abc",
			DataHash:      "0x1234567890abcdef",
			DataSizeBytes: 1024,
			Status:        StatusCertifying,
		}

		req.Status = StatusCompleted
		err := req.Validate()
		assert.NoError(t, err)
	})

	t.Run("can transition to failed", func(t *testing.T) {
		req := &CertificationRequest{
			RequestID:     "req_test_12345",
			ClientID:      "client_abc",
			DataHash:      "0x1234567890abcdef",
			DataSizeBytes: 1024,
			Status:        StatusCertifying,
		}

		req.Status = StatusFailed
		err := req.Validate()
		assert.NoError(t, err)
	})
}
