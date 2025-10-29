package config

import (
	"fmt"
	"regexp"
)

// NetworkConfig contains network-specific parameters for payment processing
type NetworkConfig struct {
	ChainID        uint64 `yaml:"chain_id"`        // EIP-155 chain ID
	USDCContract   string `yaml:"usdc_contract"`   // Native USDC address
	FacilitatorURL string `yaml:"facilitator_url"` // x402 facilitator endpoint
	RPCURL         string `yaml:"rpc_url"`         // Blockchain RPC for nonces
	PayeeAddress   string `yaml:"payee_address"`   // Certification service payee
}

// Allowed chain IDs per data-model.md validation rules
var allowedChainIDs = map[uint64]bool{
	8453:  true, // Base
	84532: true, // Base Sepolia
	42161: true, // Arbitrum
}

// Ethereum address pattern: 0x prefix + 40 hex characters
var addressPattern = regexp.MustCompile(`^0x[a-fA-F0-9]{40}$`)

// URL pattern: http or https
var urlPattern = regexp.MustCompile(`^https?://`)

// Validate checks that all required network config fields are valid
func (n *NetworkConfig) Validate() error {
	// Chain ID must be in allowlist
	if !allowedChainIDs[n.ChainID] {
		return fmt.Errorf("chain_id %d not in allowed list (8453, 84532, 42161)", n.ChainID)
	}

	// USDC contract must be valid Ethereum address
	if !addressPattern.MatchString(n.USDCContract) {
		return fmt.Errorf("usdc_contract must be valid Ethereum address (0x + 40 hex chars)")
	}

	// Payee address must be valid Ethereum address
	if !addressPattern.MatchString(n.PayeeAddress) {
		return fmt.Errorf("payee_address must be valid Ethereum address (0x + 40 hex chars)")
	}

	// RPC URL must be valid HTTP/HTTPS URL
	if !urlPattern.MatchString(n.RPCURL) {
		return fmt.Errorf("rpc_url must be valid HTTP/HTTPS URL")
	}

	// Facilitator URL must be valid HTTP/HTTPS URL
	if !urlPattern.MatchString(n.FacilitatorURL) {
		return fmt.Errorf("facilitator_url must be valid HTTP/HTTPS URL")
	}

	return nil
}
