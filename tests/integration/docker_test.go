package integration

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDockerComposeServiceStartup verifies that docker-compose can start
// PostgreSQL and Redis services successfully with health checks passing
func TestDockerComposeServiceStartup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Step 1: Start services
	t.Log("Starting Docker Compose services...")
	upCmd := exec.CommandContext(ctx, "docker-compose", "up", "-d")
	upOutput, err := upCmd.CombinedOutput()
	require.NoError(t, err, "docker-compose up failed: %s", string(upOutput))

	// Cleanup: Stop services after test
	defer func() {
		t.Log("Stopping Docker Compose services...")
		downCmd := exec.Command("docker-compose", "down", "-v")
		downOutput, _ := downCmd.CombinedOutput()
		t.Logf("docker-compose down output: %s", string(downOutput))
	}()

	// Step 2: Wait for health checks to pass
	t.Log("Waiting for health checks to pass (max 60 seconds)...")
	healthyPostgres := false
	healthyRedis := false

	for i := 0; i < 12; i++ { // 12 attempts * 5 seconds = 60 seconds max
		time.Sleep(5 * time.Second)

		psCmd := exec.Command("docker-compose", "ps", "--format", "json")
		psOutput, err := psCmd.CombinedOutput()
		if err != nil {
			t.Logf("docker-compose ps failed (attempt %d/12): %v", i+1, err)
			continue
		}

		output := string(psOutput)
		if contains(output, "certify-postgres") && contains(output, "healthy") {
			healthyPostgres = true
		}
		if contains(output, "certify-redis") && contains(output, "healthy") {
			healthyRedis = true
		}

		if healthyPostgres && healthyRedis {
			t.Log("✅ All services healthy")
			break
		}

		t.Logf("Waiting... (Postgres: %v, Redis: %v)", healthyPostgres, healthyRedis)
	}

	// Step 3: Verify both services are healthy
	assert.True(t, healthyPostgres, "PostgreSQL service should be healthy")
	assert.True(t, healthyRedis, "Redis service should be healthy")

	// Step 4: Verify PostgreSQL port is accessible
	t.Log("Verifying PostgreSQL connection...")
	pgCmd := exec.CommandContext(ctx, "docker", "exec", "certify-postgres", "pg_isready", "-U", "certify")
	pgOutput, err := pgCmd.CombinedOutput()
	assert.NoError(t, err, "PostgreSQL should be accepting connections: %s", string(pgOutput))

	// Step 5: Verify Redis ping works
	t.Log("Verifying Redis connection...")
	redisCmd := exec.CommandContext(ctx, "docker", "exec", "certify-redis", "redis-cli", "ping")
	redisOutput, err := redisCmd.CombinedOutput()
	assert.NoError(t, err, "Redis should respond to PING: %s", string(redisOutput))
	assert.Contains(t, string(redisOutput), "PONG", "Redis should return PONG")
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && hasSubstring(s, substr))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
