// internal/cache/config.go
package cache

import "time"

// Config holds configuration for all cache layers used by the service.
//
// We have two (and potentially more) semantically different caches:
//   1. Responses  → final rendered output (HTML, JSON, PNG, text, etc.) — should stay fast/in-memory
//   2. Weather    → raw upstream data from weather backends (WWO, etc.) — can be disk-backed and persistent
//   3. Future     → location, IP geodata, translations, etc.
type Config struct {
	// Responses is the cache for fully generated responses (what the client receives).
	// This is the existing LRU cache passed to NewWeatherService.
	Responses CacheLayer `yaml:"responses"`

	// Weather is the cache for raw weather data coming from the upstream API(s).
	// This is what the new CachedWeatherClient will use.
	Weather CacheLayer `yaml:"weather"`

	// Future data-source caches can be added here without breaking the structure:
	// Location CacheLayer `yaml:"location"`
	// IP       CacheLayer `yaml:"ip"`
	// ...
}

// CacheLayer defines the settings for a single cache layer.
// Different layers can use completely different backends and TTLs.
type CacheLayer struct {
	// Type selects the implementation: "lru", "disk", "memory", or "disabled"
	Type string `yaml:"type"`

	// LRU-specific
	Size int `yaml:"size"`

	// Disk-specific
	Dir             string        `yaml:"dir"`
	MaxSizeMB       int64         `yaml:"max_size_mb"`
	CleanupInterval time.Duration `yaml:"cleanup_interval"`

	// Common to all layers
	TTL     time.Duration `yaml:"ttl"`     // default TTL for entries in this layer
	Enabled bool          `yaml:"enabled"` // quick way to turn a layer off
}

// Helper methods (optional but very convenient)

func (c CacheLayer) IsEnabled() bool {
	return c.Enabled && c.Type != "disabled"
}

func (c CacheLayer) IsDisk() bool {
	return c.Type == "disk" || c.Type == "diskv"
}

func (c CacheLayer) IsLRU() bool {
	return c.Type == "lru" || c.Type == ""
}