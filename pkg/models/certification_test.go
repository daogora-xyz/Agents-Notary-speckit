package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCertificationValidation(t *testing.T) {
	t.Run("valid certification (confirmed)", func(t *testing.T) {
		cert := &Certification{
			RequestID:    "req_test_12345",
			CIRXTxID:     "cirx_tx_abc123",
			CIRXBlockID:  "cirx_block_456",
			CIRXFeePaid:  "4.000000", // Fixed 4 CIRX fee
			Status:       CertStatusConfirmed,
			RetryCount:   0,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := cert.Validate()
		assert.NoError(t, err, "valid certification should pass validation")
	})

	t.Run("valid certification (pending)", func(t *testing.T) {
		cert := &Certification{
			RequestID:  "req_test_12345",
			Status:     CertStatusPending,
			RetryCount: 0,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		err := cert.Validate()
		assert.NoError(t, err, "pending certification without tx_id should be valid")
	})

	t.Run("missing request_id", func(t *testing.T) {
		cert := &Certification{
			RequestID:  "", // Missing
			Status:     CertStatusPending,
			RetryCount: 0,
		}

		err := cert.Validate()
		require.Error(t, err, "certification without request_id should fail validation")
		assert.Contains(t, err.Error(), "request_id")
	})

	t.Run("invalid status", func(t *testing.T) {
		cert := &Certification{
			RequestID:  "req_test_12345",
			Status:     "invalid_status", // Invalid
			RetryCount: 0,
		}

		err := cert.Validate()
		require.Error(t, err, "certification with invalid status should fail validation")
		assert.Contains(t, err.Error(), "status")
	})

	t.Run("all valid statuses", func(t *testing.T) {
		validStatuses := []CertificationStatus{
			CertStatusPending,
			CertStatusSubmitted,
			CertStatusConfirmed,
			CertStatusFailed,
		}

		for _, status := range validStatuses {
			cert := &Certification{
				RequestID:  "req_test_12345",
				Status:     status,
				RetryCount: 0,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			err := cert.Validate()
			assert.NoError(t, err, "status %s should be valid", status)
		}
	})

	t.Run("negative retry_count", func(t *testing.T) {
		cert := &Certification{
			RequestID:  "req_test_12345",
			Status:     CertStatusPending,
			RetryCount: -1, // Invalid
		}

		err := cert.Validate()
		require.Error(t, err, "certification with negative retry_count should fail validation")
		assert.Contains(t, err.Error(), "retry_count")
	})

	t.Run("valid retry counts", func(t *testing.T) {
		for i := 0; i <= 5; i++ {
			cert := &Certification{
				RequestID:  "req_test_12345",
				Status:     CertStatusPending,
				RetryCount: i,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			err := cert.Validate()
			assert.NoError(t, err, "retry_count %d should be valid", i)
		}
	})

	t.Run("failed certification with error message", func(t *testing.T) {
		cert := &Certification{
			RequestID:  "req_test_12345",
			Status:     CertStatusFailed,
			RetryCount: 3,
			LastError:  "Transaction failed: insufficient CIRX balance",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		err := cert.Validate()
		assert.NoError(t, err, "failed certification with last_error should be valid")
	})

	t.Run("confirmed certification with all fields", func(t *testing.T) {
		cert := &Certification{
			RequestID:    "req_test_12345",
			CIRXTxID:     "cirx_tx_abc123",
			CIRXBlockID:  "cirx_block_456",
			CIRXFeePaid:  "4.000000",
			Status:       CertStatusConfirmed,
			RetryCount:   1, // Succeeded after 1 retry
			LastError:    "",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := cert.Validate()
		assert.NoError(t, err, "confirmed certification with all fields should be valid")
	})
}

func TestCertificationStatusLifecycle(t *testing.T) {
	t.Run("pending to submitted", func(t *testing.T) {
		cert := &Certification{
			RequestID:  "req_test_12345",
			Status:     CertStatusPending,
			RetryCount: 0,
		}

		cert.Status = CertStatusSubmitted
		cert.CIRXTxID = "cirx_tx_abc123"

		err := cert.Validate()
		assert.NoError(t, err, "transition from pending to submitted should be valid")
	})

	t.Run("submitted to confirmed", func(t *testing.T) {
		cert := &Certification{
			RequestID:   "req_test_12345",
			CIRXTxID:    "cirx_tx_abc123",
			Status:      CertStatusSubmitted,
			RetryCount:  0,
		}

		cert.Status = CertStatusConfirmed
		cert.CIRXBlockID = "cirx_block_456"
		cert.CIRXFeePaid = "4.000000"

		err := cert.Validate()
		assert.NoError(t, err, "transition from submitted to confirmed should be valid")
	})

	t.Run("pending to failed (with error)", func(t *testing.T) {
		cert := &Certification{
			RequestID:  "req_test_12345",
			Status:     CertStatusPending,
			RetryCount: 0,
		}

		cert.Status = CertStatusFailed
		cert.RetryCount = 3
		cert.LastError = "Maximum retries exceeded"

		err := cert.Validate()
		assert.NoError(t, err, "transition to failed with error should be valid")
	})
}
