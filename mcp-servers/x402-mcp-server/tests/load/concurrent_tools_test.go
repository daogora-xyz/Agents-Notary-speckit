package load

import (
	"sync"
	"testing"
	"time"

	"github.com/yourusername/x402-mcp-server/internal/config"
	"github.com/yourusername/x402-mcp-server/internal/x402"
	"github.com/yourusername/x402-mcp-server/tools"
)

// TestConcurrentToolCalls tests the system under concurrent load
func TestConcurrentToolCalls(t *testing.T) {
	// Load config
	cfg, err := config.LoadConfig("../../config.yaml.example")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test with different concurrency levels
	concurrencyLevels := []int{10, 50, 100}

	for _, concurrency := range concurrencyLevels {
		conc := concurrency
		t.Run("Concurrency", func(t *testing.T) {
			testConcurrentPaymentRequirements(t, cfg, conc)
		})
	}
}

func testConcurrentPaymentRequirements(t *testing.T, cfg *config.Config, concurrency int) {
	var wg sync.WaitGroup
	errors := make(chan error, concurrency)
	durations := make(chan time.Duration, concurrency)

	// Create payment requirement generator
	generator := x402.NewPaymentRequirementGenerator(cfg)
	tool := tools.NewCreatePaymentRequirementTool(generator)

	// Launch concurrent requests
	startTime := time.Now()
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()

			reqStart := time.Now()

			// Execute tool
			args := map[string]interface{}{
				"amount":  "100000",
				"network": "base-sepolia",
			}

			result, err := tool.Execute(args)
			reqDuration := time.Since(reqStart)
			durations <- reqDuration

			if err != nil {
				errors <- err
				return
			}

			// Validate result
			resultMap, ok := result.(map[string]interface{})
			if !ok {
				errors <- err
				return
			}

			if resultMap["network"] != "base-sepolia" {
				errors <- err
				return
			}
		}(i)
	}

	// Wait for all to complete
	wg.Wait()
	close(errors)
	close(durations)
	totalDuration := time.Since(startTime)

	// Check for errors
	errorCount := 0
	for err := range errors {
		errorCount++
		t.Logf("Error during concurrent execution: %v", err)
	}

	// Calculate statistics
	var sumDuration time.Duration
	var maxDuration time.Duration
	var minDuration time.Duration = time.Hour
	count := 0

	for d := range durations {
		count++
		sumDuration += d
		if d > maxDuration {
			maxDuration = d
		}
		if d < minDuration {
			minDuration = d
		}
	}

	avgDuration := sumDuration / time.Duration(count)
	throughput := float64(concurrency) / totalDuration.Seconds()

	// Report results
	t.Logf("Concurrency: %d", concurrency)
	t.Logf("Total Duration: %v", totalDuration)
	t.Logf("Avg Response Time: %v", avgDuration)
	t.Logf("Min Response Time: %v", minDuration)
	t.Logf("Max Response Time: %v", maxDuration)
	t.Logf("Throughput: %.2f req/sec", throughput)
	t.Logf("Error Rate: %d/%d (%.2f%%)", errorCount, concurrency, float64(errorCount)/float64(concurrency)*100)

	// Assert acceptable performance
	if errorCount > 0 {
		t.Errorf("Found %d errors during concurrent execution", errorCount)
	}

	// Assert reasonable response times (< 1 second average)
	if avgDuration > time.Second {
		t.Errorf("Average response time too high: %v (expected < 1s)", avgDuration)
	}
}

// TestConcurrentMixedTools tests concurrent calls to different tools
func TestConcurrentMixedTools(t *testing.T) {
	cfg, err := config.LoadConfig("../../config.yaml.example")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	var wg sync.WaitGroup
	concurrency := 30 // 10 of each tool type
	errors := make(chan error, concurrency)

	generator := x402.NewPaymentRequirementGenerator(cfg)
	createTool := tools.NewCreatePaymentRequirementTool(generator)

	startTime := time.Now()

	// Launch concurrent requests for create_payment_requirement
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			args := map[string]interface{}{
				"amount":  "50000",
				"network": "base",
			}
			_, err := createTool.Execute(args)
			if err != nil {
				errors <- err
			}
		}()
	}

	// Launch concurrent requests for verify_payment (mock scenarios)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Simulate verify_payment work (actual test would need valid signatures)
			time.Sleep(10 * time.Millisecond)
		}()
	}

	// Launch concurrent requests for settle_payment (mock scenarios)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Simulate settle_payment work (actual test would need valid blockchain state)
			time.Sleep(10 * time.Millisecond)
		}()
	}

	wg.Wait()
	close(errors)
	duration := time.Since(startTime)

	// Check for errors
	errorCount := 0
	for err := range errors {
		errorCount++
		t.Logf("Error: %v", err)
	}

	t.Logf("Mixed tool test completed in %v", duration)
	t.Logf("Errors: %d/%d", errorCount, concurrency)

	if errorCount > 0 {
		t.Errorf("Found %d errors during mixed tool execution", errorCount)
	}
}
