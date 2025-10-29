# Data Model: Foundation Infrastructure

**Feature**: Project Foundation Infrastructure
**Branch**: 001-foundation-setup
**Date**: 2025-10-28

## Overview

This document specifies the database schema for the four core tables required by the blockchain certification platform, along with Go model structs, validation rules, and relationships.

## Entity Relationship Diagram

```
┌─────────────────────────┐
│ certification_requests  │
│ ──────────────────────  │
│ id (PK)                 │
│ request_id (UNIQUE)     │───┐
│ client_id               │   │
│ data_hash               │   │ 1:N
│ data_size_bytes         │   │
│ status                  │   │
│ created_at              │   │
│ updated_at              │   │
└─────────────────────────┘   │
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌─────────────────┐  ┌──────────────────┐  ┌─────────────────────┐
│    payments     │  │ certifications   │  │  wallet_balances    │
│ ─────────────── │  │ ──────────────── │  │  ───────────────── │
│ id (PK)         │  │ id (PK)          │  │ id (PK)            │
│ request_id (FK) │  │ request_id (FK)  │  │ asset              │
│ payment_nonce   │  │ cirx_tx_id       │  │ network            │
│ from_address    │  │ cirx_block_id    │  │ wallet_address     │
│ to_address      │  │ cirx_fee_paid    │  │ balance            │
│ amount_usdc     │  │ status           │  │ last_updated       │
│ network         │  │ retry_count      │  └────────────────────┘
│ evm_tx_hash     │  │ created_at       │
│ status          │  │ updated_at       │
│ created_at      │  └──────────────────┘
│ updated_at      │
└─────────────────┘

Legend: PK = Primary Key, FK = Foreign Key, UNIQUE = Unique constraint
```

## Database Schema (PostgreSQL)

### Table 1: `certification_requests`

**Purpose**: Tracks the lifecycle of a certification request from initiation through payment to completion.

**Schema**:
```sql
CREATE TABLE certification_requests (
    id                  BIGSERIAL PRIMARY KEY,
    request_id          TEXT NOT NULL UNIQUE,  -- Client-provided idempotency key
    client_id           TEXT NOT NULL,         -- Identifies the requesting client/agent
    data_hash           TEXT NOT NULL,         -- SHA-256 hash of data to certify (hex-encoded)
    data_size_bytes     BIGINT NOT NULL,       -- Size of original data in bytes
    status              TEXT NOT NULL,         -- 'initiated', 'payment_pending', 'payment_verified', 'certifying', 'completed', 'failed'
    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for query performance
CREATE INDEX idx_certification_requests_request_id ON certification_requests(request_id);
CREATE INDEX idx_certification_requests_client_id ON certification_requests(client_id);
CREATE INDEX idx_certification_requests_status ON certification_requests(status);
CREATE INDEX idx_certification_requests_created_at ON certification_requests(created_at DESC);

-- Constraints
ALTER TABLE certification_requests
    ADD CONSTRAINT chk_status CHECK (status IN ('initiated', 'payment_pending', 'payment_verified', 'certifying', 'completed', 'failed'));
ALTER TABLE certification_requests
    ADD CONSTRAINT chk_data_size_bytes CHECK (data_size_bytes > 0 AND data_size_bytes <= 10485760); -- Max 10MB per spec
```

**Go Model**:
```go
package models

import (
    "errors"
    "regexp"
    "time"
)

type CertificationRequestStatus string

const (
    StatusInitiated        CertificationRequestStatus = "initiated"
    StatusPaymentPending   CertificationRequestStatus = "payment_pending"
    StatusPaymentVerified  CertificationRequestStatus = "payment_verified"
    StatusCertifying       CertificationRequestStatus = "certifying"
    StatusCompleted        CertificationRequestStatus = "completed"
    StatusFailed           CertificationRequestStatus = "failed"
)

type CertificationRequest struct {
    ID            int64                      `json:"id" db:"id"`
    RequestID     string                     `json:"request_id" db:"request_id"`
    ClientID      string                     `json:"client_id" db:"client_id"`
    DataHash      string                     `json:"data_hash" db:"data_hash"`
    DataSizeBytes int64                      `json:"data_size_bytes" db:"data_size_bytes"`
    Status        CertificationRequestStatus `json:"status" db:"status"`
    CreatedAt     time.Time                  `json:"created_at" db:"created_at"`
    UpdatedAt     time.Time                  `json:"updated_at" db:"updated_at"`
}

var sha256Regex = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

func (r *CertificationRequest) Validate() error {
    if r.RequestID == "" {
        return errors.New("request_id is required")
    }
    if r.ClientID == "" {
        return errors.New("client_id is required")
    }
    if !sha256Regex.MatchString(r.DataHash) {
        return errors.New("data_hash must be a valid SHA-256 hex string (64 characters)")
    }
    if r.DataSizeBytes <= 0 {
        return errors.New("data_size_bytes must be positive")
    }
    if r.DataSizeBytes > 10485760 { // 10MB
        return errors.New("data_size_bytes exceeds maximum of 10MB")
    }
    validStatuses := []CertificationRequestStatus{
        StatusInitiated, StatusPaymentPending, StatusPaymentVerified,
        StatusCertifying, StatusCompleted, StatusFailed,
    }
    valid := false
    for _, s := range validStatuses {
        if r.Status == s {
            valid = true
            break
        }
    }
    if !valid {
        return errors.New("invalid status value")
    }
    return nil
}
```

---

### Table 2: `payments`

**Purpose**: Tracks payment authorizations (EIP-3009) for certification services.

**Schema**:
```sql
CREATE TABLE payments (
    id              BIGSERIAL PRIMARY KEY,
    request_id      TEXT NOT NULL,         -- Links to certification_requests.request_id
    payment_nonce   TEXT NOT NULL UNIQUE,  -- EIP-3009 nonce for idempotency/replay protection
    from_address    TEXT NOT NULL,         -- Ethereum address of payer (0x...)
    to_address      TEXT NOT NULL,         -- Ethereum address of payee (service wallet)
    amount_usdc     NUMERIC(20, 6) NOT NULL, -- Amount in USDC (6 decimals)
    network         TEXT NOT NULL,         -- 'base', 'base-sepolia', 'arbitrum'
    evm_tx_hash     TEXT,                  -- EVM transaction hash after settlement (nullable until settled)
    status          TEXT NOT NULL,         -- 'pending', 'verified', 'settled', 'failed'
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),

    FOREIGN KEY (request_id) REFERENCES certification_requests(request_id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_payments_request_id ON payments(request_id);
CREATE INDEX idx_payments_payment_nonce ON payments(payment_nonce);
CREATE INDEX idx_payments_status ON payments(status);

-- Constraints
ALTER TABLE payments
    ADD CONSTRAINT chk_payment_status CHECK (status IN ('pending', 'verified', 'settled', 'failed'));
ALTER TABLE payments
    ADD CONSTRAINT chk_network CHECK (network IN ('base', 'base-sepolia', 'arbitrum'));
ALTER TABLE payments
    ADD CONSTRAINT chk_amount_usdc CHECK (amount_usdc > 0);
```

**Go Model**:
```go
package models

import (
    "errors"
    "regexp"
    "strings"
    "time"
)

type PaymentStatus string

const (
    PaymentPending  PaymentStatus = "pending"
    PaymentVerified PaymentStatus = "verified"
    PaymentSettled  PaymentStatus = "settled"
    PaymentFailed   PaymentStatus = "failed"
)

type Payment struct {
    ID            int64         `json:"id" db:"id"`
    RequestID     string        `json:"request_id" db:"request_id"`
    PaymentNonce  string        `json:"payment_nonce" db:"payment_nonce"`
    FromAddress   string        `json:"from_address" db:"from_address"`
    ToAddress     string        `json:"to_address" db:"to_address"`
    AmountUSDC    float64       `json:"amount_usdc" db:"amount_usdc"`
    Network       string        `json:"network" db:"network"`
    EVMTxHash     *string       `json:"evm_tx_hash,omitempty" db:"evm_tx_hash"` // Nullable
    Status        PaymentStatus `json:"status" db:"status"`
    CreatedAt     time.Time     `json:"created_at" db:"created_at"`
    UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
}

var ethAddressRegex = regexp.MustCompile(`^0x[a-fA-F0-9]{40}$`)

func (p *Payment) Validate() error {
    if p.RequestID == "" {
        return errors.New("request_id is required")
    }
    if p.PaymentNonce == "" {
        return errors.New("payment_nonce is required")
    }
    if !ethAddressRegex.MatchString(p.FromAddress) {
        return errors.New("from_address must be a valid Ethereum address (0x...)")
    }
    if !ethAddressRegex.MatchString(p.ToAddress) {
        return errors.New("to_address must be a valid Ethereum address (0x...)")
    }
    if p.AmountUSDC <= 0 {
        return errors.New("amount_usdc must be positive")
    }
    validNetworks := []string{"base", "base-sepolia", "arbitrum"}
    networkValid := false
    for _, n := range validNetworks {
        if strings.ToLower(p.Network) == n {
            networkValid = true
            break
        }
    }
    if !networkValid {
        return errors.New("network must be one of: base, base-sepolia, arbitrum")
    }
    validStatuses := []PaymentStatus{PaymentPending, PaymentVerified, PaymentSettled, PaymentFailed}
    statusValid := false
    for _, s := range validStatuses {
        if p.Status == s {
            statusValid = true
            break
        }
    }
    if !statusValid {
        return errors.New("invalid payment status")
    }
    return nil
}
```

---

### Table 3: `certifications`

**Purpose**: Records completed blockchain certification transactions on Circular Protocol.

**Schema**:
```sql
CREATE TABLE certifications (
    id              BIGSERIAL PRIMARY KEY,
    request_id      TEXT NOT NULL,         -- Links to certification_requests.request_id
    cirx_tx_id      TEXT NOT NULL UNIQUE,  -- Circular Protocol transaction ID
    cirx_block_id   TEXT NOT NULL,         -- Circular Protocol block ID containing the transaction
    cirx_fee_paid   NUMERIC(20, 6) NOT NULL, -- CIRX fee paid (4 CIRX per spec)
    status          TEXT NOT NULL,         -- 'pending', 'executed', 'failed'
    retry_count     INT NOT NULL DEFAULT 0, -- Number of retry attempts (max 10 per spec)
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),

    FOREIGN KEY (request_id) REFERENCES certification_requests(request_id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_certifications_request_id ON certifications(request_id);
CREATE INDEX idx_certifications_cirx_tx_id ON certifications(cirx_tx_id);
CREATE INDEX idx_certifications_status ON certifications(status);

-- Constraints
ALTER TABLE certifications
    ADD CONSTRAINT chk_certification_status CHECK (status IN ('pending', 'executed', 'failed'));
ALTER TABLE certifications
    ADD CONSTRAINT chk_retry_count CHECK (retry_count >= 0 AND retry_count <= 10);
ALTER TABLE certifications
    ADD CONSTRAINT chk_cirx_fee_paid CHECK (cirx_fee_paid > 0);
```

**Go Model**:
```go
package models

import (
    "errors"
    "time"
)

type CertificationStatus string

const (
    CertPending  CertificationStatus = "pending"
    CertExecuted CertificationStatus = "executed"
    CertFailed   CertificationStatus = "failed"
)

type Certification struct {
    ID           int64               `json:"id" db:"id"`
    RequestID    string              `json:"request_id" db:"request_id"`
    CIRXTxID     string              `json:"cirx_tx_id" db:"cirx_tx_id"`
    CIRXBlockID  string              `json:"cirx_block_id" db:"cirx_block_id"`
    CIRXFeePaid  float64             `json:"cirx_fee_paid" db:"cirx_fee_paid"`
    Status       CertificationStatus `json:"status" db:"status"`
    RetryCount   int                 `json:"retry_count" db:"retry_count"`
    CreatedAt    time.Time           `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time           `json:"updated_at" db:"updated_at"`
}

func (c *Certification) Validate() error {
    if c.RequestID == "" {
        return errors.New("request_id is required")
    }
    if c.CIRXTxID == "" {
        return errors.New("cirx_tx_id is required")
    }
    if c.CIRXBlockID == "" {
        return errors.New("cirx_block_id is required")
    }
    if c.CIRXFeePaid <= 0 {
        return errors.New("cirx_fee_paid must be positive")
    }
    if c.RetryCount < 0 || c.RetryCount > 10 {
        return errors.New("retry_count must be between 0 and 10")
    }
    validStatuses := []CertificationStatus{CertPending, CertExecuted, CertFailed}
    statusValid := false
    for _, s := range validStatuses {
        if c.Status == s {
            statusValid = true
            break
        }
    }
    if !statusValid {
        return errors.New("invalid certification status")
    }
    return nil
}
```

---

### Table 4: `wallet_balances`

**Purpose**: Tracks service wallet balances across different blockchain networks for monitoring and alerting.

**Schema**:
```sql
CREATE TABLE wallet_balances (
    id              BIGSERIAL PRIMARY KEY,
    asset           TEXT NOT NULL,         -- 'CIRX', 'USDC'
    network         TEXT NOT NULL,         -- 'circular-mainnet', 'circular-testnet', 'base', 'base-sepolia', 'arbitrum'
    wallet_address  TEXT NOT NULL,         -- Blockchain address of service wallet
    balance         NUMERIC(30, 6) NOT NULL, -- Current balance (6 decimals precision)
    last_updated    TIMESTAMP NOT NULL DEFAULT NOW(),

    UNIQUE(asset, network, wallet_address)  -- One balance record per asset/network/address combo
);

-- Indexes
CREATE INDEX idx_wallet_balances_asset ON wallet_balances(asset);
CREATE INDEX idx_wallet_balances_network ON wallet_balances(network);
CREATE INDEX idx_wallet_balances_last_updated ON wallet_balances(last_updated DESC);

-- Constraints
ALTER TABLE wallet_balances
    ADD CONSTRAINT chk_balance CHECK (balance >= 0);
```

**Go Model**:
```go
package models

import (
    "errors"
    "time"
)

type WalletBalance struct {
    ID            int64     `json:"id" db:"id"`
    Asset         string    `json:"asset" db:"asset"`
    Network       string    `json:"network" db:"network"`
    WalletAddress string    `json:"wallet_address" db:"wallet_address"`
    Balance       float64   `json:"balance" db:"balance"`
    LastUpdated   time.Time `json:"last_updated" db:"last_updated"`
}

func (w *WalletBalance) Validate() error {
    if w.Asset == "" {
        return errors.New("asset is required")
    }
    if w.Network == "" {
        return errors.New("network is required")
    }
    if w.WalletAddress == "" {
        return errors.New("wallet_address is required")
    }
    if w.Balance < 0 {
        return errors.New("balance cannot be negative")
    }
    return nil
}
```

---

## Migration Files

### Up Migration: `001_init.up.sql`

```sql
-- Migration: 001_init
-- Description: Create initial database schema for certification platform

CREATE TABLE certification_requests (
    id                  BIGSERIAL PRIMARY KEY,
    request_id          TEXT NOT NULL UNIQUE,
    client_id           TEXT NOT NULL,
    data_hash           TEXT NOT NULL,
    data_size_bytes     BIGINT NOT NULL,
    status              TEXT NOT NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_certification_requests_request_id ON certification_requests(request_id);
CREATE INDEX idx_certification_requests_client_id ON certification_requests(client_id);
CREATE INDEX idx_certification_requests_status ON certification_requests(status);
CREATE INDEX idx_certification_requests_created_at ON certification_requests(created_at DESC);

ALTER TABLE certification_requests
    ADD CONSTRAINT chk_status CHECK (status IN ('initiated', 'payment_pending', 'payment_verified', 'certifying', 'completed', 'failed'));
ALTER TABLE certification_requests
    ADD CONSTRAINT chk_data_size_bytes CHECK (data_size_bytes > 0 AND data_size_bytes <= 10485760);

CREATE TABLE payments (
    id              BIGSERIAL PRIMARY KEY,
    request_id      TEXT NOT NULL,
    payment_nonce   TEXT NOT NULL UNIQUE,
    from_address    TEXT NOT NULL,
    to_address      TEXT NOT NULL,
    amount_usdc     NUMERIC(20, 6) NOT NULL,
    network         TEXT NOT NULL,
    evm_tx_hash     TEXT,
    status          TEXT NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),

    FOREIGN KEY (request_id) REFERENCES certification_requests(request_id) ON DELETE CASCADE
);

CREATE INDEX idx_payments_request_id ON payments(request_id);
CREATE INDEX idx_payments_payment_nonce ON payments(payment_nonce);
CREATE INDEX idx_payments_status ON payments(status);

ALTER TABLE payments
    ADD CONSTRAINT chk_payment_status CHECK (status IN ('pending', 'verified', 'settled', 'failed'));
ALTER TABLE payments
    ADD CONSTRAINT chk_network CHECK (network IN ('base', 'base-sepolia', 'arbitrum'));
ALTER TABLE payments
    ADD CONSTRAINT chk_amount_usdc CHECK (amount_usdc > 0);

CREATE TABLE certifications (
    id              BIGSERIAL PRIMARY KEY,
    request_id      TEXT NOT NULL,
    cirx_tx_id      TEXT NOT NULL UNIQUE,
    cirx_block_id   TEXT NOT NULL,
    cirx_fee_paid   NUMERIC(20, 6) NOT NULL,
    status          TEXT NOT NULL,
    retry_count     INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),

    FOREIGN KEY (request_id) REFERENCES certification_requests(request_id) ON DELETE CASCADE
);

CREATE INDEX idx_certifications_request_id ON certifications(request_id);
CREATE INDEX idx_certifications_cirx_tx_id ON certifications(cirx_tx_id);
CREATE INDEX idx_certifications_status ON certifications(status);

ALTER TABLE certifications
    ADD CONSTRAINT chk_certification_status CHECK (status IN ('pending', 'executed', 'failed'));
ALTER TABLE certifications
    ADD CONSTRAINT chk_retry_count CHECK (retry_count >= 0 AND retry_count <= 10);
ALTER TABLE certifications
    ADD CONSTRAINT chk_cirx_fee_paid CHECK (cirx_fee_paid > 0);

CREATE TABLE wallet_balances (
    id              BIGSERIAL PRIMARY KEY,
    asset           TEXT NOT NULL,
    network         TEXT NOT NULL,
    wallet_address  TEXT NOT NULL,
    balance         NUMERIC(30, 6) NOT NULL,
    last_updated    TIMESTAMP NOT NULL DEFAULT NOW(),

    UNIQUE(asset, network, wallet_address)
);

CREATE INDEX idx_wallet_balances_asset ON wallet_balances(asset);
CREATE INDEX idx_wallet_balances_network ON wallet_balances(network);
CREATE INDEX idx_wallet_balances_last_updated ON wallet_balances(last_updated DESC);

ALTER TABLE wallet_balances
    ADD CONSTRAINT chk_balance CHECK (balance >= 0);
```

### Down Migration: `001_init.down.sql`

```sql
-- Rollback: 001_init
-- Description: Drop all tables created in 001_init.up.sql

DROP TABLE IF EXISTS wallet_balances CASCADE;
DROP TABLE IF EXISTS certifications CASCADE;
DROP TABLE IF EXISTS payments CASCADE;
DROP TABLE IF EXISTS certification_requests CASCADE;
```

---

## Relationships

1. **certification_requests → payments**: One-to-many (one request can have multiple payment attempts)
2. **certification_requests → certifications**: One-to-many (one request can have multiple certification attempts due to retries)
3. **wallet_balances**: Standalone table (no foreign keys), updated periodically by monitoring job

---

## Validation Summary

| Model | Required Fields | Format Validation | Business Rules |
|-------|----------------|-------------------|----------------|
| CertificationRequest | request_id, client_id, data_hash, data_size_bytes, status | data_hash: 64-char hex (SHA-256) | data_size_bytes: 0-10MB, status: enum |
| Payment | request_id, payment_nonce, from/to_address, amount_usdc, network, status | addresses: 0x+40 hex chars | network: allowlist, amount > 0 |
| Certification | request_id, cirx_tx_id, cirx_block_id, cirx_fee_paid, status | None | retry_count: 0-10, fee > 0 |
| WalletBalance | asset, network, wallet_address, balance | None | balance >= 0 |

---

## Query Patterns (for future reference)

```sql
-- Find all pending certifications for retry queue
SELECT * FROM certifications WHERE status = 'pending' AND retry_count < 10;

-- Find certification request with all related payments and certifications
SELECT cr.*, p.*, c.*
FROM certification_requests cr
LEFT JOIN payments p ON cr.request_id = p.request_id
LEFT JOIN certifications c ON cr.request_id = c.request_id
WHERE cr.request_id = '...';

-- Get CIRX balance for alerting (< 100 threshold per spec)
SELECT balance FROM wallet_balances
WHERE asset = 'CIRX' AND network = 'circular-mainnet';
```

---

## Testing Strategy

- **Unit Tests**: Model validation logic (pkg/models/*_test.go)
- **Integration Tests**: Database CRUD operations with dockertest (tests/integration/models_test.go)
- **Migration Tests**: Up/down migration success and rollback integrity (tests/integration/migration_test.go)
