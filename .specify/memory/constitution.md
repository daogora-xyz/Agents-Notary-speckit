<!--
Sync Impact Report
==================
Version change: N/A → 1.0.0 (Initial Constitution)
Modified principles: N/A (initial creation)
Added sections:
  - I. Specification-Driven Development
  - II. Test-First Development
  - III. MCP Architecture
  - IV. Security-First Design
  - V. Observability & Monitoring
  - VI. Blockchain Integration Standards
Removed sections: None
Templates requiring updates:
  ✅ plan-template.md - Constitution Check section already present
  ✅ spec-template.md - User Scenarios & Requirements align with principles
  ✅ tasks-template.md - Test-first workflow already emphasized
Follow-up TODOs: None - all placeholders filled
-->

# Agents Notary Blockchain Certification Platform Constitution

## Core Principles

### I. Specification-Driven Development

**Mandatory Workflow:**
- MUST use spec-kit workflow for all features: `/speckit.specify` → `/speckit.clarify` → `/speckit.plan` → `/speckit.tasks` → `/speckit.implement`
- All features MUST begin with a complete specification in `/specs/[###-feature-name]/spec.md`
- User stories MUST be prioritized (P1, P2, P3) and independently testable
- Each user story MUST be implementable and deployable as a standalone MVP increment
- Implementation MUST NOT begin until specification is approved

**Rationale:** Ensures clear requirements, prevents scope creep, enables incremental delivery, and maintains alignment between intent and implementation.

### II. Test-First Development (NON-NEGOTIABLE)

**Mandatory TDD Cycle:**
- Tests MUST be written before implementation code
- Tests MUST fail initially (Red phase)
- Implementation proceeds only after failing tests exist (Green phase)
- Refactoring occurs only with passing tests (Refactor phase)
- Contract tests MUST verify API/MCP tool contracts
- Integration tests MUST verify end-to-end user journeys
- Unit tests cover edge cases and business logic

**Coverage Requirements:**
- Minimum 80% code coverage across all packages
- Critical paths (payment, certification) require 95%+ coverage
- Security-sensitive code (signature verification, key management) requires 100% coverage

**Rationale:** Test-first prevents defects, documents behavior, enables safe refactoring, and ensures reliability in financial and blockchain operations.

### III. MCP Architecture

**MCP Server Design:**
- Each MCP server MUST be independently deployable
- MCP servers MUST use stdio transport for host communication
- Tool definitions MUST follow JSON Schema specification
- Tools MUST be discoverable via `tools/list` protocol
- MCP servers MUST NOT maintain state between tool calls
- Error responses MUST include actionable error codes and messages

**MCP Host (Proxy) Design:**
- certify.ar4s.com acts as MCP Host, never as MCP server
- MUST connect to all required MCP servers on startup
- MUST implement connection pooling and retry logic
- MUST gracefully handle MCP server disconnections
- Tool calls MUST be orchestrated through workflow state machine

**Rationale:** Ensures loose coupling, independent scalability, testability, and alignment with MCP protocol standards.

### IV. Security-First Design

**Wallet Key Management:**
- Service wallet private keys MUST be encrypted at rest (AES-256)
- Encryption keys MUST be sourced from environment or KMS/Vault
- Private keys MUST NEVER appear in logs, responses, or error messages
- Separate keys MUST be used for testnet and mainnet
- Key rotation procedures MUST be documented and tested

**Payment Validation:**
- EIP-3009 signatures MUST be verified using cryptographic signature recovery
- All authorization fields MUST be validated: amounts, addresses, nonces, time windows
- Payment nonces MUST be stored to prevent replay attacks
- Failed validation MUST NOT process payments or certifications
- Rate limiting MUST prevent brute force attacks (10 requests/minute per API key)

**Input Validation:**
- All user inputs MUST be validated and sanitized
- Data size MUST be limited (max 10MB)
- Network parameters MUST use allowlist validation
- Callback URLs MUST enforce HTTPS in production
- SQL injection resistance MUST be verified via automated testing

**Rationale:** Blockchain and payment operations require cryptographic security; compromised keys or invalid payments result in financial loss.

### V. Observability & Monitoring

**Logging Standards:**
- MUST use structured JSON logging (Zap library)
- Log levels: DEBUG (development), INFO (operations), WARN (recoverable errors), ERROR (failures)
- MUST NEVER log sensitive data: private keys, full payment authorizations, user PII
- All certification requests MUST log: request_id, client_id, status transitions, error codes
- Correlation IDs MUST propagate through MCP tool calls

**Metrics Requirements:**
- Prometheus metrics MUST be exposed at `/metrics`
- Required metrics: certification success rate, payment verification rate, average certification time, CIRX wallet balance, HTTP request latency
- Custom metrics for critical paths: payment settlement duration, blockchain confirmation time
- Metrics MUST enable SLA monitoring (99.5% uptime, <10s P95 latency)

**Alerting:**
- MUST alert on: CIRX balance < 100 (critical), certification failure rate > 5%, payment verification failure rate > 10%
- Alerts MUST include: current value, threshold, time window, runbook link
- Dead letter queue items MUST trigger operator review

**Rationale:** Real-time visibility enables proactive operations, rapid incident response, and SLA compliance.

### VI. Blockchain Integration Standards

**Circular Protocol:**
- Transactions MUST use Type="C_TYPE_CERTIFICATE"
- Nonce MUST be fetched immediately before each transaction to prevent desync
- Transaction signing MUST use Secp256k1 (matching Circular Protocol)
- CIRX fee is fixed at 4 CIRX per certification
- Transaction status MUST be polled until "Executed" or timeout (30 seconds)
- Failed certifications MUST enter retry queue (exponential backoff, max 10 attempts)

**x402 Payment Protocol:**
- Payment requirements MUST conform to x402 specification version 1
- Supported networks: base, base-sepolia, arbitrum (allowlist validation)
- Payment settlement MUST use x402 facilitator API
- Settlement failures MUST trigger retry (max 10 attempts over 24 hours)
- Idempotency MUST be enforced via `request_id` to prevent double-charging

**Retry & Failure Handling:**
- Payment settled but certification failed = CRITICAL PATH requiring retry
- Retry queue uses exponential backoff: 5s, 10s, 20s, 40s, ..., max 60s
- After max retries, move to dead letter queue for operator intervention
- Operators MUST have dashboard to manually retry, refund, or credit accounts

**Rationale:** Blockchain operations are eventually consistent and require robust retry logic; payment+certification atomicity is critical for user trust.

## Development Workflow

**Feature Lifecycle:**
1. User request → `/speckit.specify` creates spec.md with prioritized user stories
2. Ambiguities → `/speckit.clarify` resolves underspecified areas
3. Specification approved → `/speckit.plan` generates research, design, contracts
4. Implementation plan → `/speckit.tasks` generates dependency-ordered tasks
5. Execution → `/speckit.implement` processes tasks with TDD enforcement
6. Validation → `/speckit.checklist` verifies compliance before merge

**Branching Strategy:**
- Feature branches: `###-feature-name` (matches spec directory)
- All work occurs on feature branch
- Merge to `main` requires: passing tests, spec compliance, constitution check
- Main branch MUST always be deployable

**Code Review Requirements:**
- Constitution compliance MUST be verified in PR review
- Security-sensitive changes (wallet, payments, signatures) require security review
- Test coverage report MUST be included in PR
- Breaking changes require documentation update and migration plan

## Technology Constraints

**Language & Frameworks:**
- Go 1.23+ for all services (type safety, performance, tooling)
- MCP: mcp-go library (github.com/mark3labs/mcp-go)
- HTTP Framework: Gin or Echo (high-performance, middleware support)
- Database: PostgreSQL 16+ (ACID compliance, connection pooling)
- Cache: Redis 7+ (price caching, rate limiting)

**Deployment:**
- Containerization: Docker with multi-stage builds
- Orchestration: Kubernetes with HPA (horizontal pod autoscaling)
- CI/CD: GitHub Actions (test → build → deploy)
- Monitoring: Prometheus + Grafana
- Secrets: Environment variables or HashiCorp Vault

**Performance Standards:**
- P95 latency: < 10 seconds for full certification flow
- Throughput: 100 concurrent requests
- Uptime: 99.5% SLA
- Database connections: max 50 (connection pooling required)

## Governance

**Constitutional Authority:**
- This constitution supersedes all other development practices
- Violations MUST be documented in implementation plan "Complexity Tracking" section with justification
- Amendments require: specification update, approval, template synchronization, version increment

**Amendment Process:**
- Propose change via `/speckit.constitution` command
- Specify: rationale, affected principles, template impacts
- Version increment: MAJOR (breaking changes), MINOR (new principles), PATCH (clarifications)
- Update dependent templates: plan-template.md, spec-template.md, tasks-template.md
- Commit with sync impact report

**Compliance Verification:**
- All PRs MUST verify constitution compliance
- `/speckit.analyze` command checks cross-artifact consistency
- Complexity violations require approval with documented rationale

**Version**: 1.0.0 | **Ratified**: 2025-10-28 | **Last Amended**: 2025-10-28
