# Quickstart Guide: Foundation Infrastructure Setup

**Feature**: Project Foundation Infrastructure
**Branch**: 001-foundation-setup
**Target Time**: < 30 minutes (per spec success criterion SC-010)

## Prerequisites

Before you begin, ensure your system meets these requirements:

- **Operating System**: Linux, macOS, or Windows with WSL2
- **RAM**: Minimum 8GB
- **Disk Space**: Minimum 10GB free
- **Required Software**:
  - Docker Desktop (20.10+) - [Install Guide](https://docs.docker.com/get-docker/)
  - Docker Compose (2.0+) - Included with Docker Desktop
  - Go 1.23+ - [Install Guide](https://go.dev/doc/install)
  - Make (optional, for convenience commands)

## Step 1: Clone the Repository

```bash
# Clone the repository
git clone <repository-url>
cd Agents-Notary-speckit

# Verify you're on the correct branch
# (Once git is set up - currently not detected per setup script)
```

## Step 2: Install golang-migrate CLI

The migration tool is required for database schema management.

### macOS
```bash
brew install golang-migrate
```

### Linux
```bash
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/migrate
chmod +x /usr/local/bin/migrate
```

### Windows (WSL2)
```bash
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/migrate
chmod +x /usr/local/bin/migrate
```

### Verify Installation
```bash
migrate -version
# Expected output: v4.17.0 (or higher)
```

## Step 3: Configure Environment Variables

Create a `.env` file in the project root from the template:

```bash
cp .env.example .env
```

Edit `.env` with your preferred text editor:

```bash
# Database Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=certify
POSTGRES_PASSWORD=dev_password_change_in_production
POSTGRES_DB=certify_platform
POSTGRES_MAX_CONNS=50

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Development Settings
LOG_LEVEL=debug
ENVIRONMENT=development
```

**Security Note**: Never commit `.env` to version control. It's already in `.gitignore`.

## Step 4: Start Infrastructure Services

Use Docker Compose to start PostgreSQL and Redis:

```bash
# Start services in detached mode
docker-compose up -d

# Verify services are running
docker-compose ps

# Expected output:
# NAME                 COMMAND                  SERVICE    STATUS    PORTS
# certify-postgres     "docker-entrypoint.s…"   postgres   Up        0.0.0.0:5432->5432/tcp
# certify-redis        "redis-server --maxm…"   redis      Up        0.0.0.0:6379->6379/tcp
```

**Troubleshooting**:
- **Port conflicts**: If ports 5432 or 6379 are in use, edit `docker-compose.yml` to use different ports
- **Docker not running**: Start Docker Desktop and wait for it to finish initializing

## Step 5: Run Database Migrations

Apply the initial database schema:

```bash
# Run migrations
migrate -path ./migrations \
        -database "postgres://certify:dev_password_change_in_production@localhost:5432/certify_platform?sslmode=disable" \
        up

# Expected output:
# 001/u init (123.456ms)
```

**Verify migrations succeeded**:

```bash
# Connect to PostgreSQL
docker exec -it certify-postgres psql -U certify -d certify_platform

# List tables
\dt

# Expected output:
#                   List of relations
#  Schema |         Name              | Type  | Owner
# --------+---------------------------+-------+--------
#  public | certification_requests    | table | certify
#  public | certifications            | table | certify
#  public | payments                  | table | certify
#  public | schema_migrations         | table | certify
#  public | wallet_balances           | table | certify

# Exit psql
\q
```

**Migration Failure Handling**:

Per spec clarification, failed migrations **block environment startup**. If a migration fails:

1. **Check error output**: Migration errors include SQL details and line numbers
2. **Fix the migration file**: Edit `migrations/001_init.up.sql` to resolve the issue
3. **Rollback if needed**:
   ```bash
   migrate -path ./migrations \
           -database "postgres://..." \
           down 1
   ```
4. **Re-run migration**:
   ```bash
   migrate -path ./migrations \
           -database "postgres://..." \
           up
   ```

**Do not proceed** to the next step until migrations succeed.

## Step 6: Initialize Go Modules

Install Go dependencies:

```bash
# Initialize Go module (if not already done)
go mod init github.com/<your-org>/Agents-Notary-speckit

# Add dependencies
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/pgxpool
go get github.com/redis/go-redis/v9
go get github.com/btcsuite/btcd/btcec/v2
go get github.com/stretchr/testify
go get github.com/ory/dockertest/v3
go get github.com/golang-migrate/migrate/v4

# Download dependencies
go mod download

# Verify no errors
go mod tidy
```

## Step 7: Verify Shared Packages

Test that shared packages can be imported and validated:

```bash
# Run unit tests for shared packages
go test ./pkg/models -v
go test ./pkg/crypto -v
go test ./pkg/errors -v

# Expected output:
# === RUN   TestCertificationRequestValidation
# --- PASS: TestCertificationRequestValidation (0.00s)
# === RUN   TestPaymentValidation
# --- PASS: TestPaymentValidation (0.00s)
# ...
# PASS
# ok      github.com/<your-org>/Agents-Notary-speckit/pkg/models    0.123s
```

## Step 8: Run Integration Tests

Verify database and cache operations work end-to-end:

```bash
# Run integration tests (requires Docker)
go test ./tests/integration -v

# Expected tests:
# === RUN   TestMigrationUpDown
# --- PASS: TestMigrationUpDown (2.34s)
# === RUN   TestCacheOperations
# --- PASS: TestCacheOperations (0.56s)
# === RUN   TestCacheGracefulDegradation
# --- PASS: TestCacheGracefulDegradation (0.12s)
# === RUN   TestModelSerialization
# --- PASS: TestModelSerialization (0.89s)
# PASS
# ok      github.com/<your-org>/Agents-Notary-speckit/tests/integration    4.123s
```

**Note**: Integration tests use `dockertest` to spin up isolated PostgreSQL and Redis containers. They will download Docker images on first run (~2 minutes).

## Step 9: Verify Environment Health

Run the health check script to verify all services are operational:

```bash
# Run health check
./scripts/health-check.sh

# Expected output:
# ✅ PostgreSQL: Connected (version 16.2)
# ✅ Redis: Connected (version 7.2.4)
# ✅ Migrations: Up-to-date (1 applied)
# ✅ Shared packages: Importable
# ✅ Environment: READY
```

## Step 10: Review Project Structure

Your development environment is now ready! Here's the project structure:

```
Agents-Notary-speckit/
├── .env                     # Environment variables (not in git)
├── docker-compose.yml       # Service orchestration
├── go.mod                   # Go dependencies
├── go.sum                   # Dependency checksums
├── Makefile                 # Convenience commands
├── README.md                # Project documentation
├── migrations/              # Database migrations
│   ├── 001_init.up.sql
│   └── 001_init.down.sql
├── pkg/                     # Shared packages
│   ├── models/              # Data models
│   ├── crypto/              # Secp256k1 utilities
│   └── errors/              # Custom error types
├── tests/                   # Test suite
│   ├── integration/
│   └── unit/
└── specs/                   # Feature specifications
    └── 001-foundation-setup/
```

## Common Commands (Makefile)

For convenience, common commands are available via `make`:

```bash
# Start all services
make up

# Stop all services
make down

# Run migrations
make migrate-up

# Rollback last migration
make migrate-down

# Run all tests
make test

# Run only unit tests
make test-unit

# Run only integration tests
make test-integration

# Health check
make health

# View logs
make logs

# Clean up (stop services, remove volumes)
make clean
```

## Troubleshooting

### Problem: "Cannot connect to Docker daemon"
**Solution**: Ensure Docker Desktop is running. On macOS, check the menu bar icon.

### Problem: "Port 5432 already in use"
**Solution**: Another PostgreSQL instance is running. Either stop it or change the port in `docker-compose.yml`:
```yaml
services:
  postgres:
    ports:
      - "5433:5432"  # Use port 5433 instead
```
Then update `.env`:
```
POSTGRES_PORT=5433
```

### Problem: "Migration failed: constraint violation"
**Solution**: This indicates a bug in the migration SQL. Check the error message for the specific constraint that failed. Common issues:
- Duplicate unique values
- NULL values in NOT NULL columns
- Invalid enum values in CHECK constraints

### Problem: "Redis unavailable" warning in logs
**Expected Behavior**: Per spec clarification, services continue operating without cache. This is graceful degradation, not an error. Verify Redis is running:
```bash
docker-compose ps redis
```

### Problem: Go module download fails
**Solution**: Check your internet connection and Go proxy settings:
```bash
go env GOPROXY
# Should be: https://proxy.golang.org,direct

# If behind corporate proxy:
export GOPROXY=https://your-corporate-proxy.com
```

## Next Steps

Now that your foundation environment is ready:

1. **Explore the codebase**: Read through `pkg/models/` to understand data structures
2. **Run tests in watch mode**: Use `go test ./... -v -count=1` to see live test output
3. **Start building features**: The next milestone (002-x402-mcp) builds on this foundation
4. **Read the data model**: See `specs/001-foundation-setup/data-model.md` for detailed schema docs

## Performance Benchmarks

After completing setup, your environment should meet these criteria (per spec success criteria):

- ✅ **SC-001**: Environment initialization < 5 minutes (typical: 2-3 minutes)
- ✅ **SC-003**: Database supports 50 concurrent connections
- ✅ **SC-004**: Cache operations < 10ms P95 (typical: 1-2ms)
- ✅ **SC-009**: Services start/stop cleanly without orphaned processes
- ✅ **SC-010**: New developer setup < 30 minutes

Run benchmarks:
```bash
go test ./tests/integration -bench=. -benchmem
```

## Support

If you encounter issues not covered in this guide:

1. Check the [troubleshooting section](#troubleshooting) above
2. Review `docs/OVERVIEW.md` for system architecture details
3. Read `specs/001-foundation-setup/plan.md` for implementation notes
4. Open an issue with:
   - Error message
   - Output of `docker-compose ps`
   - Output of `go version`
   - Operating system and version

## Success Checklist

Before moving to the next milestone, verify:

- ☐ Docker services running (`docker-compose ps` shows "Up")
- ☐ Migrations applied (`psql` shows 4 tables)
- ☐ Go modules downloaded (`go mod tidy` succeeds)
- ☐ Unit tests passing (`go test ./pkg/... -v`)
- ☐ Integration tests passing (`go test ./tests/integration -v`)
- ☐ Health check script succeeds (`./scripts/health-check.sh`)
- ☐ Can import shared packages in new code
- ☐ Environment setup completed in < 30 minutes

If all items are checked, congratulations! Your foundation infrastructure is ready for feature development. 🎉

Proceed to **Milestone 2: x402 MCP Server** when ready.
