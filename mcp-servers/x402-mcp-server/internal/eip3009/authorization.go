package eip3009

import (
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"

	"github.com/ethereum/go-ethereum/common"
)

// EIP3009Authorization represents the payment authorization data structure
// per EIP-3009 receiveWithAuthorization spec
type EIP3009Authorization struct {
	From        string `json:"from"`         // Payer address (hex string)
	To          string `json:"to"`           // Payee address (hex string)
	Value       string `json:"value"`        // Amount in atomic units (decimal string)
	ValidAfter  uint64 `json:"validAfter"`   // Unix timestamp (seconds)
	ValidBefore uint64 `json:"validBefore"`  // Unix timestamp (seconds)
	Nonce       string `json:"nonce"`        // bytes32 as hex string
	V           uint8  `json:"v"`            // Signature parameter (27 or 28)
	R           string `json:"r"`            // Signature parameter (bytes32 hex)
	S           string `json:"s"`            // Signature parameter (bytes32 hex)
}

// VerifyPaymentOutput represents the verification result
type VerifyPaymentOutput struct {
	IsValid       bool   `json:"is_valid"`
	SignerAddress string `json:"signer_address,omitempty"` // Recovered from signature
	Error         string `json:"error,omitempty"`
}

var (
	// addressPattern validates Ethereum addresses
	addressPatternAuth = regexp.MustCompile(`^0x[a-fA-F0-9]{40}$`)

	// bytes32Pattern validates 32-byte hex strings
	bytes32Pattern = regexp.MustCompile(`^0x[a-fA-F0-9]{64}$`)

	// amountPatternAuth validates positive integer amounts
	amountPatternAuth = regexp.MustCompile(`^[1-9][0-9]*$`)
)

// Validate performs input validation on the authorization
func (a *EIP3009Authorization) Validate() error {
	// Validate From address
	if !addressPatternAuth.MatchString(a.From) {
		return fmt.Errorf("invalid from address format: %s", a.From)
	}

	// Validate To address
	if !addressPatternAuth.MatchString(a.To) {
		return fmt.Errorf("invalid to address format: %s", a.To)
	}

	// Validate Value
	if !amountPatternAuth.MatchString(a.Value) {
		return fmt.Errorf("invalid value format: must be positive integer string")
	}

	// Validate Nonce format
	if !bytes32Pattern.MatchString(a.Nonce) {
		return fmt.Errorf("invalid nonce format: must be 32-byte hex string")
	}

	// Validate V parameter
	if a.V != 27 && a.V != 28 {
		return fmt.Errorf("invalid v parameter: must be 27 or 28, got %d", a.V)
	}

	// Validate R parameter
	if !bytes32Pattern.MatchString(a.R) {
		return fmt.Errorf("invalid r parameter: must be 32-byte hex string")
	}

	// Validate S parameter
	if !bytes32Pattern.MatchString(a.S) {
		return fmt.Errorf("invalid s parameter: must be 32-byte hex string")
	}

	// Validate time bounds
	if a.ValidAfter >= a.ValidBefore {
		return fmt.Errorf("validAfter must be less than validBefore")
	}

	return nil
}

// ToMessage converts the authorization to an EIP-712 message struct
func (a *EIP3009Authorization) ToMessage() (*ReceiveWithAuthorizationMessage, error) {
	// Convert addresses
	from := common.HexToAddress(a.From)
	to := common.HexToAddress(a.To)

	// Convert value
	value, ok := new(big.Int).SetString(a.Value, 10)
	if !ok {
		return nil, fmt.Errorf("failed to parse value: %s", a.Value)
	}

	// Convert time bounds
	validAfter := new(big.Int).SetUint64(a.ValidAfter)
	validBefore := new(big.Int).SetUint64(a.ValidBefore)

	// Convert nonce
	nonceBytes := common.FromHex(a.Nonce)
	if len(nonceBytes) != 32 {
		return nil, fmt.Errorf("nonce must be 32 bytes, got %d", len(nonceBytes))
	}

	var nonce [32]byte
	copy(nonce[:], nonceBytes)

	return &ReceiveWithAuthorizationMessage{
		From:        from,
		To:          to,
		Value:       value,
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
		Nonce:       nonce,
	}, nil
}

// GetSignature returns the signature components in the format expected by crypto.Sign
func (a *EIP3009Authorization) GetSignature() ([]byte, error) {
	// Parse R and S
	rBytes := common.FromHex(a.R)
	sBytes := common.FromHex(a.S)

	if len(rBytes) != 32 {
		return nil, fmt.Errorf("R must be 32 bytes, got %d", len(rBytes))
	}
	if len(sBytes) != 32 {
		return nil, fmt.Errorf("S must be 32 bytes, got %d", len(sBytes))
	}

	// Construct signature: R || S || V
	// Note: crypto.Sign uses v as 0 or 1, but Ethereum uses 27 or 28
	// We need to convert back
	signature := make([]byte, 65)
	copy(signature[0:32], rBytes)
	copy(signature[32:64], sBytes)

	// Convert v from 27/28 to 0/1 for go-ethereum
	if a.V == 27 {
		signature[64] = 0
	} else if a.V == 28 {
		signature[64] = 1
	} else {
		return nil, fmt.Errorf("invalid v value: %d", a.V)
	}

	return signature, nil
}

// ToJSON converts the authorization to JSON
func (a *EIP3009Authorization) ToJSON() ([]byte, error) {
	return json.Marshal(a)
}

// ToMap converts the verification output to a map for MCP response
func (v *VerifyPaymentOutput) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"is_valid": v.IsValid,
	}

	if v.SignerAddress != "" {
		result["signer_address"] = v.SignerAddress
	}

	if v.Error != "" {
		result["error"] = v.Error
	}

	return result
}
