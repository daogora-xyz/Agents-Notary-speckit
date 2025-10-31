# MCP Tool Contracts

**Feature**: Circular Protocol MCP Server
**Created**: 2025-10-30
**Status**: Phase 1b Complete

## Overview

This directory contains JSON Schema definitions for all 4 MCP tools provided by the Circular Protocol MCP Server. These contracts define the input parameters, output formats, and error responses for each tool following the Model Context Protocol (MCP) specification.

## Tool Schemas

### 1. get_wallet_nonce.json

**Purpose**: Retrieve current nonce for wallet transaction ordering

**Priority**: P3 (Supporting capability)

**Input**:
- `wallet_address` (string, required): Circular Protocol wallet address
- `network` (enum, required): "testnet" | "mainnet"

**Output**:
- `wallet_address`: Address queried
- `current_nonce`: Next nonce value to use (integer ≥ 0)
- `network`: Network queried
- `last_updated`: RFC3339 timestamp

**Use Case**: Fetch nonce before calling `certify_data` to ensure correct transaction ordering.

---

### 2. certify_data.json

**Purpose**: Submit data for immutable blockchain certification

**Priority**: P1 (Core value proposition)

**Input**:
- `data` (string, required): Data to certify (max 1 MB)
- `network` (enum, required): "testnet" | "mainnet"
- `wallet_address` (string, optional): Sender address

**Output**:
- `transaction_id`: Blockchain-assigned identifier
- `status`: Initial status ("Pending")
- `network`: Network where submitted
- `submitted_at`: RFC3339 timestamp
- `sender`: Wallet address (optional)

**Use Case**: Primary certification operation - records data immutably on Circular Protocol.

---

### 3. get_transaction_status.json

**Purpose**: Poll transaction status until execution

**Priority**: P2 (Essential for workflows)

**Input**:
- `transaction_id` (string, required): Transaction identifier from `certify_data`
- `network` (enum, required): "testnet" | "mainnet"
- `wait_for_execution` (boolean, optional): Poll until "Executed" (default: true)

**Output**:
- `transaction_id`: Transaction queried
- `status`: "Pending" | "Verified" | "Executed" | "Failed"
- `block_id`: Block identifier (when executed)
- `timestamp`: Block timestamp (Circular format)
- `confirmations`: Block confirmations count
- `network`: Network queried
- `executed_at`: RFC3339 timestamp (when executed)

**Use Case**: Monitor certification progress, wait for blockchain confirmation (target: < 60 seconds).

**Polling Strategy**:
- Interval: 5 seconds
- Max attempts: 12 (60 second timeout)
- Terminal states: "Executed", "Failed"

---

### 4. get_certification_proof.json

**Purpose**: Generate verifiable proof for executed transactions

**Priority**: P2 (Verification capability)

**Input**:
- `transaction_id` (string, required): Executed transaction identifier
- `network` (enum, required): "testnet" | "mainnet"
- `include_data` (boolean, optional): Include certified data in proof (default: false)

**Output**:
- `transaction_id`: Transaction identifier
- `block_id`: Block containing transaction
- `timestamp`: Block timestamp (Circular format)
- `block_height`: Block number (optional)
- `explorer_url`: Blockchain explorer link
- `network`: Network (testnet/mainnet)
- `certified_data`: Hex-encoded data (if include_data = true)
- `sender`: Wallet address (optional)

**Use Case**: Generate shareable proof of certification for external verification.

**Explorer URL Pattern**:
```
https://circularlabs.io/Explorer?network={network}&tx={transaction_id}
```

---

## Error Handling

All tools use a consistent error response format:

```json
{
  "type": "ERROR_TYPE",
  "status_code": 400,
  "message": "Human-readable error description",
  "retry_suggestion": "Guidance for agent (optional)"
}
```

### Common Error Types

| Error Type | Status | Retry? | Tools Affected |
|------------|--------|--------|----------------|
| `API_UNAVAILABLE` | 503 | Yes (30s) | All |
| `NETWORK_ERROR` | 503 | Yes (30s) | All |
| `INVALID_INPUT` | 400 | No | All |
| `INVALID_ADDRESS` | 400 | No | get_wallet_nonce |
| `INVALID_SIGNATURE` | 400 | No | certify_data |
| `INSUFFICIENT_BALANCE` | 402 | No | certify_data |
| `NONCE_MISMATCH` | 400 | Yes (fetch nonce) | certify_data |
| `RATE_LIMIT` | 429 | Yes (60s) | certify_data |
| `TRANSACTION_TIMEOUT` | 504 | Yes (later) | get_transaction_status |
| `TRANSACTION_FAILED` | 400 | No | get_transaction_status, get_certification_proof |

---

## Usage Flow

### Complete Certification Workflow

```
1. get_wallet_nonce
   Input: { wallet_address, network }
   Output: { current_nonce }

2. certify_data
   Input: { data, network }
   Output: { transaction_id, status: "Pending" }

3. get_transaction_status (poll)
   Input: { transaction_id, network, wait_for_execution: true }
   Output: { status: "Executed", block_id, timestamp }

4. get_certification_proof
   Input: { transaction_id, network }
   Output: { block_id, timestamp, explorer_url }
```

### Minimal Certification (Auto-Nonce)

```
1. certify_data
   Input: { data, network }
   (Server fetches nonce automatically)
   Output: { transaction_id }

2. get_transaction_status
   Input: { transaction_id, network }
   Output: { status: "Executed" }
```

---

## Schema Validation

All JSON schemas follow JSON Schema Draft 07 specification:
- Input validation via `required` and `properties` fields
- Type safety with `type`, `enum`, `minLength`, `maxLength`
- Examples provided for documentation
- Error scenarios documented

**Testing**: Contract tests must validate:
1. Schema structure matches MCP specification
2. All required fields are present
3. Enum values are correct
4. Examples are valid against schema
5. Error types cover all failure scenarios

---

## Network Configuration

Both testnet and mainnet are supported:

**Testnet**:
- For testing and development
- Free CIRX tokens from faucet
- Explorer: https://circularlabs.io/Explorer?network=testnet

**Mainnet**:
- For production use
- Requires real CIRX balance
- Explorer: https://circularlabs.io/Explorer?network=mainnet

---

## Contract Validation Checklist

- [x] get_wallet_nonce.json created and validated
- [x] certify_data.json created and validated
- [x] get_transaction_status.json created and validated
- [x] get_certification_proof.json created and validated
- [x] All schemas follow JSON Schema Draft 07
- [x] Input/output types defined
- [x] Required fields specified
- [x] Error scenarios documented
- [x] Examples provided for each tool
- [x] Consistent error format across all tools

**Status**: **PHASE 1b COMPLETE** - All tool contracts defined

---

## Next Steps

1. Implement tools in Go (`tools/*.go`)
2. Create contract tests (`tests/contract/*_test.go`)
3. Validate schemas against test vectors
4. Update quickstart.md with usage examples

---

**Contracts Completed**: 2025-10-30
**Designer**: Claude (Sonnet 4.5)
**Approved for Implementation**: ✅ Ready
