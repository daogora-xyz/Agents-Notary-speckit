# Fixing Remaining Test Signature Issues

## Summary

✅ **COMPLETED:**
- Created `tests/contract/test_helpers.go` with signature generation functions
- Fixed `TestSettlePayment_Execute_SuccessfulSettlement` - **PASSING**
- Fixed `TestSettlePayment_Execute_FacilitatorError` - Ready to test

## Pattern to Apply

Replace hardcoded signature blocks with real signature generation:

### Before (hardcoded):
```go
input := map[string]interface{}{
    "authorization": map[string]interface{}{
        "from":        "0x1111111111111111111111111111111111111111",
        "to":          "0x2222222222222222222222222222222222222222",
        "value":       "50000",
        "validAfter":  float64(now - 3600),
        "validBefore": float64(now + 3600),
        "nonce":       "0x0000000000000000000000000000000000000000000000000000000000000001",
        "v":           float64(27),
        "r":           "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
        "s":           "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
    },
    "network": "base",
}
```

### After (real signatures):
```go
// Generate valid signature for testing
privateKey, fromAddr, err := createTestPrivateKeyAndAddress()
if err != nil {
    t.Fatalf("Failed to create test private key: %v", err)
}

toAddr := common.HexToAddress("0x2222222222222222222222222222222222222222")
value := big.NewInt(50000)
now := time.Now().Unix()
validAfter := big.NewInt(now - 3600)
validBefore := big.NewInt(now + 3600)
var nonce [32]byte
copy(nonce[:], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})

chainID := big.NewInt(8453) // Base mainnet (or 84532 for Base Sepolia)
usdcContract := common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913")

v, r, s, err := generateValidSignature(privateKey, fromAddr, toAddr, value, validAfter, validBefore, nonce, chainID, usdcContract)
if err != nil {
    t.Fatalf("Failed to generate valid signature: %v", err)
}

input := map[string]interface{}{
    "authorization": map[string]interface{}{
        "from":        fromAddr.Hex(),
        "to":          toAddr.Hex(),
        "value":       "50000",
        "validAfter":  float64(validAfter.Int64()),
        "validBefore": float64(validBefore.Int64()),
        "nonce":       common.BytesToHash(nonce[:]).Hex(),
        "v":           float64(v),
        "r":           common.BytesToHash(r.Bytes()).Hex(),
        "s":           common.BytesToHash(s.Bytes()).Hex(),
    },
    "network": "base",
}
```

## Remaining Tests to Fix

### Contract Tests (`tests/contract/settle_payment_test.go`):
- [x] TestSettlePayment_Execute_SuccessfulSettlement - **FIXED & PASSING** ✅
- [x] TestSettlePayment_Execute_FacilitatorError - **FIXED & PASSING** ✅
- [x] TestSettlePayment_Execute_Idempotency - **FIXED & PASSING** ✅
- [x] TestSettlePayment_Execute_NetworkTimeout - **FIXED & PASSING** ✅
- [x] TestSettlePayment_JSONOutput - **FIXED & PASSING** ✅

### Integration Tests (`tests/integration/payment_flow_integration_test.go`):
Need to add the same helper file and apply pattern to all tests.

## Step-by-Step Instructions

### 1. For Remaining Contract Tests

Search for hardcoded signatures:
```bash
grep -n "0x1234567890abcdef" tests/contract/settle_payment_test.go
```

For each occurrence, apply the pattern above. Make sure to:
- Change the nonce byte (last byte) for uniqueness:
  - Test 1: `..., 0, 0, 0, 1}`
  - Test 2: `..., 0, 0, 0, 2}`
  - Test 3: `..., 0, 0, 0, 3}`

### 2. For Integration Tests

Copy `test_helpers.go` to integration test directory:
```bash
cp tests/contract/test_helpers.go tests/integration/
```

Update package declaration:
```go
package integration  // Change from 'package contract'
```

Apply the same signature generation pattern to all integration tests.

### 3. Run All Tests

After fixes:
```bash
nix develop --command go test ./tests/contract/... -v
nix develop --command go test ./tests/integration/... -v
```

## Network Configuration Reference

**Base Mainnet:**
- Chain ID: 8453
- USDC Contract: 0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913

**Base Sepolia:**
- Chain ID: 84532
- USDC Contract: 0x036CbD53842c5426634e7929541eC2318f3dCF7e

**Arbitrum:**
- Chain ID: 42161
- USDC Contract: 0xaf88d065e77c8cC2239327C5EDb3A432268e5831

## Quick Test Command

Test all contract tests:
```bash
nix develop --command go test ./tests/contract/... -v 2>&1 | grep -E "(PASS|FAIL|RUN)"
```

Test specific integration test:
```bash
nix develop --command go test ./tests/integration/... -v -run TestEndToEndPaymentFlow
```

## Success Criteria

All tests should show:
```
--- PASS: TestName (X.XXs)
PASS
ok  	github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/tests/contract	X.XXXs
```

No signature verification errors should appear.
