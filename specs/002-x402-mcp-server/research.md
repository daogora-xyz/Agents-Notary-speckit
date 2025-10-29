# Technical Research: x402 Payment MCP Server

**Date**: 2025-10-28
**Purpose**: Resolve technical unknowns for x402 payment tool implementation

## Summary of Decisions

| Topic | Decision | Source |
|-------|----------|--------|
| x402 Protocol | Version 1, "exact" scheme | [x402.org](https://www.x402.org/), [Coinbase CDP](https://docs.cdp.coinbase.com) |
| EIP-3009 Function | `receiveWithAuthorization` | [EIP-3009](https://eips.ethereum.org/EIPS/eip-3009), [Circle SDK](https://www.circle.com/blog/four-ways-to-authorize-usdc-smart-contract-interactions-with-circle-sdk) |
| EIP-712 Domain | name="USD Coin", version="2" | [FiatTokenV2_2](https://github.com/FraxFinance/fraxtal-usdc/blob/master/contracts/v2/FiatTokenV2_2.sol) |
| EIP-681 Format | Standard ERC-20 transfer URI | [EIP-681](https://eips.ethereum.org/EIPS/eip-681) |
| USDC Contracts | Native USDC (not bridged) | [Circle Docs](https://developers.circle.com/stablecoins/usdc-contract-addresses) |
| MetaMask Links | link.metamask.io/send format | [MetaMask Docs](https://docs.metamask.io/sdk/guides/use-deeplinks/) |

## 1. x402 Facilitator API

### Decision
Use Coinbase CDP facilitator endpoint for production, x402.org for Base Sepolia testnet.

### Endpoints

**Base Sepolia Testnet:**
```
Base URL: https://x402.org/facilitator
```

**Production (Base, Arbitrum):**
```
Base URL: https://api.cdp.coinbase.com/platform/v2/x402/
```

### API Methods

**POST /settle**
```json
Request:
{
  "x402Version": 1,
  "paymentPayload": {
    "from": "0x...",
    "to": "0x...",
    "value": "50000",
    "validAfter": 0,
    "validBefore": 1730000000,
    "nonce": "0x...",
    "v": 27,
    "r": "0x...",
    "s": "0x..."
  },
  "paymentRequirements": {
    "scheme": "exact",
    "network": "base",
    "maxAmountRequired": "50000",
    "payTo": "0x...",
    "asset": "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
  }
}

Response (success):
{
  "success": true,
  "txHash": "0x...",
  "networkId": "base",
  "payer": "0x..."
}

Response (error):
{
  "success": false,
  "error": "insufficient_funds" | "invalid_network" | "invalid_x402_version"
}
```

### Rationale
- Coinbase CDP provides production-grade facilitator with SLA
- x402.org reference implementation suitable for testnet
- Abstracts blockchain settlement complexity

##  2. EIP-3009 Authorization

### Decision
Use `receiveWithAuthorization` function (NOT `transferWithAuthorization`)

### Function Signature
```solidity
function receiveWithAuthorization(
    address from,        // Payer
    address to,          // Payee (must be msg.sender - prevents front-running)
    uint256 value,       // Amount in atomic units
    uint256 validAfter,  // Unix timestamp
    uint256 validBefore, // Unix timestamp
    bytes32 nonce,       // Random 32 bytes
    uint8 v,
    bytes32 r,
    bytes32 s
) external;
```

### EIP-712 Message Structure
```javascript
{
  types: {
    EIP712Domain: [
      { name: "name", type: "string" },
      { name: "version", type: "string" },
      { name: "chainId", type: "uint256" },
      { name: "verifyingContract", type: "address" }
    ],
    ReceiveWithAuthorization: [
      { name: "from", type: "address" },
      { name: "to", type: "address" },
      { name: "value", type: "uint256" },
      { name: "validAfter", type: "uint256" },
      { name: "validBefore", type: "uint256" },
      { name: "nonce", type: "bytes32" }
    ]
  },
  primaryType: "ReceiveWithAuthorization",
  domain: {
    name: "USD Coin",
    version: "2",
    chainId: 8453,  // or 84532, 42161
    verifyingContract: "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
  },
  message: {
    from: "0x...",
    to: "0x...",
    value: "50000",
    validAfter: 0,
    validBefore: 1730000000,
    nonce: "0x..."
  }
}
```

### Rationale
- `receiveWithAuthorization` prevents front-running (payee must be caller)
- Random nonces more flexible than sequential (EIP-2612)
- Gasless for payer (facilitator pays gas)

## 3. Network Configurations

### USDC Contract Addresses (Native)

| Network | Chain ID | USDC Contract | EIP-712 Domain |
|---------|----------|---------------|----------------|
| **Base** | 8453 | `0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913` | name="USD Coin", version="2" |
| **Base Sepolia** | 84532 | `0x036CbD53842c5426634e7929541eC2318f3dCF7e` | name="USD Coin", version="2" |
| **Arbitrum** | 42161 | `0xaf88d065e77c8cC2239327C5EDb3A432268e5831` | name="USD Coin", version="2" |

**Important**: Use **native USDC**, NOT bridged USDC.e on Arbitrum (`0xFF970A61A04b1cA14834A43f5dE4533eBDDB5CC8` is deprecated)

### Configuration Template (config.yaml)

```yaml
networks:
  base:
    chain_id: 8453
    usdc_contract: "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
    facilitator_url: "https://api.cdp.coinbase.com/platform/v2/x402/"
    rpc_url: "https://mainnet.base.org"
    payee_address: "${PAYEE_ADDRESS_BASE}"  # From environment

  base-sepolia:
    chain_id: 84532
    usdc_contract: "0x036CbD53842c5426634e7929541eC2318f3dCF7e"
    facilitator_url: "https://x402.org/facilitator"
    rpc_url: "https://sepolia.base.org"
    payee_address: "${PAYEE_ADDRESS_SEPOLIA}"

  arbitrum:
    chain_id: 42161
    usdc_contract: "0xaf88d065e77c8cC2239327C5EDb3A432268e5831"
    facilitator_url: "https://api.cdp.coinbase.com/platform/v2/x402/"
    rpc_url: "https://arb1.arbitrum.io/rpc"
    payee_address: "${PAYEE_ADDRESS_ARBITRUM}"

eip712:
  domain_name: "USD Coin"
  domain_version: "2"

logging:
  level: "INFO"  # DEBUG, INFO, WARN, ERROR
  format: "json"

cache:
  settlement_ttl_minutes: 10
```

## 4. EIP-681 Payment URIs

### Format
```
ethereum:<token_contract>@<chain_id>/transfer?address=<recipient>&uint256=<amount>
```

### Examples

**Base:**
```
ethereum:0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913@8453/transfer?address=0xPayee&uint256=50000
```

**Base Sepolia:**
```
ethereum:0x036CbD53842c5426634e7929541eC2318f3dCF7e@84532/transfer?address=0xPayee&uint256=50000
```

**Arbitrum:**
```
ethereum:0xaf88d065e77c8cC2239327C5EDb3A432268e5831@42161/transfer?address=0xPayee&uint256=50000
```

### Rationale
- Recognized by MetaMask Mobile, Rainbow, Coinbase Wallet
- Chain ID enables automatic network switching
- Amount in atomic units (6 decimals for USDC)

## 5. MetaMask Deep Links

### Format
```
https://link.metamask.io/send/<token>@<chain_id>/transfer?address=<recipient>&uint256=<amount>
```

### Examples

**Base:**
```
https://link.metamask.io/send/0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913@8453/transfer?address=0xPayee&uint256=50000
```

### Callback URL Handling

**Status**: ⚠️ No native callback support in MetaMask deep links

**Workarounds**:
1. **Client-side polling**: JavaScript monitors MetaMask events, redirects with tx hash
2. **Blockchain polling**: HTTP proxy queries blockchain for expected payment
3. **WebSocket subscription**: Subscribe to mempool/blocks for payee address

**Recommendation for MVP**: Client-side JavaScript polling (simplest for User Story 4)

## 6. Signature Verification Process

### Steps

1. **Parse Authorization**:
   ```go
   type Authorization struct {
       From        string `json:"from"`
       To          string `json:"to"`
       Value       string `json:"value"`
       ValidAfter  uint64 `json:"validAfter"`
       ValidBefore uint64 `json:"validBefore"`
       Nonce       string `json:"nonce"`
       V           uint8  `json:"v"`
       R           string `json:"r"`
       S           string `json:"s"`
   }
   ```

2. **Construct EIP-712 Hash**:
   ```go
   domainSeparator := hashStruct(EIP712Domain{
       Name: "USD Coin",
       Version: "2",
       ChainId: network.ChainID,
       VerifyingContract: network.USDCContract,
   })

   messageHash := hashStruct(ReceiveWithAuthorization{
       From: auth.From,
       To: auth.To,
       Value: auth.Value,
       ValidAfter: auth.ValidAfter,
       ValidBefore: auth.ValidBefore,
       Nonce: auth.Nonce,
   })

   digest := keccak256("\x19\x01" + domainSeparator + messageHash)
   ```

3. **Recover Signer**:
   ```go
   publicKey := ecrecover(digest, auth.V, auth.R, auth.S)
   signerAddress := publicKeyToAddress(publicKey)
   ```

4. **Validate**:
   - `signerAddress == auth.From` (signature matches claimed payer)
   - `validAfter <= now < validBefore` (time bounds)
   - Signature components valid (v=27 or 28, r/s non-zero)

### Go Libraries

```go
import (
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/signer/core/apitypes"
)
```

## 7. Nonce Fetching via RPC

### Method
```
eth_getTransactionCount(address, "latest")
```

### Example (Base RPC)
```bash
curl https://mainnet.base.org \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "eth_getTransactionCount",
    "params": ["0xPayerAddress", "latest"],
    "id": 1
  }'

# Response: {"jsonrpc":"2.0","id":1,"result":"0x5"} # 5 transactions
```

### Go Implementation
```go
client, _ := ethclient.Dial(network.RPCURL)
nonce, _ := client.NonceAt(context.Background(), common.HexToAddress(payerAddress), nil)
// nonce is uint64, use as part of payment requirement nonce generation
```

### Retry Logic (Edge Case)
```go
const maxRetries = 3
for attempt := 0; attempt < maxRetries; attempt++ {
    nonce, err := fetchNonce(address)
    if err == nil {
        return nonce, nil
    }
    time.Sleep(time.Second * time.Duration(attempt+1))
}
return 0, errors.New("nonce retrieval failed after 3 attempts")
```

## Resolved Unknowns

All technical unknowns from spec Open Questions (Q1, Q3-Q7) have been resolved:

- **Q1 (x402 API spec)**: ✅ Resolved - Coinbase CDP documented
- **Q2 (nonce generation)**: ✅ Resolved via clarification - blockchain-sourced
- **Q3 (additional testnets)**: ℹ️ Deferred - Base Sepolia sufficient for MVP
- **Q4 (EIP-712 domain)**: ✅ Resolved - name="USD Coin", version="2"
- **Q5 (facilitator URL config)**: ✅ Resolved - per-network configuration
- **Q6 (facilitator version mismatch)**: ℹ️ Operational concern - monitor facilitator API versions
- **Q7 (EIP-681 parameter ordering)**: ✅ Resolved - no specific ordering required

## Next Steps

Phase 1 artifacts ready for generation:
- `data-model.md` - Entity definitions with validated field types
- `contracts/*.json` - MCP tool JSON schemas with researched parameters
- `quickstart.md` - Developer setup with network configuration examples
