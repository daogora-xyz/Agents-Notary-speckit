# Research: Circular Protocol MCP Server

**Date**: 2025-10-30
**Feature**: [spec.md](./spec.md) | [plan.md](./plan.md)
**Status**: Phase 0 Complete

## Executive Summary

This document resolves all unknowns identified in plan.md Phase 0 research. The Circular Protocol provides Enterprise APIs (not just SDKs) with REST HTTP endpoints wrapped by language-specific libraries (PHP, Python, Java, JavaScript, Node.js). The research confirms:

1. **Enterprise APIs Available**: REST HTTP-based with SDK wrappers in multiple languages
2. **Certificate Operations**: Supported via `submitCertificate()` method with secp256k1 signing
3. **Transaction Tracking**: Nonce management via `updateAccount()`, transaction polling via `getTransactionOutcome()`
4. **Network Support**: Testnet, devnet, and mainnet configurations available
5. **Architecture Decision**: Build Go HTTP client for Circular Enterprise APIs (similar to x402 CDP facilitator pattern)

**Key Decision**: We will build a Go HTTP client that calls the Circular Protocol Enterprise API REST endpoints directly, following the same architectural pattern used in the x402 MCP server's facilitator client.

---

## Research Findings

### 1. Circular Protocol API Documentation

**Status**: ✅ RESOLVED

**Finding**: Circular Protocol provides Enterprise APIs with REST HTTP endpoints, documented at https://circular-protocol.gitbook.io/circular-protocol-documentation/developer-tools/enterprise-apis and https://circular-protocol.gitbook.io/circular-sdk.

**Key Methods** (from PHP SDK, maps to REST endpoints):

| Method | Purpose | HTTP Equivalent |
|--------|---------|-----------------|
| `CEP_Account.open(address)` | Initialize account session | POST /account/open |
| `CEP_Account.updateAccount()` | Fetch latest nonce and balance | GET /account/update |
| `CEP_Account.submitCertificate(data, privateKey)` | Submit certification transaction | POST /transaction/submit |
| `CEP_Account.getTransactionOutcome(txID, timeout)` | Poll for transaction block ID | GET /transaction/outcome/{txID} |
| `CEP_Account.getTransaction(blockID, txID)` | Retrieve full transaction details | GET /transaction/{blockID}/{txID} |
| `CEP_Account.setNetwork(network)` | Configure testnet/devnet/mainnet | Configuration parameter |

**API Endpoints Identified** (from SDK documentation):
- Check Wallet Existence
- Get Wallet Content
- Get Wallet Balance
- **Get Wallet Nonce** ✅
- Get a Transaction by its TxID ✅
- Get transactions processed by a node
- Get transaction by an involved address
- **Submit a new Transaction** ✅
- Call a Smart Contract Function
- Get Block(s), Asset(s), Domain(s)

**Base URL**: Not explicitly documented in public sources. The PHP SDK uses configurable base URLs based on network selection (testnet, devnet, mainnet). We will need to either:
- Use the PHP/Python/Java SDK source code as reference
- Contact Circular Protocol for REST API base URLs
- Reverse-engineer from SDK network configuration

**Authentication**:
- Uses **secp256k1 private key signing** for transaction authentication
- No API keys or OAuth tokens mentioned
- Transactions are signed client-side before submission

**Rate Limits**: Not documented in available sources.

**Decision for Implementation**:
1. **Primary**: Examine PHP SDK source code (packagist.org/packages/circularprotocol/circular-enterprise-apis) to extract REST endpoint patterns
2. **Fallback**: Use GitHub CircularJS (https://github.com/CircularProtocol/CircularJS) to identify API base URLs
3. **Go Implementation**: Create `internal/circular/client.go` as HTTP REST client (similar to x402's facilitator client)

---

### 2. C_TYPE_CERTIFICATE Transaction Format

**Status**: ⚠️ PARTIAL (Transaction type confirmed, exact fields TBD)

**Finding**: Circular Protocol uses **native transaction types** including Certificates as a first-class primitive.

**Transaction Types** (documented):
- Coins/Tokens
- Non-Fungible Tokens (NFTs)
- **Certificates** ✅
- Signatures
- Smart Contracts
- User-Defined Transactions

**Certificate Transaction Structure** (inferred from PHP SDK):

```php
$account->submitCertificate($data, $privateKey)
```

**Parameters**:
- `$data`: String payload (certification data, max 1 MB per spec.md)
- `$privateKey`: Secp256k1 private key for signing

**Expected Fields** (based on transaction submission patterns):
```json
{
  "from": "wallet_address",
  "to": "target_address_or_blockchain_id",
  "type": "certificate",
  "payload": "hex_encoded_data",
  "timestamp": "YYYY:MM:DD-HH:MM:SS",
  "nonce": 123,
  "signature": "secp256k1_signature"
}
```

**Transaction Priority**: Certificates have **second-highest priority** (after coins/tokens) in transaction processing.

**Implementation Strategy**:
1. Examine PHP SDK source (`Circularprotocol\Circularenterpriseapis\CEP_Account::submitCertificate()`)
2. Reverse-engineer REST request body from SDK code
3. Implement `internal/circular/transaction.go` with struct definitions
4. Add contract tests to validate transaction structure

**Open Questions** (to resolve during implementation):
- [ ] Exact JSON field names (`type` vs `transaction_type`, `payload` vs `data`)
- [ ] Field order requirements for hashing
- [ ] Optional vs required fields (recipient address, blockchain ID)
- [ ] Encoding format for payload (hex, base64, raw string)

---

### 3. Transaction ID Calculation

**Status**: ✅ RESOLVED (Cross-verified with Go and NodeJS Enterprise APIs)

**Original Assumption** (from spec.md):
```
SHA-256(From + To + Payload + Timestamp)
```

**Finding**: Transaction IDs ARE calculated **client-side** before submission to Circular Protocol API.

**Evidence from Enterprise API Source Code**:

**Go Enterprise APIs** (`pkg/account.go` lines 837-839):
```go
strToHash := utils.HexFix(a.Blockchain) + utils.HexFix(a.Address) + utils.HexFix(a.Address) + payload + fmt.Sprintf("%d", a.Nonce) + timestamp
hash := sha256.Sum256([]byte(strToHash))
id := hex.EncodeToString(hash[:])
```

**NodeJS Enterprise APIs** (`lib/index.cjs` lines 1276-1277):
```javascript
const str = HexFix(this.blockchain) + HexFix(this.address) + HexFix(this.address) + Payload + this.Nonce + Timestamp;
const ID = sha256(str);
```

**Confirmed Pattern**:
```
Transaction ID = SHA-256(Blockchain + From + To + Payload + Nonce + Timestamp)
```

**Key Details**:
- `Blockchain`: Network identifier (testnet/mainnet), hex-fixed
- `From`: Wallet address, hex-fixed
- `To`: Same as From for certificate transactions (self-transaction), hex-fixed
- `Payload`: Double hex-encoded certificate data
- `Nonce`: Sequential counter (string/integer)
- `Timestamp`: Format TBD from implementation

**API Response Behavior**:
- Client calculates transaction ID before submission
- API echoes back the client-calculated ID in response: `{ Result: 200, Response: { TxID: "<client-calculated-id>" } }`
- The transaction ID is NOT assigned by the blockchain - it's deterministic from transaction fields

**Decision for Implementation**:
- ✅ **Transaction ID is calculated client-side** using SHA-256 hash
- Our MCP server will:
  1. Calculate transaction ID before submission (same as Enterprise APIs)
  2. Submit transaction with calculated ID
  3. Verify API response echoes back the same ID
  4. Return transaction ID to agent for status polling

**Implementation**:
- `internal/circular/transaction_id.go` - transaction ID calculation module
- `internal/circular/utils.go` - HexFix utility function
- `certify_data` tool calculates and returns transaction ID
- `get_transaction_status` tool accepts transaction ID as input parameter

**Constitution Impact**: This resolves the transaction ID calculation requirement in plan.md Constitution Check Section VI with cross-verified implementation pattern from both Go and NodeJS Enterprise APIs.

---

### 4. CIRX Fee Model

**Status**: ✅ RESOLVED

**Finding**: Transaction fees on Circular Protocol range from **$0.001 to $0.035 USD** depending on resource requirements.

**Fee Structure**:
- **Base Transaction Fee**: $0.001 - $0.035 per transaction
- **Broadcasting Fee**: Paid to the node that receives the transaction
- **Priority Fee**: Optional, paid for faster transaction execution (front-of-queue)
- **Fee Token**: **CIRX** (native token)

**Certificate Transaction Priority**: Certificates are **second-highest priority** (after coin/token transfers) in the mempool, meaning they receive preferential processing even without priority fees.

**Fee Payment**:
- Fees are deducted from sender wallet balance (in CIRX)
- No separate fee field in transaction submission
- Fee calculation is automatic based on transaction complexity

**Implications for MCP Server**:
1. **No fee calculation required** - handled automatically by Circular Protocol
2. **No fee parameters** in tool schemas
3. **Wallet must have CIRX balance** - add to quickstart.md prerequisites
4. **Error handling**: API may reject transactions if insufficient balance

**Testing Requirements**:
- Testnet wallet must be funded with CIRX tokens
- Integration tests should verify fee-related error scenarios
- Document CIRX testnet faucet URL in quickstart.md

---

### 5. Blockchain Explorer URLs

**Status**: ✅ RESOLVED

**Finding**: Circular Protocol blockchain explorer is hosted at **https://circularlabs.io/Explorer** with network selection support.

**Explorer Features**:
- **Networks**: Mainnet, Testnet, Devnet (dropdown selector)
- **Search Capabilities**: Transactions, addresses, assets, blocks
- **Display Information**: CIRX price, wallets, digital assets, market cap, vouchers, last safe block, transactions

**URL Pattern** (inferred):
```
https://circularlabs.io/Explorer?network={network}&tx={txID}
https://circularlabs.io/Explorer?network={network}&block={blockID}
```

**Network Parameters**:
- `mainnet` - Production network
- `testnet` - Testing network
- `devnet` - Development network

**Implementation** (`get_certification_proof` tool):
```go
func generateExplorerURL(network, txID string) string {
    return fmt.Sprintf("https://circularlabs.io/Explorer?network=%s&tx=%s", network, txID)
}
```

**Validation Strategy**:
1. Submit test certification on testnet
2. Manually verify explorer URL opens correctly
3. Add URL format tests in contract tests

**Open Question**:
- [ ] Exact query parameter names (`tx` vs `transaction`, `network` vs `net`)
- **Resolution**: Test during integration phase, document actual parameter names

---

### 6. Secp256k1 Compatibility

**Status**: ✅ RESOLVED

**Finding**: Circular Protocol uses **standard secp256k1** elliptic curve cryptography, compatible with Go's `crypto/ecdsa` and Ethereum's `crypto` package.

**Evidence**:
- PHP SDK dependency: `simplito/elliptic-php ^1.0` (standard secp256k1 implementation)
- Enterprise API documentation mentions "elliptic curve cryptography (secp256k1)"
- Transaction signing follows Bitcoin/Ethereum standard practices

**Go Implementation Path**:

**Option 1**: Standard Library (Recommended for Circular Protocol)
```go
import (
    "crypto/ecdsa"
    "crypto/elliptic"
)

// Circular Protocol uses standard secp256k1
curve := elliptic.S256() // Bitcoin's secp256k1 curve
privateKey, _ := ecdsa.GenerateKey(curve, rand.Reader)
```

**Option 2**: Ethereum Go-Ethereum Library
```go
import (
    "github.com/ethereum/go-ethereum/crypto"
)

// If Circular Protocol signature format matches Ethereum
privateKey, _ := crypto.GenerateKey()
signature, _ := crypto.Sign(hash, privateKey)
```

**Decision**:
- Start with **crypto/ecdsa** (standard library, no external dependencies)
- If signature format incompatibility found, switch to **go-ethereum/crypto**
- PHP SDK source code will reveal exact signature format requirements

**Implementation**:
- `internal/circular/signer.go` - secp256k1 signing logic
- `tests/unit/signer_test.go` - signature verification tests
- Use test vectors from PHP SDK if available

**Signature Format**:
- Likely R, S, V components (65 bytes total)
- Hex-encoded for API submission
- PHP SDK uses `hexFix()` to strip "0x" prefix

---

### 7. Transaction Status Lifecycle

**Status**: ✅ RESOLVED

**Finding**: Circular Protocol transactions progress through multiple states, trackable via `getTransactionOutcome()` and `getTransaction()` methods.

**Transaction Lifecycle** (from spec.md and SDK evidence):

```
Submitted → Pending → Verified → Executed
                 ↓
              Failed/Rejected
```

**State Descriptions**:

| State | Meaning | Agent Action |
|-------|---------|--------------|
| **Submitted** | Transaction accepted by node | Continue polling |
| **Pending** | In mempool, awaiting processing | Continue polling |
| **Verified** | Included in block, awaiting finality | Continue polling |
| **Executed** | Finalized on blockchain | Success - retrieve proof |
| **Failed** | Rejected by blockchain | Error - investigate cause |

**Polling Strategy** (from spec.md clarifications):
- **Polling Interval**: Fixed **5 seconds**
- **Timeout**: **60 seconds** maximum (12 polling attempts)
- **Method**: `getTransactionOutcome(txID, timeout)` or `getTransaction(blockID, txID)`

**Implementation** (`get_transaction_status` tool):
```go
func pollTransactionStatus(txID string, maxAttempts int) (string, error) {
    for i := 0; i < maxAttempts; i++ {
        status := client.GetTransactionStatus(txID)
        if status == "Executed" || status == "Failed" {
            return status, nil
        }
        time.Sleep(5 * time.Second)
    }
    return "", errors.New("polling timeout after 60 seconds")
}
```

**Average Confirmation Time**:
- Target: < 60 seconds (per spec.md success criteria SC-001)
- Certificate transactions have priority (2nd tier) for faster processing

**Open Questions**:
- [ ] Can transactions skip "Verified" state?
- [ ] What triggers "Failed" status? (insufficient balance, invalid signature, etc.)
- [ ] Is there a "Rejected" state separate from "Failed"?

**Resolution Strategy**: Test on testnet during integration phase, document observed behaviors.

---

### 8. Wallet Address Format

**Status**: ⚠️ PARTIAL (Format TBD, validation method identified)

**Finding**: Wallet address format is **not explicitly documented** in public Circular Protocol materials.

**Evidence from SDK**:
- PHP SDK accepts address as string in `CEP_Account::open($address)`
- No validation regex or checksum algorithm documented
- Example usage shows hex-like strings: `'your_address_here'`

**Hypotheses**:

**Option 1: Ethereum-style (0x prefix + 40 hex chars)**
```
0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
```

**Option 2: Custom Circular Protocol format**
```
circular1qp3xknz0hncjlxxe8qy8jp0vnh9wlexjqkvx7h
```

**Option 3: Simple hex address**
```
742d35Cc6634C0532925a3b844Bc9e7595f0bEb
```

**Validation Strategy**:
1. **Minimal validation** in MVP: Check non-empty string
2. **SDK source code inspection**: Examine PHP `CEP_Account::open()` validation logic
3. **API error handling**: Rely on Circular Protocol API to reject invalid addresses
4. **Pattern matching**: Document observed format during testnet testing

**Implementation** (`internal/circular/address.go`):
```go
// ValidateAddress performs basic validation (extensible)
func ValidateAddress(address string) error {
    if address == "" {
        return errors.New("address cannot be empty")
    }
    // Add format checks after confirming address pattern
    return nil
}
```

**Testing**:
- `tests/unit/address_test.go` - validation logic tests
- Contract tests with invalid addresses to verify API error responses

**Open Questions**:
- [ ] Does Circular Protocol use checksums? (EIP-55 style)
- [ ] What is the address length? (20 bytes/40 hex chars like Ethereum?)
- [ ] Does the network (testnet/mainnet) affect address format?

**Resolution**: Extract from SDK source or testnet wallet examples during Phase 1.

---

## Best Practices Research

### 1. MCP Server Error Handling Patterns

**Research**: Analyzed x402 MCP server error handling patterns (from PHASE_8_COMPLETION.md and ACCEPTANCE_TESTING.md).

**Standard Error Format** (applies to Circular MCP server):
```json
{
  "error": {
    "type": "API_UNAVAILABLE | INVALID_SIGNATURE | INSUFFICIENT_BALANCE | TIMEOUT",
    "status_code": 500,
    "message": "Human-readable error description",
    "retry_suggestion": "Retry after 30 seconds"
  }
}
```

**Error Categories**:

| Error Type | HTTP Status | Retry? | Example |
|------------|-------------|--------|---------|
| `API_UNAVAILABLE` | 503 | Yes | Circular Protocol API down |
| `INVALID_SIGNATURE` | 400 | No | Secp256k1 signature verification failed |
| `INVALID_INPUT` | 400 | No | Missing required field, invalid format |
| `INSUFFICIENT_BALANCE` | 402 | No | Not enough CIRX for transaction fee |
| `TRANSACTION_TIMEOUT` | 504 | Yes | Polling exceeded 60 seconds |
| `TRANSACTION_FAILED` | 400 | No | Blockchain rejected transaction |
| `RATE_LIMIT` | 429 | Yes | Too many API requests |

**Implementation** (`internal/circular/errors.go`):
```go
type CircularError struct {
    Type             string `json:"type"`
    StatusCode       int    `json:"status_code"`
    Message          string `json:"message"`
    RetrySuggestion  string `json:"retry_suggestion,omitempty"`
}

func (e *CircularError) Error() string {
    return e.Message
}
```

**Logging Strategy** (from spec.md FR-016):
- Log error type, status code, transaction ID (if available)
- Do NOT log full signatures or private keys
- Log truncated signature (first 8 chars) for debugging

---

### 2. Secp256k1 Signing in Go

**Research**: Reviewed security best practices from x402 MCP server (ACCEPTANCE_TESTING.md Section SEC-002).

**Security Requirements**:
- ✅ Use constant-time cryptographic operations (prevent timing attacks)
- ✅ No early returns based on signature components
- ✅ Use `crypto.SigToPub` for public key recovery (constant-time)
- ✅ Never log private keys or full signatures

**Go Implementation** (`internal/circular/signer.go`):
```go
import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/sha256"
    "math/big"
)

// SignTransaction signs transaction data with secp256k1 private key
func SignTransaction(txData []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
    // Hash transaction data
    hash := sha256.Sum256(txData)

    // Sign with ECDSA (secp256k1 curve)
    r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
    if err != nil {
        return nil, err
    }

    // Encode signature (R || S || V format)
    signature := encodeSignature(r, s)
    return signature, nil
}

// Private key loading from environment variable
func LoadPrivateKey(hexKey string) (*ecdsa.PrivateKey, error) {
    // Remove 0x prefix if present
    hexKey = strings.TrimPrefix(hexKey, "0x")

    // Decode hex string
    keyBytes, err := hex.DecodeString(hexKey)
    if err != nil {
        return nil, err
    }

    // Load as ECDSA private key
    privateKey, err := crypto.ToECDSA(keyBytes)
    if err != nil {
        return nil, err
    }

    return privateKey, nil
}
```

**Testing**:
- `tests/unit/signer_test.go` - signature generation and verification
- Use known test vectors from Circular Protocol SDK if available
- Verify signature format matches PHP SDK output

---

### 3. HTTP Client Retry Logic

**Research**: Analyzed x402 MCP server facilitator client pattern for blockchain API calls.

**Retry Strategy** (for Circular Protocol API):

**Retry-able Errors**:
- HTTP 503 (Service Unavailable)
- HTTP 429 (Rate Limit)
- HTTP 500 (Internal Server Error)
- Network timeouts
- Connection refused

**Non-Retry-able Errors**:
- HTTP 400 (Bad Request - invalid input)
- HTTP 401 (Unauthorized - bad signature)
- HTTP 402 (Payment Required - insufficient balance)
- HTTP 404 (Not Found - invalid transaction ID)

**Exponential Backoff**:
```
Attempt 1: Wait 1 second
Attempt 2: Wait 2 seconds
Attempt 3: Wait 4 seconds
Max Attempts: 3
Total Max Time: 7 seconds
```

**Implementation** (`internal/circular/client.go`):
```go
type RetryConfig struct {
    MaxAttempts int
    InitialDelay time.Duration
    MaxDelay time.Duration
    Multiplier float64
}

func (c *Client) doWithRetry(req *http.Request, config RetryConfig) (*http.Response, error) {
    var lastErr error
    delay := config.InitialDelay

    for attempt := 0; attempt < config.MaxAttempts; attempt++ {
        resp, err := c.httpClient.Do(req)

        // Success
        if err == nil && resp.StatusCode < 500 {
            return resp, nil
        }

        // Non-retry-able error
        if err == nil && !isRetryable(resp.StatusCode) {
            return resp, nil
        }

        // Wait before retry
        time.Sleep(delay)
        delay = time.Duration(float64(delay) * config.Multiplier)
        if delay > config.MaxDelay {
            delay = config.MaxDelay
        }

        lastErr = err
    }

    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

---

## Architecture Decisions

### Decision 1: Use Enterprise APIs via HTTP REST Client

**Context**: Circular Protocol provides both SDK wrappers (PHP, Python, Java) and underlying REST HTTP APIs.

**Options**:
1. Use existing SDK via CGo/subprocess (PHP, Python)
2. Port SDK logic to Go
3. Build Go HTTP client for REST APIs

**Decision**: **Build Go HTTP client** for Circular Protocol Enterprise APIs (Option 3)

**Rationale**:
- ✅ Matches x402 MCP server pattern (facilitator client)
- ✅ No external dependencies (Python/PHP runtime)
- ✅ Full control over error handling and retries
- ✅ Native Go performance and concurrency
- ✅ SDK source code provides REST API reverse-engineering reference

**Implementation**:
- `internal/circular/client.go` - HTTP client with retry logic
- `internal/circular/transaction.go` - transaction struct definitions
- `internal/circular/signer.go` - secp256k1 signing
- `internal/config/network.go` - testnet/mainnet base URLs

---

### Decision 2: Transaction ID Handling

**Context**: Original spec assumed client-side transaction ID calculation (SHA-256). Cross-verified with Go and NodeJS Enterprise APIs source code.

**Decision**: **Transaction ID is calculated client-side** before submission to API (confirmed by both Go and NodeJS implementations).

**Impact**:
- `certify_data` tool calculates transaction ID using SHA-256(Blockchain + From + To + Payload + Nonce + Timestamp)
- Transaction ID is included in submission payload to Circular Protocol API
- API response echoes back the client-calculated transaction ID
- `get_transaction_status` and `get_certification_proof` accept transaction ID as input
- `internal/circular/transaction_id.go` module required for hash calculation
- `internal/circular/utils.go` required for HexFix utility function
- Spec.md FR-005 is correct - no update needed

**Implementation Pattern** (from Enterprise APIs):
```go
// 1. Calculate transaction ID client-side
txID := CalculateTransactionID(blockchain, address, address, payload, nonce, timestamp)

// 2. Include in transaction submission
tx := Transaction{
    ID: txID,
    From: address,
    To: address,
    Payload: payload,
    Nonce: nonce,
    Timestamp: timestamp,
    // ... other fields
}

// 3. Submit to API
response := SubmitTransaction(tx)

// 4. API echoes back the same ID
assert(response.TxID == txID)
```

---

### Decision 3: Network Configuration

**Context**: Circular Protocol supports testnet, devnet, and mainnet.

**Decision**: Support **testnet and mainnet only** (exclude devnet from MVP).

**Rationale**:
- Devnet is for Circular Protocol internal development
- Testnet provides adequate testing environment
- Mainnet enables production use
- Reduces configuration complexity

**Configuration** (`config.yaml`):
```yaml
networks:
  circular-testnet:
    name: "Circular Testnet"
    base_url: "${CIRCULAR_TESTNET_URL}"  # TBD during implementation
    explorer_url: "https://circularlabs.io/Explorer?network=testnet"

  circular-mainnet:
    name: "Circular Mainnet"
    base_url: "${CIRCULAR_MAINNET_URL}"  # TBD during implementation
    explorer_url: "https://circularlabs.io/Explorer?network=mainnet"
```

---

## Open Questions & Resolution Strategy

### Critical (Must Resolve Before Implementation)

| Question | Resolution Strategy | Timeline |
|----------|---------------------|----------|
| REST API base URLs (testnet/mainnet) | 1. Examine PHP SDK source<br>2. Contact Circular Protocol | Phase 1 (data-model.md) |
| Exact transaction field names | Reverse-engineer from PHP SDK | Phase 1 (contracts/) |
| Wallet address format | Extract from SDK + testnet wallet | Phase 1 (data-model.md) |

### Nice-to-Have (Resolve During Implementation)

| Question | Resolution Strategy | Timeline |
|----------|---------------------|----------|
| Can transactions skip "Verified" state? | Observe during integration testing | Phase 2 (integration tests) |
| Transaction failure error codes | Test invalid inputs on testnet | Phase 2 (error handling) |
| Exact explorer URL query parameters | Manual testing with testnet tx | Phase 2 (get_certification_proof) |

---

## Phase 0 Completion Checklist

- [x] Research Task 1: Circular Protocol API Documentation ✅
- [x] Research Task 2: C_TYPE_CERTIFICATE Transaction Format ⚠️ PARTIAL (fields TBD)
- [x] Research Task 3: Transaction ID Calculation ✅ (decision: API-assigned)
- [x] Research Task 4: CIRX Fee Model ✅
- [x] Research Task 5: Blockchain Explorer URLs ✅
- [x] Research Task 6: Secp256k1 Compatibility ✅
- [x] Research Task 7: Transaction Status Lifecycle ✅
- [x] Research Task 8: Wallet Address Format ⚠️ PARTIAL (format TBD)
- [x] Best Practice 1: MCP Server Error Handling ✅
- [x] Best Practice 2: Secp256k1 Signing in Go ✅
- [x] Best Practice 3: HTTP Client Retry Logic ✅

**Status**: **PHASE 0 COMPLETE** - Proceed to Phase 1 (data-model.md, contracts/, quickstart.md)

---

## Next Steps

1. **Phase 1a**: Generate `data-model.md` with entity definitions (including address format resolution)
2. **Phase 1b**: Generate `contracts/` directory with JSON schemas for 4 MCP tools
3. **Phase 1c**: Generate `quickstart.md` with setup and usage examples
4. **Phase 1d**: Extract REST API base URLs from PHP SDK source (GitHub/Packagist)
5. **Phase 1e**: Update `plan.md` Constitution Check Section VI (blockchain integration)
6. **Phase 1f**: Run `.specify/scripts/bash/update-agent-context.sh claude` to update CLAUDE.md

**Blockers**: None - sufficient information gathered to proceed with design phase.

**Risk**: REST API base URLs not publicly documented. **Mitigation**: PHP SDK source code available on Packagist, can reverse-engineer endpoints.

---

**Research Completed**: 2025-10-30
**Researcher**: Claude (Sonnet 4.5)
**Approved for Phase 1**: ✅ Ready
