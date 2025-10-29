# Implementation Plan: x402 Payment MCP Server

**Branch**: `002-x402-mcp-server` | **Date**: 2025-10-28 | **Spec**: [spec.md](./spec.md)

## Summary

Build an MCP server that provides 5 payment-related tools for the certify.ar4s.com HTTP proxy: generating x402 payment requirements, verifying EIP-3009 payment signatures, settling payments via x402 facilitator, generating MetaMask browser deep links, and encoding EIP-681 URIs for QR code mobile payments. The server supports base, base-sepolia, and arbitrum networks using native USDC tokens.

**Technical Approach**: Implement stateless MCP server using mcp-go with stdio transport, in-memory caching for settlement idempotency (10-minute TTL), blockchain RPC integration for nonce retrieval, structured JSON logging, and YAML configuration for network parameters.

## Technical Context

**Language/Version**: Go 1.23+ (from Nix flake)
**Primary Dependencies**:
- `github.com/mark3labs/mcp-go` - MCP server framework
- `github.com/ethereum/go-ethereum` - EIP-712 signing, crypto utilities, RPC client
- `gopkg.in/yaml.v3` - Configuration file parsing

**Storage**: In-memory cache with TTL for settlement idempotency (no persistent database)
**Testing**: Go standard testing (`go test`), testify for assertions, dockertest for integration tests
**Target Platform**: Linux server (containerized), runs as subprocess of HTTP proxy via stdio
**Project Type**: Single Go project (MCP server component)
**Performance Goals**:
- <100ms tool response time (excluding facilitator I/O per SC-001)
- 10 concurrent tool calls without race conditions (SC-007)
- 100% signature verification accuracy across 1000 test cases (SC-002)
- 95% first-attempt settlement success rate (SC-003)

**Constraints**:
- No authentication (stdio security model - inherits proxy context)
- Stateless between tool calls (MCP architecture requirement)
- Configuration-driven network parameters (no hardcoded addresses)
- 5-second timeout for facilitator HTTP calls (FR-011)

**Scale/Scope**:
- 5 MCP tools
- 3 supported networks (base, base-sepolia, arbitrum)
- ~2000 LOC estimated (tools, config, crypto utilities, tests)

## Constitution Check

### I. Specification-Driven Development ✅ PASS

- [x] Feature began with complete specification (`specs/002-x402-mcp-server/spec.md`)
- [x] 5 user stories prioritized (P1: payment primitives, P2: UX, P3: mobile)
- [x] Each story independently testable (verified in spec acceptance scenarios)
- [x] Spec approved before implementation planning

### II. Test-First Development (NON-NEGOTIABLE) ✅ PASS

**TDD Cycle Planned**:
- [x] Tests MUST be written before implementation (enforced in tasks.md generation)
- [x] Contract tests planned for MCP tool schemas (QM-001)
- [x] Integration tests planned for facilitator interaction
- [x] Unit tests planned for signature verification, nonce generation

**Coverage Requirements**:
- [x] Target: 90%+ code coverage (SC-008)
- [x] Signature verification: 100% coverage required (security-sensitive per Constitution IV)
- [x] Each tool: happy path + 3 error scenarios minimum (QM-005)

### III. MCP Architecture ✅ PASS

- [x] Independent MCP server (not part of HTTP proxy codebase)
- [x] stdio transport (FR-001)
- [x] JSON Schema tool definitions (QM-001)
- [x] Tools discoverable via mcp-go `tools/list`
- [x] Stateless design (clarification: in-memory cache is transient, not persistent state)
- [x] Actionable error responses (QM-002)

### IV. Security-First Design ✅ PASS

**Wallet Key Management**: N/A (agents bring their own keys)

**Payment Validation**:
- [x] EIP-3009 signature verification via secp256k1 recovery (FR-007)
- [x] All authorization fields validated (FR-009: time bounds, FR-010: signer recovery)
- [x] Nonces tracked via blockchain RPC (FR-005 clarification)
- [x] Failed validation prevents settlement (edge case handling)
- [x] Rate limiting: delegated to HTTP proxy layer

**Input Validation**:
- [x] All tool inputs validated (FR-019)
- [x] Network parameters use allowlist (FR-004: base, base-sepolia, arbitrum)
- [x] No SQL injection risk (no database)

### V. Observability & Monitoring ✅ PASS

**Logging Standards**:
- [x] Structured JSON logging (QM-004 clarification)
- [x] Log levels: DEBUG/INFO/WARN/ERROR
- [x] No sensitive data logged (signature components excluded from logs)
- [x] All tool calls logged with context: tool_name, network, duration_ms
- [x] Facilitator interactions logged (request/response)

**Metrics Requirements**: ⚠️ DEFERRED TO HTTP PROXY
- Prometheus metrics endpoint delegated to proxy layer
- MCP server logs provide data source for proxy metrics aggregation
- Justification: MCP servers are stateless subprocesses, metrics aggregation better suited to long-running proxy

**Alerting**: ⚠️ DELEGATED TO HTTP PROXY
- Settlement failures logged at ERROR level for proxy alerting
- MCP server provides observability data, proxy implements alerting logic

### VI. Blockchain Integration Standards ✅ PASS

**x402 Payment Protocol**:
- [x] Conforms to x402 v1 specification (FR-003)
- [x] Supported networks: base, base-sepolia, arbitrum (FR-004)
- [x] Facilitator API integration (FR-011: POST with 5s timeout)
- [x] Idempotency enforced (FR-013: in-memory cache by nonce)

**Retry & Failure Handling**:
- [x] Facilitator errors return structured responses (FR-014)
- [x] Network timeouts handled gracefully (edge case: tool returns retriable error)
- [x] Nonce fetch retry logic (edge case: 3 attempts then error)

**Note**: Circular Protocol integration is out of scope for this MCP server (handled by separate `circular-protocol-mcp-server`)

---

**GATE RESULT**: ✅ ALL CHECKS PASS

Minor deviations justified:
- Prometheus metrics deferred to HTTP proxy (MCP subprocess model makes this appropriate)
- Alerting delegated to proxy (MCP server is stateless component)

## Project Structure

### Documentation (this feature)

```text
specs/002-x402-mcp-server/
├── plan.md              # This file
├── research.md          # Phase 0 output (technical research on x402, EIP-3009, EIP-712, EIP-681)
├── data-model.md        # Phase 1 output (entities, configuration schema)
├── quickstart.md        # Phase 1 output (developer setup guide)
├── contracts/           # Phase 1 output (MCP tool JSON schemas)
│   ├── create_payment_requirement.json
│   ├── verify_payment.json
│   ├── settle_payment.json
│   ├── generate_browser_link.json
│   └── encode_payment_for_qr.json
└── tasks.md             # Phase 2 output (NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
mcp-servers/
└── x402-mcp-server/
    ├── main.go                    # MCP server entry point, stdio transport setup
    ├── config.yaml.example        # Network configuration template
    ├── tools/
    │   ├── create_payment_requirement.go
    │   ├── verify_payment.go
    │   ├── settle_payment.go
    │   ├── generate_browser_link.go
    │   └── encode_payment_for_qr.go
    ├── internal/
    │   ├── x402/
    │   │   ├── payment_requirement.go  # x402 JSON generation
    │   │   └── facilitator_client.go   # HTTP client for x402 facilitator API
    │   ├── eip3009/
    │   │   ├── authorization.go        # EIP-3009 data structures
    │   │   └── signature_verifier.go   # EIP-712 + secp256k1 verification
    │   ├── eip681/
    │   │   └── uri_encoder.go          # EIP-681 payment URI generation
    │   ├── browser/
    │   │   └── link_generator.go       # MetaMask deep link generation
    │   ├── config/
    │   │   ├── config.go               # YAML configuration loading
    │   │   └── network.go              # Network-specific parameters
    │   ├── cache/
    │   │   └── ttl_cache.go            # In-memory cache with expiration
    │   ├── rpc/
    │   │   └── nonce_fetcher.go        # Blockchain RPC nonce queries
    │   └── logger/
    │       └── structured.go            # JSON logging setup
    └── tests/
        ├── contract/
        │   └── mcp_tool_schemas_test.go
        ├── integration/
        │   ├── facilitator_integration_test.go
        │   ├── rpc_integration_test.go
        │   └── browser_link_integration_test.go
        └── unit/
            ├── x402_test.go
            ├── eip3009_test.go
            ├── eip681_test.go
            ├── browser_test.go
            ├── cache_test.go
            └── config_test.go
```

**Structure Decision**: Single Go project under `mcp-servers/x402-mcp-server/` directory. This follows the monorepo pattern established in OVERVIEW.md where each MCP server is a self-contained module. The `internal/` package ensures implementation details are not exported, while `tools/` contains MCP tool handlers. Configuration is external (YAML) to support multiple networks without code changes (FR-020, SC-010).

## Complexity Tracking

> **No Constitution violations requiring justification**

All architecture decisions align with constitution principles. The choice to defer Prometheus metrics to the HTTP proxy layer is a design decision within the MCP architecture principle (stateless subprocesses), not a violation.
