package main

import (
	"fmt"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/config"
	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/x402"
)

func main() {
	// Load config with Polygon network
	cfg, err := config.LoadConfig("config-test-polygon.yaml")
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		return
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("❌ Config validation failed: %v\n", err)
		return
	}

	fmt.Println("✅ Config loaded and validated successfully")

	// Check that Polygon network is present
	polygonCfg, exists := cfg.Networks["polygon"]
	if !exists {
		fmt.Println("❌ Polygon network not found in config")
		return
	}

	fmt.Printf("✅ Polygon network found:\n")
	fmt.Printf("   - Chain ID: %d\n", polygonCfg.ChainID)
	fmt.Printf("   - USDC Contract: %s\n", polygonCfg.USDCContract)
	fmt.Printf("   - RPC URL: %s\n", polygonCfg.RPCURL)
	fmt.Printf("   - Payee Address: %s\n", polygonCfg.PayeeAddress)

	// Try to create a payment requirement for Polygon (without actual RPC call)
	generator := x402.NewPaymentRequirementGenerator(cfg)
	
	// This should work without any code changes!
	req, err := generator.Generate("50000", "polygon")
	if err != nil {
		fmt.Printf("❌ Failed to generate payment requirement: %v\n", err)
		return
	}

	fmt.Println("✅ Payment requirement generated successfully for Polygon:")
	fmt.Printf("   - Network: %s\n", req.Network)
	fmt.Printf("   - Amount: %s\n", req.MaxAmountRequired)
	fmt.Printf("   - Asset: %s\n", req.Asset)
	fmt.Printf("   - PayTo: %s\n", req.PayTo)

	fmt.Println("\n🎉 Extensibility test PASSED - No code changes needed to add Polygon network!")
}
