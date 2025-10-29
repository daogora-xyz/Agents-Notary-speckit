package facilitator

// FacilitatorResponse represents the result of a payment settlement attempt
type FacilitatorResponse struct {
	Status      string `json:"status"`                 // settled | pending | failed
	TxHash      string `json:"tx_hash,omitempty"`      // Transaction hash (if settled)
	BlockNumber uint64 `json:"block_number,omitempty"` // Block number (if settled)
	Error       string `json:"error,omitempty"`        // Error message (if failed)
	RetryAfter  int    `json:"retry_after,omitempty"`  // Seconds until retry (if pending)
}

// ToMap converts the response to a map for MCP tool output
func (r *FacilitatorResponse) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"status": r.Status,
	}

	if r.TxHash != "" {
		result["tx_hash"] = r.TxHash
	}

	if r.BlockNumber > 0 {
		result["block_number"] = r.BlockNumber
	}

	if r.Error != "" {
		result["error"] = r.Error
	}

	if r.RetryAfter > 0 {
		result["retry_after"] = r.RetryAfter
	}

	return result
}
