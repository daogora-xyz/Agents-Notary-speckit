package models

import (
	"fmt"
	"strconv"
	"time"
)

// WalletBalance represents the current balance of service wallets across different blockchain networks
type WalletBalance struct {
	ID            int64     `json:"id" db:"id"`
	Asset         string    `json:"asset" db:"asset"`
	Network       string    `json:"network" db:"network"`
	WalletAddress string    `json:"wallet_address" db:"wallet_address"`
	Balance       string    `json:"balance" db:"balance"` // DECIMAL stored as string for precision
	LastUpdated   time.Time `json:"last_updated" db:"last_updated"`
}

// Validate checks if the WalletBalance has all required fields and valid values
func (w *WalletBalance) Validate() error {
	if w.Asset == "" {
		return fmt.Errorf("asset is required")
	}

	if w.Network == "" {
		return fmt.Errorf("network is required")
	}

	if w.WalletAddress == "" {
		return fmt.Errorf("wallet_address is required")
	}

	if w.Balance == "" {
		return fmt.Errorf("balance is required")
	}

	// Validate balance is non-negative
	balance, err := strconv.ParseFloat(w.Balance, 64)
	if err != nil {
		return fmt.Errorf("balance must be a valid number: %w", err)
	}
	if balance < 0 {
		return fmt.Errorf("balance must be non-negative (got: %s)", w.Balance)
	}

	return nil
}
