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
// per official Coinbase x402 specification
type PaymentRequirement struct {
	// Core x402 v1 fields (official spec - github.com/coinbase/x402)
	Scheme            string                 `json:"scheme"`
	Network           string                 `json:"network"`
	MaxAmountRequired string                 `json:"maxAmountRequired"`
	Resource          string                 `json:"resource"`
	Description       string                 `json:"description"`
	MimeType          string                 `json:"mimeType"`
	OutputSchema      map[string]interface{} `json:"outputSchema,omitempty"`
	PayTo             string                 `json:"payTo"`
	MaxTimeoutSeconds int                    `json:"maxTimeoutSeconds"`
	Asset             string                 `json:"asset"`
	Extra             ExtraMetadata          `json:"extra"`

	// Extension fields (for backward compatibility and internal use)
	X402Version int    `json:"x402_version"`
	ValidUntil  string `json:"valid_until"`
	Nonce       string `json:"nonce"`
}

// ExtraMetadata contains scheme-specific payment details
// For "exact" scheme on EVM networks: name and version pertain to the asset
type ExtraMetadata struct {
	Name    string `json:"name,omitempty"`    // Asset name (e.g., "USD Coin")
	Version string `json:"version,omitempty"` // Asset version (e.g., "2")
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
// per official Coinbase x402 specification
func NewPaymentRequirement(
	amount string,
	network string,
	payTo string,
	asset string,
	resource string,
	description string,
	mimeType string,
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

	// Validate payTo address
	if !addressPattern.MatchString(payTo) {
		return nil, fmt.Errorf("invalid payTo address format")
	}

	// Validate asset address
	if !addressPattern.MatchString(asset) {
		return nil, fmt.Errorf("invalid asset address format")
	}

	// Validate required fields
	if resource == "" {
		return nil, fmt.Errorf("resource URL is required")
	}
	if description == "" {
		return nil, fmt.Errorf("description is required")
	}
	if mimeType == "" {
		mimeType = "application/json" // Default MIME type
	}

	// Generate unique nonce
	nonce, err := generateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Calculate expiration time
	validUntil := time.Now().UTC().Add(validity)

	return &PaymentRequirement{
		// Official x402 v1 fields
		Scheme:            "exact",
		Network:           network,
		MaxAmountRequired: amount,
		Resource:          resource,
		Description:       description,
		MimeType:          mimeType,
		OutputSchema:      nil, // Optional, can be set by caller
		PayTo:             payTo,
		MaxTimeoutSeconds: 60, // Reasonable default for API responses
		Asset:             asset,
		Extra: ExtraMetadata{
			Name:    "USD Coin", // Standard USDC name
			Version: "2",        // USDC version from EIP-712 domain
		},

		// Extension fields
		X402Version: 1,
		ValidUntil:  validUntil.Format(time.RFC3339),
		Nonce:       nonce,
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
	result := map[string]interface{}{
		// Official x402 v1 fields
		"scheme":            pr.Scheme,
		"network":           pr.Network,
		"maxAmountRequired": pr.MaxAmountRequired,
		"resource":          pr.Resource,
		"description":       pr.Description,
		"mimeType":          pr.MimeType,
		"payTo":             pr.PayTo,
		"maxTimeoutSeconds": pr.MaxTimeoutSeconds,
		"asset":             pr.Asset,
		"extra": map[string]interface{}{
			"name":    pr.Extra.Name,
			"version": pr.Extra.Version,
		},

		// Extension fields
		"x402_version": pr.X402Version,
		"valid_until":  pr.ValidUntil,
		"nonce":        pr.Nonce,
	}

	// Add outputSchema if present
	if pr.OutputSchema != nil {
		result["outputSchema"] = pr.OutputSchema
	}

	return result
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

	// Validate official x402 fields
	if pr.Resource == "" {
		return fmt.Errorf("resource is required")
	}

	if pr.Description == "" {
		return fmt.Errorf("description is required")
	}

	if pr.MimeType == "" {
		return fmt.Errorf("mimeType is required")
	}

	if !addressPattern.MatchString(pr.PayTo) {
		return fmt.Errorf("invalid payTo address format")
	}

	if !addressPattern.MatchString(pr.Asset) {
		return fmt.Errorf("invalid asset address format")
	}

	if pr.MaxTimeoutSeconds <= 0 {
		return fmt.Errorf("maxTimeoutSeconds must be positive")
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
