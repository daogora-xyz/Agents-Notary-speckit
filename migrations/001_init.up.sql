-- Migration: 001_init
-- Description: Initialize database schema with certification_requests, payments, certifications, and wallet_balances tables
-- Created: 2025-10-28

-- Table: certification_requests
-- Purpose: Track certification requests from clients through their lifecycle
CREATE TABLE certification_requests (
    id BIGSERIAL PRIMARY KEY,
    request_id TEXT NOT NULL UNIQUE,
    client_id TEXT NOT NULL,
    data_hash TEXT NOT NULL,
    data_size_bytes BIGINT NOT NULL CHECK (data_size_bytes > 0),
    status TEXT NOT NULL CHECK (status IN ('pending', 'payment_received', 'certifying', 'completed', 'failed')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for querying by client_id
CREATE INDEX idx_certification_requests_client_id ON certification_requests(client_id);

-- Index for querying by status
CREATE INDEX idx_certification_requests_status ON certification_requests(status);

-- Index for time-based queries
CREATE INDEX idx_certification_requests_created_at ON certification_requests(created_at DESC);

-- Table: payments
-- Purpose: Track payment authorizations for certification services
CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    request_id TEXT NOT NULL,
    payment_nonce TEXT NOT NULL UNIQUE,
    from_address TEXT NOT NULL,
    to_address TEXT NOT NULL,
    amount_usdc DECIMAL(20, 6) NOT NULL CHECK (amount_usdc > 0),
    network TEXT NOT NULL CHECK (network IN ('ethereum', 'polygon', 'base', 'arbitrum', 'optimism')),
    evm_tx_hash TEXT,
    status TEXT NOT NULL CHECK (status IN ('pending', 'authorized', 'settled', 'failed')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (request_id) REFERENCES certification_requests(request_id) ON DELETE CASCADE
);

-- Index for querying by request_id
CREATE INDEX idx_payments_request_id ON payments(request_id);

-- Index for querying by status
CREATE INDEX idx_payments_status ON payments(status);

-- Index for time-based queries
CREATE INDEX idx_payments_created_at ON payments(created_at DESC);

-- Table: certifications
-- Purpose: Track blockchain certification transactions on Circular Protocol
CREATE TABLE certifications (
    id BIGSERIAL PRIMARY KEY,
    request_id TEXT NOT NULL,
    cirx_tx_id TEXT UNIQUE,
    cirx_block_id TEXT,
    cirx_fee_paid DECIMAL(20, 6),
    status TEXT NOT NULL CHECK (status IN ('pending', 'submitted', 'confirmed', 'failed')),
    retry_count INTEGER NOT NULL DEFAULT 0 CHECK (retry_count >= 0),
    last_error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (request_id) REFERENCES certification_requests(request_id) ON DELETE CASCADE
);

-- Index for querying by request_id
CREATE INDEX idx_certifications_request_id ON certifications(request_id);

-- Index for querying by status
CREATE INDEX idx_certifications_status ON certifications(status);

-- Index for time-based queries
CREATE INDEX idx_certifications_created_at ON certifications(created_at DESC);

-- Table: wallet_balances
-- Purpose: Track service wallet balances across blockchain networks
CREATE TABLE wallet_balances (
    id BIGSERIAL PRIMARY KEY,
    asset TEXT NOT NULL,
    network TEXT NOT NULL,
    wallet_address TEXT NOT NULL,
    balance DECIMAL(30, 18) NOT NULL CHECK (balance >= 0),
    last_updated TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (asset, network, wallet_address)
);

-- Index for querying by asset and network
CREATE INDEX idx_wallet_balances_asset_network ON wallet_balances(asset, network);

-- Index for time-based queries
CREATE INDEX idx_wallet_balances_last_updated ON wallet_balances(last_updated DESC);
