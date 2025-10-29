# Tasks: x402 Payment MCP Server

**Input**: Design documents from `/specs/002-x402-mcp-server/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

**Tests**: The specification requires Test-First Development (TDD) per Constitution II with 90%+ code coverage (SC-008). All test tasks MUST be written and verified to FAIL before implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Convention

All paths relative to `mcp-servers/x402-mcp-server/` per plan.md structure.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure per plan.md

- [x] T001 Create project directory structure at mcp-servers/x402-mcp-server/ with subdirs: tools/, internal/, tests/
- [x] T002 Initialize Go module with `go mod init` in mcp-servers/x402-mcp-server/
- [x] T003 [P] Add mcp-go dependency: `go get github.com/mark3labs/mcp-go`
- [x] T004 [P] Add go-ethereum dependency: `go get github.com/ethereum/go-ethereum`
- [x] T005 [P] Add yaml.v3 dependency: `go get gopkg.in/yaml.v3`
- [x] T006 [P] Create config.yaml.example in mcp-servers/x402-mcp-server/ with network configs from research.md
- [x] T007 [P] Configure golangci-lint in mcp-servers/x402-mcp-server/.golangci.yml

**Checkpoint**: Project structure ready for foundational implementation

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Configuration & Logging Infrastructure

- [x] T008 Implement Config struct and YAML loading in internal/config/config.go per data-model.md
- [x] T009 Implement NetworkConfig validation (chain IDs, addresses) in internal/config/network.go
- [x] T010 [P] Implement structured JSON logger in internal/logger/structured.go with DEBUG/INFO/WARN/ERROR levels (QM-004)
- [x] T011 [P] Create TTL-based in-memory cache in internal/cache/ttl_cache.go for settlement idempotency (FR-013)

### Tests for Foundational Infrastructure

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T012 [P] Unit test for config loading in tests/unit/config_test.go (happy path + invalid YAML)
- [x] T013 [P] Unit test for network validation in tests/unit/config_test.go (invalid chain ID, malformed address)
- [x] T014 [P] Unit test for TTL cache in tests/unit/cache_test.go (set, get, expiry, idempotency)

### EIP-712 & Crypto Utilities

- [ ] T015 Implement EIP712Domain struct and domain separator hashing in internal/eip3009/eip712.go per research.md
- [ ] T016 Implement ReceiveWithAuthorizationMessage struct in internal/eip3009/authorization.go per data-model.md
- [ ] T017 Implement EIP-712 typed data hash construction in internal/eip3009/eip712.go
- [ ] T018 [P] Unit test for EIP-712 domain hashing in tests/unit/eip3009_test.go (base, base-sepolia, arbitrum)
- [ ] T019 [P] Unit test for typed data hash in tests/unit/eip3009_test.go (known test vectors)

### Blockchain RPC Integration

- [ ] T020 Implement nonce fetcher via eth_getTransactionCount in internal/rpc/nonce_fetcher.go with 3-retry logic
- [ ] T021 Integration test for nonce fetching in tests/integration/rpc_integration_test.go (Base Sepolia testnet)

### MCP Server Entry Point

- [ ] T022 Implement MCP server main.go with stdio transport setup using mcp-go
- [ ] T023 Implement tool registration framework in main.go (empty tool handlers)
- [ ] T024 Contract test for MCP tool discovery in tests/contract/mcp_tool_schemas_test.go (verify 5 tools registered)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Generate Payment Requirements (Priority: P1) 🎯 MVP

**Goal**: Implement `create_payment_requirement` tool to generate x402-compliant payment JSON with blockchain-sourced nonces

**Independent Test**: Call tool with various amounts/networks, validate output JSON schema compliance (FR-003, FR-004)

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T025 [P] [US1] Contract test for create_payment_requirement tool schema in tests/contract/mcp_tool_schemas_test.go (input/output validation per contracts/create_payment_requirement.json)
- [ ] T026 [P] [US1] Unit test for PaymentRequirement generation in tests/unit/x402_test.go (base network, 50000 USDC, verify all fields)
- [ ] T027 [P] [US1] Unit test for nonce uniqueness in tests/unit/x402_test.go (multiple calls should generate different nonces)
- [ ] T028 [P] [US1] Unit test for invalid network handling in tests/unit/x402_test.go (expect error for unsupported network)
- [ ] T029 [P] [US1] Integration test for end-to-end payment requirement creation in tests/integration/create_payment_test.go (Base Sepolia, verify blockchain nonce fetch)

### Implementation for User Story 1

- [ ] T030 [P] [US1] Create PaymentRequirement struct in internal/x402/payment_requirement.go per data-model.md
- [ ] T031 [US1] Implement create_payment_requirement tool handler in tools/create_payment_requirement.go (FR-003, FR-005, FR-006)
- [ ] T032 [US1] Integrate nonce fetcher from internal/rpc/ into payment requirement generation
- [ ] T033 [US1] Implement valid_until timestamp generation (+5 minutes per FR-006)
- [ ] T034 [US1] Add input validation (amount > 0, network in allowlist) per FR-019
- [ ] T035 [US1] Add structured logging for tool calls (tool_name, network, amount, duration_ms)
- [ ] T036 [US1] Register create_payment_requirement tool in main.go

**Checkpoint**: At this point, User Story 1 should pass all tests and be independently usable

---

## Phase 4: User Story 2 - Verify EIP-3009 Payment Signatures (Priority: P1) 🎯 MVP

**Goal**: Implement `verify_payment` tool to validate EIP-3009 authorization signatures using secp256k1 ECDSA recovery

**Independent Test**: Generate valid/invalid signatures with test private keys, verify tool correctly identifies each (FR-007, FR-008)

### Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T037 [P] [US2] Contract test for verify_payment tool schema in tests/contract/mcp_tool_schemas_test.go (input/output validation per contracts/verify_payment.json)
- [ ] T038 [P] [US2] Unit test for valid signature verification in tests/unit/eip3009_test.go (known test vector with valid v/r/s)
- [ ] T039 [P] [US2] Unit test for invalid signature detection in tests/unit/eip3009_test.go (tampered value field)
- [ ] T040 [P] [US2] Unit test for wrong network detection in tests/unit/eip3009_test.go (signature from different chain ID)
- [ ] T041 [P] [US2] Unit test for time bound validation in tests/unit/eip3009_test.go (expired validBefore, future validAfter)
- [ ] T042 [P] [US2] Integration test with 1000 random test cases in tests/integration/signature_verification_test.go (SC-002: 100% accuracy)

### Implementation for User Story 2

- [ ] T043 [P] [US2] Create EIP3009Authorization struct in internal/eip3009/authorization.go per data-model.md
- [ ] T044 [P] [US2] Create VerifyPaymentOutput struct in internal/eip3009/authorization.go per data-model.md
- [ ] T045 [US2] Implement signature verification in internal/eip3009/signature_verifier.go using go-ethereum/crypto (FR-007)
- [ ] T046 [US2] Implement EIP-712 domain matching in internal/eip3009/signature_verifier.go (FR-008: name, version, chainId, verifyingContract)
- [ ] T047 [US2] Implement signer recovery via secp256k1 ECDSA in internal/eip3009/signature_verifier.go (FR-010)
- [ ] T048 [US2] Implement time bound validation in internal/eip3009/signature_verifier.go (FR-009)
- [ ] T049 [US2] Implement verify_payment tool handler in tools/verify_payment.go
- [ ] T050 [US2] Add input validation for authorization fields (addresses, v/r/s format) per FR-019
- [ ] T051 [US2] Add structured logging for verification attempts (is_valid, signer_address, duration_ms)
- [ ] T052 [US2] Register verify_payment tool in main.go

**Checkpoint**: At this point, User Stories 1 AND 2 (MVP primitives) should be fully functional

---

## Phase 5: User Story 3 - Settle Payments via Facilitator (Priority: P2)

**Goal**: Implement `settle_payment` tool to submit verified authorizations to x402 facilitator for on-chain settlement

**Independent Test**: Test against x402 facilitator testnet endpoint with test USDC, verify HTTP interaction and idempotency (FR-011, FR-013)

### Tests for User Story 3

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T053 [P] [US3] Contract test for settle_payment tool schema in tests/contract/mcp_tool_schemas_test.go (input/output validation per contracts/settle_payment.json)
- [ ] T054 [P] [US3] Unit test for facilitator request construction in tests/unit/x402_test.go (verify POST body format from research.md)
- [ ] T055 [P] [US3] Unit test for facilitator success response parsing in tests/unit/x402_test.go (tx_hash, block_number)
- [ ] T056 [P] [US3] Unit test for facilitator error handling in tests/unit/x402_test.go (400, 500, timeout)
- [ ] T057 [P] [US3] Unit test for idempotency cache in tests/unit/x402_test.go (same nonce returns cached result)
- [ ] T058 [US3] Integration test for facilitator settlement in tests/integration/facilitator_integration_test.go (Base Sepolia testnet, SC-003)

### Implementation for User Story 3

- [ ] T059 [P] [US3] Create FacilitatorResponse struct in internal/x402/facilitator_client.go per data-model.md
- [ ] T060 [P] [US3] Create SettlePaymentInput/Output structs in internal/x402/facilitator_client.go per data-model.md
- [ ] T061 [US3] Implement HTTP client for x402 facilitator API in internal/x402/facilitator_client.go with 5s timeout (FR-011)
- [ ] T062 [US3] Implement facilitator POST request formatting in internal/x402/facilitator_client.go per research.md API spec
- [ ] T063 [US3] Implement facilitator response parsing in internal/x402/facilitator_client.go (success, pending, failed states)
- [ ] T064 [US3] Implement settle_payment tool handler in tools/settle_payment.go (FR-012, FR-014)
- [ ] T065 [US3] Integrate verify_payment signature check before settlement (dependency on Phase 4)
- [ ] T066 [US3] Integrate idempotency cache from internal/cache/ (FR-013: 10-minute TTL by nonce)
- [ ] T067 [US3] Add error handling for facilitator errors (400/500/timeout) per FR-014
- [ ] T068 [US3] Add structured logging for settlement attempts (status, tx_hash, retry_after, duration_ms)
- [ ] T069 [US3] Register settle_payment tool in main.go

**Checkpoint**: Core payment flow (create → verify → settle) now complete

---

## Phase 6: User Story 4 - Generate Browser Payment Links (Priority: P2)

**Goal**: Implement `generate_browser_link` tool to create MetaMask deep links for browser-based payments

**Independent Test**: Generate link, verify URL format, test in browser with MetaMask extension (FR-015, FR-016)

### Tests for User Story 4

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T070 [P] [US4] Contract test for generate_browser_link tool schema in tests/contract/mcp_tool_schemas_test.go (input/output validation per contracts/generate_browser_link.json)
- [ ] T071 [P] [US4] Unit test for MetaMask link generation in tests/unit/browser_test.go (base network, verify URL format from research.md)
- [ ] T072 [P] [US4] Unit test for callback URL encoding in tests/unit/browser_test.go (special characters properly escaped)
- [ ] T073 [P] [US4] Unit test for multi-network link generation in tests/unit/browser_test.go (base, base-sepolia, arbitrum with correct chain IDs)
- [ ] T074 [US4] Manual integration test for MetaMask pre-fill in tests/integration/browser_link_integration_test.go (SC-005: Chrome, Firefox, Brave)

### Implementation for User Story 4

- [ ] T075 [P] [US4] Create BrowserPaymentLink struct in internal/browser/link_generator.go per data-model.md
- [ ] T076 [US4] Implement MetaMask deep link formatter in internal/browser/link_generator.go per research.md (FR-015)
- [ ] T077 [US4] Implement callback URL encoding in internal/browser/link_generator.go (FR-016)
- [ ] T078 [US4] Implement generate_browser_link tool handler in tools/generate_browser_link.go
- [ ] T079 [US4] Add input validation (payment requirement fields, callback URL format) per FR-019
- [ ] T080 [US4] Add structured logging for link generation (network, chain_id, callback_url, duration_ms)
- [ ] T081 [US4] Register generate_browser_link tool in main.go

**Checkpoint**: Browser payment workflow enabled

---

## Phase 7: User Story 5 - Encode Payments for QR Codes (Priority: P3)

**Goal**: Implement `encode_payment_for_qr` tool to generate EIP-681 URIs for mobile wallet QR scanning

**Independent Test**: Generate URI, encode as QR, scan with mobile wallets (MetaMask Mobile, Rainbow, Coinbase Wallet) per SC-006

### Tests for User Story 5

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T082 [P] [US5] Contract test for encode_payment_for_qr tool schema in tests/contract/mcp_tool_schemas_test.go (input/output validation per contracts/encode_payment_for_qr.json)
- [ ] T083 [P] [US5] Unit test for EIP-681 URI encoding in tests/unit/eip681_test.go (base network, verify URI format from research.md)
- [ ] T084 [P] [US5] Unit test for multi-network URI encoding in tests/unit/eip681_test.go (correct USDC contract per network per FR-018)
- [ ] T085 [P] [US5] Unit test for callback parameter handling in tests/unit/eip681_test.go (optional callback URL)
- [ ] T086 [P] [US5] Unit test for QR version estimation in tests/unit/eip681_test.go (URI length → QR version 10 target)
- [ ] T087 [US5] Manual integration test for mobile wallet parsing in tests/integration/qr_integration_test.go (SC-006: 3 wallets)

### Implementation for User Story 5

- [ ] T088 [P] [US5] Create QRPaymentEncoding struct in internal/eip681/uri_encoder.go per data-model.md
- [ ] T089 [US5] Implement EIP-681 URI formatter in internal/eip681/uri_encoder.go per research.md (FR-017)
- [ ] T090 [US5] Implement network-specific USDC contract mapping in internal/eip681/uri_encoder.go (FR-018)
- [ ] T091 [US5] Implement QR version estimation (URI length → QR version) in internal/eip681/uri_encoder.go
- [ ] T092 [US5] Implement encode_payment_for_qr tool handler in tools/encode_payment_for_qr.go
- [ ] T093 [US5] Add input validation (payment requirement fields, optional callback) per FR-019
- [ ] T094 [US5] Add structured logging for QR encoding (network, usdc_contract, estimated_qr_version, duration_ms)
- [ ] T095 [US5] Register encode_payment_for_qr tool in main.go

**Checkpoint**: All 5 user stories (all payment workflows) now complete

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories and finalize deliverables

### Performance & Concurrency Testing

- [ ] T096 [P] Load test for concurrent tool calls in tests/integration/concurrency_test.go (SC-007: 10 simultaneous calls, no race conditions)
- [ ] T097 [P] Performance test for tool response times in tests/integration/performance_test.go (SC-001: <100ms excluding I/O)
- [ ] T098 Run all tests with race detector: `go test -race ./...`

### Coverage & Quality

- [ ] T099 Generate coverage report: `go test ./... -coverprofile=coverage.out` (SC-008: 90%+ target)
- [ ] T100 Review coverage report: `go tool cover -html=coverage.out` and add missing tests if below 90%
- [ ] T101 Run golangci-lint: `golangci-lint run ./...` and fix issues

### Documentation & Configuration

- [ ] T102 [P] Validate config.yaml.example has all required fields from data-model.md
- [ ] T103 [P] Create README.md in mcp-servers/x402-mcp-server/ with quickstart instructions from quickstart.md
- [ ] T104 [P] Add inline documentation (godoc) for all exported functions
- [ ] T105 Validate quickstart.md examples work end-to-end

### Security Hardening

- [ ] T106 Review signature verification for timing attacks - use crypto/subtle.ConstantTimeCompare for signature component comparisons, verify via code review that no early returns leak timing info
- [ ] T107 Validate all address inputs are checksummed or case-insensitive
- [ ] T108 Ensure no sensitive data (private keys, full signatures) logged (QM-004)

### Final Validation

- [ ] T109 Run acceptance scenarios from spec.md for all 5 user stories manually
- [ ] T110 Verify all 5 tools registered and discoverable via `tools/list` MCP request
- [ ] T111 Test extensibility: Add 'polygon' network (chainId 137, USDC 0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359) to config.yaml and verify MCP server works without code changes (SC-010)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phases 3-7)**: All depend on Foundational phase completion
  - Phase 3 (US1): Independent - can start after Foundational
  - Phase 4 (US2): Independent - can start after Foundational
  - Phase 5 (US3): Depends on Phase 4 (needs verify_payment for pre-settlement check)
  - Phase 6 (US4): Depends on Phase 3 (needs payment requirements from US1)
  - Phase 7 (US5): Depends on Phase 3 (needs payment requirements from US1)
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - Independent from US1
- **User Story 3 (P2)**: DEPENDS ON User Story 2 (verify_payment integration) - Start after Phase 4
- **User Story 4 (P2)**: DEPENDS ON User Story 1 (payment requirements input) - Start after Phase 3
- **User Story 5 (P3)**: DEPENDS ON User Story 1 (payment requirements input) - Start after Phase 3

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD per Constitution II)
- Structs/models before services
- Services before tool handlers
- Tool handlers before registration in main.go
- Story complete before moving to next priority

### Parallel Opportunities

**Phase 1 (Setup)**: Tasks T003, T004, T005, T006, T007 can run in parallel (different dependency fetches)

**Phase 2 (Foundational)**: Tasks T010, T011, T012, T013, T014, T018, T019 can run in parallel (different packages)

**After Foundational Complete**:
- User Story 1 (Phase 3) and User Story 2 (Phase 4) can run in parallel (independent implementations)
- Once US1 done: US4 and US5 can start in parallel
- Once US2 done: US3 can start

**Within User Story Test Phases**: All tests marked [P] can run in parallel (different test files)

**Within User Story Implementation**: Tasks marked [P] can run in parallel (different files/packages)

---

## Parallel Example: User Story 1 (Payment Requirements)

```bash
# Tests (can all run in parallel after writing):
Task T025: Contract test for create_payment_requirement tool schema
Task T026: Unit test for PaymentRequirement generation
Task T027: Unit test for nonce uniqueness
Task T028: Unit test for invalid network handling
Task T029: Integration test for end-to-end payment requirement

# Implementation (T030 and T031-T036 can overlap):
Task T030: Create PaymentRequirement struct (independent file)
Task T031-T036: Implement tool handler, validation, logging (depends on T030)
```

---

## Parallel Example: Foundational Phase

```bash
# Can run in parallel:
Task T010: Structured logger (internal/logger/)
Task T011: TTL cache (internal/cache/)
Task T012: Config test (tests/unit/config_test.go)
Task T013: Network validation test (tests/unit/config_test.go)
Task T014: Cache test (tests/unit/cache_test.go)
Task T018: EIP-712 domain test (tests/unit/eip3009_test.go)
Task T019: Typed data hash test (tests/unit/eip3009_test.go)
```

---

## Implementation Strategy

### MVP First (User Stories 1 & 2 Only - Core Payment Primitives)

1. Complete Phase 1: Setup (T001-T007)
2. Complete Phase 2: Foundational (T008-T024) - CRITICAL - blocks all stories
3. Complete Phase 3: User Story 1 - Payment Requirements (T025-T036)
4. Complete Phase 4: User Story 2 - Signature Verification (T037-T052)
5. **STOP and VALIDATE**: Test US1 + US2 independently (create requirements + verify signatures)
6. Deploy/demo MVP (agent can generate requirements and verify payments)

**MVP Deliverable**: Agents can generate payment requirements and verify payment signatures (foundational capabilities)

### Incremental Delivery (Add Settlement, Browser, Mobile)

1. Foundation + US1 + US2 → MVP deployed
2. Add Phase 5: User Story 3 - Settlement (T053-T069) → Test independently → Deploy (complete agent workflow)
3. Add Phase 6: User Story 4 - Browser Links (T070-T081) → Test independently → Deploy (browser support)
4. Add Phase 7: User Story 5 - QR Encoding (T082-T095) → Test independently → Deploy (mobile support)
5. Complete Phase 8: Polish (T096-T111) → Final release

### Parallel Team Strategy

With multiple developers after Foundational phase complete:

**Stage 1 (MVP - P1 Stories)**:
- Developer A: Phase 3 (User Story 1 - Payment Requirements)
- Developer B: Phase 4 (User Story 2 - Signature Verification)
- Stories complete and integrate independently

**Stage 2 (P2 Stories)**:
- Developer A: Phase 5 (User Story 3 - Settlement) - after US2 complete
- Developer B: Phase 6 (User Story 4 - Browser Links) - after US1 complete

**Stage 3 (P3 Stories)**:
- Developer A or B: Phase 7 (User Story 5 - QR Encoding) - after US1 complete

**Stage 4 (Polish)**:
- All developers: Phase 8 together

---

## Notes

- **[P]** tasks = different files, no dependencies, can run in parallel
- **[Story]** label maps task to specific user story for traceability
- Each user story should be independently completable and testable per Constitution
- **TDD Non-Negotiable**: Verify tests FAIL before implementing (Constitution II)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- **Coverage Requirement**: 90%+ code coverage (SC-008), 100% for signature verification (security-sensitive)
- **Concurrency**: All tests must pass with `-race` flag (SC-007)
- Follow go-ethereum/crypto patterns for EIP-712 and secp256k1 operations
