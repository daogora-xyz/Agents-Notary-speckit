#!/usr/bin/env bash
# Health check script for development environment
# Usage: ./scripts/health-check.sh
# Exit code: 0 if all critical checks pass, 1 if critical failure

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Track overall health status
ALL_HEALTHY=true

echo "🏥 Agents Notary Health Check"
echo "=============================="
echo ""

# Load environment variables from .env if it exists
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs 2>/dev/null || true)
fi

# Check 1: PostgreSQL connection
echo -n "PostgreSQL: "
if docker exec certify-postgres pg_isready -U "${POSTGRES_USER:-certify}" &> /dev/null; then
    # Get PostgreSQL version
    PG_VERSION=$(docker exec certify-postgres psql -U "${POSTGRES_USER:-certify}" -d "${POSTGRES_DB:-certify_platform}" -t -c "SELECT version();" 2>/dev/null | head -1 | awk '{print $2}')
    echo -e "${GREEN}✅ Connected${NC} (version $PG_VERSION)"
else
    echo -e "${RED}❌ Not connected${NC}"
    ALL_HEALTHY=false
fi

# Check 2: Redis connection
echo -n "Redis: "
if docker exec certify-redis redis-cli ping &> /dev/null; then
    # Get Redis version
    REDIS_VERSION=$(docker exec certify-redis redis-cli INFO server 2>/dev/null | grep redis_version | cut -d: -f2 | tr -d '\r')
    echo -e "${GREEN}✅ Connected${NC} (version $REDIS_VERSION)"
else
    # Per spec clarification: Redis unavailability is graceful degradation, not critical failure
    echo -e "${YELLOW}⚠️  Unavailable (graceful degradation mode)${NC}"
    echo "   Services will operate without cache (degraded performance)"
fi

# Check 3: Database migration status
echo -n "Migrations: "
if [ -z "$DATABASE_URL" ]; then
    DATABASE_URL="postgres://${POSTGRES_USER:-certify}:${POSTGRES_PASSWORD:-dev_password_change_in_production}@${POSTGRES_HOST:-localhost}:${POSTGRES_PORT:-5432}/${POSTGRES_DB:-certify_platform}?sslmode=disable"
fi

if command -v migrate &> /dev/null; then
    MIGRATION_OUTPUT=$(migrate -path ./migrations -database "$DATABASE_URL" version 2>&1 || true)

    if echo "$MIGRATION_OUTPUT" | grep -q "error"; then
        # Per spec clarification: Failed migrations MUST block environment startup
        echo -e "${RED}❌ Migration failed${NC}"
        echo "   $MIGRATION_OUTPUT"
        echo ""
        echo "⛔ CRITICAL: Failed migrations block environment startup"
        echo "   Fix the migration and re-run: ./scripts/migrate.sh up"
        ALL_HEALTHY=false
    elif echo "$MIGRATION_OUTPUT" | grep -q "dirty"; then
        echo -e "${RED}❌ Dirty migration state${NC}"
        echo "   Force version: migrate -path ./migrations -database \$DATABASE_URL force VERSION"
        ALL_HEALTHY=false
    else
        # Extract version number
        MIGRATION_VERSION=$(echo "$MIGRATION_OUTPUT" | grep -o '[0-9]\+' | head -1 || echo "0")
        echo -e "${GREEN}✅ Up-to-date${NC} (version $MIGRATION_VERSION)"

        # Verify expected tables exist
        EXPECTED_TABLES=("certification_requests" "payments" "certifications" "wallet_balances")
        for TABLE in "${EXPECTED_TABLES[@]}"; do
            TABLE_EXISTS=$(docker exec certify-postgres psql -U "${POSTGRES_USER:-certify}" -d "${POSTGRES_DB:-certify_platform}" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '$TABLE');" 2>/dev/null | xargs || echo "f")

            if [ "$TABLE_EXISTS" != "t" ]; then
                echo -e "   ${RED}❌ Table '$TABLE' missing${NC}"
                ALL_HEALTHY=false
            fi
        done
    fi
else
    echo -e "${YELLOW}⚠️  Cannot verify (migrate CLI not installed)${NC}"
    echo "   Install: brew install golang-migrate"
fi

# Check 4: Shared packages importable
echo -n "Shared packages: "
if [ -f go.mod ]; then
    # Check if Go packages can be imported (basic syntax check)
    if go list ./pkg/... &> /dev/null; then
        echo -e "${GREEN}✅ Importable${NC}"
    else
        echo -e "${RED}❌ Import errors${NC}"
        ALL_HEALTHY=false
    fi
else
    echo -e "${YELLOW}⚠️  go.mod not found${NC}"
fi

echo ""
echo "=============================="

# Final status
if [ "$ALL_HEALTHY" = true ]; then
    echo -e "${GREEN}✅ Environment: READY${NC}"
    echo ""
    echo "All critical checks passed. You can start developing!"
    echo ""
    echo "Quick commands:"
    echo "  - Run tests: make test"
    echo "  - View logs: make logs"
    echo "  - Stop services: make down"
    exit 0
else
    echo -e "${RED}❌ Environment: NOT READY${NC}"
    echo ""
    echo "Some critical checks failed. Please fix the issues above before continuing."
    echo ""
    echo "Common troubleshooting:"
    echo "  - Check service logs: docker-compose logs"
    echo "  - Restart services: docker-compose down && docker-compose up -d"
    echo "  - Fix migrations: see specs/001-foundation-setup/quickstart.md"
    exit 1
fi
