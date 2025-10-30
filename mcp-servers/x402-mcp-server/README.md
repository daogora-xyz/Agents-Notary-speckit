# x402 Payment MCP Server

Model Context Protocol (MCP) server providing blockchain payment capabilities for AI agents via the x402 payment protocol and EIP-3009 payment authorizations.

## Overview

The x402 Payment MCP Server enables AI agents to:
- Generate **x402 v1 compliant** payment requirements with full spec support
- Verify cryptographic payment authorizations (EIP-3009)
- Submit verified payments to the x402 facilitator for on-chain settlement

**Standards Compliance:**
- ✅ Full [Coinbase x402 v1 specification](https://github.com/coinbase/x402) support
- ✅ EIP-3009: Transfer With Authorization
- ✅ EIP-712: Typed Structured Data Hashing and Signing

**Supported Networks:** USDC payments on Base, Base Sepolia, and Arbitrum

## Features

### MCP Tools

1. **create_payment_requirement** - Generate x402 v1 compliant payment requirements
   - Creates structured payment metadata with resource URL and description
   - Returns complete x402 specification fields (scheme, network, payTo, asset, extra metadata)
   - Generates unique nonces for payment authorization
   - Supports custom MIME types and timeout configuration

2. **verify_payment** - Verify EIP-3009 signatures
   - Validates ECDSA signatures using secp256k1 recovery
   - Checks EIP-712 domain parameters
   - Validates time bounds (validAfter/validBefore)

3. **settle_payment** - Submit payments to facilitator
   - Verifies signature before submission (FR-011)
   - Submits to x402 facilitator for on-chain settlement
   - Implements idempotency caching (prevents duplicate submissions)
   - Returns settlement status (settled/pending/failed)

### Architecture

- **EIP-712 Typed Data**: Structured signature verification
- **EIP-3009**: `receiveWithAuthorization` payment standard
- **Facilitator Integration**: HTTP client for x402 facilitator API
- **Multi-Network**: Base (mainnet), Base Sepolia (testnet), Arbitrum
- **Security**: Signature verification before settlement, time-bound authorizations

## Quick Start

### Prerequisites

- Go 1.21 or later
- Access to the Nix development environment (optional but recommended)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/lessuseless/agents-notary.git
cd agents-notary/mcp-servers/x402-mcp-server
```

2. Create configuration file:
```bash
cp config.yaml.example config.yaml
```

3. Configure environment variables in `config.yaml`:
```yaml
networks:
  base:
    payee_address: "0xYourPayeeAddress"
  base-sepolia:
    payee_address: "0xYourTestPayeeAddress"
```

### Running the Server

**With Nix (recommended):**
```bash
nix develop /path/to/Agents-Notary-speckit --command go run ./cmd/server
```

**Standard Go:**
```bash
go run ./cmd/server
```

The server will start listening on stdio for MCP protocol messages.

### Testing

**Run all tests:**
```bash
nix develop --command go test ./...
```

**Run with coverage:**
```bash
nix develop --command go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**Run with race detection:**
```bash
nix develop --command go test -race ./...
```

## Configuration

Configuration is loaded from `config.yaml` in the server directory.

### Example Configuration

```yaml
networks:
  base:
    chain_id: 8453
    usdc_contract: "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
    facilitator_url: "https://api.cdp.coinbase.com/platform/v2/x402/"
    rpc_url: "https://mainnet.base.org"
    payee_address: "${PAYEE_ADDRESS_BASE}"

  base-sepolia:
    chain_id: 84532
    usdc_contract: "0x036CbD53842c5426634e7929541eC2318f3dCF7e"
    facilitator_url: "https://x402.org/facilitator"
    rpc_url: "https://sepolia.base.org"
    payee_address: "${PAYEE_ADDRESS_SEPOLIA}"

eip712:
  domain_name: "USD Coin"
  domain_version: "2"

logging:
  level: "INFO"  # DEBUG, INFO, WARN, ERROR
  format: "json"

cache:
  settlement_ttl_minutes: 10
```

### Environment Variables

Environment variables can be referenced in `config.yaml` using `${VARIABLE_NAME}` syntax:

- `PAYEE_ADDRESS_BASE` - Your payee address for Base mainnet
- `PAYEE_ADDRESS_SEPOLIA` - Your payee address for Base Sepolia testnet
- `PAYEE_ADDRESS_ARBITRUM` - Your payee address for Arbitrum

## Usage Example

### 1. Create Payment Requirement

```json
{
  "method": "tools/call",
  "params": {
    "name": "create_payment_requirement",
    "arguments": {
      "amount": "5000000",
      "network": "base-sepolia",
      "resource": "https://api.example.com/protected-resource",
      "description": "API access fee for premium features",
      "mime_type": "application/json"
    }
  }
}
```

Response (x402 v1 compliant):
```json
{
  "scheme": "exact",
  "network": "base-sepolia",
  "maxAmountRequired": "5000000",
  "resource": "https://api.example.com/protected-resource",
  "description": "API access fee for premium features",
  "mimeType": "application/json",
  "payTo": "0xYourPayeeAddress",
  "maxTimeoutSeconds": 60,
  "asset": "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
  "extra": {
    "name": "USD Coin",
    "version": "2"
  },
  "x402_version": 1,
  "valid_until": "2025-10-30T12:00:00Z",
  "nonce": "0x..."
}
```

### 2. Verify Payment Authorization

```json
{
  "method": "tools/call",
  "params": {
    "name": "verify_payment",
    "arguments": {
      "authorization": {
        "from": "0xPayerAddress",
        "to": "0xPayeeAddress",
        "value": "5000000",
        "validAfter": 0,
        "validBefore": 1735689600,
        "nonce": "0x...",
        "v": 27,
        "r": "0x...",
        "s": "0x..."
      },
      "network": "base-sepolia"
    }
  }
}
```

Response:
```json
{
  "is_valid": true,
  "signer_address": "0xPayerAddress",
  "matches_from": true
}
```

### 3. Settle Payment

```json
{
  "method": "tools/call",
  "params": {
    "name": "settle_payment",
    "arguments": {
      "authorization": {
        "from": "0xPayerAddress",
        "to": "0xPayeeAddress",
        "value": "5000000",
        "validAfter": 0,
        "validBefore": 1735689600,
        "nonce": "0x...",
        "v": 27,
        "r": "0x...",
        "s": "0x..."
      },
      "network": "base-sepolia"
    }
  }
}
```

Response:
```json
{
  "status": "settled",
  "tx_hash": "0x...",
  "block_number": 12345678,
  "network": "base-sepolia"
}
```

## Project Structure

```
x402-mcp-server/
├── cmd/
│   └── server/
│       └── main.go              # Server entry point
├── internal/
│   ├── config/                  # Configuration loading and validation
│   ├── eip3009/                 # EIP-3009 signature verification
│   ├── eip712/                  # EIP-712 typed data handling
│   ├── facilitator/             # x402 facilitator HTTP client
│   ├── logger/                  # Structured logging
│   ├── server/                  # Core server implementation
│   └── x402/                    # x402 protocol implementation
│       └── payment_requirement.go
├── tools/
│   ├── create_payment_requirement.go
│   ├── verify_payment.go
│   └── settle_payment.go
├── tests/
│   ├── unit/                    # Unit tests
│   ├── contract/                # Contract tests
│   └── integration/             # Integration tests
├── config.yaml.example          # Example configuration
├── go.mod                       # Go module definition
└── README.md                    # This file
```

## Security Considerations

1. **Signature Verification**: All payments are cryptographically verified before settlement
2. **Time Bounds**: Authorizations have validAfter/validBefore timestamps
3. **Nonce Uniqueness**: Each authorization uses a unique nonce to prevent replay attacks
4. **Idempotency**: Settlement caching prevents duplicate submissions
5. **Domain Separation**: EIP-712 domain parameters ensure signatures are network-specific

## Development

### Adding New Networks

1. Add network configuration to `config.yaml`:
```yaml
networks:
  new-network:
    chain_id: 1234
    usdc_contract: "0x..."
    facilitator_url: "https://..."
    rpc_url: "https://..."
    payee_address: "0x..."
```

2. Update network enum in tool schemas (if using enum validation)

### Running Tests in Development

```bash
# Unit tests only
nix develop --command go test ./tests/unit/... -v

# Contract tests only
nix develop --command go test ./tests/contract/... -v

# Specific test
nix develop --command go test ./tests/unit/eip3009_verification_test.go -v
```

## Troubleshooting

### "Failed to load config" Error

Ensure `config.yaml` exists in the server directory and is valid YAML.

### "Invalid config" Error

Check that all required fields are present:
- Network configurations (chain_id, usdc_contract, facilitator_url, payee_address)
- EIP-712 domain parameters (domain_name, domain_version)

### "Signature verification failed" Error

Verify:
- Correct network is specified
- EIP-712 domain parameters match the signing wallet's expectations
- Signature components (v, r, s) are correctly formatted (0x-prefixed hex)
- Time bounds are valid (current time is between validAfter and validBefore)

## License

See the main repository LICENSE file.

## References

- [Model Context Protocol](https://modelcontextprotocol.io/)
- [EIP-3009: Transfer With Authorization](https://eips.ethereum.org/EIPS/eip-3009)
- [EIP-712: Typed Structured Data Hashing](https://eips.ethereum.org/EIPS/eip-712)
- [x402 Protocol Documentation](https://x402.org/docs)
- [Circle USDC Developer Docs](https://developers.circle.com/)
