package errors

import "fmt"

// ValidationError represents an error during data validation
type ValidationError struct {
	Field   string
	Message string
	Wrapped error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on '%s': %s", e.Field, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Wrapped
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) error {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// WrapValidationError wraps an error with validation context
func WrapValidationError(validationErr error, originalErr error) error {
	if ve, ok := validationErr.(*ValidationError); ok {
		ve.Wrapped = originalErr
		return ve
	}
	return &ValidationError{
		Field:   "unknown",
		Message: validationErr.Error(),
		Wrapped: originalErr,
	}
}

// NetworkError represents an error during network communication
type NetworkError struct {
	Network   string
	Message   string
	Retryable bool
	Wrapped   error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error on '%s': %s", e.Network, e.Message)
}

func (e *NetworkError) Unwrap() error {
	return e.Wrapped
}

// NewNetworkError creates a new network error
func NewNetworkError(network, message string) error {
	return &NetworkError{
		Network:   network,
		Message:   message,
		Retryable: false,
	}
}

// WrapNetworkError wraps an error with network context
func WrapNetworkError(networkErr error, originalErr error) error {
	if ne, ok := networkErr.(*NetworkError); ok {
		ne.Wrapped = originalErr
		return ne
	}
	return &NetworkError{
		Network: "unknown",
		Message: networkErr.Error(),
		Wrapped: originalErr,
	}
}

// BlockchainError represents an error during blockchain operations
type BlockchainError struct {
	Chain   string
	Message string
	TxHash  string
	Wrapped error
}

func (e *BlockchainError) Error() string {
	if e.TxHash != "" {
		return fmt.Sprintf("blockchain error on '%s': %s (tx: %s)", e.Chain, e.Message, e.TxHash)
	}
	return fmt.Sprintf("blockchain error on '%s': %s", e.Chain, e.Message)
}

func (e *BlockchainError) Unwrap() error {
	return e.Wrapped
}

// NewBlockchainError creates a new blockchain error
func NewBlockchainError(chain, message, txHash string) error {
	return &BlockchainError{
		Chain:   chain,
		Message: message,
		TxHash:  txHash,
	}
}

// WrapBlockchainError wraps an error with blockchain context
func WrapBlockchainError(blockchainErr error, originalErr error) error {
	if be, ok := blockchainErr.(*BlockchainError); ok {
		be.Wrapped = originalErr
		return be
	}
	return &BlockchainError{
		Chain:   "unknown",
		Message: blockchainErr.Error(),
		Wrapped: originalErr,
	}
}

// PaymentError represents an error during payment processing
type PaymentError struct {
	PaymentNonce string
	Message      string
	Wrapped      error
}

func (e *PaymentError) Error() string {
	return fmt.Sprintf("payment error for nonce '%s': %s", e.PaymentNonce, e.Message)
}

func (e *PaymentError) Unwrap() error {
	return e.Wrapped
}

// NewPaymentError creates a new payment error
func NewPaymentError(paymentNonce, message string) error {
	return &PaymentError{
		PaymentNonce: paymentNonce,
		Message:      message,
	}
}

// WrapPaymentError wraps an error with payment context
func WrapPaymentError(paymentErr error, originalErr error) error {
	if pe, ok := paymentErr.(*PaymentError); ok {
		pe.Wrapped = originalErr
		return pe
	}
	return &PaymentError{
		PaymentNonce: "unknown",
		Message:      paymentErr.Error(),
		Wrapped:      originalErr,
	}
}
