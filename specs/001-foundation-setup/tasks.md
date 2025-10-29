# Tasks: Project Foundation Infrastructure

**Input**: Design documents from `/specs/001-foundation-setup/`
**Prerequisites**: plan.md (required), spec.md (required), data-model.md, research.md, quickstart.md

**Tests**: Per constitution Principle II (Test-First Development), this milestone includes comprehensive test tasks. Tests are written BEFORE implementation code.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- **Go monorepo**: Root level contains `pkg/`, `migrations/`, `tests/`, `scripts/`
- **Shared packages**: `pkg/models/`, `pkg/crypto/`, `pkg/errors/`
- **Migrations**: `migrations/001_init.up.sql`, `migrations/001_init.down.sql`
- **Tests**: `tests/integration/`, `tests/unit/` (unit tests colocated with packages)
- **Scripts**: `scripts/setup-dev.sh`, `scripts/migrate.sh`, `scripts/health-check.sh`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Initialize Go module with `go mod init` at repository root
- [ ] T002 Create directory structure: `pkg/`, `migrations/`, `tests/integration/`, `tests/unit/`, `scripts/`, `docs/`
- [ ] T003 [P] Create `docker-compose.yml` with PostgreSQL 16 and Redis 7 service definitions
- [ ] T004 [P] Create `.env.example` template with all required environment variables
- [ ] T005 [P] Create `Makefile` with common commands (up, down, migrate-up, migrate-down, test, health)
- [ ] T006 [P] Create `.gitignore` with Go, Docker, and IDE-specific entries

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T007 Add Go dependencies to `go.mod`: pgx, go-redis, btcsuite/btcd, testify, dockertest, golang-migrate
- [ ] T008 Run `go mod download` and `go mod tidy` to fetch all dependencies
- [ ] T009 Create database migration files: `migrations/001_init.up.sql` and `migrations/001_init.down.sql` with all 4 table schemas
- [ ] T010 Configure Docker Compose health checks for PostgreSQL (port ready) and Redis (ping response)
- [ ] T011 Create `scripts/migrate.sh` wrapper script for golang-migrate CLI
- [ ] T012 Document PostgreSQL connection string format in `.env.example`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Development Environment Setup (Priority: P1) 🎯 MVP

**Goal**: Developers can initialize complete local development environment with single command

**Independent Test**: Run setup script from fresh checkout, verify all services running, migrations applied, health checks pass

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T013 [P] [US1] Integration test for Docker Compose service startup in `tests/integration/docker_test.go`
- [ ] T014 [P] [US1] Integration test for migration up/down in `tests/integration/migration_test.go`
- [ ] T015 [P] [US1] Integration test for health check script in `tests/integration/health_test.go`

### Implementation for User Story 1

- [ ] T016 [US1] Implement `scripts/setup-dev.sh`: check prerequisites, copy .env.example to .env, run docker-compose up, run migrations
- [ ] T017 [US1] Implement `scripts/health-check.sh`: verify PostgreSQL connection, verify Redis connection, check migration status
- [ ] T018 [US1] Update README.md with quickstart instructions linking to `specs/001-foundation-setup/quickstart.md`
- [ ] T019 [US1] Test full setup workflow: clean environment → run setup-dev.sh → verify health check passes
- [ ] T020 [US1] Document troubleshooting steps in `docs/TROUBLESHOOTING.md` for common Docker/migration issues

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Data Persistence Layer (Priority: P2)

**Goal**: Reliable versioned database schema with 4 tables for certification workflow persistence

**Independent Test**: Run migrations, insert/query test data in all 4 tables, verify data integrity after restart

### Tests for User Story 2 ⚠️

- [ ] T021 [P] [US2] Unit test for migration rollback in `tests/integration/migration_rollback_test.go`
- [ ] T022 [P] [US2] Integration test for certification_requests CRUD in `tests/integration/db_certification_requests_test.go`
- [ ] T023 [P] [US2] Integration test for payments CRUD in `tests/integration/db_payments_test.go`
- [ ] T024 [P] [US2] Integration test for certifications CRUD in `tests/integration/db_certifications_test.go`
- [ ] T025 [P] [US2] Integration test for wallet_balances CRUD in `tests/integration/db_wallet_balances_test.go`

### Implementation for User Story 2

- [ ] T026 [US2] Implement certification_requests table schema in `migrations/001_init.up.sql` with indexes and constraints
- [ ] T027 [US2] Implement payments table schema in `migrations/001_init.up.sql` with foreign key to certification_requests
- [ ] T028 [US2] Implement certifications table schema in `migrations/001_init.up.sql` with foreign key to certification_requests
- [ ] T029 [US2] Implement wallet_balances table schema in `migrations/001_init.up.sql` with unique constraint
- [ ] T030 [US2] Implement DROP TABLE statements in `migrations/001_init.down.sql` for rollback
- [ ] T031 [US2] Test migration failure blocks startup: intentionally break migration, verify docker-compose health check fails
- [ ] T032 [US2] Test migration rollback: apply migration, insert data, rollback, verify tables dropped and data gone
- [ ] T033 [US2] Document migration failure recovery procedure in `docs/TROUBLESHOOTING.md`

**Checkpoint**: At this point, User Story 2 should work independently - database schema fully migrated and tested

---

## Phase 5: User Story 4 - Shared Code Utilities (Priority: P2)

**Goal**: Reusable Go packages for models, crypto utilities, and error types

**Independent Test**: Import shared packages in test code, create instances, validate, sign data, verify serialization

**Note**: US4 implemented before US3 because caching (US3) depends on shared models

### Tests for User Story 4 ⚠️

- [ ] T034 [P] [US4] Unit test for CertificationRequest model validation in `pkg/models/request_test.go`
- [ ] T035 [P] [US4] Unit test for Payment model validation in `pkg/models/payment_test.go`
- [ ] T036 [P] [US4] Unit test for Certification model validation in `pkg/models/certification_test.go`
- [ ] T037 [P] [US4] Unit test for WalletBalance model validation in `pkg/models/wallet_test.go`
- [ ] T038 [P] [US4] Unit test for Secp256k1 signing/verification in `pkg/crypto/secp256k1_test.go`
- [ ] T039 [P] [US4] Unit test for error type wrapping in `pkg/errors/types_test.go`
- [ ] T040 [P] [US4] Integration test for cross-service serialization in `tests/integration/models_serialization_test.go`

### Implementation for User Story 4

- [ ] T041 [P] [US4] Implement CertificationRequest model with validation in `pkg/models/request.go`
- [ ] T042 [P] [US4] Implement Payment model with validation in `pkg/models/payment.go`
- [ ] T043 [P] [US4] Implement Certification model with validation in `pkg/models/certification.go`
- [ ] T044 [P] [US4] Implement WalletBalance model with validation in `pkg/models/wallet.go`
- [ ] T045 [P] [US4] Implement Secp256k1 signing utilities in `pkg/crypto/secp256k1.go` using btcsuite/btcd
- [ ] T046 [P] [US4] Implement Secp256k1 verification utilities in `pkg/crypto/secp256k1.go`
- [ ] T047 [P] [US4] Implement custom error types (ValidationError, NetworkError, BlockchainError, PaymentError) in `pkg/errors/types.go`
- [ ] T048 [US4] Verify all models can be imported without errors: create test main.go that imports all packages
- [ ] T049 [US4] Benchmark crypto signing performance: ensure <100ms per operation per spec success criterion SC-006

**Checkpoint**: All shared packages fully functional and independently testable

---

## Phase 6: User Story 3 - Caching Infrastructure (Priority: P3)

**Goal**: Redis caching layer with TTL support and graceful degradation

**Independent Test**: Set cache values with TTL, retrieve before/after expiration, verify graceful degradation when Redis unavailable

### Tests for User Story 3 ⚠️

- [ ] T050 [P] [US3] Integration test for Redis connection and basic operations in `tests/integration/cache_test.go`
- [ ] T051 [P] [US3] Integration test for TTL expiration in `tests/integration/cache_ttl_test.go`
- [ ] T052 [P] [US3] Integration test for graceful degradation (Redis stopped) in `tests/integration/cache_degradation_test.go`
- [ ] T053 [P] [US3] Integration test for allkeys-lru eviction policy in `tests/integration/cache_eviction_test.go`

### Implementation for User Story 3

- [ ] T054 [US3] Create Redis configuration in `docker-compose.yml` with `--maxmemory-policy allkeys-lru` command
- [ ] T055 [US3] Create `pkg/cache/` directory and implement Redis client wrapper in `pkg/cache/client.go`
- [ ] T056 [US3] Implement Set operation with TTL in `pkg/cache/client.go`
- [ ] T057 [US3] Implement Get operation with graceful fallback in `pkg/cache/client.go`
- [ ] T058 [US3] Implement Delete operation in `pkg/cache/client.go`
- [ ] T059 [US3] Add connection error handling: log warning, set fallback flag, continue without cache
- [ ] T060 [US3] Test graceful degradation: stop Redis container, verify services still start and log warnings
- [ ] T061 [US3] Test cache performance: measure P95 latency for Set/Get operations, verify <10ms per spec SC-004
- [ ] T062 [US3] Document Redis configuration and fallback behavior in README.md

**Checkpoint**: Caching infrastructure complete and independently functional

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T063 [P] Update README.md with complete setup instructions, architecture diagram, and links to specs
- [ ] T064 [P] Create `docs/ARCHITECTURE.md` documenting Go monorepo structure and shared package design
- [ ] T065 [P] Verify all unit tests pass: `go test ./pkg/... -v -cover`
- [ ] T066 [P] Verify all integration tests pass: `go test ./tests/integration/... -v`
- [ ] T067 Verify test coverage meets 80% minimum per constitution: `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out`
- [ ] T068 Run quickstart guide validation: follow `specs/001-foundation-setup/quickstart.md` step-by-step on clean environment
- [ ] T069 Benchmark environment initialization time: verify <5 minutes per spec SC-001
- [ ] T070 Verify Docker Compose supports 50 concurrent database connections per spec SC-003
- [ ] T071 Create `CHANGELOG.md` with Milestone 1 completion notes
- [ ] T072 [P] Code cleanup: run `gofmt -w .` and `go vet ./...`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User Story 1 (P1): Can start after Foundational - No dependencies on other stories
  - User Story 4 (P2): Can start after Foundational - No dependencies on other stories
  - User Story 2 (P2): Can start after Foundational - May use US4 models for testing
  - User Story 3 (P3): Can start after Foundational - May cache US4 models
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - Independent, no cross-story dependencies
- **User Story 4 (P2)**: Can start after Foundational (Phase 2) - Independent, provides models for US2/US3
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - May integrate with US4 models for CRUD tests
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - May cache US4 models for serialization tests

### Within Each User Story

- Tests (marked with ⚠️) MUST be written and FAIL before implementation
- Models before services
- Database schema before CRUD operations
- Error handling after core implementation
- Integration tests after unit tests
- Story complete before moving to next priority

### Parallel Opportunities

- **Setup Phase**: All tasks T003-T006 can run in parallel (different files)
- **Foundational Phase**: T007-T008 sequential, T010 can run in parallel with T009
- **User Story 1 Tests**: T013-T015 all parallel (different test files)
- **User Story 2 Tests**: T021-T025 all parallel (different test files)
- **User Story 4 Tests**: T034-T040 all parallel (different test files)
- **User Story 4 Implementation**: T041-T047 all parallel (different packages/files)
- **User Story 3 Tests**: T050-T053 all parallel (different test files)
- **Polish Phase**: T063-T066, T072 all parallel (different docs/test suites)
- **Cross-Story**: After Foundational phase completes, US1, US2, US3, US4 can all start in parallel

---

## Parallel Example: User Story 4 (Shared Code Utilities)

```bash
# Launch all tests for User Story 4 together:
Task: "Unit test for CertificationRequest model validation in pkg/models/request_test.go"
Task: "Unit test for Payment model validation in pkg/models/payment_test.go"
Task: "Unit test for Certification model validation in pkg/models/certification_test.go"
Task: "Unit test for WalletBalance model validation in pkg/models/wallet_test.go"
Task: "Unit test for Secp256k1 signing/verification in pkg/crypto/secp256k1_test.go"
Task: "Unit test for error type wrapping in pkg/errors/types_test.go"
Task: "Integration test for cross-service serialization in tests/integration/models_serialization_test.go"

# Launch all model implementations for User Story 4 together:
Task: "Implement CertificationRequest model with validation in pkg/models/request.go"
Task: "Implement Payment model with validation in pkg/models/payment.go"
Task: "Implement Certification model with validation in pkg/models/certification.go"
Task: "Implement WalletBalance model with validation in pkg/models/wallet.go"
Task: "Implement Secp256k1 signing utilities in pkg/crypto/secp256k1.go"
Task: "Implement custom error types in pkg/errors/types.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Environment Setup)
4. **STOP and VALIDATE**: Test User Story 1 independently
   - Run `scripts/setup-dev.sh` from clean checkout
   - Verify `scripts/health-check.sh` passes
   - Verify Docker Compose services running
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Document (MVP: developers can set up environment)
3. Add User Story 4 → Test independently → Document (shared packages available)
4. Add User Story 2 → Test independently → Document (database schema ready for use)
5. Add User Story 3 → Test independently → Document (caching layer available)
6. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (Environment Setup)
   - Developer B: User Story 4 (Shared Packages)
   - Developer C: User Story 2 (Database Schema)
   - Developer D: User Story 3 (Caching)
3. Stories complete and integrate independently
4. All stories meet at Polish phase for final integration testing

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing (TDD: RED → GREEN → REFACTOR)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- **Constitution Compliance**: This task list enforces 80% test coverage minimum per Principle II
- **Migration Failure**: Per clarification, failed migrations block startup (T031 validates this behavior)
- **Cache Degradation**: Per clarification, Redis unavailability doesn't block startup (T060 validates this behavior)
- **Redis Eviction**: Per clarification, allkeys-lru policy configured (T054 implements, T053 validates)
