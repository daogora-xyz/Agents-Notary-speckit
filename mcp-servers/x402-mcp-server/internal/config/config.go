package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the complete MCP server configuration
type Config struct {
	Networks map[string]NetworkConfig `yaml:"networks"`
	EIP712   EIP712Config             `yaml:"eip712"`
	Logging  LoggingConfig            `yaml:"logging"`
	Cache    CacheConfig              `yaml:"cache"`
}

// EIP712Config contains EIP-712 domain parameters
type EIP712Config struct {
	DomainName    string `yaml:"domain_name"`    // "USD Coin"
	DomainVersion string `yaml:"domain_version"` // "2"
}

// LoggingConfig defines logging behavior
type LoggingConfig struct {
	Level  string `yaml:"level"`  // DEBUG, INFO, WARN, ERROR
	Format string `yaml:"format"` // json
}

// CacheConfig defines cache behavior for settlement idempotency
type CacheConfig struct {
	SettlementTTLMinutes int `yaml:"settlement_ttl_minutes"` // 10
}

// LoadConfig reads and parses the YAML configuration file
func LoadConfig(path string) (*Config, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables
	expanded := os.ExpandEnv(string(data))

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// Validate checks that required configuration is present
func (c *Config) Validate() error {
	if len(c.Networks) == 0 {
		return fmt.Errorf("at least one network must be configured")
	}

	for name, network := range c.Networks {
		if err := network.Validate(); err != nil {
			return fmt.Errorf("network %s: %w", name, err)
		}
	}

	if c.EIP712.DomainName == "" {
		return fmt.Errorf("eip712.domain_name is required")
	}

	if c.EIP712.DomainVersion == "" {
		return fmt.Errorf("eip712.domain_version is required")
	}

	if c.Cache.SettlementTTLMinutes <= 0 {
		return fmt.Errorf("cache.settlement_ttl_minutes must be > 0")
	}

	return nil
}
