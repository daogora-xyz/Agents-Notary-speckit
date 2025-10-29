# Feature Specification: Project Foundation Infrastructure

**Feature Branch**: `001-foundation-setup`
**Created**: 2025-10-28
**Status**: Draft
**Input**: User description: "Foundation setup: Initialize Go monorepo with project structure per docs/OVERVIEW.md Section 3.1. Implement PostgreSQL schema with migrations for certification_requests, payments, certifications, and wallet_balances tables (Section 2.3.1). Set up Redis for caching. Create shared packages: models (Request, Payment, Certification), errors (custom types), and crypto (signing utilities). Reference: docs/OVERVIEW.md Milestone 1, Tasks T001-T004."

## Clarifications

### Session 2025-10-28

- Q: Should failed migrations block environment startup completely, or should the system attempt graceful degradation? → A: Block startup entirely until migrations are manually fixed and re-run
- Q: When Redis cache is unavailable, should services fail immediately or continue operating without caching? → A: Graceful degradation - log warnings and operate without cache (slower but functional)
- Q: Which Redis eviction policy should be configured for the caching layer? → A: allkeys-lru - evict least recently used keys first (recommended for cache workloads)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Development Environment Setup (Priority: P1)

As a developer joining the blockchain certification platform project, I need a fully functional local development environment so that I can start building features immediately without manual infrastructure configuration.

**Why this priority**: Without a working development environment, no feature development can proceed. This is the absolute foundation that blocks all other work.

**Independent Test**: Can be fully tested by running the environment initialization script and verifying that all services start successfully, migrations execute, and basic health checks pass.

**Acceptance Scenarios**:

1. **Given** a fresh checkout of the repository, **When** a developer runs the setup script, **Then** the development environment starts with all required services running and accessible
2. **Given** the development environment is running, **When** a developer connects to the database, **Then** all required tables exist with correct schemas
3. **Given** the development environment is running, **When** a developer imports shared packages in code, **Then** all model types and utilities are available without errors

---

### User Story 2 - Data Persistence Layer (Priority: P2)

As a developer implementing certification workflows, I need a reliable data persistence layer with versioned schema migrations so that certification requests, payments, and blockchain transactions are safely stored and retrievable across application restarts.

**Why this priority**: The certification workflow requires persistent storage of payment records and blockchain transactions for audit, compliance, and retry logic. This must work before any API endpoints can be built.

**Independent Test**: Can be tested by running migration scripts, inserting test records into each table, querying them back, and verifying data integrity after service restarts.

**Acceptance Scenarios**:

1. **Given** the database migration system is initialized, **When** a developer runs the migration command, **Then** all tables are created with correct columns, indexes, and constraints
2. **Given** database tables exist, **When** a certification request is stored, **Then** it can be retrieved with all fields intact including timestamps and status
3. **Given** multiple payments are stored, **When** querying by request_id, **Then** only payments for that specific request are returned
4. **Given** the database is running, **When** the application restarts, **Then** all previously stored data remains accessible

---

### User Story 3 - Caching Infrastructure (Priority: P3)

As a developer implementing pricing and rate limiting features, I need a caching layer so that frequently accessed data (like CIRX token prices) doesn't require repeated API calls and rate limiting can be enforced per client.

**Why this priority**: Caching improves performance and reduces external API costs, but the system can function without it initially (albeit slower and more expensive).

**Independent Test**: Can be tested by setting cache values with expiration times, retrieving them before and after expiration, and verifying rate limit counters increment correctly.

**Acceptance Scenarios**:

1. **Given** the cache service is running, **When** a value is stored with a 5-minute expiration, **Then** it can be retrieved within 5 minutes but returns null after expiration
2. **Given** rate limiting is configured for 10 requests per minute, **When** a client makes 11 requests in one minute, **Then** the 11th request is rejected
3. **Given** cached pricing data exists, **When** the cache expires, **Then** the next request fetches fresh data and updates the cache

---

### User Story 4 - Shared Code Utilities (Priority: P2)

As a developer building MCP servers and API handlers, I need reusable data models and utility functions so that common operations (data validation, cryptographic signing, error handling) are consistent across all services.

**Why this priority**: Shared packages prevent code duplication and ensure consistency in how data is structured and operations are performed. This is critical before building multiple services that must interoperate.

**Independent Test**: Can be tested by importing shared packages in test code, creating model instances, calling utility functions, and verifying expected behaviors match specifications.

**Acceptance Scenarios**:

1. **Given** shared model packages are available, **When** a developer creates a CertificationRequest model, **Then** it enforces required fields and provides validation methods
2. **Given** the crypto utilities package exists, **When** a developer signs data using the Secp256k1 function, **Then** the signature can be verified using the corresponding verification function
3. **Given** custom error types are defined, **When** a developer wraps an error with context, **Then** the error chain preserves original error information
4. **Given** multiple services import shared packages, **When** data is serialized by one service and deserialized by another, **Then** field types and values match exactly

---

### Edge Cases

- Database migrations that fail partway through (e.g., due to constraint violations) MUST block environment startup entirely. The system MUST display clear error messages indicating which migration failed and require manual intervention to fix and re-run migrations before services can start.
- Cache service (Redis) unavailability MUST NOT block service startup. Services MUST log warnings about cache unavailability and operate without caching (with degraded performance - direct API calls instead of cached values, no rate limiting enforcement).
- Redis cache MUST be configured with allkeys-lru eviction policy so that when memory limits are reached, the least recently used keys are evicted first to make room for new entries.
- What happens when shared packages are updated in a breaking way (how are dependent services notified)?
- How does the system handle concurrent migration attempts by multiple developers?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a reproducible development environment that can be initialized with a single command
- **FR-002**: System MUST implement a versioned database migration framework that tracks which migrations have been applied and blocks environment startup if migrations fail
- **FR-003**: System MUST create certification_requests table with fields for request tracking (id, request_id, client_id, data_hash, data_size_bytes, status, timestamps)
- **FR-004**: System MUST create payments table with fields for payment tracking (id, request_id, payment_nonce, from_address, to_address, amount_usdc, network, evm_tx_hash, status, timestamps)
- **FR-005**: System MUST create certifications table with fields for blockchain certification records (id, request_id, cirx_tx_id, cirx_block_id, cirx_fee_paid, status, retry_count, timestamps)
- **FR-006**: System MUST create wallet_balances table with fields for tracking service wallet balances (id, asset, network, wallet_address, balance, last_updated)
- **FR-007**: System MUST provide database connection pooling configuration with maximum connection limits
- **FR-008**: System MUST provide cache service with configurable time-to-live (TTL) for stored values, with graceful degradation when cache is unavailable
- **FR-009**: System MUST provide cache service with key-based storage and retrieval operations, falling back to direct operations if cache is unavailable
- **FR-010**: System MUST configure Redis with allkeys-lru eviction policy to automatically evict least recently used keys when memory limits are reached
- **FR-011**: System MUST provide shared data models for Request, Payment, and Certification entities with field validation
- **FR-012**: System MUST provide cryptographic signing utilities compatible with Secp256k1 curve (for Circular Protocol compatibility)
- **FR-013**: System MUST provide custom error types for common failure scenarios (validation errors, network errors, blockchain errors)
- **FR-014**: System MUST support both forward migrations (applying schema changes) and rollback migrations (reverting schema changes)
- **FR-015**: System MUST create indexes on frequently queried fields (request_id, payment_nonce, client_id, status fields)
- **FR-016**: System MUST enforce unique constraints where required (request_id, payment_nonce, cirx_tx_id)

### Key Entities

- **Certification Request**: Represents a user's request to certify data on the blockchain. Tracks request lifecycle from initiation through payment to completion. Key attributes: unique request ID, client identifier, data hash, data size, current status, timestamps.

- **Payment**: Represents a payment authorization for certification service. Links to a certification request and tracks payment lifecycle. Key attributes: payment nonce (for idempotency), source and destination addresses, amount in USDC, blockchain network, settlement transaction hash, payment status.

- **Certification**: Represents a completed blockchain certification transaction on Circular Protocol. Links to a certification request and contains blockchain proof details. Key attributes: Circular Protocol transaction ID, block ID, CIRX fee paid, confirmation status, retry count for failed attempts.

- **Wallet Balance**: Represents the current balance of service wallets across different blockchain networks. Used for monitoring and alerting. Key attributes: asset type (e.g., CIRX), network identifier, wallet address, current balance, last update timestamp.

- **Shared Models**: Reusable data structures used across multiple services for consistency. Provide validation, serialization, and type safety.

- **Error Types**: Categorized error representations for different failure modes (validation, network, blockchain, payment). Enable consistent error handling and user-friendly messages.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Development environment initialization completes in under 5 minutes on standard hardware
- **SC-002**: Database migrations execute successfully on both empty databases and databases with existing data
- **SC-003**: All database tables support concurrent read/write operations from 50 simultaneous connections without errors
- **SC-004**: Cache operations (set, get, delete) complete in under 10 milliseconds for 95% of requests
- **SC-005**: Shared packages can be imported and used by test code without compilation errors
- **SC-006**: Cryptographic signing operations complete in under 100 milliseconds per signature
- **SC-007**: Data model validation catches invalid data 100% of the time in test scenarios
- **SC-008**: Database rollback migrations successfully revert schema changes without data loss
- **SC-009**: All services can start and stop cleanly without leaving orphaned processes or locked resources
- **SC-010**: Documentation allows a new developer to set up the environment in under 30 minutes

### Assumptions

- Development environment will run on Linux or macOS (Windows via WSL2)
- Developers have sufficient disk space (minimum 10GB) and memory (minimum 8GB RAM)
- Database schema follows PostgreSQL 16+ capabilities
- Cache service follows Redis 7+ capabilities
- Migration framework will be selected based on ecosystem best practices
- Cryptographic utilities prioritize correctness over performance (blockchain signing is not latency-critical)
- All configuration will use environment variables for secrets and config files for non-sensitive settings
