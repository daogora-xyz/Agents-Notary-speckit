#!/usr/bin/env bash
# Migration wrapper script for golang-migrate CLI
# Usage: ./scripts/migrate.sh [up|down|version|force VERSION]

set -e

# Load environment variables from .env if it exists
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

# Check if migrate CLI is installed
if ! command -v migrate &> /dev/null; then
    echo "Error: golang-migrate CLI not found"
    echo "Install it with:"
    echo "  macOS: brew install golang-migrate"
    echo "  Linux: curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && sudo mv migrate /usr/local/bin/"
    exit 1
fi

# Validate DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
    echo "Error: DATABASE_URL not set"
    echo "Set it in .env file or export it:"
    echo "  export DATABASE_URL='postgres://user:pass@localhost:5432/dbname?sslmode=disable'"
    exit 1
fi

# Migration directory
MIGRATIONS_PATH="./migrations"

# Check migrations directory exists
if [ ! -d "$MIGRATIONS_PATH" ]; then
    echo "Error: Migrations directory not found: $MIGRATIONS_PATH"
    exit 1
fi

# Parse command
COMMAND=${1:-up}

case "$COMMAND" in
    up)
        echo "Running migrations (up)..."
        migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" up
        echo "✅ Migrations applied successfully"
        ;;
    down)
        echo "Rolling back last migration (down)..."
        migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" down 1
        echo "✅ Migration rolled back successfully"
        ;;
    version)
        echo "Checking migration version..."
        migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" version
        ;;
    force)
        VERSION=$2
        if [ -z "$VERSION" ]; then
            echo "Error: force command requires version number"
            echo "Usage: ./scripts/migrate.sh force VERSION"
            exit 1
        fi
        echo "Forcing migration version to $VERSION..."
        migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" force "$VERSION"
        echo "✅ Migration version forced to $VERSION"
        ;;
    *)
        echo "Usage: ./scripts/migrate.sh [up|down|version|force VERSION]"
        echo ""
        echo "Commands:"
        echo "  up       - Apply all pending migrations"
        echo "  down     - Rollback the last migration"
        echo "  version  - Show current migration version"
        echo "  force N  - Force set migration version to N (use with caution)"
        exit 1
        ;;
esac
