# Quick Start: x402 Payment MCP Server

## Prerequisites

- Go 1.23+ (available in `nix develop`)
- Base Sepolia testnet USDC (for testing)
- RPC endpoint access (e.g., Alchemy, Infura, or public RPCs)

## Setup

### 1. Create Configuration

```bash
cd mcp-servers/x402-mcp-server
cp config.yaml.example config.yaml
```

Edit `config.yaml`:

```yaml
networks:
  base-sepolia:
    chain_id: 84532
    usdc_contract: "0x036CbD53842c5426634e7929541eC2318f3dCF7e"
    facilitator_url: "https://x402.org/facilitator"
    rpc_url: "https://sepolia.base.org"
    payee_address: "YOUR_TEST_WALLET_ADDRESS"

eip712:
  domain_name: "USD Coin"
  domain_version: "2"

logging:
  level: "DEBUG"
  format: "json"

cache:
  settlement_ttl_minutes: 10
```

### 2. Install Dependencies

```bash
go mod download
go mod tidy
```

### 3. Run MCP Server

```bash
# Standalone (stdio)
go run main.go

# Or via HTTP proxy integration (production mode)
# The HTTP proxy will spawn this as subprocess
```

### 4. Test Tools

```bash
# Send MCP request via stdio
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"create_payment_requirement","arguments":{"amount":"50000","network":"base-sepolia"}}}' | go run main.go
```

## Development Workflow

### TDD Cycle

1. **Write failing test** (Red phase):
   ```bash
   cd tests/unit
   # Edit x402_test.go - add test for new functionality
   go test ./... -v # Should fail
   ```

2. **Implement minimum code** (Green phase):
   ```bash
   # Edit internal/x402/payment_requirement.go
   go test ./... -v # Should pass
   ```

3. **Refactor** (Refactor phase):
   ```bash
   # Improve code quality
   go test ./... -v # Should still pass
   ```

### Run Tests

```bash
# Unit tests only
go test ./tests/unit/... -v

# Integration tests (requires network access)
go test ./tests/integration/... -v

# All tests with coverage
go test ./... -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Logging

Structured JSON logs output to stdout:

```json
{
  "level": "INFO",
  "ts": "2025-10-28T12:34:56Z",
  "msg": "tool called",
  "tool_name": "create_payment_requirement",
  "network": "base-sepolia",
  "amount": "50000",
  "duration_ms": 45
}
```

## Testing Payment Flow

### 1. Generate Payment Requirement

```go
// Tool: create_payment_requirement
{
  "amount": "50000",  // 0.05 USDC
  "network": "base-sepolia"
}

// Response:
{
  "x402_version": 1,
  "scheme": "exact",
  "network": "base-sepolia",
  "maxAmountRequired": "50000",
  "payee": "0xYourPayeeAddress",
  "valid_until": "2025-10-28T12:40:00Z",
  "nonce": "0x5",
  "asset": "0x036CbD53842c5426634e7929541eC2318f3dCF7e"
}
```

### 2. Sign Authorization (Agent Wallet)

Use agent's private key to sign EIP-712 message (see research.md for structure).

### 3. Verify Payment

```go
// Tool: verify_payment
{
  "authorization": {
    "from": "0xPayerAddress",
    "to": "0xPayeeAddress",
    "value": "50000",
    "validAfter": 0,
    "validBefore": 1730123456,
    "nonce": "0x...",
    "v": 27,
    "r": "0x...",
    "s": "0x..."
  },
  "network": "base-sepolia"
}

// Response:
{
  "is_valid": true,
  "signer_address": "0xPayerAddress"
}
```

### 4. Settle Payment

```go
// Tool: settle_payment
{
  "authorization": { /* same as verify */ },
  "payment_requirement": { /* from step 1 */ }
}

// Response:
{
  "status": "settled",
  "tx_hash": "0x123...",
  "block_number": 12345
}
```

## Configuration Examples

### Production (Base Mainnet)

```yaml
networks:
  base:
    chain_id: 8453
    usdc_contract: "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
    facilitator_url: "https://api.cdp.coinbase.com/platform/v2/x402/"
    rpc_url: "${BASE_RPC_URL}"  # Set via environment
    payee_address: "${PAYEE_ADDRESS_BASE}"

logging:
  level: "INFO"  # Less verbose for production
```

### Multiple Networks

```yaml
networks:
  base:
    chain_id: 8453
    usdc_contract: "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
    facilitator_url: "https://api.cdp.coinbase.com/platform/v2/x402/"
    rpc_url: "https://mainnet.base.org"
    payee_address: "${PAYEE_ADDRESS_BASE}"

  arbitrum:
    chain_id: 42161
    usdc_contract: "0xaf88d065e77c8cC2239327C5EDb3A432268e5831"
    facilitator_url: "https://api.cdp.coinbase.com/platform/v2/x402/"
    rpc_url: "https://arb1.arbitrum.io/rpc"
    payee_address: "${PAYEE_ADDRESS_ARBITRUM}"
```

## Troubleshooting

### Nonce Retrieval Fails

**Symptom**: `create_payment_requirement` returns error "nonce retrieval failed"

**Solutions**:
- Check RPC URL is accessible: `curl https://sepolia.base.org`
- Verify network connectivity
- Check RPC endpoint rate limits

### Signature Verification Fails

**Symptom**: `verify_payment` returns `is_valid: false`

**Solutions**:
- Verify EIP-712 domain parameters match network (see research.md)
- Ensure signature was created for correct chainId
- Check `from` address matches signer
- Validate time bounds (validAfter, validBefore)

### Settlement Timeout

**Symptom**: `settle_payment` returns `status: "pending"`

**Solutions**:
- Facilitator may be processing - check again after retry_after seconds
- Verify facilitator_url is correct for network
- Check network connectivity to facilitator

## Next Steps

1. **Implement User Story 1 (P1)**: `create_payment_requirement` tool
2. **Implement User Story 2 (P1)**: `verify_payment` tool
3. **Run `/speckit.tasks`** to generate detailed task breakdown
4. **Follow TDD workflow** for each task

See `plan.md` for complete architecture and `research.md` for technical details.
