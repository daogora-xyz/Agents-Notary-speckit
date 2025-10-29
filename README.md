# Agents Notary spec-kit

Development environment for the **certify.ar4s.com** Blockchain Certification Platform.

## Overview

This project implements a Nix flake that provides a reproducible development environment with **spec-kit** (GitHub's specification-driven development toolkit) and all necessary tools for building the blockchain certification platform described in `docs/OVERVIEW.md`.

## Quick Start

### Prerequisites

- Nix package manager with flakes enabled
- (Optional) direnv for automatic environment loading

### Using the Development Environment

#### Option 1: Using `nix develop` (Manual)

```bash
# Enter the development shell
nix develop

# You now have access to spec-kit and all development tools
specify --version
go version
```

#### Option 2: Using direnv (Automatic)

If you have direnv installed:

```bash
# Allow direnv for this project
direnv allow

# The environment will load automatically when you cd into the directory
```

## Available Tools

The development environment includes:

- **spec-kit** (`specify` command) - GitHub's toolkit for Spec-Driven Development
- **Go** (latest version) - For building the Go services
- **gopls**, **gotools**, **go-tools** - Go development utilities
- **PostgreSQL 16** - Database
- **Redis** - Caching and rate limiting
- **Docker** & **docker-compose** - Container management
- **git**, **make**, **jq**, **curl** - Standard development utilities

## Using spec-kit

Once in the development environment, you can use spec-kit commands:

```bash
# Initialize a new spec-driven project
specify init <PROJECT_NAME>

# Check your specifications
specify check

# View help
specify --help
```

## Project Structure

```
.
├── flake.nix          # Nix flake configuration
├── flake.lock         # Locked dependency versions
├── .envrc             # direnv configuration
├── docs/
│   └── OVERVIEW.md    # Project specification
└── README.md          # This file
```

## Foundation Infrastructure Setup (Milestone 1)

The foundation infrastructure has been implemented following TDD approach. See `specs/001-foundation-setup/` for complete documentation.

### Quick Setup (Podman - No Daemon Required!)

```bash
# 1. Enter development environment (includes Podman)
nix develop

# 2. Initialize Podman (first time only)
./scripts/setup-podman.sh

# 3. Create environment file
cp .env.example .env

# 4. Start services using Podman
podman-compose up -d
# Or use Makefile (auto-detects): make up

# 5. Run migrations
./scripts/migrate.sh up

# 6. Run health check
./scripts/health-check.sh

# 7. Run tests
make test
```

### Why Podman?

**Podman is the recommended container runtime** for this project because:

- ✅ **No daemon required** - Works immediately, no system rebuild needed
- ✅ **Rootless by default** - More secure, no root privileges
- ✅ **Included in Nix** - Available instantly in `nix develop`
- ✅ **Docker-compatible** - Same commands, same compose files
- ✅ **NixOS friendly** - Native support, zero configuration

See [docs/PODMAN-SETUP.md](docs/PODMAN-SETUP.md) for detailed guide.

### Alternative: Docker or Arion

If you prefer Docker or Arion, they work too:

```bash
# Docker Compose (requires Docker daemon)
docker-compose up -d

# Arion (NixOS Docker Compose)
arion up -d
```

The `Makefile` automatically detects: **Podman** > Arion > Docker

### What's Been Implemented

**Phase 1: Setup**
- ✅ Go module initialized
- ✅ Directory structure created (`pkg/`, `migrations/`, `tests/`, `scripts/`)
- ✅ Docker Compose with PostgreSQL 16 and Redis 7
- ✅ Environment variable template (`.env.example`)
- ✅ Makefile with common commands
- ✅ .gitignore for Go, Docker, and IDE files

**Phase 2: Foundational**
- ✅ Go dependencies added (pgx, go-redis, btcsuite/btcd, testify, dockertest)
- ✅ Database migrations (4 tables: certification_requests, payments, certifications, wallet_balances)
- ✅ Migration wrapper script (`scripts/migrate.sh`)

**Phase 3: User Story 1 - Development Environment**
- ✅ Integration tests for Docker Compose, migrations, health checks
- ✅ Setup script (`scripts/setup-dev.sh`)
- ✅ Health check script (`scripts/health-check.sh`)

**Phase 4 & 5: User Story 4 - Shared Code Utilities**
- ✅ CertificationRequest model with validation (`pkg/models/request.go`)
- ✅ Payment model with validation (`pkg/models/payment.go`)
- ✅ Certification model with validation (`pkg/models/certification.go`)
- ✅ WalletBalance model with validation (`pkg/models/wallet.go`)
- ✅ Secp256k1 crypto utilities (`pkg/crypto/secp256k1.go`)
- ✅ Custom error types (`pkg/errors/types.go`)
- ✅ **All unit tests passing** (100% of implemented features)

### Test Results

```
pkg/crypto:  PASS (3.878s) - Secp256k1 signing avg 1.25ms (spec req: <100ms)
pkg/errors:  PASS (0.068s) - All error types working correctly
pkg/models:  PASS (0.074s) - All model validations working correctly
```

### Architecture

```
Agents-Notary-speckit/
├── pkg/                    # Shared packages
│   ├── models/            # Data models with validation
│   ├── crypto/            # Secp256k1 signing utilities
│   └── errors/            # Custom error types
├── migrations/            # Database migrations
│   ├── 001_init.up.sql
│   └── 001_init.down.sql
├── tests/
│   ├── integration/       # Integration tests
│   └── unit/              # Unit tests (colocated)
├── scripts/
│   ├── setup-dev.sh       # Development environment setup
│   ├── health-check.sh    # Health check script
│   └── migrate.sh         # Migration wrapper
├── specs/001-foundation-setup/  # Feature specification
│   ├── spec.md            # Requirements & user stories
│   ├── plan.md            # Implementation plan
│   ├── tasks.md           # Task breakdown
│   ├── data-model.md      # Database schema
│   ├── research.md        # Technology selection
│   └── quickstart.md      # Developer guide
└── docker-compose.yml     # PostgreSQL + Redis services
```

### Documentation

- **Quickstart Guide**: `specs/001-foundation-setup/quickstart.md`
- **Data Model**: `specs/001-foundation-setup/data-model.md`
- **Technology Research**: `specs/001-foundation-setup/research.md`
- **Project Constitution**: `.specify/memory/constitution.md`

##  Next Steps

1. **Start Docker daemon** and run integration tests (requires system Docker)
2. **Implement User Story 2** (Data Persistence Layer - migrations integration)
3. **Implement User Story 3** (Caching Infrastructure - Redis client wrapper)
4. **Move to Milestone 2**: x402 MCP Server implementation
5. Review the full specification in `docs/OVERVIEW.md`

## Environment Variables

The development shell sets the following:

- `PGDATA=$PWD/.postgres` - PostgreSQL data directory
- `REDIS_DATA=$PWD/.redis` - Redis data directory

## Building from Source

The flake automatically builds spec-kit from the GitHub repository. If you need to update it:

```bash
# Update all flake inputs
nix flake update

# Rebuild the environment
nix develop
```

## Troubleshooting

### First build takes a long time

The first time you run `nix develop`, it will download and build all dependencies. This can take several minutes. Subsequent runs will be much faster due to caching.

### spec-kit not found

If the `specify` command is not available, the build may still be in progress. Wait for all dependencies to finish downloading and building.

## Learn More

- [spec-kit Repository](https://github.com/github/spec-kit)
- [Nix Flakes](https://nixos.wiki/wiki/Flakes)
- [direnv](https://direnv.net/)
- [Kiro Spec-Driven Development](https://martinfowler.com/articles/exploring-gen-ai/sdd-3-tools.html)

## License

See project specification for details.
