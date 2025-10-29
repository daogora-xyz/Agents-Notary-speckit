package facilitator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/config"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/eip3009"
)

// Client handles interaction with the x402 facilitator API
type Client struct {
	config     *config.Config
	httpClient *http.Client
	cache      *settlementCache
}

// settlementCache provides idempotency via nonce-based caching
type settlementCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	ttl     time.Duration
}

type cacheEntry struct {
	response  *FacilitatorResponse
	timestamp time.Time
}

// NewClient creates a new facilitator client
func NewClient(cfg *config.Config, timeout time.Duration) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		cache: &settlementCache{
			entries: make(map[string]*cacheEntry),
			ttl:     time.Duration(cfg.Cache.SettlementTTLMinutes) * time.Minute,
		},
	}
}

// BuildSettlementRequest constructs the JSON request body for facilitator submission
func (c *Client) BuildSettlementRequest(auth *eip3009.EIP3009Authorization, network string) ([]byte, error) {
	// Validate authorization
	if err := auth.Validate(); err != nil {
		return nil, fmt.Errorf("invalid authorization: %w", err)
	}

	// Build request body matching EIP-3009 receiveWithAuthorization parameters
	requestBody := map[string]interface{}{
		"from":        auth.From,
		"to":          auth.To,
		"value":       auth.Value,
		"validAfter":  auth.ValidAfter,
		"validBefore": auth.ValidBefore,
		"nonce":       auth.Nonce,
		"v":           auth.V,
		"r":           auth.R,
		"s":           auth.S,
	}

	return json.Marshal(requestBody)
}

// SubmitSettlement submits a payment authorization to the x402 facilitator
func (c *Client) SubmitSettlement(auth *eip3009.EIP3009Authorization, network string) (*FacilitatorResponse, error) {
	// Check cache for idempotency
	if cached := c.cache.get(auth.Nonce); cached != nil {
		return cached, nil
	}

	// Get network configuration
	networkCfg, exists := c.config.Networks[network]
	if !exists {
		return nil, fmt.Errorf("unsupported network: %s", network)
	}

	// Build request body
	requestBody, err := c.BuildSettlementRequest(auth, network)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	// Create HTTP POST request
	req, err := http.NewRequest(http.MethodPost, networkCfg.FacilitatorURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Submit request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("facilitator request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	result, err := c.parseResponse(resp.StatusCode, body)
	if err != nil {
		return nil, err
	}

	// Cache successful settlements
	if result.Status == "settled" {
		c.cache.set(auth.Nonce, result)
	}

	return result, nil
}

// parseResponse parses the facilitator HTTP response
func (c *Client) parseResponse(statusCode int, body []byte) (*FacilitatorResponse, error) {
	var response FacilitatorResponse

	// Handle different status codes
	switch {
	case statusCode == http.StatusOK || statusCode == http.StatusAccepted:
		// Parse JSON response
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		return &response, nil

	case statusCode == http.StatusBadRequest:
		// Parse error response
		if err := json.Unmarshal(body, &response); err != nil {
			// If parsing fails, return generic error
			return nil, fmt.Errorf("facilitator returned 400 Bad Request: %s", string(body))
		}
		// Return parsed error response
		if response.Status == "" {
			response.Status = "failed"
		}
		return &response, nil

	case statusCode >= 500:
		// Server error
		return nil, fmt.Errorf("facilitator server error (%d): %s", statusCode, string(body))

	default:
		// Unexpected status code
		return nil, fmt.Errorf("unexpected facilitator response (%d): %s", statusCode, string(body))
	}
}

// get retrieves a cached settlement result by nonce
func (sc *settlementCache) get(nonce string) *FacilitatorResponse {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	entry, exists := sc.entries[nonce]
	if !exists {
		return nil
	}

	// Check if entry has expired
	if time.Since(entry.timestamp) > sc.ttl {
		// Entry expired, will be cleaned up later
		return nil
	}

	return entry.response
}

// set stores a settlement result in cache
func (sc *settlementCache) set(nonce string, response *FacilitatorResponse) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.entries[nonce] = &cacheEntry{
		response:  response,
		timestamp: time.Now(),
	}

	// Cleanup expired entries (simple inline cleanup)
	sc.cleanup()
}

// cleanup removes expired entries from cache
func (sc *settlementCache) cleanup() {
	now := time.Now()
	for nonce, entry := range sc.entries {
		if now.Sub(entry.timestamp) > sc.ttl {
			delete(sc.entries, nonce)
		}
	}
}
