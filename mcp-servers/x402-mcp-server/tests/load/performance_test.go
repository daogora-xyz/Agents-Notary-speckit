package load

import (
	"testing"
	"time"

	"github.com/yourusername/x402-mcp-server/internal/config"
	"github.com/yourusername/x402-mcp-server/internal/x402"
	"github.com/yourusername/x402-mcp-server/tools"
)

// TestToolResponseTimes benchmarks individual tool response times
func TestToolResponseTimes(t *testing.T) {
	cfg, err := config.LoadConfig("../../config.yaml.example")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	t.Run("CreatePaymentRequirement", func(t *testing.T) {
		testCreatePaymentRequirementPerformance(t, cfg)
	})
}

func testCreatePaymentRequirementPerformance(t *testing.T, cfg *config.Config) {
	generator := x402.NewPaymentRequirementGenerator(cfg)
	tool := tools.NewCreatePaymentRequirementTool(generator)

	iterations := 100
	durations := make([]time.Duration, iterations)

	// Warm-up
	args := map[string]interface{}{
		"amount":  "100000",
		"network": "base-sepolia",
	}
	_, _ = tool.Execute(args)

	// Run performance test
	for i := 0; i < iterations; i++ {
		start := time.Now()
		_, err := tool.Execute(args)
		durations[i] = time.Since(start)

		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}
	}

	// Calculate statistics
	var total time.Duration
	min := durations[0]
	max := durations[0]

	for _, d := range durations {
		total += d
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
	}

	avg := total / time.Duration(iterations)

	// Calculate p95 and p99
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	p95Index := int(float64(iterations) * 0.95)
	p99Index := int(float64(iterations) * 0.99)
	p95 := sorted[p95Index]
	p99 := sorted[p99Index]

	// Report results
	t.Logf("=== CreatePaymentRequirement Performance ===")
	t.Logf("Iterations: %d", iterations)
	t.Logf("Avg Response Time: %v", avg)
	t.Logf("Min Response Time: %v", min)
	t.Logf("Max Response Time: %v", max)
	t.Logf("P95 Response Time: %v", p95)
	t.Logf("P99 Response Time: %v", p99)

	// Performance assertions
	if avg > 100*time.Millisecond {
		t.Errorf("Average response time too high: %v (expected < 100ms)", avg)
	}

	if p95 > 200*time.Millisecond {
		t.Errorf("P95 response time too high: %v (expected < 200ms)", p95)
	}

	if p99 > 500*time.Millisecond {
		t.Errorf("P99 response time too high: %v (expected < 500ms)", p99)
	}
}

// BenchmarkCreatePaymentRequirement benchmarks the create payment requirement tool
func BenchmarkCreatePaymentRequirement(b *testing.B) {
	cfg, err := config.LoadConfig("../../config.yaml.example")
	if err != nil {
		b.Fatalf("Failed to load config: %v", err)
	}

	generator := x402.NewPaymentRequirementGenerator(cfg)
	tool := tools.NewCreatePaymentRequirementTool(generator)

	args := map[string]interface{}{
		"amount":  "100000",
		"network": "base-sepolia",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tool.Execute(args)
		if err != nil {
			b.Fatalf("Failed to execute tool: %v", err)
		}
	}
}

// BenchmarkPaymentRequirementGeneration benchmarks just the generation logic
func BenchmarkPaymentRequirementGeneration(b *testing.B) {
	cfg, err := config.LoadConfig("../../config.yaml.example")
	if err != nil {
		b.Fatalf("Failed to load config: %v", err)
	}

	generator := x402.NewPaymentRequirementGenerator(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.Generate("base-sepolia", "100000", "https://api.example.com/resource", "Payment for service")
		if err != nil {
			b.Fatalf("Failed to generate payment requirement: %v", err)
		}
	}
}
