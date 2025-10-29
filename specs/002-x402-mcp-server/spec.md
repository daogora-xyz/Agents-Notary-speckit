# Feature Specification: x402 Payment MCP Server

**Feature Branch**: `002-x402-mcp-server`
**Created**: 2025-10-28
**Status**: Draft
**Input**: User description: "Build MCP server using mcp-go with 5 tools - create_payment_requirement (generates x402 JSON), verify_payment (EIP-3009 signature validation), settle_payment (calls facilitator API), generate_browser_link (MetaMask deep links), encode_payment_for_qr (EIP-681 format). Support base, base-sepolia, arbitrum networks."

## Clarifications

### Session 2025-10-28

- Q: Where should the `settle_payment` tool store cached settlement results for idempotency (FR-013)? → A: In-memory cache with TTL (5-10 minutes)
- Q: How should payment requirement nonces be generated (FR-005, Open Question Q2)? → A: Blockchain-sourced nonces - Query wallet's transaction count from RPC endpoint for each payment
- Q: What data should wallet callbacks include when users complete payments via browser links or QR codes (User Story 4, acceptance scenario 5)? → A: Transaction hash only via URL param (e.g., callback_url?tx=0x123...)
- Q: How should the MCP server control access to its tools (security consideration)? → A: No authentication - stdio transport security (inherits HTTP proxy's security context)
- Q: What observability instrumentation should the MCP server implement beyond basic logging (QM-004)? → A: Structured logging with levels - JSON logs with DEBUG/INFO/WARN/ERROR levels, context fields (tool, network, duration)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Generate Payment Requirements (Priority: P1)

The certify.ar4s.com HTTP proxy service needs to create x402-compliant payment requirements when an AI agent or browser user requests blockchain certification.

**Why this priority**: This is the foundational capability - without payment requirements generation, the entire x402 payment flow cannot begin. It's the first step in every payment workflow.

**Independent Test**: Can be fully tested by calling `create_payment_requirement` tool with various amounts and networks, then validating the output JSON conforms to x402 specification. Delivers immediately usable payment requirement JSON for integration testing.

**Acceptance Scenarios**:

1. **Given** a certification quote of 0.05 USDC on base network, **When** `create_payment_requirement` is called with amount=50000 (atomic units), network="base", **Then** returns valid x402 JSON with scheme="exact", maxAmountRequired="50000", payee address, valid_until timestamp (+5 min), and nonce
2. **Given** different networks (base, base-sepolia, arbitrum), **When** tool is called with same amount on each network, **Then** returns network-specific payee addresses and correct chain IDs
3. **Given** an invalid network name, **When** tool is called, **Then** returns error indicating unsupported network
4. **Given** payment requirement generation, **When** called multiple times with same inputs, **Then** generates unique nonces each time to prevent replay attacks

---

### User Story 2 - Verify EIP-3009 Payment Signatures (Priority: P1)

The HTTP proxy must validate that payment authorizations submitted by agents are cryptographically valid before attempting blockchain settlement.

**Why this priority**: Prevents wasted gas fees and settlement attempts with invalid signatures. Critical for security - without this, malicious actors could submit invalid payments.

**Independent Test**: Can be tested by generating valid EIP-3009 signatures (using test private keys) and invalid ones (wrong signature, modified data), then verifying the tool correctly identifies each. No blockchain interaction needed.

**Acceptance Scenarios**:

1. **Given** a valid EIP-3009 authorization (from, to, value, validAfter, validBefore, nonce, signature), **When** `verify_payment` is called, **Then** returns `{is_valid: true, signer_address: "0x..."}` with recovered signer address
2. **Given** an authorization where signature doesn't match the data, **When** `verify_payment` is called, **Then** returns `{is_valid: false, error: "signature verification failed"}`
3. **Given** an authorization with modified value (after signing), **When** verification is called, **Then** detects tampering and returns invalid
4. **Given** an authorization signed for a different network, **When** verified against current network's chain ID, **Then** returns invalid due to EIP-712 domain mismatch
5. **Given** an expired authorization (validBefore < current_time), **When** verified, **Then** returns `{is_valid: false, error: "authorization expired"}`

---

### User Story 3 - Settle Payments via Facilitator (Priority: P2)

After verification, the system must submit valid payment authorizations to the x402 facilitator for on-chain settlement.

**Why this priority**: Completes the payment flow and enables funds transfer. While critical, it depends on Story 2 (verification) completing first. P2 because payment can be "verified but pending settlement" temporarily.

**Independent Test**: Can test against x402 facilitator testnet endpoint with test USDC. Verifies HTTP interaction, error handling, and async settlement tracking without needing full certification workflow.

**Acceptance Scenarios**:

1. **Given** a verified EIP-3009 authorization, **When** `settle_payment` is called with facilitator_url="https://x402.org/facilitator", **Then** POSTs authorization to facilitator and returns `{status: "settled", tx_hash: "0x...", block_number: 12345}`
2. **Given** facilitator returns timeout (>5 seconds), **When** settlement is attempted, **Then** returns `{status: "pending", retry_after: 30}` for async retry
3. **Given** facilitator returns error (insufficient balance, nonce reused), **When** settlement is called, **Then** returns `{status: "failed", error: "facilitator error: <details>"}`
4. **Given** network connectivity issues, **When** settlement POST fails, **Then** tool handles exception gracefully and returns retriable error
5. **Given** successful settlement, **When** queried multiple times with same authorization (idempotency), **Then** returns cached result without re-submitting to blockchain

---

### User Story 4 - Generate Browser Payment Links (Priority: P2)

Browser users need MetaMask deep links to complete payments without programmatic wallet access.

**Why this priority**: Enables browser user workflow (US-004, US-005 from OVERVIEW.md). Important for user experience but can be implemented after core payment primitives (P1 stories). Desktop users represent secondary workflow.

**Independent Test**: Call `generate_browser_link` with payment requirements and verify returned URL opens MetaMask with pre-filled transaction. Test on various browsers (Chrome, Firefox, Brave) with MetaMask extension.

**Acceptance Scenarios**:

1. **Given** x402 payment requirements (from Story 1), **When** `generate_browser_link` is called with callback_url="https://certify.ar4s.com/callback/abc123", **Then** returns MetaMask deep link like `https://metamask.app.link/send/0xPayeeAddress@<chain_id>?value=50000&...`
2. **Given** browser user clicks the generated link, **When** MetaMask opens, **Then** transaction is pre-filled with correct recipient, amount (USDC atomic units), and network
3. **Given** different networks (base=8453, base-sepolia=84532), **When** link is generated, **Then** includes correct chain_id parameter for network switching
4. **Given** callback URL with special characters, **When** link is generated, **Then** URL-encodes callback properly
5. **Given** user completes MetaMask transaction, **When** wallet calls callback URL with tx hash parameter (e.g., `callback_url?tx=0x123...`), **Then** HTTP proxy can parse transaction hash and verify settlement on-chain

---

### User Story 5 - Encode Payments for QR Codes (Priority: P3)

Mobile users need EIP-681 formatted payment data that mobile wallets can parse from QR codes.

**Why this priority**: Enables mobile workflow (US-006 from OVERVIEW.md). Lowest priority as it serves tertiary user segment. Can be implemented after agent and browser workflows are complete.

**Independent Test**: Generate EIP-681 URI, encode as QR code, scan with mobile wallet (Rainbow, Coinbase Wallet, MetaMask Mobile) and verify transaction pre-fills correctly.

**Acceptance Scenarios**:

1. **Given** x402 payment requirements, **When** `encode_payment_for_qr` is called, **Then** returns EIP-681 URI like `ethereum:0xPayeeAddress@8453/transfer?address=0xUSDC&uint256=50000`
2. **Given** the encoded URI, **When** embedded in QR code and scanned by mobile wallet, **Then** wallet recognizes ERC-20 transfer request with correct token contract and amount
3. **Given** callback URL parameter, **When** encoding for QR, **Then** includes callback in URI parameters for post-payment notification
4. **Given** different networks, **When** encoded, **Then** uses correct chain ID in URI format and correct USDC contract address per network
5. **Given** URI length limits for QR code density, **When** encoding complex payment, **Then** optimizes URI format to fit in standard QR code size (version 10 or lower)

---

### Edge Cases

- **What happens when** payment requirement generation is called during network congestion and nonce retrieval times out?
  - Tool should retry nonce fetch (3 attempts) then return error indicating network unavailability

- **What happens when** EIP-3009 authorization has validAfter in the future?
  - Verification returns `{is_valid: false, error: "authorization not yet valid", valid_after: <timestamp>}`

- **What happens when** facilitator settlement succeeds but response parsing fails?
  - Tool stores raw response and returns `{status: "unknown", raw_response: "..."}` for manual reconciliation

- **What happens when** browser link is generated for unsupported network?
  - Tool returns error listing supported networks (base, base-sepolia, arbitrum)

- **What happens when** QR encoding is called with payment exceeding mobile wallet's max displayable amount?
  - Tool warns about large amount but still generates valid URI with amount in atomic units

- **What happens when** multiple concurrent calls attempt to settle the same payment?
  - First call proceeds, subsequent calls return cached result (idempotency via nonce tracking)

- **What happens when** user's wallet has insufficient USDC balance at settlement time?
  - Facilitator returns error, tool propagates with `{status: "failed", error: "insufficient balance", required: "50000", balance: "10000"}`

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST implement MCP server using mcp-go library with stdio transport (runs as subprocess of HTTP proxy, no separate authentication required)
- **FR-002**: System MUST register exactly 5 tools: `create_payment_requirement`, `verify_payment`, `settle_payment`, `generate_browser_link`, `encode_payment_for_qr`
- **FR-003**: Tool `create_payment_requirement` MUST generate x402-compliant JSON with fields: `x402_version`, `scheme`, `network`, `maxAmountRequired`, `payee`, `valid_until`, `nonce`
- **FR-004**: Tool `create_payment_requirement` MUST support networks: base (chain_id=8453), base-sepolia (chain_id=84532), arbitrum (chain_id=42161)
- **FR-005**: Payment requirements MUST use unique nonces to prevent replay attacks (query blockchain RPC for wallet transaction count)
- **FR-006**: Payment requirements MUST have `valid_until` timestamp exactly 5 minutes from creation time
- **FR-007**: Tool `verify_payment` MUST validate EIP-3009 authorization signatures using secp256k1 ECDSA recovery
- **FR-008**: Tool `verify_payment` MUST verify EIP-712 domain (name, version, chainId, verifyingContract) matches expected values
- **FR-009**: Tool `verify_payment` MUST check authorization time bounds (validAfter <= now < validBefore)
- **FR-010**: Tool `verify_payment` MUST recover signer address from signature and return it for balance verification
- **FR-011**: Tool `settle_payment` MUST POST verified authorization to x402 facilitator endpoint with timeout of 5 seconds
- **FR-012**: Tool `settle_payment` MUST parse facilitator response containing tx_hash and block_number
- **FR-013**: Tool `settle_payment` MUST implement idempotency using in-memory cache with TTL (10 minutes) - cache settlement results by nonce to prevent duplicate submissions
- **FR-014**: Tool `settle_payment` MUST handle facilitator errors (400 Bad Request, 500 Server Error) and return structured error responses
- **FR-015**: Tool `generate_browser_link` MUST create MetaMask deep links using format `https://link.metamask.io/send/{usdc_contract}@{chain_id}/transfer?address={payee}&uint256={amount}`
- **FR-016**: Browser links MUST include parameters: value (amount in atomic units), callback URL (URL-encoded) - callback expects tx hash as URL parameter (e.g., `?tx=0x...`)
- **FR-017**: Tool `encode_payment_for_qr` MUST generate EIP-681 URIs with format `ethereum:{usdc_contract}@{chain_id}/transfer?address={payee}&uint256={amount}` (first address is USDC token contract, address parameter is payee)
- **FR-018**: QR encodings MUST use network-specific USDC contract addresses (base: 0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913, etc.)
- **FR-019**: All tools MUST validate input parameters and return descriptive errors for invalid inputs
- **FR-020**: System MUST use configuration file (YAML/JSON) for network-specific settings (payee addresses, chain IDs, USDC contracts, facilitator URLs)

### Key Entities

- **PaymentRequirement**: x402-formatted JSON containing payment terms (amount, payee, deadline, network)
  - Attributes: x402_version (int), scheme (string), network (string), maxAmountRequired (string), payee (address), valid_until (ISO8601), nonce (string)
  - Relationships: Generated for each certification quote, consumed by verify/settle tools

- **EIP3009Authorization**: Cryptographically signed payment authorization for gasless USDC transfer
  - Attributes: from (address), to (address), value (uint256), validAfter (uint256), validBefore (uint256), nonce (bytes32), signature (bytes), v/r/s components
  - Relationships: Created by agent wallets, verified by verify_payment tool, submitted by settle_payment tool

- **FacilitatorResponse**: Result of payment settlement attempt via x402 facilitator
  - Attributes: status (settled|pending|failed), tx_hash (string), block_number (uint64), error (string), retry_after (int)
  - Relationships: Returned by settle_payment, stored for idempotency checks

- **NetworkConfiguration**: Network-specific parameters for payment processing
  - Attributes: network_name (string), chain_id (uint64), payee_address (address), usdc_contract (address), facilitator_url (string)
  - Relationships: One config per supported network, referenced by all tools

- **BrowserPaymentLink**: MetaMask deep link for browser-based payment initiation
  - Attributes: url (string), callback_url (string), expiry (ISO8601)
  - Relationships: Generated from PaymentRequirement, consumed by browser users

- **QRPaymentEncoding**: EIP-681 formatted URI for mobile wallet QR scanning
  - Attributes: eip681_uri (string), callback_url (string), estimated_qr_version (int)
  - Relationships: Generated from PaymentRequirement, encoded into QR images by qr-code-mcp-server

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All 5 MCP tools respond to requests within 100ms (excluding facilitator network I/O)
- **SC-002**: `verify_payment` achieves 100% accuracy in detecting invalid signatures during testing (0 false positives/negatives across 1000 test cases)
- **SC-003**: `settle_payment` successfully settles 95% of verified payments on first attempt (5% retry due to network issues)
- **SC-004**: Payment requirements generated by `create_payment_requirement` conform to x402 JSON schema with 100% validation pass rate
- **SC-005**: Browser links generated by `generate_browser_link` successfully pre-fill MetaMask transactions in 100% of manual tests across Chrome, Firefox, Brave
- **SC-006**: QR encodings from `encode_payment_for_qr` are successfully parsed by 3 tested mobile wallets (MetaMask Mobile, Rainbow, Coinbase Wallet) with 100% success rate
- **SC-007**: Server handles concurrent requests (10 simultaneous tool calls) without race conditions or duplicate settlement attempts
- **SC-008**: Integration test suite achieves 90%+ code coverage for all tool handlers (exceeds Constitution II minimum of 80% for all packages, maintains 100% for security-sensitive signature verification)
- **SC-009**: Settlement idempotency prevents duplicate facilitator calls - 100% of re-submitted payments return cached results within 10ms
- **SC-010**: Configuration file supports adding new networks without code changes (extensibility test)

### Quality Metrics

- **QM-001**: All tool input schemas validated against MCP specification
- **QM-002**: Error responses include actionable error messages (not just "invalid input")
- **QM-003**: No hardcoded network parameters (addresses, chain IDs) in tool code - all from config
- **QM-004**: Structured JSON logging with DEBUG/INFO/WARN/ERROR levels captures all facilitator interactions (request/response), tool invocations, and errors with context fields (tool_name, network, duration_ms)
- **QM-005**: Unit tests exist for each tool covering happy path + 3 error scenarios minimum

## Out of Scope *(optional)*

- Blockchain node interaction (delegated to x402 facilitator)
- Wallet private key management (agents bring their own keys)
- USDC token contract deployment
- Gas price estimation or optimization
- Frontend UI for payment workflows (handled by certify.ar4s.com HTTP proxy)
- Multi-token support (only USDC in v1)
- Refund processing for failed certifications

## Open Questions *(optional)*

- **Q1**: What is the exact x402 facilitator API specification? (Need documentation from x402.org)
- **Q3**: Do we need to support testnets beyond base-sepolia (e.g., arbitrum-sepolia)?
- **Q4**: What EIP-712 domain name/version should be used for signature verification? (Need to align with USDC contract)
- **Q5**: Should facilitator URL be configurable per-network or globally?
- **Q6**: How should we handle facilitator version mismatches (if they update their API)?
- **Q7**: Do mobile wallets require specific EIP-681 parameter ordering for optimal parsing?

## Dependencies *(optional)*

**External Services:**
- x402 Facilitator API (`https://x402.org/facilitator`) - for payment settlement
- Blockchain JSON-RPC endpoints (base, base-sepolia, arbitrum) - for querying wallet transaction count (nonces)

**Libraries:**
- `github.com/mark3labs/mcp-go` - MCP server framework
- `github.com/ethereum/go-ethereum` - EIP-712 signing, crypto utilities
- `github.com/btcsuite/btcd/btcec` - secp256k1 signature verification (or go-ethereum/crypto)

**Internal:**
- Configuration file (config.yaml) with network parameters
- pkg/crypto/secp256k1.go (existing from Milestone 1) - may be reused for signature ops

**Development:**
- Go 1.23+ (from Nix flake)
- Test USDC contracts on base-sepolia
- MetaMask extension for browser link testing
- Mobile wallet apps for QR encoding testing
