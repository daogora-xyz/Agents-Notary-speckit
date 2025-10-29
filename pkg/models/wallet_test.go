package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalletBalanceValidation(t *testing.T) {
	t.Run("valid wallet balance", func(t *testing.T) {
		wallet := &WalletBalance{
			Asset:         "CIRX",
			Network:       "circular",
			WalletAddress: "cirx1abcdef1234567890",
			Balance:       "100.500000",
			LastUpdated:   time.Now(),
		}

		err := wallet.Validate()
		assert.NoError(t, err, "valid wallet balance should pass validation")
	})

	t.Run("missing asset", func(t *testing.T) {
		wallet := &WalletBalance{
			Asset:         "", // Missing
			Network:       "circular",
			WalletAddress: "cirx1abcdef1234567890",
			Balance:       "100.500000",
			LastUpdated:   time.Now(),
		}

		err := wallet.Validate()
		require.Error(t, err, "wallet balance without asset should fail validation")
		assert.Contains(t, err.Error(), "asset")
	})

	t.Run("missing network", func(t *testing.T) {
		wallet := &WalletBalance{
			Asset:         "CIRX",
			Network:       "", // Missing
			WalletAddress: "cirx1abcdef1234567890",
			Balance:       "100.500000",
			LastUpdated:   time.Now(),
		}

		err := wallet.Validate()
		require.Error(t, err, "wallet balance without network should fail validation")
		assert.Contains(t, err.Error(), "network")
	})

	t.Run("missing wallet_address", func(t *testing.T) {
		wallet := &WalletBalance{
			Asset:         "CIRX",
			Network:       "circular",
			WalletAddress: "", // Missing
			Balance:       "100.500000",
			LastUpdated:   time.Now(),
		}

		err := wallet.Validate()
		require.Error(t, err, "wallet balance without wallet_address should fail validation")
		assert.Contains(t, err.Error(), "wallet_address")
	})

	t.Run("invalid balance (empty)", func(t *testing.T) {
		wallet := &WalletBalance{
			Asset:         "CIRX",
			Network:       "circular",
			WalletAddress: "cirx1abcdef1234567890",
			Balance:       "", // Invalid
			LastUpdated:   time.Now(),
		}

		err := wallet.Validate()
		require.Error(t, err, "wallet balance with empty balance should fail validation")
		assert.Contains(t, err.Error(), "balance")
	})

	t.Run("invalid balance (negative)", func(t *testing.T) {
		wallet := &WalletBalance{
			Asset:         "CIRX",
			Network:       "circular",
			WalletAddress: "cirx1abcdef1234567890",
			Balance:       "-10.50", // Invalid
			LastUpdated:   time.Now(),
		}

		err := wallet.Validate()
		require.Error(t, err, "wallet balance with negative balance should fail validation")
		assert.Contains(t, err.Error(), "balance")
	})

	t.Run("valid zero balance", func(t *testing.T) {
		wallet := &WalletBalance{
			Asset:         "CIRX",
			Network:       "circular",
			WalletAddress: "cirx1abcdef1234567890",
			Balance:       "0.000000", // Valid (zero is acceptable)
			LastUpdated:   time.Now(),
		}

		err := wallet.Validate()
		assert.NoError(t, err, "wallet balance with zero should be valid (for monitoring)")
	})

	t.Run("multiple assets on same network", func(t *testing.T) {
		assets := []string{"CIRX", "USDC", "ETH"}

		for _, asset := range assets {
			wallet := &WalletBalance{
				Asset:         asset,
				Network:       "ethereum",
				WalletAddress: "0x1234567890abcdef1234567890abcdef12345678",
				Balance:       "100.000000",
				LastUpdated:   time.Now(),
			}

			err := wallet.Validate()
			assert.NoError(t, err, "asset %s should be valid", asset)
		}
	})

	t.Run("same asset on multiple networks", func(t *testing.T) {
		networks := []string{"ethereum", "polygon", "base", "arbitrum", "optimism"}

		for _, network := range networks {
			wallet := &WalletBalance{
				Asset:         "USDC",
				Network:       network,
				WalletAddress: "0x1234567890abcdef1234567890abcdef12345678",
				Balance:       "100.000000",
				LastUpdated:   time.Now(),
			}

			err := wallet.Validate()
			assert.NoError(t, err, "network %s should be valid", network)
		}
	})

	t.Run("high precision balance", func(t *testing.T) {
		wallet := &WalletBalance{
			Asset:         "CIRX",
			Network:       "circular",
			WalletAddress: "cirx1abcdef1234567890",
			Balance:       "123456789.123456789012345678", // 18 decimal places
			LastUpdated:   time.Now(),
		}

		err := wallet.Validate()
		assert.NoError(t, err, "high precision balance should be valid")
	})
}

func TestWalletBalanceUniqueness(t *testing.T) {
	t.Run("uniqueness by asset, network, and address", func(t *testing.T) {
		// This test documents the unique constraint requirement
		// Actual uniqueness is enforced at the database level
		// Here we just verify the model can represent different combinations

		wallet1 := &WalletBalance{
			Asset:         "CIRX",
			Network:       "circular",
			WalletAddress: "cirx1address1",
			Balance:       "100.0",
			LastUpdated:   time.Now(),
		}

		wallet2 := &WalletBalance{
			Asset:         "CIRX",
			Network:       "circular",
			WalletAddress: "cirx1address2", // Different address
			Balance:       "200.0",
			LastUpdated:   time.Now(),
		}

		wallet3 := &WalletBalance{
			Asset:         "USDC",
			Network:       "circular",
			WalletAddress: "cirx1address1", // Different asset
			Balance:       "300.0",
			LastUpdated:   time.Now(),
		}

		wallet4 := &WalletBalance{
			Asset:         "CIRX",
			Network:       "ethereum", // Different network
			WalletAddress: "cirx1address1",
			Balance:       "400.0",
			LastUpdated:   time.Now(),
		}

		// All should be valid individually
		assert.NoError(t, wallet1.Validate())
		assert.NoError(t, wallet2.Validate())
		assert.NoError(t, wallet3.Validate())
		assert.NoError(t, wallet4.Validate())
	})
}
