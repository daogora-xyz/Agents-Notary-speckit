package models

import (
	"fmt"
	"time"
)

// CertificationStatus represents the lifecycle status of a blockchain certification
type CertificationStatus string

const (
	CertStatusPending   CertificationStatus = "pending"
	CertStatusSubmitted CertificationStatus = "submitted"
	CertStatusConfirmed CertificationStatus = "confirmed"
	CertStatusFailed    CertificationStatus = "failed"
)

// ValidCertificationStatuses lists all valid certification statuses
var ValidCertificationStatuses = []CertificationStatus{
	CertStatusPending,
	CertStatusSubmitted,
	CertStatusConfirmed,
	CertStatusFailed,
}

// Certification represents a completed blockchain certification transaction on Circular Protocol
type Certification struct {
	ID          int64               `json:"id" db:"id"`
	RequestID   string              `json:"request_id" db:"request_id"`
	CIRXTxID    string              `json:"cirx_tx_id,omitempty" db:"cirx_tx_id"`
	CIRXBlockID string              `json:"cirx_block_id,omitempty" db:"cirx_block_id"`
	CIRXFeePaid string              `json:"cirx_fee_paid,omitempty" db:"cirx_fee_paid"` // DECIMAL stored as string for precision
	Status      CertificationStatus `json:"status" db:"status"`
	RetryCount  int                 `json:"retry_count" db:"retry_count"`
	LastError   string              `json:"last_error,omitempty" db:"last_error"`
	CreatedAt   time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at" db:"updated_at"`
}

// Validate checks if the Certification has all required fields and valid values
func (c *Certification) Validate() error {
	if c.RequestID == "" {
		return fmt.Errorf("request_id is required")
	}

	// Validate status is one of the valid statuses
	validStatus := false
	for _, validSt := range ValidCertificationStatuses {
		if c.Status == validSt {
			validStatus = true
			break
		}
	}
	if !validStatus {
		return fmt.Errorf("invalid status '%s' (valid: %v)", c.Status, ValidCertificationStatuses)
	}

	// Validate retry_count is non-negative
	if c.RetryCount < 0 {
		return fmt.Errorf("retry_count must be non-negative (got: %d)", c.RetryCount)
	}

	return nil
}
