# Podman Setup Guide - Daemonless Containers

This guide explains how to use **Podman** for the Agents Notary project. Podman is a daemonless container runtime that works without root privileges - perfect for NixOS and environments where Docker daemon isn't available.

## What is Podman?

Podman is a drop-in replacement for Docker that:

- ✅ **Runs without a daemon** - No background service required
- ✅ **Rootless by default** - More secure, no root privileges needed
- ✅ **Docker-compatible** - Same commands and syntax as Docker
- ✅ **Works in Nix** - Fully supported in the nix develop environment
- ✅ **No system rebuild** - Install and use immediately

## Quick Start

### 1. Enter Nix Environment

```bash
nix develop
# Podman is now available!
```

### 2. Initialize Podman (First Time Only)

```bash
./scripts/setup-podman.sh
```

This script:
- Creates Podman socket directory
- Starts rootless Podman service
- Configures Docker compatibility
- Tests the Podman installation

### 3. Start Services

```bash
# Use podman-compose directly
podman-compose up -d

# Or use the Makefile (auto-detects Podman)
make up
```

### 4. Verify Services

```bash
podman ps
# Should show certify-postgres and certify-redis containers
```

## Common Commands

### Container Management

```bash
# List running containers
podman ps

# List all containers (including stopped)
podman ps -a

# View logs
podman logs certify-postgres
podman logs certify-redis

# Follow logs (like tail -f)
podman logs -f certify-postgres

# Stop containers
podman stop certify-postgres certify-redis

# Remove containers
podman rm certify-postgres certify-redis
```

### Compose Operations

```bash
# Start all services
podman-compose up -d

# Stop all services
podman-compose down

# View logs from all services
podman-compose logs -f

# Restart specific service
podman-compose restart postgres

# Pull latest images
podman-compose pull
```

### Makefile Commands (Recommended)

The Makefile automatically detects Podman:

```bash
make up      # Start services (uses podman-compose if available)
make down    # Stop services
make logs    # View logs
make ps      # List containers
make clean   # Remove containers and volumes
```

## Configuration

### Environment Variables

Podman reads from `.env` file just like Docker Compose:

```bash
# .env
POSTGRES_PORT=5432
REDIS_PORT=6379
POSTGRES_PASSWORD=dev_password_change_in_production
```

### Docker Compatibility

Podman can emulate Docker commands:

```bash
# Set Docker compatibility in your shell
export DOCKER_HOST="unix:///run/user/$(id -u)/podman/podman.sock"

# Now Docker commands use Podman
docker ps      # Actually runs: podman ps
docker images  # Actually runs: podman images
```

This is automatically set in `nix develop`.

## Troubleshooting

### Podman Not Found

**Solution**: Ensure you're in the Nix development environment:

```bash
nix develop
podman --version
```

### Permission Denied

Podman runs rootless, but if you see permission errors:

```bash
# Check Podman info
podman info

# Verify rootless mode
podman info | grep rootless
# Should show: true
```

### Services Not Starting

**1. Check Podman socket:**

```bash
podman system service --time=0 &
```

**2. Check container status:**

```bash
podman-compose ps
```

**3. View detailed logs:**

```bash
podman-compose logs
```

### Port Already in Use

If ports 5432 or 6379 are in use:

```bash
# Edit .env
POSTGRES_PORT=5433
REDIS_PORT=6380

# Restart services
podman-compose down
podman-compose up -d
```

### Clean Slate Restart

To completely reset:

```bash
# Stop and remove everything
podman-compose down -v

# Remove all Podman containers (optional)
podman rm -af

# Remove all volumes (optional)
podman volume prune -f

# Start fresh
podman-compose up -d
```

## Podman vs Docker Comparison

| Feature | Podman | Docker |
|---------|--------|--------|
| Daemon | ❌ No (daemonless) | ✅ Yes (required) |
| Root access | ❌ Not required | ⚠️ Often required |
| System rebuild | ❌ Not required | ⚠️ May require |
| NixOS friendly | ✅ Yes | ⚠️ Complicated |
| Command syntax | Same as Docker | Standard |
| docker-compose | Via podman-compose | Native |

## Advanced: Podman-Specific Features

### Pods (Like Kubernetes Pods)

Podman supports pods - groups of containers:

```bash
# Create a pod
podman pod create --name certify-pod -p 5432:5432 -p 6379:6379

# Run containers in the pod
podman run -d --pod certify-pod postgres:16-alpine
podman run -d --pod certify-pod redis:7-alpine
```

### Systemd Integration

Generate systemd units for containers:

```bash
# Generate user systemd unit
podman generate systemd --new --name certify-postgres > ~/.config/systemd/user/certify-postgres.service

# Enable on boot
systemctl --user enable certify-postgres
systemctl --user start certify-postgres
```

### Image Management

```bash
# List images
podman images

# Remove unused images
podman image prune

# Pull specific image
podman pull postgres:16-alpine

# Save image to file
podman save -o postgres.tar postgres:16-alpine

# Load image from file
podman load -i postgres.tar
```

## Integration Tests with Podman

The integration tests work with Podman:

```bash
# Set Docker compatibility
export DOCKER_HOST="unix:///run/user/$(id -u)/podman/podman.sock"

# Run integration tests
go test ./tests/integration/... -v
```

Dockertest library (used in integration tests) automatically detects Podman via `DOCKER_HOST`.

## Migrating from Docker to Podman

If you were using Docker:

1. **Stop Docker services:**
   ```bash
   docker-compose down
   ```

2. **Enter Nix environment:**
   ```bash
   nix develop
   ```

3. **Initialize Podman:**
   ```bash
   ./scripts/setup-podman.sh
   ```

4. **Start with Podman:**
   ```bash
   podman-compose up -d
   # Or: make up
   ```

5. **Verify migration:**
   ```bash
   podman ps
   ./scripts/health-check.sh
   ```

## Why Podman for This Project?

1. **No system rebuild** - Works immediately in nix develop
2. **Rootless** - More secure, better for development
3. **NixOS friendly** - Native support, no daemon complications
4. **Resource efficient** - No background daemon consuming resources
5. **Docker compatible** - Existing docker-compose.yml works as-is

## References

- [Podman Documentation](https://docs.podman.io/)
- [Podman vs Docker](https://docs.podman.io/en/latest/Introduction.html)
- [Rootless Podman](https://github.com/containers/podman/blob/main/docs/tutorials/rootless_tutorial.md)
- [podman-compose](https://github.com/containers/podman-compose)
