package eip3009

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// EIP712Domain represents the EIP-712 domain separator parameters
type EIP712Domain struct {
	Name              string
	Version           string
	ChainID           *big.Int
	VerifyingContract common.Address
}

// DomainSeparator computes the EIP-712 domain separator hash
// Per EIP-712: keccak256(typeHash || encodeData(domain))
func (d *EIP712Domain) DomainSeparator() common.Hash {
	// EIP712Domain typeHash
	// keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)")
	typeHash := crypto.Keccak256Hash([]byte(
		"EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)",
	))

	// Encode domain struct data
	nameHash := crypto.Keccak256Hash([]byte(d.Name))
	versionHash := crypto.Keccak256Hash([]byte(d.Version))

	// Pack: typeHash || nameHash || versionHash || chainId || verifyingContract
	packed := make([]byte, 0, 160) // 32 + 32 + 32 + 32 + 32
	packed = append(packed, typeHash.Bytes()...)
	packed = append(packed, nameHash.Bytes()...)
	packed = append(packed, versionHash.Bytes()...)
	packed = append(packed, common.LeftPadBytes(d.ChainID.Bytes(), 32)...)
	packed = append(packed, common.LeftPadBytes(d.VerifyingContract.Bytes(), 32)...)

	return crypto.Keccak256Hash(packed)
}

// ReceiveWithAuthorizationMessage represents the EIP-3009 authorization message
type ReceiveWithAuthorizationMessage struct {
	From        common.Address
	To          common.Address
	Value       *big.Int
	ValidAfter  *big.Int
	ValidBefore *big.Int
	Nonce       [32]byte
}

// TypeHash returns the EIP-712 type hash for ReceiveWithAuthorization
func (m *ReceiveWithAuthorizationMessage) TypeHash() common.Hash {
	return crypto.Keccak256Hash([]byte(
		"ReceiveWithAuthorization(address from,address to,uint256 value,uint256 validAfter,uint256 validBefore,bytes32 nonce)",
	))
}

// StructHash computes the EIP-712 struct hash for the message
// Per EIP-712: keccak256(typeHash || encodeData(message))
func (m *ReceiveWithAuthorizationMessage) StructHash() common.Hash {
	typeHash := m.TypeHash()

	// Pack: typeHash || from || to || value || validAfter || validBefore || nonce
	packed := make([]byte, 0, 224) // 32 + 32 + 32 + 32 + 32 + 32 + 32
	packed = append(packed, typeHash.Bytes()...)
	packed = append(packed, common.LeftPadBytes(m.From.Bytes(), 32)...)
	packed = append(packed, common.LeftPadBytes(m.To.Bytes(), 32)...)
	packed = append(packed, common.LeftPadBytes(m.Value.Bytes(), 32)...)
	packed = append(packed, common.LeftPadBytes(m.ValidAfter.Bytes(), 32)...)
	packed = append(packed, common.LeftPadBytes(m.ValidBefore.Bytes(), 32)...)
	packed = append(packed, m.Nonce[:]...)

	return crypto.Keccak256Hash(packed)
}

// TypedDataHash computes the full EIP-712 typed data hash
// Per EIP-712: keccak256("\x19\x01" || domainSeparator || structHash)
func TypedDataHash(domain *EIP712Domain, message *ReceiveWithAuthorizationMessage) (common.Hash, error) {
	if domain == nil {
		return common.Hash{}, fmt.Errorf("domain cannot be nil")
	}
	if message == nil {
		return common.Hash{}, fmt.Errorf("message cannot be nil")
	}

	domainSeparator := domain.DomainSeparator()
	structHash := message.StructHash()

	// Pack: 0x19 0x01 || domainSeparator || structHash
	packed := make([]byte, 0, 66) // 2 + 32 + 32
	packed = append(packed, 0x19, 0x01)
	packed = append(packed, domainSeparator.Bytes()...)
	packed = append(packed, structHash.Bytes()...)

	return crypto.Keccak256Hash(packed), nil
}
