# Data Model: Circular Protocol MCP Server

**Date**: 2025-10-30
**Feature**: [spec.md](./spec.md) | [plan.md](./plan.md) | [research.md](./research.md)
**Status**: Phase 1 Complete

## Overview

This document defines all data entities, structures, and relationships for the Circular Protocol MCP Server. The system handles blockchain certification operations through four primary entities: CertificationTransaction, TransactionStatus, CertificationProof, and NetworkConfig.

**Architecture**: The MCP server is stateless - all data is transient within tool execution context. No persistent storage or database is used.

---

## Core Entities

### 1. CertificationTransaction

**Purpose**: Represents a C_TYPE_CERTIFICATE transaction submitted to Circular Protocol blockchain for immutable data certification.

**Lifecycle**: Created by `certify_data` tool → Submitted to blockchain → Tracked by transaction ID

**Structure**:
```go
type CertificationTransaction struct {
    // Transaction Identification
    TransactionID string `json:"transaction_id"` // Assigned by blockchain (not calculated)

    // Transaction Content
    From          string `json:"from"`           // Sender wallet address
    To            string `json:"to,omitempty"`   // Recipient address (may be blockchain ID)
    Payload       string `json:"payload"`        // Hex-encoded certification data
    Timestamp     string `json:"timestamp"`      // Format: YYYY:MM:DD-HH:MM:SS

    // Transaction Metadata
    Type          string `json:"type"`           // "certificate"
    Nonce         int64  `json:"nonce"`          // Wallet nonce (from updateAccount)
    Signature     string `json:"signature"`      // Secp256k1 signature (hex-encoded)

    // Network Context
    Network       string `json:"network"`        // "testnet" or "mainnet"
    BlockchainID  string `json:"blockchain_id,omitempty"` // Target blockchain identifier
}
```

**Field Constraints**:
- `Payload`: Maximum 1 MB (1,048,576 bytes) when decoded from hex
- `Timestamp`: Must use Circular Protocol format `YYYY:MM:DD-HH:MM:SS`
- `Nonce`: Sequential counter, must match wallet's current nonce
- `Signature`: 65 bytes (130 hex chars) - R (32) + S (32) + V (1)
- `TransactionID`: Assigned by blockchain after submission (empty before submission)

**Validation Rules**:
```go
func (tx *CertificationTransaction) Validate() error {
    // Check required fields
    if tx.From == "" {
        return errors.New("from address is required")
    }
    if tx.Payload == "" {
        return errors.New("payload is required")
    }

    // Check payload size limit (1 MB)
    payloadBytes, err := hex.DecodeString(tx.Payload)
    if err != nil {
        return fmt.Errorf("invalid hex payload: %w", err)
    }
    if len(payloadBytes) > 1048576 {
        return errors.New("payload exceeds 1 MB limit")
    }

    // Check nonce (must be non-negative)
    if tx.Nonce < 0 {
        return errors.New("nonce must be non-negative")
    }

    // Check signature presence
    if tx.Signature == "" {
        return errors.New("signature is required")
    }

    return nil
}
```

**Usage Example** (from `certify_data` tool):
```go
tx := &CertificationTransaction{
    From:      wallet.Address,
    Payload:   hexEncodedData,
    Timestamp: getFormattedTimestamp(),
    Type:      "certificate",
    Nonce:     wallet.Nonce,
    Network:   "testnet",
}

// Sign transaction
signature, err := signer.Sign(tx, privateKey)
tx.Signature = signature

// Submit to blockchain
response, err := client.SubmitCertificate(tx)
tx.TransactionID = response.TransactionID
```

---

### 2. TransactionStatus

**Purpose**: Represents the current state of a transaction in the Circular Protocol blockchain lifecycle.

**Lifecycle**: Polling target for `get_transaction_status` tool

**Structure**:
```go
type TransactionStatus struct {
    // Identification
    TransactionID string `json:"transaction_id"`

    // Status Information
    Status        string `json:"status"`        // "Pending" | "Verified" | "Executed" | "Failed"
    BlockID       string `json:"block_id,omitempty"` // Set when status = "Executed"
    Timestamp     string `json:"timestamp,omitempty"` // Block timestamp (when executed)

    // Progress Tracking
    Confirmations int    `json:"confirmations"` // Number of block confirmations
    ExecutedAt    string `json:"executed_at,omitempty"` // RFC3339 timestamp

    // Network Context
    Network       string `json:"network"`       // "testnet" | "mainnet"
}
```

**Status State Machine**:
```
┌──────────┐
│ Pending  │ ← Transaction in mempool
└────┬─────┘
     │
     ▼
┌──────────┐
│ Verified │ ← Transaction included in block
└────┬─────┘
     │
     ▼
┌──────────┐
│ Executed │ ← Transaction finalized (SUCCESS)
└──────────┘

     OR

┌──────────┐
│  Failed  │ ← Transaction rejected (ERROR)
└──────────┘
```

**Status Descriptions**:

| Status | Meaning | Agent Action | Next State |
|--------|---------|--------------|------------|
| `Pending` | In mempool, awaiting block inclusion | Continue polling | Verified or Failed |
| `Verified` | Included in block, awaiting finality | Continue polling | Executed |
| `Executed` | Finalized, immutable on blockchain | Success - get proof | Terminal state |
| `Failed` | Rejected by blockchain | Error - investigate | Terminal state |

**Polling Logic** (from spec.md FR-006a):
```go
func PollTransactionStatus(txID string, network string) (*TransactionStatus, error) {
    const maxAttempts = 12 // 60 seconds / 5 seconds per attempt
    const pollInterval = 5 * time.Second

    for attempt := 0; attempt < maxAttempts; attempt++ {
        status, err := client.GetTransactionStatus(txID, network)
        if err != nil {
            return nil, err
        }

        // Terminal states
        if status.Status == "Executed" || status.Status == "Failed" {
            return status, nil
        }

        // Wait before next poll
        time.Sleep(pollInterval)
    }

    return nil, errors.New("transaction status polling timeout after 60 seconds")
}
```

**Usage Example** (from `get_transaction_status` tool):
```go
status, err := PollTransactionStatus(txID, network)
if err != nil {
    return ErrorResponse{
        Type:    "TRANSACTION_TIMEOUT",
        Message: "Transaction did not reach Executed status within 60 seconds",
    }
}

if status.Status == "Failed" {
    return ErrorResponse{
        Type:    "TRANSACTION_FAILED",
        Message: "Transaction was rejected by blockchain",
    }
}

// Success case
return status // Status = "Executed"
```

---

### 3. CertificationProof

**Purpose**: Verifiable evidence that data was certified on blockchain at a specific time.

**Lifecycle**: Generated by `get_certification_proof` tool after transaction reaches "Executed" status

**Structure**:
```go
type CertificationProof struct {
    // Transaction Reference
    TransactionID string `json:"transaction_id"`
    BlockID       string `json:"block_id"`

    // Temporal Evidence
    Timestamp     string `json:"timestamp"`     // Block timestamp (RFC3339)
    BlockHeight   int64  `json:"block_height,omitempty"` // Optional: block number

    // Verification Links
    ExplorerURL   string `json:"explorer_url"`  // Blockchain explorer link
    Network       string `json:"network"`       // "testnet" | "mainnet"

    // Certificate Data (optional, for convenience)
    CertifiedData string `json:"certified_data,omitempty"` // Original payload (hex)
    Sender        string `json:"sender,omitempty"` // Wallet address
}
```

**Generation Logic**:
```go
func GenerateCertificationProof(txID string, network string) (*CertificationProof, error) {
    // 1. Get transaction details
    tx, err := client.GetTransaction(txID, network)
    if err != nil {
        return nil, err
    }

    // 2. Verify transaction is executed
    if tx.Status != "Executed" {
        return nil, errors.New("transaction not yet executed")
    }

    // 3. Generate explorer URL
    explorerURL := fmt.Sprintf(
        "https://circularlabs.io/Explorer?network=%s&tx=%s",
        network,
        txID,
    )

    // 4. Assemble proof
    proof := &CertificationProof{
        TransactionID: txID,
        BlockID:       tx.BlockID,
        Timestamp:     tx.Timestamp,
        ExplorerURL:   explorerURL,
        Network:       network,
        CertifiedData: tx.Payload, // Optional
        Sender:        tx.From,     // Optional
    }

    return proof, nil
}
```

**Usage Example** (from `get_certification_proof` tool):
```go
proof, err := GenerateCertificationProof(txID, network)
if err != nil {
    return ErrorResponse{
        Type:    "PROOF_GENERATION_FAILED",
        Message: err.Error(),
    }
}

// Return proof to agent
return proof
```

**Verification Process** (external party):
1. Receive `CertificationProof` JSON
2. Visit `explorer_url` to view transaction on blockchain
3. Compare `block_id` and `timestamp` with blockchain data
4. Verify `transaction_id` matches certified data hash

---

### 4. Nonce

**Purpose**: Sequential counter for wallet transactions, ensures transaction ordering and prevents replay attacks.

**Lifecycle**: Retrieved via `get_wallet_nonce` tool before transaction submission

**Structure**:
```go
type Nonce struct {
    // Wallet Reference
    WalletAddress string `json:"wallet_address"`

    // Nonce Value
    CurrentNonce  int64  `json:"current_nonce"`  // Next nonce to use

    // Network Context
    Network       string `json:"network"`         // "testnet" | "mainnet"

    // Metadata
    LastUpdated   string `json:"last_updated"`    // RFC3339 timestamp
}
```

**Nonce Management**:
```go
func GetWalletNonce(address string, network string) (*Nonce, error) {
    // Call Circular Protocol API: updateAccount()
    account, err := client.UpdateAccount(address, network)
    if err != nil {
        return nil, err
    }

    nonce := &Nonce{
        WalletAddress: address,
        CurrentNonce:  account.Nonce,
        Network:       network,
        LastUpdated:   time.Now().Format(time.RFC3339),
    }

    return nonce, nil
}
```

**Nonce Incrementation**:
- Nonce automatically increments on Circular Protocol after each transaction
- Agent must fetch fresh nonce before each `certify_data` call
- Reusing nonces causes transaction rejection

**Usage Example** (from `get_wallet_nonce` tool):
```go
nonce, err := GetWalletNonce(walletAddress, network)
if err != nil {
    return ErrorResponse{
        Type:    "NONCE_FETCH_FAILED",
        Message: "Unable to retrieve wallet nonce",
    }
}

// Agent uses nonce.CurrentNonce for next transaction
return nonce
```

---

### 5. NetworkConfig

**Purpose**: Configuration for Circular Protocol network environments (testnet, mainnet).

**Lifecycle**: Loaded at server startup from `config.yaml`

**Structure**:
```go
type NetworkConfig struct {
    // Network Identification
    Name        string `yaml:"name" json:"name"`               // "Circular Testnet"
    NetworkID   string `yaml:"network_id" json:"network_id"`   // "testnet" | "mainnet"

    // Enterprise API Endpoints (NAG Discovery)
    NAGDiscoveryURL string `yaml:"nag_discovery_url" json:"nag_discovery_url"` // NAG URL discovery endpoint
    BaseURL         string `yaml:"base_url" json:"base_url"`                   // Discovered REST API base URL
    ExplorerURL     string `yaml:"explorer_url" json:"explorer_url"`           // Blockchain explorer

    // Blockchain Identification
    BlockchainID string `yaml:"blockchain_id" json:"blockchain_id"` // Testnet sandbox or mainnet ID

    // Wallet Configuration
    PayeeAddress string `yaml:"payee_address" json:"payee_address"` // Default sender address

    // Network Parameters (optional)
    ChainID     int    `yaml:"chain_id,omitempty" json:"chain_id,omitempty"`
    CurrencySymbol string `yaml:"currency_symbol" json:"currency_symbol"` // "CIRX"
}

type Config struct {
    Networks map[string]*NetworkConfig `yaml:"networks"`
    LogLevel string                    `yaml:"log_level"`
}
```

**Configuration File** (`config.yaml.example`):
```yaml
networks:
  circular-testnet:
    name: "Circular Testnet (Sandbox)"
    network_id: "testnet"
    nag_discovery_url: "https://circularlabs.io/network/getNAG"
    blockchain_id: "0x8a20baa40c45dc5055aeb26197c203e576ef389d9acb171bd62da11dc5ad72b2"  # Sandbox chain
    explorer_url: "https://circularlabs.io/Explorer?network=testnet"
    payee_address: "${CIRCULAR_CEP_TESTNET_ADDRESS}"
    currency_symbol: "CIRX"
    # base_url is dynamically discovered via NAG: getNAG?network=testnet

  circular-mainnet:
    name: "Circular Mainnet"
    network_id: "mainnet"
    nag_discovery_url: "https://circularlabs.io/network/getNAG"
    blockchain_id: "mainnet"  # Discovered via NAG
    explorer_url: "https://circularlabs.io/Explorer?network=mainnet"
    payee_address: "${CIRCULAR_CEP_MAINNET_ADDRESS}"
    currency_symbol: "CIRX"
    # base_url is dynamically discovered via NAG: getNAG?network=mainnet

log_level: "info"
```

**NAG Discovery Process**:
1. Query `{nag_discovery_url}?network={network_id}` at startup
2. Parse response: `{"status": "success", "url": "https://nag.circularlabs.io/NAG.php?cep="}`
3. Construct endpoint URLs: `{nag_url}Circular_{MethodName}_{network}`
   - Example: `https://nag.circularlabs.io/NAG.php?cep=Circular_GetWalletNonce_testnet`

**Loading Logic**:
```go
func LoadConfig(path string) (*Config, error) {
    // Read config file
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    // Expand environment variables
    expanded := os.ExpandEnv(string(data))

    // Parse YAML
    var config Config
    if err := yaml.Unmarshal([]byte(expanded), &config); err != nil {
        return nil, err
    }

    // Validate configuration
    if err := config.Validate(); err != nil {
        return nil, err
    }

    return &config, nil
}

func (c *Config) Validate() error {
    if len(c.Networks) == 0 {
        return errors.New("at least one network must be configured")
    }

    for name, network := range c.Networks {
        if network.BaseURL == "" {
            return fmt.Errorf("network %s: base_url is required", name)
        }
        if network.PayeeAddress == "" {
            return fmt.Errorf("network %s: payee_address is required", name)
        }
    }

    return nil
}
```

**Environment Variables** (`.env` file):
```bash
# Circular Protocol Enterprise API (CEP) - Testnet/Development
CIRCULAR_CEP_TESTNET_PRIVATE_KEY=your_testnet_private_key_hex
CIRCULAR_CEP_TESTNET_SEED_PHRASE="your twelve word seed phrase here"
CIRCULAR_CEP_TESTNET_BLOCKCHAIN_ID=0x8a20baa40c45dc5055aeb26197c203e576ef389d9acb171bd62da11dc5ad72b2
CIRCULAR_CEP_NAG_DISCOVERY_URL=https://circularlabs.io/network/getNAG

# Circular Protocol Enterprise API - Mainnet
CIRCULAR_CEP_MAINNET_PRIVATE_KEY=your_mainnet_private_key_hex
CIRCULAR_CEP_MAINNET_BLOCKCHAIN_ID=mainnet  # Discovered via NAG

# Server Configuration
CIRCULAR_CEP_NETWORK=testnet  # Options: testnet, mainnet
LOG_LEVEL=info
```

---

### 6. ErrorResponse

**Purpose**: Structured error information returned to MCP agent when operations fail.

**Lifecycle**: Created on error, returned as tool response

**Structure** (from spec.md FR-013):
```go
type ErrorResponse struct {
    // Error Classification
    Type            string `json:"type"`             // Error category (see below)

    // HTTP Context
    StatusCode      int    `json:"status_code"`      // HTTP status code (if API error)

    // Error Details
    Message         string `json:"message"`          // Human-readable description
    RetrySuggestion string `json:"retry_suggestion,omitempty"` // Guidance for agent

    // Context (optional)
    TransactionID   string `json:"transaction_id,omitempty"`
    Network         string `json:"network,omitempty"`
}
```

**Error Types** (from research.md):

| Type | Status Code | Retry? | Example Message |
|------|-------------|--------|-----------------|
| `API_UNAVAILABLE` | 503 | Yes | "Circular Protocol API is temporarily unavailable" |
| `INVALID_SIGNATURE` | 400 | No | "Transaction signature verification failed" |
| `INVALID_INPUT` | 400 | No | "Payload exceeds 1 MB size limit" |
| `INSUFFICIENT_BALANCE` | 402 | No | "Wallet has insufficient CIRX balance for transaction fee" |
| `TRANSACTION_TIMEOUT` | 504 | Yes | "Transaction did not reach Executed status within 60 seconds" |
| `TRANSACTION_FAILED` | 400 | No | "Blockchain rejected transaction" |
| `RATE_LIMIT` | 429 | Yes | "API rate limit exceeded" |
| `NONCE_MISMATCH` | 400 | Yes | "Transaction nonce does not match wallet nonce" |
| `INVALID_ADDRESS` | 400 | No | "Wallet address format is invalid" |
| `NETWORK_ERROR` | 503 | Yes | "Network connection failed" |

**Error Constructor**:
```go
func NewErrorResponse(errType string, statusCode int, message string) ErrorResponse {
    err := ErrorResponse{
        Type:       errType,
        StatusCode: statusCode,
        Message:    message,
    }

    // Add retry suggestion for retry-able errors
    switch errType {
    case "API_UNAVAILABLE", "NETWORK_ERROR":
        err.RetrySuggestion = "Retry after 30 seconds"
    case "RATE_LIMIT":
        err.RetrySuggestion = "Retry after 60 seconds"
    case "TRANSACTION_TIMEOUT":
        err.RetrySuggestion = "Transaction may still be pending, check status later"
    case "NONCE_MISMATCH":
        err.RetrySuggestion = "Fetch fresh nonce and retry transaction"
    }

    return err
}
```

**Usage Example** (from tool error handling):
```go
if len(payloadBytes) > 1048576 {
    return NewErrorResponse(
        "INVALID_INPUT",
        400,
        "Payload exceeds 1 MB size limit",
    )
}

if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == 503 {
    return NewErrorResponse(
        "API_UNAVAILABLE",
        503,
        "Circular Protocol API is temporarily unavailable",
    )
}
```

---

## Helper Types

### 7. Wallet

**Purpose**: Represents a Circular Protocol wallet account with balance and nonce.

**Structure**:
```go
type Wallet struct {
    Address  string `json:"address"`
    Balance  int64  `json:"balance"`  // CIRX balance (in smallest unit)
    Nonce    int64  `json:"nonce"`
    Network  string `json:"network"`
}
```

**Usage**: Returned by `updateAccount()` API call, used internally for nonce management.

---

### 8. APIResponse

**Purpose**: Generic wrapper for Circular Protocol API responses.

**Structure**:
```go
type APIResponse struct {
    Result   int         `json:"Result"`   // Status code (0 = success)
    Response interface{} `json:"Response"` // Actual response data
    Error    string      `json:"Error,omitempty"`
}
```

**Usage**: Parse API responses from Circular Protocol Enterprise APIs.

---

## Data Flow Diagrams

### Flow 1: Certify Data (certify_data tool)

```
┌─────────┐
│  Agent  │
└────┬────┘
     │ certify_data(data, network)
     ▼
┌─────────────────┐
│ get_wallet_nonce│ ──────────► Circular API: updateAccount()
└────┬────────────┘                 │
     │ Nonce: 42                   │
     │◄────────────────────────────┘
     ▼
┌──────────────────┐
│ Build Transaction│
│ - Payload (hex)  │
│ - Timestamp      │
│ - Nonce: 42      │
└────┬─────────────┘
     │
     ▼
┌────────────┐
│ Sign with  │
│ Private Key│ (Secp256k1)
└────┬───────┘
     │ Signature: 0xabc...
     ▼
┌────────────────────┐
│ Submit Certificate │ ──────────► Circular API: submitCertificate()
└────┬───────────────┘                 │
     │ TransactionID: "tx123"         │
     │◄────────────────────────────────┘
     ▼
┌─────────┐
│  Agent  │ ← Returns transaction_id
└─────────┘
```

### Flow 2: Verify Status (get_transaction_status tool)

```
┌─────────┐
│  Agent  │
└────┬────┘
     │ get_transaction_status(tx_id, network)
     ▼
┌──────────────┐       Attempt 1 (t=0s)
│ Poll Status  │ ──────────► Circular API: getTransactionOutcome()
└──────┬───────┘                 │ Status: Pending
       │◄────────────────────────┘
       │ Wait 5 seconds
       │
       │       Attempt 2 (t=5s)
       │ ──────────► Circular API: getTransactionOutcome()
       │◄────────────┘ Status: Verified
       │ Wait 5 seconds
       │
       │       Attempt 3 (t=10s)
       │ ──────────► Circular API: getTransactionOutcome()
       │◄────────────┘ Status: Executed, BlockID: "blk456"
       ▼
┌─────────┐
│  Agent  │ ← Returns status=Executed, block_id
└─────────┘
```

### Flow 3: Generate Proof (get_certification_proof tool)

```
┌─────────┐
│  Agent  │
└────┬────┘
     │ get_certification_proof(tx_id, network)
     ▼
┌─────────────────┐
│ Get Transaction │ ──────────► Circular API: getTransaction(blockID, txID)
└────┬────────────┘                 │
     │ Transaction: {...}           │
     │◄─────────────────────────────┘
     ▼
┌─────────────────┐
│ Verify Executed │
└────┬────────────┘
     │ Status: Executed ✓
     ▼
┌────────────────────┐
│ Generate Explorer  │
│ URL                │
└────┬───────────────┘
     │ https://circularlabs.io/Explorer?network=testnet&tx=tx123
     ▼
┌───────────────────┐
│ Assemble Proof    │
│ - tx_id           │
│ - block_id        │
│ - timestamp       │
│ - explorer_url    │
└────┬──────────────┘
     ▼
┌─────────┐
│  Agent  │ ← Returns CertificationProof JSON
└─────────┘
```

---

## Entity Relationships

```
┌──────────────────────────┐
│   NetworkConfig          │
│  (loaded at startup)     │
└──────────┬───────────────┘
           │
           │ network parameter
           ▼
┌──────────────────────────┐
│   Wallet / Nonce         │ (get_wallet_nonce)
│  - address               │
│  - nonce                 │
└──────────┬───────────────┘
           │
           │ nonce value
           ▼
┌──────────────────────────┐
│  CertificationTransaction│ (certify_data)
│  - from, payload         │
│  - nonce, signature      │
│  → transaction_id        │
└──────────┬───────────────┘
           │
           │ transaction_id
           ▼
┌──────────────────────────┐
│   TransactionStatus      │ (get_transaction_status)
│  - status: Executed      │
│  → block_id              │
└──────────┬───────────────┘
           │
           │ transaction_id + block_id
           ▼
┌──────────────────────────┐
│   CertificationProof     │ (get_certification_proof)
│  - tx_id, block_id       │
│  - timestamp             │
│  - explorer_url          │
└──────────────────────────┘

     All operations flow through:
┌──────────────────────────┐
│   ErrorResponse          │ (on failure)
│  - type, message         │
│  - retry_suggestion      │
└──────────────────────────┘
```

---

## Persistence Strategy

**Decision**: **No persistent storage** in MCP server.

**Rationale**:
- MCP servers are stateless by design
- All data is transient within tool execution
- Blockchain is the source of truth (immutable ledger)
- Agent is responsible for storing transaction IDs if needed

**Implications**:
- No database required (aligns with x402 pattern)
- No cache required (unlike x402's settlement cache)
- All state retrieved fresh from Circular Protocol API
- Simplified testing and deployment

---

## Security Considerations

### Private Key Handling

**Requirement** (from spec.md FR-003a): Load private key from environment variable at server startup.

**Implementation**:
```go
// Load once at startup
privateKey, err := LoadPrivateKey(os.Getenv("CIRCULAR_PRIVATE_KEY"))
if err != nil {
    log.Fatal("Failed to load private key:", err)
}

// Use for all transaction signing
signature, err := signer.Sign(txData, privateKey)
```

**Security Rules**:
- ✅ Private key loaded from `CIRCULAR_PRIVATE_KEY` environment variable
- ✅ Private key NEVER logged (not even truncated)
- ✅ Signatures truncated to first 8 chars in logs for debugging
- ✅ Private key stored in memory only (never persisted)

### Sensitive Data Logging (from spec.md FR-016)

**Allowed in Logs**:
- ✅ Tool invocations (tool name, network parameter)
- ✅ Transaction IDs (public, on blockchain)
- ✅ Wallet addresses (public, on blockchain)
- ✅ Status transitions (Pending → Verified → Executed)
- ✅ Error types and messages
- ✅ Truncated signatures (first 8 chars for debugging)

**NEVER Logged**:
- ❌ Full signatures
- ❌ Private keys
- ❌ Raw transaction payloads (may contain sensitive data)

---

## Data Model Completion Checklist

- [x] CertificationTransaction entity defined
- [x] TransactionStatus entity defined
- [x] CertificationProof entity defined
- [x] Nonce entity defined
- [x] NetworkConfig entity defined
- [x] ErrorResponse entity defined
- [x] Helper types (Wallet, APIResponse) defined
- [x] Data flow diagrams created
- [x] Entity relationships documented
- [x] Security considerations documented
- [x] Validation rules specified
- [x] Persistence strategy defined (stateless)

**Status**: **PHASE 1a COMPLETE** - Proceed to Phase 1b (contracts/)

---

**Data Model Completed**: 2025-10-30
**Designer**: Claude (Sonnet 4.5)
**Approved for Phase 1b**: ✅ Ready
