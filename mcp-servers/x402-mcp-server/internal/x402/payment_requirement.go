package x402

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"time"
)

// PaymentRequirement represents an x402-compliant payment requirement
type PaymentRequirement struct {
	X402Version       int    `json:"x402_version"`
	Scheme            string `json:"scheme"`
	Network           string `json:"network"`
	MaxAmountRequired string `json:"maxAmountRequired"`
	Payee             string `json:"payee"`
	ValidUntil        string `json:"valid_until"`
	Nonce             string `json:"nonce"`
	Asset             string `json:"asset"`
}

var (
	// addressPattern validates Ethereum addresses (0x + 40 hex characters)
	addressPattern = regexp.MustCompile(`^0x[a-fA-F0-9]{40}$`)

	// amountPattern validates positive integer amounts
	amountPattern = regexp.MustCompile(`^[1-9][0-9]*$`)

	// Supported networks
	supportedNetworks = map[string]bool{
		"base":         true,
		"base-sepolia": true,
		"arbitrum":     true,
	}
)

// NewPaymentRequirement creates a new x402-compliant payment requirement
func NewPaymentRequirement(
	amount string,
	network string,
	payee string,
	asset string,
	validity time.Duration,
) (*PaymentRequirement, error) {
	// Validate amount
	if !amountPattern.MatchString(amount) {
		return nil, fmt.Errorf("invalid amount: must be a positive integer")
	}

	// Validate network
	if !supportedNetworks[network] {
		return nil, fmt.Errorf("unsupported network: %s", network)
	}

	// Validate payee address
	if !addressPattern.MatchString(payee) {
		return nil, fmt.Errorf("invalid payee address format")
	}

	// Validate asset address
	if !addressPattern.MatchString(asset) {
		return nil, fmt.Errorf("invalid asset address format")
	}

	// Generate unique nonce
	nonce, err := generateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Calculate expiration time
	validUntil := time.Now().UTC().Add(validity)

	return &PaymentRequirement{
		X402Version:       1,
		Scheme:            "exact",
		Network:           network,
		MaxAmountRequired: amount,
		Payee:             payee,
		ValidUntil:        validUntil.Format(time.RFC3339),
		Nonce:             nonce,
		Asset:             asset,
	}, nil
}

// generateNonce creates a cryptographically secure random nonce
// Combines timestamp with random bytes for uniqueness
func generateNonce() (string, error) {
	// Use 16 bytes of randomness (128 bits)
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Add timestamp component for additional uniqueness
	timestamp := time.Now().UnixNano()
	timestampBig := big.NewInt(timestamp)

	// Combine random bytes with timestamp
	combined := append(randomBytes, timestampBig.Bytes()...)

	// Return as 0x-prefixed hex string
	return "0x" + hex.EncodeToString(combined), nil
}

// ToJSON converts the payment requirement to JSON
func (pr *PaymentRequirement) ToJSON() ([]byte, error) {
	return json.Marshal(pr)
}

// ToMap converts the payment requirement to a map for MCP tool output
func (pr *PaymentRequirement) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"x402_version":       pr.X402Version,
		"scheme":             pr.Scheme,
		"network":            pr.Network,
		"maxAmountRequired":  pr.MaxAmountRequired,
		"payee":              pr.Payee,
		"valid_until":        pr.ValidUntil,
		"nonce":              pr.Nonce,
		"asset":              pr.Asset,
	}
}

// Validate checks if the payment requirement is valid
func (pr *PaymentRequirement) Validate() error {
	if pr.X402Version != 1 {
		return fmt.Errorf("invalid x402_version: expected 1, got %d", pr.X402Version)
	}

	if pr.Scheme != "exact" {
		return fmt.Errorf("invalid scheme: expected 'exact', got %s", pr.Scheme)
	}

	if !supportedNetworks[pr.Network] {
		return fmt.Errorf("unsupported network: %s", pr.Network)
	}

	if !amountPattern.MatchString(pr.MaxAmountRequired) {
		return fmt.Errorf("invalid maxAmountRequired format")
	}

	if !addressPattern.MatchString(pr.Payee) {
		return fmt.Errorf("invalid payee address format")
	}

	if !addressPattern.MatchString(pr.Asset) {
		return fmt.Errorf("invalid asset address format")
	}

	// Validate valid_until is a valid RFC3339 timestamp
	_, err := time.Parse(time.RFC3339, pr.ValidUntil)
	if err != nil {
		return fmt.Errorf("invalid valid_until format: %w", err)
	}

	// Validate nonce is hex-encoded
	if len(pr.Nonce) < 3 || pr.Nonce[:2] != "0x" {
		return fmt.Errorf("invalid nonce format: must be 0x-prefixed hex string")
	}

	return nil
}
