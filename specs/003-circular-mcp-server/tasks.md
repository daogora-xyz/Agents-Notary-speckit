# Tasks: Circular Protocol MCP Server

**Input**: Design documents from `/specs/003-circular-mcp-server/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: This feature follows TDD (Test-Driven Development) per Constitution Section II. All test tasks are REQUIRED and must be written BEFORE implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

Project root: `mcp-servers/circular-protocol-mcp-server/`
- `cmd/server/` - Server entrypoint
- `internal/` - Private packages (config, circular client, server, logger)
- `tools/` - MCP tool implementations
- `tests/` - Contract, integration, and unit tests

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Create project directory structure per plan.md in mcp-servers/circular-protocol-mcp-server/
- [ ] T002 Initialize Go module with go.mod and go.sum in mcp-servers/circular-protocol-mcp-server/
- [ ] T003 [P] Add mcp-go dependency (github.com/mark3labs/mcp-go) to go.mod
- [ ] T004 [P] Create config.yaml.example with testnet/mainnet network configurations in mcp-servers/circular-protocol-mcp-server/
- [ ] T005 [P] Create .gitignore for Go project (exclude config.yaml, .env, vendor/, binaries)
- [ ] T006 [P] Create README.md with project overview and quickstart reference

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Configuration & Logging

- [ ] T007 Implement NetworkConfig struct and YAML loader in internal/config/config.go
- [ ] T008 [P] Implement NAG discovery client for dynamic API URL resolution in internal/config/network.go
- [ ] T009 [P] Implement structured JSON logger in internal/logger/structured.go
- [ ] T010 [P] Implement private key loader from environment variables in internal/config/keys.go

### Core Circular Protocol Client

- [ ] T011 Implement HTTP client with retry logic in internal/circular/client.go
- [ ] T012 [P] Implement HexFix utility function in internal/circular/utils.go
- [ ] T013 [P] Implement ErrorResponse struct and constructor in internal/circular/errors.go
- [ ] T014 Implement NAG URL discovery at startup in internal/circular/client.go

### Cryptography & Transaction Handling

- [ ] T015 [P] Implement Secp256k1 transaction signing in internal/circular/signer.go
- [ ] T016 [P] Implement client-side transaction ID calculation (SHA-256) in internal/circular/transaction_id.go
- [ ] T017 [P] Implement timestamp formatting (YYYY:MM:DD-HH:MM:SS) in internal/circular/utils.go
- [ ] T018 Implement CertificationTransaction struct in internal/circular/transaction.go

### MCP Server Infrastructure

- [ ] T019 Implement MCP server initialization in internal/server/server.go
- [ ] T020 Create server entrypoint with stdio transport in cmd/server/main.go

### Foundational Unit Tests

- [ ] T021 [P] Unit test for config loading and validation in tests/unit/config_test.go
- [ ] T022 [P] Unit test for NAG discovery in tests/unit/network_test.go
- [ ] T023 [P] Unit test for private key loading in tests/unit/keys_test.go
- [ ] T024 [P] Unit test for HTTP client retry logic in tests/unit/client_test.go
- [ ] T025 [P] Unit test for Secp256k1 signing in tests/unit/signer_test.go
- [ ] T026 [P] Unit test for transaction ID calculation in tests/unit/transaction_id_test.go
- [ ] T027 [P] Unit test for timestamp formatting in tests/unit/utils_test.go

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 4 - Manage Nonce (Priority: P3) 🔧 Foundation

**Goal**: Enable AI agents to retrieve current wallet nonce for transaction construction

**Independent Test**: Call get_wallet_nonce for test wallet address and verify nonce matches blockchain state

**Why First**: US1 (Certify Data) depends on nonce retrieval, so this must be implemented before core certification

### Contract Tests for User Story 4 (TDD - Write FIRST)

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T028 [P] [US4] Contract test: get_wallet_nonce JSON schema validation in tests/contract/get_wallet_nonce_test.go
- [ ] T029 [P] [US4] Contract test: get_wallet_nonce with valid address returns nonce in tests/contract/get_wallet_nonce_test.go
- [ ] T030 [P] [US4] Contract test: get_wallet_nonce with invalid address returns INVALID_ADDRESS error in tests/contract/get_wallet_nonce_test.go

### Implementation for User Story 4

- [ ] T031 [US4] Implement Circular_GetWalletNonce_ API call in internal/circular/client.go
- [ ] T032 [US4] Implement Nonce struct and validation in internal/circular/nonce.go
- [ ] T033 [US4] Implement get_wallet_nonce MCP tool in tools/get_wallet_nonce.go
- [ ] T034 [US4] Register get_wallet_nonce tool with MCP server in internal/server/server.go
- [ ] T035 [US4] Add address format validation in internal/circular/address.go

### Integration Tests for User Story 4

- [ ] T036 [US4] Integration test: Retrieve nonce from testnet for configured wallet in tests/integration/nonce_test.go
- [ ] T037 [US4] Integration test: Handle API_UNAVAILABLE error gracefully in tests/integration/nonce_test.go

**Checkpoint**: Nonce retrieval fully functional - US1 can now be implemented

---

## Phase 4: User Story 1 - Certify Data (Priority: P1) 🎯 MVP

**Goal**: Enable AI agents to certify important data by recording it immutably on Circular Protocol blockchain

**Independent Test**: Submit test data payload to testnet, receive transaction ID, verify transaction reaches "Executed" status

### Contract Tests for User Story 1 (TDD - Write FIRST)

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T038 [P] [US1] Contract test: certify_data JSON schema validation in tests/contract/certify_data_test.go
- [ ] T039 [P] [US1] Contract test: certify_data with valid payload returns transaction_id in tests/contract/certify_data_test.go
- [ ] T040 [P] [US1] Contract test: certify_data with >1MB payload returns INVALID_INPUT error in tests/contract/certify_data_test.go
- [ ] T041 [P] [US1] Contract test: certify_data with insufficient balance returns INSUFFICIENT_BALANCE error in tests/contract/certify_data_test.go

### Implementation for User Story 1

- [ ] T042 [P] [US1] Implement CertificationTransaction validation in internal/circular/transaction.go
- [ ] T043 [P] [US1] Implement TransactionStatus struct in internal/circular/status.go
- [ ] T044 [US1] Implement transaction builder with nonce fetching in internal/circular/transaction.go
- [ ] T045 [US1] Implement Circular_AddTransaction_ API call in internal/circular/client.go
- [ ] T046 [US1] Implement certify_data MCP tool in tools/certify_data.go
- [ ] T047 [US1] Register certify_data tool with MCP server in internal/server/server.go
- [ ] T048 [US1] Add payload size validation (1 MB limit) in tools/certify_data.go
- [ ] T049 [US1] Add hex encoding for data payloads in internal/circular/utils.go

### Integration Tests for User Story 1

- [ ] T050 [US1] Integration test: End-to-end certification on testnet (submit + verify executed) in tests/integration/certification_flow_test.go
- [ ] T051 [US1] Integration test: Measure confirmation time (<60 seconds target) in tests/integration/certification_flow_test.go
- [ ] T052 [US1] Integration test: Transaction ID calculation matches API response in tests/integration/certification_flow_test.go

**Checkpoint**: Core certification functionality complete - MVP ready for testing

---

## Phase 5: User Story 2 - Verify Transaction Status (Priority: P2)

**Goal**: Enable AI agents to monitor certification transactions until blockchain confirmation

**Independent Test**: Submit test transaction, poll status until "Executed", measure time to confirmation

### Contract Tests for User Story 2 (TDD - Write FIRST)

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T053 [P] [US2] Contract test: get_transaction_status JSON schema validation in tests/contract/get_transaction_status_test.go
- [ ] T054 [P] [US2] Contract test: get_transaction_status with valid tx_id returns status in tests/contract/get_transaction_status_test.go
- [ ] T055 [P] [US2] Contract test: get_transaction_status with invalid tx_id returns INVALID_INPUT error in tests/contract/get_transaction_status_test.go
- [ ] T056 [P] [US2] Contract test: get_transaction_status polling timeout returns TRANSACTION_TIMEOUT error in tests/contract/get_transaction_status_test.go

### Implementation for User Story 2

- [ ] T057 [P] [US2] Implement Circular_GetTransactionbyID_ API call in internal/circular/client.go
- [ ] T058 [P] [US2] Implement TransactionStatus lifecycle (Pending→Verified→Executed) in internal/circular/status.go
- [ ] T059 [US2] Implement status polling logic with 5-second interval in internal/circular/status.go
- [ ] T060 [US2] Implement get_transaction_status MCP tool in tools/get_transaction_status.go
- [ ] T061 [US2] Register get_transaction_status tool with MCP server in internal/server/server.go
- [ ] T062 [US2] Add timeout handling (60 seconds max) in tools/get_transaction_status.go

### Integration Tests for User Story 2

- [ ] T063 [US2] Integration test: Poll transaction from Pending to Executed in tests/integration/status_polling_test.go
- [ ] T064 [US2] Integration test: Verify status transitions are accurate in tests/integration/status_polling_test.go
- [ ] T065 [US2] Integration test: Confirm 95% of transactions complete within 60 seconds in tests/integration/status_polling_test.go

**Checkpoint**: Transaction monitoring fully functional

---

## Phase 6: User Story 3 - Retrieve Certification Proof (Priority: P2)

**Goal**: Enable AI agents to generate verifiable proof of certification for external parties

**Independent Test**: Use completed transaction ID to generate proof, manually verify explorer URL works

### Contract Tests for User Story 3 (TDD - Write FIRST)

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T066 [P] [US3] Contract test: get_certification_proof JSON schema validation in tests/contract/get_certification_proof_test.go
- [ ] T067 [P] [US3] Contract test: get_certification_proof with executed tx returns proof in tests/contract/get_certification_proof_test.go
- [ ] T068 [P] [US3] Contract test: get_certification_proof with pending tx returns error in tests/contract/get_certification_proof_test.go
- [ ] T069 [P] [US3] Contract test: get_certification_proof testnet URL points to testnet explorer in tests/contract/get_certification_proof_test.go

### Implementation for User Story 3

- [ ] T070 [P] [US3] Implement CertificationProof struct in internal/circular/proof.go
- [ ] T071 [P] [US3] Implement explorer URL generation (testnet/mainnet) in internal/circular/proof.go
- [ ] T072 [US3] Implement proof generation logic in internal/circular/proof.go
- [ ] T073 [US3] Implement get_certification_proof MCP tool in tools/get_certification_proof.go
- [ ] T074 [US3] Register get_certification_proof tool with MCP server in internal/server/server.go
- [ ] T075 [US3] Add validation for executed status requirement in tools/get_certification_proof.go

### Integration Tests for User Story 3

- [ ] T076 [US3] Integration test: Generate proof for testnet transaction and verify explorer URL in tests/integration/proof_generation_test.go
- [ ] T077 [US3] Integration test: Verify proof contains all required fields (block_id, timestamp, explorer_url) in tests/integration/proof_generation_test.go

**Checkpoint**: All 4 MCP tools fully functional - complete certification workflow operational

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories and production readiness

### Documentation & Examples

- [ ] T078 [P] Create comprehensive README.md with installation, configuration, and usage in mcp-servers/circular-protocol-mcp-server/
- [ ] T079 [P] Add code comments and godoc documentation to all exported functions
- [ ] T080 [P] Verify quickstart.md examples work end-to-end in specs/003-circular-mcp-server/quickstart.md

### Testing & Quality

- [ ] T081 [P] Run all contract tests and verify 100% pass
- [ ] T082 [P] Run all integration tests on testnet and verify 100% pass
- [ ] T083 [P] Run all unit tests and measure code coverage (target: 90%+)
- [ ] T084 [P] Add edge case tests for error handling scenarios in tests/unit/errors_test.go
- [ ] T085 [P] Perform load testing: 100 concurrent tool invocations in tests/load/concurrent_test.go

### Security & Performance

- [ ] T086 [P] Security audit: Verify no private keys or signatures in logs
- [ ] T087 [P] Security audit: Verify constant-time cryptographic operations in tests/unit/signer_test.go
- [ ] T088 [P] Performance: Verify tool calls respond within 5 seconds in tests/integration/performance_test.go
- [ ] T089 [P] Performance: Verify NAG discovery caches URLs (not queried per request)

### Configuration & Deployment

- [ ] T090 [P] Create production config.yaml.example for mainnet
- [ ] T091 [P] Add environment variable validation at startup in cmd/server/main.go
- [ ] T092 [P] Add graceful shutdown handling in cmd/server/main.go
- [ ] T093 [P] Build and test server binary: go build -o circular-mcp-server cmd/server/main.go

### Final Validation

- [ ] T094 Run complete certification workflow on testnet (all 4 tools)
- [ ] T095 Verify all success criteria from spec.md (SC-001 through SC-007)
- [ ] T096 Update CLAUDE.md with final tech stack via .specify/scripts/bash/update-agent-context.sh

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 4 (Phase 3)**: Depends on Foundational completion
- **User Story 1 (Phase 4)**: Depends on Foundational + US4 completion
- **User Story 2 (Phase 5)**: Depends on Foundational + US1 completion (needs transactions to monitor)
- **User Story 3 (Phase 6)**: Depends on Foundational + US1 + US2 completion (needs executed transactions)
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 4 (Nonce)**: Can start after Foundational - No dependencies on other stories
- **User Story 1 (Certify Data)**: Depends on US4 (needs nonce retrieval)
- **User Story 2 (Status)**: Depends on US1 (needs transactions to monitor)
- **User Story 3 (Proof)**: Depends on US1 + US2 (needs executed transactions)

### Within Each User Story (TDD Workflow)

1. Contract tests MUST be written and FAIL before implementation
2. Implement data structures (entities)
3. Implement service layer (business logic)
4. Implement MCP tool (interface)
5. Register tool with server
6. Integration tests to validate end-to-end
7. Story complete before moving to next priority

### Parallel Opportunities

**Phase 1 (Setup)**: All tasks marked [P] can run in parallel

**Phase 2 (Foundational)**: Within each subsection, tasks marked [P] can run in parallel:
- T007-T010 (Configuration & Logging) - 3 parallel tasks
- T012-T013 (Client utilities) - 2 parallel tasks
- T015-T017 (Cryptography) - 3 parallel tasks
- T021-T027 (Foundational tests) - 7 parallel tasks

**Phase 3-6 (User Stories)**: Within each story:
- Contract tests marked [P] can run in parallel (after TDD setup)
- Implementation tasks marked [P] can run in parallel
- Integration tests marked [P] can run in parallel

**Phase 7 (Polish)**: Most tasks marked [P] can run in parallel

**Cross-Story Parallelism**: Different user stories CANNOT run in parallel due to dependencies (US4 → US1 → US2 → US3)

---

## Parallel Example: User Story 1

```bash
# Phase 4a: Launch all contract tests for US1 together (TDD - Write FIRST):
Task T038: Contract test - certify_data JSON schema validation
Task T039: Contract test - certify_data with valid payload
Task T040: Contract test - certify_data with >1MB payload
Task T041: Contract test - certify_data with insufficient balance

# Phase 4b: Launch parallel implementation tasks for US1:
Task T042: Implement CertificationTransaction validation
Task T043: Implement TransactionStatus struct

# Phase 4c: Sequential implementation (dependencies):
Task T044: Implement transaction builder (depends on T042, T043)
Task T045: Implement Circular_AddTransaction_ API
Task T046: Implement certify_data MCP tool (depends on T044, T045)
...
```

---

## Implementation Strategy

### MVP First (User Story 4 + User Story 1 Only)

1. Complete Phase 1: Setup (6 tasks)
2. Complete Phase 2: Foundational (27 tasks) - CRITICAL
3. Complete Phase 3: User Story 4 - Nonce (10 tasks)
4. Complete Phase 4: User Story 1 - Certify Data (15 tasks)
5. **STOP and VALIDATE**: Test end-to-end certification on testnet
6. **MVP READY**: Agents can certify data on blockchain

**Total MVP Tasks**: 58 tasks

### Incremental Delivery

1. **Foundation** (Setup + Foundational) → Infrastructure ready (33 tasks)
2. **+US4** (Nonce) → Nonce retrieval works (+10 tasks = 43 total)
3. **+US1** (Certify) → Certification works (+15 tasks = 58 total) **← MVP**
4. **+US2** (Status) → Status monitoring works (+13 tasks = 71 total)
5. **+US3** (Proof) → Proof generation works (+12 tasks = 83 total)
6. **Polish** → Production ready (+19 tasks = 102 total)

### Sequential Implementation (Recommended)

Due to story dependencies, recommended order:

1. Phase 1: Setup (parallel tasks where marked)
2. Phase 2: Foundational (parallel tasks within subsections)
3. Phase 3: US4 (Nonce) - Required by US1
4. Phase 4: US1 (Certify Data) - Core functionality **← STOP HERE FOR MVP**
5. Phase 5: US2 (Status) - Monitors US1 transactions
6. Phase 6: US3 (Proof) - Proves US1 transactions
7. Phase 7: Polish - Production hardening

---

## Success Criteria Mapping

Tasks explicitly designed to meet spec.md success criteria:

- **SC-001** (60s certification time): T051 measures confirmation time
- **SC-002** (100% execution rate): T081-T083 verify all tests pass
- **SC-003** (95% within 60s): T065 confirms polling performance
- **SC-004** (100% proof generation): T076-T077 validate proof generation
- **SC-005** (All 4 tools pass tests): T081-T083 validate all tools
- **SC-006** (Graceful error handling): T084 adds error handling tests
- **SC-007** (Deterministic TX IDs): T052 validates TX ID calculation

---

## Notes

- **[P]** tasks = different files, no dependencies within that phase subsection
- **[Story]** label maps task to specific user story for traceability
- **TDD Required**: All contract and integration tests must be written BEFORE implementation (per Constitution Section II)
- Verify tests FAIL before implementing (Red → Green → Refactor cycle)
- Each user story builds on previous stories (sequential dependencies)
- Stop at Phase 4 completion for MVP validation
- Commit after each task or logical group
- Total tasks: 96 (excluding checkpoint validations)
- MVP tasks (Phases 1-4): 58 tasks
- Full implementation: 96 tasks

**Constitution Compliance**:
- ✅ Section I: Specification-driven (all tasks map to spec.md requirements)
- ✅ Section II: Test-first (TDD workflow enforced, 90%+ coverage target)
- ✅ Section III: MCP architecture (stdio transport, stateless, JSON Schema)
- ✅ Section IV: Security-first (private key handling, no sensitive logging)
- ✅ Section V: Observability (structured logging throughout)
- ✅ Section VI: Blockchain integration (NAG discovery, client-side TX ID, Enterprise APIs)
