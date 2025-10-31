# certify.ar4s.com - Blockchain Certification Platform Specification

**Version:** 1.0
**Status:** Draft
**Date:** 2025-10-28
**Methodology:** Kiro Spec-Driven Development

---

## Executive Summary

**certify.ar4s.com** is an MCP Host proxy service that enables ANY LLM (regardless of native MCP support) to certify data on the Circular Protocol blockchain using Enterprise APIs (Go implementation, see docs/GO-CEP-APIS.xml) with automatic payment handling via x402. The service acts as a universal gateway, supporting three distinct user workflows: AI agents (programmatic), browser users (wallet extensions), and mobile users (QR codes).

### Core Value Proposition
Transform any LLM into a blockchain-capable entity without requiring:
- Native MCP protocol support
- Direct blockchain integration knowledge
- Manual payment processing
- Cryptocurrency wallet management expertise

### Architecture Overview
```
Any LLM (GPT-4, Gemini, Claude, etc.)
    ↓ HTTP/REST API
┌────────────────────────────────────────────────────┐
│  certify.ar4s.com (MCP Host Proxy)                 │
│  - HTTP gateway for non-MCP LLMs                   │
│  - Orchestrates payment → certification workflow   │
│  - Built with: mcp-go client library               │
└──────┬──────────┬──────────┬──────────────────────┘
       │          │          │
       │ MCP      │ MCP      │ MCP
       │          │          │
┌──────▼────┐ ┌──▼────┐ ┌───▼──────┐ ┌─▼────────────┐
│ x402-mcp  │ │circular│ │data-quote│ │qr-code-mcp   │
│ server    │ │-protocol│ │-mcp      │ │server        │
│           │ │-mcp     │ │server    │ │              │
│[5 tools]  │ │server   │ │          │ │[3 tools]     │
│           │ │[4 tools]│ │[3 tools] │ │              │
└───────────┘ └─────────┘ └──────────┘ └──────────────┘
```

---

## Phase 1: Requirements

### 1.1 User Stories

#### US-001: AI Agent Requests Certification Quote
**As an** AI agent (LLM like GPT-4, Claude, Gemini)
**I want to** request a pricing quote for certifying my data
**So that** I can understand costs before initiating payment

**Acceptance Criteria:**
- [ ] Agent can POST data to `/v1/quote` endpoint
- [ ] Response includes USDC cost, CIRX fee breakdown, data size
- [ ] Quote includes expiration timestamp (5 minute validity)
- [ ] Quote ID is returned for tracking
- [ ] Supported networks are listed (base, base-sepolia, arbitrum)

---

#### US-002: AI Agent Pays and Certifies Data Programmatically
**As an** AI agent with programmatic wallet access
**I want to** pay for certification and receive proof automatically
**So that** I can integrate blockchain certification into my workflows without human intervention

**Acceptance Criteria:**
- [ ] Agent receives 402 Payment Required with x402 payment requirements
- [ ] Agent can sign EIP-3009 authorization using private key
- [ ] Agent retries POST `/v1/certify` with X-PAYMENT header
- [ ] Payment is verified and settled on-chain within 2 seconds
- [ ] Certification is submitted to Circular Protocol via Enterprise APIs within 5 seconds
- [ ] Response includes certification proof (tx hash, block number, timestamp)
- [ ] All operations complete in < 10 seconds (happy path)

---

#### US-003: Agent Checks Certification Status
**As an** AI agent
**I want to** check the status of my certification request
**So that** I can handle async workflows and retry failures

**Acceptance Criteria:**
- [ ] Agent can GET `/v1/status/{request_id}`
- [ ] Response includes current status (payment_pending, certifying, completed, failed)
- [ ] If completed: certification proof is included
- [ ] If failed: error reason and retry eligibility are provided
- [ ] Status updates in real-time as state transitions occur

---

#### US-004: Browser User Initiates Certification
**As a** developer using an LLM interface in a browser
**I want to** initiate certification for data my LLM generated
**So that** I can prove data provenance without leaving my browser

**Acceptance Criteria:**
- [ ] LLM makes POST `/v1/certify` with `client_type: "browser"`
- [ ] 402 response includes MetaMask deep link
- [ ] Deep link pre-fills payment parameters
- [ ] User clicks link → MetaMask opens → user approves
- [ ] Browser automatically retries request with X-PAYMENT header
- [ ] Certification completes and proof is displayed in browser

---

#### US-005: Mobile User Scans QR Code to Pay
**As a** user with a mobile wallet app
**I want to** scan a QR code to pay for certification
**So that** I can use my mobile wallet without typing addresses or amounts

**Acceptance Criteria:**
- [ ] LLM makes POST `/v1/certify` with `client_type: "mobile"`
- [ ] 402 response includes ASCII QR code for terminal display
- [ ] QR code encodes payment requirements in EIP-681 format
- [ ] User scans QR with Rainbow/Coinbase Wallet/MetaMask mobile
- [ ] Wallet parses payment data and prompts user
- [ ] After approval, wallet POST to `/v1/pay` with X-PAYMENT
- [ ] Certification completes and result available via status endpoint

---

#### US-006: Operator Monitors CIRX Wallet Balance
**As a** service operator
**I want to** receive alerts when CIRX balance is low
**So that** I can replenish funds before service interruption occurs

**Acceptance Criteria:**
- [ ] System monitors service wallet balance every 5 minutes
- [ ] Alert triggers when balance < 100 CIRX (25 certifications remaining)
- [ ] Alert includes current balance, daily burn rate, estimated time until depletion
- [ ] Dashboard displays real-time CIRX balance
- [ ] Operator can manually trigger balance check

---

#### US-007: Operator Handles Failed Certifications
**As a** service operator
**I want to** view and retry failed certifications
**So that** users who paid receive their certification even after transient failures

**Acceptance Criteria:**
- [ ] Dashboard lists all failed certifications with error details
- [ ] Failed certifications marked "paid_not_certified" are auto-retried
- [ ] Retry queue uses exponential backoff (max 10 attempts over 24h)
- [ ] Operator can manually retry from dashboard
- [ ] Dead letter queue captures permanently failed items
- [ ] Refund workflow initiated for unrecoverable failures

---

### 1.2 Functional Requirements

#### FR-001: HTTP API Gateway
- Accept HTTP POST/GET requests from any HTTP client
- Support JSON request/response bodies
- Implement CORS for browser clients
- Return standard HTTP status codes
- Support idempotency via `request_id` parameter

#### FR-002: x402 Payment Protocol Integration
- Generate x402 payment requirements conforming to spec
- Validate EIP-3009 payment authorizations
- Verify signature matches authorization data
- Settle payments via x402 facilitator server
- Handle payment timeouts and retries

#### FR-003: MCP Client Operations (mcp-go)
- Connect to 4 MCP servers via stdio transport
- Discover tools dynamically on startup
- Execute MCP tool calls with proper request/response formatting
- Handle MCP server disconnections and reconnects
- Implement connection pooling for performance

#### FR-004: Circular Protocol Enterprise API Integration
- Use Circular Protocol Enterprise APIs (Go implementation pattern from docs/GO-CEP-APIS.xml)
- Maintain hot wallet for signing certification transactions
- Fetch wallet nonce before each transaction via Enterprise API endpoint
- Sign transactions using Secp256k1 (matches Circular Protocol Enterprise APIs)
- Calculate transaction ID client-side: SHA-256(Blockchain+From+To+Payload+Nonce+Timestamp)
- Submit certification transactions with Type="C_TYPE_CERTIFICATE" via Enterprise API
- Poll transaction status until "Executed" or timeout
- Pay 4 CIRX flat fee per certification

#### FR-005: Multi-Workflow Orchestration
- Coordinate: quote → payment → certification flow
- Support agent, browser, and mobile payment workflows
- Generate payment links for browser users
- Generate QR codes for mobile users
- Implement state machine: INITIATED → QUOTED → PAYMENT_PENDING → PAYMENT_VERIFIED → CERTIFYING → COMPLETED/FAILED

#### FR-006: Data and Pricing Management
- Calculate data size from input (base64/hex/plaintext)
- Fetch current CIRX/USD price from CoinGecko
- Calculate USDC quote: (4 CIRX × CIRX_price_USD) × (1 + margin%)
- Default margin: 65%
- Cache price data for 5 minutes to reduce API calls

#### FR-007: QR Code Generation
- Generate ASCII QR codes for terminal display using block characters (█)
- Generate PNG/SVG QR code images
- Encode x402 payment data in EIP-681 format for wallet compatibility
- Include callback URL in QR data for mobile wallet POST

#### FR-008: Error Handling and Retry Logic
- Implement retry queue for failed certifications
- Use exponential backoff (5s, 10s, 20s, 40s, ..., max 60s)
- Max retries: 10 attempts per certification
- Store failed items in dead letter queue after max retries
- Alert operators on high failure rates (>5%)

---

### 1.3 Non-Functional Requirements

#### NFR-001: Performance
- **Latency:** Total certification flow < 10 seconds (P95)
  - Payment verification: < 1 second
  - Payment settlement (via facilitator): < 2 seconds
  - Certification submission: < 3 seconds
  - Certification confirmation: < 4 seconds
- **Throughput:** Support 100 concurrent certification requests
- **Availability:** 99.5% uptime SLA

#### NFR-002: Security
- **Wallet Security:** Service wallet private key encrypted at rest (Vault/KMS)
- **Payment Validation:** Strict EIP-3009 signature verification
  - Validate: signature, nonce uniqueness, validBefore/validAfter, amounts
- **API Authentication:** All endpoints require X-API-Key header
- **Rate Limiting:** 10 requests/minute per API key (configurable)
- **Input Validation:** Max data size 10MB, sanitize all inputs
- **CORS:** Whitelist-based origin validation

#### NFR-003: Reliability
- **Idempotency:** Duplicate `request_id` returns cached result (no double-charge)
- **Retry Logic:** Automatic retry for transient Circular Protocol failures
- **Graceful Degradation:** Queue requests if blockchain temporarily unavailable
- **Data Persistence:** All requests/payments/certifications logged in PostgreSQL
- **Monitoring:** Real-time alerts for failures, low balance, high latency

#### NFR-004: Observability
- **Logging:** Structured JSON logs (Zap library)
- **Metrics:** Prometheus metrics
  - Certification success rate
  - Payment verification rate
  - Average certification time
  - CIRX wallet balance
  - HTTP request latency
- **Tracing:** OpenTelemetry distributed tracing
- **Dashboards:** Grafana dashboard for operators

#### NFR-005: Scalability
- **Horizontal Scaling:** Stateless service design
- **Database:** PostgreSQL with connection pooling (max 50 connections)
- **Caching:** Redis for CIRX price, payment status, quote data
- **Load Balancing:** Deploy behind Cloudflare or AWS ALB

---

## Phase 2: Design

### 2.1 System Architecture

#### 2.1.1 Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    External Clients                             │
│  - LLMs (GPT-4, Gemini, Claude)                                 │
│  - Browser applications (JavaScript)                            │
│  - Mobile wallet apps (Rainbow, Coinbase Wallet)                │
└────────────────────────────┬────────────────────────────────────┘
                             │ HTTPS
┌────────────────────────────▼────────────────────────────────────┐
│            Load Balancer (Cloudflare / AWS ALB)                 │
└────────────────────────────┬────────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────────┐
│           certify.ar4s.com (MCP Host Proxy)                     │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              HTTP API Layer (Gin/Echo)                    │  │
│  │  - POST /v1/quote                                         │  │
│  │  - POST /v1/certify                                       │  │
│  │  - GET /v1/status/:id                                     │  │
│  │  - GET /v1/qr/:id                                         │  │
│  │  Middleware: Auth, CORS, Rate Limiting                    │  │
│  └────────────────┬─────────────────────────────────────────┘  │
│                   │                                             │
│  ┌────────────────▼─────────────────────────────────────────┐  │
│  │          Orchestration Layer (State Machine)             │  │
│  │  - Request lifecycle management                          │  │
│  │  - MCP tool call coordination                            │  │
│  │  - Error handling & retry queue                          │  │
│  └────────┬───────────┬───────────┬───────────┬─────────────┘  │
│           │           │           │           │                 │
│  ┌────────▼──┐  ┌─────▼────┐ ┌───▼─────┐ ┌──▼──────────┐     │
│  │MCP Client │  │MCP Client│ │MCP      │ │MCP Client    │     │
│  │(x402)     │  │(circular)│ │Client   │ │(qr-code)     │     │
│  └────────┬──┘  └─────┬────┘ │(quote)  │ └──┬──────────┘     │
│           │           │       └───┬─────┘    │                 │
│  ┌────────▼───────────▼───────────▼──────────▼─────────────┐  │
│  │         Data Persistence Layer                           │  │
│  │  - PostgreSQL (certification_requests, payments, certs)  │  │
│  │  - Redis (cache: prices, quotes, payment status)         │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                             │ MCP Protocol (stdio)
        ┌────────────────────┼────────────────────┐
        │                    │                    │
┌───────▼──────────┐ ┌───────▼───────────────┐ ┌──────▼────────────┐
│ x402-mcp-server  │ │circular-protocol      │ │data-quote-mcp     │
│ (Go + mcp-go)    │ │-enterprise-mcp-server │ │-server            │
│                  │ │(Go + mcp-go)          │ │(Go + mcp-go)      │
│ Tools:           │ │                       │ │                   │
│ -create_payment  │ │Tools:                 │ │Tools:             │
│ -verify_payment  │ │-get_nonce             │ │-check_size        │
│ -settle_payment  │ │-certify_data          │ │-get_cirx_price    │
│ -browser_link    │ │-get_tx_status         │ │-calculate_quote   │
│ -encode_for_qr   │ │-get_proof             │ │                   │
└────────┬─────────┘ └────────┬──────────────┘ └────────┬──────────┘
         │                    │                         │
         │                    │                         │
┌────────▼─────────┐ ┌────────▼──────────┐ ┌───────▼──────────────────┐
│ qr-code-mcp      │ │ x402 Facilitator  │ │ Circular Protocol        │
│ -server          │ │ (x402.org)        │ │ Enterprise APIs          │
│ (Go + mcp-go)    │ │                   │ │ (HTTP REST)              │
│                  │ │ Wraps:            │ │                          │
│ Tools:           │ │ - EIP-3009 calls  │ │ Enterprise Endpoints:    │
│ -generate_ascii  │ │ - Blockchain      │ │ - Circular_AddTransaction│
│ -generate_image  │ │   settlement      │ │ - Circular_GetWalletNonce│
│ -encode_x402     │ │                   │ │ - Circular_GetTransaction│
└──────────────────┘ └───────────────────┘ └──────────────────────────┘
```

---

### 2.2 Sequence Diagrams

#### 2.2.1 Happy Path: Agent Payment Workflow

```
LLM Agent  certify.ar4s.com  data-quote  x402-mcp  circular-ent-mcp  Facilitator  Circular Enterprise API
    │              │              │           │           │               │              │
    │ POST /certify│              │           │           │            │              │
    ├──────────────>              │           │           │            │              │
    │              │ check_size() │           │           │            │              │
    │              ├──────────────>           │           │            │              │
    │              │<──────────────┤          │           │            │              │
    │              │ get_cirx_price()         │           │            │              │
    │              ├──────────────>           │           │            │              │
    │              │<──────────────┤          │           │            │              │
    │              │ calculate_quote()        │           │            │              │
    │              ├──────────────>           │           │            │              │
    │              │<──────────────┤          │           │            │              │
    │              │ create_payment()         │           │            │              │
    │              ├──────────────────────────>           │            │              │
    │              │<──────────────────────────┤          │            │              │
    │ 402 Payment  │              │           │           │            │              │
    │  Required    │              │           │           │            │              │
    │<──────────────              │           │           │            │              │
    │              │              │           │           │            │              │
    │ POST /certify│              │           │           │            │              │
    │ X-PAYMENT    │              │           │           │            │              │
    ├──────────────>              │           │           │            │              │
    │              │ verify_payment()         │           │            │              │
    │              ├──────────────────────────>           │            │              │
    │              │ {is_valid: true}         │           │            │              │
    │              │<──────────────────────────┤          │            │              │
    │              │ settle_payment()         │           │            │              │
    │              ├──────────────────────────>           │            │              │
    │              │              │           │ POST /settle          │              │
    │              │              │           ├────────────────────────>             │
    │              │              │           │ EIP-3009 call          │             │
    │              │              │           │<────────────────────────┤            │
    │              │ {tx_hash: 0x...}        │           │            │              │
    │              │<──────────────────────────┤          │            │              │
    │              │ get_wallet_nonce()       │           │            │              │
    │              ├────────────────────────────────────────>          │              │
    │              │              │           │           │ GET Nonce  │              │
    │              │              │           │           ├────────────────────────────>
    │              │              │           │           │ {nonce: 42}│              │
    │              │              │           │           │<────────────────────────────┤
    │              │ {nonce: 42}  │           │           │            │              │
    │              │<────────────────────────────────────────┤         │              │
    │              │ [Sign tx locally]       │           │            │              │
    │              │ certify_data(hash, sig) │           │            │              │
    │              ├────────────────────────────────────────>          │              │
    │              │              │           │           │ POST AddTx │              │
    │              │              │           │           ├────────────────────────────>
    │              │              │           │           │ {tx_id}    │              │
    │              │              │           │           │<────────────────────────────┤
    │              │ {tx_id: 0xabc}          │           │            │              │
    │              │<────────────────────────────────────────┤         │              │
    │              │ get_tx_status()         │           │            │              │
    │              ├────────────────────────────────────────>          │              │
    │              │              │           │           │ GET Status │              │
    │              │              │           │           ├────────────────────────────>
    │              │              │           │           │{Executed}  │              │
    │              │              │           │           │<────────────────────────────┤
    │              │ {confirmed: true}       │           │            │              │
    │              │<────────────────────────────────────────┤         │              │
    │ 200 OK       │              │           │           │            │              │
    │ {cert proof} │              │           │           │            │              │
    │<──────────────              │           │           │            │              │
```

**Flow Duration:** ~7-10 seconds total
- Steps 1-10 (quote + payment verification): ~2 seconds
- Steps 11-14 (settlement): ~2 seconds
- Steps 15-23 (certification): ~3-5 seconds

---

#### 2.2.2 Browser User Workflow

```
Browser LLM  certify.ar4s.com  x402-mcp  qr-code-mcp
    │              │              │           │
    │ POST /certify│              │           │
    │ client_type: │              │           │
    │  "browser"   │              │           │
    ├──────────────>              │           │
    │              │ [Get quote as in 2.2.1]  │
    │              │ create_payment()         │
    │              ├──────────────>           │
    │              │<──────────────┤          │
    │              │ generate_browser_link()  │
    │              ├──────────────>           │
    │              │ {deep_link}  │           │
    │              │<──────────────┤          │
    │ 402 Payment  │              │           │
    │ Required +   │              │           │
    │ payment_link │              │           │
    │<──────────────              │           │
    │              │              │           │
    │ User clicks link            │           │
    │ ──> MetaMask opens ──>      │           │
    │ ──> User approves ──>       │           │
    │ ──> MetaMask signs ──>      │           │
    │              │              │           │
    │ POST /certify│              │           │
    │ X-PAYMENT    │              │           │
    ├──────────────>              │           │
    │ [Continue as in 2.2.1]      │           │
```

**Key Difference:** After 402 response, browser client handles MetaMask interaction. Payment flow is identical to agent workflow once X-PAYMENT header is generated.

---

#### 2.2.3 Mobile User Workflow

```
Terminal/CLI  certify.ar4s.com  qr-code-mcp  Mobile Wallet
    │              │              │              │
    │ POST /certify│              │              │
    │ client_type: │              │              │
    │  "mobile"    │              │              │
    ├──────────────>              │              │
    │              │ [Get quote]  │              │
    │              │ encode_x402_to_qr()         │
    │              ├──────────────>              │
    │              │ {qr_data}    │              │
    │              │<──────────────┤             │
    │              │ generate_qr_ascii()         │
    │              ├──────────────>              │
    │              │ {ascii_qr}   │              │
    │              │<──────────────┤             │
    │ 402 Payment  │              │              │
    │ + ASCII QR   │              │              │
    │<──────────────              │              │
    │              │              │              │
    │ Display QR:  │              │              │
    │ █████████    │              │              │
    │ ██   ██      │              │              │
    │ (User scans) │              │              │
    │───────────────────────────────────────────>│
    │              │              │ Wallet parses│
    │              │              │ payment data │
    │              │              │ User approves│
    │              │              │ Wallet signs │
    │              │              │              │
    │              │ POST /pay    │              │
    │              │ X-PAYMENT    │              │
    │              │<──────────────────────────────┤
    │ [Continue normal flow]      │              │
```

**Key Difference:** QR code displayed in terminal, mobile wallet makes the callback POST to certify.ar4s.com after user approval.

---

### 2.3 Data Models

#### 2.3.1 Database Schema (PostgreSQL)

**Table: certification_requests**
```sql
CREATE TABLE certification_requests (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id              TEXT UNIQUE NOT NULL,
    client_id               TEXT NOT NULL, -- API key
    client_type             TEXT, -- 'agent', 'browser', 'mobile'
    data_hash               TEXT NOT NULL,
    data_size_bytes         BIGINT NOT NULL,
    status                  TEXT NOT NULL,
        -- 'initiated', 'quoted', 'payment_pending', 'payment_verified',
        -- 'certifying', 'completed', 'failed'
    quote_usdc              NUMERIC(20, 6),
    cirx_price_at_quote     NUMERIC(20, 8),
    cirx_fee                NUMERIC(20, 6) DEFAULT 4.0,
    quote_expires_at        TIMESTAMP,
    callback_url            TEXT,
    created_at              TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_request_id ON certification_requests(request_id);
CREATE INDEX idx_client_id ON certification_requests(client_id);
CREATE INDEX idx_status ON certification_requests(status);
CREATE INDEX idx_created_at ON certification_requests(created_at DESC);
```

**Table: payments**
```sql
CREATE TABLE payments (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id              TEXT NOT NULL REFERENCES certification_requests(request_id),
    payment_nonce           TEXT UNIQUE NOT NULL,
    from_address            TEXT NOT NULL,
    to_address              TEXT NOT NULL,
    amount_usdc             NUMERIC(20, 6) NOT NULL,
    network                 TEXT NOT NULL, -- 'base', 'base-sepolia', 'arbitrum'
    evm_tx_hash             TEXT,
    status                  TEXT NOT NULL,
        -- 'pending', 'verified', 'settled', 'failed'
    payment_requirements    JSONB NOT NULL,
    payment_header          JSONB,
    verified_at             TIMESTAMP,
    settled_at              TIMESTAMP,
    created_at              TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payment_request ON payments(request_id);
CREATE INDEX idx_payment_nonce ON payments(payment_nonce);
CREATE INDEX idx_payment_status ON payments(status);
```

**Table: certifications**
```sql
CREATE TABLE certifications (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id              TEXT NOT NULL REFERENCES certification_requests(request_id),
    cirx_tx_id              TEXT UNIQUE NOT NULL,
    cirx_block_id           TEXT,
    cirx_fee_paid           NUMERIC(20, 6) NOT NULL,
    status                  TEXT NOT NULL,
        -- 'pending', 'confirmed', 'failed'
    retry_count             INT DEFAULT 0,
    last_retry_at           TIMESTAMP,
    error_message           TEXT,
    confirmed_at            TIMESTAMP,
    created_at              TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cert_request ON certifications(request_id);
CREATE INDEX idx_cert_tx ON certifications(cirx_tx_id);
CREATE INDEX idx_cert_status ON certifications(status);
```

**Table: wallet_balances**
```sql
CREATE TABLE wallet_balances (
    id                      SERIAL PRIMARY KEY,
    asset                   TEXT NOT NULL, -- 'CIRX'
    network                 TEXT NOT NULL, -- 'circular-mainnet'
    wallet_address          TEXT NOT NULL,
    balance                 NUMERIC(20, 6) NOT NULL,
    last_updated            TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_asset_network ON wallet_balances(asset, network);
```

---

#### 2.3.2 MCP Tool Definitions

**x402-mcp-server Tools:**

1. `create_payment_requirement`
```json
{
  "name": "create_payment_requirement",
  "description": "Generate x402 payment requirements for a given amount and network",
  "inputSchema": {
    "type": "object",
    "properties": {
      "amount_usdc": {
        "type": "string",
        "description": "USDC amount like '0.05' or '0.050000'",
        "pattern": "^\\d+\\.\\d{1,6}$"
      },
      "network": {
        "type": "string",
        "enum": ["base", "base-sepolia", "arbitrum"],
        "description": "EVM network for payment"
      },
      "pay_to": {
        "type": "string",
        "pattern": "^0x[a-fA-F0-9]{40}$",
        "description": "Recipient wallet address"
      },
      "resource": {
        "type": "string",
        "format": "uri",
        "description": "Full URL of protected resource"
      },
      "description": {
        "type": "string",
        "description": "Human-readable description of purchase"
      }
    },
    "required": ["amount_usdc", "network", "pay_to", "resource"]
  }
}
```

2. `verify_payment`
```json
{
  "name": "verify_payment",
  "description": "Verify EIP-3009 payment authorization signature",
  "inputSchema": {
    "type": "object",
    "properties": {
      "payment_header": {
        "type": "string",
        "description": "Base64-encoded X-PAYMENT header value"
      },
      "payment_requirements": {
        "type": "object",
        "description": "Original payment requirements"
      }
    },
    "required": ["payment_header", "payment_requirements"]
  }
}
```

3. `settle_payment`
```json
{
  "name": "settle_payment",
  "description": "Settle payment on-chain via facilitator",
  "inputSchema": {
    "type": "object",
    "properties": {
      "payment_header": {"type": "string"},
      "payment_requirements": {"type": "object"}
    },
    "required": ["payment_header", "payment_requirements"]
  }
}
```

4. `generate_browser_link`
```json
{
  "name": "generate_browser_link",
  "description": "Generate deep link for browser wallet (MetaMask, Coinbase Wallet)",
  "inputSchema": {
    "type": "object",
    "properties": {
      "payment_requirements": {"type": "object"},
      "wallet_type": {
        "type": "string",
        "enum": ["metamask", "coinbase", "rainbow"],
        "default": "metamask"
      }
    },
    "required": ["payment_requirements"]
  }
}
```

5. `encode_payment_for_qr`
```json
{
  "name": "encode_payment_for_qr",
  "description": "Encode payment data for QR code generation",
  "inputSchema": {
    "type": "object",
    "properties": {
      "payment_requirements": {"type": "object"},
      "callback_url": {
        "type": "string",
        "format": "uri",
        "description": "URL for wallet to POST signed payment"
      }
    },
    "required": ["payment_requirements", "callback_url"]
  }
}
```

**circular-protocol-enterprise-mcp-server Tools:**

1. `get_wallet_nonce`
```json
{
  "name": "get_wallet_nonce",
  "description": "Get current nonce for a Circular Protocol Enterprise API wallet",
  "inputSchema": {
    "type": "object",
    "properties": {
      "wallet_address": {
        "type": "string",
        "pattern": "^[a-fA-F0-9]{64}$",
        "description": "Circular Protocol wallet address (no 0x prefix)"
      }
    },
    "required": ["wallet_address"]
  }
}
```

2. `certify_data`
```json
{
  "name": "certify_data",
  "description": "Submit certification transaction to Circular Protocol via Enterprise API",
  "inputSchema": {
    "type": "object",
    "properties": {
      "data_hash": {
        "type": "string",
        "pattern": "^0x[a-fA-F0-9]{64}$",
        "description": "SHA256 hash of data to certify"
      },
      "from_wallet": {"type": "string"},
      "to_wallet": {"type": "string"},
      "nonce": {"type": "integer"},
      "signature": {"type": "string"},
      "public_key": {"type": "string"}
    },
    "required": ["data_hash", "from_wallet", "to_wallet", "nonce", "signature", "public_key"]
  }
}
```

3. `get_transaction_status`
```json
{
  "name": "get_transaction_status",
  "description": "Poll Circular Protocol transaction status via Enterprise API",
  "inputSchema": {
    "type": "object",
    "properties": {
      "tx_id": {
        "type": "string",
        "pattern": "^[a-fA-F0-9]{64}$",
        "description": "Transaction ID (no 0x prefix)"
      }
    },
    "required": ["tx_id"]
  }
}
```

4. `get_certification_proof`
```json
{
  "name": "get_certification_proof",
  "description": "Retrieve full certification proof for verified transaction",
  "inputSchema": {
    "type": "object",
    "properties": {
      "tx_id": {"type": "string"}
    },
    "required": ["tx_id"]
  }
}
```

**data-quote-mcp-server Tools:**

1. `check_data_size`
```json
{
  "name": "check_data_size",
  "description": "Calculate byte size of data",
  "inputSchema": {
    "type": "object",
    "properties": {
      "data": {
        "type": "string",
        "description": "Data in base64, hex, or plaintext"
      }
    },
    "required": ["data"]
  }
}
```

2. `get_cirx_price`
```json
{
  "name": "get_cirx_price",
  "description": "Fetch current CIRX/USD price from CoinGecko",
  "inputSchema": {
    "type": "object",
    "properties": {}
  }
}
```

3. `calculate_quote`
```json
{
  "name": "calculate_quote",
  "description": "Calculate USDC quote for certification",
  "inputSchema": {
    "type": "object",
    "properties": {
      "data_size_bytes": {"type": "integer"},
      "cirx_price_usd": {
        "type": "number",
        "description": "Current CIRX price (fetched if not provided)"
      },
      "margin_percent": {
        "type": "number",
        "default": 65.0,
        "description": "Service margin percentage"
      }
    },
    "required": ["data_size_bytes"]
  }
}
```

**qr-code-mcp-server Tools:**

1. `generate_qr_ascii`
```json
{
  "name": "generate_qr_ascii",
  "description": "Generate ASCII art QR code for terminal display",
  "inputSchema": {
    "type": "object",
    "properties": {
      "data": {"type": "string"},
      "size": {
        "type": "integer",
        "default": 256,
        "description": "QR code size in pixels"
      }
    },
    "required": ["data"]
  }
}
```

2. `generate_qr_image`
```json
{
  "name": "generate_qr_image",
  "description": "Generate QR code as PNG or SVG image",
  "inputSchema": {
    "type": "object",
    "properties": {
      "data": {"type": "string"},
      "format": {
        "type": "string",
        "enum": ["png", "svg"],
        "default": "png"
      },
      "size": {"type": "integer", "default": 256}
    },
    "required": ["data"]
  }
}
```

3. `encode_x402_to_qr`
```json
{
  "name": "encode_x402_to_qr",
  "description": "Encode x402 payment data in EIP-681 format for wallet scanning",
  "inputSchema": {
    "type": "object",
    "properties": {
      "payment_requirements": {"type": "object"},
      "callback_url": {"type": "string", "format": "uri"}
    },
    "required": ["payment_requirements", "callback_url"]
  }
}
```

---

### 2.4 HTTP API Specification

#### Endpoint: `POST /v1/quote`
**Purpose:** Get pricing quote without initiating payment

**Request:**
```json
{
  "data": "base64_or_hex_data",
  "data_hash": "0xsha256..." // optional
}
```

**Response 200 OK:**
```json
{
  "quote_id": "quote_abc123",
  "data_hash": "0x...",
  "data_size_bytes": 1024,
  "quote": {
    "usdc_amount": "0.050000",
    "cirx_fee": 4.0,
    "cirx_price_usd": 0.0044,
    "service_fee_usd": 0.0324,
    "margin_percent": 65.0,
    "breakdown": {
      "cirx_cost": "4 CIRX × $0.0044 = $0.0176",
      "service_margin": "$0.0324 (65%)",
      "total": "$0.05"
    }
  },
  "valid_until": "2025-10-28T12:40:00Z",
  "networks_supported": ["base", "base-sepolia", "arbitrum"]
}
```

---

#### Endpoint: `POST /v1/certify`
**Purpose:** Initiate certification or process payment

**Request (Initial):**
```json
{
  "request_id": "req_unique_123", // idempotency key
  "data": "base64...",
  "data_hash": "0x...", // optional
  "network": "base",
  "callback_url": "https://example.com/webhook", // optional
  "client_type": "agent" | "browser" | "mobile" // optional
}
```

**Response 402 Payment Required:**
```json
{
  "error": "payment_required",
  "request_id": "req_unique_123",
  "quote": {
    "usdc_amount": "0.050000",
    "cirx_fee": 4.0,
    "valid_until": "2025-10-28T12:40:00Z"
  },
  "payment": {
    "x402_version": 1,
    "scheme": "exact",
    "network": "base",
    "maxAmountRequired": "50000", // atomic units (6 decimals)
    "asset": "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
    "payTo": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "maxTimeoutSeconds": 300,
    "resource": "https://certify.ar4s.com/v1/certify",
    "description": "Certify data on Circular Protocol via Enterprise APIs"
  },
  "payment_options": {
    "agent": {
      "instructions": "Sign EIP-3009 authorization and retry with X-PAYMENT header"
    },
    "browser": {
      "deep_link": "https://metamask.app.link/send?...",
      "instructions": "Click link to pay with MetaMask"
    },
    "mobile": {
      "qr_code_ascii": "█████████\n██   ██ \n...",
      "qr_code_url": "/v1/qr/req_unique_123",
      "instructions": "Scan QR code with mobile wallet"
    }
  }
}
```

**Request (With Payment):**
Headers: `X-PAYMENT: base64_encoded_payment_payload`

**Response 200 OK (Sync Success):**
```json
{
  "request_id": "req_unique_123",
  "status": "completed",
  "payment": {
    "tx_hash": "0xevm_payment_hash",
    "network": "base",
    "amount_paid_usdc": "0.050000",
    "settled_at": "2025-10-28T12:35:54Z"
  },
  "certification": {
    "tx_id": "0xcirx_tx_id",
    "block_id": "12345",
    "timestamp": "2025-10-28T12:35:56Z",
    "data_hash": "0x...",
    "verification_url": "https://explorer.circular-protocol.org/tx/0xcirx_tx_id"
  },
  "completed_at": "2025-10-28T12:35:56Z"
}
```

**Response 202 Accepted (Async Processing):**
```json
{
  "request_id": "req_unique_123",
  "status": "certifying",
  "message": "Payment verified and settled. Certification in progress.",
  "payment": {
    "tx_hash": "0xevm_payment_hash",
    "settled_at": "2025-10-28T12:35:54Z"
  },
  "status_url": "/v1/status/req_unique_123",
  "estimated_completion": "2025-10-28T12:36:30Z"
}
```

---

#### Endpoint: `GET /v1/status/{request_id}`
**Purpose:** Check certification status

**Response 200 OK:**
```json
{
  "request_id": "req_unique_123",
  "status": "completed", // or "payment_pending", "certifying", "failed"
  "certification": {
    // ... (same as POST /v1/certify success response)
  },
  "created_at": "2025-10-28T12:35:50Z",
  "updated_at": "2025-10-28T12:35:56Z"
}
```

**Response 200 OK (Failed):**
```json
{
  "request_id": "req_unique_123",
  "status": "failed",
  "error": {
    "code": "CIRCULAR_TX_FAILED",
    "message": "Circular Protocol Enterprise API transaction failed: insufficient balance",
    "retryable": true,
    "retry_after": "2025-10-28T12:40:00Z"
  },
  "payment": {
    "status": "settled",
    "tx_hash": "0x..."
  }
}
```

---

#### Endpoint: `GET /v1/qr/{request_id}`
**Purpose:** Get QR code for mobile payment

**Query Parameters:**
- `format`: `png` | `svg` | `ascii` (default: `png`)
- `size`: integer (default: `256`)

**Response 200 OK (PNG):**
- Content-Type: `image/png`
- Body: Binary PNG data

**Response 200 OK (ASCII):**
- Content-Type: `text/plain`
- Body:
```
█████████████████████████████
█████████████████████████████
███ ▄▄▄▄▄ █▀█ █▄▄▀▄█ ▄▄▄▄▄ ███
███ █   █ ██▄█ ▀██▄▄█ █   █ ███
███ █▄▄▄█ █▀▄▀▀▀ ▀▀▄█ █▄▄▄█ ███
...
```

---

### 2.5 Error Handling Strategy

#### Payment Verification Failures
**Error:** `x402-mcp.verify_payment` returns `{is_valid: false}`

**Handling:**
1. Return 402 with detailed error in `payment_required.error` field
2. Do NOT store as failed certification (user can retry)
3. Common errors:
   - `invalid_signature`: EIP-3009 signature verification failed
   - `invalid_nonce`: Nonce already used or invalid
   - `expired_authorization`: validBefore timestamp passed
   - `amount_mismatch`: Payment value doesn't match requirements

---

#### Payment Settlement Failures
**Error:** `x402-mcp.settle_payment` fails (facilitator timeout, network error)

**Handling:**
1. Store payment record with `status='verified_pending_settlement'`
2. Return 202 Accepted to client
3. Background worker retries settlement:
   - Interval: 30s, 60s, 120s, 240s, 300s (exponential backoff, max 5min)
   - Max attempts: 10
4. If all retries fail:
   - Mark payment as `failed_settlement`
   - Trigger operator alert
   - Initiate refund process

---

#### Certification Failures (Critical Path)
**Error:** Circular Protocol Enterprise API certification fails AFTER payment settled

**Handling:**
1. Store certification record with `status='failed'`
2. Increment `retry_count`
3. Add to retry queue:
   - Delays: 5s, 10s, 20s, 40s, 80s, 160s, 300s, 600s, 1200s, 1800s
   - Max retries: 10 attempts over ~4 hours
4. If max retries exceeded:
   - Move to dead letter queue
   - Operator dashboard shows for manual review
   - Options:
     - Manual retry with investigation
     - Refund to user
     - Credit user account

**Common Causes:**
- Nonce desync (fetch fresh nonce from Enterprise API before retry)
- Circular Protocol Enterprise API network congestion (wait and retry)
- Service wallet out of CIRX (alert operator immediately)
- Invalid transaction format (requires code fix)

---

#### Idempotency Handling
**Scenario:** Client retries with same `request_id`

**Handling:**
```go
func (s *Service) HandleCertify(req CertifyRequest) Response {
    existing, err := s.db.GetRequest(req.RequestID)
    if err == nil {
        // Request already exists
        switch existing.Status {
        case "completed":
            return s.getCachedResult(req.RequestID) // 200 OK
        case "failed":
            // Allow retry
            return s.processCertification(req)
        case "certifying", "payment_pending":
            return Response{
                Status: 202,
                Message: "Request already processing",
                StatusURL: fmt.Sprintf("/v1/status/%s", req.RequestID),
            }
        }
    }

    // New request
    return s.processCertification(req)
}
```

---

### 2.6 Security Architecture

#### Wallet Key Management
**Service Wallet Private Key:**
- Stored encrypted at rest using AES-256
- Encryption key from environment variable or Vault
- Never logged or exposed in responses
- Separate keys for testnet and mainnet
- Key rotation process documented

**Example Configuration:**
```yaml
circular:
  service_wallet:
    address_env: CIRX_WALLET_ADDRESS
    encrypted_key_path: /secrets/cirx_key.enc
    encryption_passphrase_env: CIRX_KEY_PASSPHRASE
```

---

#### EIP-3009 Signature Verification
**Critical Validation Steps:**
```go
func verifyPayment(payment PaymentHeader, requirements PaymentRequirements) error {
    // 1. Decode base64 payment header
    decoded, err := base64.StdEncoding.DecodeString(payment.Header)

    // 2. Parse authorization data
    auth := parseAuthorization(decoded)

    // 3. Validate amounts match
    if auth.Value != requirements.MaxAmountRequired {
        return errors.New("amount_mismatch")
    }

    // 4. Validate addresses
    if auth.To != requirements.PayTo {
        return errors.New("recipient_mismatch")
    }

    // 5. Validate time window
    now := time.Now().Unix()
    if now < auth.ValidAfter || now >= auth.ValidBefore {
        return errors.New("expired_authorization")
    }

    // 6. Check nonce uniqueness (prevent replay)
    exists, _ := db.NonceExists(auth.Nonce)
    if exists {
        return errors.New("duplicate_nonce")
    }

    // 7. Recover signer from signature
    signer, err := recoverSigner(auth, payment.Signature)
    if err != nil {
        return errors.New("invalid_signature")
    }

    // 8. Verify signer matches from address
    if signer != auth.From {
        return errors.New("signer_mismatch")
    }

    // 9. Store nonce to prevent reuse
    db.StoreNonce(auth.Nonce)

    return nil
}
```

---

#### API Authentication & Rate Limiting
**Implementation:**
```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        apiKey := r.Header.Get("X-API-Key")
        if apiKey == "" {
            http.Error(w, "Missing API key", 401)
            return
        }

        // Validate API key (bcrypt hash comparison)
        clientID, valid := validateAPIKey(apiKey)
        if !valid {
            http.Error(w, "Invalid API key", 401)
            return
        }

        // Rate limiting (Redis-based)
        allowed, err := rateLimiter.Allow(clientID, 10, time.Minute)
        if !allowed {
            w.Header().Set("Retry-After", "60")
            http.Error(w, "Rate limit exceeded", 429)
            return
        }

        // Store client ID in request context
        ctx := context.WithValue(r.Context(), "client_id", clientID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

---

#### Input Validation
**Data Size Limits:**
```go
const (
    MaxDataSizeMB = 10
    MaxDataSizeBytes = MaxDataSizeMB * 1024 * 1024
)

func validateCertifyRequest(req CertifyRequest) error {
    // Validate request_id format
    if !regexp.MustCompile(`^req_[a-zA-Z0-9_-]{10,50}$`).MatchString(req.RequestID) {
        return errors.New("invalid_request_id_format")
    }

    // Validate data size
    dataBytes := base64DecodedSize(req.Data)
    if dataBytes > MaxDataSizeBytes {
        return fmt.Errorf("data_too_large: max %dMB", MaxDataSizeMB)
    }

    // Validate network
    validNetworks := []string{"base", "base-sepolia", "arbitrum"}
    if !contains(validNetworks, req.Network) {
        return errors.New("invalid_network")
    }

    // Validate callback URL (must be HTTPS in prod)
    if req.CallbackURL != "" {
        u, err := url.Parse(req.CallbackURL)
        if err != nil || u.Scheme != "https" {
            return errors.New("invalid_callback_url")
        }
    }

    return nil
}
```

---

### 2.7 Monitoring & Observability

#### Prometheus Metrics
```go
var (
    certificationTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "certify_certifications_total",
            Help: "Total number of certification attempts",
        },
        []string{"status"}, // completed, failed
    )

    certificationDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "certify_certification_duration_seconds",
            Help: "Time taken for full certification flow",
            Buckets: []float64{1, 2, 5, 10, 20, 30, 60},
        },
        []string{"client_type"}, // agent, browser, mobile
    )

    paymentVerificationRate = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "certify_payment_verifications_total",
            Help: "Payment verification attempts",
        },
        []string{"result"}, // valid, invalid
    )

    cirxWalletBalance = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "certify_cirx_wallet_balance",
            Help: "Current CIRX balance in service wallet",
        },
    )

    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "certify_http_request_duration_seconds",
            Help: "HTTP request latency",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path", "status"},
    )
)
```

#### Alert Rules (Prometheus Alertmanager)
```yaml
groups:
  - name: certify_alerts
    interval: 1m
    rules:
      - alert: LowCIRXBalance
        expr: certify_cirx_wallet_balance < 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "CIRX wallet balance critically low"
          description: "Balance: {{ $value }} CIRX (< 100)"

      - alert: HighCertificationFailureRate
        expr: |
          rate(certify_certifications_total{status="failed"}[5m])
          / rate(certify_certifications_total[5m]) > 0.05
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "Certification failure rate > 5%"

      - alert: HighPaymentVerificationFailureRate
        expr: |
          rate(certify_payment_verifications_total{result="invalid"}[5m])
          / rate(certify_payment_verifications_total[5m]) > 0.10
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Payment verification failure rate > 10%"

      - alert: SlowCertificationTime
        expr: |
          histogram_quantile(0.95,
            certify_certification_duration_seconds_bucket
          ) > 15
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "P95 certification time > 15 seconds"
```

---

## Phase 3: Implementation

### 3.1 Project Structure

```
certify-ar4s/
├── mcp-servers/
│   ├── x402-mcp-server/
│   │   ├── main.go
│   │   ├── config.yaml
│   │   ├── tools/
│   │   │   ├── create_payment.go
│   │   │   ├── verify_payment.go
│   │   │   ├── settle_payment.go
│   │   │   ├── browser_link.go
│   │   │   └── encode_qr.go
│   │   ├── internal/
│   │   │   ├── x402/
│   │   │   ├── eip3009/
│   │   │   └── facilitator/
│   │   └── tests/
│   │
│   ├── circular-protocol-mcp-server/
│   │   ├── main.go
│   │   ├── config.yaml
│   │   ├── tools/
│   │   │   ├── get_nonce.go
│   │   │   ├── certify_data.go
│   │   │   ├── tx_status.go
│   │   │   └── get_proof.go
│   │   ├── internal/
│   │   │   ├── circular/
│   │   │   └── crypto/
│   │   └── tests/
│   │
│   ├── data-quote-mcp-server/
│   │   ├── main.go
│   │   ├── config.yaml
│   │   ├── tools/
│   │   │   ├── check_size.go
│   │   │   ├── get_price.go
│   │   │   └── calculate_quote.go
│   │   ├── internal/
│   │   │   ├── pricing/
│   │   │   └── coingecko/
│   │   └── tests/
│   │
│   └── qr-code-mcp-server/
│       ├── main.go
│       ├── config.yaml
│       ├── tools/
│       │   ├── generate_ascii.go
│       │   ├── generate_image.go
│       │   └── encode_x402.go
│       ├── internal/
│       │   └── qr/
│       └── tests/
│
├── proxy/                        # certify.ar4s.com
│   ├── main.go
│   ├── config.yaml
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── quote.go
│   │   │   ├── certify.go
│   │   │   ├── status.go
│   │   │   └── qr.go
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   ├── ratelimit.go
│   │   │   └── logging.go
│   │   └── router.go
│   ├── internal/
│   │   ├── mcp/
│   │   │   └── client.go
│   │   ├── orchestration/
│   │   │   ├── workflow.go
│   │   │   └── state_machine.go
│   │   ├── storage/
│   │   │   ├── postgres.go
│   │   │   └── redis.go
│   │   └── retry/
│   │       └── queue.go
│   ├── migrations/
│   │   ├── 001_init.up.sql
│   │   └── 001_init.down.sql
│   └── tests/
│
├── pkg/                          # Shared packages
│   ├── models/
│   │   ├── request.go
│   │   ├── payment.go
│   │   └── certification.go
│   ├── crypto/
│   │   └── signing.go
│   └── errors/
│       └── errors.go
│
├── docs/
│   ├── SPEC.md                   # This file
│   ├── API.md
│   ├── ARCHITECTURE.md
│   ├── DEPLOYMENT.md
│   └── RUNBOOK.md
│
├── scripts/
│   ├── setup-dev.sh
│   ├── run-all-servers.sh
│   └── migrate.sh
│
├── deployments/
│   ├── docker/
│   │   ├── Dockerfile.proxy
│   │   └── Dockerfile.mcp-server
│   ├── kubernetes/
│   │   ├── proxy-deployment.yaml
│   │   ├── mcp-servers-deployment.yaml
│   │   └── configmap.yaml
│   └── docker-compose.yml
│
├── .github/
│   └── workflows/
│       ├── ci.yaml
│       └── deploy.yaml
│
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

### 3.2 Implementation Milestones

#### Milestone 1: Foundation (Week 1)
**Goal:** Project setup and base infrastructure

**TASK-001: Repository Setup**
- [ ] Initialize Git monorepo
- [ ] Set up Go module (`go mod init github.com/yourusername/certify-ar4s`)
- [ ] Create directory structure as per 3.1
- [ ] Set up `.gitignore` for Go, IDE files, secrets
- [ ] Write initial README with architecture overview

**TASK-002: Development Environment**
- [ ] Create `docker-compose.yml` with PostgreSQL, Redis
- [ ] Write `scripts/setup-dev.sh` to initialize local environment
- [ ] Document development workflow in `docs/DEVELOPMENT.md`
- [ ] Set up VS Code / GoLand configuration files

**TASK-003: Database Setup**
- [ ] Write SQL migration: `001_init.up.sql` (schema from 2.3.1)
- [ ] Write SQL migration: `001_init.down.sql` (rollback)
- [ ] Integrate golang-migrate
- [ ] Test migrations: up, down, re-up

**TASK-004: Shared Packages**
- [ ] Implement `pkg/models` (Request, Payment, Certification structs)
- [ ] Implement `pkg/errors` (custom error types)
- [ ] Implement `pkg/crypto` (signing utilities)
- [ ] Write unit tests for shared packages

---

#### Milestone 2: MCP Server - x402 (Week 2)
**Goal:** Build x402 payment MCP server

**TASK-005: x402-mcp-server Setup**
- [ ] Initialize server with mcp-go: `server.NewMCPServer(...)`
- [ ] Implement stdio transport
- [ ] Add tool registration skeleton
- [ ] Write basic test: start server, list tools

**TASK-006: Tool - create_payment_requirement**
- [ ] Implement `handleCreatePaymentRequirement`
- [ ] Generate x402-compliant payment requirements JSON
- [ ] Support multiple networks (base, base-sepolia, arbitrum)
- [ ] Write unit tests with different networks and amounts

**TASK-007: Tool - verify_payment**
- [ ] Implement EIP-3009 signature recovery
- [ ] Validate authorization fields (amount, addresses, time window)
- [ ] Check nonce uniqueness (in-memory cache for testing)
- [ ] Write unit tests with valid and invalid signatures

**TASK-008: Tool - settle_payment**
- [ ] Implement HTTP client for facilitator API
- [ ] POST `/settle` with payment payload
- [ ] Parse settlement response (tx hash, success status)
- [ ] Handle facilitator errors (timeout, network issues)
- [ ] Write integration test with mock facilitator

**TASK-009: Tool - generate_browser_link**
- [ ] Generate MetaMask deep link format
- [ ] Support Coinbase Wallet and Rainbow formats
- [ ] URL encode payment parameters
- [ ] Write unit tests for link generation

**TASK-010: Tool - encode_payment_for_qr**
- [ ] Format payment data in EIP-681 URI scheme
- [ ] Include callback URL for wallet POST
- [ ] Write unit tests

**TASK-011: x402-mcp-server Integration Testing**
- [ ] End-to-end test: create → verify → settle
- [ ] Test with x402 facilitator (testnet)
- [ ] Verify payment on blockchain explorer

---

#### Milestone 3: MCP Server - Circular Protocol Enterprise APIs (Week 2-3)
**Goal:** Build Circular Protocol Enterprise API MCP server (using Go implementation pattern from docs/GO-CEP-APIS.xml)

**TASK-012: circular-protocol-enterprise-mcp-server Setup**
- [ ] Initialize server with mcp-go
- [ ] Implement Circular Protocol Enterprise HTTP REST API client
- [ ] Test connectivity to Enterprise API testnet endpoints

**TASK-013: Tool - get_wallet_nonce**
- [ ] Implement `GET Circular_GetWalletNonce_` Enterprise API call
- [ ] Parse response (nonce value)
- [ ] Handle API errors
- [ ] Write unit tests with mock API

**TASK-014: Tool - certify_data**
- [ ] Construct transaction payload:
  - Type: "C_TYPE_CERTIFICATE"
  - Payload: `{"Action":"CP_CERTIFICATE","Data":"<hex_encoded_hash>"}`
- [ ] Calculate transaction ID client-side: `sha256(Blockchain + From + To + Payload + Nonce + Timestamp)` (as per Enterprise APIs pattern)
- [ ] Sign transaction ID with Secp256k1
- [ ] POST `Circular_AddTransaction_` Enterprise API endpoint
- [ ] Write unit tests with mock signing

**TASK-015: Tool - get_transaction_status**
- [ ] Implement `GET transaction by ID` Enterprise API call
- [ ] Parse status field ("Executed", "Pending")
- [ ] Write unit tests

**TASK-016: Tool - get_certification_proof**
- [ ] Fetch full transaction details from Enterprise API
- [ ] Extract block ID, timestamp, payload
- [ ] Generate verification URL (block explorer)
- [ ] Write unit tests

**TASK-017: circular-protocol-enterprise-mcp-server Integration Testing**
- [ ] End-to-end test on Circular Protocol Enterprise API testnet
- [ ] Submit actual certification transaction
- [ ] Verify on explorer
- [ ] Measure confirmation time

---

#### Milestone 4: MCP Servers - Quote & QR (Week 3)
**Goal:** Build data-quote and qr-code MCP servers

**TASK-018: data-quote-mcp-server Setup**
- [ ] Initialize server with mcp-go
- [ ] Set up CoinGecko API client

**TASK-019: Tool - check_data_size**
- [ ] Decode base64/hex data
- [ ] Calculate byte size
- [ ] Format human-readable size (KB, MB)
- [ ] Write unit tests

**TASK-020: Tool - get_cirx_price**
- [ ] Call CoinGecko API: `/simple/price?ids=circular&vs_currencies=usd`
- [ ] Parse response
- [ ] Cache price in memory (5 minute TTL)
- [ ] Write unit tests with mock API

**TASK-021: Tool - calculate_quote**
- [ ] Formula: `(4 CIRX × CIRX_price_USD) × (1 + margin%)`
- [ ] Return detailed breakdown
- [ ] Support custom margin parameter
- [ ] Write unit tests

**TASK-022: qr-code-mcp-server Setup**
- [ ] Initialize server with mcp-go
- [ ] Integrate QR code library (e.g., `github.com/skip2/go-qrcode`)

**TASK-023: Tool - generate_qr_ascii**
- [ ] Generate QR code bitmap
- [ ] Convert to ASCII art using █ character
- [ ] Make terminal-friendly (ensure proper display)
- [ ] Write unit tests

**TASK-024: Tool - generate_qr_image**
- [ ] Generate PNG QR code
- [ ] Generate SVG QR code
- [ ] Base64 encode for JSON response
- [ ] Write unit tests

**TASK-025: Tool - encode_x402_to_qr**
- [ ] Format payment data as EIP-681 URI
- [ ] Include callback URL
- [ ] Write unit tests

---

#### Milestone 5: MCP Host Proxy - Core (Week 4)
**Goal:** Build certify.ar4s.com proxy service

**TASK-026: Proxy Setup**
- [ ] Initialize Gin/Echo HTTP server
- [ ] Set up configuration loading (config.yaml + env vars)
- [ ] Implement health check endpoint `/health`

**TASK-027: MCP Client Connections**
- [ ] Create mcp-go client for each of 4 servers
- [ ] Use stdio transport
- [ ] Implement connection management (reconnect on failure)
- [ ] Test: start all 4 servers, verify connections

**TASK-028: Database Layer**
- [ ] Implement PostgreSQL connection pool
- [ ] Implement CRUD for `certification_requests`
- [ ] Implement CRUD for `payments`
- [ ] Implement CRUD for `certifications`
- [ ] Write integration tests

**TASK-029: Redis Layer**
- [ ] Implement Redis connection
- [ ] Implement cache functions (get, set, expire)
- [ ] Cache CIRX price (5 min TTL)
- [ ] Cache payment status
- [ ] Write integration tests

**TASK-030: Middleware**
- [ ] Implement `authMiddleware` (API key validation)
- [ ] Implement `rateLimitMiddleware` (Redis-based)
- [ ] Implement `corsMiddleware`
- [ ] Implement `loggingMiddleware` (structured logs)
- [ ] Write unit tests for each middleware

---

#### Milestone 6: MCP Host Proxy - API Handlers (Week 4-5)
**Goal:** Implement HTTP API endpoints

**TASK-031: Handler - POST /v1/quote**
- [ ] Parse request body
- [ ] Call `data-quote-mcp.check_data_size`
- [ ] Call `data-quote-mcp.get_cirx_price`
- [ ] Call `data-quote-mcp.calculate_quote`
- [ ] Store quote in database (optional)
- [ ] Return quote response
- [ ] Write integration tests

**TASK-032: Handler - POST /v1/certify (Initial Request)**
- [ ] Parse and validate request
- [ ] Check for existing `request_id` (idempotency)
- [ ] Get quote (reuse TASK-031 logic)
- [ ] Call `x402-mcp.create_payment_requirement`
- [ ] If `client_type=browser`: call `x402-mcp.generate_browser_link`
- [ ] If `client_type=mobile`: call `qr-code-mcp` tools
- [ ] Return 402 Payment Required response
- [ ] Store request in database
- [ ] Write integration tests

**TASK-033: Handler - POST /v1/certify (With Payment)**
- [ ] Parse X-PAYMENT header
- [ ] Call `x402-mcp.verify_payment`
- [ ] If invalid: return 402 with error
- [ ] Call `x402-mcp.settle_payment`
- [ ] Store payment in database
- [ ] Trigger certification workflow (async)
- [ ] Return 202 Accepted or 200 OK (if sync)
- [ ] Write integration tests

**TASK-034: Handler - GET /v1/status/:id**
- [ ] Fetch request from database
- [ ] Fetch payment and certification records
- [ ] Build status response
- [ ] Write integration tests

**TASK-035: Handler - GET /v1/qr/:id**
- [ ] Fetch request from database
- [ ] Call `qr-code-mcp.encode_x402_to_qr`
- [ ] Call `qr-code-mcp.generate_qr_ascii` or `generate_qr_image`
- [ ] Return QR code (image or text)
- [ ] Write integration tests

---

#### Milestone 7: Orchestration & State Machine (Week 5)
**Goal:** Implement certification workflow orchestration

**TASK-036: State Machine Implementation**
- [ ] Define state enum and transitions
- [ ] Implement state transition validation
- [ ] Store state in database
- [ ] Emit state change events (for monitoring)
- [ ] Write unit tests

**TASK-037: Certification Workflow**
- [ ] After payment settled, call `circular-protocol-enterprise-mcp.get_wallet_nonce`
- [ ] Sign certification transaction locally
- [ ] Call `circular-protocol-enterprise-mcp.certify_data`
- [ ] Poll `circular-protocol-enterprise-mcp.get_transaction_status` until "Executed"
- [ ] Call `circular-protocol-enterprise-mcp.get_certification_proof`
- [ ] Update database records
- [ ] Trigger webhook callback if provided
- [ ] Write integration tests

**TASK-038: Retry Queue**
- [ ] Implement background worker (goroutine)
- [ ] Scan database for failed certifications
- [ ] Implement exponential backoff
- [ ] Max 10 retry attempts
- [ ] Move to dead letter queue after max retries
- [ ] Write integration tests

**TASK-039: Webhook Callbacks**
- [ ] HTTP POST to `callback_url` on completion
- [ ] Include certification proof in body
- [ ] Sign webhook payload (HMAC)
- [ ] Retry webhook on failure (3 attempts)
- [ ] Write integration tests

---

#### Milestone 8: Monitoring & Operations (Week 6)
**Goal:** Operational readiness

**TASK-040: Prometheus Metrics**
- [ ] Implement metrics (as per 2.7)
- [ ] Expose `/metrics` endpoint
- [ ] Test metrics collection

**TASK-041: CIRX Wallet Balance Monitor**
- [ ] Background job: check balance every 5 minutes
- [ ] Update `cirx_wallet_balance` metric
- [ ] Trigger alert if < 100 CIRX
- [ ] Write integration tests

**TASK-042: Logging**
- [ ] Integrate Zap logger
- [ ] Structured JSON logs
- [ ] Log levels: DEBUG, INFO, WARN, ERROR
- [ ] Never log sensitive data (private keys, full payment headers)

**TASK-043: Grafana Dashboard**
- [ ] Create dashboard JSON
- [ ] Panels:
  - Certification success rate (7-day trend)
  - Average certification time (P50, P95, P99)
  - Payment verification rate
  - CIRX wallet balance
  - HTTP request latency
  - Error rate by endpoint
- [ ] Document dashboard setup

**TASK-044: Alertmanager Rules**
- [ ] Write alert rules (as per 2.7)
- [ ] Configure PagerDuty or Slack integration
- [ ] Test alert firing

---

#### Milestone 9: Testing & QA (Week 7)
**Goal:** Comprehensive testing

**TASK-045: Unit Test Coverage**
- [ ] Achieve 80%+ coverage across all packages
- [ ] Run: `go test ./... -cover`
- [ ] Fix failing tests

**TASK-046: Integration Tests**
- [ ] Test full flow with all 4 MCP servers
- [ ] Use testnet blockchains
- [ ] Verify end-to-end: quote → payment → certification
- [ ] Test error paths

**TASK-047: Load Testing**
- [ ] Use `k6` or `vegeta` for load testing
- [ ] Test 100 concurrent requests
- [ ] Measure P95 latency
- [ ] Identify bottlenecks

**TASK-048: Security Testing**
- [ ] Run `gosec` for vulnerability scanning
- [ ] Test API key authentication bypass attempts
- [ ] Test payment signature forgery attempts
- [ ] Test rate limiting effectiveness
- [ ] Test SQL injection resistance

---

#### Milestone 10: Deployment (Week 8)
**Goal:** Production deployment

**TASK-049: Docker Images**
- [ ] Write `Dockerfile.proxy`
- [ ] Write `Dockerfile.mcp-server`
- [ ] Multi-stage builds for size optimization
- [ ] Test local Docker builds

**TASK-050: Kubernetes Manifests**
- [ ] Write Deployment for proxy (3 replicas)
- [ ] Write Deployment for MCP servers
- [ ] Write Service manifests
- [ ] Write ConfigMap for config.yaml
- [ ] Write Secret for wallet keys
- [ ] Write HPA (Horizontal Pod Autoscaler)
- [ ] Test on local Kubernetes (minikube)

**TASK-051: CI/CD Pipeline**
- [ ] GitHub Actions: `.github/workflows/ci.yaml`
  - Run tests on PR
  - Lint with `golangci-lint`
  - Build Docker images
- [ ] GitHub Actions: `.github/workflows/deploy.yaml`
  - Deploy to staging on merge to `main`
  - Deploy to production on tag release
- [ ] Test CI/CD workflow

**TASK-052: Testnet Deployment**
- [ ] Deploy to staging environment
- [ ] Use base-sepolia, Circular Protocol Enterprise API testnet
- [ ] Run smoke tests
- [ ] Verify monitoring works

**TASK-053: Mainnet Deployment**
- [ ] Purchase CIRX for production wallet
- [ ] Update config for mainnet networks
- [ ] Deploy to production
- [ ] Run smoke tests
- [ ] Monitor for 24 hours

**TASK-054: Documentation**
- [ ] Finalize API.md (OpenAPI spec)
- [ ] Write DEPLOYMENT.md (how to deploy)
- [ ] Write RUNBOOK.md (operations procedures)
- [ ] Update README.md

---

### 3.3 Success Criteria

#### Technical Metrics
- ✅ **P95 Latency:** < 10 seconds for full certification flow
- ✅ **Throughput:** 100 concurrent requests without errors
- ✅ **Uptime:** 99.5% over 30 days
- ✅ **Error Rate:** < 1% payment failures, < 2% certification failures
- ✅ **Test Coverage:** > 80% across all packages

#### Business Metrics
- ✅ **Successful Certifications:** 1,000+ in first month
- ✅ **Payment Success Rate:** > 95%
- ✅ **User Types:** All 3 workflows functional (agent, browser, mobile)
- ✅ **Revenue:** Positive cash flow (revenue > costs)

#### Operational Metrics
- ✅ **CIRX Balance:** Never drop below critical threshold (100 CIRX)
- ✅ **Monitoring:** All alerts configured and tested
- ✅ **Documentation:** Complete (spec, API docs, runbook, architecture)
- ✅ **Security:** No vulnerabilities in security audit

---

## Appendix

### A. Glossary

- **MCP (Model Context Protocol):** Protocol for LLMs to interact with external tools
- **MCP Host:** Application that connects to MCP servers (certify.ar4s.com)
- **MCP Server:** Service providing tools via MCP (x402, circular-protocol-enterprise, etc.)
- **x402:** HTTP payment protocol using 402 status code
- **EIP-3009:** Ethereum standard for gasless token transfers via signed authorization
- **CIRX:** Native token of Circular Protocol blockchain
- **Circular Protocol:** Layer 1 blockchain for data certification (accessed via Enterprise APIs)
- **Enterprise APIs:** Circular Protocol's HTTP REST APIs for blockchain interaction (Go implementation pattern)
- **Facilitator:** Service that settles x402 payments on blockchain
- **Idempotency:** Property ensuring duplicate requests return same result without side effects

### B. Technology Stack Summary

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Language | Go 1.23+ | All services |
| MCP Framework | mcp-go (github.com/mark3labs/mcp-go) | MCP server/client implementation |
| HTTP Framework | Gin or Echo | certify.ar4s.com HTTP API |
| Database | PostgreSQL 16+ | Persistent storage |
| Cache | Redis 7+ | Price cache, rate limiting |
| Blockchain (EVM) | go-ethereum (geth) | EVM chain interactions |
| Blockchain (Circular) | Enterprise HTTP REST API | Circular Protocol Enterprise API integration (Go pattern) |
| Monitoring | Prometheus + Grafana | Metrics and dashboards |
| Logging | Zap | Structured logging |
| Tracing | OpenTelemetry | Distributed tracing |
| Deployment | Docker + Kubernetes | Container orchestration |
| CI/CD | GitHub Actions | Automation |

### C. External Dependencies

**Required Services:**
- x402 facilitator: `https://x402.org/facilitator`
- Circular Protocol Enterprise API: NAG URL discovery endpoint (see docs/GO-CEP-APIS.xml)
- CoinGecko API: `https://api.coingecko.com/api/v3`

**Blockchain Networks:**
- Base Sepolia (testnet): Chain ID 84532
- Base Mainnet: Chain ID 8453
- Arbitrum: Chain ID 42161
- Circular Protocol: Custom L1 (accessed via Enterprise APIs - testnet/mainnet)

**Required Credentials:**
- Service wallet (CIRX): Private key for signing certifications
- Service wallet (EVM): Address for receiving USDC payments
- CoinGecko API key: For CIRX price fetching
- API keys: For client authentication

### D. Cost Model

**Operating Costs (at 100 certs/day):**
- CIRX fees: 400 CIRX/day × $0.0044 = $1.76/day
- Infrastructure: $280/month = $9.33/day
- Total: ~$11/day = $330/month

**Revenue (at $0.05/cert):**
- 100 certs/day × $0.05 = $5/day
- Monthly: $150

**Break-even:** ~284 certs/day

**At 1,000 certs/day:**
- Revenue: $50/day = $1,500/month
- Costs: $27/day = $810/month
- Profit: $690/month

### E. Future Enhancements

**Phase 2 (Post-MVP):**
- Batch certifications (multiple items in one transaction)
- Additional EVM networks (Polygon, Avalanche)
- Subscription pricing tiers
- White-label API for enterprise

**Phase 3 (Advanced):**
- Support for other L1s beyond Circular Protocol
- Zero-knowledge proof certifications
- Mobile app for certification management
- Browser extension for one-click certifications
- SDK libraries (Python, TypeScript, Rust)

### F. Risk Assessment

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| CIRX price volatility | Medium | High | Update pricing frequently, maintain margin buffer |
| x402 facilitator downtime | High | Low | Queue requests, retry logic, alert operators |
| Circular Protocol Enterprise API issues | High | Low | Retry queue, dead letter queue, operator intervention |
| Payment succeeds but cert fails | Critical | Medium | Robust retry logic, operator dashboard, refund process |
| Security breach (wallet key theft) | Critical | Low | Encrypted storage, Vault integration, monitoring |

---

## Approval & Sign-off

**Specification Author:** Claude Code + User
**Version:** 1.0
**Date:** 2025-10-28
**Status:** Draft - Ready for Review

**Next Steps:**
1. ✅ Review and approve specification
2. ⏳ Validate external dependencies (x402 facilitator, Circular Protocol Enterprise APIs)
3. ⏳ Set up development environment
4. ⏳ Begin Milestone 1 implementation

---

**End of Specification**
