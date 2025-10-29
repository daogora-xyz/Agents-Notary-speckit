package models

import (
	"fmt"
	"strconv"
	"time"
)

// PaymentStatus represents the lifecycle status of a payment
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusAuthorized PaymentStatus = "authorized"
	PaymentStatusSettled    PaymentStatus = "settled"
	PaymentStatusFailed     PaymentStatus = "failed"
)

// ValidPaymentStatuses lists all valid payment statuses
var ValidPaymentStatuses = []PaymentStatus{
	PaymentStatusPending,
	PaymentStatusAuthorized,
	PaymentStatusSettled,
	PaymentStatusFailed,
}

// Network represents a blockchain network
type Network string

const (
	NetworkEthereum Network = "ethereum"
	NetworkPolygon  Network = "polygon"
	NetworkBase     Network = "base"
	NetworkArbitrum Network = "arbitrum"
	NetworkOptimism Network = "optimism"
)

// ValidNetworks lists all valid networks
var ValidNetworks = []Network{
	NetworkEthereum,
	NetworkPolygon,
	NetworkBase,
	NetworkArbitrum,
	NetworkOptimism,
}

// Payment represents a payment authorization for certification service
type Payment struct {
	ID           int64         `json:"id" db:"id"`
	RequestID    string        `json:"request_id" db:"request_id"`
	PaymentNonce string        `json:"payment_nonce" db:"payment_nonce"`
	FromAddress  string        `json:"from_address" db:"from_address"`
	ToAddress    string        `json:"to_address" db:"to_address"`
	AmountUSDC   string        `json:"amount_usdc" db:"amount_usdc"` // DECIMAL stored as string for precision
	Network      Network       `json:"network" db:"network"`
	EVMTxHash    string        `json:"evm_tx_hash,omitempty" db:"evm_tx_hash"`
	Status       PaymentStatus `json:"status" db:"status"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at" db:"updated_at"`
}

// Validate checks if the Payment has all required fields and valid values
func (p *Payment) Validate() error {
	if p.RequestID == "" {
		return fmt.Errorf("request_id is required")
	}

	if p.PaymentNonce == "" {
		return fmt.Errorf("payment_nonce is required")
	}

	if p.FromAddress == "" {
		return fmt.Errorf("from_address is required")
	}

	if p.ToAddress == "" {
		return fmt.Errorf("to_address is required")
	}

	if p.AmountUSDC == "" {
		return fmt.Errorf("amount_usdc is required")
	}

	// Validate amount is positive
	amount, err := strconv.ParseFloat(p.AmountUSDC, 64)
	if err != nil {
		return fmt.Errorf("amount_usdc must be a valid number: %w", err)
	}
	if amount <= 0 {
		return fmt.Errorf("amount_usdc must be positive (got: %s)", p.AmountUSDC)
	}

	// Validate network is one of the valid networks
	validNetwork := false
	for _, validNet := range ValidNetworks {
		if p.Network == validNet {
			validNetwork = true
			break
		}
	}
	if !validNetwork {
		return fmt.Errorf("invalid network '%s' (valid: %v)", p.Network, ValidNetworks)
	}

	// Validate status is one of the valid statuses
	validStatus := false
	for _, validSt := range ValidPaymentStatuses {
		if p.Status == validSt {
			validStatus = true
			break
		}
	}
	if !validStatus {
		return fmt.Errorf("invalid status '%s' (valid: %v)", p.Status, ValidPaymentStatuses)
	}

	return nil
}
