package models

import (
	"fmt"
	"time"
)

// RequestStatus represents the lifecycle status of a certification request
type RequestStatus string

const (
	StatusPending         RequestStatus = "pending"
	StatusPaymentReceived RequestStatus = "payment_received"
	StatusCertifying      RequestStatus = "certifying"
	StatusCompleted       RequestStatus = "completed"
	StatusFailed          RequestStatus = "failed"
)

// ValidRequestStatuses lists all valid request statuses
var ValidRequestStatuses = []RequestStatus{
	StatusPending,
	StatusPaymentReceived,
	StatusCertifying,
	StatusCompleted,
	StatusFailed,
}

// CertificationRequest represents a user's request to certify data on the blockchain
type CertificationRequest struct {
	ID            int64         `json:"id" db:"id"`
	RequestID     string        `json:"request_id" db:"request_id"`
	ClientID      string        `json:"client_id" db:"client_id"`
	DataHash      string        `json:"data_hash" db:"data_hash"`
	DataSizeBytes int64         `json:"data_size_bytes" db:"data_size_bytes"`
	Status        RequestStatus `json:"status" db:"status"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
}

// Validate checks if the CertificationRequest has all required fields and valid values
func (r *CertificationRequest) Validate() error {
	if r.RequestID == "" {
		return fmt.Errorf("request_id is required")
	}

	if r.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}

	if r.DataHash == "" {
		return fmt.Errorf("data_hash is required")
	}

	if r.DataSizeBytes <= 0 {
		return fmt.Errorf("data_size_bytes must be positive (got: %d)", r.DataSizeBytes)
	}

	// Validate status is one of the valid statuses
	validStatus := false
	for _, validSt := range ValidRequestStatuses {
		if r.Status == validSt {
			validStatus = true
			break
		}
	}
	if !validStatus {
		return fmt.Errorf("invalid status '%s' (valid: %v)", r.Status, ValidRequestStatuses)
	}

	return nil
}
