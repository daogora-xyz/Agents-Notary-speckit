package unit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/config"
)

func TestLoadConfig_HappyPath(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
networks:
  base:
    chain_id: 8453
    usdc_contract: "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
    facilitator_url: "https://api.cdp.coinbase.com"
    rpc_url: "https://mainnet.base.org"
    payee_address: "0x1234567890123456789012345678901234567890"

eip712:
  domain_name: "USD Coin"
  domain_version: "2"

logging:
  level: "INFO"
  format: "json"

cache:
  settlement_ttl_minutes: 10
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Load config
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Validate loaded values
	if cfg.EIP712.DomainName != "USD Coin" {
		t.Errorf("Expected domain name 'USD Coin', got %s", cfg.EIP712.DomainName)
	}

	if cfg.Cache.SettlementTTLMinutes != 10 {
		t.Errorf("Expected settlement TTL 10, got %d", cfg.Cache.SettlementTTLMinutes)
	}

	// Validate network config
	base, exists := cfg.Networks["base"]
	if !exists {
		t.Fatal("Expected 'base' network to exist")
	}

	if base.ChainID != 8453 {
		t.Errorf("Expected chain ID 8453, got %d", base.ChainID)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	invalidContent := `
networks:
  base:
    - invalid yaml structure
`

	if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	_, err := config.LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestNetworkConfig_Validate_InvalidChainID(t *testing.T) {
	nc := config.NetworkConfig{
		ChainID:        9999, // Invalid
		USDCContract:   "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
		FacilitatorURL: "https://api.cdp.coinbase.com",
		RPCURL:         "https://mainnet.base.org",
		PayeeAddress:   "0x1234567890123456789012345678901234567890",
	}

	err := nc.Validate()
	if err == nil {
		t.Error("Expected validation error for invalid chain ID")
	}
}

func TestNetworkConfig_Validate_MalformedAddress(t *testing.T) {
	nc := config.NetworkConfig{
		ChainID:        8453,
		USDCContract:   "invalid_address", // Malformed
		FacilitatorURL: "https://api.cdp.coinbase.com",
		RPCURL:         "https://mainnet.base.org",
		PayeeAddress:   "0x1234567890123456789012345678901234567890",
	}

	err := nc.Validate()
	if err == nil {
		t.Error("Expected validation error for malformed USDC address")
	}
}

func TestNetworkConfig_Validate_AllowedChainIDs(t *testing.T) {
	allowedIDs := []uint64{8453, 84532, 42161}

	for _, chainID := range allowedIDs {
		nc := config.NetworkConfig{
			ChainID:        chainID,
			USDCContract:   "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
			FacilitatorURL: "https://api.cdp.coinbase.com",
			RPCURL:         "https://mainnet.base.org",
			PayeeAddress:   "0x1234567890123456789012345678901234567890",
		}

		if err := nc.Validate(); err != nil {
			t.Errorf("Chain ID %d should be valid, got error: %v", chainID, err)
		}
	}
}
