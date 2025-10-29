package unit

import (
	"testing"
	"time"

	"github.com/lessuseless/agents-notary/mcp-servers/x402-mcp-server/internal/cache"
)

func TestTTLCache_SetAndGet(t *testing.T) {
	c := cache.NewTTLCache(1 * time.Minute)

	// Set value
	c.Set("key1", "value1")

	// Get value
	val, found := c.Get("key1")
	if !found {
		t.Error("Expected to find key1")
	}

	if val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}
}

func TestTTLCache_Expiry(t *testing.T) {
	c := cache.NewTTLCache(100 * time.Millisecond)

	// Set value with short TTL
	c.Set("temp_key", "temp_value")

	// Should exist immediately
	if _, found := c.Get("temp_key"); !found {
		t.Error("Key should exist immediately after set")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	if _, found := c.Get("temp_key"); found {
		t.Error("Key should have expired")
	}
}

func TestTTLCache_Idempotency(t *testing.T) {
	c := cache.NewTTLCache(1 * time.Minute)

	// Simulate settlement result caching by nonce
	nonce := "0x1234567890abcdef"
	settlementResult := map[string]interface{}{
		"status":  "settled",
		"tx_hash": "0xabcdef",
	}

	// First settlement
	c.Set(nonce, settlementResult)

	// Subsequent calls should get cached result
	cached, found := c.Get(nonce)
	if !found {
		t.Error("Settlement result should be cached")
	}

	cachedMap, ok := cached.(map[string]interface{})
	if !ok {
		t.Fatal("Cached value should be map")
	}

	if cachedMap["status"] != "settled" {
		t.Errorf("Expected status 'settled', got %v", cachedMap["status"])
	}

	// Verify second get returns same result (idempotency)
	cached2, found2 := c.Get(nonce)
	if !found2 {
		t.Error("Second get should also find cached result")
	}

	if cached != cached2 {
		t.Error("Should return same cached instance")
	}
}

func TestTTLCache_Delete(t *testing.T) {
	c := cache.NewTTLCache(1 * time.Minute)

	c.Set("key_to_delete", "value")

	// Verify it exists
	if _, found := c.Get("key_to_delete"); !found {
		t.Error("Key should exist before delete")
	}

	// Delete
	c.Delete("key_to_delete")

	// Verify it's gone
	if _, found := c.Get("key_to_delete"); found {
		t.Error("Key should not exist after delete")
	}
}

func TestTTLCache_Clear(t *testing.T) {
	c := cache.NewTTLCache(1 * time.Minute)

	// Add multiple entries
	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Set("key3", "value3")

	if c.Size() != 3 {
		t.Errorf("Expected size 3, got %d", c.Size())
	}

	// Clear all
	c.Clear()

	if c.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", c.Size())
	}

	// Verify none exist
	if _, found := c.Get("key1"); found {
		t.Error("key1 should not exist after clear")
	}
}

func TestTTLCache_CustomTTL(t *testing.T) {
	c := cache.NewTTLCache(1 * time.Minute) // Default TTL

	// Set with custom shorter TTL
	c.SetWithTTL("short_lived", "value", 50*time.Millisecond)

	// Should exist immediately
	if _, found := c.Get("short_lived"); !found {
		t.Error("Key should exist immediately")
	}

	// Wait for custom TTL expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	if _, found := c.Get("short_lived"); found {
		t.Error("Key should have expired with custom TTL")
	}
}
