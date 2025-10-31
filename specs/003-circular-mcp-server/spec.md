# Feature Specification: Circular Protocol MCP Server

**Feature Branch**: `003-circular-mcp-server`
**Created**: 2025-10-30
**Status**: Draft
**Input**: User description: "Circular Protocol MCP server: Build MCP server using mcp-go with 4 tools for blockchain certification operations. Tools: get_wallet_nonce (fetches current nonce via Circular_GetWalletNonce_ API), certify_data (constructs C_TYPE_CERTIFICATE transaction with Secp256k1 signing, posts via Circular_AddTransaction_), get_transaction_status (polls transaction status until "Executed"), get_certification_proof (extracts block ID, timestamp, generates explorer URL). Must handle Circular Protocol HTTP REST API, implement transaction ID calculation (sha256 of From+To+Payload+Timestamp), support testnet and mainnet. Testing: end-to-end certification on testnet, measure confirmation time. Reference: docs/OVERVIEW.md Milestone 3 (Week 2-3), Section 2.3.2 tool schemas, Tasks T012-T017."

## Clarifications

### Session 2025-10-30

- Q: How should the MCP server access the private key for signing certification transactions? → A: Load from environment variable (CIRCULAR_CEP_TESTNET_PRIVATE_KEY for testnet, CIRCULAR_CEP_MAINNET_PRIVATE_KEY for mainnet) at server startup
- Q: What is the maximum size limit for data payloads that can be certified in a single transaction? → A: 1 MB (medium documents, certificates, proofs)
- Q: What should be the polling interval when checking transaction status with get_transaction_status? → A: Fixed 5 seconds between each poll
- Q: When the Circular Protocol API returns an error or is unavailable, what information should the system provide to the agent? → A: Structured error with: error type, HTTP status code, error message, and retry suggestion
- Q: What level of logging/observability should the MCP server provide for certification operations? → A: Standard logging (tool calls, transaction IDs, status changes, errors)

## User Scenarios & Testing

### User Story 1 - Certify Data on Blockchain (Priority: P1)

An AI agent needs to certify important data (documents, proofs, attestations) by recording them immutably on the Circular Protocol blockchain. The agent creates a certification transaction, signs it cryptographically, and submits it to the blockchain, receiving proof of certification.

**Why this priority**: This is the core value proposition - without the ability to certify data, no other functionality matters. This story delivers immediate value by enabling immutable data certification.

**Independent Test**: Can be fully tested by submitting a test data payload to testnet, receiving a transaction ID, and verifying the transaction reaches "Executed" status. Delivers complete certification capability.

**Acceptance Scenarios**:

1. **Given** an AI agent has data to certify, **When** the agent calls the certify_data tool with the data payload, **Then** the system creates a signed C_TYPE_CERTIFICATE transaction and returns a transaction ID
2. **Given** a certification transaction has been submitted, **When** the agent checks the transaction status, **Then** the system reports the current status (Pending, Verified, or Executed)
3. **Given** a certification transaction has reached "Executed" status, **When** the agent requests certification proof, **Then** the system returns block ID, timestamp, and an explorer URL for verification

---

### User Story 2 - Verify Transaction Status (Priority: P2)

An AI agent needs to monitor certification transactions to know when they have been confirmed on the blockchain. The agent polls the transaction status until it reaches "Executed" state, ensuring the certification is complete before proceeding.

**Why this priority**: Status monitoring is essential for reliable certification workflows, but an agent could manually check the blockchain explorer as a workaround. This automates and simplifies the process.

**Independent Test**: Can be fully tested by submitting a test transaction, then polling its status until "Executed", measuring the time to confirmation. Delivers complete status tracking.

**Acceptance Scenarios**:

1. **Given** a pending transaction ID, **When** the agent polls get_transaction_status, **Then** the system returns the current status without errors
2. **Given** a transaction transitions from Pending to Verified to Executed, **When** the agent polls repeatedly, **Then** the system reports each status change accurately
3. **Given** an invalid transaction ID, **When** the agent requests status, **Then** the system returns a clear error message

---

### User Story 3 - Retrieve Certification Proof (Priority: P2)

An AI agent needs to provide verifiable proof of certification to external parties. The agent retrieves the block ID, timestamp, and blockchain explorer URL for a completed certification transaction.

**Why this priority**: Proof generation is necessary for verification but could be manually constructed from transaction details. This automates proof assembly for convenience.

**Independent Test**: Can be fully tested by using a completed transaction ID to generate proof, then manually verifying the explorer URL works. Delivers complete proof generation.

**Acceptance Scenarios**:

1. **Given** an executed transaction ID, **When** the agent calls get_certification_proof, **Then** the system returns block ID, timestamp, and explorer URL
2. **Given** a transaction not yet executed, **When** the agent requests proof, **Then** the system returns an error indicating the transaction is not yet complete
3. **Given** a testnet transaction, **When** proof is generated, **Then** the explorer URL points to the testnet explorer
4. **Given** a mainnet transaction, **When** proof is generated, **Then** the explorer URL points to the mainnet explorer

---

### User Story 4 - Manage Nonce for Transactions (Priority: P3)

An AI agent needs to retrieve the current nonce for a wallet address to construct valid transactions. The nonce ensures transactions are processed in the correct order and prevents replay attacks.

**Why this priority**: While nonce management is necessary for transaction construction, it's a supporting capability. The agent could cache nonces or use defaults, though this would be less reliable.

**Independent Test**: Can be fully tested by calling get_wallet_nonce for a test wallet address and verifying the returned nonce matches the blockchain state. Delivers nonce retrieval capability.

**Acceptance Scenarios**:

1. **Given** a valid wallet address, **When** the agent calls get_wallet_nonce, **Then** the system returns the current nonce value
2. **Given** an invalid wallet address format, **When** the agent requests the nonce, **Then** the system returns a validation error
3. **Given** multiple sequential transaction submissions, **When** the agent retrieves the nonce after each, **Then** the nonce increments correctly

---

### Edge Cases

- What happens when the Circular Protocol API is temporarily unavailable or returns errors?
- How does the system handle transaction submission failures (network errors, insufficient balance, invalid signatures)?
- What happens when polling a transaction status times out before reaching "Executed"?
- How does the system handle malformed data payloads for certification?
- How does the system handle data payloads exceeding the 1 MB limit? (Answer: Reject with clear error message before submission)
- What happens when requesting proof for a transaction that failed or was rejected?
- How does the system distinguish between testnet and mainnet transactions in proof generation?
- What happens when the nonce API returns stale or incorrect values?

## Requirements

### Functional Requirements

- **FR-001**: System MUST provide a get_wallet_nonce tool that retrieves the current nonce for a wallet address via Circular_GetWalletNonce_ API
- **FR-002**: System MUST provide a certify_data tool that constructs a C_TYPE_CERTIFICATE transaction with the provided data payload
- **FR-002a**: System MUST reject data payloads exceeding 1 MB with a clear error message
- **FR-003**: System MUST sign certification transactions using Secp256k1 cryptographic signing
- **FR-003a**: System MUST load the private key from an environment variable (CIRCULAR_CEP_TESTNET_PRIVATE_KEY for testnet, CIRCULAR_CEP_MAINNET_PRIVATE_KEY for mainnet) at server startup for transaction signing
- **FR-004**: System MUST submit signed transactions to the Circular Protocol blockchain via Circular_AddTransaction_ API
- **FR-005**: System MUST calculate transaction IDs client-side as SHA-256 hash of concatenated Blockchain+From+To+Payload+Nonce+Timestamp fields (as per Circular Protocol Enterprise APIs pattern)
- **FR-006**: System MUST provide a get_transaction_status tool that polls transaction status until it reaches "Executed" state
- **FR-006a**: System MUST use a fixed 5-second interval between status polling attempts
- **FR-007**: System MUST handle transaction status transitions: Pending → Verified → Executed
- **FR-008**: System MUST provide a get_certification_proof tool that extracts block ID and timestamp from executed transactions
- **FR-009**: System MUST generate blockchain explorer URLs appropriate to the network (testnet or mainnet)
- **FR-010**: System MUST support both Circular Protocol testnet and mainnet environments
- **FR-011**: System MUST communicate with Circular Protocol HTTP REST API endpoints
- **FR-012**: System MUST validate wallet addresses before making API calls
- **FR-013**: System MUST handle API errors gracefully and return structured error responses containing: error type, HTTP status code, error message, and retry suggestion
- **FR-014**: System MUST prevent transaction resubmission when a transaction ID already exists
- **FR-015**: Certification transactions MUST be immutable once executed on the blockchain
- **FR-016**: System MUST provide standard logging that captures: tool invocations, transaction IDs, status transitions, and error events

### Key Entities

- **Certification Transaction**: A C_TYPE_CERTIFICATE transaction containing data payload, sender address, nonce, timestamp, and cryptographic signature. Represents an immutable record on the blockchain.
- **Transaction ID**: A unique identifier calculated client-side as SHA-256 hash of Blockchain+From+To+Payload+Nonce+Timestamp (per Enterprise APIs pattern). Used to track and retrieve transaction status.
- **Nonce**: A sequential counter for each wallet address ensuring transaction ordering and preventing replay attacks.
- **Certification Proof**: Contains block ID, timestamp, and explorer URL. Provides verifiable evidence that data was certified at a specific time.
- **Network Configuration**: Distinguishes between testnet and mainnet, including different API endpoints and explorer URLs.
- **Error Response**: Structured error information containing error type (e.g., "API_UNAVAILABLE", "INVALID_SIGNATURE"), HTTP status code, human-readable error message, and retry suggestion (e.g., "Retry after 30 seconds").

## Success Criteria

### Measurable Outcomes

- **SC-001**: AI agents can certify data on testnet within 60 seconds from submission to "Executed" status (measured during testing)
- **SC-002**: 100% of valid certification transactions submitted to testnet successfully reach "Executed" status
- **SC-003**: Transaction status polling completes within 5 attempts (25 seconds maximum at 5-second intervals) for 95% of transactions
- **SC-004**: Certification proof generation succeeds for 100% of executed transactions
- **SC-005**: All 4 tools (get_wallet_nonce, certify_data, get_transaction_status, get_certification_proof) pass end-to-end integration testing on testnet
- **SC-006**: System handles API errors gracefully with clear error messages in 100% of failure scenarios
- **SC-007**: Transaction ID calculation produces consistent, deterministic results for identical transaction inputs

## Assumptions

- The Circular Protocol HTTP REST API (Circular_GetWalletNonce_, Circular_AddTransaction_, and status endpoints) is available and documented
- A test wallet with sufficient balance exists for testnet certification transactions
- Secp256k1 signing is compatible with Circular Protocol's signature verification
- Transaction status progresses Pending → Verified → Executed in a reasonable timeframe (under 60 seconds on average)
- The blockchain explorer URLs follow predictable patterns for both testnet and mainnet
- C_TYPE_CERTIFICATE is a valid transaction type in the Circular Protocol
- The Circular Protocol API accepts SHA-256 transaction IDs
- Network configuration (testnet vs mainnet) can be determined from environment variables or configuration files

## Dependencies

- **Circular Protocol API Access**: Requires active HTTP REST API endpoints for both testnet and mainnet
- **Cryptographic Library**: Requires Secp256k1 implementation for transaction signing (standard Go crypto libraries assumed)
- **MCP Framework**: Depends on mcp-go framework for tool registration and execution
- **Test Wallet**: Requires a funded testnet wallet for end-to-end testing

## Out of Scope

- User interface for managing wallets or viewing certifications (agent-only interface)
- Wallet creation or private key management (assumes wallet already exists)
- Gas fee estimation or optimization
- Batch certification of multiple data payloads in a single transaction
- Historical certification queries or search functionality
- Integration with other blockchain networks beyond Circular Protocol
- Automatic retry logic for failed transactions (agent handles retry strategy)
