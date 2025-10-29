# Research: Technology Selection for Foundation Infrastructure

**Feature**: Project Foundation Infrastructure
**Branch**: 001-foundation-setup
**Date**: 2025-10-28

## Purpose

This document records technology selection decisions for the foundational infrastructure milestone, including database migration framework, PostgreSQL driver, Redis client, cryptographic libraries, and testing tools.

## Research Questions

1. Which Go database migration framework best supports PostgreSQL with rollbacks and CLI+library usage?
2. Which PostgreSQL driver provides the best performance, connection pooling, and error handling?
3. Which Redis client for Go supports graceful degradation when Redis is unavailable?
4. Which Secp256k1 library is compatible with Circular Protocol's signing requirements?
5. Which testing tools support Docker container lifecycle management for integration tests?

---

## Decision 1: Database Migration Framework

### Requirement
- Must support PostgreSQL 16+
- Must support both forward (up) and rollback (down) migrations
- Must provide CLI tool for manual migration management
- Must provide Go library for programmatic migration checks
- Must prevent concurrent migration attempts
- Must track which migrations have been applied

### Options Evaluated

| Library | Pros | Cons | Score |
|---------|------|------|-------|
| **golang-migrate/migrate** | Industry standard, 12k+ stars, supports 15+ databases, CLI + Go library, advisory locks for concurrency, active maintenance | Requires separate CLI installation | 9/10 |
| **pressly/goose** | Simple API, embedded migrations in Go code, good docs | Smaller community (5k stars), fewer database drivers | 7/10 |
| **rubenv/sql-migrate** | Embedded migrations, Go-based DSL | Limited PostgreSQL-specific features, less popular | 6/10 |
| **go-gormigrate/gormigrate** | Integrates with GORM ORM | Requires GORM dependency (overkill for this use case) | 5/10 |

### Decision: **golang-migrate/migrate**

**Rationale**:
- De facto standard in Go ecosystem (most GitHub stars, widest adoption)
- Dual CLI + library usage: developers run `migrate` CLI manually, services check migration status programmatically
- Built-in PostgreSQL advisory locks prevent concurrent migrations (addresses edge case from spec)
- Supports file-based SQL migrations (up/down pairs) - clearer than Go-embedded for SQL experts
- Used by major projects: HashiCorp, GitLab

**Installation**:
```bash
go get -u github.com/golang-migrate/migrate/v4
go get -u github.com/golang-migrate/migrate/v4/database/postgres
go get -u github.com/golang-migrate/migrate/v4/source/file
```

**CLI**:
```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
mv migrate /usr/local/bin/migrate
```

---

## Decision 2: PostgreSQL Driver

### Requirement
- Must support PostgreSQL 16+ features
- Must support connection pooling (max 50 connections per spec)
- Must support prepared statements (SQL injection prevention)
- Must provide excellent error handling and diagnostics
- Must integrate with golang-migrate

### Options Evaluated

| Library | Pros | Cons | Score |
|---------|------|------|-------|
| **jackc/pgx** | Modern (v5), best performance, native PostgreSQL protocol, built-in connection pool, detailed error types, supports LISTEN/NOTIFY | Slightly less adoption than lib/pq | 9/10 |
| **lib/pq** | Most popular (8k+ stars), battle-tested, simple API | Older, slower, less detailed errors, maintenance mode | 7/10 |
| **go-pg/pg** | ORM-like features, connection pooling | ORM overhead, not needed for this use case | 6/10 |

### Decision: **jackc/pgx**

**Rationale**:
- **Performance**: 2-3x faster than lib/pq in benchmarks (native PostgreSQL protocol vs. database/sql layer)
- **Connection Pooling**: Built-in `pgxpool` with sensible defaults, easy max connection limiting
- **Error Handling**: Rich error types (`pgconn.PgError`) with detailed constraint violation info (critical for debugging migration failures)
- **Modern**: Active development, supports all PostgreSQL 16 features
- **golang-migrate compatibility**: Fully supported as migration driver

**Configuration Example**:
```go
import "github.com/jackc/pgx/v5/pgxpool"

config, _ := pgxpool.ParseConfig("postgres://user:pass@localhost:5432/dbname")
config.MaxConns = 50  // Per spec: FR-007
pool, _ := pgxpool.NewWithConfig(context.Background(), config)
```

---

## Decision 3: Redis Client

### Requirement
- Must support Redis 7+ features
- Must support TTL-based key expiration
- Must support graceful degradation (continue operating if Redis unavailable per spec clarification)
- Must support connection pooling
- Must configure eviction policy (allkeys-lru per spec clarification)

### Options Evaluated

| Library | Pros | Cons | Score |
|---------|------|------|-------|
| **go-redis/redis** | Most popular (18k+ stars), supports Redis 7, excellent docs, connection pooling, pipeline support, active maintenance | None significant | 9/10 |
| **gomodule/redigo** | Lightweight, simple API | Smaller community, fewer advanced features | 7/10 |
| **mediocregopher/radix** | High performance, cluster support | Less popular, steeper learning curve | 6/10 |

### Decision: **go-redis/redis**

**Rationale**:
- **Graceful Degradation**: Easy error handling pattern:
  ```go
  val, err := rdb.Get(ctx, "key").Result()
  if err == redis.Nil {
      // Key doesn't exist - fetch from source
  } else if err != nil {
      // Redis unavailable - log warning, fall back
      log.Warn("Redis unavailable, operating without cache")
      return fetchFromSource()
  }
  ```
- **Connection Pooling**: Built-in with `redis.NewClient()`, automatic reconnection
- **TTL Support**: Native `Set(key, value, 5*time.Minute)` for 5-minute cache expiration (per spec: FR-008)
- **Eviction Policy**: Configured on Redis server side (docker-compose.yml), client-agnostic

**Configuration Example**:
```go
import "github.com/redis/go-redis/v9"

rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "", // no password for local dev
    DB:       0,  // default DB
})
```

---

## Decision 4: Secp256k1 Cryptography Library

### Requirement
- Must implement Secp256k1 elliptic curve (Bitcoin/Ethereum standard)
- Must be compatible with Circular Protocol's signing requirements
- Must support signing and signature verification
- Must be well-tested (blockchain-grade reliability)
- Must complete signing in < 100ms (per spec: SC-006)

### Options Evaluated

| Library | Pros | Cons | Score |
|---------|------|------|-------|
| **btcsuite/btcd/btcec/v2** | Bitcoin Core implementation in Go, battle-tested, widely used in blockchain projects, supports Secp256k1 signing/verification | Requires btcd dependency (but lightweight) | 9/10 |
| **ethereum/go-ethereum/crypto** | Ethereum's crypto package, includes Secp256k1 | Heavier dependency (full Ethereum client libs) | 7/10 |
| **decred/dcrd/dcrec/secp256k1/v4** | High-performance, pure Go | Smaller community than btcsuite | 7/10 |

### Decision: **btcsuite/btcd/btcec/v2**

**Rationale**:
- **Proven**: Used by Bitcoin full node implementation (btcd), audited and battle-tested
- **Performance**: Sub-1ms signing on modern hardware (well under 100ms requirement)
- **Circular Protocol Compatibility**: Secp256k1 is the same curve used by Bitcoin, Ethereum, and Circular Protocol - ensures signature interoperability
- **API Simplicity**:
  ```go
  import "github.com/btcsuite/btcd/btcec/v2"
  import "github.com/btcsuite/btcd/btcec/v2/ecdsa"

  privKey, _ := btcec.NewPrivateKey()
  hash := sha256.Sum256(data)
  signature := ecdsa.Sign(privKey, hash[:])
  verified := signature.Verify(hash[:], privKey.PubKey())
  ```
- **Adoption**: Used by Lightning Network, various blockchain projects

**Alternative Considered**: `go-ethereum/crypto` was close second, but brings unnecessary Ethereum-specific dependencies (RLP encoding, trie structures, etc.)

---

## Decision 5: Testing Framework

### Requirement
- Must support Go's built-in testing framework
- Must provide assertion helpers (reduce boilerplate)
- Must support Docker container lifecycle for integration tests (PostgreSQL, Redis)
- Must clean up test containers after tests complete
- Must support parallel test execution

### Options Evaluated

| Library | Pros | Cons | Score |
|---------|------|------|-------|
| **testify/assert + dockertest** | testify: 21k+ stars, fluent assertions; dockertest: automatic Docker container management | Requires Docker daemon | 9/10 |
| **testify/assert + testcontainers-go** | testcontainers: official Testcontainers port, more features | Heavier, slower startup | 7/10 |
| **ginkgo + gomega** | BDD-style testing, rich matchers | Non-standard Go testing style, steeper learning curve | 6/10 |

### Decision: **testify/assert + ory/dockertest**

**Rationale**:
- **testify/assert**: De facto standard for assertions in Go
  ```go
  import "github.com/stretchr/testify/assert"

  func TestMigration(t *testing.T) {
      assert.NoError(t, err)
      assert.Equal(t, expected, actual)
  }
  ```
- **dockertest**: Lightweight, fast container startup
  ```go
  import "github.com/ory/dockertest/v3"

  pool, _ := dockertest.NewPool("")
  resource, _ := pool.Run("postgres", "16", []string{"POSTGRES_PASSWORD=secret"})
  defer pool.Purge(resource) // Auto-cleanup
  ```
- **Integration**: Works with Go's standard `testing` package (no framework lock-in)
- **Parallel Safe**: Each test gets isolated Docker containers
- **CI/CD Ready**: Works in GitHub Actions (Docker-in-Docker)

---

## Summary Table

| Decision | Technology | Rationale (1 sentence) |
|----------|-----------|------------------------|
| Migration Framework | **golang-migrate/migrate** | Industry standard with CLI + library, built-in concurrency locks |
| PostgreSQL Driver | **jackc/pgx** | Best performance, rich error types, modern connection pooling |
| Redis Client | **go-redis/redis** | Most popular, easy graceful degradation, excellent docs |
| Secp256k1 Crypto | **btcsuite/btcd/btcec/v2** | Bitcoin-grade reliability, Circular Protocol compatible |
| Testing Framework | **testify + dockertest** | Standard Go assertions + lightweight Docker container management |

---

## Alternatives Not Pursued

### ORMs (GORM, ent, sqlboiler)
**Why Not**: This milestone requires only schema setup, no complex queries yet. Raw SQL migrations + pgx driver provide more control and transparency. Future milestones may introduce query builders if needed, but full ORMs add unnecessary abstraction.

### Configuration Libraries (viper, envconfig)
**Why Defer**: Spec requires environment variables for configuration. Standard `os.Getenv()` sufficient for this milestone. Future milestones (Milestone 5: Proxy) will evaluate structured config libraries when HTTP service requires complex configuration.

### Logging (zap, logrus)
**Why Defer**: Per constitution, structured logging deferred to Milestone 8 (Monitoring & Operations). This milestone uses standard `log` package for basic error output during development.

---

## Open Questions (for Future Milestones)

1. **Observability**: Prometheus client library selection deferred to Milestone 8
2. **HTTP Framework**: Gin vs. Echo decision deferred to Milestone 5 (Proxy HTTP API)
3. **MCP Library**: mcp-go evaluation deferred to Milestone 2 (x402 MCP Server)

---

## References

- golang-migrate: https://github.com/golang-migrate/migrate
- pgx: https://github.com/jackc/pgx
- go-redis: https://github.com/redis/go-redis
- btcsuite/btcd: https://github.com/btcsuite/btcd
- testify: https://github.com/stretchr/testify
- dockertest: https://github.com/ory/dockertest
