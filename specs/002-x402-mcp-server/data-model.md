# Data Model: x402 Payment MCP Server

**Purpose**: Define data structures, configuration schema, and validation rules

## Configuration Schema

### NetworkConfig (YAML)

```go
type Config struct {
    Networks map[string]NetworkConfig `yaml:"networks"`
    EIP712   EIP712Config             `yaml:"eip712"`
    Logging  LoggingConfig            `yaml:"logging"`
    Cache    CacheConfig              `yaml:"cache"`
}

type NetworkConfig struct {
    ChainID        uint64 `yaml:"chain_id"`        // EIP-155 chain ID
    USDCContract   string `yaml:"usdc_contract"`   // Native USDC address
    FacilitatorURL string `yaml:"facilitator_url"` // x402 facilitator endpoint
    RPCURL         string `yaml:"rpc_url"`         // Blockchain RPC for nonces
    PayeeAddress   string `yaml:"payee_address"`   // Certification service payee
}

type EIP712Config struct {
    DomainName    string `yaml:"domain_name"`    // "USD Coin"
    DomainVersion string `yaml:"domain_version"` // "2"
}

type LoggingConfig struct {
    Level  string `yaml:"level"`  // DEBUG, INFO, WARN, ERROR
    Format string `yaml:"format"` // json
}

type CacheConfig struct {
    SettlementTTLMinutes int `yaml:"settlement_ttl_minutes"` // 10
}
```

**Validation Rules**:
- Networks map MUST contain at least one entry
- ChainID MUST match allowed values: 8453, 84532, 42161
- USDCContract MUST be valid Ethereum address (0x prefix, 40 hex chars)
- PayeeAddress MUST be valid Ethereum address
- RPCURL MUST be valid HTTP/HTTPS URL
- SettlementTTLMinutes MUST be > 0

## MCP Tool Data Structures

### 1. create_payment_requirement

**Input**:
```go
type CreatePaymentRequirementInput struct {
    Amount  string `json:"amount"`  // USDC atomic units (e.g., "50000" = 0.05 USDC)
    Network string `json:"network"` // "base" | "base-sepolia" | "arbitrum"
}
```

**Output**:
```go
type PaymentRequirement struct {
    X402Version       int    `json:"x402_version"`         // Always 1
    Scheme            string `json:"scheme"`                // Always "exact"
    Network           string `json:"network"`               // Input network
    MaxAmountRequired string `json:"maxAmountRequired"`     // Input amount
    Payee             string `json:"payee"`                 // From config
    ValidUntil        string `json:"valid_until"`           // ISO8601, +5 min
    Nonce             string `json:"nonce"`                 // Blockchain nonce (hex)
    Asset             string `json:"asset"`                 // USDC contract address
}
```

**Validation**:
- Amount MUST be numeric string, > 0
- Network MUST be in config.Networks map
- Nonce MUST be fetched from blockchain RPC (FR-005 clarification)

### 2. verify_payment

**Input**:
```go
type VerifyPaymentInput struct {
    Authorization EIP3009Authorization `json:"authorization"`
    Network       string               `json:"network"`
}

type EIP3009Authorization struct {
    From        string `json:"from"`         // Payer address
    To          string `json:"to"`           // Payee address
    Value       string `json:"value"`        // Amount (atomic units)
    ValidAfter  uint64 `json:"validAfter"`   // Unix timestamp
    ValidBefore uint64 `json:"validBefore"`  // Unix timestamp
    Nonce       string `json:"nonce"`        // bytes32 hex
    V           uint8  `json:"v"`            // 27 or 28
    R           string `json:"r"`            // bytes32 hex
    S           string `json:"s"`            // bytes32 hex
}
```

**Output**:
```go
type VerifyPaymentOutput struct {
    IsValid       bool   `json:"is_valid"`
    SignerAddress string `json:"signer_address"` // Recovered from signature
    Error         string `json:"error,omitempty"`
}
```

**Validation**:
- All address fields MUST be valid Ethereum addresses
- Value MUST be numeric string, > 0
- ValidAfter MUST be <= current time
- ValidBefore MUST be > current time
- V MUST be 27 or 28
- R, S MUST be 32-byte hex strings (66 chars with 0x)
- Signature recovery MUST yield signer == From

### 3. settle_payment

**Input**:
```go
type SettlePaymentInput struct {
    Authorization        EIP3009Authorization `json:"authorization"`
    PaymentRequirement   PaymentRequirement   `json:"payment_requirement"`
}
```

**Output**:
```go
type SettlePaymentOutput struct {
    Status      string `json:"status"`       // "settled" | "pending" | "failed"
    TxHash      string `json:"tx_hash,omitempty"`
    BlockNumber uint64 `json:"block_number,omitempty"`
    Error       string `json:"error,omitempty"`
    RetryAfter  int    `json:"retry_after,omitempty"` // Seconds
}
```

**Validation**:
- Authorization MUST pass verify_payment checks
- Idempotency: Check cache by nonce before settling
- HTTP timeout: 5 seconds (FR-011)

### 4. generate_browser_link

**Input**:
```go
type GenerateBrowserLinkInput struct {
    PaymentRequirement PaymentRequirement `json:"payment_requirement"`
    CallbackURL        string             `json:"callback_url"`
}
```

**Output**:
```go
type BrowserPaymentLink struct {
    URL        string `json:"url"`         // MetaMask deep link
    ExpiresAt  string `json:"expires_at"`  // ISO8601 (same as payment requirement)
}
```

**Validation**:
- CallbackURL MUST be valid HTTPS URL (production) or HTTP (testnet)
- URL MUST be properly encoded

**Format**:
```
https://link.metamask.io/send/{usdc_contract}@{chain_id}/transfer?address={payee}&uint256={amount}
```

### 5. encode_payment_for_qr

**Input**:
```go
type EncodePaymentForQRInput struct {
    PaymentRequirement PaymentRequirement `json:"payment_requirement"`
    CallbackURL        string             `json:"callback_url,omitempty"`
}
```

**Output**:
```go
type QRPaymentEncoding struct {
    EIP681URI            string `json:"eip681_uri"`
    EstimatedQRVersion   int    `json:"estimated_qr_version"`  // ~10 for typical payments
}
```

**Validation**:
- CallbackURL optional (not part of EIP-681 standard)
- URI length SHOULD be < 300 characters for QR Version 10

**Format**:
```
ethereum:{usdc_contract}@{chain_id}/transfer?address={payee}&uint256={amount}
```

## Internal Entities

### SettlementCacheEntry

```go
type SettlementCacheEntry struct {
    Nonce       string    `json:"nonce"`
    Result      SettlePaymentOutput `json:"result"`
    CachedAt    time.Time `json:"cached_at"`
    ExpiresAt   time.Time `json:"expires_at"`
}
```

**Cache Key**: `settlement:{nonce}`
**TTL**: 10 minutes (from config)
**Purpose**: Idempotency (FR-013 clarification)

### EIP712Domain

```go
type EIP712Domain struct {
    Name              string         `json:"name"`              // "USD Coin"
    Version           string         `json:"version"`           // "2"
    ChainId           *big.Int       `json:"chainId"`
    VerifyingContract common.Address `json:"verifyingContract"` // USDC contract
}
```

### ReceiveWithAuthorizationMessage

```go
type ReceiveWithAuthorizationMessage struct {
    From        common.Address `json:"from"`
    To          common.Address `json:"to"`
    Value       *big.Int       `json:"value"`
    ValidAfter  *big.Int       `json:"validAfter"`
    ValidBefore *big.Int       `json:"validBefore"`
    Nonce       [32]byte       `json:"nonce"`
}
```

## State Transitions

### Payment Verification Flow

```
[Authorization Received]
    ↓
[Parse & Validate Fields] → Invalid → {is_valid: false, error: "..."}
    ↓ Valid
[Construct EIP-712 Hash]
    ↓
[Recover Signer via ECDSA]
    ↓
[Check Signer == From] → Mismatch → {is_valid: false, error: "signature mismatch"}
    ↓ Match
[Check Time Bounds] → Expired/Future → {is_valid: false, error: "..."}
    ↓ Valid
{is_valid: true, signer_address: "0x..."}
```

### Payment Settlement Flow

```
[Authorization + PaymentRequirement]
    ↓
[Check Cache by Nonce] → Hit → Return cached result (10ms)
    ↓ Miss
[Verify Payment] → Invalid → {status: "failed", error: "..."}
    ↓ Valid
[POST to Facilitator (5s timeout)]
    ↓
    ├─ Success → {status: "settled", tx_hash: "...", block_number: ...}
    ├─ Timeout → {status: "pending", retry_after: 30}
    └─ Error → {status: "failed", error: "facilitator error: ..."}
    ↓
[Cache Result with TTL]
    ↓
Return result
```

## Error Types

```go
// Standard MCP error format
type MCPError struct {
    Code    int    `json:"code"`    // MCP error code
    Message string `json:"message"` // Human-readable
    Data    any    `json:"data,omitempty"`
}

// Error codes
const (
    InvalidParams      = -32602 // Invalid tool parameters
    InternalError      = -32603 // Server error
    NetworkUnavailable = -32001 // Custom: RPC/facilitator unreachable
    ValidationFailed   = -32002 // Custom: Signature/authorization invalid
)
```

## Validation Summary

| Field Type | Rules |
|------------|-------|
| **Ethereum Address** | 0x prefix, 40 hex chars, case-insensitive |
| **Amount (USDC)** | Numeric string, > 0, <= uint256 max |
| **Chain ID** | Allowlist: 8453, 84532, 42161 |
| **Network Name** | Allowlist: "base", "base-sepolia", "arbitrum" |
| **Timestamp (Unix)** | uint64, reasonable range (2020-2100) |
| **Nonce (bytes32)** | 0x prefix, 64 hex chars |
| **Signature (v,r,s)** | v in [27,28], r/s 32-byte hex |
| **URL** | Valid HTTP/HTTPS, length < 2048 |
| **ISO8601** | RFC 3339 format with timezone |

## Persistence

**None** - MCP server is stateless. Only in-memory cache for idempotency (TTL-based expiration).

HTTP proxy is responsible for persistent storage (PostgreSQL database for certification requests, payments, etc.).
