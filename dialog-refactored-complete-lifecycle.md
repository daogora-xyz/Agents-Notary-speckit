# Complete Certification Lifecycle: User → Agent → certify.ar4s.com → 4 MCP Servers

**Architecture Context:**
- User's LLM/Agent has NO native MCP support, NO wallet/private keys
- certify.ar4s.com is an HTTP proxy that acts as an MCP Host (using mcp-go client library)
- 4 MCP servers run as subprocesses, communicate via stdio transport
- All MCP tool calls orchestrated by certify.ar4s.com's mcp-go client

**Notation:**
- `User →` / `← User` : Natural language conversation
- `LLM/Agent [HTTP] →` / `← [HTTP]` : HTTP REST API calls
- `HTTP Proxy [MCP-GO] →` / `← [MCP-GO]` : mcp-go client tool invocations via stdio
- `MCP Server [External] →` / `← [External]` : MCP server calls to external services
- `{Internal Logic}` : Processing within a component

---

## Phase 1: Initial Request & Quote Generation

**User → LLM/Agent:**
```
"Hey, could you help me certify 'ashley-barr_offerletter-contract--unsigned.pdf'
using certify.ar4s.com? The file should be attached."
```

**LLM/Agent → User:**
```
"Sure, no problem! Let me get a quote first to see how much this will cost."
```

### LLM/Agent initiates HTTP request

**LLM/Agent [HTTP] → POST https://certify.ar4s.com/v1/certify**
```json
{
  "request_id": "req_offerletter_20251031_abc123",
  "data": "base64_encoded_pdf_content...",
  "network": "base",
  "client_type": "agent"
}
```

---

## Phase 2: HTTP Proxy Orchestrates Quote via MCP Servers

### HTTP Proxy receives request, starts orchestration

**HTTP Proxy {Internal Logic}:**
```go
// certify.ar4s.com proxy/internal/orchestration/workflow.go
func (s *Service) HandleCertifyRequest(req CertifyRequest) Response {
    // Check idempotency
    if existing := s.db.GetRequest(req.RequestID); existing != nil {
        return s.handleExistingRequest(existing)
    }

    // Initialize mcp-go clients (stdio transport to 4 MCP servers)
    dataQuoteClient := s.mcpClients["data-quote-mcp"]
    x402Client := s.mcpClients["x402-mcp"]

    // Phase 1: Get quote via data-quote-mcp-server
    // ...
}
```

### Tool Call 1: Check data size

**HTTP Proxy [MCP-GO] → data-quote-mcp-server.check_data_size**
```json
{
  "tool": "check_data_size",
  "arguments": {
    "data": "base64_encoded_pdf_content..."
  }
}
```

**data-quote-mcp-server {Processing}:**
```go
// mcp-servers/data-quote-mcp-server/tools/check_size.go
func handleCheckDataSize(args CheckSizeArgs) (CheckSizeResult, error) {
    decoded, err := base64.StdEncoding.DecodeString(args.Data)
    sizeBytes := len(decoded)
    sizeKB := float64(sizeBytes) / 1024

    return CheckSizeResult{
        SizeBytes: sizeBytes,
        SizeKB: sizeKB,
        SizeHuman: fmt.Sprintf("%.2f KB", sizeKB),
    }, nil
}
```

**data-quote-mcp-server [MCP-GO] ← HTTP Proxy:**
```json
{
  "size_bytes": 45678,
  "size_kb": 44.61,
  "size_human": "44.61 KB"
}
```

### Tool Call 2: Get CIRX price

**HTTP Proxy [MCP-GO] → data-quote-mcp-server.get_cirx_price**
```json
{
  "tool": "get_cirx_price",
  "arguments": {}
}
```

**data-quote-mcp-server [External] → CoinGecko API**
```
GET https://api.coingecko.com/api/v3/simple/price?ids=circular&vs_currencies=usd
```

**data-quote-mcp-server [External] ← CoinGecko API:**
```json
{
  "circular": {
    "usd": 0.0044
  }
}
```

**data-quote-mcp-server {Caching}:**
```go
// Cache price for 5 minutes
s.priceCache.Set("cirx_usd", 0.0044, 5*time.Minute)
```

**data-quote-mcp-server [MCP-GO] ← HTTP Proxy:**
```json
{
  "cirx_price_usd": 0.0044,
  "cached_at": "2025-10-31T14:30:00Z",
  "expires_at": "2025-10-31T14:35:00Z"
}
```

### Tool Call 3: Calculate quote

**HTTP Proxy [MCP-GO] → data-quote-mcp-server.calculate_quote**
```json
{
  "tool": "calculate_quote",
  "arguments": {
    "data_size_bytes": 45678,
    "cirx_price_usd": 0.0044,
    "margin_percent": 65.0
  }
}
```

**data-quote-mcp-server {Calculation}:**
```go
// Formula: (4 CIRX × CIRX_price_USD) × (1 + margin%)
cirxCostUSD := 4 * 0.0044          // $0.0176
serviceMargin := cirxCostUSD * 0.65 // $0.01144
totalUSD := cirxCostUSD + serviceMargin // $0.02904
roundedUSDC := math.Ceil(totalUSD * 100) / 100 // $0.03 (round up)
```

**data-quote-mcp-server [MCP-GO] ← HTTP Proxy:**
```json
{
  "usdc_amount": "0.030000",
  "cirx_fee": 4.0,
  "breakdown": {
    "cirx_cost_usd": 0.0176,
    "service_margin_usd": 0.01144,
    "margin_percent": 65.0,
    "total_usd": 0.02904,
    "rounded_usdc": 0.03
  },
  "valid_until": "2025-10-31T14:35:00Z"
}
```

---

## Phase 3: Payment Requirement Generation

### Tool Call 4: Create payment requirement (x402 protocol)

**HTTP Proxy [MCP-GO] → x402-mcp-server.create_payment_requirement**
```json
{
  "tool": "create_payment_requirement",
  "arguments": {
    "amount_usdc": "0.030000",
    "network": "base",
    "pay_to": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "resource": "https://certify.ar4s.com/v1/certify",
    "description": "Certify ashley-barr_offerletter-contract--unsigned.pdf"
  }
}
```

**x402-mcp-server {Nonce Generation}:**
```go
// Query blockchain for wallet nonce (prevents replay attacks)
rpcClient := s.getEthClient("base")
nonce, err := rpcClient.PendingNonceAt(context.Background(), payeeAddress)
```

**x402-mcp-server [External] → Base RPC Endpoint**
```
POST https://mainnet.base.org
{
  "jsonrpc": "2.0",
  "method": "eth_getTransactionCount",
  "params": ["0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", "pending"],
  "id": 1
}
```

**x402-mcp-server [External] ← Base RPC:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": "0x2a"
}
```

**x402-mcp-server {x402 JSON Construction}:**
```go
paymentReq := X402PaymentRequirement{
    X402Version:       1,
    Scheme:            "exact",
    Network:           "base",
    MaxAmountRequired: "30000", // 0.03 USDC in atomic units (6 decimals)
    Asset:             "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", // USDC on Base
    PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    Nonce:             fmt.Sprintf("0x%x", nonce),
    ValidUntil:        time.Now().Add(5 * time.Minute).Unix(),
    Resource:          "https://certify.ar4s.com/v1/certify",
    Description:       "Certify ashley-barr_offerletter-contract--unsigned.pdf",
}
```

**x402-mcp-server [MCP-GO] ← HTTP Proxy:**
```json
{
  "x402_version": 1,
  "scheme": "exact",
  "network": "base",
  "maxAmountRequired": "30000",
  "asset": "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
  "payTo": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "nonce": "0x2a",
  "valid_until": 1730388900,
  "resource": "https://certify.ar4s.com/v1/certify",
  "description": "Certify ashley-barr_offerletter-contract--unsigned.pdf"
}
```

### HTTP Proxy returns 402 Payment Required

**HTTP Proxy {Store Request in DB}:**
```sql
INSERT INTO certification_requests
  (request_id, client_id, data_hash, data_size_bytes, status, quote_usdc, cirx_price_at_quote)
VALUES
  ('req_offerletter_20251031_abc123', 'api_key_client_1', '0xsha256...', 45678, 'payment_pending', 0.03, 0.0044);
```

**LLM/Agent [HTTP] ← 402 Payment Required**
```json
{
  "error": "payment_required",
  "request_id": "req_offerletter_20251031_abc123",
  "quote": {
    "usdc_amount": "0.030000",
    "cirx_fee": 4.0,
    "cirx_price_usd": 0.0044,
    "valid_until": "2025-10-31T14:35:00Z"
  },
  "payment": {
    "x402_version": 1,
    "scheme": "exact",
    "network": "base",
    "maxAmountRequired": "30000",
    "asset": "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
    "payTo": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "nonce": "0x2a",
    "valid_until": 1730388900,
    "resource": "https://certify.ar4s.com/v1/certify",
    "description": "Certify ashley-barr_offerletter-contract--unsigned.pdf"
  },
  "instructions": {
    "message": "Payment required to proceed. Sign an EIP-3009 authorization and retry with X-PAYMENT header.",
    "eip3009_fields": [
      "from: your_wallet_address",
      "to: 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
      "value: 30000",
      "validAfter: 0",
      "validBefore: 1730388900",
      "nonce: 0x2a",
      "signature: <sign with your private key>"
    ]
  }
}
```

---

## Phase 4: Agent Discovers Payment Requirement & Limitation

**LLM/Agent {Attempts to Pay}:**
```javascript
// LLM/Agent tries to construct EIP-3009 authorization
// Realizes: NO access to private keys, NO wallet, NO native signing capability
```

**LLM/Agent → User:**
```
"I've received a quote for certifying your document: $0.03 USDC (0.03 USDC processing fee +
4 CIRX blockchain fee at current rate of $0.0044/CIRX).

However, I need payment to proceed. I cannot make payments directly because I don't have
access to a wallet or private keys. Here are your options:

Option 1 (Agent Workflow - Programmatic):
  - You'll need to sign an EIP-3009 payment authorization using your wallet
  - Required signature fields are in the response above
  - Once signed, I can submit it on your behalf

Option 2 (Browser Workflow - MetaMask):
  - I can generate a MetaMask deep link for you
  - Click the link, approve the payment in MetaMask
  - Your browser will automatically retry the request with payment proof

Option 3 (Mobile Workflow - QR Code):
  - I can generate a QR code for mobile wallet payment
  - Scan with your mobile wallet (MetaMask, Rainbow, Coinbase Wallet)
  - Your wallet will make the callback with payment proof

Which option would you prefer?"
```

**User → LLM/Agent:**
```
"Let's use Option 1. I have a wallet with private keys. How do I sign the authorization?"
```

**LLM/Agent → User:**
```
"You'll need to sign an EIP-3009 'TransferWithAuthorization' message with these parameters:

from: <your_wallet_address>
to: 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
value: 30000
validAfter: 0
validBefore: 1730388900
nonce: 0x2a

Use EIP-712 typed data signing. Once you provide the signature (v, r, s components),
I'll submit it to complete the certification."
```

**User → LLM/Agent:**
```
"Here's my signed authorization:
from: 0xUserWalletAddress123...
v: 28
r: 0xabcdef1234567890...
s: 0x1234567890abcdef..."
```

---

## Phase 5: Payment Verification & Settlement

**LLM/Agent [HTTP] → POST https://certify.ar4s.com/v1/certify**
```
Headers:
  X-PAYMENT: base64_encoded_eip3009_authorization

Body:
{
  "request_id": "req_offerletter_20251031_abc123"
}
```

**HTTP Proxy {Decode X-PAYMENT header}:**
```go
paymentHeader := r.Header.Get("X-PAYMENT")
decodedPayment, err := base64.StdEncoding.DecodeString(paymentHeader)
// Parse EIP-3009 authorization fields
```

### Tool Call 5: Verify payment

**HTTP Proxy [MCP-GO] → x402-mcp-server.verify_payment**
```json
{
  "tool": "verify_payment",
  "arguments": {
    "payment_header": "base64_encoded_eip3009_authorization",
    "payment_requirements": {
      "maxAmountRequired": "30000",
      "payTo": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
      "valid_until": 1730388900,
      "nonce": "0x2a"
    }
  }
}
```

**x402-mcp-server {EIP-3009 Verification}:**
```go
// 1. Validate amount matches
if authorization.Value != requirements.MaxAmountRequired {
    return VerifyResult{IsValid: false, Error: "amount_mismatch"}
}

// 2. Validate recipient
if authorization.To != requirements.PayTo {
    return VerifyResult{IsValid: false, Error: "recipient_mismatch"}
}

// 3. Validate time window
now := time.Now().Unix()
if now < authorization.ValidAfter || now >= authorization.ValidBefore {
    return VerifyResult{IsValid: false, Error: "expired_authorization"}
}

// 4. Check nonce uniqueness (prevent replay attacks)
if s.db.NonceExists(authorization.Nonce) {
    return VerifyResult{IsValid: false, Error: "duplicate_nonce"}
}

// 5. Recover signer from EIP-712 signature
domainSeparator := eip712.TypedDataDomain{
    Name:              "USD Coin",
    Version:           "2",
    ChainId:           math.NewHexOrDecimal256(8453), // Base mainnet
    VerifyingContract: common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"),
}
signerAddress, err := recoverEIP712Signer(authorization, domainSeparator)

// 6. Verify signer matches 'from' address
if signerAddress != authorization.From {
    return VerifyResult{IsValid: false, Error: "signer_mismatch"}
}
```

**x402-mcp-server [MCP-GO] ← HTTP Proxy:**
```json
{
  "is_valid": true,
  "signer_address": "0xUserWalletAddress123...",
  "verified_at": "2025-10-31T14:32:15Z"
}
```

### Tool Call 6: Settle payment via x402 facilitator

**HTTP Proxy [MCP-GO] → x402-mcp-server.settle_payment**
```json
{
  "tool": "settle_payment",
  "arguments": {
    "payment_header": "base64_encoded_eip3009_authorization",
    "facilitator_url": "https://x402.org/facilitator",
    "network": "base"
  }
}
```

**x402-mcp-server [External] → x402 Facilitator API**
```
POST https://x402.org/facilitator/settle
{
  "network": "base",
  "token": "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
  "authorization": {
    "from": "0xUserWalletAddress123...",
    "to": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "value": "30000",
    "validAfter": "0",
    "validBefore": "1730388900",
    "nonce": "0x2a",
    "v": 28,
    "r": "0xabcdef1234567890...",
    "s": "0x1234567890abcdef..."
  }
}
```

**x402-mcp-server [External] ← Facilitator API:**
```json
{
  "status": "settled",
  "tx_hash": "0xevm_transaction_hash_abc123...",
  "block_number": 12345678,
  "settled_at": "2025-10-31T14:32:18Z"
}
```

**x402-mcp-server {Cache Settlement Result}:**
```go
// Store in-memory cache for idempotency (TTL: 10 minutes)
s.settlementCache.Set(authorization.Nonce, SettlementResult{
    Status:      "settled",
    TxHash:      "0xevm_transaction_hash_abc123...",
    BlockNumber: 12345678,
}, 10*time.Minute)
```

**x402-mcp-server [MCP-GO] ← HTTP Proxy:**
```json
{
  "status": "settled",
  "tx_hash": "0xevm_transaction_hash_abc123...",
  "block_number": 12345678,
  "network": "base",
  "settled_at": "2025-10-31T14:32:18Z"
}
```

**HTTP Proxy {Update DB}:**
```sql
INSERT INTO payments
  (request_id, payment_nonce, from_address, to_address, amount_usdc, network, evm_tx_hash, status)
VALUES
  ('req_offerletter_20251031_abc123', '0x2a', '0xUserWallet...', '0x742d35...', 0.03, 'base', '0xevmtx...', 'settled');

UPDATE certification_requests
SET status = 'payment_verified'
WHERE request_id = 'req_offerletter_20251031_abc123';
```

---

## Phase 6: Blockchain Certification via Circular Protocol Enterprise APIs

**HTTP Proxy {Initiate Certification Workflow}:**
```go
// Payment successful, now certify on Circular Protocol blockchain
circularClient := s.mcpClients["circular-protocol-mcp"]
```

### Tool Call 7: Get wallet nonce

**HTTP Proxy [MCP-GO] → circular-protocol-mcp-server.get_wallet_nonce**
```json
{
  "tool": "get_wallet_nonce",
  "arguments": {
    "wallet_address": "service_wallet_address_64_char_hex"
  }
}
```

**circular-protocol-mcp-server [External] → Circular Protocol Enterprise API**
```
POST https://circular-protocol-api.example.com/Circular_GetWalletNonce_
{
  "Blockchain": "CEP",
  "WalletPublicKey": "service_wallet_address_64_char_hex"
}
```

**circular-protocol-mcp-server [External] ← Circular Enterprise API:**
```json
{
  "Status": "Success",
  "Nonce": 42
}
```

**circular-protocol-mcp-server [MCP-GO] ← HTTP Proxy:**
```json
{
  "nonce": 42,
  "wallet_address": "service_wallet_address_64_char_hex"
}
```

### HTTP Proxy signs transaction locally

**HTTP Proxy {Local Transaction Signing}:**
```go
// Construct certification transaction
timestamp := time.Now().Unix()
txPayload := map[string]interface{}{
    "Action": "CP_CERTIFICATE",
    "Data":   hex.EncodeToString(dataHash), // SHA256 of document
}
payloadJSON, _ := json.Marshal(txPayload)

// Calculate transaction ID client-side (Circular Protocol pattern)
// txID = SHA256(Blockchain + From + To + Payload + Nonce + Timestamp)
txIDInput := fmt.Sprintf("CEP%s%s%s%d%d",
    serviceWalletAddress,
    serviceWalletAddress, // To = From for certificate
    string(payloadJSON),
    nonce,
    timestamp,
)
txID := sha256.Sum256([]byte(txIDInput))

// Sign transaction ID with Secp256k1
privateKey := s.loadPrivateKey("CIRCULAR_CEP_MAINNET_PRIVATE_KEY")
signature, err := crypto.Sign(txID[:], privateKey)
```

### Tool Call 8: Certify data

**HTTP Proxy [MCP-GO] → circular-protocol-mcp-server.certify_data**
```json
{
  "tool": "certify_data",
  "arguments": {
    "data_hash": "0xsha256_of_document...",
    "from_wallet": "service_wallet_address_64_char_hex",
    "to_wallet": "service_wallet_address_64_char_hex",
    "nonce": 42,
    "timestamp": 1730388738,
    "signature": "hex_encoded_secp256k1_signature",
    "public_key": "service_wallet_public_key"
  }
}
```

**circular-protocol-mcp-server {Construct Transaction}:**
```go
tx := CircularTransaction{
    Blockchain:      "CEP",
    Type:            "C_TYPE_CERTIFICATE",
    From:            args.FromWallet,
    To:              args.ToWallet,
    Payload:         fmt.Sprintf(`{"Action":"CP_CERTIFICATE","Data":"%s"}`, args.DataHash),
    Nonce:           args.Nonce,
    Timestamp:       args.Timestamp,
    Signature:       args.Signature,
    SenderPublicKey: args.PublicKey,
}
```

**circular-protocol-mcp-server [External] → Circular Protocol Enterprise API**
```
POST https://circular-protocol-api.example.com/Circular_AddTransaction_
{
  "Blockchain": "CEP",
  "Type": "C_TYPE_CERTIFICATE",
  "From": "service_wallet_address_64_char_hex",
  "To": "service_wallet_address_64_char_hex",
  "Payload": "{\"Action\":\"CP_CERTIFICATE\",\"Data\":\"0xsha256...\"}",
  "Nonce": 42,
  "Timestamp": 1730388738,
  "Signature": "hex_encoded_secp256k1_signature",
  "SenderPublicKey": "service_wallet_public_key"
}
```

**circular-protocol-mcp-server [External] ← Circular Enterprise API:**
```json
{
  "Status": "Success",
  "TransactionID": "calculated_sha256_tx_id_hex",
  "Message": "Transaction submitted successfully"
}
```

**circular-protocol-mcp-server [MCP-GO] ← HTTP Proxy:**
```json
{
  "tx_id": "calculated_sha256_tx_id_hex",
  "status": "submitted",
  "submitted_at": "2025-10-31T14:32:25Z"
}
```

**HTTP Proxy {Update DB}:**
```sql
INSERT INTO certifications
  (request_id, cirx_tx_id, cirx_fee_paid, status)
VALUES
  ('req_offerletter_20251031_abc123', 'calc_sha256_tx_id...', 4.0, 'pending');

UPDATE certification_requests
SET status = 'certifying'
WHERE request_id = 'req_offerletter_20251031_abc123';
```

### Tool Call 9: Poll transaction status

**HTTP Proxy [MCP-GO] → circular-protocol-mcp-server.get_transaction_status**
```json
{
  "tool": "get_transaction_status",
  "arguments": {
    "tx_id": "calculated_sha256_tx_id_hex"
  }
}
```

**circular-protocol-mcp-server [External] → Circular Protocol Enterprise API**
```
POST https://circular-protocol-api.example.com/Circular_GetTransaction_
{
  "Blockchain": "CEP",
  "TransactionID": "calculated_sha256_tx_id_hex"
}
```

**circular-protocol-mcp-server [External] ← Circular Enterprise API (1st poll):**
```json
{
  "Status": "Success",
  "Transaction": {
    "ID": "calculated_sha256_tx_id_hex",
    "Status": "Pending",
    "From": "service_wallet...",
    "To": "service_wallet...",
    "Timestamp": 1730388738
  }
}
```

**circular-protocol-mcp-server [MCP-GO] ← HTTP Proxy:**
```json
{
  "tx_id": "calculated_sha256_tx_id_hex",
  "status": "Pending",
  "confirmed": false
}
```

**HTTP Proxy {Wait 5 seconds, poll again}:**
```go
time.Sleep(5 * time.Second)
// Retry get_transaction_status
```

**HTTP Proxy [MCP-GO] → circular-protocol-mcp-server.get_transaction_status**
```json
{
  "tool": "get_transaction_status",
  "arguments": {
    "tx_id": "calculated_sha256_tx_id_hex"
  }
}
```

**circular-protocol-mcp-server [External] → Circular Protocol Enterprise API**
```
POST https://circular-protocol-api.example.com/Circular_GetTransaction_
{
  "Blockchain": "CEP",
  "TransactionID": "calculated_sha256_tx_id_hex"
}
```

**circular-protocol-mcp-server [External] ← Circular Enterprise API (2nd poll):**
```json
{
  "Status": "Success",
  "Transaction": {
    "ID": "calculated_sha256_tx_id_hex",
    "Status": "Verified",
    "From": "service_wallet...",
    "To": "service_wallet...",
    "Timestamp": 1730388738
  }
}
```

**circular-protocol-mcp-server [MCP-GO] ← HTTP Proxy:**
```json
{
  "tx_id": "calculated_sha256_tx_id_hex",
  "status": "Verified",
  "confirmed": false
}
```

**HTTP Proxy {Wait 5 seconds, poll again}:**

**HTTP Proxy [MCP-GO] → circular-protocol-mcp-server.get_transaction_status**
```json
{
  "tool": "get_transaction_status",
  "arguments": {
    "tx_id": "calculated_sha256_tx_id_hex"
  }
}
```

**circular-protocol-mcp-server [External] → Circular Protocol Enterprise API**
```
POST https://circular-protocol-api.example.com/Circular_GetTransaction_
{
  "Blockchain": "CEP",
  "TransactionID": "calculated_sha256_tx_id_hex"
}
```

**circular-protocol-mcp-server [External] ← Circular Enterprise API (3rd poll):**
```json
{
  "Status": "Success",
  "Transaction": {
    "ID": "calculated_sha256_tx_id_hex",
    "Status": "Executed",
    "BlockID": "block_987654",
    "From": "service_wallet...",
    "To": "service_wallet...",
    "Timestamp": 1730388738,
    "Payload": "{\"Action\":\"CP_CERTIFICATE\",\"Data\":\"0xsha256...\"}"
  }
}
```

**circular-protocol-mcp-server [MCP-GO] ← HTTP Proxy:**
```json
{
  "tx_id": "calculated_sha256_tx_id_hex",
  "status": "Executed",
  "confirmed": true,
  "block_id": "block_987654"
}
```

### Tool Call 10: Get certification proof

**HTTP Proxy [MCP-GO] → circular-protocol-mcp-server.get_certification_proof**
```json
{
  "tool": "get_certification_proof",
  "arguments": {
    "tx_id": "calculated_sha256_tx_id_hex"
  }
}
```

**circular-protocol-mcp-server {Generate Proof}:**
```go
// Extract proof details from executed transaction
proof := CertificationProof{
    TransactionID:  txDetails.ID,
    BlockID:        txDetails.BlockID,
    Timestamp:      time.Unix(txDetails.Timestamp, 0).Format(time.RFC3339),
    DataHash:       extractDataHashFromPayload(txDetails.Payload),
    ExplorerURL:    fmt.Sprintf("https://explorer.circular-protocol.org/tx/%s", txDetails.ID),
    Network:        "mainnet",
}
```

**circular-protocol-mcp-server [MCP-GO] ← HTTP Proxy:**
```json
{
  "tx_id": "calculated_sha256_tx_id_hex",
  "block_id": "block_987654",
  "timestamp": "2025-10-31T14:32:43Z",
  "data_hash": "0xsha256_of_document...",
  "explorer_url": "https://explorer.circular-protocol.org/tx/calculated_sha256_tx_id_hex",
  "network": "mainnet"
}
```

**HTTP Proxy {Update DB - Final}:**
```sql
UPDATE certifications
SET status = 'confirmed', cirx_block_id = 'block_987654', confirmed_at = NOW()
WHERE request_id = 'req_offerletter_20251031_abc123';

UPDATE certification_requests
SET status = 'completed'
WHERE request_id = 'req_offerletter_20251031_abc123';
```

---

## Phase 7: Final Response to Agent/User

**LLM/Agent [HTTP] ← 200 OK**
```json
{
  "request_id": "req_offerletter_20251031_abc123",
  "status": "completed",
  "payment": {
    "tx_hash": "0xevm_transaction_hash_abc123...",
    "network": "base",
    "amount_paid_usdc": "0.030000",
    "settled_at": "2025-10-31T14:32:18Z",
    "block_number": 12345678
  },
  "certification": {
    "tx_id": "calculated_sha256_tx_id_hex",
    "block_id": "block_987654",
    "timestamp": "2025-10-31T14:32:43Z",
    "data_hash": "0xsha256_of_document...",
    "cirx_fee_paid": 4.0,
    "verification_url": "https://explorer.circular-protocol.org/tx/calculated_sha256_tx_id_hex",
    "network": "mainnet"
  },
  "completed_at": "2025-10-31T14:32:43Z",
  "total_duration_seconds": 28
}
```

**LLM/Agent → User:**
```
"Done! Your document 'ashley-barr_offerletter-contract--unsigned.pdf' has been successfully
certified on the Circular Protocol blockchain.

Certification Details:
- Payment: 0.03 USDC settled on Base (tx: 0xevmtx...abc123)
- Blockchain Fee: 4 CIRX
- Certification Transaction: calculated_sha256_tx_id_hex
- Block: block_987654
- Timestamp: 2025-10-31 14:32:43 UTC
- Total Time: 28 seconds

You can verify your certification here:
https://explorer.circular-protocol.org/tx/calculated_sha256_tx_id_hex

This proof is now permanently recorded on the blockchain and can be independently verified
by anyone."
```

**User → LLM/Agent:**
```
"Perfect, thank you!"
```

---

## Summary: MCP Architecture & Tool Usage

### 4 MCP Servers Connected via mcp-go (stdio transport):

1. **data-quote-mcp-server** (3 tools):
   - `check_data_size` - Calculate byte size of data
   - `get_cirx_price` - Fetch CIRX/USD from CoinGecko (cached 5 min)
   - `calculate_quote` - Compute USDC quote with margin

2. **x402-mcp-server** (5 tools):
   - `create_payment_requirement` - Generate x402 JSON with blockchain nonce
   - `verify_payment` - EIP-3009 signature verification (Secp256k1 recovery)
   - `settle_payment` - Call x402 facilitator API for on-chain settlement
   - `generate_browser_link` - MetaMask deep links (not used in agent workflow)
   - `encode_payment_for_qr` - EIP-681 format (not used in agent workflow)

3. **circular-protocol-mcp-server** (4 tools):
   - `get_wallet_nonce` - Query Circular_GetWalletNonce_ Enterprise API
   - `certify_data` - Submit C_TYPE_CERTIFICATE via Circular_AddTransaction_
   - `get_transaction_status` - Poll Circular_GetTransaction_ until "Executed"
   - `get_certification_proof` - Extract block ID, timestamp, generate explorer URL

4. **qr-code-mcp-server** (3 tools - unused in agent workflow):
   - `generate_qr_ascii` - ASCII art QR for terminal
   - `generate_qr_image` - PNG/SVG QR generation
   - `encode_x402_to_qr` - Format x402 for QR scanning

### mcp-go Client Logic (HTTP Proxy):

```go
// proxy/internal/mcp/client.go
type MCPClientManager struct {
    clients map[string]*mcpgoclient.StdioMCPClient
}

func (m *MCPClientManager) Initialize() error {
    // Start 4 MCP servers as subprocesses (stdio transport)
    m.clients["data-quote-mcp"] = mcpgoclient.NewStdioClient("./mcp-servers/data-quote-mcp-server/main")
    m.clients["x402-mcp"] = mcpgoclient.NewStdioClient("./mcp-servers/x402-mcp-server/main")
    m.clients["circular-protocol-mcp"] = mcpgoclient.NewStdioClient("./mcp-servers/circular-protocol-mcp-server/main")
    m.clients["qr-code-mcp"] = mcpgoclient.NewStdioClient("./mcp-servers/qr-code-mcp-server/main")

    // Discover tools from each server
    for name, client := range m.clients {
        tools, err := client.ListTools()
        log.Infof("MCP server %s: discovered %d tools", name, len(tools))
    }

    return nil
}

func (m *MCPClientManager) CallTool(serverName, toolName string, args map[string]interface{}) (interface{}, error) {
    client := m.clients[serverName]
    result, err := client.CallTool(toolName, args)
    return result, err
}
```

### Total Tool Calls in Complete Flow:
1. data-quote-mcp.check_data_size
2. data-quote-mcp.get_cirx_price
3. data-quote-mcp.calculate_quote
4. x402-mcp.create_payment_requirement
5. x402-mcp.verify_payment
6. x402-mcp.settle_payment
7. circular-protocol-mcp.get_wallet_nonce
8. circular-protocol-mcp.certify_data
9. circular-protocol-mcp.get_transaction_status (3x polling)
10. circular-protocol-mcp.get_certification_proof

**Total: 13 MCP tool invocations orchestrated by mcp-go client**
