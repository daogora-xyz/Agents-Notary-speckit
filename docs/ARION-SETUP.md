# Arion Setup Guide

Arion is a NixOS-based Docker Compose tool that provides better integration with the Nix ecosystem. This guide explains how to use Arion for the Agents Notary project.

## What is Arion?

Arion allows you to define Docker Compose configurations using Nix expressions instead of YAML. Benefits include:

- **Type safety**: Nix catches configuration errors at evaluation time
- **Reproducibility**: Nix ensures consistent container configurations
- **Integration**: Better integration with Nix development environments
- **Backwards compatible**: Arion can generate standard `docker-compose.yml` files

## Files

The Arion configuration consists of:

- `arion-compose.nix` - Main service definitions (PostgreSQL, Redis)
- `arion-pkgs.nix` - Nixpkgs import for Arion
- `flake.nix` - Includes Arion as a development tool

## Basic Usage

### Starting Services

```bash
# Enter the Nix development environment first
nix develop

# Start all services in detached mode
arion up -d

# View service status
arion ps
```

### Viewing Logs

```bash
# Follow logs from all services
arion logs -f

# View logs from specific service
arion logs postgres
arion logs redis
```

### Stopping Services

```bash
# Stop services (keeps data volumes)
arion down

# Stop services and remove volumes (clean slate)
arion down -v
```

## Configuration

### Service Definitions

The `arion-compose.nix` file defines two services:

#### PostgreSQL 16
- **Image**: `postgres:16-alpine`
- **Container name**: `certify-postgres`
- **Port**: 5432 (configurable via `POSTGRES_PORT`)
- **Volumes**: `postgres_data` for persistent storage
- **Health check**: `pg_isready` command

#### Redis 7
- **Image**: `redis:7-alpine`
- **Container name**: `certify-redis`
- **Port**: 6379 (configurable via `REDIS_PORT`)
- **Configuration**:
  - Max memory: 256MB
  - Eviction policy: `allkeys-lru` (per spec requirement)
- **Volumes**: `redis_data` for persistent storage
- **Health check**: `redis-cli ping` command

### Environment Variables

Arion reads environment variables from your `.env` file:

```bash
# Database
POSTGRES_USER=certify
POSTGRES_PASSWORD=dev_password_change_in_production
POSTGRES_DB=certify_platform
POSTGRES_PORT=5432

# Redis
REDIS_PORT=6379
```

## Makefile Integration

The project's `Makefile` automatically detects Arion:

```bash
# These commands use Arion if available, docker-compose otherwise
make up      # Start services
make down    # Stop services
make logs    # View logs
make clean   # Clean up volumes
```

To see which backend is being used:

```bash
make help
# Output will show: Docker backend: arion
```

## Troubleshooting

### Arion not found

If you get "command not found: arion", ensure you're in the Nix development environment:

```bash
nix develop
# Now arion should be available
```

### Services not starting

Check if Docker daemon is running:

```bash
docker info
```

View detailed logs:

```bash
arion logs
```

### Port conflicts

If ports 5432 or 6379 are already in use, update your `.env` file:

```bash
POSTGRES_PORT=5433
REDIS_PORT=6380
```

Then restart:

```bash
arion down
arion up -d
```

### Clean slate restart

To completely reset the environment:

```bash
arion down -v                    # Remove containers and volumes
rm -rf .postgres .redis          # Remove local data directories
cp .env.example .env             # Reset environment variables
arion up -d                      # Start fresh
./scripts/migrate.sh up          # Re-run migrations
```

## Advanced: Modifying Services

To add or modify services, edit `arion-compose.nix`:

```nix
services = {
  # ... existing services ...

  myservice = {
    service.image = "myimage:latest";
    service.container_name = "my-container";
    service.ports = [ "8080:8080" ];
  };
};
```

After editing, rebuild with:

```bash
arion up -d
```

## Comparison: Arion vs Docker Compose

| Feature | Arion | docker-compose |
|---------|-------|----------------|
| Configuration | Nix expressions | YAML |
| Type checking | Yes | No |
| Nix integration | Native | Via wrapper |
| Learning curve | Steeper (requires Nix knowledge) | Gentler |
| Error messages | Nix evaluation errors | Runtime errors |

## References

- [Arion Documentation](https://docs.hercules-ci.com/arion/)
- [Arion GitHub](https://github.com/hercules-ci/arion)
- [NixOS Docker Integration](https://nixos.wiki/wiki/Docker)
