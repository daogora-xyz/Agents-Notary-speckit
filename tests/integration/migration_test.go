package integration

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigrationUpDown verifies that database migrations can be applied
// and rolled back successfully
func TestMigrationUpDown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Setup: Start Docker Compose services
	t.Log("Starting Docker Compose services...")
	upCmd := exec.CommandContext(ctx, "docker-compose", "up", "-d")
	upOutput, err := upCmd.CombinedOutput()
	require.NoError(t, err, "docker-compose up failed: %s", string(upOutput))

	defer func() {
		t.Log("Cleaning up Docker Compose services...")
		downCmd := exec.Command("docker-compose", "down", "-v")
		downCmd.Run()
	}()

	// Wait for PostgreSQL to be ready
	t.Log("Waiting for PostgreSQL to be ready...")
	time.Sleep(10 * time.Second)

	// Get database connection string
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://certify:dev_password_change_in_production@localhost:5432/certify_platform?sslmode=disable"
	}

	// Test 1: Apply migrations (up)
	t.Log("Running migrations (up)...")
	migrateUpCmd := exec.CommandContext(ctx, "./scripts/migrate.sh", "up")
	migrateUpCmd.Env = append(os.Environ(), "DATABASE_URL="+databaseURL)
	migrateUpOutput, err := migrateUpCmd.CombinedOutput()
	require.NoError(t, err, "migration up failed: %s", string(migrateUpOutput))

	// Test 2: Verify tables were created
	t.Log("Verifying tables were created...")
	db, err := sql.Open("postgres", databaseURL)
	require.NoError(t, err, "failed to connect to database")
	defer db.Close()

	expectedTables := []string{
		"certification_requests",
		"payments",
		"certifications",
		"wallet_balances",
	}

	for _, tableName := range expectedTables {
		var exists bool
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = $1
		)`
		err := db.QueryRowContext(ctx, query, tableName).Scan(&exists)
		require.NoError(t, err, "failed to check if table %s exists", tableName)
		assert.True(t, exists, "table %s should exist after migration up", tableName)
	}

	// Test 3: Verify migration version
	t.Log("Checking migration version...")
	var version int
	var dirty bool
	err = db.QueryRowContext(ctx, "SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty)
	require.NoError(t, err, "failed to read schema_migrations table")
	assert.Equal(t, 1, version, "migration version should be 1")
	assert.False(t, dirty, "migration should not be dirty")

	// Test 4: Rollback migrations (down)
	t.Log("Rolling back migrations (down)...")
	migrateDownCmd := exec.CommandContext(ctx, "./scripts/migrate.sh", "down")
	migrateDownCmd.Env = append(os.Environ(), "DATABASE_URL="+databaseURL)
	migrateDownOutput, err := migrateDownCmd.CombinedOutput()
	require.NoError(t, err, "migration down failed: %s", string(migrateDownOutput))

	// Test 5: Verify tables were dropped
	t.Log("Verifying tables were dropped...")
	for _, tableName := range expectedTables {
		var exists bool
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = $1
		)`
		err := db.QueryRowContext(ctx, query, tableName).Scan(&exists)
		require.NoError(t, err, "failed to check if table %s exists", tableName)
		assert.False(t, exists, "table %s should NOT exist after migration down", tableName)
	}

	t.Log("✅ Migration up/down test completed successfully")
}

// TestMigrationFailureBlocksStartup verifies that failed migrations prevent
// the environment from starting (per spec clarification)
func TestMigrationFailureBlocksStartup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("TODO: Implement migration failure test - requires intentionally breaking migration")
	// This test should:
	// 1. Create a migration file with invalid SQL
	// 2. Attempt to run migrations
	// 3. Verify that migration fails with clear error message
	// 4. Verify that health check script detects failed migration
	// 5. Verify that services cannot start until migration is fixed
}
