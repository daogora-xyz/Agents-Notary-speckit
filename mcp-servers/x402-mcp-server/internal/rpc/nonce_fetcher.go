package rpc

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// NonceFetcher retrieves transaction nonces from an Ethereum RPC endpoint
type NonceFetcher struct {
	client     *ethclient.Client
	maxRetries int
	retryDelay time.Duration
	timeout    time.Duration
}

// NewNonceFetcher creates a new nonce fetcher with the specified RPC URL
func NewNonceFetcher(rpcURL string) (*NonceFetcher, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}

	return &NonceFetcher{
		client:     client,
		maxRetries: 3,
		retryDelay: 1 * time.Second,
		timeout:    10 * time.Second,
	}, nil
}

// Close closes the RPC client connection
func (nf *NonceFetcher) Close() {
	nf.client.Close()
}

// GetNonce retrieves the pending nonce for the given address
// Uses eth_getTransactionCount with "pending" block parameter
// Implements exponential backoff retry logic
func (nf *NonceFetcher) GetNonce(address common.Address) (*big.Int, error) {
	var lastErr error

	for attempt := 0; attempt <= nf.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			delay := nf.retryDelay * time.Duration(1<<(attempt-1))
			time.Sleep(delay)
		}

		ctx, cancel := context.WithTimeout(context.Background(), nf.timeout)
		defer cancel()

		// Get pending nonce (includes pending transactions)
		nonce, err := nf.client.PendingNonceAt(ctx, address)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d failed: %w", attempt+1, err)
			continue
		}

		// Success
		return big.NewInt(int64(nonce)), nil
	}

	// All retries exhausted
	return nil, fmt.Errorf("failed after %d attempts: %w", nf.maxRetries+1, lastErr)
}

// GetNonceAt retrieves the nonce at a specific block number
func (nf *NonceFetcher) GetNonceAt(address common.Address, blockNumber *big.Int) (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), nf.timeout)
	defer cancel()

	nonce, err := nf.client.NonceAt(ctx, address, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce at block %s: %w", blockNumber.String(), err)
	}

	return big.NewInt(int64(nonce)), nil
}

// GetChainID retrieves the chain ID from the RPC endpoint
func (nf *NonceFetcher) GetChainID() (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), nf.timeout)
	defer cancel()

	chainID, err := nf.client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	return chainID, nil
}

// VerifyConnection checks if the RPC connection is working
func (nf *NonceFetcher) VerifyConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), nf.timeout)
	defer cancel()

	// Try to get the latest block number as a connection test
	_, err := nf.client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("RPC connection verification failed: %w", err)
	}

	return nil
}
