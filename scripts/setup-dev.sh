#!/usr/bin/env bash
# Development environment setup script
# Usage: ./scripts/setup-dev.sh

set -e

echo "🚀 Setting up Agents Notary development environment..."
echo ""

# Step 1: Check prerequisites
echo "📋 Checking prerequisites..."

# Check for container runtime (prefer Podman, fallback to Docker)
if command -v podman &> /dev/null; then
    CONTAINER_CMD="podman"
    COMPOSE_CMD="podman-compose"
    echo "✅ Podman found: $(podman --version) (daemonless!)"

    if ! command -v podman-compose &> /dev/null; then
        echo "⚠️  Warning: podman-compose not found"
        echo "Install it or use docker-compose as fallback"
    else
        echo "✅ Podman Compose found"
    fi
elif command -v docker &> /dev/null; then
    CONTAINER_CMD="docker"
    COMPOSE_CMD="docker-compose"
    echo "✅ Docker found: $(docker --version)"

    if ! command -v docker-compose &> /dev/null; then
        echo "❌ Error: docker-compose not found"
        echo "Install Docker Compose: https://docs.docker.com/compose/install/"
        exit 1
    fi
    echo "✅ Docker Compose found: $(docker-compose --version)"
else
    echo "❌ Error: No container runtime found"
    echo "Install one of:"
    echo "  - Podman (daemonless): Available in 'nix develop'"
    echo "  - Docker Desktop: https://docs.docker.com/get-docker/"
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo "❌ Error: Go not found"
    echo "Install Go 1.23+: https://go.dev/doc/install"
    echo "Or use: nix develop (if using Nix flake)"
    exit 1
fi
echo "✅ Go found: $(go version)"

if ! command -v migrate &> /dev/null; then
    echo "⚠️  Warning: golang-migrate CLI not found"
    echo "Install it with:"
    echo "  macOS: brew install golang-migrate"
    echo "  Linux: curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && sudo mv migrate /usr/local/bin/"
    echo ""
    read -p "Continue without migrate CLI? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo "✅ golang-migrate found: $(migrate -version)"
fi

echo ""

# Step 2: Create .env file if it doesn't exist
echo "📝 Setting up environment variables..."
if [ ! -f .env ]; then
    echo "Creating .env from .env.example..."
    cp .env.example .env
    echo "✅ .env file created"
    echo "⚠️  Review .env and update passwords/secrets as needed"
else
    echo "✅ .env file already exists"
fi

echo ""

# Step 3: Start container services
echo "🐳 Starting container services..."
echo "Using: $COMPOSE_CMD"

if [ "$CONTAINER_CMD" = "podman" ]; then
    # Initialize Podman socket if needed
    if ! podman info >/dev/null 2>&1; then
        echo "Initializing Podman..."
        ./scripts/setup-podman.sh
    fi
fi

$COMPOSE_CMD up -d

echo "Waiting for services to be healthy (max 60 seconds)..."
TIMEOUT=60
ELAPSED=0
INTERVAL=5

while [ $ELAPSED -lt $TIMEOUT ]; do
    sleep $INTERVAL
    ELAPSED=$((ELAPSED + INTERVAL))

    # Check if both services are healthy
    POSTGRES_HEALTHY=$($CONTAINER_CMD inspect --format='{{.State.Health.Status}}' certify-postgres 2>/dev/null || echo "unknown")
    REDIS_HEALTHY=$($CONTAINER_CMD inspect --format='{{.State.Health.Status}}' certify-redis 2>/dev/null || echo "unknown")

    if [ "$POSTGRES_HEALTHY" = "healthy" ] && [ "$REDIS_HEALTHY" = "healthy" ]; then
        echo "✅ All services healthy"
        break
    fi

    echo "Waiting... (PostgreSQL: $POSTGRES_HEALTHY, Redis: $REDIS_HEALTHY)"
done

if [ "$POSTGRES_HEALTHY" != "healthy" ] || [ "$REDIS_HEALTHY" != "healthy" ]; then
    echo "⚠️  Warning: Services may not be fully healthy yet"
    echo "PostgreSQL: $POSTGRES_HEALTHY"
    echo "Redis: $REDIS_HEALTHY"
    echo "Check logs with: $COMPOSE_CMD logs"
fi

echo ""

# Step 4: Run database migrations
if command -v migrate &> /dev/null; then
    echo "📊 Running database migrations..."

    # Load DATABASE_URL from .env
    if [ -f .env ]; then
        export $(grep DATABASE_URL .env | xargs)
    fi

    if [ -z "$DATABASE_URL" ]; then
        echo "❌ Error: DATABASE_URL not set in .env"
        exit 1
    fi

    ./scripts/migrate.sh up
    echo "✅ Migrations completed"
else
    echo "⚠️  Skipping migrations (migrate CLI not installed)"
    echo "Run migrations manually: ./scripts/migrate.sh up"
fi

echo ""

# Step 5: Download Go dependencies
echo "📦 Downloading Go dependencies..."
go mod download
go mod tidy
echo "✅ Go dependencies ready"

echo ""

# Step 6: Run health check
echo "🏥 Running health check..."
if [ -f ./scripts/health-check.sh ]; then
    ./scripts/health-check.sh
else
    echo "⚠️  Health check script not found"
    echo "Verifying services manually..."
    $COMPOSE_CMD ps
fi

echo ""
echo "🎉 Development environment setup complete!"
echo ""
echo "Next steps:"
echo "  - Review .env file and update secrets"
echo "  - Run tests: make test"
echo "  - View logs: make logs"
echo "  - Stop services: make down"
echo ""
echo "See docs/OVERVIEW.md for architecture details"
echo "See specs/001-foundation-setup/quickstart.md for detailed guide"
