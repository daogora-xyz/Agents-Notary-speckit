# Implementation Plan: Circular Protocol MCP Server

**Branch**: `003-circular-mcp-server` | **Date**: 2025-10-30 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-circular-mcp-server/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build an MCP server that provides 4 tools for AI agents to perform blockchain certification operations on Circular Protocol: get_wallet_nonce (fetch nonce), certify_data (create signed C_TYPE_CERTIFICATE transactions), get_transaction_status (poll until "Executed"), and get_certification_proof (generate proof with block ID, timestamp, explorer URL). Uses mcp-go framework, Secp256k1 signing, HTTP REST API integration via NAG (Network Access Gateway) discovery, client-side transaction ID calculation (SHA-256 hash), and supports testnet/mainnet environments.

**Technical Approach**: Implements Circular Protocol Enterprise APIs pattern with dynamic NAG URL discovery at startup (query https://circularlabs.io/network/getNAG?network={testnet|mainnet} to get base API URL), then construct Enterprise API endpoints as {NAG_URL}Circular_{MethodName}_{network}. For testnet development, uses sandbox blockchain ID 0x8a20baa40c45dc5055aeb26197c203e576ef389d9acb171bd62da11dc5ad72b2.

## Technical Context

**Language/Version**: Go 1.23+
**Primary Dependencies**: mcp-go (github.com/mark3labs/mcp-go), crypto/ecdsa (Secp256k1), net/http (Circular Protocol Enterprise REST API client with NAG discovery)
**Storage**: In-memory only (no persistent storage required for MCP server)
**Testing**: Go testing framework with contract, integration, and unit test coverage
**Target Platform**: Linux server (stdio transport for MCP host communication)
**Project Type**: Single MCP server (no UI, backend-only)
**Performance Goals**: <5 second response time per tool call, handle 100 concurrent tool invocations
**Constraints**: 60 second max certification time (submission to "Executed"), 1 MB payload size limit, 5-second fixed polling interval
**Scale/Scope**: 4 MCP tools, 2 network environments (testnet/mainnet), stateless server design

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Specification-Driven Development ✅

- ✅ Spec created via `/speckit.specify` with complete user stories
- ✅ User stories prioritized (P1: Certify Data, P2: Status/Proof, P3: Nonce)
- ✅ Each user story independently testable
- ✅ Clarifications completed via `/speckit.clarify` (5 questions answered)
- ✅ Implementation blocked until planning complete

**Status**: PASS - Full spec-kit workflow followed

### II. Test-First Development (NON-NEGOTIABLE) ✅

- ✅ Will write contract tests for all 4 MCP tool schemas
- ✅ Will write integration tests for end-to-end certification flow
- ✅ Will write unit tests for transaction signing, ID calculation, error handling
- ✅ Target 90%+ coverage (blockchain certification = critical path)
- ✅ TDD cycle: Red (write failing test) → Green (implement) → Refactor

**Status**: PASS - TDD workflow will be enforced in tasks.md

### III. MCP Architecture ✅

- ✅ MCP server uses stdio transport (not HTTP server)
- ✅ Tools use JSON Schema for input validation
- ✅ Server is stateless (no state between tool calls)
- ✅ Error responses will include structured error format (type, status, message, retry suggestion)
- ✅ Tool discovery via standard MCP protocol

**Status**: PASS - Aligns with MCP protocol standards

### IV. Security-First Design ✅

- ✅ Private key loaded from environment variable (CIRCULAR_CEP_TESTNET_PRIVATE_KEY for testnet, CIRCULAR_CEP_MAINNET_PRIVATE_KEY for mainnet)
- ✅ Private key never logged or exposed in responses
- ✅ Secp256k1 signature verification for transaction signing
- ✅ Input validation: 1 MB payload limit, wallet address format validation
- ✅ Transaction ID deduplication prevents replay

**Status**: PASS - Security requirements met

### V. Observability & Monitoring ✅

- ✅ Standard structured logging: tool calls, transaction IDs, status transitions, errors
- ✅ Will use JSON logging format
- ✅ No sensitive data in logs (private keys, full signatures)
- ✅ Correlation via transaction IDs

**Status**: PASS - Logging requirements defined

### VI. Blockchain Integration Standards ✅

- ✅ Uses Secp256k1 signing (Circular Protocol Enterprise APIs compatible - standard curve)
- ✅ Transaction status polling until "Executed" (5s interval, 60s timeout)
- ✅ Testnet and mainnet support (via config.yaml networks with NAG discovery)
- ✅ Circular Protocol C_TYPE_CERTIFICATE transactions (confirmed via Enterprise API research)
- ✅ CIRX fee model: $0.001-$0.035 per transaction, automatic deduction from sender wallet
- ✅ Retry logic designed: exponential backoff for API errors, 3 max attempts
- ✅ Transaction ID handling: calculated CLIENT-SIDE using SHA-256(Blockchain+From+To+Payload+Nonce+Timestamp) per Enterprise APIs pattern (cross-verified with Go and NodeJS implementations)
- ✅ REST API architecture: Go HTTP client for Circular Enterprise APIs with NAG discovery

**Status**: PASS - All blockchain integration requirements resolved via research.md

## Project Structure

### Documentation (this feature)

```text
specs/003-circular-mcp-server/
├── spec.md              # Feature specification (✅ completed)
├── plan.md              # This file (/speckit.plan command output - ✅ completed)
├── research.md          # Phase 0 output (✅ completed)
├── data-model.md        # Phase 1 output (✅ completed)
├── quickstart.md        # Phase 1 output (✅ completed)
├── contracts/           # Phase 1 output (✅ completed - 4 JSON schemas + README)
│   ├── get_wallet_nonce.json
│   ├── certify_data.json
│   ├── get_transaction_status.json
│   ├── get_certification_proof.json
│   └── README.md
└── tasks.md             # Phase 2 output (/speckit.tasks command - NEXT STEP)
```

### Source Code (repository root)

```text
mcp-servers/circular-protocol-mcp-server/
├── cmd/
│   └── server/
│       └── main.go                # Server entrypoint
├── internal/
│   ├── config/
│   │   ├── config.go              # Network configuration (testnet/mainnet)
│   │   └── network.go             # Network-specific settings
│   ├── circular/
│   │   ├── client.go              # Circular Protocol HTTP client
│   │   ├── transaction.go         # C_TYPE_CERTIFICATE transaction builder
│   │   ├── signer.go              # Secp256k1 transaction signing
│   │   └── transaction_id.go      # SHA-256 hash calculation
│   ├── server/
│   │   └── server.go              # MCP server initialization
│   └── logger/
│       └── structured.go          # Structured JSON logging
├── tools/
│   ├── get_wallet_nonce.go        # MCP tool: nonce retrieval
│   ├── certify_data.go            # MCP tool: certification
│   ├── get_transaction_status.go  # MCP tool: status polling
│   └── get_certification_proof.go # MCP tool: proof generation
├── tests/
│   ├── contract/
│   │   ├── tool_schemas_test.go   # Verify MCP tool JSON schemas
│   │   ├── get_wallet_nonce_test.go
│   │   ├── certify_data_test.go
│   │   ├── get_transaction_status_test.go
│   │   └── get_certification_proof_test.go
│   ├── integration/
│   │   ├── certification_flow_test.go  # End-to-end testnet certification
│   │   └── status_polling_test.go      # Status polling behavior
│   └── unit/
│       ├── transaction_id_test.go      # SHA-256 calculation
│       ├── signer_test.go              # Secp256k1 signing
│       ├── config_test.go              # Configuration loading
│       └── client_test.go              # HTTP client error handling
├── config.yaml.example      # Example configuration
├── go.mod
├── go.sum
└── README.md
```

**Structure Decision**: Single MCP server project following Go standard layout. Uses `cmd/` for entrypoint, `internal/` for private packages, `tools/` for MCP tool implementations, and `tests/` for contract/integration/unit tests. Mirrors the x402-mcp-server structure for consistency.

## Complexity Tracking

**No violations** - All constitution requirements met.

## Phase 0: Research & Unknowns

### Research Tasks

1. **Circular Protocol API Documentation**
   - **Unknown**: Complete API specification for Circular_GetWalletNonce_, Circular_AddTransaction_, and transaction status endpoints
   - **Research**: Locate official Circular Protocol API docs or OpenAPI spec
   - **Questions**: Base URL for testnet/mainnet? Authentication requirements? Rate limits?

2. **C_TYPE_CERTIFICATE Transaction Format**
   - **Unknown**: Exact fields and structure for C_TYPE_CERTIFICATE transactions
   - **Research**: Find Circular Protocol transaction type specification
   - **Questions**: Required fields? Optional fields? Encoding format?

3. **Transaction ID Calculation**
   - **Unknown**: Whether SHA-256(From+To+Payload+Timestamp) is the official Circular Protocol method
   - **Research**: Verify transaction ID calculation in Circular Protocol docs
   - **Questions**: Field concatenation order? Encoding (hex, base64, raw bytes)?

4. **CIRX Fee Model**
   - **Unknown**: Fee amount for certification transactions, payment method
   - **Research**: Circular Protocol fee structure documentation
   - **Questions**: Fixed fee per tx? Dynamic? Paid from sender wallet?

5. **Blockchain Explorer URLs**
   - **Unknown**: Explorer URL patterns for testnet and mainnet
   - **Research**: Find official Circular Protocol block explorer
   - **Questions**: URL format? Does tx ID map directly to explorer path?

6. **Secp256k1 Compatibility**
   - **Unknown**: Specific Secp256k1 curve parameters used by Circular Protocol
   - **Research**: Verify Go crypto/ecdsa compatibility
   - **Questions**: Standard secp256k1 curve? Custom parameters?

7. **Transaction Status Lifecycle**
   - **Unknown**: Detailed state machine for Pending → Verified → Executed
   - **Research**: Circular Protocol transaction finality documentation
   - **Questions**: Average time in each state? Can tx skip "Verified"? Failure states?

8. **Wallet Address Format**
   - **Unknown**: Address format and validation rules
   - **Research**: Circular Protocol address specification
   - **Questions**: Hex? Checksum? Length? Example addresses?

### Best Practices Research

1. **MCP Server Error Handling Patterns**
   - Research: Best practices for structured error responses in mcp-go
   - Goal: Define standard error format across all 4 tools

2. **Secp256k1 Signing in Go**
   - Research: crypto/ecdsa best practices for blockchain signing
   - Goal: Ensure deterministic, secure signature generation

3. **HTTP Client Retry Logic**
   - Research: Exponential backoff patterns for blockchain API calls
   - Goal: Define retry strategy for Circular Protocol API failures

## Phase 1: Design Outputs

### Artifacts Generated ✅

1. ✅ **research.md** - 8 research tasks completed, all blockchain integration questions resolved
2. ✅ **data-model.md** - 6 core entities defined with validation rules and data flows
3. ✅ **contracts/** - JSON Schema for all 4 MCP tools with error definitions
4. ✅ **quickstart.md** - Complete setup guide with 4 usage examples
5. ✅ **Update CLAUDE.md** - Added Go 1.23+, mcp-go, crypto/ecdsa, in-memory storage

**Key Decisions**:
- Use Circular Protocol Enterprise APIs via Go HTTP client (not SDK)
- Transaction IDs assigned by blockchain (not calculated client-side)
- Support testnet and mainnet only (exclude devnet from MVP)
- Stateless server design (no persistent storage)
- Exponential backoff retry logic for API errors

## Next Steps

1. ✅ Phase 0: Generate research.md (resolved 8 unknowns + 3 best practices)
2. ✅ Phase 1: Generate data-model.md, contracts/, quickstart.md (all design artifacts complete)
3. ✅ Phase 1: Update agent context via `.specify/scripts/bash/update-agent-context.sh claude`
4. ✅ Re-evaluate Constitution Check post-design (Section VI now PASS)
5. ⏳ **Phase 2**: Generate tasks.md via `/speckit.tasks` (NEXT COMMAND - separate invocation)
